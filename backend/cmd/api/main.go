package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/supercakecrumb/tg-analysis-tool/internal/api"
	"github.com/supercakecrumb/tg-analysis-tool/internal/config"
	"github.com/supercakecrumb/tg-analysis-tool/internal/infrastructure/postgres"
	"github.com/supercakecrumb/tg-analysis-tool/internal/infrastructure/redis"
)

func main() {
	// Load .env file in development
	if err := godotenv.Load(); err != nil {
		// It's okay if .env doesn't exist in production
		log.Debug().Err(err).Msg("No .env file found")
	}

	// Initialize logger
	initLogger()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Create application context
	ctx := context.Background()

	// Initialize PostgreSQL
	pgDB, err := postgres.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to PostgreSQL")
	}
	defer pgDB.Close()

	// Initialize Redis
	redisClient, err := redis.New(ctx, cfg.RedisURL)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to Redis")
	}
	defer redisClient.Close()

	// Create HTTP server
	server := api.NewServer(cfg, pgDB, redisClient)

	// Start server
	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      server,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Info().Msgf("Starting server on port %d", cfg.Port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("Server exited")
}

func initLogger() {
	// Configure zerolog
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// Set log level from environment
	logLevel := os.Getenv("LOG_LEVEL")
	switch logLevel {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	// Pretty logging for development
	if os.Getenv("ENV") != "production" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}
}
