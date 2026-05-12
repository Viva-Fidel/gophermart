package balance

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"gophermart/pkg/middleware"
)

func TestHandler_GetBalance_Unauthorized(t *testing.T) {
	h := NewHandler(HandlerDeps{Service: New(&stubBalanceRepo{})})
	rec := httptest.NewRecorder()
	h.GetBalance(rec, httptest.NewRequest(http.MethodGet, "/b", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatal(rec.Code)
	}
}

func TestHandler_GetBalance_OK(t *testing.T) {
	repo := &stubBalanceRepo{
		getFn: func(ctx context.Context, userID int64) (Balance, error) {
			return Balance{Current: 10, Withdrawn: 2}, nil
		},
	}
	h := NewHandler(HandlerDeps{Service: New(repo)})
	req := httptest.NewRequest(http.MethodGet, "/b", nil)
	req = req.WithContext(middleware.WithUserID(req.Context(), 1))
	rec := httptest.NewRecorder()
	h.GetBalance(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatal(rec.Code)
	}
}

func TestHandler_Withdraw_BadJSON(t *testing.T) {
	h := NewHandler(HandlerDeps{Service: New(&stubBalanceRepo{})})
	req := httptest.NewRequest(http.MethodPost, "/w", bytes.NewReader([]byte(`{`)))
	req = req.WithContext(middleware.WithUserID(req.Context(), 1))
	rec := httptest.NewRecorder()
	h.Withdraw(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatal(rec.Code)
	}
}

func TestHandler_Withdraw_500(t *testing.T) {
	repo := &stubBalanceRepo{
		withdrawFn: func(ctx context.Context, userID int64, order string, sum float64) error {
			return errors.New("db")
		},
	}
	h := NewHandler(HandlerDeps{Service: New(repo)})
	body, _ := json.Marshal(WithdrawRequest{Order: "79927398713", Sum: 1})
	req := httptest.NewRequest(http.MethodPost, "/w", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(middleware.WithUserID(req.Context(), 1))
	rec := httptest.NewRecorder()
	h.Withdraw(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Fatal(rec.Code)
	}
}

func TestHandler_ListWithdrawals_500(t *testing.T) {
	repo := &stubBalanceRepo{
		listFn: func(ctx context.Context, userID int64) ([]Withdrawal, error) {
			return nil, errors.New("db")
		},
	}
	h := NewHandler(HandlerDeps{Service: New(repo)})
	req := httptest.NewRequest(http.MethodGet, "/w", nil)
	req = req.WithContext(middleware.WithUserID(req.Context(), 1))
	rec := httptest.NewRecorder()
	h.ListWithdrawals(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Fatal(rec.Code)
	}
}

func TestHandler_Withdraw_402(t *testing.T) {
	repo := &stubBalanceRepo{
		withdrawFn: func(ctx context.Context, userID int64, order string, sum float64) error {
			return ErrNotEnoughBalance
		},
	}
	h := NewHandler(HandlerDeps{Service: New(repo)})
	body, _ := json.Marshal(WithdrawRequest{Order: "79927398713", Sum: 1})
	req := httptest.NewRequest(http.MethodPost, "/w", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(middleware.WithUserID(req.Context(), 1))
	rec := httptest.NewRecorder()
	h.Withdraw(rec, req)
	if rec.Code != http.StatusPaymentRequired {
		t.Fatal(rec.Code)
	}
}

func TestHandler_List_NoContent(t *testing.T) {
	repo := &stubBalanceRepo{
		listFn: func(ctx context.Context, userID int64) ([]Withdrawal, error) {
			return nil, nil
		},
	}
	h := NewHandler(HandlerDeps{Service: New(repo)})
	req := httptest.NewRequest(http.MethodGet, "/w", nil)
	req = req.WithContext(middleware.WithUserID(req.Context(), 1))
	rec := httptest.NewRecorder()
	h.ListWithdrawals(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatal(rec.Code)
	}
}

func TestHandler_GetBalance_500(t *testing.T) {
	repo := &stubBalanceRepo{
		getFn: func(ctx context.Context, userID int64) (Balance, error) {
			return Balance{}, errors.New("db")
		},
	}
	h := NewHandler(HandlerDeps{Service: New(repo)})
	req := httptest.NewRequest(http.MethodGet, "/b", nil)
	req = req.WithContext(middleware.WithUserID(req.Context(), 1))
	rec := httptest.NewRecorder()
	h.GetBalance(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Fatal(rec.Code)
	}
}
