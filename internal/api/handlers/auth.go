package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/BjornOnGit/payment-gateway/internal/service"
	"github.com/BjornOnGit/payment-gateway/internal/util"
	"go.uber.org/zap"
)

type AuthHandler struct {
	svc    *service.AuthService
	logger *zap.Logger
}

func NewAuthHandler(svc *service.AuthService, logger *zap.Logger) *AuthHandler {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &AuthHandler{svc: svc, logger: logger}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	log := util.WithTraceFromContext(r.Context(), h.logger)

	var payload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	user, token, err := h.svc.Register(r.Context(), strings.TrimSpace(payload.Email), payload.Password)
	if err != nil {
		switch err {
		case service.ErrEmailTaken:
			http.Error(w, "email already registered", http.StatusConflict)
			return
		case service.ErrInvalidCredentials:
			http.Error(w, "invalid email or password", http.StatusBadRequest)
			return
		case service.ErrJWTNotConfigured:
			http.Error(w, "jwt not configured", http.StatusServiceUnavailable)
			return
		default:
			log.Error("register failed", zap.Error(err))
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
	}

	resp := map[string]any{
		"user": map[string]any{
			"id":    user.ID,
			"email": user.Email,
		},
		"access_token": token,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	log := util.WithTraceFromContext(r.Context(), h.logger)

	var payload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	user, token, err := h.svc.Login(r.Context(), strings.TrimSpace(payload.Email), payload.Password)
	if err != nil {
		switch err {
		case service.ErrInvalidCredentials:
			http.Error(w, "invalid email or password", http.StatusUnauthorized)
			return
		case service.ErrJWTNotConfigured:
			http.Error(w, "jwt not configured", http.StatusServiceUnavailable)
			return
		default:
			log.Error("login failed", zap.Error(err))
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
	}

	resp := map[string]any{
		"user": map[string]any{
			"id":    user.ID,
			"email": user.Email,
		},
		"access_token": token,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
