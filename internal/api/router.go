package api

import (
	"net/http"
	"strings"

	"github.com/BjornOnGit/payment-gateway/internal/api/handlers"
	"github.com/BjornOnGit/payment-gateway/internal/api/middleware"
	"github.com/BjornOnGit/payment-gateway/internal/auth"
	"github.com/BjornOnGit/payment-gateway/internal/service"
	"go.uber.org/zap"
)

type RouterConfig struct {
	TxService        *service.TransactionService
	SettlementSvc    *service.SettlementService   // optional - if nil, settlement endpoints disabled
	AuthService      *service.AuthService         // optional - if nil, auth endpoints disabled
	JWTManager       *auth.JWTManager             // optional - if nil, no auth middleware
	IdempotencyStore *middleware.IdempotencyStore // optional - if nil, no idempotency middleware
	OAuthServer      *auth.OAuthServer            // optional - if nil, /oauth/token disabled
	Logger           *zap.Logger                  // optional - if nil, handlers use no-op logger
}

func NewRouter(txService *service.TransactionService) http.Handler {
	return NewRouterWithConfig(RouterConfig{TxService: txService})
}

type apiRouter struct {
	cfg       RouterConfig
	txHandler *handlers.TransactionHandler
	sHandler  *handlers.SettlementHandler
	oauthH    *handlers.OAuthHandler
	authH     *handlers.AuthHandler
}

func (ar *apiRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// Public routes
	if path == "/health" {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
		return
	}

	if path == "/oauth/token" && ar.cfg.OAuthServer != nil {
		ar.oauthH.Token(w, r)
		return
	}

	if ar.authH != nil {
		if r.Method == http.MethodPost && path == "/auth/register" {
			ar.authH.Register(w, r)
			return
		}
		if r.Method == http.MethodPost && path == "/auth/login" {
			ar.authH.Login(w, r)
			return
		}
	}

	// Transaction routes
	if strings.HasPrefix(path, "/v1/transactions") {
		if r.Method == http.MethodPost && path == "/v1/transactions" {
			ar.serveCreateTransaction(w, r)
			return
		}
		if r.Method == http.MethodGet && path == "/v1/transactions/list" {
			ar.serveListTransactions(w, r)
			return
		}
		if r.Method == http.MethodGet && strings.HasPrefix(path, "/v1/transactions/") && path != "/v1/transactions/list" {
			ar.serveGetTransaction(w, r)
			return
		}
	}

	// Settlement routes
	if strings.HasPrefix(path, "/v1/settlements") && ar.cfg.SettlementSvc != nil {
		if r.Method == http.MethodGet && path == "/v1/settlements/list" {
			ar.serveListSettlements(w, r)
			return
		}
		if r.Method == http.MethodGet && strings.HasPrefix(path, "/v1/settlements/") && path != "/v1/settlements/list" {
			ar.serveGetSettlement(w, r)
			return
		}
	}

	http.NotFound(w, r)
}

func (ar *apiRouter) serveCreateTransaction(w http.ResponseWriter, r *http.Request) {
	var createTxHandler http.Handler = http.HandlerFunc(ar.txHandler.Create)

	if ar.cfg.IdempotencyStore != nil {
		createTxHandler = ar.cfg.IdempotencyStore.NewIdempotencyMiddleware(createTxHandler)
	}

	if ar.cfg.JWTManager != nil {
		authn := &middleware.Authenticator{JWT: ar.cfg.JWTManager}
		createTxHandler = authn.NewAuthMiddleware(createTxHandler)
	}

	createTxHandler.ServeHTTP(w, r)
}

func (ar *apiRouter) serveListTransactions(w http.ResponseWriter, r *http.Request) {
	var handler http.Handler = http.HandlerFunc(ar.txHandler.List)

	if ar.cfg.JWTManager != nil {
		authn := &middleware.Authenticator{JWT: ar.cfg.JWTManager}
		handler = authn.NewAuthMiddleware(handler)
	}

	handler.ServeHTTP(w, r)
}

func (ar *apiRouter) serveGetTransaction(w http.ResponseWriter, r *http.Request) {
	var handler http.Handler = http.HandlerFunc(ar.txHandler.GetByID)

	if ar.cfg.JWTManager != nil {
		authn := &middleware.Authenticator{JWT: ar.cfg.JWTManager}
		handler = authn.NewAuthMiddleware(handler)
	}

	handler.ServeHTTP(w, r)
}

func (ar *apiRouter) serveListSettlements(w http.ResponseWriter, r *http.Request) {
	var handler http.Handler = http.HandlerFunc(ar.sHandler.List)

	if ar.cfg.JWTManager != nil {
		authn := &middleware.Authenticator{JWT: ar.cfg.JWTManager}
		handler = authn.NewAuthMiddleware(handler)
	}

	handler.ServeHTTP(w, r)
}

func (ar *apiRouter) serveGetSettlement(w http.ResponseWriter, r *http.Request) {
	var handler http.Handler = http.HandlerFunc(ar.sHandler.GetByID)

	if ar.cfg.JWTManager != nil {
		authn := &middleware.Authenticator{JWT: ar.cfg.JWTManager}
		handler = authn.NewAuthMiddleware(handler)
	}

	handler.ServeHTTP(w, r)
}

func NewRouterWithConfig(cfg RouterConfig) http.Handler {
	txHandler := handlers.NewTransactionHandler(cfg.TxService, cfg.Logger)
	var sHandler *handlers.SettlementHandler
	if cfg.SettlementSvc != nil {
		sHandler = handlers.NewSettlementHandler(cfg.SettlementSvc, cfg.Logger)
	}

	oauthHandler := handlers.NewOAuthHandler(cfg.OAuthServer, cfg.Logger)
	var authHandler *handlers.AuthHandler
	if cfg.AuthService != nil {
		authHandler = handlers.NewAuthHandler(cfg.AuthService, cfg.Logger)
	}

	return &apiRouter{
		cfg:       cfg,
		txHandler: txHandler,
		sHandler:  sHandler,
		oauthH:    oauthHandler,
		authH:     authHandler,
	}
}
