package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/BjornOnGit/payment-gateway/internal/service"
	"github.com/BjornOnGit/payment-gateway/internal/service/dto"
	"github.com/BjornOnGit/payment-gateway/internal/util"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type TransactionHandler struct {
	svc    *service.TransactionService
	logger *zap.Logger
}

func NewTransactionHandler(s *service.TransactionService, logger *zap.Logger) *TransactionHandler {
	if logger == nil {
		logger = zap.NewNop() // no-op logger if nil
	}
	return &TransactionHandler{svc: s, logger: logger}
}

func (h *TransactionHandler) Create(w http.ResponseWriter, r *http.Request) {
	log := util.WithTraceFromContext(r.Context(), h.logger)
	log.Info("received request", zap.String("path", r.URL.Path), zap.String("method", r.Method))

	var payload dto.CreateTransactionDTO

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		log.Error("invalid json", zap.Error(err))
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	id, err := h.svc.CreateTransaction(r.Context(), payload)
	if err != nil {
		log.Error("failed to create transaction", zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Info("transaction created", zap.String("id", id.String()))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{
		"id": id,
	})
}

func (h *TransactionHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	log := util.WithTraceFromContext(r.Context(), h.logger)
	log.Info("received request", zap.String("path", r.URL.Path), zap.String("method", r.Method))

	idStr := r.PathValue("id")
	if idStr == "" {
		// Fallback: extract last path segment for custom router
		p := r.URL.Path
		if i := strings.LastIndex(p, "/"); i != -1 && i+1 < len(p) {
			idStr = p[i+1:]
		}
	}
	if idStr == "" {
		log.Error("missing transaction id")
		http.Error(w, "missing transaction id", http.StatusBadRequest)
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		log.Error("invalid transaction id", zap.String("id", idStr), zap.Error(err))
		http.Error(w, "invalid transaction id", http.StatusBadRequest)
		return
	}

	tx, err := h.svc.GetTransaction(r.Context(), id)
	if err != nil {
		log.Error("failed to get transaction", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if tx == nil {
		log.Warn("transaction not found", zap.String("id", id.String()))
		http.Error(w, "transaction not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tx)
}

func (h *TransactionHandler) List(w http.ResponseWriter, r *http.Request) {
	log := util.WithTraceFromContext(r.Context(), h.logger)
	log.Info("received request", zap.String("path", r.URL.Path), zap.String("method", r.Method))

	limit := 10
	offset := 0

	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	transactions, err := h.svc.ListTransactions(r.Context(), limit, offset)
	if err != nil {
		log.Error("failed to list transactions", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"data":   transactions,
		"limit":  limit,
		"offset": offset,
	})
}
