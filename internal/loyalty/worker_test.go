package loyalty

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"gophermart/internal/order"
)

type recStore struct {
	orders  []order.Order
	accrual float64
}

func (r *recStore) ClaimOrdersForProcessing(ctx context.Context, limit int) ([]order.Order, error) {
	return r.orders, nil
}

func (r *recStore) SetOrderStatus(ctx context.Context, number string, st order.OrderStatus) error {
	return nil
}

func (r *recStore) ApplyAccrual(ctx context.Context, number string, amount float64) error {
	r.accrual = amount
	return nil
}

func TestWorker_ProcessBatch_Processed(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		a := 12.0
		_ = json.NewEncoder(w).Encode(Response{Order: "79927398713", Status: "PROCESSED", Accrual: &a})
	}))
	defer srv.Close()

	st := &recStore{
		orders: []order.Order{{Number: "79927398713"}},
	}
	w := NewWorker(st, NewClient(srv.URL))
	w.processBatch(context.Background())
	if st.accrual != 12 {
		t.Fatal(st.accrual)
	}
}

func TestWorker_ProcessBatch_Invalid(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		_ = json.NewEncoder(w).Encode(Response{Order: "79927398713", Status: "INVALID"})
	}))
	defer srv.Close()
	st := &recStore{orders: []order.Order{{Number: "79927398713"}}}
	NewWorker(st, NewClient(srv.URL)).processBatch(context.Background())
}

type errClaim struct{}

func (errClaim) ClaimOrdersForProcessing(ctx context.Context, limit int) ([]order.Order, error) {
	return nil, errors.New("claim")
}
func (errClaim) SetOrderStatus(ctx context.Context, number string, st order.OrderStatus) error {
	return nil
}
func (errClaim) ApplyAccrual(ctx context.Context, number string, amount float64) error { return nil }

func TestWorker_ProcessBatch_ClaimError(t *testing.T) {
	NewWorker(errClaim{}, NewClient("http://localhost:1")).processBatch(context.Background())
}

func TestWorker_ProcessBatch_NoContent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()
	st := &recStore{orders: []order.Order{{Number: "79927398713"}}}
	NewWorker(st, NewClient(srv.URL)).processBatch(context.Background())
}

func TestWorker_ProcessBatch_Registered(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(Response{Order: "79927398713", Status: "REGISTERED"})
	}))
	defer srv.Close()
	st := &recStore{orders: []order.Order{{Number: "79927398713"}}}
	NewWorker(st, NewClient(srv.URL)).processBatch(context.Background())
}
