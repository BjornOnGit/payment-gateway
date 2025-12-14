package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/BjornOnGit/payment-gateway/internal/api"
	"github.com/BjornOnGit/payment-gateway/internal/api/middleware"
	"github.com/BjornOnGit/payment-gateway/internal/auth"
	"github.com/BjornOnGit/payment-gateway/internal/bus"
	"github.com/BjornOnGit/payment-gateway/internal/bus/rabbitmq"
	"github.com/BjornOnGit/payment-gateway/internal/db"
	"github.com/BjornOnGit/payment-gateway/internal/db/repo"
	"github.com/BjornOnGit/payment-gateway/internal/service"
	"github.com/BjornOnGit/payment-gateway/internal/util"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	serviceName := os.Getenv("LOG_SERVICE_NAME")
	if serviceName == "" {
		serviceName = "payment-gateway-api"
	}
	logger, err := util.NewLogger(serviceName)
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	logger.Info("starting payment gateway API server")

	util.LoadEnv() // loads DB_URL etc

	dsn := os.Getenv("DB_URL")
	if dsn == "" {
		logger.Fatal("DB_URL not set")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conn, err := db.OpenPostgres(ctx, dsn)
	if err != nil {
		logger.Fatal("failed to connect to postgres", zap.Error(err))
	}
	defer conn.Close()

	logger.Info("postgres connection successful")

	// Initialize repositories
	txRepo := repo.NewPostgresTransactionRepository(conn)
	settlementRepo := repo.NewPostgresSettlementRepository(conn)
	accountRepo := repo.NewPostgresAccountRepository(conn)

	// Initialize RabbitMQ bus
	rabbitmqURL := os.Getenv("RABBITMQ_URL")
	if rabbitmqURL == "" {
		rabbitmqURL = "amqp://guest:guest@localhost:5672/"
	}
	msgBus, err := rabbitmq.NewRabbitMQBus(rabbitmqURL)
	if err != nil {
		logger.Fatal("failed to initialize RabbitMQ bus", zap.Error(err))
	}
	defer msgBus.Close()
	logger.Info("RabbitMQ bus initialized", zap.String("url", rabbitmqURL))

	// Initialize services with RabbitMQ bus
	txService := service.NewTransactionService(txRepo, msgBus)
	routingService := service.NewRoutingService(txRepo, msgBus)
	settlementService := service.NewSettlementService(txRepo, settlementRepo, accountRepo, msgBus, logger)

	// Start transaction-worker subscriber
	startTransactionWorker(ctx, routingService, msgBus, logger)

	// Start settlement-worker subscriber
	startSettlementWorker(ctx, settlementService, msgBus, logger)

	// Initialize Redis client for idempotency
	rdb := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_URL"), // "localhost:6379"
	})
	defer rdb.Close()

	// Ping Redis to verify connection
	if err := rdb.Ping(ctx).Err(); err != nil {
		logger.Warn("redis connection failed (idempotency disabled)", zap.Error(err))
		rdb = nil // Disable idempotency if Redis is unavailable
	} else {
		logger.Info("redis connection successful")
	}

	// Create idempotency store with 24h TTL
	var idempStore *middleware.IdempotencyStore
	if rdb != nil {
		idempStore = &middleware.IdempotencyStore{Redis: rdb, TTL: 24 * time.Hour}
	}

	// Optional: Initialize JWT manager for authentication
	// Only initialize if cert paths are provided
	var jwtManager *auth.JWTManager
	privKeyPath := os.Getenv("JWT_PRIVATE_KEY_PATH")
	pubKeyPath := os.Getenv("JWT_PUBLIC_KEY_PATH")

	if privKeyPath != "" && pubKeyPath != "" {
		var err error
		jwtManager, err = auth.NewJWTManager(
			privKeyPath,
			pubKeyPath,
			"payment-gateway",
			1*time.Hour,
		)
		if err != nil {
			logger.Warn("JWT initialization failed (authentication disabled)", zap.Error(err))
			jwtManager = nil
		} else {
			logger.Info("JWT manager initialized")
		}
	} else {
		logger.Warn("JWT_PRIVATE_KEY_PATH and JWT_PUBLIC_KEY_PATH not set (authentication disabled)")
	}

	// Dev-only OAuth server (client registry in-memory)
	var oauthServer *auth.OAuthServer
	if jwtManager != nil {
		oauthServer = auth.NewOAuthServer(jwtManager)
		// Register a dev client (use env overrides if present)
		clientID := os.Getenv("DEV_CLIENT_ID")
		if clientID == "" {
			clientID = "dev-client"
		}
		clientSecret := os.Getenv("DEV_CLIENT_SECRET")
		if clientSecret == "" {
			clientSecret = "dev-secret"
		}
		oauthServer.RegisterClient(clientID, clientSecret, "Dev Client", []string{"client"})
		logger.Info("OAuth dev client registered", zap.String("client_id", clientID), zap.String("client_secret", clientSecret))
	} else {
		logger.Warn("JWT not initialized; /oauth/token disabled")
	}

	// Initialize router with service and middleware
	router := api.NewRouterWithConfig(api.RouterConfig{
		TxService:        txService,
		SettlementSvc:    settlementService,
		JWTManager:       jwtManager,
		IdempotencyStore: idempStore,
		OAuthServer:      oauthServer,
		Logger:           logger,
	})

	// Create a mux to add metrics endpoint and wrap with metrics middleware
	mux := http.NewServeMux()

	// Add metrics endpoint (public, no auth required)
	mux.Handle("/metrics", util.MetricsHandler())

	// Delegate all other requests to the main router
	mux.Handle("/", router)

	// Wrap entire mux with metrics middleware to track all requests
	handler := util.MetricsMiddleware(mux)

	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8080"
	}

	// TLS certificate paths
	cert := os.Getenv("TLS_CERT_PATH")
	key := os.Getenv("TLS_KEY_PATH")
	if cert == "" {
		cert = "dev-certs/server.crt"
	}
	if key == "" {
		key = "dev-certs/server.key"
	}

	// Start HTTP server in a goroutine
	server := &http.Server{
		Addr:    ":" + port,
		Handler: handler, // Use metrics-wrapped handler
	}

	// Allow switching between HTTPS (self-signed inside container) and plain HTTP for Nginx TLS termination
	tlsMode := os.Getenv("API_TLS")
	if tlsMode == "off" || tlsMode == "false" { // HTTP mode
		go func() {
			logger.Info("API server running", zap.String("port", port), zap.String("protocol", "http"))
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logger.Fatal("HTTP server error", zap.Error(err))
			}
		}()
	} else { // HTTPS mode (default)
		go func() {
			logger.Info("API server running", zap.String("port", port), zap.String("protocol", "https"))
			if err := server.ListenAndServeTLS(cert, key); err != nil && err != http.ErrServerClosed {
				logger.Fatal("HTTPS server error", zap.Error(err))
			}
		}()
	}

	// Wait for interrupt signal
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig

	logger.Info("shutting down gracefully...")
	cancel()
	server.Shutdown(context.Background())
	msgBus.Close()
}

