package http

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yourorg/md-spec-tool/internal/config"
	"github.com/yourorg/md-spec-tool/internal/http/handlers"
	"github.com/yourorg/md-spec-tool/internal/http/middleware"
	"github.com/yourorg/md-spec-tool/internal/repositories"
	"github.com/yourorg/md-spec-tool/internal/services"
)

func SetupRouter(cfg *config.Config, pool *pgxpool.Pool) *gin.Engine {
	router := gin.Default()

	// Initialize repositories
	userRepo := repositories.NewUserRepository(pool)
	specRepo := repositories.NewSpecRepository(pool)
	templateRepo := repositories.NewTemplateRepository(pool)
	shareRepo := repositories.NewShareRepository(pool)
	commentRepo := repositories.NewCommentRepository(pool)
	activityRepo := repositories.NewActivityRepository(pool)
	notificationRepo := repositories.NewNotificationRepository(pool)
	mentionRepo := repositories.NewMentionRepository(pool)

	// Initialize services
	authService := services.NewAuthService(userRepo, cfg.JWTSecret)
	excelService := services.NewExcelService()
	specService := services.NewSpecService(specRepo)
	templateService := services.NewTemplateService(templateRepo)
	shareService := services.NewShareService(shareRepo, specRepo, userRepo)
	commentService := services.NewCommentService(commentRepo, specRepo, shareRepo, mentionRepo, notificationRepo, userRepo)
	activityService := services.NewActivityService(activityRepo, notificationRepo, userRepo)
	notificationService := services.NewNotificationService(notificationRepo, mentionRepo, userRepo)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	importHandler := handlers.NewImportHandler(excelService)
	convertHandler := handlers.NewConvertHandler(templateService)
	specHandler := handlers.NewSpecHandler(specService)
	templateHandler := handlers.NewTemplateHandler(templateService)
	shareHandler := handlers.NewShareHandler(shareService)
	commentHandler := handlers.NewCommentHandler(commentService)
	activityHandler := handlers.NewActivityHandler(activityService)
	notificationHandler := handlers.NewNotificationHandler(notificationService)

	// Apply middlewares
	router.Use(middleware.CORS(cfg))

	// Public routes
	router.GET("/health", handlers.HealthHandler)

	// MDFlow converter routes (public, no auth required)
	mdflowHandler := handlers.NewMDFlowHandler()
	mdflow := router.Group("/api/mdflow")
	{
		mdflow.POST("/paste", mdflowHandler.ConvertPaste)
		mdflow.POST("/xlsx", mdflowHandler.ConvertXLSX)
		mdflow.POST("/xlsx/sheets", mdflowHandler.GetXLSXSheets)
		mdflow.GET("/templates", mdflowHandler.GetTemplates)
	}

	// Auth routes
	auth := router.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
	}

	// Protected routes
	protected := router.Group("")
	protected.Use(middleware.Auth(authService))
	{
		// Import routes
		protected.POST("/import/excel", importHandler.UploadExcel)
		protected.GET("/spec/preview/:id", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "TODO: preview"})
		})

		// Conversion routes
		protected.POST("/convert/markdown", convertHandler.ConvertToMarkdown)

		// Template routes
		protected.GET("/templates", templateHandler.ListTemplates)
		protected.GET("/templates/:id", templateHandler.GetTemplate)
		protected.POST("/templates", templateHandler.CreateTemplate)

		// Spec routes
		protected.POST("/spec", specHandler.SaveSpec)
		protected.GET("/spec/:id", specHandler.GetSpec)
		protected.GET("/spec", specHandler.ListSpecs)
		protected.GET("/spec/:id/versions", specHandler.GetVersions)
		protected.PUT("/spec/:id", specHandler.UpdateSpec)
		protected.DELETE("/spec/:id", specHandler.DeleteSpec)
		protected.POST("/spec/search", specHandler.SearchSpecs)
		protected.GET("/spec/:id/download", specHandler.DownloadSpec)

		// Share routes (Phase 5)
		protected.POST("/spec/:id/share", shareHandler.ShareSpec)
		protected.DELETE("/spec/:id/share/:user_id", shareHandler.UnshareSpec)
		protected.GET("/spec/:id/shares", shareHandler.GetSpecShares)
		protected.PUT("/spec/:id/share/:user_id", shareHandler.UpdateSharePermission)
		protected.GET("/spec/shared/mine", shareHandler.GetSharedSpecs)

		// Comment routes (Phase 5)
		protected.POST("/spec/:id/comments", commentHandler.AddComment)
		protected.GET("/spec/:id/comments", commentHandler.GetComments)
		protected.PUT("/spec/:id/comments/:comment_id", commentHandler.UpdateComment)
		protected.DELETE("/spec/:id/comments/:comment_id", commentHandler.DeleteComment)
		protected.POST("/spec/:id/comments/:comment_id/reply", commentHandler.AddReply)

		// Comment edit history routes (Phase 7)
		protected.PUT("/spec/:id/comments/:comment_id/edit", commentHandler.EditComment) // NEW: Edit comment with history
		protected.GET("/comments/:comment_id/edits", commentHandler.GetCommentEdits)      // NEW: View edit history

		// Activity routes (Phase 6)
		protected.GET("/spec/:id/activity", activityHandler.GetSpecActivity)
		protected.GET("/activity/mine", activityHandler.GetUserActivity)

		// Activity filtering & export routes (Phase 7)
		protected.POST("/spec/:id/activities/filter", activityHandler.FilterActivities) // NEW: Filter activities
		protected.GET("/spec/:id/activities/stats", activityHandler.GetActivityStats)    // NEW: Activity stats
		protected.POST("/spec/:id/activities/export", activityHandler.ExportActivities)  // NEW: Export activities

		// Notification routes (Phase 6)
		protected.GET("/notifications", notificationHandler.GetNotifications)
		protected.POST("/notifications/read", notificationHandler.MarkAsRead)
		protected.POST("/notifications/read-all", notificationHandler.MarkAllAsRead)
		protected.DELETE("/notifications/:id", notificationHandler.DeleteNotification)
	}

	return router
}
