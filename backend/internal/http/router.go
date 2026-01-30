package http

import (
	"github.com/gin-gonic/gin"
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
		mdflow.POST("/diff", handlers.DiffMDFlow())
		mdflow.POST("/gsheet", mdflowHandler.FetchGoogleSheet)
		mdflow.POST("/gsheet/convert", mdflowHandler.ConvertGoogleSheet)
	}

	return router
}
