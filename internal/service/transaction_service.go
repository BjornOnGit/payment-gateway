package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/BjornOnGit/payment-gateway/internal/bus"
	"github.com/BjornOnGit/payment-gateway/internal/db/repo"
	"github.com/BjornOnGit/payment-gateway/internal/model"
	"github.com/BjornOnGit/payment-gateway/internal/service/dto"
	"github.com/google/uuid"
)

var (
	ErrInvalidAmount   = errors.New("invalid amount")
	ErrInvalidCurrency = errors.New("unsupported currency")
)

var allowedCurrencies = map[string]bool{
	"NGN": true,
}

type TransactionService struct {
	repo repo.TransactionRepository
	bus  bus.Bus
}

func NewTransactionService(r repo.TransactionRepository, b bus.Bus) *TransactionService {
	return &TransactionService{repo: r, bus: b}
}

func (s *TransactionService) CreateTransaction(ctx context.Context,
	input dto.CreateTransactionDTO) (uuid.UUID, error) {

	// ---- Validation ----
	if input.Amount <= 0 {
		return uuid.Nil, ErrInvalidAmount
	}

	if !allowedCurrencies[strings.ToUpper(input.Currency)] {
		return uuid.Nil, ErrInvalidCurrency
	}

	// Initialize metadata if nil
	metadata := input.Metadata
	if metadata == nil {
		metadata = make(map[string]any)
	}

	tx := &model.Transaction{
		ID:         uuid.New(),
		Amount:     input.Amount,
		Currency:   strings.ToUpper(input.Currency),
		UserID:     input.UserID,
		MerchantID: input.MerchantID,
		Status:     model.TransactionStatusPending,
		Metadata:   metadata,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}

	// ---- Persist ----
	if err := s.repo.CreateTransaction(ctx, tx); err != nil {
		return uuid.Nil, err
	}

	// Publish event if bus is configured
	if s.bus != nil {
		go func() {
			payload, _ := json.Marshal(tx)
			_ = s.bus.Publish(context.Background(), "transaction.created", tx.ID.String(), payload)
		}()
	}

	return tx.ID, nil
}

func (s *TransactionService) GetTransaction(ctx context.Context, id uuid.UUID) (*model.Transaction, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *TransactionService) ListTransactions(ctx context.Context, limit int, offset int) ([]*model.Transaction, error) {
	return s.repo.List(ctx, limit, offset)
}
