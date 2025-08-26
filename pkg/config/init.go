package config

import (
	"fmt"

	"github.com/danghamo/life/pkg/logger"
)

// Initialize loads configuration and sets up global logger
func Initialize() (*Config, *logger.Logger, error) {
	// Load configuration
	cfg, err := Load()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Create logger based on configuration
	loggerCfg := logger.Config{
		Level:       logger.ParseLevel(cfg.Log.Level),
		Environment: cfg.Log.Environment,
		Encoding:    cfg.Log.Encoding,
	}

	appLogger, err := logger.New(loggerCfg)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create logger: %w", err)
	}

	// Set global logger
	logger.SetGlobalLogger(appLogger)

	// Log successful initialization
	fields := map[string]interface{}{
		"environment":  cfg.Server.Environment,
		"server_port":  cfg.Server.Port,
		"redis_host":   cfg.Redis.Host,
		"redis_port":   cfg.Redis.Port,
		"log_level":    cfg.Log.Level,
		"log_encoding": cfg.Log.Encoding,
	}
	appLogger.WithFields(fields).Info("Configuration and logger initialized successfully")

	return cfg, appLogger, nil
}

// MustInitialize is like Initialize but panics on error
func MustInitialize() (*Config, *logger.Logger) {
	cfg, appLogger, err := Initialize()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize application: %v", err))
	}
	return cfg, appLogger
}