func startTransactionWorker(ctx context.Context, routingService *service.RoutingService, b bus.Bus, logger *zap.Logger) {
	if err := b.Subscribe(ctx, "transaction.created", func(ctx context.Context, topic, key string, payload []byte) error {
		logger.Info("transaction-worker received event", zap.String("topic", topic), zap.String("key", key), zap.String("payload", string(payload)))
		// Use background context to avoid cancellation issues from HTTP request context
		return routingService.ProcessTransaction(context.Background(), key)
	}); err != nil {
		logger.Fatal("transaction-worker subscribe failed", zap.Error(err))
	}
	logger.Info("transaction worker subscribed to transaction.created")
}

func startSettlementWorker(ctx context.Context, settlementService *service.SettlementService, b bus.Bus, logger *zap.Logger) {
	if err := b.Subscribe(ctx, "settlement.requested", func(ctx context.Context, topic, key string, payload []byte) error {
		logger.Info("settlement-worker received event", zap.String("topic", topic), zap.String("key", key))
		// Use background context to avoid cancellation issues from HTTP request context
		return settlementService.ProcessSettlement(context.Background(), payload)
	}); err != nil {
		logger.Fatal("settlement-worker subscribe failed", zap.Error(err))
	}
	logger.Info("settlement worker subscribed to settlement.requested")
}
