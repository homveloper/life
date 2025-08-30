package api

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"

	"github.com/danghamo/life/internal/api/handlers"
	"github.com/danghamo/life/internal/api/jsonrpcx"
	"github.com/danghamo/life/internal/api/middleware"
	cqrshandlers "github.com/danghamo/life/internal/cqrs/handlers"
	"github.com/danghamo/life/internal/domain/account"
	"github.com/danghamo/life/internal/domain/trainer"
	"github.com/danghamo/life/pkg/logger"
	"github.com/danghamo/life/pkg/redisx"
	"github.com/danghamo/life/pkg/sse"
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
	serverHandler  *handlers.ServerHandler
	authMiddleware *middleware.AuthMiddleware
	sseBroadcaster *sse.SSEBroadcaster
	// Watermill CQRS components
	commandBus       *cqrs.CommandBus
	eventBus         *cqrs.EventBus
	commandProcessor *cqrs.CommandProcessor
	eventProcessor   *cqrs.EventProcessor
	router          *message.Router
	sseEventHandler *cqrshandlers.SSEEventHandler
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

	// Generate unique server ID
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "unknown"
	}
	serverID := fmt.Sprintf("%s-%d", hostname, time.Now().UnixNano())

	// Create Watermill logger
	watermillLogger := watermill.NewStdLogger(false, false)

	// Create Redis publisher and subscriber
	publisher, err := redisstream.NewPublisher(
		redisstream.PublisherConfig{
			Client: redisClient.Client,
		},
		watermillLogger,
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to create publisher: %v", err))
	}

	subscriber, err := redisstream.NewSubscriber(
		redisstream.SubscriberConfig{
			Client:        redisClient.Client,
			ConsumerGroup: fmt.Sprintf("game-server-%s", serverID),
		},
		watermillLogger,
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to create subscriber: %v", err))
	}

	// Create message router with short close timeout
	router, err := message.NewRouter(message.RouterConfig{
		CloseTimeout: 5 * time.Second, // Short timeout for graceful shutdown
	}, watermillLogger)
	if err != nil {
		panic(fmt.Sprintf("Failed to create router: %v", err))
	}

	// Create command bus
	commandBus, err := cqrs.NewCommandBusWithConfig(
		publisher,
		cqrs.CommandBusConfig{
			GeneratePublishTopic: func(params cqrs.CommandBusGeneratePublishTopicParams) (string, error) {
				return fmt.Sprintf("game-commands.%s", params.CommandName), nil
			},
			Marshaler: cqrs.JSONMarshaler{},
			Logger:    watermillLogger,
		},
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to create command bus: %v", err))
	}

	// Create event bus
	eventBus, err := cqrs.NewEventBusWithConfig(
		publisher,
		cqrs.EventBusConfig{
			GeneratePublishTopic: func(params cqrs.GenerateEventPublishTopicParams) (string, error) {
				return fmt.Sprintf("game-events.%s", params.EventName), nil
			},
			Marshaler: cqrs.JSONMarshaler{},
			Logger:    watermillLogger,
		},
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to create event bus: %v", err))
	}

	// Create command processor
	commandProcessor, err := cqrs.NewCommandProcessorWithConfig(
		router,
		cqrs.CommandProcessorConfig{
			GenerateSubscribeTopic: func(params cqrs.CommandProcessorGenerateSubscribeTopicParams) (string, error) {
				return fmt.Sprintf("game-commands.%s", params.CommandName), nil
			},
			SubscriberConstructor: func(params cqrs.CommandProcessorSubscriberConstructorParams) (message.Subscriber, error) {
				return subscriber, nil
			},
			Marshaler: cqrs.JSONMarshaler{},
			Logger:    watermillLogger,
		},
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to create command processor: %v", err))
	}

	// Create event processor
	eventProcessor, err := cqrs.NewEventProcessorWithConfig(
		router,
		cqrs.EventProcessorConfig{
			GenerateSubscribeTopic: func(params cqrs.EventProcessorGenerateSubscribeTopicParams) (string, error) {
				return fmt.Sprintf("game-events.%s", params.EventName), nil
			},
			SubscriberConstructor: func(params cqrs.EventProcessorSubscriberConstructorParams) (message.Subscriber, error) {
				return subscriber, nil
			},
			Marshaler: cqrs.JSONMarshaler{},
			Logger:    watermillLogger,
		},
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to create event processor: %v", err))
	}

	// Create SSE broadcaster
	sseBroadcaster := sse.NewSSEBroadcaster(apiLogger)

	// Create event handlers
	sseEventHandler := cqrshandlers.NewSSEEventHandler(
		sseBroadcaster, // SSEBroadcaster interface
		eventBus,       // EventPublisher interface
		apiLogger,
	)

	server := &Server{
		httpServer: &http.Server{
			Addr:         fmt.Sprintf("%s:%d", config.Host, config.Port),
			Handler:      mux,
			ReadTimeout:  config.ReadTimeout,
			WriteTimeout: config.WriteTimeout,
			IdleTimeout:  config.IdleTimeout,
		},
		logger:            apiLogger,
		redisClient:       redisClient,
		mux:               mux,
		trainerHandler:    handlers.NewTrainerHandler(apiLogger, trainerRepo, eventBus),
		animalHandler:     handlers.NewAnimalHandler(apiLogger),
		worldHandler:      handlers.NewWorldHandler(apiLogger),
		authHandler:       handlers.NewAuthHandler(apiLogger, accountRepo, jwtService, oauthConfig),
		serverHandler:     handlers.NewServerHandler(),
		authMiddleware:    authMiddleware,
		sseBroadcaster:    sseBroadcaster,
		commandBus:        commandBus,
		eventBus:          eventBus,
		commandProcessor:  commandProcessor,
		eventProcessor:    eventProcessor,
		router:            router,
		sseEventHandler:   sseEventHandler,
	}

	// Register only event handlers for SSE broadcasting
	err = eventProcessor.AddHandlers(
		cqrs.NewEventHandler("TrainerMovedEvent", sseEventHandler.HandleTrainerMovedEvent),
		cqrs.NewEventHandler("TrainerStoppedEvent", sseEventHandler.HandleTrainerStoppedEvent),
		cqrs.NewEventHandler("TrainerCreatedEvent", sseEventHandler.HandleTrainerCreatedEvent),
		cqrs.NewEventHandler("SSENotificationEvent", sseEventHandler.HandleSSENotificationEvent),
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to register event handlers: %v", err))
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

	// Server info endpoint (no auth required)
	s.mux.HandleFunc("/api/v1/server.Info", s.serverHandler.HandleServerInfo)

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
	s.mux.Handle("/api/v1/trainer.FetchPosition", s.authMiddleware.RequireAuth(http.HandlerFunc(s.trainerHandler.HandleFetchPosition)))
	s.mux.Handle("/api/v1/trainer.List", s.authMiddleware.RequireAuth(http.HandlerFunc(s.trainerHandler.HandleList)))
	s.mux.Handle("/api/v1/trainer.Status", s.authMiddleware.RequireAuth(http.HandlerFunc(s.trainerHandler.HandleStatus)))

	// Animal endpoints (JWT auth required)
	s.mux.Handle("/api/v1/animal.Spawn", s.authMiddleware.RequireAuth(http.HandlerFunc(s.animalHandler.HandleSpawn)))
	s.mux.Handle("/api/v1/animal.Get", s.authMiddleware.RequireAuth(http.HandlerFunc(s.animalHandler.HandleGet)))
	s.mux.Handle("/api/v1/animal.Capture", s.authMiddleware.RequireAuth(http.HandlerFunc(s.animalHandler.HandleCapture)))
	s.mux.Handle("/api/v1/animal.List", s.authMiddleware.RequireAuth(http.HandlerFunc(s.animalHandler.HandleList)))

	// World endpoints (JWT auth required)
	s.mux.Handle("/api/v1/world.Get", s.authMiddleware.RequireAuth(http.HandlerFunc(s.worldHandler.HandleGet)))

	// Static file serving for client
	s.mux.HandleFunc("/", s.handleStaticFiles)

	// SSE endpoint for real-time updates (uses dedicated SSE auth middleware)
	s.mux.Handle("/api/v1/stream/positions", s.authMiddleware.RequireSSEAuth(http.HandlerFunc(s.sseBroadcaster.HandleSSE)))
}

