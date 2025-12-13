package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/BjornOnGit/payment-gateway/internal/service"
	"github.com/BjornOnGit/payment-gateway/internal/util"
	"go.uber.org/zap"
)

type SettlementHandler struct {
	svc    *service.SettlementService
	logger *zap.Logger
}

func NewSettlementHandler(s *service.SettlementService, logger *zap.Logger) *SettlementHandler {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &SettlementHandler{svc: s, logger: logger}
}

func (h *SettlementHandler) GetByID(w http.ResponseWriter, r *http.Request) {
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
		log.Error("missing settlement id")
		http.Error(w, "missing settlement id", http.StatusBadRequest)
		return
	}

	settlement, err := h.svc.GetSettlement(r.Context(), idStr)
	if err != nil {
		log.Error("failed to get settlement", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if settlement == nil {
		log.Warn("settlement not found", zap.String("id", idStr))
		http.Error(w, "settlement not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(settlement)
}

func (h *SettlementHandler) List(w http.ResponseWriter, r *http.Request) {
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

	settlements, err := h.svc.ListSettlements(r.Context(), limit, offset)
	if err != nil {
		log.Error("failed to list settlements", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"data":   settlements,
		"limit":  limit,
		"offset": offset,
	})
}
