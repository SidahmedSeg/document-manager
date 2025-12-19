package logger

import (
	"context"
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ContextKey is a custom type for context keys to avoid collisions
type ContextKey string

const (
	// RequestIDKey is the context key for request ID
	RequestIDKey ContextKey = "request_id"
	// TenantIDKey is the context key for tenant ID
	TenantIDKey ContextKey = "tenant_id"
	// UserIDKey is the context key for user ID
	UserIDKey ContextKey = "user_id"
)

// Logger wraps zap.Logger with additional context methods
type Logger struct {
	*zap.Logger
}

// New creates a new logger instance
func New(environment, level, format string) (*Logger, error) {
	var zapConfig zap.Config

	if environment == "production" {
		zapConfig = zap.NewProductionConfig()
		zapConfig.EncoderConfig.TimeKey = "timestamp"
		zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		zapConfig = zap.NewDevelopmentConfig()
		zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	// Set log level
	zapLevel, err := zapcore.ParseLevel(level)
	if err != nil {
		return nil, fmt.Errorf("invalid log level %s: %w", level, err)
	}
	zapConfig.Level = zap.NewAtomicLevelAt(zapLevel)

	// Set encoding format
	if format == "json" {
		zapConfig.Encoding = "json"
	} else {
		zapConfig.Encoding = "console"
	}

	// Build logger
	zapLogger, err := zapConfig.Build(
		zap.AddCallerSkip(1), // Skip one caller to get correct line numbers
	)
	if err != nil {
		return nil, fmt.Errorf("failed to build logger: %w", err)
	}

	return &Logger{Logger: zapLogger}, nil
}

// NewDefault creates a logger with default settings
func NewDefault() *Logger {
	logger, _ := New("development", "info", "console")
	return logger
}

// WithContext creates a logger with context fields
func (l *Logger) WithContext(ctx context.Context) *zap.Logger {
	logger := l.Logger

	// Add request ID if present
	if requestID, ok := ctx.Value(RequestIDKey).(string); ok && requestID != "" {
		logger = logger.With(zap.String("request_id", requestID))
	}

	// Add tenant ID if present
	if tenantID, ok := ctx.Value(TenantIDKey).(string); ok && tenantID != "" {
		logger = logger.With(zap.String("tenant_id", tenantID))
	}

	// Add user ID if present
	if userID, ok := ctx.Value(UserIDKey).(string); ok && userID != "" {
		logger = logger.With(zap.String("user_id", userID))
	}

	return logger
}

// WithRequestID adds request ID to context
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey, requestID)
}

// WithTenantID adds tenant ID to context
func WithTenantID(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, TenantIDKey, tenantID)
}

// WithUserID adds user ID to context
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

// GetRequestID retrieves request ID from context
func GetRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(RequestIDKey).(string); ok {
		return requestID
	}
	return ""
}

// GetTenantID retrieves tenant ID from context
func GetTenantID(ctx context.Context) string {
	if tenantID, ok := ctx.Value(TenantIDKey).(string); ok {
		return tenantID
	}
	return ""
}

// GetUserID retrieves user ID from context
func GetUserID(ctx context.Context) string {
	if userID, ok := ctx.Value(UserIDKey).(string); ok {
		return userID
	}
	return ""
}

// InfoContext logs an info message with context
func (l *Logger) InfoContext(ctx context.Context, msg string, fields ...zap.Field) {
	l.WithContext(ctx).Info(msg, fields...)
}

// ErrorContext logs an error message with context
func (l *Logger) ErrorContext(ctx context.Context, msg string, fields ...zap.Field) {
	l.WithContext(ctx).Error(msg, fields...)
}

// WarnContext logs a warning message with context
func (l *Logger) WarnContext(ctx context.Context, msg string, fields ...zap.Field) {
	l.WithContext(ctx).Warn(msg, fields...)
}

// DebugContext logs a debug message with context
func (l *Logger) DebugContext(ctx context.Context, msg string, fields ...zap.Field) {
	l.WithContext(ctx).Debug(msg, fields...)
}

// FatalContext logs a fatal message with context and exits
func (l *Logger) FatalContext(ctx context.Context, msg string, fields ...zap.Field) {
	l.WithContext(ctx).Fatal(msg, fields...)
}

// Global logger instance
var global *Logger

// init initializes the global logger
func init() {
	global = NewDefault()
}

// SetGlobal sets the global logger
func SetGlobal(logger *Logger) {
	global = logger
}

// Global returns the global logger
func Global() *Logger {
	return global
}

// Info logs an info message using the global logger
func Info(msg string, fields ...zap.Field) {
	global.Info(msg, fields...)
}

// Error logs an error message using the global logger
func Error(msg string, fields ...zap.Field) {
	global.Error(msg, fields...)
}

// Warn logs a warning message using the global logger
func Warn(msg string, fields ...zap.Field) {
	global.Warn(msg, fields...)
}

// Debug logs a debug message using the global logger
func Debug(msg string, fields ...zap.Field) {
	global.Debug(msg, fields...)
}

// Fatal logs a fatal message using the global logger and exits
func Fatal(msg string, fields ...zap.Field) {
	global.Fatal(msg, fields...)
	os.Exit(1)
}

// InfoContext logs an info message with context using the global logger
func InfoContext(ctx context.Context, msg string, fields ...zap.Field) {
	global.InfoContext(ctx, msg, fields...)
}

// ErrorContext logs an error message with context using the global logger
func ErrorContext(ctx context.Context, msg string, fields ...zap.Field) {
	global.ErrorContext(ctx, msg, fields...)
}

// WarnContext logs a warning message with context using the global logger
func WarnContext(ctx context.Context, msg string, fields ...zap.Field) {
	global.WarnContext(ctx, msg, fields...)
}

// DebugContext logs a debug message with context using the global logger
func DebugContext(ctx context.Context, msg string, fields ...zap.Field) {
	global.DebugContext(ctx, msg, fields...)
}

// FatalContext logs a fatal message with context using the global logger and exits
func FatalContext(ctx context.Context, msg string, fields ...zap.Field) {
	global.FatalContext(ctx, msg, fields...)
}
