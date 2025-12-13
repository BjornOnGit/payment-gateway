package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/BjornOnGit/payment-gateway/internal/bus"
	"github.com/BjornOnGit/payment-gateway/internal/db/repo"
	"github.com/BjornOnGit/payment-gateway/internal/model"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type SettlementService struct {
	txRepo         repo.TransactionRepository
	settlementRepo repo.SettlementRepository
	accountRepo    repo.AccountRepository
	bus            bus.Bus
	logger         *zap.Logger
}

func NewSettlementService(
	txRepo repo.TransactionRepository,
	settlementRepo repo.SettlementRepository,
	accountRepo repo.AccountRepository,
	b bus.Bus,
	logger *zap.Logger,
) *SettlementService {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &SettlementService{
		txRepo:         txRepo,
		settlementRepo: settlementRepo,
		accountRepo:    accountRepo,
		bus:            b,
		logger:         logger,
	}
}

type SettlementPayload struct {
	TransactionID uuid.UUID      `json:"transaction_id"`
	Routing       map[string]any `json:"routing"`
	RequestedAt   time.Time      `json:"requested_at"`
}

func (s *SettlementService) ProcessSettlement(ctx context.Context, payload []byte) error {
	var sp SettlementPayload
	if err := json.Unmarshal(payload, &sp); err != nil {
		s.logger.Error("failed to unmarshal settlement payload", zap.Error(err))
		return fmt.Errorf("invalid settlement payload: %w", err)
	}

	// Extract retry count from context (set by RabbitMQ bus from x-death header)
	retryCount := 0
	if rc, ok := ctx.Value("retry_count").(int); ok {
		retryCount = rc
	}

	s.logger.Info("processing settlement",
		zap.String("transaction_id", sp.TransactionID.String()),
		zap.Int("retry_count", retryCount),
	)

	// Check max retry limit
	const maxRetries = 3
	if retryCount >= maxRetries {
		s.logger.Error("settlement exhausted max retries",
			zap.String("transaction_id", sp.TransactionID.String()),
			zap.Int("retry_count", retryCount),
		)
		// Publish to DLQ
		if s.bus != nil {
			if err := s.bus.Publish(context.Background(), "dlq.settlement.requested", sp.TransactionID.String(), payload); err != nil {
				s.logger.Error("failed to publish to DLQ", zap.Error(err))
			} else {
				s.logger.Info("published to DLQ", zap.String("transaction_id", sp.TransactionID.String()))
			}
		}
		return fmt.Errorf("settlement failed permanently after %d retries", maxRetries)
	}

	// Fetch the transaction
	tx, err := s.txRepo.GetByID(ctx, sp.TransactionID)
	if err != nil {
		s.logger.Error("failed to fetch transaction", zap.String("transaction_id", sp.TransactionID.String()), zap.Error(err))
		// Transient error - trigger retry
		return fmt.Errorf("failed to fetch transaction: %w", err)
	}

	// Get or create merchant account
	merchantAccount, err := s.accountRepo.GetOrCreateMerchantAccount(ctx, tx.MerchantID, tx.Currency)
	if err != nil {
		s.logger.Error("failed to get/create merchant account",
			zap.String("merchant_id", tx.MerchantID.String()),
			zap.Error(err),
		)
		return fmt.Errorf("failed to get/create merchant account: %w", err)
	}

	// Create settlement attempt
	settlement := &model.Settlements{
		ID:                uuid.New(),
		MerchantAccountID: merchantAccount.ID,
		ExternalReference: sp.TransactionID.String(),
		Status:            string(model.SettlementPending),
		Amount:            tx.Amount,
		Metadata: map[string]any{
			"routing":      sp.Routing,
			"requested_at": sp.RequestedAt,
		},
		Attempts:  1,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	if err := s.settlementRepo.CreateSettlementAttempt(ctx, settlement); err != nil {
		s.logger.Error("failed to create settlement attempt", zap.String("settlement_id", settlement.ID.String()), zap.Error(err))
		return fmt.Errorf("failed to create settlement attempt: %w", err)
	}

	s.logger.Info("settlement attempt created", zap.String("settlement_id", settlement.ID.String()))

	// Simulate settlement processing (in real world, this would call external APIs)
	// For now, mark as success and update transaction status
	success := s.simulateSettlementProcessing(ctx, settlement)

	if success {
		// Update transaction status to completed
		if err := s.txRepo.UpdateStatus(ctx, tx.ID, "completed"); err != nil {
			s.logger.Error("failed to update transaction status", zap.String("transaction_id", tx.ID.String()), zap.Error(err))
			return err
		}

		s.logger.Info("settlement successful", zap.String("transaction_id", tx.ID.String()), zap.String("settlement_id", settlement.ID.String()))

		// Publish settlement.completed event
		if s.bus != nil {
			completedPayload := map[string]any{
				"transaction_id": tx.ID.String(),
				"settlement_id":  settlement.ID.String(),
				"status":         "completed",
				"completed_at":   time.Now().UTC(),
			}
			bs, _ := json.Marshal(completedPayload)
			if err := s.bus.Publish(ctx, "settlement.completed", tx.ID.String(), bs); err != nil {
				s.logger.Warn("failed to publish settlement.completed event", zap.Error(err))
			}
		}
	} else {
		// Update transaction status to failed
		if err := s.txRepo.UpdateStatus(ctx, tx.ID, "failed"); err != nil {
			s.logger.Error("failed to update transaction status to failed", zap.String("transaction_id", tx.ID.String()), zap.Error(err))
			return err
		}

		s.logger.Warn("settlement failed", zap.String("transaction_id", tx.ID.String()), zap.String("settlement_id", settlement.ID.String()))

		// Publish settlement.failed event
		if s.bus != nil {
			failedPayload := map[string]any{
				"transaction_id": tx.ID.String(),
				"settlement_id":  settlement.ID.String(),
				"status":         "failed",
				"failed_at":      time.Now().UTC(),
			}
			bs, _ := json.Marshal(failedPayload)
			if err := s.bus.Publish(ctx, "settlement.failed", tx.ID.String(), bs); err != nil {
				s.logger.Warn("failed to publish settlement.failed event", zap.Error(err))
			}
		}
	}

	return nil
}

// simulateSettlementProcessing simulates calling external settlement APIs
// In a real system, this would make HTTP calls to payment processors
func (s *SettlementService) simulateSettlementProcessing(ctx context.Context, settlement *model.Settlements) bool {
	// Simulate realistic success rate with occasional failures
	// In production, this would actually call external APIs
	s.logger.Info("simulating settlement API call", zap.String("settlement_id", settlement.ID.String()))

	// Simulate 90% success rate (10% failure for testing DLQ)
	// In real world: make HTTP calls, handle retries, timeouts, etc.
	// For testing: uncomment next line to force failures
	// return false
	return true // Change to false to test DLQ flow
}

func (s *SettlementService) GetSettlement(ctx context.Context, id string) (*model.Settlements, error) {
	return s.settlementRepo.GetByID(ctx, id)
}

func (s *SettlementService) ListSettlements(ctx context.Context, limit int, offset int) ([]*model.Settlements, error) {
	return s.settlementRepo.List(ctx, limit, offset)
}
