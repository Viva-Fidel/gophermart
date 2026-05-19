package user

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"gophermart/internal/auth"
)

type stubUserRepo struct {
	createFn   func(ctx context.Context, login, hash string) (int64, error)
	getByLogin func(ctx context.Context, login string) (*User, error)
}

func (s *stubUserRepo) CreateUser(ctx context.Context, login, passwordHash string) (int64, error) {
	if s.createFn != nil {
		return s.createFn(ctx, login, passwordHash)
	}
	return 0, errors.New("ni")
}

func (s *stubUserRepo) GetUserByLogin(ctx context.Context, login string) (*User, error) {
	if s.getByLogin != nil {
		return s.getByLogin(ctx, login)
	}
	return nil, errors.New("ni")
}

func TestService_RegisterAndParse(t *testing.T) {
	repo := &stubUserRepo{
		createFn: func(ctx context.Context, login, hash string) (int64, error) {
			return 5, nil
		},
	}
	svc := New(repo, "secret-key-test-123456789", time.Hour)
	tok, err := svc.Register(context.Background(), "u", "p")
	if err != nil {
		t.Fatal(err)
	}
	id, err := svc.ParseToken(tok)
	if err != nil || id != 5 {
		t.Fatalf("id=%d err=%v", id, err)
	}
}

func TestService_Register_ErrNoRows(t *testing.T) {
	repo := &stubUserRepo{
		createFn: func(ctx context.Context, login, hash string) (int64, error) {
			return 0, sql.ErrNoRows
		},
	}
	svc := New(repo, "s", time.Hour)
	_, err := svc.Register(context.Background(), "u", "p")
	if !errors.Is(err, ErrUserExists) {
		t.Fatal(err)
	}
}

func TestService_Register_UserExists(t *testing.T) {
	repo := &stubUserRepo{
		createFn: func(ctx context.Context, login, hash string) (int64, error) {
			return 0, errors.New("user exists")
		},
	}
	svc := New(repo, "s", time.Hour)
	_, err := svc.Register(context.Background(), "u", "p")
	if !errors.Is(err, ErrUserExists) {
		t.Fatal(err)
	}
}

func TestService_Login_OK(t *testing.T) {
	h, _ := RegisterHashForTest(t, "pw")
	repo := &stubUserRepo{
		getByLogin: func(ctx context.Context, login string) (*User, error) {
			return &User{ID: 3, Login: "u", PasswordHash: h}, nil
		},
	}
	svc := New(repo, "secret-key-test-123456789", time.Hour)
	tok, err := svc.Login(context.Background(), "u", "pw")
	if err != nil || tok == "" {
		t.Fatal(err, tok)
	}
}

func TestService_Login_OtherRepoError(t *testing.T) {
	repo := &stubUserRepo{
		getByLogin: func(ctx context.Context, login string) (*User, error) {
			return nil, errors.New("db")
		},
	}
	svc := New(repo, "s", time.Hour)
	_, err := svc.Login(context.Background(), "u", "p")
	if err == nil || errors.Is(err, ErrInvalidCredentials) {
		t.Fatal(err)
	}
}

func TestService_Login_NoUser(t *testing.T) {
	repo := &stubUserRepo{
		getByLogin: func(ctx context.Context, login string) (*User, error) {
			return nil, sql.ErrNoRows
		},
	}
	svc := New(repo, "s", time.Hour)
	_, err := svc.Login(context.Background(), "u", "p")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatal(err)
	}
}

func TestService_New_DefaultTTL(t *testing.T) {
	svc := New(&stubUserRepo{}, "s", 0)
	if svc.TokenTTL() != 24*time.Hour {
		t.Fatal(svc.TokenTTL())
	}
}

func TestService_ParseToken_Error(t *testing.T) {
	svc := New(&stubUserRepo{}, "secret-parse", time.Hour)
	_, err := svc.ParseToken("not-a-jwt")
	if err == nil {
		t.Fatal("want err")
	}
}

// RegisterHashForTest обходит экспорт только для теста: хеш с тем же алгоритмом, что и сервис.
func RegisterHashForTest(t *testing.T, password string) (string, error) {
	t.Helper()
	return auth.HashPassword(password)
}
