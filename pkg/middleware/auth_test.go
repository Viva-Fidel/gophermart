package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type stubParser struct {
	id  int64
	err error
}

func (s stubParser) ParseToken(string) (int64, error) { return s.id, s.err }

func TestAuth_NoToken(t *testing.T) {
	h := Auth(stubParser{id: 1})(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("next")
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatal(rec.Code)
	}
}

func TestAuth_BadToken(t *testing.T) {
	h := Auth(stubParser{err: errors.New("e")})(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("next")
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer x")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatal(rec.Code)
	}
}

func TestAuth_OK(t *testing.T) {
	var uid int64
	h := Auth(stubParser{id: 77})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var ok bool
		uid, ok = UserIDFromContext(r.Context())
		if !ok {
			t.Fatal("ctx")
		}
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer t")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || uid != 77 {
		t.Fatal(rec.Code, uid)
	}
}

func TestAuth_CookieToken(t *testing.T) {
	h := Auth(stubParser{id: 5})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: AuthCookieName, Value: "from-cookie"})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatal(rec.Code)
	}
}

func TestWithUserID(t *testing.T) {
	ctx := WithUserID(context.Background(), 3)
	id, ok := UserIDFromContext(ctx)
	if !ok || id != 3 {
		t.Fatal(id, ok)
	}
}
