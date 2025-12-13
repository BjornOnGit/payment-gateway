package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/BjornOnGit/payment-gateway/internal/auth"
	"github.com/BjornOnGit/payment-gateway/internal/util"
	"go.uber.org/zap"
)

// OAuthHandler handles dev-only client credentials token exchange.
type OAuthHandler struct {
	oauth  *auth.OAuthServer
	logger *zap.Logger
}

func NewOAuthHandler(oauth *auth.OAuthServer, logger *zap.Logger) *OAuthHandler {
	if logger == nil {
		logger = zap.NewNop() // no-op logger if nil
	}
	return &OAuthHandler{oauth: oauth, logger: logger}
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
}

// Token handles POST /oauth/token with grant_type=client_credentials.
func (h *OAuthHandler) Token(w http.ResponseWriter, r *http.Request) {
	log := util.WithTraceFromContext(r.Context(), h.logger)
	log.Info("received request", zap.String("path", r.URL.Path), zap.String("method", r.Method))

	if h.oauth == nil {
		log.Error("oauth server not configured")
		http.Error(w, "oauth server not configured", http.StatusServiceUnavailable)
		return
	}

	if err := r.ParseForm(); err != nil {
		log.Error("invalid form", zap.Error(err))
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}

	grant := strings.TrimSpace(r.Form.Get("grant_type"))
	clientID := strings.TrimSpace(r.Form.Get("client_id"))
	clientSecret := strings.TrimSpace(r.Form.Get("client_secret"))

	if grant != "client_credentials" {
		log.Error("unsupported grant_type", zap.String("grant_type", grant))
		http.Error(w, "unsupported grant_type", http.StatusBadRequest)
		return
	}
	if clientID == "" || clientSecret == "" {
		log.Error("missing client credentials")
		http.Error(w, "client_id and client_secret required", http.StatusBadRequest)
		return
	}

	token, err := h.oauth.ExchangeClientCredentials(r.Context(), clientID, clientSecret)
	if err != nil {
		log.Error("invalid client credentials", zap.String("client_id", clientID), zap.Error(err))
		http.Error(w, "invalid client credentials", http.StatusUnauthorized)
		return
	}

	log.Info("token issued", zap.String("client_id", clientID))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tokenResponse{
		AccessToken: token,
		TokenType:   "bearer",
		ExpiresIn:   3600,
	})
}