// setupMiddleware applies middleware to all routes
func (s *Server) setupMiddleware() {
	// Apply middleware chain using functional composition
	middlewareChain := middleware.Chain(
		// middleware.RateLimit(s.logger), // Disabled for real-time movement
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

	// Start Watermill router first
	go func() {
		if err := s.router.Run(ctx); err != nil {
			s.logger.Error("Watermill router error", zap.Error(err))
		}
	}()

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

	// Shutdown SSE broadcaster first to close client connections
	if s.sseBroadcaster != nil {
		s.logger.Debug("Closing SSE broadcaster")
		s.sseBroadcaster.Close()
	}

	// Shutdown HTTP server with shorter timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.logger.Error("Server shutdown error", zap.Error(err))
		return err
	}

	// Shutdown Watermill router (with CloseTimeout already configured)
	if s.router != nil {
		s.logger.Info("Closing Watermill router")
		if err := s.router.Close(); err != nil {
			s.logger.Error("Router shutdown error", zap.Error(err))
			return err
		}
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

// handleStaticFiles serves static files with proper MIME types
func (s *Server) handleStaticFiles(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/":
		// Serve the trainer client HTML file
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		http.ServeFile(w, r, "trainer-client.html")
		return
	case "/trainer-client.js":
		// Serve JavaScript file with correct MIME type
		w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
		http.ServeFile(w, r, "trainer-client.js")
		return
	case "/trainer-client.html":
		// Also allow direct access to HTML file
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		http.ServeFile(w, r, "trainer-client.html")
		return
	default:
		// 404 for other paths
		http.NotFound(w, r)
	}
}
