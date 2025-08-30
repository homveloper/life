package middleware

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
	"golang.org/x/time/rate"

	"github.com/danghamo/life/internal/api/jsonrpcx"
	"github.com/danghamo/life/pkg/logger"
)

// Middleware is a function that wraps an http.Handler
type Middleware func(http.Handler) http.Handler

// Chain applies middleware in order (last applied, first executed)
func Chain(middlewares ...Middleware) Middleware {
	return func(final http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			final = middlewares[i](final)
		}
		return final
	}
}

// Logging middleware
func Logging(logger *logger.Logger) Middleware {
	l := logger.WithComponent("logging-middleware")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Create a response writer that captures status code
			wrapped := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			next.ServeHTTP(wrapped, r)

			duration := time.Since(start)

			l.Info("HTTP request",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("remote_addr", r.RemoteAddr),
				zap.String("user_agent", r.UserAgent()),
				zap.Int("status_code", wrapped.statusCode),
				zap.Duration("duration", duration),
			)
		})
	}
}

// CORS middleware
func CORS() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Set CORS headers
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Max-Age", "86400")

			// Handle preflight request
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Recovery middleware (renamed from ErrorHandling)
func Recovery(logger *logger.Logger) Middleware {
	l := logger.WithComponent("recovery-middleware")
	errorAdapter := jsonrpcx.NewErrorAdapter()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					l.Error("HTTP handler panic",
						zap.Any("error", err),
						zap.String("path", r.URL.Path),
						zap.String("method", r.Method),
					)

					// Set generic 500 error response using ErrorAdapter
					errorAdapter.SendError(w, nil, jsonrpcx.InternalError, "Internal server error")
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// ErrorAdapter middleware handles errors and converts them to JSON-RPC responses
func ErrorAdapter(logger *logger.Logger) Middleware {
	l := logger.WithComponent("error-adapter-middleware")
	errorAdapter := jsonrpcx.NewErrorAdapter()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)

			// Check if there's a JSON-RPC error in the context
			if rpcResponse, ok := r.Context().Value("jsonrpc_error").(*jsonrpcx.Response); ok {
				jsonrpcx.Write(w, *rpcResponse)
				return
			}

			// Check if there's a generic error in the context
			if err, ok := r.Context().Value("error").(error); ok {
				// Non-specific error - return 500
				l.Error("Error encountered", zap.Error(err))
				errorAdapter.SendError(w, nil, jsonrpcx.InternalError, "Internal server error")
			}
		})
	}
}

// RateLimit middleware
func RateLimit(logger *logger.Logger) Middleware {
	l := logger.WithComponent("ratelimit-middleware")

	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	// Cleanup old clients every minute
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			mu.Lock()
			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := getClientIP(r)

			mu.Lock()
			if _, exists := clients[ip]; !exists {
				// Allow 600 requests per minute per IP (10 per second)
				clients[ip] = &client{
					limiter: rate.NewLimiter(rate.Every(time.Second/10), 50),
				}
			}
			clients[ip].lastSeen = time.Now()
			limiter := clients[ip].limiter
			mu.Unlock()

			if !limiter.Allow() {
				l.Warn("Rate limit exceeded",
					zap.String("ip", ip),
					zap.String("path", r.URL.Path),
				)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"error": "Rate limit exceeded"}`))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code while preserving interfaces
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Flush implements http.Flusher interface for SSE support
func (rw *responseWriter) Flush() {
	if flusher, ok := rw.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

// Hijack implements http.Hijacker interface if the underlying ResponseWriter supports it
func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := rw.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, fmt.Errorf("responseWriter does not implement http.Hijacker")
}

// getClientIP extracts the client IP from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}
