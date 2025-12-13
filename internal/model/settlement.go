package model

import (
	"time"

	"github.com/google/uuid"
)

type SettlementStatus string

const (
	SettlementPending SettlementStatus = "pending"
	SettlementSuccess SettlementStatus = "success"
	SettlementFailed SettlementStatus = "failed"
)

type Settlements struct {
	ID uuid.UUID `json:"id" db:"id"`
	MerchantAccountID uuid.UUID `json:"merchant_account_id" db:"merchant_account_id"`
	ExternalReference string `json:"external_reference" db:"external_reference"`
	Status string `json:"status" db:"status"`
	Amount int64 `json:"amount" db:"amount"`
	Metadata map[string]any `json:"metadata" db:"metadata"`
	Attempts int64 `json:"attempts" db:"attempts"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}