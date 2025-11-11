package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/manish-npx/go-echo-pg/internal/config"
)

func main() {

	// 🧩 Load config
	cfg := config.MustLoad()

	// 🧩 Initialize Echo
	e := echo.New()

	// 🧩 Example route (replace later with real routes)
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "🚀 Go Echo PG Server is running!")
	})

	// 🧩 Setup HTTP server
	server := &http.Server{
		Addr:    cfg.HTTPServer.Addr,
		Handler: e,
	}

	slog.Info("✅ Main file started")

	// 🧩 Choose database based on config

	// Channel for graceful shutdown
	// 🧩 Graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Start server in background
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Failed to start server: %v", err)
		}
	}()

	<-done // Block until shutdown signal

	slog.Info("📴 Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("❌ Failed to gracefully shutdown server", slog.String("error", err.Error()))
	} else {
		slog.Info("✅ Server shutdown successfully")
	}
}
