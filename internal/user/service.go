package user

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"gophermart/internal/auth"
)

type Repository interface {
	CreateUser(ctx context.Context, login, passwordHash string) (int64, error)
	GetUserByLogin(ctx context.Context, login string) (*User, error)
}

type Service struct {
	repo        Repository
	jwtSecret   string
	jwtTokenTTL time.Duration
}

func New(repo Repository, secret string, tokenTTL time.Duration) *Service {
	if tokenTTL <= 0 {
		tokenTTL = 24 * time.Hour
	}
	return &Service{repo: repo, jwtSecret: secret, jwtTokenTTL: tokenTTL}
}

func (s *Service) TokenTTL() time.Duration {
	return s.jwtTokenTTL
}

// Регистрация пользователя
func (s *Service) Register(ctx context.Context, login, password string) (string, error) {
	hash, err := auth.HashPassword(password)
	if err != nil {
		return "", err
	}
	id, err := s.repo.CreateUser(ctx, login, hash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrUserExists
		}
		if err.Error() == "user exists" {
			return "", ErrUserExists
		}
		return "", err
	}
	return auth.NewToken(id, s.jwtSecret, s.jwtTokenTTL)
}

// Вход пользователя
func (s *Service) Login(ctx context.Context, login, password string) (string, error) {
	u, err := s.repo.GetUserByLogin(ctx, login)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrInvalidCredentials
		}
		return "", err
	}
	if !auth.VerifyPassword(u.PasswordHash, password) {
		return "", ErrInvalidCredentials
	}
	return auth.NewToken(u.ID, s.jwtSecret, s.jwtTokenTTL)
}

// Парсит токен
func (s *Service) ParseToken(token string) (int64, error) {
	claims, err := auth.ParseToken(token, s.jwtSecret)
	if err != nil {
		return 0, err
	}
	return claims.UserID, nil
}
