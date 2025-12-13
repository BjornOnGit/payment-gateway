package model

import (
	"time"
	"github.com/google/uuid"
)

type AuditLogs struct {
	ID uuid.UUID `json:"id" db:"id"`
	ActorID uuid.UUID `json:"actor_id" db:"actor_id"`
	Action string    `json:"action" db:"action"`
	Entity string    `json:"entity" db:"entity"`
	EntityID uuid.UUID `json:"entity_id" db:"entity_id"`
	Details string    `json:"details" db:"details"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type ReconciliationLogs struct {
	ID uuid.UUID `json:"id" db:"id"`
	EntityType string    `json:"entity_type" db:"entity_type"`
	EntityID uuid.UUID `json:"entity_id" db:"entity_id"`
	ExpectedAmount int64 `json:"expected_amount" db:"expected_amount"`
	ActualAmount int64 `json:"actual_amount" db:"actual_amount"`
	Status string    `json:"status" db:"status"`
	Notes string    `json:"notes" db:"notes"`
	Metadata string    `json:"metadata" db:"metadata"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}