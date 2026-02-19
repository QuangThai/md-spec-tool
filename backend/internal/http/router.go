package http

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/md-spec-tool/internal/ai"
	"github.com/yourorg/md-spec-tool/internal/config"
	"github.com/yourorg/md-spec-tool/internal/converter"
	"github.com/yourorg/md-spec-tool/internal/http/handlers"
	"github.com/yourorg/md-spec-tool/internal/http/middleware"
	"github.com/yourorg/md-spec-tool/internal/share"
	"github.com/yourorg/md-spec-tool/internal/suggest"
)

// SetupRouterWithCleanup sets up the router and returns a cleanup function for graceful shutdown.
// The cleanup function must be called before exiting to properly close resources like goroutines.
func SetupRouterWithCleanup(cfg *config.Config) (*gin.Engine, func()) {
	router, cleanup := setupRouterInternal(cfg, true)
	return router, cleanup
}

func SetupRouter(cfg *config.Config) *gin.Engine {
	router, _ := setupRouterInternal(cfg, false)
	return router
}

func setupRouterInternal(cfg *config.Config, withCleanup bool) (*gin.Engine, func()) {
	router := gin.Default()
	if err := router.SetTrustedProxies(cfg.TrustedProxies); err != nil {
		slog.Error("Failed to set trusted proxies", "error", err)
	}
	// MaxMultipartMemory controls when to spill to disk, should be much smaller than MaxUploadBytes
	// to avoid OOM under concurrent load. Use 8MB buffer for safety under concurrent uploads.
	router.MaxMultipartMemory = 8 * 1024 * 1024 // 8MB

	// Initialize quota store
	quotaStore := handlers.NewInMemoryQuotaStore()

	// Apply middlewares (order matters: CORS first, then RequestID, metrics, then error handler)
	router.Use(middleware.CORS(cfg))
	router.Use(middleware.RequestID())
	router.Use(middleware.SessionID()) // Add session ID middleware
	router.Use(middleware.MetricsMiddleware())
	router.Use(middleware.APITelemetryEvents())
	router.Use(middleware.ErrorHandler())

	// Public routes
	router.GET("/health", handlers.HealthHandler)
	router.GET("/metrics", handlers.MetricsHandler)
	telemetryHandler := handlers.NewTelemetryHandler()
	router.POST("/api/telemetry/events", telemetryHandler.IngestEvents)
	router.GET("/api/telemetry/dashboard", telemetryHandler.Dashboard)

	// Create quota handler (will be injected into handlers below)
	quotaHandler := handlers.NewQuotaHandler(quotaStore)
	router.GET("/api/quota/status", quotaHandler.GetQuotaStatus)
	router.GET("/api/quota/daily-report", quotaHandler.GetDailyReport)

	// Create shared HTTP client for outbound requests (Google Sheets, etc.)
	httpClient := &http.Client{
		Timeout: cfg.HTTPClientTimeout,
	}

	// MDFlow converter routes (public, no auth required)
	resolveModel := func(primary, fallback string) string {
		if primary != "" {
			return primary
		}
		return fallback
	}
	buildAIService := func(model string, timeout time.Duration, maxTokens int) ai.Service {
		if !cfg.AIEnabled {
			return nil
		}
		aiConfig := ai.DefaultConfig()
		aiConfig.Model = model
		aiConfig.APIKey = cfg.OpenAIAPIKey
		aiConfig.RequestTimeout = timeout
		aiConfig.MaxRetries = cfg.AIMaxRetries
		aiConfig.CacheTTL = cfg.AICacheTTL
		aiConfig.MaxCacheSize = cfg.AIMaxCacheSize
		aiConfig.RetryBaseDelay = cfg.AIRetryBaseDelay
		aiConfig.MaxCompletionTokens = maxTokens
		aiConfig.PromptProfile = cfg.AIPromptProfile

		svc, err := ai.NewService(aiConfig)
		if err != nil {
			slog.Warn("AI service initialization failed", "model", model, "error", err)
			return nil
		}
		slog.Info("AI service initialized", "model", model, "timeout", timeout, "max_tokens", maxTokens)
		return svc
	}

	convertModel := resolveModel(cfg.OpenAIConvertModel, cfg.OpenAIModel)
	previewModel := resolveModel(cfg.OpenAIPreviewModel, convertModel)
	suggestModel := resolveModel(cfg.OpenAISuggestModel, convertModel)

	// Create shared AI service for all AI operations (auto-enabled when OPENAI_API_KEY is set)
	convertAIService := buildAIService(convertModel, cfg.AIRequestTimeout, cfg.AIConvertMaxTokens)
	previewAIService := buildAIService(previewModel, cfg.AIPreviewTimeout, cfg.AIPreviewMaxTokens)
	suggestAIService := buildAIService(suggestModel, cfg.AISuggestTimeout, cfg.AISuggestMaxTokens)

	convForConvert := converter.NewConverter().WithAIService(convertAIService)
	convForPreview := converter.NewConverter().WithAIService(previewAIService)

	// Create AIServiceProvider for BYOK support across all handlers
	aiProvider := handlers.NewAIServiceProvider(cfg)
	aiProvider.SetDefaultAIService(convertAIService)

	// Create specialized handlers (Phase 5.3 decomposition)
	convertHandler := handlers.NewConvertHandler(convForConvert, cfg, aiProvider)
	previewHandler := handlers.NewPreviewHandler(convForPreview, cfg, aiProvider)
	templateHandler := handlers.NewTemplateHandler(converter.NewMDFlowRenderer(), cfg)
	validationHandler := handlers.NewValidationHandler(cfg, aiProvider)

	// Inject quota handler for token tracking (created above)
	convertHandler.SetQuotaHandler(quotaHandler)

	// Create MDFlowHandler for legacy AI suggestions (will be refactored in future)
	// Inject shared AIServiceProvider to avoid duplicate caches
	mdflowHandler := handlers.NewMDFlowHandlerWithProvider(
		convForConvert,
		converter.NewMDFlowRenderer(),
		httpClient,
		cfg,
		aiProvider,
	)

	// Create GSheetHandler with injected AI service factory.
	// Use shared provider cache to avoid creating a new BYOK service on every request.
	getAIService := func(apiKey string) (handlers.Service, error) {
		return aiProvider.GetAIServiceForAPIKey(apiKey)
	}

	gsheetHandler := handlers.NewGSheetHandler(
		convForConvert,
		converter.NewMDFlowRenderer(),
		httpClient,
		cfg,
		getAIService,
	)

	audioHandler := handlers.NewAudioTranscribeHandler(cfg)

	// Create share handler and store (needed early for clone-template route)
	shareStore := share.NewStore(cfg.ShareStorePath)
	shareHandler := handlers.NewShareHandler(shareStore)

	// Create diff handler (always created; supports BYOK even when no server AI key)
	diffHandler := handlers.NewDiffHandler(aiProvider, cfg)
	if suggestAIService != nil {
		aiSuggester := suggest.NewSuggester(suggestAIService)
		mdflowHandler.SetAISuggester(aiSuggester)
		mdflowHandler.SetAIService(suggestAIService)
	}

	// API v1 routes (versioned endpoint for future compatibility)
	previewRateLimit := middleware.RateLimit(cfg.PreviewRateLimit, cfg.RateLimitWindow)
	convertRateLimit := middleware.RateLimit(cfg.ConvertRateLimit, cfg.RateLimitWindow)
	aiSuggestRateLimit := middleware.RateLimit(cfg.AISuggestRateLimit, cfg.RateLimitWindow)
	quotaCheck := middleware.QuotaMiddleware(quotaHandler)

	v1 := router.Group("/api/v1/mdflow")
	{
		// Convert endpoints (Phase 5.3: specialized ConvertHandler)
		v1.POST("/paste", convertRateLimit, quotaCheck, convertHandler.ConvertPaste)
		v1.POST("/tsv", convertRateLimit, quotaCheck, convertHandler.ConvertTSV)
		v1.POST("/xlsx", convertRateLimit, quotaCheck, convertHandler.ConvertXLSX)
		v1.POST("/xlsx/sheets", convertHandler.GetXLSXSheets)

		// Preview endpoints (Phase 5.3: specialized PreviewHandler)
		v1.POST("/preview", previewRateLimit, quotaCheck, previewHandler.PreviewPaste)
		v1.POST("/tsv/preview", previewRateLimit, quotaCheck, previewHandler.PreviewTSV)
		v1.POST("/xlsx/preview", previewRateLimit, quotaCheck, previewHandler.PreviewXLSX)

		// Template endpoints (Phase 5.3: specialized TemplateHandler)
		v1.GET("/templates", templateHandler.GetTemplates)
		v1.GET("/templates/info", templateHandler.GetTemplateInfo)
		v1.GET("/templates/:name", templateHandler.GetTemplateContent)
		v1.POST("/templates/preview", templateHandler.PreviewTemplate)
		v1.POST("/clone-template", shareHandler.CloneTemplate)

		// Validation endpoints (Phase 5.3: specialized ValidationHandler)
		v1.POST("/validate", validationHandler.Validate)

		// Other routes (diff, gsheet, suggestions)
		v1.POST("/diff", quotaCheck, diffHandler.DiffMDFlow)
		v1.POST("/gsheet", gsheetHandler.FetchGoogleSheet)
		v1.POST("/gsheet/sheets", gsheetHandler.GetGoogleSheetSheets)
		v1.POST("/gsheet/preview", previewRateLimit, quotaCheck, gsheetHandler.PreviewGoogleSheet)
		v1.POST("/gsheet/convert", convertRateLimit, quotaCheck, gsheetHandler.ConvertGoogleSheet)
		v1.POST("/ai/suggest", aiSuggestRateLimit, quotaCheck, mdflowHandler.GetAISuggestions)
	}

	// Legacy routes (deprecated, point to v1) - support for backward compatibility
	mdflow := router.Group("/api/mdflow")
	{
		// Convert endpoints
		mdflow.POST("/paste", convertRateLimit, quotaCheck, convertHandler.ConvertPaste)
		mdflow.POST("/tsv", convertRateLimit, quotaCheck, convertHandler.ConvertTSV)
		mdflow.POST("/xlsx", convertRateLimit, quotaCheck, convertHandler.ConvertXLSX)
		mdflow.POST("/xlsx/sheets", convertHandler.GetXLSXSheets)

		// Preview endpoints
		mdflow.POST("/preview", previewRateLimit, quotaCheck, previewHandler.PreviewPaste)
		mdflow.POST("/tsv/preview", previewRateLimit, quotaCheck, previewHandler.PreviewTSV)
		mdflow.POST("/xlsx/preview", previewRateLimit, quotaCheck, previewHandler.PreviewXLSX)

		// Template endpoints
		mdflow.GET("/templates", templateHandler.GetTemplates)
		mdflow.GET("/templates/info", templateHandler.GetTemplateInfo)
		mdflow.GET("/templates/:name", templateHandler.GetTemplateContent)
		mdflow.POST("/templates/preview", templateHandler.PreviewTemplate)
		mdflow.POST("/clone-template", shareHandler.CloneTemplate)

		// Validation endpoints
		mdflow.POST("/validate", validationHandler.Validate)

		// Other routes
		mdflow.POST("/diff", quotaCheck, diffHandler.DiffMDFlow)
		mdflow.POST("/gsheet", gsheetHandler.FetchGoogleSheet)
		mdflow.POST("/gsheet/sheets", gsheetHandler.GetGoogleSheetSheets)
		mdflow.POST("/gsheet/preview", previewRateLimit, quotaCheck, gsheetHandler.PreviewGoogleSheet)
		mdflow.POST("/gsheet/convert", convertRateLimit, quotaCheck, gsheetHandler.ConvertGoogleSheet)
		mdflow.POST("/ai/suggest", aiSuggestRateLimit, quotaCheck, mdflowHandler.GetAISuggestions)
	}

	shareRoutes := router.Group("/api/share")
	{
		shareRoutes.POST("", middleware.RateLimit(cfg.ShareCreateRateLimit, cfg.RateLimitWindow), shareHandler.CreateShare)
		shareRoutes.GET("/public", shareHandler.ListPublic)
		shareRoutes.GET("/:key", shareHandler.GetShare)
		shareRoutes.PATCH("/:key", middleware.RateLimit(cfg.ShareUpdateRateLimit, cfg.RateLimitWindow), shareHandler.UpdateShare)
		shareRoutes.GET("/:key/comments", shareHandler.ListComments)
		shareRoutes.GET("/:key/events", shareHandler.GetShareEvents)
		shareRoutes.POST("/:key/comments", middleware.RateLimit(cfg.ShareCommentRateLimit, cfg.RateLimitWindow), shareHandler.CreateComment)
		shareRoutes.PATCH("/:key/comments/:commentId", middleware.RateLimit(cfg.ShareCommentRateLimit, cfg.RateLimitWindow), shareHandler.UpdateComment)
	}

	audio := router.Group("/api/audio")
	{
		audio.POST("/transcribe", audioHandler.Transcribe)
	}

	// Return cleanup function that closes all handlers with lifecycle management
	cleanup := func() {
		if aiProvider != nil {
			aiProvider.Close()
		}
		if mdflowHandler != nil {
			mdflowHandler.Close()
		}
		// Quota store cleanup (in-memory, no external resources)
		if quotaStore != nil {
			_ = quotaStore.Cleanup(context.Background())
		}
		slog.Debug("All handlers closed successfully")
	}

	return router, cleanup
}
