package service

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/BjornOnGit/payment-gateway/internal/bus"
	"github.com/BjornOnGit/payment-gateway/internal/db/repo"
	"github.com/google/uuid"
)

type RoutingService struct {
	repo repo.TransactionRepository
	bus  bus.Bus
}

func NewRoutingService(r repo.TransactionRepository, b bus.Bus) *RoutingService {
	return &RoutingService{repo: r, bus: b}
}

func (s *RoutingService) ProcessTransaction(ctx context.Context, transaction_id string) error {
	id, err := uuid.Parse(strings.TrimSpace(transaction_id))
	if err != nil {
		return err
	}

	if err := s.repo.UpdateStatus(ctx, id, "processing"); err != nil {
		return err
	}

	routingInfo := map[string]any{
		"route":    "simulated-acquirer",
		"priority": "normal",
	}

	payload := map[string]any{
		"transaction_id": id.String(),
		"routing":        routingInfo,
		"requested_at":   time.Now().UTC(),
	}

	bs, _ := json.Marshal(payload)

	if s.bus != nil {
		if err := s.bus.Publish(ctx, "settlement.requested", id.String(), bs); err != nil {
			log.Printf("[routing] publish settlement.requested failed: %v", err)
		}
	}
	return nil
}
