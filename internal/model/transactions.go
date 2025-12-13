package model

import (
	"time"

	"github.com/google/uuid"
)

// TransactionStatus represents the state of a transaction.
type TransactionStatus string

const (
	TransactionStatusPending   TransactionStatus = "pending"
	TransactionStatusCompleted TransactionStatus = "completed"
	TransactionStatusFailed    TransactionStatus = "failed"
	TransactionStatusCancelled TransactionStatus = "cancelled"
)

// Transaction is the domain model for a payment transaction.
type Transaction struct {
	ID         uuid.UUID         `json:"id" db:"id"`
	Amount     int64             `json:"amount" db:"amount"` // stored in smallest currency unit (e.g., cents)
	Currency   string            `json:"currency" db:"currency"`
	UserID     uuid.UUID         `json:"user_id" db:"user_id"`
	MerchantID uuid.UUID         `json:"merchant_id" db:"merchant_id"`
	Status     TransactionStatus `json:"status" db:"status"`
	Metadata   map[string]any    `json:"metadata" db:"metadata"`
	CreatedAt  time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at" db:"updated_at"`
}

func NewTransaction(amount int64, currency string, userID, merchantID uuid.UUID, status TransactionStatus) *Transaction {
	return &Transaction{
		ID:         uuid.New(),
		Amount:     amount,
		Currency:   currency,
		UserID:     userID,
		MerchantID: merchantID,
		Status:     TransactionStatusPending,
		Metadata:   map[string]any{},
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}
