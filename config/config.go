package main

import (
	"ExeProcessManager/api"
	"ExeProcessManager/command"
	"ExeProcessManager/config"
	"ExeProcessManager/process"
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// 1. Load Configuration
	cfg, err := config.Load("config.json")
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	// 2. Setup Structured Logger
	logger := setupLogger(cfg.LogLevel)
	slog.SetDefault(logger)

	logger.Info("ExeProcessManager starting up...")
	logger.Info("Configuration loaded successfully")

	// 3. Setup Context for Graceful Shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Listen for interrupt signals (e.g., Ctrl+C)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		logger.Info("shutdown signal received", "signal", sig)
		cancel() // Trigger context cancellation
	}()

	// 4. Initialize Core Components with Dependencies
	// Note: We need a mechanism to load existing processes from disk on startup.
	// This is a placeholder for that logic.
	processManager := process.NewProcessManager(logger, cfg)
	if err := processManager.LoadProcessesFromDisk(); err != nil {
		logger.Error("failed to load existing processes", "error", err)
		// Decide if you want to exit or continue with an empty manager
	}

	// 5. Start API Server in a Goroutine
	processAPI := api.NewProcessAPI(processManager, logger, cfg)
	server := &http.Server{
		Addr:    cfg.ApiListenAddress,
		Handler: processAPI.Routes(),
	}

	go func() {
		logger.Info("starting API server", "address", cfg.ApiListenAddress)
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			logger.Error("API server crashed", "error", err)
		}
	}()

	// 6. Start the Command Line Interface (CLI)
	cli := command.NewCLI(processManager, logger)
	go cli.Start(ctx)

	// 7. Wait for context to be cancelled (shutdown signal)
	<-ctx.Done()

	// 8. Perform Graceful Shutdown
	logger.Info("shutting down gracefully...")

	// Shutdown the API server with a timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("API server shutdown failed", "error", err)
	} else {
		logger.Info("API server stopped")
	}

	// Add any other cleanup logic here (e.g., ensuring all processes are saved)
	logger.Info("ExeProcessManager has been shut down. Goodbye!")
}

// setupLogger initializes and returns a new slog.Logger based on the configured log level.
func setupLogger(level string) *slog.Logger {
	var logLevel slog.Level
	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: logLevel,
	}
	handler := slog.NewTextHandler(os.Stdout, opts)
	return slog.New(handler)
}