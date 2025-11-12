package routes

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/manish-npx/go-echo-pg/internal/config"
	"github.com/manish-npx/go-echo-pg/internal/handler"
	customMiddleware "github.com/manish-npx/go-echo-pg/internal/middleware"
	"go.uber.org/zap"
)

type Routes struct {
	cfg         *config.Config
	authHandler *handler.AuthHandler
	logger      *zap.Logger
}

func NewRoutes(cfg *config.Config, authHandler *handler.AuthHandler, logger *zap.Logger) *Routes {
	return &Routes{
		cfg:         cfg,
		authHandler: authHandler,
		logger:      logger,
	}
}

func (r *Routes) RegisterRoutes(e *echo.Echo) {
	// Global middleware
	e.Use(middleware.RequestID())
	e.Use(customMiddleware.Logger(r.logger))
	e.Use(customMiddleware.CORS(r.cfg))
	e.Use(middleware.Secure())
	e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(100)))

	// Health checks (no auth)
	e.GET("/health", r.authHandler.Health)
	e.GET("/ready", r.authHandler.Ready)

	// Auth routes (public)
	auth := e.Group("/auth")
	{
		auth.POST("/register", r.authHandler.Register)
		auth.POST("/login", r.authHandler.Login)
	}

	// API v1 routes (protected)
	apiV1 := e.Group("/api/v1")
	apiV1.Use(customMiddleware.AuthMiddleware(r.cfg))
	{
		// User routes
		users := apiV1.Group("/users")
		{
			users.GET("/profile", r.authHandler.GetProfile)
			users.PUT("/profile", r.authHandler.UpdateProfile)
			users.POST("/change-password", r.authHandler.ChangePassword)
		}
	}

	// Not found handler
	e.Any("*", func(c echo.Context) error {
		return echo.ErrNotFound
	})
}
