package order

import (
	"context"
	"errors"
	"testing"
	"time"
)

type stubOrderRepo struct {
	createFn func(ctx context.Context, userID int64, number string) (bool, int64, error)
	listFn   func(ctx context.Context, userID int64) ([]Order, error)
}

func (s *stubOrderRepo) CreateOrder(ctx context.Context, userID int64, number string) (bool, int64, error) {
	if s.createFn != nil {
		return s.createFn(ctx, userID, number)
	}
	return false, 0, errors.New("not implemented")
}

func (s *stubOrderRepo) ListOrders(ctx context.Context, userID int64) ([]Order, error) {
	if s.listFn != nil {
		return s.listFn(ctx, userID)
	}
	return nil, errors.New("not implemented")
}

func TestService_SubmitOrder_InvalidLuhn(t *testing.T) {
	svc := New(&stubOrderRepo{})
	code, err := svc.SubmitOrder(context.Background(), 1, "123")
	if err != nil {
		t.Fatal(err)
	}
	if code != 422 {
		t.Fatalf("status = %d, want 422", code)
	}
}

func TestService_SubmitOrder_Created(t *testing.T) {
	repo := &stubOrderRepo{
		createFn: func(ctx context.Context, userID int64, number string) (bool, int64, error) {
			return true, userID, nil
		},
	}
	svc := New(repo)
	code, err := svc.SubmitOrder(context.Background(), 5, "79927398713")
	if err != nil {
		t.Fatal(err)
	}
	if code != 202 {
		t.Fatalf("status = %d, want 202", code)
	}
}

func TestService_SubmitOrder_SameOwner(t *testing.T) {
	repo := &stubOrderRepo{
		createFn: func(ctx context.Context, userID int64, number string) (bool, int64, error) {
			return false, userID, nil
		},
	}
	svc := New(repo)
	code, err := svc.SubmitOrder(context.Background(), 5, "79927398713")
	if err != nil {
		t.Fatal(err)
	}
	if code != 200 {
		t.Fatalf("status = %d, want 200", code)
	}
}

func TestService_SubmitOrder_Conflict(t *testing.T) {
	repo := &stubOrderRepo{
		createFn: func(ctx context.Context, userID int64, number string) (bool, int64, error) {
			return false, 99, nil
		},
	}
	svc := New(repo)
	code, err := svc.SubmitOrder(context.Background(), 5, "79927398713")
	if err != nil {
		t.Fatal(err)
	}
	if code != 409 {
		t.Fatalf("status = %d, want 409", code)
	}
}

func TestService_SubmitOrder_RepoError(t *testing.T) {
	repo := &stubOrderRepo{
		createFn: func(ctx context.Context, userID int64, number string) (bool, int64, error) {
			return false, 0, errors.New("db down")
		},
	}
	svc := New(repo)
	code, err := svc.SubmitOrder(context.Background(), 5, "79927398713")
	if err == nil {
		t.Fatal("expected error")
	}
	if code != 500 {
		t.Fatalf("status = %d, want 500", code)
	}
}

func TestService_ListOrders(t *testing.T) {
	want := []Order{{Number: "1", Status: OrderStatusNew, UploadedAt: time.Unix(1, 0)}}
	repo := &stubOrderRepo{
		listFn: func(ctx context.Context, userID int64) ([]Order, error) {
			return want, nil
		},
	}
	svc := New(repo)
	got, err := svc.ListOrders(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Number != "1" {
		t.Fatalf("got %+v", got)
	}
}
