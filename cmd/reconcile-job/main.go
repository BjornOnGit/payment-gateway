package main

import (
	"context"
	"database/sql"
	"log"
	"os"

	"github.com/BjornOnGit/payment-gateway/internal/util"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

type ReconRecord struct {
	TransactionID string `json:"transaction_id"`
	Expected      int64  `json:"expected_amount"`
	Actual        int64  `json:"actual_amount"`
	Diff          int64  `json:"difference"`
}

// Kept for backward compatibility but using AlertClient now

func main() {
	// Initialize logger
	serviceName := os.Getenv("LOG_SERVICE_NAME")
	if serviceName == "" {
		serviceName = "reconcile-job"
	}
	logger, err := util.NewLogger(serviceName)
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	logger.Info("starting reconciliation job")

	util.LoadEnv()

	dsn := os.Getenv("DB_URL")
	if dsn == "" {
		logger.Fatal("DB_URL not set")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		logger.Fatal("failed to open database", zap.Error(err))
	}
	defer db.Close()

	ctx := context.Background()

	// Initialize alert client
	alertWebhook := os.Getenv("ALERT_WEBHOOK_URL")
	alertClient := util.NewAlertClient(alertWebhook, logger)

	if err := runReconcile(ctx, db, logger, alertClient); err != nil {
		logger.Fatal("reconciliation failed", zap.Error(err))
	}

	logger.Info("reconciliation job completed successfully")
}

func runReconcile(ctx context.Context, db *sql.DB, logger *zap.Logger, alertClient *util.AlertClient) error {
	// Compare transactions.amount vs sum of settlement amounts
	// Looking for completed or failed transactions
	rows, err := db.QueryContext(ctx, `
		SELECT t.id::text, t.amount::bigint, COALESCE(SUM(s.amount), 0)::bigint
		FROM transactions t
		LEFT JOIN settlements s ON s.external_reference = t.id::text
		WHERE t.status IN ('completed', 'failed')
		GROUP BY t.id, t.amount
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	var mismatchCount int
	var totalChecked int

	for rows.Next() {
		totalChecked++
		var id string
		var expected int64
		var actual int64
		if err := rows.Scan(&id, &expected, &actual); err != nil {
			return err
		}
		if expected != actual {
			mismatchCount++
			diff := expected - actual
			logger.Warn("reconciliation mismatch detected",
				zap.String("transaction_id", id),
				zap.Int64("expected", expected),
				zap.Int64("actual", actual),
				zap.Int64("difference", diff),
			)

			if err := insertReconciliationLog(ctx, db, id, expected, actual, diff); err != nil {
				logger.Error("failed to insert reconciliation log", zap.String("transaction_id", id), zap.Error(err))
				return err
			}

			// Send alert for mismatch
			alertDetails := map[string]any{
				"transaction_id":  id,
				"expected_amount": expected,
				"actual_amount":   actual,
				"difference":      diff,
			}
			if err := alertClient.SendWarning(ctx, "reconcile-job",
				"Transaction amount mismatch detected",
				alertDetails); err != nil {
				logger.Warn("failed to send alert", zap.String("transaction_id", id), zap.Error(err))
			}
		}
	}

	logger.Info("reconciliation summary",
		zap.Int("total_checked", totalChecked),
		zap.Int("mismatches", mismatchCount),
	)

	return nil
}

func insertReconciliationLog(ctx context.Context, db *sql.DB, txID string, expected, actual, diff int64) error {
	_, err := db.ExecContext(ctx, `
		INSERT INTO reconciliation_logs (id, entity_type, entity_id, expected_amount, actual_amount, status, notes, created_at)
		VALUES (uuid_generate_v4(), 'transactions', $1, $2, $3, 'mismatch', $4, now())
	`, txID, expected, actual, "automated-check")
	return err
}
