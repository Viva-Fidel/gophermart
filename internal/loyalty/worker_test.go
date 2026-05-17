package loyalty

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

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
		_ = json.NewEncoder(w).Encode(Response{Order: "79991234567", Status: "INVALID"})
	}))
	defer srv.Close()
	st := &recStore{orders: []order.Order{{Number: "79991234567"}}}
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
	st := &recStore{orders: []order.Order{{Number: "79991234567"}}}
	NewWorker(st, NewClient(srv.URL)).processBatch(context.Background())
}

func TestWorker_ProcessBatch_Parallel(t *testing.T) {
	var inFlight int32
	var peak int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cur := atomic.AddInt32(&inFlight, 1)
		defer atomic.AddInt32(&inFlight, -1)
		for {
			old := atomic.LoadInt32(&peak)
			if cur <= old || atomic.CompareAndSwapInt32(&peak, old, cur) {
				break
			}
		}
		time.Sleep(50 * time.Millisecond)
		_ = json.NewEncoder(w).Encode(Response{Order: r.URL.Path, Status: "PROCESSING"})
	}))
	defer srv.Close()

	orders := make([]order.Order, 8)
	for i := range orders {
		orders[i] = order.Order{Number: fmt.Sprintf("n%d", i)}
	}
	NewWorker(&recStore{orders: orders}, NewClient(srv.URL)).processBatch(context.Background())

	if atomic.LoadInt32(&peak) < 2 {
		t.Fatalf("peak concurrency=%d", peak)
	}
}

func TestWorker_ProcessBatch_429_StopsExtraRequests(t *testing.T) {
	var total atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := total.Add(1)
		if n == 1 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		time.Sleep(10 * time.Millisecond)
		_ = json.NewEncoder(w).Encode(Response{Status: "PROCESSING"})
	}))
	defer srv.Close()

	orders := make([]order.Order, 12)
	for i := range orders {
		orders[i] = order.Order{Number: fmt.Sprintf("n%d", i)}
	}

	start := time.Now()
	NewWorker(&recStore{orders: orders}, NewClient(srv.URL)).processBatch(context.Background())
	elapsed := time.Since(start)

	// 429 + in-flight completions + at most one more request per worker.
	maxRequests := 1 + defaultPoolSize*2
	if total.Load() > int32(maxRequests) {
		t.Fatalf("too many requests: %d, want <= %d", total.Load(), maxRequests)
	}
	if elapsed < time.Second {
		t.Fatalf("batch finished too early: %v", elapsed)
	}
}

func TestWorker_Run_StopsOnCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	w := NewWorker(&recStore{}, NewClient("http://127.0.0.1:1"))

	done := make(chan struct{})
	go func() {
		w.Run(ctx)
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("worker did not stop after context cancel")
	}
}

func TestWorker_ProcessBatch_Registered(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(Response{Order: "79991234567", Status: "REGISTERED"})
	}))
	defer srv.Close()
	st := &recStore{orders: []order.Order{{Number: "79991234567"}}}
	NewWorker(st, NewClient(srv.URL)).processBatch(context.Background())
}
