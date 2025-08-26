package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	Server ServerConfig `mapstructure:"server"`
	Redis  RedisConfig  `mapstructure:"redis"`
	Asynq  AsynqConfig  `mapstructure:"asynq"`
	Game   GameConfig   `mapstructure:"game"`
	Auth   AuthConfig   `mapstructure:"auth"`
	CORS   CORSConfig   `mapstructure:"cors"`
	Log    LogConfig    `mapstructure:"log"`
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port            int    `mapstructure:"port"`
	Host            string `mapstructure:"host"`
	Environment     string `mapstructure:"environment"`
	MetricsEnabled  bool   `mapstructure:"metrics_enabled"`
	MetricsPort     int    `mapstructure:"metrics_port"`
	HealthCheckPath string `mapstructure:"health_check_path"`
}

// RedisConfig holds Redis-related configuration
type RedisConfig struct {
	Host         string             `mapstructure:"host"`
	Port         int                `mapstructure:"port"`
	Password     string             `mapstructure:"password"`
	DB           int                `mapstructure:"db"`
	MaxRetries   int                `mapstructure:"max_retries"`
	PoolSize     int                `mapstructure:"pool_size"`
	DialTimeout  time.Duration      `mapstructure:"dial_timeout"`
	ReadTimeout  time.Duration      `mapstructure:"read_timeout"`
	WriteTimeout time.Duration      `mapstructure:"write_timeout"`
	Streams      RedisStreamsConfig `mapstructure:"streams"`
}

// RedisStreamsConfig holds Redis Streams specific configuration
type RedisStreamsConfig struct {
	MaxLen        int64  `mapstructure:"max_len"`
	ConsumerGroup string `mapstructure:"consumer_group"`
}

// AsynqConfig holds Asynq task queue configuration
type AsynqConfig struct {
	RedisAddr      string `mapstructure:"redis_addr"`
	Concurrency    int    `mapstructure:"concurrency"`
	StrictPriority bool   `mapstructure:"strict_priority"`
}

// GameConfig holds game-specific configuration
type GameConfig struct {
	MapWidth            int     `mapstructure:"map_width"`
	MapHeight           int     `mapstructure:"map_height"`
	MaxAnimalsPerPlayer int     `mapstructure:"max_animals_per_player"`
	AnimalSpawnRate     float64 `mapstructure:"animal_spawn_rate"`
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	JWTSecret     string        `mapstructure:"jwt_secret"`
	JWTExpiration time.Duration `mapstructure:"jwt_expiration"`
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins []string `mapstructure:"allowed_origins"`
	AllowedMethods []string `mapstructure:"allowed_methods"`
	AllowedHeaders []string `mapstructure:"allowed_headers"`
}

// LogConfig holds logging configuration
type LogConfig struct {
	Level       string `mapstructure:"level"`
	Environment string `mapstructure:"environment"`
	Encoding    string `mapstructure:"encoding"`
}

