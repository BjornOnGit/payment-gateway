package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Simple in-memory client store for dev.
// In production use DB and hash secrets.
type OAuthClient struct {
	ClientID     string
	ClientSecret string
	Name         string
	Scopes       []string
}

type OAuthServer struct {
	clients map[string]*OAuthClient
	mu      sync.RWMutex
	jwt     *JWTManager
}

// NewOAuthServer with injected JWT manager
func NewOAuthServer(jwt *JWTManager) *OAuthServer {
	return &OAuthServer{
		clients: make(map[string]*OAuthClient),
		jwt:     jwt,
	}
}

func (s *OAuthServer) RegisterClient(id, secret, name string, scopes []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clients[id] = &OAuthClient{ClientID: id, ClientSecret: secret, Name: name, Scopes: scopes}
}

func (s *OAuthServer) ValidateClientCreds(clientID, clientSecret string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	c, ok := s.clients[clientID]
	if !ok {
		return false
	}
	return c.ClientSecret == clientSecret
}

// ExchangeClientCredentials issues a JWT for valid client credentials (client_credentials grant)
func (s *OAuthServer) ExchangeClientCredentials(ctx context.Context, clientID, clientSecret string) (string, error) {
	if !s.ValidateClientCreds(clientID, clientSecret) {
		return "", errors.New("invalid client credentials")
	}
	claims := CustomClaims{
		Subject: clientID,
		Scope:   "client",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.jwt.issuer,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	return s.jwt.SignClaims(claims)
}

// Helper to generate random secret for dev
func GenerateClientSecret(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
