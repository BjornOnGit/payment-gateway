package dto

import "github.com/google/uuid"

type CreateTransactionDTO struct {
	Amount     int64          `json:"amount"`
	Currency   string         `json:"currency"`
	UserID     uuid.UUID      `json:"user_id"`
	MerchantID uuid.UUID      `json:"merchant_id"`
	Metadata   map[string]any `json:"metadata"`
}