// Load loads configuration from environment variables and config files
func Load() (*Config, error) {
	// Set default values
	setDefaults()

	// Setup Viper
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("/etc/life")

	// Enable environment variable reading
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Try to read config file (optional)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		// Config file not found, continue with env vars and defaults
	}

	// Unmarshal config
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate configuration
	if err := validateConfig(&cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

// setDefaults sets default configuration values
func setDefaults() {
	// Server defaults
	viper.SetDefault("server.port", 80)
	viper.SetDefault("server.host", "localhost")
	viper.SetDefault("server.environment", "development")
	viper.SetDefault("server.metrics_enabled", true)
	viper.SetDefault("server.metrics_port", 9090)
	viper.SetDefault("server.health_check_path", "/health")

	// Redis defaults
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("redis.max_retries", 3)
	viper.SetDefault("redis.pool_size", 10)
	viper.SetDefault("redis.dial_timeout", "5s")
	viper.SetDefault("redis.read_timeout", "3s")
	viper.SetDefault("redis.write_timeout", "3s")
	viper.SetDefault("redis.streams.max_len", 10000)
	viper.SetDefault("redis.streams.consumer_group", "life-game-server")

	// Asynq defaults
	viper.SetDefault("asynq.redis_addr", "localhost:6379")
	viper.SetDefault("asynq.concurrency", 10)
	viper.SetDefault("asynq.strict_priority", false)

	// Game defaults
	viper.SetDefault("game.map_width", 30)
	viper.SetDefault("game.map_height", 20)
	viper.SetDefault("game.max_animals_per_player", 6)
	viper.SetDefault("game.animal_spawn_rate", 0.1)

	// Auth defaults
	viper.SetDefault("auth.jwt_secret", "dev-jwt-secret-change-in-production")
	viper.SetDefault("auth.jwt_expiration", "24h")

	// CORS defaults
	viper.SetDefault("cors.allowed_origins", []string{"http://localhost:3000", "http://localhost:8080"})
	viper.SetDefault("cors.allowed_methods", []string{"GET", "POST", "OPTIONS"})
	viper.SetDefault("cors.allowed_headers", []string{"Content-Type", "Authorization"})

	// Log defaults
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.environment", "development")
	viper.SetDefault("log.encoding", "console")
}

// validateConfig validates the loaded configuration
func validateConfig(cfg *Config) error {
	// Validate server config
	if cfg.Server.Port < 1 || cfg.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", cfg.Server.Port)
	}

	if cfg.Server.Host == "" {
		return fmt.Errorf("server host cannot be empty")
	}

	// Validate Redis config
	if cfg.Redis.Host == "" {
		return fmt.Errorf("redis host cannot be empty")
	}

	if cfg.Redis.Port < 1 || cfg.Redis.Port > 65535 {
		return fmt.Errorf("invalid redis port: %d", cfg.Redis.Port)
	}

	if cfg.Redis.PoolSize < 1 {
		return fmt.Errorf("redis pool size must be at least 1")
	}

	// Validate game config
	if cfg.Game.MapWidth < 10 || cfg.Game.MapWidth > 100 {
		return fmt.Errorf("map width must be between 10 and 100")
	}

	if cfg.Game.MapHeight < 10 || cfg.Game.MapHeight > 100 {
		return fmt.Errorf("map height must be between 10 and 100")
	}

	if cfg.Game.MaxAnimalsPerPlayer < 1 || cfg.Game.MaxAnimalsPerPlayer > 10 {
		return fmt.Errorf("max animals per player must be between 1 and 10")
	}

	if cfg.Game.AnimalSpawnRate < 0 || cfg.Game.AnimalSpawnRate > 1 {
		return fmt.Errorf("animal spawn rate must be between 0 and 1")
	}

	// Validate auth config
	if len(cfg.Auth.JWTSecret) < 8 {
		return fmt.Errorf("JWT secret must be at least 8 characters long")
	}

	if cfg.Auth.JWTExpiration < time.Minute {
		return fmt.Errorf("JWT expiration must be at least 1 minute")
	}

	// Validate log config
	validLogLevels := []string{"debug", "info", "warn", "error"}
	if !contains(validLogLevels, cfg.Log.Level) {
		return fmt.Errorf("invalid log level: %s", cfg.Log.Level)
	}

	validEncodings := []string{"json", "console"}
	if !contains(validEncodings, cfg.Log.Encoding) {
		return fmt.Errorf("invalid log encoding: %s", cfg.Log.Encoding)
	}

	return nil
}

// GetRedisAddr returns the Redis address in host:port format
func (r *RedisConfig) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

// GetServerAddr returns the server address in host:port format
func (s *ServerConfig) GetServerAddr() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

// IsProduction returns true if the environment is production
func (s *ServerConfig) IsProduction() bool {
	return strings.ToLower(s.Environment) == "production"
}

// IsDevelopment returns true if the environment is development
func (s *ServerConfig) IsDevelopment() bool {
	return strings.ToLower(s.Environment) == "development"
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if strings.ToLower(s) == strings.ToLower(item) {
			return true
		}
	}
	return false
}
