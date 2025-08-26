package redisx

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/danghamo/life/pkg/config"
	"github.com/danghamo/life/pkg/logger"
)

// Client wraps redis.Client with additional functionality
type Client struct {
	*redis.Client
	url    string
	logger *logger.Logger
}

// ClientOption represents an option for creating a new Redis client
type ClientOption func(*clientOptions)

// clientOptions holds configuration options for the Redis client
type clientOptions struct {
	usePrivateDB bool
}

// WithPrivate enables private DB isolation for development environments
// This will automatically assign a unique DB number based on hostname
func WithPrivate() ClientOption {
	return func(opts *clientOptions) {
		opts.usePrivateDB = true
	}
}

// NewClient creates a new Redis client from URL with options
func NewClient(redisURL string, log *logger.Logger, opts ...ClientOption) (*Client, error) {
	if redisURL == "" {
		return nil, fmt.Errorf("redis URL cannot be empty")
	}

	if log == nil {
		log = logger.GetGlobalLogger()
	}

	// Process options
	options := &clientOptions{}
	for _, opt := range opts {
		opt(options)
	}

	// Apply private DB if requested
	finalURL := redisURL
	if options.usePrivateDB {
		var err error
		finalURL, err = PrivateUrl(redisURL)
		if err != nil {
			return nil, fmt.Errorf("failed to get private URL: %w", err)
		}
	}

	// Parse Redis URL and create client
	redisOptions, err := redis.ParseURL(finalURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	rdb := redis.NewClient(redisOptions)

	client := &Client{
		Client: rdb,
		url:    finalURL,
		logger: log.WithComponent("redisx"),
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logFields := []zap.Field{
		zap.String("addr", redisOptions.Addr),
		zap.Int("db", redisOptions.DB),
		zap.Int("pool_size", redisOptions.PoolSize),
	}

	if options.usePrivateDB {
		logFields = append(logFields, zap.Bool("private_db", true))
	}

	client.logger.Info("Redis client connected successfully", logFields...)

	return client, nil
}

// NewClientFromConfig creates a new Redis client from config (deprecated, use NewClient with URL)
func NewClientFromConfig(cfg *config.RedisConfig, log *logger.Logger) (*Client, error) {
	if cfg == nil {
		return nil, fmt.Errorf("redis config cannot be nil")
	}

	// Convert config to URL format
	redisURL := fmt.Sprintf("redis://%s", cfg.GetRedisAddr())
	if cfg.Password != "" {
		redisURL = fmt.Sprintf("redis://:%s@%s", cfg.Password, cfg.GetRedisAddr())
	}
	if cfg.DB != 0 {
		redisURL = fmt.Sprintf("%s/%d", redisURL, cfg.DB)
	}

	return NewClient(redisURL, log)
}

// Close closes the Redis client connection
func (c *Client) Close() error {
	c.logger.Info("Closing Redis connection")
	return c.Client.Close()
}

// HealthCheck performs a health check on the Redis connection
func (c *Client) HealthCheck(ctx context.Context) error {
	start := time.Now()
	err := c.Ping(ctx).Err()
	duration := time.Since(start)

	if err != nil {
		c.logger.Error("Redis health check failed",
			zap.Error(err),
			zap.Duration("duration", duration),
		)
		return err
	}

	c.logger.Debug("Redis health check passed",
		zap.Duration("duration", duration),
	)

	return nil
}

// SetWithExpiration sets a key-value pair with expiration
func (c *Client) SetWithExpiration(ctx context.Context, key string, value any, expiration time.Duration) error {
	start := time.Now()
	err := c.Set(ctx, key, value, expiration).Err()
	duration := time.Since(start)

	if err != nil {
		c.logger.Error("Failed to set key with expiration",
			zap.String("key", key),
			zap.Duration("expiration", expiration),
			zap.Duration("duration", duration),
			zap.Error(err),
		)
		return err
	}

	c.logger.Debug("Set key with expiration",
		zap.String("key", key),
		zap.Duration("expiration", expiration),
		zap.Duration("duration", duration),
	)

	return nil
}

// GetWithLogging gets a value with logging
func (c *Client) GetWithLogging(ctx context.Context, key string) (string, error) {
	start := time.Now()
	result := c.Get(ctx, key)
	duration := time.Since(start)

	if result.Err() != nil {
		if result.Err() == redis.Nil {
			c.logger.Debug("Key not found",
				zap.String("key", key),
				zap.Duration("duration", duration),
			)
		} else {
			c.logger.Error("Failed to get key",
				zap.String("key", key),
				zap.Duration("duration", duration),
				zap.Error(result.Err()),
			)
		}
		return "", result.Err()
	}

	c.logger.Debug("Got key",
		zap.String("key", key),
		zap.Duration("duration", duration),
	)

	return result.Val(), nil
}

// DelWithLogging deletes a key with logging
func (c *Client) DelWithLogging(ctx context.Context, keys ...string) (int64, error) {
	start := time.Now()
	result := c.Del(ctx, keys...)
	duration := time.Since(start)

	if result.Err() != nil {
		c.logger.Error("Failed to delete keys",
			zap.Strings("keys", keys),
			zap.Duration("duration", duration),
			zap.Error(result.Err()),
		)
		return 0, result.Err()
	}

	c.logger.Debug("Deleted keys",
		zap.Strings("keys", keys),
		zap.Int64("deleted_count", result.Val()),
		zap.Duration("duration", duration),
	)

	return result.Val(), nil
}

// HSetWithLogging sets a hash field with logging
func (c *Client) HSetWithLogging(ctx context.Context, key string, values ...any) error {
	start := time.Now()
	err := c.HSet(ctx, key, values...).Err()
	duration := time.Since(start)

	if err != nil {
		c.logger.Error("Failed to set hash field",
			zap.String("key", key),
			zap.Duration("duration", duration),
			zap.Error(err),
		)
		return err
	}

	c.logger.Debug("Set hash field",
		zap.String("key", key),
		zap.Duration("duration", duration),
	)

	return nil
}

// HGetAllWithLogging gets all hash fields with logging
func (c *Client) HGetAllWithLogging(ctx context.Context, key string) (map[string]string, error) {
	start := time.Now()
	result := c.HGetAll(ctx, key)
	duration := time.Since(start)

	if result.Err() != nil {
		c.logger.Error("Failed to get all hash fields",
			zap.String("key", key),
			zap.Duration("duration", duration),
			zap.Error(result.Err()),
		)
		return nil, result.Err()
	}

	c.logger.Debug("Got all hash fields",
		zap.String("key", key),
		zap.Int("field_count", len(result.Val())),
		zap.Duration("duration", duration),
	)

	return result.Val(), nil
}

// PrivateUrl provides development isolation by assigning unique DB numbers based on hostname
// This allows multiple developers to use the same Redis instance without conflicts
// It uses Redis DB 0 to store hostname->DB mappings and auto-increments DB numbers
func PrivateUrl(redisURL string) (string, error) {
	if redisURL == "" {
		return "", fmt.Errorf("redis URL cannot be empty")
	}

	// Get current system hostname
	hostname, err := os.Hostname()
	if err != nil {
		return "", fmt.Errorf("failed to get hostname: %w", err)
	}

	return privateUrlWithHostname(redisURL, hostname)
}

// privateUrlWithHostname is a testable version that accepts hostname as parameter
func privateUrlWithHostname(redisURL, hostname string) (string, error) {
	if redisURL == "" {
		return "", fmt.Errorf("redis URL cannot be empty")
	}

	if hostname == "" {
		return "", fmt.Errorf("hostname cannot be empty")
	}

	// Parse the URL
	parsedURL, err := url.Parse(redisURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	// Connect to Redis DB 0 to manage private DB assignments
	db0URL := *parsedURL
	db0URL.Path = "/0"

	options, err := redis.ParseURL(db0URL.String())
	if err != nil {
		return "", fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	rdb := redis.NewClient(options)
	defer rdb.Close()

	ctx := context.Background()

	// Check if hostname already has an assigned DB
	dbNumber, err := rdb.HGet(ctx, "private_db", hostname).Result()
	if err == redis.Nil {
		// Hostname not found, assign new DB number
		// Use HINCRBY to atomically get next available DB number
		nextDB, err := rdb.HIncrBy(ctx, "private_db:counter", "next", 1).Result()
		if err != nil {
			return "", fmt.Errorf("failed to get next DB number: %w", err)
		}

		// DB 0 is reserved for management, start from 1
		// Redis supports up to 16 databases by default, but this can be configured higher

		// Assign the DB number to this hostname
		err = rdb.HSet(ctx, "private_db", hostname, nextDB).Err()
		if err != nil {
			return "", fmt.Errorf("failed to assign DB to hostname: %w", err)
		}

		dbNumber = strconv.FormatInt(nextDB, 10)
	} else if err != nil {
		return "", fmt.Errorf("failed to check existing DB assignment: %w", err)
	}

	// Build the new URL with the assigned DB number
	newURL := *parsedURL
	newURL.Path = "/" + dbNumber

	return newURL.String(), nil
}
