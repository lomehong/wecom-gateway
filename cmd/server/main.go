package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"wecom-gateway/internal/admin"
	"wecom-gateway/internal/apikey"
	"wecom-gateway/internal/audit"
	"wecom-gateway/internal/auth"
	"wecom-gateway/internal/config"
	"wecom-gateway/internal/contact"
	"wecom-gateway/internal/crypto"
	"wecom-gateway/internal/document"
	"wecom-gateway/internal/grpcserver"
	"wecom-gateway/internal/meeting"
	"wecom-gateway/internal/message"
	"wecom-gateway/internal/mcp"
	"wecom-gateway/internal/openapi"
	"wecom-gateway/internal/ratelimit"
	"wecom-gateway/internal/schedule"
	"wecom-gateway/internal/store"
	"wecom-gateway/internal/todo"
	"wecom-gateway/internal/wecom"
)

func main() {
	// Load configuration
	configPath := "config.yaml"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	cfg, err := config.LoadWithEnv(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Set Gin mode
	gin.SetMode(cfg.Server.Mode)

	// Initialize encryption key
	var encKey []byte
	if key := os.Getenv("WECOM_MASTER_KEY"); key != "" {
		// Use provided key
		encKey = []byte(key)
	} else {
		// Generate or load key (for production, this should be properly secured)
		encKey = crypto.GenerateKeyFromPassphrase("default-change-me")
	}

	// Initialize JWT secret
	var jwtSecret string
	if secret := os.Getenv("WECOM_JWT_SECRET"); secret != "" {
		jwtSecret = secret
	} else {
		// Generate a default JWT secret (for production, this should be properly secured)
		jwtSecret = "wecom-gateway-jwt-secret-change-me-in-production"
	}

	// Initialize database
	db, err := store.New(&store.Config{
		Driver: cfg.Database.Driver,
		DSN:    cfg.Database.DSN,
	})
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initialize services
	wecomClient := wecom.NewClientWithDB(cfg, db, encKey)
	authenticator := auth.NewAPIKeyAuthenticator(db)
	rateLimiter := ratelimit.NewRateLimiter()
	auditLogger := audit.NewLogger(db)
	auditQuer := audit.NewQuerier(db)
	authManager := admin.NewAuthManager(db, jwtSecret)

	apiKeySvc := apikey.NewService(db, cfg)
	scheduleSvc := schedule.NewService(wecomClient)
	meetingSvc := meeting.NewService(wecomClient)
	messageSvc := message.NewService(wecomClient)
	contactSvc := contact.NewService(wecomClient)
	todoSvc := todo.NewService(wecomClient)
	documentClient := document.NewClient(wecomClient)
	documentSvc := document.NewService(documentClient)
	adminSvc := admin.NewService(db, cfg, apiKeySvc, auditLogger, encKey)

	// Initialize handlers
	scheduleHandler := schedule.NewHandler(scheduleSvc)
	meetingHandler := meeting.NewHandler(meetingSvc)
	messageHandler := message.NewHandler(messageSvc)
	contactHandler := contact.NewHandler(contactSvc)
	todoHandler := todo.NewHandler(todoSvc)
	documentHandler := document.NewHandler(documentSvc)
	adminHandler := admin.NewHandler(adminSvc, apiKeySvc, auditQuer, authManager)

	// Initialize gRPC server
	grpcServer := grpcserver.NewServer(
		cfg,
		wecomClient,
		scheduleSvc,
		meetingSvc,
		messageSvc,
		apiKeySvc,
		auditLogger,
		authenticator,
	)

	// Initialize system if needed
	if err := adminSvc.InitializeSystem(context.Background()); err != nil {
		log.Printf("Warning: Failed to initialize system: %v", err)
	}

	// Create default admin user if not exists
	if err := createDefaultAdminUser(db); err != nil {
		log.Printf("Warning: Failed to create default admin user: %v", err)
	}

	// Create Gin router
	router := gin.Default()

	// Apply global middleware
	router.Use(corsMiddleware())
	router.Use(rateLimitMiddleware(rateLimiter, cfg.Auth.RateLimit))
	router.Use(auditMiddleware(auditLogger))
	router.Use(recoveryMiddleware())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		checks := gin.H{}

		// Check database connection
		dbStatus := "ok"
		if err := db.Ping(c.Request.Context()); err != nil {
			dbStatus = "unhealthy: " + err.Error()
		}
		checks["database"] = dbStatus

		overallStatus := "ok"
		if dbStatus != "ok" {
			overallStatus = "unhealthy"
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":  overallStatus,
				"version": "1.0.0",
				"checks":  checks,
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":  overallStatus,
			"version": "1.0.0",
			"checks":  checks,
		})
	})

	// Serve management UI
	if cfg.UI.Enabled {
		router.Static("/ui", "./ui")
		router.GET("/", func(c *gin.Context) {
			c.Redirect(http.StatusMovedPermanently, "/ui/")
		})
	}

	// API v1 routes
	v1 := router.Group("/v1")
	{
		// Schedule routes
		scheduleGroup := v1.Group("/schedules")
		scheduleGroup.Use(auth.GinMiddleware(authenticator))
		{
			scheduleGroup.POST("", auth.RequirePermission("calendar:write"), scheduleHandler.CreateSchedule)
			scheduleGroup.GET("", auth.RequirePermission("calendar:read"), scheduleHandler.GetSchedules)
			scheduleGroup.GET("/:id", auth.RequirePermission("calendar:read"), scheduleHandler.GetScheduleByID)
			scheduleGroup.PATCH("/:id", auth.RequirePermission("calendar:write"), scheduleHandler.UpdateSchedule)
			scheduleGroup.DELETE("/:id", auth.RequirePermission("calendar:write"), scheduleHandler.DeleteSchedule)
			scheduleGroup.POST("/availability", auth.RequirePermission("calendar:read"), scheduleHandler.CheckAvailability)
		}

		// Meeting room routes
		meetingGroup := v1.Group("/meeting-rooms")
		meetingGroup.Use(auth.GinMiddleware(authenticator))
		{
			meetingGroup.GET("", auth.RequirePermission("meetingroom:read"), meetingHandler.ListMeetingRooms)
			meetingGroup.GET("/:id/availability", auth.RequirePermission("meetingroom:read"), meetingHandler.GetRoomAvailability)
			meetingGroup.POST("/:id/bookings", auth.RequirePermission("meetingroom:write"), meetingHandler.BookMeetingRoom)
		}

		// Meeting appointment routes
		meetingApptGroup := v1.Group("/meetings")
		meetingApptGroup.Use(auth.GinMiddleware(authenticator))
		{
			meetingApptGroup.POST("", auth.RequirePermission("meeting:write"), meetingHandler.CreateMeeting)
			meetingApptGroup.DELETE("/:id", auth.RequirePermission("meeting:write"), meetingHandler.CancelMeeting)
			meetingApptGroup.PUT("/:id/invitees", auth.RequirePermission("meeting:write"), meetingHandler.UpdateInvitees)
			meetingApptGroup.GET("", auth.RequirePermission("meeting:read"), meetingHandler.ListMeetings)
			meetingApptGroup.GET("/:id", auth.RequirePermission("meeting:read"), meetingHandler.GetMeetingInfo)
		}

		// Message routes
		messageGroup := v1.Group("/messages")
		messageGroup.Use(auth.GinMiddleware(authenticator))
		{
			messageGroup.POST("/text", auth.RequirePermission("message:send"), messageHandler.SendText)
			messageGroup.POST("/markdown", auth.RequirePermission("message:send"), messageHandler.SendMarkdown)
			messageGroup.POST("/image", auth.RequirePermission("message:send"), messageHandler.SendImage)
			messageGroup.POST("/file", auth.RequirePermission("message:send"), messageHandler.SendFile)
			messageGroup.POST("/card", auth.RequirePermission("message:send"), messageHandler.SendCard)
			messageGroup.GET("/chats", auth.RequirePermission("message:read"), messageHandler.GetChatList)
			messageGroup.GET("/chats/:chatid/messages", auth.RequirePermission("message:read"), messageHandler.GetChatMessages)
			messageGroup.GET("/media/:mediaid", auth.RequirePermission("message:read"), messageHandler.DownloadMedia)
		}

		// Contact routes
		contactGroup := v1.Group("/contacts")
		contactGroup.Use(auth.GinMiddleware(authenticator))
		{
			contactGroup.GET("/users", auth.RequirePermission("contact:read"), contactHandler.GetUserList)
			contactGroup.GET("/users/search", auth.RequirePermission("contact:read"), contactHandler.SearchUser)
		}

		// Todo routes
		todoGroup := v1.Group("/todos")
		todoGroup.Use(auth.GinMiddleware(authenticator))
		{
			todoGroup.GET("", auth.RequirePermission("todo:read"), todoHandler.GetTodoList)
			todoGroup.GET("/:id", auth.RequirePermission("todo:read"), todoHandler.GetTodoDetail)
			todoGroup.POST("", auth.RequirePermission("todo:write"), todoHandler.CreateTodo)
			todoGroup.PUT("/:id", auth.RequirePermission("todo:write"), todoHandler.UpdateTodo)
			todoGroup.DELETE("/:id", auth.RequirePermission("todo:write"), todoHandler.DeleteTodo)
			todoGroup.PUT("/:id/status", auth.RequirePermission("todo:write"), todoHandler.ChangeUserStatus)
		}

		// Document management routes
		docGroup := v1.Group("/docs")
		docGroup.Use(auth.GinMiddleware(authenticator))
		{
			docGroup.GET("", auth.RequirePermission("document:read"), documentHandler.ListDocuments)
			docGroup.POST("", auth.RequirePermission("document:write"), documentHandler.CreateDocument)
			docGroup.GET("/:docid", auth.RequirePermission("document:read"), documentHandler.GetDocument)
			docGroup.PUT("/:docid/rename", auth.RequirePermission("document:write"), documentHandler.RenameDocument)
			docGroup.DELETE("/:docid", auth.RequirePermission("document:write"), documentHandler.DeleteDocument)
			docGroup.POST("/:docid/share", auth.RequirePermission("document:write"), documentHandler.ShareDocument)
			docGroup.GET("/:docid/permissions", auth.RequirePermission("document:read"), documentHandler.GetPermissions)
			docGroup.PUT("/:docid/content", auth.RequirePermission("document:write"), documentHandler.EditContent)
			docGroup.GET("/:docid/data", auth.RequirePermission("document:read"), documentHandler.GetDocumentData)
			docGroup.POST("/:docid/images", auth.RequirePermission("document:write"), documentHandler.UploadImage)

			// Sheet operations
			docGroup.POST("/sheets/:docid/content", auth.RequirePermission("document:write"), documentHandler.EditSheetContent)
			docGroup.GET("/sheets/:docid/rows", auth.RequirePermission("document:read"), documentHandler.GetSheetRowCol)
			docGroup.GET("/sheets/:docid/data", auth.RequirePermission("document:read"), documentHandler.GetSheetData)

			// Space operations
			docGroup.POST("/spaces", auth.RequirePermission("document:write"), documentHandler.CreateSpace)
			docGroup.GET("/spaces/:spaceid", auth.RequirePermission("document:read"), documentHandler.GetSpaceInfo)
			docGroup.GET("/spaces/:spaceid/files", auth.RequirePermission("document:read"), documentHandler.GetSpaceFileList)
		}

		// Admin routes
		adminGroup := v1.Group("/admin")
		{
			// Login endpoint - no authentication required
			adminGroup.POST("/login", adminHandler.Login)

			// Authenticated admin routes - use JWT middleware
			authenticatedGroup := adminGroup.Group("")
			authenticatedGroup.Use(authManager.JWTMiddleware())
			{
				authenticatedGroup.POST("/change-password", adminHandler.ChangePassword)
				authenticatedGroup.GET("/api-keys", adminHandler.ListAPIKeys)
				authenticatedGroup.POST("/api-keys", adminHandler.CreateAPIKey)
				authenticatedGroup.PATCH("/api-keys/:id", adminHandler.UpdateAPIKey)
				authenticatedGroup.DELETE("/api-keys/:id", adminHandler.DeleteAPIKey)
				authenticatedGroup.GET("/audit-logs", adminHandler.QueryAuditLogs)
				authenticatedGroup.GET("/dashboard", adminHandler.GetDashboardStats)
				authenticatedGroup.POST("/initialize", adminHandler.InitializeSystem)

				// WeCom Corp management
				authenticatedGroup.GET("/corps", adminHandler.ListWeComCorps)
				authenticatedGroup.POST("/corps", adminHandler.CreateWeComCorp)
				authenticatedGroup.PUT("/corps/:id", adminHandler.UpdateWeComCorp)
				authenticatedGroup.DELETE("/corps/:id", adminHandler.DeleteWeComCorp)

				// WeCom App management (separate prefix to avoid Gin param conflict)
				authenticatedGroup.GET("/apps", adminHandler.ListWeComApps)
				authenticatedGroup.POST("/apps", adminHandler.CreateWeComApp)
				authenticatedGroup.PUT("/apps/:id", adminHandler.UpdateWeComApp)
				authenticatedGroup.DELETE("/apps/:id", adminHandler.DeleteWeComApp)

				// WeCom Bot management (AI Bot support)
				authenticatedGroup.POST("/corps/:corp_name/bots", adminHandler.CreateWeComBot)
			}
		}
	}

	// MCP endpoint
	mcpHandler := mcp.NewHandler(wecomClient, authenticator)
	router.POST("/mcp", mcpHandler.HandleRPC)

	// OpenAPI endpoint
	openapiHandler := openapi.NewHandler()
	router.GET("/openapi.json", openapiHandler.ServeJSON)
	router.GET("/docs", openapiHandler.ServeDocs)

	// Start HTTP server
	httpSrv := &http.Server{
		Addr:    cfg.Server.HTTPListen,
		Handler: router,
	}

	// Start gRPC server in background
	go func() {
		log.Printf("Starting gRPC server on %s", cfg.Server.GRPCListen)
		if err := grpcServer.Start(cfg.Server.GRPCListen); err != nil {
			log.Fatalf("gRPC server failed: %v", err)
		}
	}()

	// Graceful shutdown
	go func() {
		log.Printf("Starting HTTP server on %s", cfg.Server.HTTPListen)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down servers...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if err := httpSrv.Shutdown(ctx); err != nil {
		log.Fatalf("HTTP server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

// corsMiddleware applies CORS headers
func corsMiddleware() gin.HandlerFunc {
	// Read allowed origins from config or environment variable
	// Default: allow all origins in development mode
	allowedOrigin := os.Getenv("CORS_ALLOWED_ORIGIN")
	if allowedOrigin == "" {
		allowedOrigin = "*"
	}

	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", allowedOrigin)
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// rateLimitMiddleware applies rate limiting
func rateLimitMiddleware(limiter *ratelimit.RateLimiter, requestsPerMinute int) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract API key for rate limiting
		apiKey := c.GetHeader("Authorization")
		if apiKey == "" {
			c.Next()
			return
		}

		// Remove "Bearer " prefix
		if len(apiKey) > 7 && apiKey[:7] == "Bearer " {
			apiKey = apiKey[7:]
		}

		if !limiter.Allow(apiKey, requestsPerMinute) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"code":    42901,
				"message": "rate limit exceeded",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// auditMiddleware logs all requests
func auditMiddleware(logger *audit.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Extract auth context if available
		var apiKeyID, apiKeyName *string
		if authCtx, exists := c.Get("auth_context"); exists {
			if ac, ok := authCtx.(*auth.AuthContext); ok {
				apiKeyName = &ac.KeyName
				// You might want to store the ID as well
			}
		}

		// Process request
		c.Next()

		// Log after request completes
		duration := time.Since(start)
		clientIP := c.ClientIP()

		logger.Log(c.Request.Context(), &audit.LogEntry{
			Timestamp:  start,
			Protocol:   "http",
			APIKeyID:   apiKeyID,
			APIKeyName: apiKeyName,
			Method:     c.Request.Method,
			Path:       c.Request.URL.Path,
			Query:      &c.Request.URL.RawQuery,
			StatusCode: c.Writer.Status(),
			DurationMs: int(duration.Milliseconds()),
			ClientIP:   &clientIP,
		})
	}
}

// recoveryMiddleware recovers from panics
func recoveryMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		log.Printf("Panic recovered: %v", recovered)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    50001,
			"message": "internal server error",
		})
	})
}

// createDefaultAdminUser creates the default admin user if it doesn't exist
func createDefaultAdminUser(db store.Database) error {
	ctx := context.Background()

	// Check if admin user already exists
	_, err := db.GetAdminUserByUsername(ctx, "admin")
	if err == nil {
		// Admin user already exists
		return nil
	}

	// Create default admin user (password from env or default)
	defaultPassword := os.Getenv("ADMIN_DEFAULT_PASSWORD")
	if defaultPassword == "" {
		defaultPassword = "admin123" // MUST be changed on first login
	}
	hashedPassword, err := admin.HashPassword(defaultPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	adminUser := &store.AdminUser{
		ID:          "admin_default",
		Username:    "admin",
		Password:    hashedPassword,
		DisplayName: "系统管理员",
		IsActive:    true,
	}

	if err := db.CreateAdminUser(ctx, adminUser); err != nil {
		return fmt.Errorf("failed to create admin user: %w", err)
	}

	log.Println("Default admin user created (username: admin). IMPORTANT: Change the default password immediately!")
	return nil
}
