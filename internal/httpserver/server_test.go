package httpserver

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

	"gophermart/internal/balance"
	"gophermart/internal/order"
	"gophermart/internal/user"
)

type memUser struct {
	next int64
	m    map[string]*user.User
}

func (m *memUser) CreateUser(ctx context.Context, login, hash string) (int64, error) {
	if m.m == nil {
		m.m = make(map[string]*user.User)
	}
	if _, ok := m.m[login]; ok {
		return 0, errors.New("user exists")
	}
	m.next++
	id := m.next
	m.m[login] = &user.User{ID: id, Login: login, PasswordHash: hash}
	return id, nil
}

func (m *memUser) GetUserByLogin(ctx context.Context, login string) (*user.User, error) {
	u, ok := m.m[login]
	if !ok {
		return nil, sql.ErrNoRows
	}
	return u, nil
}

type memOrder struct {
	owners map[string]int64
}

func (m *memOrder) CreateOrder(ctx context.Context, userID int64, number string) (bool, int64, error) {
	if m.owners == nil {
		m.owners = make(map[string]int64)
	}
	if o, ok := m.owners[number]; ok {
		return false, o, nil
	}
	m.owners[number] = userID
	return true, userID, nil
}

func (m *memOrder) ListOrders(ctx context.Context, userID int64) ([]order.Order, error) {
	return nil, nil
}

type memBal struct{}

func (memBal) GetBalance(ctx context.Context, userID int64) (balance.Balance, error) {
	return balance.Balance{}, nil
}
func (memBal) CreateWithdrawal(ctx context.Context, userID int64, order string, sum float64) error {
	return nil
}
func (memBal) ListWithdrawals(ctx context.Context, userID int64) ([]balance.Withdrawal, error) {
	return nil, nil
}

func TestRouter_RegisterAndOrder(t *testing.T) {
	uR, oR, bR := &memUser{}, &memOrder{}, memBal{}
	uSvc := user.New(uR, "integration-jwt-secret-key-12", time.Hour)
	h := New(uSvc, order.New(oR), balance.New(&bR)).Router()

	reg, _ := json.Marshal(user.RegisterRequest{Login: "x", Password: "y"})
	req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewReader(reg))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatal(rec.Code)
	}
	var tok string
	for _, c := range rec.Result().Cookies() {
		if c.Name == "token" {
			tok = c.Value
			break
		}
	}
	if tok == "" {
		t.Fatal("cookie")
	}

	up := httptest.NewRequest(http.MethodPost, "/api/user/orders", bytes.NewReader([]byte("79927398713")))
	up.Header.Set("Authorization", "Bearer "+tok)
	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, up)
	if rec2.Code != http.StatusAccepted {
		t.Fatal(rec2.Code)
	}
}
