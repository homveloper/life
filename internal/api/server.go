package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/danghamo/life/internal/api/handlers"
	"github.com/danghamo/life/internal/api/jsonrpcx"
	"github.com/danghamo/life/internal/api/middleware"
	"github.com/danghamo/life/internal/domain/account"
	"github.com/danghamo/life/internal/domain/trainer"
	"github.com/danghamo/life/pkg/logger"
	"github.com/danghamo/life/pkg/redisx"
)

// Server represents the HTTP server
type Server struct {
	httpServer     *http.Server
	logger         *logger.Logger
	redisClient    *redisx.Client
	mux            *http.ServeMux
	trainerHandler *handlers.TrainerHandler
	animalHandler  *handlers.AnimalHandler
	worldHandler   *handlers.WorldHandler
	authHandler    *handlers.AuthHandler
	authMiddleware *middleware.AuthMiddleware
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port         int           `json:"port"`
	Host         string        `json:"host"`
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
	IdleTimeout  time.Duration `json:"idle_timeout"`
}

// NewServer creates a new HTTP server
func NewServer(config ServerConfig, logger *logger.Logger, redisClient *redisx.Client) *Server {
	mux := http.NewServeMux()
	apiLogger := logger.WithComponent("api")

	// Create repositories
	trainerRepo := trainer.NewRedisRepository(redisClient.Client)
	accountRepo := account.NewRedisRepository(redisClient.Client)

	// Create JWT service
	jwtService := account.NewJWTService(
		"your-secret-key-here", // TODO: Move to config
		"life-game-server",
		24*time.Hour, // Token expires in 24 hours
	)

	// Create auth middleware
	authMiddleware := middleware.NewAuthMiddleware(jwtService, apiLogger)

	// Create OAuth configuration (TODO: move to config file)
	oauthConfig := handlers.OAuthConfig{
		Google: handlers.ProviderConfig{
			ClientID:     "your-google-client-id",
			ClientSecret: "your-google-client-secret",
			RedirectURI:  "http://localhost:8080/auth/google/callback",
			AuthURL:      "https://accounts.google.com/o/oauth2/v2/auth",
			TokenURL:     "https://oauth2.googleapis.com/token",
			UserInfoURL:  "https://www.googleapis.com/oauth2/v2/userinfo",
			Scopes:       "openid profile email",
		},
		GitHub: handlers.ProviderConfig{
			ClientID:     "your-github-client-id",
			ClientSecret: "your-github-client-secret",
			RedirectURI:  "http://localhost:8080/auth/github/callback",
			AuthURL:      "https://github.com/login/oauth/authorize",
			TokenURL:     "https://github.com/login/oauth/access_token",
			UserInfoURL:  "https://api.github.com/user",
			Scopes:       "user:email",
		},
		Discord: handlers.ProviderConfig{
			ClientID:     "your-discord-client-id",
			ClientSecret: "your-discord-client-secret",
			RedirectURI:  "http://localhost:8080/auth/discord/callback",
			AuthURL:      "https://discord.com/api/oauth2/authorize",
			TokenURL:     "https://discord.com/api/oauth2/token",
			UserInfoURL:  "https://discord.com/api/users/@me",
			Scopes:       "identify email",
		},
	}

	server := &Server{
		httpServer: &http.Server{
			Addr:         fmt.Sprintf("%s:%d", config.Host, config.Port),
			Handler:      mux,
			ReadTimeout:  config.ReadTimeout,
			WriteTimeout: config.WriteTimeout,
			IdleTimeout:  config.IdleTimeout,
		},
		logger:         apiLogger,
		redisClient:    redisClient,
		mux:            mux,
		trainerHandler: handlers.NewTrainerHandler(apiLogger, trainerRepo),
		animalHandler:  handlers.NewAnimalHandler(apiLogger),
		worldHandler:   handlers.NewWorldHandler(apiLogger),
		authHandler:    handlers.NewAuthHandler(apiLogger, accountRepo, jwtService, oauthConfig),
		authMiddleware: authMiddleware,
	}

	server.setupRoutes()
	server.setupMiddleware()

	return server
}

