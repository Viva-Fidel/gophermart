package balance

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/mock/gomock"

	"gophermart/pkg/middleware"
)

func TestHandler_GetBalance_Unauthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	h := NewHandler(HandlerDeps{Service: New(NewMockRepository(ctrl))})
	rec := httptest.NewRecorder()
	h.GetBalance(rec, httptest.NewRequest(http.MethodGet, "/b", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatal(rec.Code)
	}
}

func TestHandler_GetBalance_OK(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := NewMockRepository(ctrl)
	repo.EXPECT().GetBalance(gomock.Any(), int64(1)).Return(Balance{Current: 10, Withdrawn: 2}, nil)
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
	ctrl := gomock.NewController(t)
	h := NewHandler(HandlerDeps{Service: New(NewMockRepository(ctrl))})
	req := httptest.NewRequest(http.MethodPost, "/w", bytes.NewReader([]byte(`{`)))
	req = req.WithContext(middleware.WithUserID(req.Context(), 1))
	rec := httptest.NewRecorder()
	h.Withdraw(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatal(rec.Code)
	}
}

func TestHandler_Withdraw_500(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := NewMockRepository(ctrl)
	repo.EXPECT().CreateWithdrawal(gomock.Any(), int64(1), "79927398713", float64(1)).Return(errors.New("db"))
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
	ctrl := gomock.NewController(t)
	repo := NewMockRepository(ctrl)
	repo.EXPECT().ListWithdrawals(gomock.Any(), int64(1)).Return(nil, errors.New("db"))
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
	ctrl := gomock.NewController(t)
	repo := NewMockRepository(ctrl)
	repo.EXPECT().CreateWithdrawal(gomock.Any(), int64(1), "79927398713", float64(1)).Return(ErrNotEnoughBalance)
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
	ctrl := gomock.NewController(t)
	repo := NewMockRepository(ctrl)
	repo.EXPECT().ListWithdrawals(gomock.Any(), int64(1)).Return(nil, nil)
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
	ctrl := gomock.NewController(t)
	repo := NewMockRepository(ctrl)
	repo.EXPECT().GetBalance(gomock.Any(), int64(1)).Return(Balance{}, errors.New("db"))
	h := NewHandler(HandlerDeps{Service: New(repo)})
	req := httptest.NewRequest(http.MethodGet, "/b", nil)
	req = req.WithContext(middleware.WithUserID(req.Context(), 1))
	rec := httptest.NewRecorder()
	h.GetBalance(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Fatal(rec.Code)
	}
}
