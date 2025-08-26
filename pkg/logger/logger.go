package logger

import (
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger wraps zap.Logger to provide our application-specific logging interface
type Logger struct {
	*zap.Logger
}

// LogLevel represents the logging level
type LogLevel string

const (
	DebugLevel LogLevel = "debug"
	InfoLevel  LogLevel = "info"
	WarnLevel  LogLevel = "warn"
	ErrorLevel LogLevel = "error"
)

// Config holds logger configuration
type Config struct {
	Level       LogLevel `mapstructure:"level"`
	Environment string   `mapstructure:"environment"`
	Encoding    string   `mapstructure:"encoding"` // json or console
}

// New creates a new logger instance based on configuration
func New(cfg Config) (*Logger, error) {
	// Default configuration
	if cfg.Level == "" {
		cfg.Level = InfoLevel
	}
	if cfg.Environment == "" {
		cfg.Environment = "development"
	}
	if cfg.Encoding == "" {
		if cfg.Environment == "production" {
			cfg.Encoding = "json"
		} else {
			cfg.Encoding = "console"
		}
	}

	// Configure zap level
	var zapLevel zapcore.Level
	switch cfg.Level {
	case DebugLevel:
		zapLevel = zapcore.DebugLevel
	case InfoLevel:
		zapLevel = zapcore.InfoLevel
	case WarnLevel:
		zapLevel = zapcore.WarnLevel
	case ErrorLevel:
		zapLevel = zapcore.ErrorLevel
	default:
		zapLevel = zapcore.InfoLevel
	}

	// Configure encoder
	var encoderConfig zapcore.EncoderConfig
	if cfg.Environment == "production" {
		encoderConfig = zap.NewProductionEncoderConfig()
	} else {
		encoderConfig = zap.NewDevelopmentEncoderConfig()
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	// Configure time format
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// Create encoder
	var encoder zapcore.Encoder
	if cfg.Encoding == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// Create core
	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(os.Stdout),
		zapLevel,
	)

	// Create logger with caller info
	zapLogger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	return &Logger{Logger: zapLogger}, nil
}

// NewDefault creates a logger with default development settings
func NewDefault() *Logger {
	logger, _ := New(Config{
		Level:       DebugLevel,
		Environment: "development",
		Encoding:    "console",
	})
	return logger
}

// NewProduction creates a logger with production settings
func NewProduction() *Logger {
	logger, _ := New(Config{
		Level:       InfoLevel,
		Environment: "production",
		Encoding:    "json",
	})
	return logger
}

// WithField adds a field to the logger context
func (l *Logger) WithField(key string, value interface{}) *Logger {
	return &Logger{Logger: l.Logger.With(zap.Any(key, value))}
}

// WithFields adds multiple fields to the logger context
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	zapFields := make([]zap.Field, 0, len(fields))
	for k, v := range fields {
		zapFields = append(zapFields, zap.Any(k, v))
	}
	return &Logger{Logger: l.Logger.With(zapFields...)}
}

// WithComponent adds a component field to help identify log sources
func (l *Logger) WithComponent(component string) *Logger {
	return l.WithField("component", component)
}

// WithRequestID adds a request ID for tracing
func (l *Logger) WithRequestID(requestID string) *Logger {
	return l.WithField("request_id", requestID)
}

// WithUserID adds a user ID for user-specific operations
func (l *Logger) WithUserID(userID string) *Logger {
	return l.WithField("user_id", userID)
}

// ParseLevel parses a string log level to LogLevel
func ParseLevel(level string) LogLevel {
	switch strings.ToLower(level) {
	case "debug":
		return DebugLevel
	case "info":
		return InfoLevel
	case "warn", "warning":
		return WarnLevel
	case "error":
		return ErrorLevel
	default:
		return InfoLevel
	}
}

// Global logger instance for convenience
var globalLogger *Logger

// SetGlobalLogger sets the global logger instance
func SetGlobalLogger(logger *Logger) {
	globalLogger = logger
}

// GetGlobalLogger returns the global logger instance
func GetGlobalLogger() *Logger {
	if globalLogger == nil {
		globalLogger = NewDefault()
	}
	return globalLogger
}

// Convenience functions using global logger
func Debug(msg string, fields ...zap.Field) {
	GetGlobalLogger().Debug(msg, fields...)
}

func Info(msg string, fields ...zap.Field) {
	GetGlobalLogger().Info(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	GetGlobalLogger().Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	GetGlobalLogger().Error(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	GetGlobalLogger().Fatal(msg, fields...)
}

func Sync() error {
	return GetGlobalLogger().Sync()
}
