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

func SetupRouter(cfg *config.Config) *gin.Engine {
	router := gin.Default()
	// MaxMultipartMemory controls when to spill to disk, should be much smaller than MaxUploadBytes
	// to avoid OOM under concurrent load. Use 8MB buffer for safety under concurrent uploads.
	router.MaxMultipartMemory = 8 * 1024 * 1024 // 8MB

	// Apply middlewares
	router.Use(middleware.CORS(cfg))

	// Public routes
	router.GET("/health", handlers.HealthHandler)

	// Create shared HTTP client for outbound requests (Google Sheets, etc.)
	httpClient := &http.Client{
		Timeout: cfg.HTTPClientTimeout,
	}

	// MDFlow converter routes (public, no auth required)
	conv := converter.NewConverter()
	if cfg.UseNewConverterPipeline {
		conv.WithNewPipeline()
	}

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

	mdflowHandler := handlers.NewMDFlowHandler(
		conv,
		converter.NewMDFlowRenderer(),
		httpClient,
		cfg,
	)

	// Create diff handler (always created; supports BYOK even when no server AI key)
	diffHandler := handlers.NewDiffHandler(aiService, cfg)
	if aiService != nil {
		aiSuggester := suggest.NewSuggester(aiService)
		mdflowHandler.SetAISuggester(aiSuggester)
		mdflowHandler.SetAIService(aiService)
	}

	mdflow := router.Group("/api/mdflow")
	{
		mdflow.POST("/paste", mdflowHandler.ConvertPaste)
		mdflow.POST("/preview", mdflowHandler.PreviewPaste)
		mdflow.POST("/tsv", mdflowHandler.ConvertTSV)
		mdflow.POST("/tsv/preview", mdflowHandler.PreviewTSV)
		mdflow.POST("/xlsx", mdflowHandler.ConvertXLSX)
		mdflow.POST("/xlsx/sheets", mdflowHandler.GetXLSXSheets)
		mdflow.POST("/xlsx/preview", mdflowHandler.PreviewXLSX)
		mdflow.GET("/templates", mdflowHandler.GetTemplates)
		mdflow.GET("/templates/info", mdflowHandler.GetTemplateInfo)
		mdflow.GET("/templates/:name", mdflowHandler.GetTemplateContent)
		mdflow.POST("/templates/preview", mdflowHandler.PreviewTemplate)
		mdflow.POST("/diff", diffHandler.DiffMDFlow)
		mdflow.POST("/gsheet", mdflowHandler.FetchGoogleSheet)
		mdflow.POST("/gsheet/sheets", mdflowHandler.GetGoogleSheetSheets)
		mdflow.POST("/gsheet/preview", mdflowHandler.PreviewGoogleSheet)
		mdflow.POST("/gsheet/convert", mdflowHandler.ConvertGoogleSheet)
		mdflow.POST("/validate", mdflowHandler.Validate)
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

	return router
}
