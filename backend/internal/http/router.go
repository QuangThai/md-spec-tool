package http

import (
	"github.com/gin-gonic/gin"
	"github.com/yourorg/md-spec-tool/internal/ai"
	"github.com/yourorg/md-spec-tool/internal/config"
	"github.com/yourorg/md-spec-tool/internal/http/handlers"
	"github.com/yourorg/md-spec-tool/internal/http/middleware"
)

func SetupRouter(cfg *config.Config) *gin.Engine {
	router := gin.Default()
	router.MaxMultipartMemory = 8 << 20

	// Apply middlewares
	router.Use(middleware.CORS(cfg))

	// Public routes
	router.GET("/health", handlers.HealthHandler)

	// MDFlow converter routes (public, no auth required)
	mdflowHandler := handlers.NewMDFlowHandler()

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

	return router
}
