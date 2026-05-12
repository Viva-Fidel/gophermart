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
	resp, code, ra, err := c.GetOrder(context.Background(), "n")
	if err != nil || code != http.StatusOK || ra != 0 || resp == nil {
		t.Fatal(err, code, resp)
	}
}

func TestClient_NoContent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()
	resp, code, ra, err := NewClient(srv.URL).GetOrder(context.Background(), "x")
	if err != nil || code != http.StatusNoContent || resp != nil || ra != 0 {
		t.Fatal(err, code, resp, ra)
	}
}

func TestClient_NonOK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()
	resp, code, _, err := NewClient(srv.URL).GetOrder(context.Background(), "x")
	if err != nil || code != 500 || resp != nil {
		t.Fatal(err, code, resp)
	}
}

func TestClient_RetryAfter_InvalidHeader(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "nope")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()
	_, code, ra, err := NewClient(srv.URL).GetOrder(context.Background(), "x")
	if err != nil || code != 429 || ra != time.Second {
		t.Fatal(err, code, ra)
	}
}

func TestClient_JSONDecodeError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`not json`))
	}))
	defer srv.Close()
	_, _, _, err := NewClient(srv.URL).GetOrder(context.Background(), "x")
	if err == nil {
		t.Fatal("want decode error")
	}
}

func TestClient_RetryAfter(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "2")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()
	_, code, ra, err := NewClient(srv.URL).GetOrder(context.Background(), "x")
	if err != nil || code != http.StatusTooManyRequests || ra != 2*time.Second {
		t.Fatal(err, code, ra)
	}
}
