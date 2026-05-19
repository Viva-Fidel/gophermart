package user

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gophermart/internal/auth"
)

func TestHandler_Register_OK(t *testing.T) {
	repo := &stubUserRepo{
		createFn: func(ctx context.Context, login, hash string) (int64, error) {
			return 1, nil
		},
	}
	svc := New(repo, "handler-jwt-secret-key-12345", time.Hour)
	h := NewHandler(HandlerDeps{Service: svc})
	body, _ := json.Marshal(RegisterRequest{Login: "a", Password: "b"})
	req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.Register(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatal(rec.Code)
	}
}

func TestHandler_Register_Conflict(t *testing.T) {
	repo := &stubUserRepo{
		createFn: func(ctx context.Context, login, hash string) (int64, error) {
			return 0, errors.New("user exists")
		},
	}
	svc := New(repo, "s", time.Hour)
	h := NewHandler(HandlerDeps{Service: svc})
	body, _ := json.Marshal(RegisterRequest{Login: "a", Password: "b"})
	req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.Register(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatal(rec.Code)
	}
}

func TestHandler_Login_Unauthorized(t *testing.T) {
	repo := &stubUserRepo{
		getByLogin: func(ctx context.Context, login string) (*User, error) {
			return nil, sql.ErrNoRows
		},
	}
	svc := New(repo, "s", time.Hour)
	h := NewHandler(HandlerDeps{Service: svc})
	body, _ := json.Marshal(LoginRequest{Login: "a", Password: "b"})
	req := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.Login(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatal(rec.Code)
	}
}

func TestHandler_Register_500(t *testing.T) {
	repo := &stubUserRepo{
		createFn: func(ctx context.Context, login, hash string) (int64, error) {
			return 0, errors.New("db")
		},
	}
	svc := New(repo, "s", time.Hour)
	h := NewHandler(HandlerDeps{Service: svc})
	body, _ := json.Marshal(RegisterRequest{Login: "a", Password: "b"})
	req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.Register(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Fatal(rec.Code)
	}
}

func TestHandler_Login_500(t *testing.T) {
	repo := &stubUserRepo{
		getByLogin: func(ctx context.Context, login string) (*User, error) {
			return nil, errors.New("db")
		},
	}
	svc := New(repo, "s", time.Hour)
	h := NewHandler(HandlerDeps{Service: svc})
	body, _ := json.Marshal(LoginRequest{Login: "a", Password: "b"})
	req := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.Login(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Fatal(rec.Code)
	}
}

func TestHandler_Login_OK(t *testing.T) {
	hash, err := auth.HashPassword("pw")
	if err != nil {
		t.Fatal(err)
	}
	repo := &stubUserRepo{
		getByLogin: func(ctx context.Context, login string) (*User, error) {
			return &User{ID: 2, PasswordHash: hash}, nil
		},
	}
	svc := New(repo, "login-jwt-secret-key-123456", time.Hour)
	h := NewHandler(HandlerDeps{Service: svc})
	body, _ := json.Marshal(LoginRequest{Login: "a", Password: "pw"})
	req := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.Login(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatal(rec.Code)
	}
}
