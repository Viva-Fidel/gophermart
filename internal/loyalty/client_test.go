package loyalty

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClient_GetOrder_OK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/orders/n" {
			t.Fatal(r.URL.Path)
		}
		a := 1.0
		_ = json.NewEncoder(w).Encode(Response{Order: "n", Status: "NEW", Accrual: &a})
	}))
	defer srv.Close()
	c := NewClient(srv.URL)
	result, err := c.GetOrder(context.Background(), "n")
	if err != nil || result.StatusCode != http.StatusOK || result.RetryAfter != 0 || result.Body == nil {
		t.Fatal(err, result)
	}
}

func TestClient_NoContent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()
	result, err := NewClient(srv.URL).GetOrder(context.Background(), "x")
	if err != nil || result.StatusCode != http.StatusNoContent || result.Body != nil || result.RetryAfter != 0 {
		t.Fatal(err, result)
	}
}

func TestClient_NonOK(t *testing.T) {
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()
	_, err := NewClient(srv.URL).GetOrder(context.Background(), "x")
	if err == nil {
		t.Fatal("want error after retries exhausted")
	}
	if calls != 4 {
		t.Fatalf("calls=%d", calls)
	}
}

func TestClient_RetryAfter_InvalidHeader(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "nope")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()
	result, err := NewClient(srv.URL).GetOrder(context.Background(), "x")
	if err != nil || result.StatusCode != 429 || result.RetryAfter != time.Second {
		t.Fatal(err, result)
	}
}

func TestClient_JSONDecodeError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`not json`))
	}))
	defer srv.Close()
	_, err := NewClient(srv.URL).GetOrder(context.Background(), "x")
	if err == nil {
		t.Fatal("want decode error")
	}
}

func TestClient_RetriesOnServerError(t *testing.T) {
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		_ = json.NewEncoder(w).Encode(Response{Order: "x", Status: "NEW"})
	}))
	defer srv.Close()

	result, err := NewClient(srv.URL).GetOrder(context.Background(), "x")
	if err != nil || calls != 2 || result.StatusCode != http.StatusOK || result.Body == nil {
		t.Fatalf("err=%v calls=%d result=%+v", err, calls, result)
	}
}

func TestClient_RetryAfter(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "2")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()
	result, err := NewClient(srv.URL).GetOrder(context.Background(), "x")
	if err != nil || result.StatusCode != http.StatusTooManyRequests || result.RetryAfter != 2*time.Second {
		t.Fatal(err, result)
	}
}
