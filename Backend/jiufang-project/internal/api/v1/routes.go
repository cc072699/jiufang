package v1

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"jiufang/internal/infrastructure/erp"
	"jiufang/internal/middleware"
	"jiufang/internal/service"
)

func RegisterRoutes(
	r *gin.RouterGroup,
	authMiddleware *middleware.AuthMiddleware,
	profileService *service.ProfileAppService,
	groupService *service.GroupAppService,
	permissionService *service.PermissionAppService,
	authService *service.AuthAppService,
	historyService *service.HistoryAppService,
	dialogFavoriteService *service.DialogFavoriteAppService,
	exportAppService *service.ExportAppService,
	reportService service.ReportServiceInterface,
	feedbackService service.FeedbackServiceInterface,
	alertService *service.AlertService,
	operationLogService *service.OperationLogService,
	queryService *service.QueryAppService,
	userService *service.UserAppService,
	erpReader erp.ERPReaderInterface,
	logger *zap.Logger,
) {
	profileHandler := NewProfileHandler(profileService)
	groupHandler := NewGroupHandler(groupService, permissionService)
	authHandler := NewAuthHandler(authService)
	historyHandler := NewHistoryHandler(historyService, logger)
	favoriteHandler := NewFavoriteHandler(historyService, logger)
	dialogFavoriteHandler := NewDialogFavoriteHandler(dialogFavoriteService, logger)
	exportHandler := NewExportHandler(exportAppService, logger)
	reportHandler := NewReportHandler(reportService, alertService, logger)
	feedbackHandler := NewFeedbackHandler(feedbackService, logger)
	alertHandler := NewAlertHandler(alertService, logger)
	operationLogHandler := NewOperationLogHandler(operationLogService, logger)
	queryHandler := NewQueryHandler(queryService, permissionService, logger)
	userHandler := NewUserHandler(userService, logger)
	metadataHandler := NewMetadataHandler(erpReader, logger)

	auth := r.Group("/auth")
	{
		auth.POST("/login", authHandler.Login)
		auth.POST("/logout", authHandler.Logout)
	}

	// Natural language query routes
	query := r.Group("/query")
	query.Use(authMiddleware.Authenticate()) // JWT authentication required
	{
		query.POST("", queryHandler.ExecuteQuery)
	}

	// ERP metadata routes (table list, schema)
	metadata := r.Group("/metadata")
	metadata.Use(authMiddleware.Authenticate())
	{
		metadata.GET("/tables", metadataHandler.GetTables)
	}

	// User management routes (Admin only)
	users := r.Group("/users")
	users.Use(authMiddleware.Authenticate()) // JWT authentication required
	users.Use(authMiddleware.RequireAdmin()) // Admin permission required
	{
		users.POST("", userHandler.CreateUser)
		users.GET("", userHandler.ListUsers)
		users.GET("/:user_id", userHandler.GetUser)
		users.PUT("/:user_id", userHandler.UpdateUser)
		users.DELETE("/:user_id", userHandler.DeleteUser)
	}

	profile := r.Group("/profile")
	profile.Use(authMiddleware.Authenticate()) // JWT authentication required
	{
		profile.GET("", profileHandler.GetProfile)
		profile.POST("/avatar", profileHandler.UploadAvatar)
		profile.PUT("/password", profileHandler.ChangePassword)
	}

	groups := r.Group("/groups")
	groups.Use(authMiddleware.Authenticate()) // JWT authentication required
	groups.Use(authMiddleware.RequireAdmin()) // Admin permission required
	{
		groups.POST("", groupHandler.CreateGroup)
		groups.GET("", groupHandler.ListGroups)
		groups.GET("/:group_id", groupHandler.GetGroup)
		groups.PUT("/:group_id", groupHandler.UpdateGroup)
		groups.DELETE("/:group_id", groupHandler.DeleteGroup)
		groups.POST("/:group_id/permissions", groupHandler.ConfigurePermissions)
		groups.GET("/:group_id/permissions", groupHandler.GetPermissions)
		groups.GET("/:group_id/members", groupHandler.GetGroupMembers)
		groups.POST("/:group_id/members", groupHandler.AddGroupMembers)
		groups.DELETE("/:group_id/members/:user_id", groupHandler.RemoveGroupMember)
	}

	// Query history routes
	history := r.Group("/history")
	history.Use(authMiddleware.Authenticate()) // JWT authentication required
	{
		history.GET("", historyHandler.GetHistoryList)
		history.GET("/session/:session_id", historyHandler.GetHistoryBySessionID)
		history.GET("/:record_id", historyHandler.GetHistoryDetail)
		history.DELETE("/:record_id", historyHandler.DeleteHistory)
	}

	// Favorite routes (for query records)
	favorites := r.Group("/favorites")
	favorites.Use(authMiddleware.Authenticate()) // JWT authentication required
	{
		favorites.POST("", favoriteHandler.CreateFavorite)
		favorites.GET("", favoriteHandler.GetFavoriteList)
		favorites.DELETE("/:favorite_id", favoriteHandler.DeleteFavorite)
	}

	// Dialog favorite routes (for dialog sessions)
	dialogFavorites := r.Group("/dialog-favorites")
	dialogFavorites.Use(authMiddleware.Authenticate()) // JWT authentication required
	{
		dialogFavorites.POST("", dialogFavoriteHandler.CreateDialogFavorite)
		dialogFavorites.GET("", dialogFavoriteHandler.GetDialogFavoriteList)
		dialogFavorites.DELETE("/:favorite_id", dialogFavoriteHandler.DeleteDialogFavorite)
	}

	// Export routes
	export := r.Group("/export")
	export.Use(authMiddleware.Authenticate()) // JWT authentication required
	{
		export.POST("", exportHandler.ExportQueryResult)
		export.GET("/records", exportHandler.GetExportRecords)
	}

	// Scheduled report routes (Admin only)
	reports := r.Group("/reports")
	reports.Use(authMiddleware.Authenticate()) // JWT authentication required
	reports.Use(authMiddleware.RequireAdmin()) // Admin permission required
	{
		reports.POST("", reportHandler.CreateReport)
		reports.GET("", reportHandler.ListReports)
		reports.GET("/:id", reportHandler.GetReport)
		reports.PUT("/:id", reportHandler.UpdateReport)
		reports.DELETE("/:id", reportHandler.DeleteReport)
	}

	// Push record routes (Admin only)
	pushRecords := r.Group("/push-records")
	pushRecords.Use(authMiddleware.Authenticate()) // JWT authentication required
	pushRecords.Use(authMiddleware.RequireAdmin()) // Admin permission required
	{
		pushRecords.GET("", reportHandler.ListPushRecords)
	}

	// Feedback routes
	feedbacks := r.Group("/feedbacks")
	feedbacks.Use(authMiddleware.Authenticate()) // JWT authentication required
	{
		feedbacks.POST("", feedbackHandler.CreateFeedback) // All users can create feedback
		// Admin only routes
		feedbacks.GET("", feedbackHandler.ListFeedbacks)
		feedbacks.GET("/:id", feedbackHandler.GetFeedback)
	}

	// Alert routes (Admin only)
	alerts := r.Group("/alerts")
	alerts.Use(authMiddleware.Authenticate()) // JWT authentication required
	alerts.Use(authMiddleware.RequireAdmin()) // Admin permission required
	{
		alerts.POST("", alertHandler.CreateAlert)
		alerts.GET("", alertHandler.ListAlerts)
		alerts.GET("/:id", alertHandler.GetAlert)
		alerts.PUT("/:id", alertHandler.UpdateAlert)
		alerts.DELETE("/:id", alertHandler.DeleteAlert)
	}

	// Operation log routes (Admin only)
	logs := r.Group("/logs")
	logs.Use(authMiddleware.Authenticate()) // JWT authentication required
	logs.Use(authMiddleware.RequireAdmin()) // Admin permission required
	{
		logs.GET("", operationLogHandler.ListOperationLogs)
		logs.GET("/:id", operationLogHandler.GetOperationLog)
	}
}
