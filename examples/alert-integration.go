// Package main demonstrates how to integrate AlertClient with payment gateway services
package main

import (
	"context"
	"log"
	"os"

	"github.com/BjornOnGit/payment-gateway/internal/util"
	"go.uber.org/zap"
)

// Example 1: DLQ Monitor with Alerts
func dlqMonitorExample(alertClient *util.AlertClient, logger *zap.Logger) {
	ctx := context.Background()

	// When a message arrives in the DLQ
	logger.Error("DLQ message received - settlement failed permanently")

	// Send critical alert
	if err := alertClient.SendCritical(ctx, "dlq-monitor",
		"Settlement failed permanently - in DLQ",
		map[string]any{
			"transaction_id": "123e4567-e89b-12d3-a456-426614174000",
			"topic":          "dlq.settlement.requested",
			"message":        "Failed after 3 retries",
		}); err != nil {
		logger.Error("failed to send alert", zap.Error(err))
	}
}

// Example 2: Settlement Worker with Alerts
func settlementWorkerExample(alertClient *util.AlertClient, logger *zap.Logger) {
	ctx := context.Background()

	// Simulating settlement processing failure
	simulatedError := "external_api_timeout"

	logger.Error("settlement processing failed", zap.String("error", simulatedError))

	// Send error alert
	if err := alertClient.SendError(ctx, "settlement-worker",
		"Settlement processing failed",
		map[string]any{
			"transaction_id": "789a0123-b456-78d9-c012-345678901234",
			"error":          simulatedError,
			"retry_count":    2,
			"next_retry_in":  "1m",
		}); err != nil {
		logger.Error("failed to send alert", zap.Error(err))
	}
}

// Example 3: API Server with Alerts
func apiServerExample(alertClient *util.AlertClient, logger *zap.Logger) {
	ctx := context.Background()

	// Alert on anomalous behavior
	transactionCount := 5000 // per minute
	threshold := 1000

	if transactionCount > threshold {
		logger.Warn("high transaction rate detected", zap.Int("rate", transactionCount))

		// Send warning alert
		if err := alertClient.SendWarning(ctx, "api-server",
			"High transaction rate detected",
			map[string]any{
				"transactions_per_minute": transactionCount,
				"threshold":               threshold,
				"percentage":              (transactionCount - threshold) / threshold * 100,
			}); err != nil {
			logger.Error("failed to send alert", zap.Error(err))
		}
	}
}

// Example 4: Infrastructure Health
func infrastructureHealthExample(alertClient *util.AlertClient, logger *zap.Logger) {
	ctx := context.Background()

	// Check database connection
	dbConnected := false
	if !dbConnected {
		logger.Error("database connection lost")

		// Send critical alert
		if err := alertClient.SendCritical(ctx, "health-check",
			"Database connection lost",
			map[string]any{
				"service":  "postgresql",
				"host":     "localhost",
				"port":     5432,
				"error":    "connection_refused",
				"retry_in": "30s",
			}); err != nil {
			logger.Error("failed to send alert", zap.Error(err))
		}
	}

	// Check RabbitMQ connection
	rabbitmqConnected := false
	if !rabbitmqConnected {
		logger.Error("rabbitmq connection lost")

		// Send critical alert
		if err := alertClient.SendCritical(ctx, "health-check",
			"RabbitMQ connection lost",
			map[string]any{
				"service": "rabbitmq",
				"url":     "amqp://guest:guest@localhost:5672/",
				"error":   "connection_refused",
			}); err != nil {
			logger.Error("failed to send alert", zap.Error(err))
		}
	}
}

// Example 5: Business Logic Alerts
func businessLogicExample(alertClient *util.AlertClient, logger *zap.Logger) {
	ctx := context.Background()

	// Alert on unusual settlement patterns
	averageAmount := 500
	outlierAmount := 50000 // Much higher than average

	if outlierAmount > (averageAmount * 10) {
		logger.Warn("unusual settlement detected",
			zap.Int64("amount", int64(outlierAmount)),
			zap.Int("average", averageAmount))

		// Send warning alert
		if err := alertClient.SendWarning(ctx, "settlement-service",
			"Unusual settlement amount detected",
			map[string]any{
				"amount":         outlierAmount,
				"average_amount": averageAmount,
				"multiple":       outlierAmount / averageAmount,
				"needs_review":   true,
			}); err != nil {
			logger.Error("failed to send alert", zap.Error(err))
		}
	}
}

func main() {
	// Initialize logger
	logger, _ := util.NewLogger("alert-examples")
	defer logger.Sync()

	// Initialize alert client
	webhookURL := os.Getenv("ALERT_WEBHOOK_URL")
	if webhookURL == "" {
		webhookURL = "http://localhost:9000/webhook"
	}
	alertClient := util.NewAlertClient(webhookURL, logger)

	logger.Info("Alert Client Integration Examples",
		zap.String("webhook_url", webhookURL),
	)
	log.Println()
	log.Println("Example 1: DLQ Monitor with Alerts")
	dlqMonitorExample(alertClient, logger)
	log.Println()

	log.Println("Example 2: Settlement Worker with Alerts")
	settlementWorkerExample(alertClient, logger)
	log.Println()

	log.Println("Example 3: API Server with Alerts")
	apiServerExample(alertClient, logger)
	log.Println()

	log.Println("Example 4: Infrastructure Health")
	infrastructureHealthExample(alertClient, logger)
	log.Println()

	log.Println("Example 5: Business Logic Alerts")
	businessLogicExample(alertClient, logger)
	log.Println()

	log.Println("All examples completed. Check http://localhost:9000/alerts to see results.")
}
