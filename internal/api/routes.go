package api

import (
	"mcs-backend/internal/config"
	"mcs-backend/internal/database"
	"mcs-backend/internal/handlers"
	"mcs-backend/internal/middleware"
	"mcs-backend/internal/services"

	"github.com/gin-gonic/gin"
)

// SetupRoutes 设置所有路由
func SetupRoutes(cfg *config.Config) *gin.Engine {
	router := gin.New()

	// 中间件
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(CORSMiddleware())

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"message": "MCS Backend is running",
		})
	})

	// API版本分组
	v1 := router.Group("/api/v1")
	{
		// 初始化处理器
		authHandler := handlers.NewAuthHandler(cfg)

		// 认证相关路由
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
			auth.GET("/validate-invite", authHandler.ValidateInviteCode)
			auth.POST("/logout", middleware.AuthMiddleware(cfg), authHandler.Logout)
			auth.GET("/profile", middleware.AuthMiddleware(cfg), authHandler.GetProfile)
		}

		// 用户管理路由
		userHandler := handlers.NewUserHandler()
		users := v1.Group("/users")
		{
			// 公开接口
			users.POST("/change-password", middleware.AuthMiddleware(cfg), userHandler.ChangePassword)
			
			// 管理员接口
			users.POST("/", middleware.AuthMiddleware(cfg), middleware.RequireAdmin(), userHandler.CreateUser)
			users.GET("/", middleware.AuthMiddleware(cfg), middleware.RequireAdmin(), userHandler.GetUserList)
			users.GET("/stats", middleware.AuthMiddleware(cfg), middleware.RequireAdmin(), userHandler.GetUserStats)
			users.GET("/:id", middleware.AuthMiddleware(cfg), middleware.RequireAdmin(), userHandler.GetUser)
			users.PUT("/:id", middleware.AuthMiddleware(cfg), middleware.RequireAdmin(), userHandler.UpdateUser)
			users.DELETE("/:id", middleware.AuthMiddleware(cfg), middleware.RequireSuperAdmin(), userHandler.DeleteUser)
		}

		// 用户组管理路由
		groupHandler := handlers.NewUserGroupHandler()
		groups := v1.Group("/user-groups")
		{
			// 认证用户接口
			groups.GET("/my-groups", middleware.AuthMiddleware(cfg), groupHandler.GetUserGroups)
			
			// 管理员接口
			groups.POST("/", middleware.AuthMiddleware(cfg), middleware.RequireAdmin(), groupHandler.CreateGroup)
			groups.GET("/", middleware.AuthMiddleware(cfg), middleware.RequireAdmin(), groupHandler.GetGroupList)
			groups.GET("/:id", middleware.AuthMiddleware(cfg), middleware.RequireAdmin(), groupHandler.GetGroup)
			groups.PUT("/:id", middleware.AuthMiddleware(cfg), middleware.RequireAdmin(), groupHandler.UpdateGroup)
			groups.DELETE("/:id", middleware.AuthMiddleware(cfg), middleware.RequireAdmin(), groupHandler.DeleteGroup)
			
			// 成员管理
			groups.GET("/:id/members", middleware.AuthMiddleware(cfg), middleware.RequireAdmin(), groupHandler.GetGroupMembers)
			groups.POST("/:id/members", middleware.AuthMiddleware(cfg), middleware.RequireAdmin(), groupHandler.AddMember)
			groups.PUT("/:id/members/:user_id", middleware.AuthMiddleware(cfg), middleware.RequireAdmin(), groupHandler.UpdateMember)
			groups.DELETE("/:id/members/:user_id", middleware.AuthMiddleware(cfg), middleware.RequireAdmin(), groupHandler.RemoveMember)
		}

		// 文件上传路由
		uploadHandler := handlers.NewUploadHandler(cfg)
		upload := v1.Group("/upload")
		upload.Use(middleware.AuthMiddleware(cfg))
		{
			upload.POST("/init", uploadHandler.InitUpload)
			upload.POST("/chunk", uploadHandler.UploadChunk)
			upload.POST("/complete/:upload_id", uploadHandler.CompleteUpload)
			upload.GET("/progress/:upload_id", uploadHandler.GetUploadProgress)
			upload.DELETE("/cancel/:upload_id", uploadHandler.CancelUpload)
		}

		// 文件管理路由
		fileService := services.NewFileService(cfg)
		fileHandler := handlers.NewFileHandler(fileService)
		files := v1.Group("/files")
		files.Use(middleware.AuthMiddleware(cfg))
		{
			files.POST("/folders", fileHandler.CreateFolder)
			files.GET("/list", fileHandler.GetFileList)
			files.GET("/:id", fileHandler.GetFile)
			files.PUT("/:id", fileHandler.UpdateFile)
			files.DELETE("/:id", fileHandler.DeleteFile)
			files.PUT("/folders/:id", fileHandler.UpdateFolder)
			files.DELETE("/folders/:id", fileHandler.DeleteFolder)
			files.GET("/search", fileHandler.SearchFiles)
			files.GET("/:id/versions", fileHandler.GetFileVersions)
			files.GET("/:id/download", fileHandler.DownloadFile)
		}

		// 工作流管理路由
		workflowService := services.NewWorkflowService(cfg)
		workflowHandler := handlers.NewWorkflowHandler(workflowService)
		workflows := v1.Group("/workflows")
		workflows.Use(middleware.AuthMiddleware(cfg))
		{
			workflows.POST("/", workflowHandler.CreateWorkflow)
			workflows.GET("/", workflowHandler.GetWorkflowList)
			workflows.GET("/:id", workflowHandler.GetWorkflow)
			workflows.PUT("/:id", workflowHandler.UpdateWorkflow)
			workflows.DELETE("/:id", workflowHandler.DeleteWorkflow)
			workflows.POST("/:id/members", workflowHandler.AddMember)
			workflows.GET("/:id/members", workflowHandler.GetMembers)
			workflows.DELETE("/:id/members/:member_id", workflowHandler.RemoveMember)
			workflows.PUT("/:id/members/:member_id/role", workflowHandler.UpdateMemberRole)
		}

		// 任务管理路由
		taskService := services.NewTaskService(cfg)
		taskHandler := handlers.NewTaskHandler(taskService)
		tasks := v1.Group("/tasks")
		tasks.Use(middleware.AuthMiddleware(cfg))
		{
			tasks.POST("/", taskHandler.CreateTask)
			tasks.GET("/", taskHandler.GetTaskList)
			tasks.GET("/:id", taskHandler.GetTask)
			tasks.PUT("/:id", taskHandler.UpdateTask)
			tasks.DELETE("/:id", taskHandler.DeleteTask)
			tasks.PUT("/:id/status", taskHandler.ChangeTaskStatus)
			tasks.POST("/:id/staging", taskHandler.AddToStagingArea)
			tasks.GET("/:id/staging", taskHandler.GetStagingArea)
			tasks.POST("/:id/staging/submit", taskHandler.SubmitStagingArea)
			tasks.DELETE("/:id/staging/clear", taskHandler.ClearStagingArea)
		}

		// 通知管理路由
		notificationService := services.NewNotificationService(cfg)
		notificationHandler := handlers.NewNotificationHandler(notificationService)
		notifications := v1.Group("/notifications")
		notifications.Use(middleware.AuthMiddleware(cfg))
		{
			notifications.POST("/", notificationHandler.CreateNotification)
			notifications.GET("/", notificationHandler.GetNotificationList)
			notifications.PUT("/:id/read", notificationHandler.MarkAsRead)
			notifications.PUT("/read-all", notificationHandler.MarkAllAsRead)
			notifications.DELETE("/:id", notificationHandler.DeleteNotification)
			notifications.GET("/stats", notificationHandler.GetNotificationStats)
			notifications.GET("/settings", notificationHandler.GetNotificationSetting)
			notifications.PUT("/settings", notificationHandler.UpdateNotificationSetting)
			notifications.POST("/system", middleware.RequireAdmin(), notificationHandler.SendSystemNotification)
		}

		// 统计报表路由
		statisticsService := services.NewStatisticsService(database.GetDB())
		statisticsHandler := handlers.NewStatisticsHandler(statisticsService)
		stats := v1.Group("/stats")
		stats.Use(middleware.AuthMiddleware(cfg))
		{
			stats.GET("/operation-logs", statisticsHandler.GetOperationLogs)
			stats.GET("/storage", statisticsHandler.GetStorageStats)
			stats.GET("/user-activity", statisticsHandler.GetUserActivityStats)
			stats.GET("/overview", statisticsHandler.GetSystemOverview)
			stats.GET("/operations", statisticsHandler.GetOperationStats)
			stats.GET("/top-users", statisticsHandler.GetTopActiveUsers)
			
			// 管理员接口
			stats.PUT("/storage", middleware.RequireAdmin(), statisticsHandler.UpdateStorageStats)
			stats.PUT("/user-activity", middleware.RequireAdmin(), statisticsHandler.UpdateUserActivityStats)
		}

		// 下载管理路由
		downloadService := services.NewDownloadService(database.GetDB(), cfg.File.UploadPath, cfg.File.DownloadPath, cfg.Server.BaseURL)
		downloadHandler := handlers.NewDownloadHandler(downloadService)
		download := v1.Group("/download")
		download.Use(middleware.AuthMiddleware(cfg))
		{
			download.POST("/batch", downloadHandler.CreateBatchDownload)
			download.GET("/tasks", downloadHandler.GetDownloadTasks)
			download.GET("/tasks/:id", downloadHandler.GetDownloadTask)
			download.GET("/zip/:filename", downloadHandler.DownloadZipFile)
			download.GET("/file/:id", downloadHandler.DownloadSingleFile)
			download.GET("/stats", downloadHandler.GetDownloadStats)
			
			// 管理员接口
			download.GET("/stats/global", middleware.RequireAdmin(), downloadHandler.GetGlobalDownloadStats)
		}
	}

	return router
}

// CORSMiddleware 跨域中间件
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}