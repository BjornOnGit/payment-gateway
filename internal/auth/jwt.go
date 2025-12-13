package auth

import (
	"crypto/rsa"
	"errors"
	"io/ioutil"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type CustomClaims struct {
	Subject string `json:"sub,omitempty"`
	Scope   string `json:"scope,omitempty"`
	jwt.RegisteredClaims
}

type JWTManager struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	issuer     string
	expiry     time.Duration
}

func NewJWTManager(privateKeyPath, publicKeyPath, issuer string, expiry time.Duration) (*JWTManager, error) {
	privBytes, err := ioutil.ReadFile(privateKeyPath)
	if err != nil {
		return nil, err
	}
	pubBytes, err := ioutil.ReadFile(publicKeyPath)
	if err != nil {
		return nil, err
	}

	priv, err := jwt.ParseRSAPrivateKeyFromPEM(privBytes)
	if err != nil {
		return nil, err
	}
	pub, err := jwt.ParseRSAPublicKeyFromPEM(pubBytes)
	if err != nil {
		return nil, err
	}

	return &JWTManager{
		privateKey: priv,
		publicKey:  pub,
		issuer:     issuer,
		expiry:     expiry,
	}, nil
}

func (m *JWTManager) VerifyToken(tokenStr string) (*CustomClaims, error) {
	parser := jwt.NewParser(jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Alg()}))
	claims := &CustomClaims{}
	_, err := parser.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (any, error) {
		return m.publicKey, nil
	})
	if err != nil {
		return nil, err
	}

	if claims.Issuer != m.issuer {
		return nil, errors.New("invalid issuer")
	}
	return claims, nil
}

func (m *JWTManager) SignClaims(claims CustomClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(m.privateKey)
}
