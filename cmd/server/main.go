// Main entry point for the metrics API server.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/roryirvine/vibe-personal-dashboard-backend/internal/api"
	"github.com/roryirvine/vibe-personal-dashboard-backend/internal/api/handlers"
	"github.com/roryirvine/vibe-personal-dashboard-backend/internal/config"
	"github.com/roryirvine/vibe-personal-dashboard-backend/internal/repository"
	"github.com/roryirvine/vibe-personal-dashboard-backend/internal/service"
)

func main() {
	// Setup logging first so all startup messages are logged
	logger := setupLogging()
	logger.Info("Starting metrics API server")

	// Load environment and configuration
	port, dbPath := loadEnvironment(logger)
	metrics, err := config.LoadConfig("./config/metrics.toml")
	if err != nil {
		logger.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Initialize repository (database)
	repo, err := repository.NewSQLiteRepository(dbPath)
	if err != nil {
		logger.Error("Failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer repo.Close()

	// Wire up dependencies: repository -> service -> handlers -> router
	svc := service.NewMetricService(repo, metrics, logger)
	h := handlers.NewMetricsHandler(svc, logger)
	router := api.NewRouter(h, logger)

	// Setup HTTP server
	srv := &http.Server{
		Addr:           fmt.Sprintf(":%d", port),
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	// Start server in a goroutine
	go func() {
		logger.Info("HTTP server listening", "address", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server error", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigChan
	logger.Info("Received signal, shutting down", "signal", sig.String())

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Error during server shutdown", "error", err)
		os.Exit(1)
	}

	logger.Info("Server stopped gracefully")
}

// setupLogging configures slog with JSON output format.
func setupLogging() *slog.Logger {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	handler := slog.NewJSONHandler(os.Stdout, opts)
	logger := slog.New(handler)
	slog.SetDefault(logger)
	return logger
}

// loadEnvironment reads PORT and DB_PATH from environment or .env file with defaults.
func loadEnvironment(logger *slog.Logger) (port int, dbPath string) {
	// PORT
	portStr := os.Getenv("PORT")
	if portStr == "" {
		portStr = "8080"
		logger.Debug("PORT not set, using default", "port", portStr)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		logger.Error("Invalid PORT value", "value", portStr, "error", err)
		os.Exit(1)
	}

	// DB_PATH
	dbPath = os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./data.db"
		logger.Debug("DB_PATH not set, using default", "path", dbPath)
	}

	return port, dbPath
}
