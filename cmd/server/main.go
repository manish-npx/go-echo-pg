package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/manish-npx/go-echo-pg/internal/config"
	"github.com/manish-npx/go-echo-pg/internal/database"
	"github.com/manish-npx/go-echo-pg/internal/handler"
	"github.com/manish-npx/go-echo-pg/internal/repository"
	"github.com/manish-npx/go-echo-pg/internal/routes"
	"github.com/manish-npx/go-echo-pg/internal/service"
	"github.com/manish-npx/go-echo-pg/internal/utils"
	"go.uber.org/zap"
)

func main() {
	// Parse command line flags
	var configPath string
	flag.StringVar(&configPath, "config", "", "Path to config file (default: ./config.yaml or ./config/config.yaml)")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ Error loading config: %v", err)
	}

	// Initialize logger
	logger, err := createLogger(cfg)
	if err != nil {
		log.Fatalf("❌ Error creating logger: %v", err)
	}
	defer logger.Sync()

	logger.Info("🚀 Starting application",
		zap.String("environment", cfg.Env),
		zap.String("config", configPath),
	)

	// Initialize database
	db, err := database.NewDB(cfg, logger)
	if err != nil {
		logger.Fatal("❌ Error connecting to database", zap.Error(err))
	}
	defer db.Close()

	// Create application
	app, err := NewApp(cfg, db, logger)
	if err != nil {
		logger.Fatal("❌ Error creating application", zap.Error(err))
	}

	// Start server
	if err := app.Start(); err != nil {
		logger.Fatal("❌ Error starting application", zap.Error(err))
	}
}

type App struct {
	cfg    *config.Config
	db     *database.DB
	logger *zap.Logger
	echo   *echo.Echo
}

func NewApp(cfg *config.Config, db *database.DB, logger *zap.Logger) (*App, error) {
	e := echo.New()

	// Configure Echo
	e.HideBanner = true
	e.HidePort = true

	// Custom error handler
	e.HTTPErrorHandler = customHTTPErrorHandler

	// Validator
	e.Validator = utils.NewValidator()

	// Initialize layers
	userRepo := repository.NewUserRepository(db, logger)
	authService := service.NewAuthService(userRepo, cfg, logger)
	authHandler := handler.NewAuthHandler(authService, logger)

	// Register routes
	routes := routes.NewRoutes(cfg, authHandler, logger)
	routes.RegisterRoutes(e)

	return &App{
		cfg:    cfg,
		db:     db,
		logger: logger,
		echo:   e,
	}, nil
}

func (a *App) Start() error {
	// Create server with timeouts
	server := &http.Server{
		Addr:         a.cfg.Server.Address,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		a.logger.Info("🚀 Starting server",
			zap.String("address", a.cfg.Server.Address),
			zap.String("environment", a.cfg.Env),
		)

		if err := a.echo.StartServer(server); err != nil && err != http.ErrServerClosed {
			a.logger.Fatal("❌ Failed to start server", zap.Error(err))
		}
	}()

	return a.waitForShutdown(server)
}

func (a *App) waitForShutdown(server *http.Server) error {
	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	a.logger.Info("⏳ Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := a.echo.Shutdown(ctx); err != nil {
		return err
	}

	a.logger.Info("✅ Server stopped gracefully")
	return nil
}

func createLogger(cfg *config.Config) (*zap.Logger, error) {
	var logger *zap.Logger
	var err error

	if cfg.Env == "production" {
		logger, err = zap.NewProduction()
	} else {
		config := zap.NewDevelopmentConfig()

		if cfg.Logging.Format == "json" {
			config.Encoding = "json"
		} else {
			config.Encoding = "console"
		}

		// Set log level based on config
		switch cfg.Logging.Level {
		case "debug":
			config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
		case "info":
			config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
		case "warn":
			config.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
		case "error":
			config.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
		default:
			config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
		}

		logger, err = config.Build()
	}

	return logger, err
}

func customHTTPErrorHandler(err error, c echo.Context) {
	code := http.StatusInternalServerError
	message := "Internal Server Error"

	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		if msg, ok := he.Message.(string); ok {
			message = msg
		}
	}

	// Send JSON error response
	if !c.Response().Committed {
		if c.Request().Method == http.MethodHead {
			err = c.NoContent(code)
		} else {
			err = c.JSON(code, map[string]any{
				"success": false,
				"error":   message,
			})
		}
		if err != nil {
			c.Logger().Error(err)
		}
	}
}
