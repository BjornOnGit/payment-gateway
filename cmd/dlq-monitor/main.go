package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/BjornOnGit/payment-gateway/internal/bus/rabbitmq"
	"github.com/BjornOnGit/payment-gateway/internal/util"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	serviceName := os.Getenv("LOG_SERVICE_NAME")
	if serviceName == "" {
		serviceName = "dlq-monitor"
	}
	logger, err := util.NewLogger(serviceName)
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	logger.Info("starting DLQ monitor")

	util.LoadEnv()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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

	// Subscribe to DLQ for settlement failures
	if err := msgBus.Subscribe(ctx, "dlq.settlement.requested", func(ctx context.Context, topic, key string, payload []byte) error {
		logger.Error("DLQ message received - settlement failed permanently",
			zap.String("topic", topic),
			zap.String("key", key),
			zap.ByteString("payload", payload),
		)

		// In a real system, you might:
		// 1. Store to a database for manual review
		// 2. Send alerts to operations team
		// 3. Trigger manual intervention workflow
		// 4. Log to external monitoring system

		// For now, just log and acknowledge
		return nil
	}); err != nil {
		logger.Fatal("failed to subscribe to DLQ", zap.Error(err))
	}

	logger.Info("DLQ monitor subscribed to dlq.settlement.requested")

	// Wait for interrupt signal
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig

	logger.Info("shutting down DLQ monitor...")
	cancel()
}
