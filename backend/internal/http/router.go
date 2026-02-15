package http

import (
	"log/slog"
	"net/http"

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
	// MaxMultipartMemory controls when to spill to disk, should be much smaller than MaxUploadBytes
	// to avoid OOM under concurrent load. Use 8MB buffer for safety under concurrent uploads.
	router.MaxMultipartMemory = 8 * 1024 * 1024 // 8MB

	// Apply middlewares (order matters: CORS first, then RequestID, metrics, then error handler)
	router.Use(middleware.CORS(cfg))
	router.Use(middleware.RequestID())
	router.Use(middleware.MetricsMiddleware())
	router.Use(middleware.ErrorHandler())

	// Public routes
	router.GET("/health", handlers.HealthHandler)
	router.GET("/metrics", handlers.MetricsHandler)

	// Create shared HTTP client for outbound requests (Google Sheets, etc.)
	httpClient := &http.Client{
		Timeout: cfg.HTTPClientTimeout,
	}

	// MDFlow converter routes (public, no auth required)
	conv := converter.NewConverter()

	// Create shared AI service for all AI operations (auto-enabled when OPENAI_API_KEY is set)
	var aiService ai.Service
	if cfg.AIEnabled {
		aiConfig := ai.DefaultConfig()
		aiConfig.Model = cfg.OpenAIModel
		aiConfig.APIKey = cfg.OpenAIAPIKey
		aiConfig.RequestTimeout = cfg.AIRequestTimeout
		aiConfig.MaxRetries = cfg.AIMaxRetries
		aiConfig.CacheTTL = cfg.AICacheTTL
		aiConfig.MaxCacheSize = cfg.AIMaxCacheSize
		aiConfig.RetryBaseDelay = cfg.AIRetryBaseDelay

		var err error
		aiService, err = ai.NewService(aiConfig)
		if err != nil {
			slog.Warn("AI service initialization failed", "error", err)
		} else {
			conv.WithAIService(aiService)
			slog.Info("AI service initialized", "model", aiConfig.Model)
		}
	}

	// Create AIServiceProvider for BYOK support across all handlers
	aiProvider := handlers.NewAIServiceProvider(cfg)
	aiProvider.SetDefaultAIService(aiService)

	// Create specialized handlers (Phase 5.3 decomposition)
	convertHandler := handlers.NewConvertHandler(conv, cfg, aiProvider)
	previewHandler := handlers.NewPreviewHandler(conv, cfg, aiProvider)
	templateHandler := handlers.NewTemplateHandler(converter.NewMDFlowRenderer(), cfg)
	validationHandler := handlers.NewValidationHandler(cfg, aiProvider)

	// Create MDFlowHandler for legacy AI suggestions (will be refactored in future)
	// Inject shared AIServiceProvider to avoid duplicate caches
	mdflowHandler := handlers.NewMDFlowHandlerWithProvider(
		conv,
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
		conv,
		converter.NewMDFlowRenderer(),
		httpClient,
		cfg,
		getAIService,
	)

	audioHandler := handlers.NewAudioTranscribeHandler(cfg)

	// Create diff handler (always created; supports BYOK even when no server AI key)
	diffHandler := handlers.NewDiffHandler(aiProvider, cfg)
	if aiService != nil {
		aiSuggester := suggest.NewSuggester(aiService)
		mdflowHandler.SetAISuggester(aiSuggester)
		mdflowHandler.SetAIService(aiService)
	}

	// API v1 routes (versioned endpoint for future compatibility)
	v1 := router.Group("/api/v1/mdflow")
	{
		// Convert endpoints (Phase 5.3: specialized ConvertHandler)
		v1.POST("/paste", convertHandler.ConvertPaste)
		v1.POST("/tsv", convertHandler.ConvertTSV)
		v1.POST("/xlsx", convertHandler.ConvertXLSX)
		v1.POST("/xlsx/sheets", convertHandler.GetXLSXSheets)

		// Preview endpoints (Phase 5.3: specialized PreviewHandler)
		v1.POST("/preview", previewHandler.PreviewPaste)
		v1.POST("/tsv/preview", previewHandler.PreviewTSV)
		v1.POST("/xlsx/preview", previewHandler.PreviewXLSX)

		// Template endpoints (Phase 5.3: specialized TemplateHandler)
		v1.GET("/templates", templateHandler.GetTemplates)
		v1.GET("/templates/info", templateHandler.GetTemplateInfo)
		v1.GET("/templates/:name", templateHandler.GetTemplateContent)
		v1.POST("/templates/preview", templateHandler.PreviewTemplate)

		// Validation endpoints (Phase 5.3: specialized ValidationHandler)
		v1.POST("/validate", validationHandler.Validate)

		// Other routes (diff, gsheet, suggestions)
		v1.POST("/diff", diffHandler.DiffMDFlow)
		v1.POST("/gsheet", gsheetHandler.FetchGoogleSheet)
		v1.POST("/gsheet/sheets", gsheetHandler.GetGoogleSheetSheets)
		v1.POST("/gsheet/preview", gsheetHandler.PreviewGoogleSheet)
		v1.POST("/gsheet/convert", gsheetHandler.ConvertGoogleSheet)
		v1.POST("/ai/suggest", mdflowHandler.GetAISuggestions)
	}

	// Legacy routes (deprecated, point to v1) - support for backward compatibility
	mdflow := router.Group("/api/mdflow")
	{
		// Convert endpoints
		mdflow.POST("/paste", convertHandler.ConvertPaste)
		mdflow.POST("/tsv", convertHandler.ConvertTSV)
		mdflow.POST("/xlsx", convertHandler.ConvertXLSX)
		mdflow.POST("/xlsx/sheets", convertHandler.GetXLSXSheets)

		// Preview endpoints
		mdflow.POST("/preview", previewHandler.PreviewPaste)
		mdflow.POST("/tsv/preview", previewHandler.PreviewTSV)
		mdflow.POST("/xlsx/preview", previewHandler.PreviewXLSX)

		// Template endpoints
		mdflow.GET("/templates", templateHandler.GetTemplates)
		mdflow.GET("/templates/info", templateHandler.GetTemplateInfo)
		mdflow.GET("/templates/:name", templateHandler.GetTemplateContent)
		mdflow.POST("/templates/preview", templateHandler.PreviewTemplate)

		// Validation endpoints
		mdflow.POST("/validate", validationHandler.Validate)

		// Other routes
		mdflow.POST("/diff", diffHandler.DiffMDFlow)
		mdflow.POST("/gsheet", gsheetHandler.FetchGoogleSheet)
		mdflow.POST("/gsheet/sheets", gsheetHandler.GetGoogleSheetSheets)
		mdflow.POST("/gsheet/preview", gsheetHandler.PreviewGoogleSheet)
		mdflow.POST("/gsheet/convert", gsheetHandler.ConvertGoogleSheet)
		mdflow.POST("/ai/suggest", mdflowHandler.GetAISuggestions)
	}

	shareStore := share.NewStore(cfg.ShareStorePath)
	shareHandler := handlers.NewShareHandler(shareStore)

	shareRoutes := router.Group("/api/share")
	{
		shareRoutes.POST("", middleware.RateLimit(cfg.ShareCreateRateLimit, cfg.RateLimitWindow), shareHandler.CreateShare)
		shareRoutes.GET("/public", shareHandler.ListPublic)
		shareRoutes.GET("/:key", shareHandler.GetShare)
		shareRoutes.PATCH("/:key", middleware.RateLimit(cfg.ShareUpdateRateLimit, cfg.RateLimitWindow), shareHandler.UpdateShare)
		shareRoutes.GET("/:key/comments", shareHandler.ListComments)
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
		slog.Debug("All handlers closed successfully")
	}

	return router, cleanup
}
