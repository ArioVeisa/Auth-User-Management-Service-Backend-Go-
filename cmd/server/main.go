package main

import (
	"fmt"
	"log"

	"github.com/auth-service/internal/config"
	"github.com/auth-service/internal/database"
	"github.com/auth-service/internal/handlers"
	"github.com/auth-service/internal/middleware"
	"github.com/auth-service/internal/repository"
	"github.com/auth-service/internal/services"
	"github.com/auth-service/internal/utils"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := database.NewPostgresDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	redisClient, err := database.NewRedisClient(cfg.RedisURL)
	if err != nil {
		log.Printf("Warning: Redis connection failed: %v. Rate limiting will be disabled.", err)
		redisClient = nil
	}

	userRepo := repository.NewUserRepository(db)
	tokenRepo := repository.NewTokenRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	auditRepo := repository.NewAuditRepository(db)

	emailService := services.NewEmailService(cfg)
	auditService := services.NewAuditService(auditRepo)
	authService := services.NewAuthService(cfg, userRepo, tokenRepo, roleRepo, emailService, auditService)
	userService := services.NewUserService(userRepo, roleRepo, auditService)
	roleService := services.NewRoleService(roleRepo)

	jwtManager := utils.NewJWTManager(cfg.JWTSigningKey, cfg.AccessTokenExpiry)
	authMiddleware := middleware.NewAuthMiddleware(jwtManager)
	rateLimiter := middleware.NewRateLimiter(redisClient, cfg.RateLimitRequests, cfg.RateLimitWindow)

	authHandler := handlers.NewAuthHandler(authService)
	userHandler := handlers.NewUserHandler(userService)
	roleHandler := handlers.NewRoleHandler(roleService)
	auditHandler := handlers.NewAuditHandler(auditService)
	healthHandler := handlers.NewHealthHandler(db, redisClient)

	e := echo.New()
	e.HideBanner = true

	e.Use(echomiddleware.Logger())
	e.Use(echomiddleware.Recover())
	e.Use(echomiddleware.CORSWithConfig(echomiddleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET, echo.POST, echo.PUT, echo.DELETE, echo.OPTIONS},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
	}))
	e.Use(echomiddleware.Secure())

	e.GET("/healthz", healthHandler.Healthz)
	e.GET("/ready", healthHandler.Ready)

	api := e.Group("/api/v1")

	auth := api.Group("/auth")
	auth.POST("/register", authHandler.Register, rateLimiter.LimitByEndpoint("register"))
	auth.POST("/verify-email", authHandler.VerifyEmail)
	auth.POST("/login", authHandler.Login, rateLimiter.LimitByEndpoint("login"))
	auth.POST("/refresh", authHandler.Refresh)
	auth.POST("/logout", authHandler.Logout, authMiddleware.Authenticate)
	auth.POST("/forgot-password", authHandler.ForgotPassword, rateLimiter.LimitByEndpoint("forgot-password"))
	auth.POST("/reset-password", authHandler.ResetPassword)

	users := api.Group("/users")
	users.Use(authMiddleware.Authenticate)
	users.GET("/me", userHandler.GetCurrentUser)
	users.PUT("/me/password", userHandler.ChangePassword)
	users.GET("/:id", userHandler.GetUser, authMiddleware.RequireRoles("admin", "auditor"))
	users.GET("", userHandler.ListUsers, authMiddleware.RequireRoles("admin"))
	users.POST("", userHandler.CreateUser, authMiddleware.RequireRoles("admin"))
	users.PUT("/:id", userHandler.UpdateUser, authMiddleware.RequireRoles("admin"))
	users.DELETE("/:id", userHandler.DeleteUser, authMiddleware.RequireRoles("admin"))
	users.POST("/:id/roles", userHandler.AssignRole, authMiddleware.RequireRoles("admin"))
	users.DELETE("/:id/roles/:role", userHandler.UnassignRole, authMiddleware.RequireRoles("admin"))

	roles := api.Group("/roles")
	roles.Use(authMiddleware.Authenticate)
	roles.GET("", roleHandler.ListRoles)
	roles.GET("/:id", roleHandler.GetRole)
	roles.POST("", roleHandler.CreateRole, authMiddleware.RequireRoles("admin"))
	roles.PUT("/:id", roleHandler.UpdateRole, authMiddleware.RequireRoles("admin"))
	roles.DELETE("/:id", roleHandler.DeleteRole, authMiddleware.RequireRoles("admin"))

	audit := api.Group("/audit")
	audit.Use(authMiddleware.Authenticate)
	audit.Use(authMiddleware.RequireRoles("admin", "auditor"))
	audit.GET("", auditHandler.ListAuditLogs)

	fmt.Printf("Server starting on port %s\n", cfg.Port)
	if err := e.Start(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
