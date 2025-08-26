// @title LIFE Game Server API
// @version 0.1.0
// @description A multiplayer FPS game server with OAuth authentication and N:1 account linking
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://github.com/homveloper/life
// @contact.email support@lifegame.com

// @license.name MIT
// @license.url https://github.com/homveloper/life/blob/main/LICENSE

// @host localhost:8082
// @BasePath /
// @schemes http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	_ "github.com/danghamo/life/docs/api"
	"github.com/danghamo/life/internal/api"
	"github.com/danghamo/life/pkg/config"
	"github.com/danghamo/life/pkg/redisx"
)

func main() {
	// Initialize configuration and logger
	cfg, log, err := config.Initialize()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize application: %v\n", err)
		os.Exit(1)
	}

	// Ensure logger is flushed on exit
	defer func() {
		_ = log.Sync()
	}()

	log.Info("Starting LIFE Game Server",
		zap.String("version", "0.1.0"),
		zap.String("environment", cfg.Server.Environment),
	)

	// Initialize Redis client (URL-based)
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379/0" // Default for development
	}

	redisClient, err := redisx.NewClient(redisURL, log)
	if err != nil {
		log.Fatal("Failed to initialize Redis client", zap.Error(err))
	}
	defer redisClient.Close()

	// Create API server
	serverConfig := api.ServerConfig{
		Port:         cfg.Server.Port,
		Host:         cfg.Server.Host,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	apiServer := api.NewServer(serverConfig, log, redisClient)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		log.Info("Shutting down server...")
		cancel()
	}()

	// Start server
	if err := apiServer.Start(ctx); err != nil {
		log.Error("Server error", zap.Error(err))
		os.Exit(1)
	}

	log.Info("Server gracefully stopped")
}
