package order

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"gophermart/pkg/middleware"
)

func TestHandler_Upload_Unauthorized(t *testing.T) {
	h := NewHandler(HandlerDeps{Service: New(&stubOrderRepo{})})
	req := httptest.NewRequest(http.MethodPost, "/api/user/orders", strings.NewReader("79927398713"))
	rec := httptest.NewRecorder()
	h.UploadOrder(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatal(rec.Code)
	}
}

func TestHandler_Upload_Accepted(t *testing.T) {
	repo := &stubOrderRepo{
		createFn: func(ctx context.Context, userID int64, number string) (bool, int64, error) {
			return true, userID, nil
		},
	}
	h := NewHandler(HandlerDeps{Service: New(repo)})
	req := httptest.NewRequest(http.MethodPost, "/api/user/orders", strings.NewReader("79927398713"))
	req = req.WithContext(middleware.WithUserID(req.Context(), 1))
	rec := httptest.NewRecorder()
	h.UploadOrder(rec, req)
	if rec.Code != http.StatusAccepted {
		t.Fatal(rec.Code)
	}
}

func TestHandler_List_NoContent(t *testing.T) {
	repo := &stubOrderRepo{
		listFn: func(ctx context.Context, userID int64) ([]Order, error) {
			return nil, nil
		},
	}
	h := NewHandler(HandlerDeps{Service: New(repo)})
	req := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
	req = req.WithContext(middleware.WithUserID(req.Context(), 1))
	rec := httptest.NewRecorder()
	h.ListOrders(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatal(rec.Code)
	}
}

func TestHandler_List_500(t *testing.T) {
	repo := &stubOrderRepo{
		listFn: func(ctx context.Context, userID int64) ([]Order, error) {
			return nil, errors.New("db")
		},
	}
	h := NewHandler(HandlerDeps{Service: New(repo)})
	req := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
	req = req.WithContext(middleware.WithUserID(req.Context(), 1))
	rec := httptest.NewRecorder()
	h.ListOrders(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Fatal(rec.Code)
	}
}

func TestHandler_List_JSON(t *testing.T) {
	ts := time.Unix(10, 0).UTC()
	repo := &stubOrderRepo{
		listFn: func(ctx context.Context, userID int64) ([]Order, error) {
			return []Order{{Number: "79927398713", Status: OrderStatusNew, UploadedAt: ts}}, nil
		},
	}
	h := NewHandler(HandlerDeps{Service: New(repo)})
	req := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
	req = req.WithContext(middleware.WithUserID(req.Context(), 1))
	rec := httptest.NewRecorder()
	h.ListOrders(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatal(rec.Code)
	}
	var raw []map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&raw); err != nil {
		t.Fatal(err)
	}
	if len(raw) != 1 {
		t.Fatal(len(raw))
	}
}

type errReader struct{}

func (errReader) Read([]byte) (int, error) { return 0, errors.New("x") }

func TestHandler_Upload_500(t *testing.T) {
	repo := &stubOrderRepo{
		createFn: func(ctx context.Context, userID int64, number string) (bool, int64, error) {
			return false, 0, errors.New("db")
		},
	}
	h := NewHandler(HandlerDeps{Service: New(repo)})
	req := httptest.NewRequest(http.MethodPost, "/api/user/orders", strings.NewReader("79927398713"))
	req = req.WithContext(middleware.WithUserID(req.Context(), 1))
	rec := httptest.NewRecorder()
	h.UploadOrder(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Fatal(rec.Code)
	}
}

func TestHandler_Upload_BadBody(t *testing.T) {
	h := NewHandler(HandlerDeps{Service: New(&stubOrderRepo{})})
	req := httptest.NewRequest(http.MethodPost, "/api/user/orders", errReader{})
	req = req.WithContext(middleware.WithUserID(req.Context(), 1))
	rec := httptest.NewRecorder()
	h.UploadOrder(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatal(rec.Code)
	}
}
