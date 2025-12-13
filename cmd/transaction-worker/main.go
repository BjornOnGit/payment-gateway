package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/BjornOnGit/payment-gateway/internal/bus/rabbitmq"
	"github.com/BjornOnGit/payment-gateway/internal/db"
	"github.com/BjornOnGit/payment-gateway/internal/db/repo"
	"github.com/BjornOnGit/payment-gateway/internal/service"
	"github.com/BjornOnGit/payment-gateway/internal/util"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	serviceName := os.Getenv("LOG_SERVICE_NAME")
	if serviceName == "" {
		serviceName = "transaction-worker"
	}
	logger, err := util.NewLogger(serviceName)
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	logger.Info("starting transaction worker")

	util.LoadEnv() // loads DB_URL etc

	dsn := os.Getenv("DB_URL")
	if dsn == "" {
		logger.Fatal("DB_URL not set")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Connect to database
	conn, err := db.OpenPostgres(ctx, dsn)
	if err != nil {
		logger.Fatal("failed to connect to postgres", zap.Error(err))
	}
	defer conn.Close()

	logger.Info("postgres connection successful")

	// Initialize repository
	txRepo := repo.NewPostgresTransactionRepository(conn)

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

	// Initialize routing service
	routingService := service.NewRoutingService(txRepo, msgBus)

	// Subscribe to transaction.created events
	if err := msgBus.Subscribe(ctx, "transaction.created", func(ctx context.Context, topic, key string, payload []byte) error {
		logger.Info("transaction-worker received event", zap.String("topic", topic), zap.String("key", key))
		// Use background context to avoid cancellation issues
		return routingService.ProcessTransaction(context.Background(), key)
	}); err != nil {
		logger.Fatal("failed to subscribe to transaction.created", zap.Error(err))
	}

	logger.Info("transaction worker subscribed to transaction.created")

	// Wait for interrupt signal
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig

	logger.Info("shutting down gracefully...")
	cancel()
}
