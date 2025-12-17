package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/BjornOnGit/payment-gateway/internal/auth"
	"github.com/BjornOnGit/payment-gateway/internal/db/repo"
	"github.com/BjornOnGit/payment-gateway/internal/model"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	users      repo.UserRepository
	jwtManager *auth.JWTManager
}

func NewAuthService(users repo.UserRepository, jwtManager *auth.JWTManager) *AuthService {
	return &AuthService{users: users, jwtManager: jwtManager}
}

func (s *AuthService) Register(ctx context.Context, email, password string) (*model.User, string, error) {
	// Basic validation
	if email == "" || password == "" {
		return nil, "", ErrInvalidCredentials
	}
	if !strings.Contains(email, "@") {
		return nil, "", ErrInvalidCredentials
	}

	existing, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		return nil, "", err
	}
	if existing != nil {
		return nil, "", ErrEmailTaken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", err
	}

	now := time.Now()
	user := &model.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: string(hash),
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.users.Create(ctx, user); err != nil {
		return nil, "", err
	}

	token, err := s.signUserToken(user)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*model.User, string, error) {
	user, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		return nil, "", err
	}
	if user == nil {
		return nil, "", ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, "", ErrInvalidCredentials
	}

	if err := s.users.UpdateLastLogin(ctx, user.ID); err != nil {
		return nil, "", err
	}

	token, err := s.signUserToken(user)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

func (s *AuthService) signUserToken(user *model.User) (string, error) {
	if s.jwtManager == nil {
		return "", ErrJWTNotConfigured
	}
	claims := s.jwtManager.BuildClaims(user.ID.String(), "user")
	return s.jwtManager.SignClaims(claims)
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailTaken         = errors.New("email already registered")
	ErrJWTNotConfigured   = errors.New("jwt manager not configured")
)
