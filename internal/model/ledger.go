package model

import (
	"time"
	"github.com/google/uuid"
)

type LedgerEntryType string

const (
	LedgerCredit LedgerEntryType = "credit"
	LedgerDebit LedgerEntryType = "debit"
)

type LedgerEntries struct {
	ID uuid.UUID `json:"id" db:"id"`
	DebitAccountID uuid.UUID `json:"debit_account_id" db:"debit_account_id"`
	CreditAccountID uuid.UUID `json:"credit_account_id" db:"credit_account_id"`
	Amount int64 `json:"amount" db:"amount"`
	Currency  string      `json:"currency" db:"currency"`
	TransactionID uuid.UUID `json:"transaction_id" db:"transaction_id"`
	Description string `json:"description" db:"description"`
	Metadata map[string] any `json:"metadata" db:"metadata"`
	EntryType LedgerEntryType `json:"entry_type" db:"entry_type"`
	CreatedAt time.Time   `json:"created_at" db:"created_at"`
}