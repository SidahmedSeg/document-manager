package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/SidahmedSeg/document-manager/backend/pkg/errors"
	"github.com/SidahmedSeg/document-manager/backend/pkg/logger"
	"github.com/SidahmedSeg/document-manager/backend/pkg/response"
	"go.uber.org/zap"
)

// Header constants for Oathkeeper-injected values
const (
	HeaderUserID       = "X-User-ID"
	HeaderUserEmail    = "X-User-Email"
	HeaderUserName     = "X-User-Name"
	HeaderRequestID    = "X-Request-ID"
	HeaderTenantID     = "X-Tenant-ID"
)

// AuthContext holds authentication information extracted from headers
type AuthContext struct {
	UserID    string
	UserEmail string
	UserName  string
	TenantID  string
}

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const (
	authContextKey contextKey = "auth_context"
)

// ExtractAuthHeaders extracts Oathkeeper headers and adds them to context
func ExtractAuthHeaders(log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := r.Header.Get(HeaderUserID)
			userEmail := r.Header.Get(HeaderUserEmail)
			userName := r.Header.Get(HeaderUserName)
			tenantID := r.Header.Get(HeaderTenantID)

			// For services that require authentication, validate user ID
			if userID == "" {
				response.Error(w, errors.ErrUnauthorized)
				return
			}

			// Create auth context
			authCtx := &AuthContext{
				UserID:    userID,
				UserEmail: userEmail,
				UserName:  userName,
				TenantID:  tenantID,
			}

			// Add auth context to request context
			ctx := context.WithValue(r.Context(), authContextKey, authCtx)

			// Also add individual values to logger context
			ctx = logger.WithUserID(ctx, userID)
			if tenantID != "" {
				ctx = logger.WithTenantID(ctx, tenantID)
			}

			// Continue with updated context
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OptionalAuth extracts headers but doesn't require them (for public endpoints)
func OptionalAuth(log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := r.Header.Get(HeaderUserID)
			userEmail := r.Header.Get(HeaderUserEmail)
			userName := r.Header.Get(HeaderUserName)
			tenantID := r.Header.Get(HeaderTenantID)

			// Create auth context even if values are empty
			authCtx := &AuthContext{
				UserID:    userID,
				UserEmail: userEmail,
				UserName:  userName,
				TenantID:  tenantID,
			}

			ctx := context.WithValue(r.Context(), authContextKey, authCtx)

			if userID != "" {
				ctx = logger.WithUserID(ctx, userID)
			}
			if tenantID != "" {
				ctx = logger.WithTenantID(ctx, tenantID)
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequestID adds a request ID to the context
func RequestID() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get(HeaderRequestID)
			if requestID == "" {
				requestID = uuid.New().String()
			}

			// Add request ID to response header
			w.Header().Set(HeaderRequestID, requestID)

			// Add to context
			ctx := logger.WithRequestID(r.Context(), requestID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Logging logs HTTP requests
func Logging(log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap response writer to capture status code
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			// Process request
			next.ServeHTTP(wrapped, r)

			// Log request
			duration := time.Since(start)
			log.InfoContext(r.Context(), "http request",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("remote_addr", r.RemoteAddr),
				zap.Int("status", wrapped.statusCode),
				zap.Duration("duration", duration),
				zap.String("user_agent", r.UserAgent()),
			)
		})
	}
}

// Recovery recovers from panics and returns 500 error
func Recovery(log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					log.ErrorContext(r.Context(), "panic recovered",
						zap.Any("error", err),
						zap.String("method", r.Method),
						zap.String("path", r.URL.Path),
					)

					response.Error(w, errors.ErrInternal)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// CORS adds CORS headers
func CORS(allowedOrigins []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Check if origin is allowed
			allowed := false
			for _, allowedOrigin := range allowedOrigins {
				if allowedOrigin == "*" || allowedOrigin == origin {
					allowed = true
					break
				}
			}

			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-Request-ID")
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Access-Control-Max-Age", "3600")
			}

			// Handle preflight
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Timeout adds a timeout to the request context
func Timeout(timeout time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetAuthContext retrieves the auth context from the request context
func GetAuthContext(ctx context.Context) *AuthContext {
	authCtx, ok := ctx.Value(authContextKey).(*AuthContext)
	if !ok {
		return &AuthContext{}
	}
	return authCtx
}

// GetUserID retrieves the user ID from context
func GetUserID(ctx context.Context) string {
	authCtx := GetAuthContext(ctx)
	return authCtx.UserID
}

// GetUserEmail retrieves the user email from context
func GetUserEmail(ctx context.Context) string {
	authCtx := GetAuthContext(ctx)
	return authCtx.UserEmail
}

// GetTenantID retrieves the tenant ID from context
func GetTenantID(ctx context.Context) string {
	authCtx := GetAuthContext(ctx)
	return authCtx.TenantID
}

// RequireTenant middleware ensures tenant ID is present
func RequireTenant() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tenantID := GetTenantID(r.Context())
			if tenantID == "" {
				response.Error(w, errors.Forbiddenf("tenant context required"))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
