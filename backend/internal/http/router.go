package http

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/md-spec-tool/internal/ai"
	"github.com/yourorg/md-spec-tool/internal/config"
	"github.com/yourorg/md-spec-tool/internal/converter"
	"github.com/yourorg/md-spec-tool/internal/http/handlers"
	"github.com/yourorg/md-spec-tool/internal/http/middleware"
	"github.com/yourorg/md-spec-tool/internal/share"
)

func SetupRouter(cfg *config.Config) *gin.Engine {
	router := gin.Default()
	router.MaxMultipartMemory = 8 << 20

	// Apply middlewares
	router.Use(middleware.CORS(cfg))

	// Public routes
	router.GET("/health", handlers.HealthHandler)

	// Create shared HTTP client for outbound requests (Google Sheets, OpenAI, etc.)
	// This avoids creating new clients per request and ensures proper timeout handling
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	// MDFlow converter routes (public, no auth required)
	// Inject dependencies: converter, renderer, httpClient
	mdflowHandler := handlers.NewMDFlowHandler(
		converter.NewConverter(),
		converter.NewMDFlowRenderer(),
		httpClient,
	)

	// Configure AI suggester if API key is available
	if cfg.OpenAIAPIKey != "" {
		aiSuggester := ai.NewSuggester(cfg.OpenAIAPIKey, cfg.OpenAIModel)
		mdflowHandler.SetAISuggester(aiSuggester)
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
		mdflow.POST("/diff", handlers.DiffMDFlow())
		mdflow.POST("/gsheet", mdflowHandler.FetchGoogleSheet)
		mdflow.POST("/gsheet/convert", mdflowHandler.ConvertGoogleSheet)
		mdflow.POST("/validate", mdflowHandler.Validate)
		mdflow.POST("/ai/suggest", mdflowHandler.GetAISuggestions)
	}

	shareStore := share.NewStore(cfg.ShareStorePath)
	shareHandler := handlers.NewShareHandler(shareStore)

	shareRoutes := router.Group("/api/share")
	{
		shareRoutes.POST("", middleware.RateLimit(10, time.Minute), shareHandler.CreateShare)
		shareRoutes.GET("/public", shareHandler.ListPublic)
		shareRoutes.GET("/:key", shareHandler.GetShare)
		shareRoutes.PATCH("/:key", middleware.RateLimit(20, time.Minute), shareHandler.UpdateShare)
		shareRoutes.GET("/:key/comments", shareHandler.ListComments)
		shareRoutes.POST("/:key/comments", middleware.RateLimit(20, time.Minute), shareHandler.CreateComment)
		shareRoutes.PATCH("/:key/comments/:commentId", middleware.RateLimit(20, time.Minute), shareHandler.UpdateComment)
	}

	return router
}
