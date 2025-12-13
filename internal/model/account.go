package model

import (
    "time"

    "github.com/google/uuid"
)

type AccountType string

const (
    AccountTypeUser     AccountType = "user"
    AccountTypeMerchant AccountType = "merchant"
    AccountTypeSystem   AccountType = "system"
)

type Accounts struct {
    ID        uuid.UUID   `json:"id" db:"id"`
    OwnerID   uuid.UUID   `json:"owner_id" db:"owner_id"` // user or merchant id
    AccountType      AccountType `json:"type" db:"type"`
    Currency  string      `json:"currency" db:"currency"`
    CreatedAt time.Time   `json:"created_at" db:"created_at"`
    UpdatedAt time.Time   `json:"updated_at" db:"updated_at"`
}
