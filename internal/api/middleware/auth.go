package middleware

import (
	"context"
	"net/http"
	"strings"

	"go.uber.org/zap"

	"github.com/danghamo/life/internal/api/jsonrpcx"
	"github.com/danghamo/life/internal/domain/account"
	"github.com/danghamo/life/pkg/logger"
)

// UserContextKey is the key for storing user info in request context
type UserContextKey string

const (
	// UserIDContextKey stores the user ID in context
	UserIDContextKey UserContextKey = "user_id"
	// UserEmailContextKey stores the user email in context  
	UserEmailContextKey UserContextKey = "user_email"
	// UserNameContextKey stores the user name in context  
	UserNameContextKey UserContextKey = "user_name"
)

// AuthMiddleware provides JWT authentication middleware
type AuthMiddleware struct {
	jwtService *account.JWTService
	logger     *logger.Logger
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(jwtService *account.JWTService, logger *logger.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		jwtService: jwtService,
		logger:     logger.WithComponent("auth-middleware"),
	}
}

// RequireAuth returns a middleware that requires JWT authentication
func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			m.logger.Debug("Missing Authorization header")
			jsonrpcx.WithError(r, nil, jsonrpcx.InvalidRequest, "Missing Authorization header")
			return
		}

		// Check Bearer token format
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			m.logger.Debug("Invalid Authorization header format")
			jsonrpcx.WithError(r, nil, jsonrpcx.InvalidRequest, "Invalid Authorization header format")
			return
		}

		tokenString := parts[1]

		// Validate JWT token
		claims, err := m.jwtService.ValidateToken(tokenString)
		if err != nil {
			m.logger.Debug("Invalid JWT token", zap.Error(err))
			jsonrpcx.WithError(r, nil, jsonrpcx.InvalidRequest, "Invalid or expired token")
			return
		}

		// Add user info to request context
		ctx := context.WithValue(r.Context(), UserIDContextKey, claims.UserID)
		ctx = context.WithValue(ctx, UserEmailContextKey, claims.Email)
		ctx = context.WithValue(ctx, UserNameContextKey, claims.Name)

		// Log successful authentication
		m.logger.Debug("JWT authentication successful", 
			zap.String("userId", claims.UserID),
			zap.String("email", claims.Email),
			zap.String("name", claims.Name))

		// Continue to next handler with user context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// OptionalAuth returns a middleware that optionally extracts user info if JWT is present
func (m *AuthMiddleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			// No auth header, continue without user context
			next.ServeHTTP(w, r)
			return
		}

		// Check Bearer token format
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			// Invalid format, continue without user context
			next.ServeHTTP(w, r)
			return
		}

		tokenString := parts[1]

		// Validate JWT token
		claims, err := m.jwtService.ValidateToken(tokenString)
		if err != nil {
			// Invalid token, continue without user context
			m.logger.Debug("Optional auth failed", zap.Error(err))
			next.ServeHTTP(w, r)
			return
		}

		// Add user info to request context
		ctx := context.WithValue(r.Context(), UserIDContextKey, claims.UserID)
		ctx = context.WithValue(ctx, UserEmailContextKey, claims.Email)
		ctx = context.WithValue(ctx, UserNameContextKey, claims.Name)

		m.logger.Debug("Optional JWT authentication successful", 
			zap.String("userId", claims.UserID),
			zap.String("email", claims.Email),
			zap.String("name", claims.Name))

		// Continue to next handler with user context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserID extracts user ID from request context
func GetUserID(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDContextKey).(string)
	return userID, ok
}

// GetUserEmail extracts user email from request context
func GetUserEmail(ctx context.Context) (string, bool) {
	email, ok := ctx.Value(UserEmailContextKey).(string)
	return email, ok
}

// GetUserName extracts user name from request context
func GetUserName(ctx context.Context) (string, bool) {
	name, ok := ctx.Value(UserNameContextKey).(string)
	return name, ok
}

// GetUserInfo extracts user ID, email, and name from request context
func GetUserInfo(ctx context.Context) (userID, email, name string, ok bool) {
	userID, hasUserID := GetUserID(ctx)
	email, hasEmail := GetUserEmail(ctx)
	name, hasName := GetUserName(ctx)
	return userID, email, name, hasUserID && hasEmail && hasName
}