// setupRoutes configures the server routes
func (s *Server) setupRoutes() {
	// Health check endpoint (pure REST)
	s.mux.HandleFunc("/health", s.healthCheckHandler)

	// Swagger documentation endpoint
	s.mux.HandleFunc("/swagger/", httpSwagger.WrapHandler)

	// General purpose ping endpoint (hybrid)
	s.mux.HandleFunc("/api/v1/ping", s.handlePing)

	// OAuth endpoints (no auth required)
	s.mux.HandleFunc("/api/v1/auth.OAuthStart", s.authHandler.HandleOAuthStart)
	s.mux.HandleFunc("/api/v1/auth.OAuthCallback", s.authHandler.HandleOAuthCallback)
	
	// Guest login (no auth required)
	s.mux.HandleFunc("/api/v1/auth.GuestLogin", s.authHandler.HandleGuestLogin)
	
	// Social linking (auth required - guest accounts only)
	s.mux.Handle("/api/v1/auth.LinkSocial", s.authMiddleware.RequireAuth(http.HandlerFunc(s.authHandler.HandleLinkSocial)))

	// Trainer endpoints (JWT auth required)
	s.mux.Handle("/api/v1/trainer.Create", s.authMiddleware.RequireAuth(http.HandlerFunc(s.trainerHandler.HandleCreate)))
	s.mux.Handle("/api/v1/trainer.Get", s.authMiddleware.RequireAuth(http.HandlerFunc(s.trainerHandler.HandleGet)))
	s.mux.Handle("/api/v1/trainer.Move", s.authMiddleware.RequireAuth(http.HandlerFunc(s.trainerHandler.HandleMove)))
	s.mux.Handle("/api/v1/trainer.List", s.authMiddleware.RequireAuth(http.HandlerFunc(s.trainerHandler.HandleList)))
	s.mux.Handle("/api/v1/trainer.Status", s.authMiddleware.RequireAuth(http.HandlerFunc(s.trainerHandler.HandleStatus)))

	// Animal endpoints (JWT auth required)
	s.mux.Handle("/api/v1/animal.Spawn", s.authMiddleware.RequireAuth(http.HandlerFunc(s.animalHandler.HandleSpawn)))
	s.mux.Handle("/api/v1/animal.Get", s.authMiddleware.RequireAuth(http.HandlerFunc(s.animalHandler.HandleGet)))
	s.mux.Handle("/api/v1/animal.Capture", s.authMiddleware.RequireAuth(http.HandlerFunc(s.animalHandler.HandleCapture)))
	s.mux.Handle("/api/v1/animal.List", s.authMiddleware.RequireAuth(http.HandlerFunc(s.animalHandler.HandleList)))

	// World endpoints (JWT auth required)
	s.mux.Handle("/api/v1/world.Get", s.authMiddleware.RequireAuth(http.HandlerFunc(s.worldHandler.HandleGet)))
}

// setupMiddleware applies middleware to all routes
func (s *Server) setupMiddleware() {
	// Apply middleware chain using functional composition
	middlewareChain := middleware.Chain(
		middleware.RateLimit(s.logger),
		middleware.Recovery(s.logger),
		middleware.ErrorAdapter(s.logger),
		middleware.CORS(),
		middleware.Logging(s.logger),
	)

	s.httpServer.Handler = middlewareChain(s.mux)
}

// Start starts the HTTP server
func (s *Server) Start(ctx context.Context) error {
	s.logger.Info("Starting HTTP server",
		zap.String("address", s.httpServer.Addr))

	// Start server in goroutine
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("HTTP server error", zap.Error(err))
		}
	}()

	// Wait for context cancellation
	<-ctx.Done()

	return s.Shutdown()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown() error {
	s.logger.Info("Shutting down HTTP server")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.logger.Error("Server shutdown error", zap.Error(err))
		return err
	}

	s.logger.Info("HTTP server stopped")
	return nil
}

// GetAddr returns the server address
func (s *Server) GetAddr() string {
	return s.httpServer.Addr
}

// healthCheckHandler handles health check requests
func (s *Server) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	// Check Redis connection
	if err := s.redisClient.HealthCheck(r.Context()); err != nil {
		s.logger.Error("Redis health check failed", zap.Error(err))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`{"status":"unhealthy","checks":{"redis":{"status":"down","error":"` + err.Error() + `"}}}`))
		return
	}

	// All checks passed
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy","checks":{"redis":{"status":"up"}}}`))
}

// handlePing handles ping requests (hybrid JSON-RPC)
func (s *Server) handlePing(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonrpcx.WithError(r, nil, jsonrpcx.MethodNotFound, "Method not allowed")
		return
	}

	req, err := jsonrpcx.ParseRequest(r)
	if err != nil {
		jsonrpcx.WithError(r, nil, jsonrpcx.ParseError, "Invalid JSON-RPC request")
		return
	}

	result := map[string]string{"message": "pong"}
	jsonrpcx.Success(w, req.ID, result)
}
