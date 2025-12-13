package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/BjornOnGit/payment-gateway/internal/auth"
)

type contextKey string

const (
	ContextKeyUserId contextKey = "user_id"
)

type Authenticator struct {
	JWT *auth.JWTManager
}

func (a *Authenticator) NewAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ah := r.Header.Get("Authorization")
		if ah == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		parts := strings.Fields(ah)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		tokenStr := parts[1]
		claims, err := a.JWT.VerifyToken(tokenStr)
		if err != nil {
			http.Error(w, "unauthorized: "+err.Error(), http.StatusUnauthorized)
			return
		}
		// Add user id to context (claim subject)
		ctx := context.WithValue(r.Context(), ContextKeyUserId, claims.Subject)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(ContextKeyUserId).(string); ok {
		return v
	}
	return ""
}