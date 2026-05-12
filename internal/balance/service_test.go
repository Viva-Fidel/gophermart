package balance

import (
	"context"
	"errors"
	"testing"
)

type stubBalanceRepo struct {
	getFn     func(ctx context.Context, userID int64) (Balance, error)
	withdrawFn func(ctx context.Context, userID int64, order string, sum float64) error
	listFn    func(ctx context.Context, userID int64) ([]Withdrawal, error)
}

func (s *stubBalanceRepo) GetBalance(ctx context.Context, userID int64) (Balance, error) {
	if s.getFn != nil {
		return s.getFn(ctx, userID)
	}
	return Balance{}, errors.New("not implemented")
}

func (s *stubBalanceRepo) CreateWithdrawal(ctx context.Context, userID int64, order string, sum float64) error {
	if s.withdrawFn != nil {
		return s.withdrawFn(ctx, userID, order, sum)
	}
	return errors.New("not implemented")
}

func (s *stubBalanceRepo) ListWithdrawals(ctx context.Context, userID int64) ([]Withdrawal, error) {
	if s.listFn != nil {
		return s.listFn(ctx, userID)
	}
	return nil, errors.New("not implemented")
}

func TestService_Withdraw_InvalidLuhn(t *testing.T) {
	svc := New(&stubBalanceRepo{})
	code, err := svc.Withdraw(context.Background(), 1, "bad", 10)
	if err != nil {
		t.Fatal(err)
	}
	if code != 422 {
		t.Fatalf("status = %d, want 422", code)
	}
}

func TestService_Withdraw_Success(t *testing.T) {
	repo := &stubBalanceRepo{
		withdrawFn: func(ctx context.Context, userID int64, order string, sum float64) error {
			return nil
		},
	}
	svc := New(repo)
	code, err := svc.Withdraw(context.Background(), 1, "79927398713", 10)
	if err != nil {
		t.Fatal(err)
	}
	if code != 200 {
		t.Fatalf("status = %d, want 200", code)
	}
}

func TestService_Withdraw_NotEnoughBalance(t *testing.T) {
	repo := &stubBalanceRepo{
		withdrawFn: func(ctx context.Context, userID int64, order string, sum float64) error {
			return ErrNotEnoughBalance
		},
	}
	svc := New(repo)
	code, err := svc.Withdraw(context.Background(), 1, "79927398713", 10)
	if err != nil {
		t.Fatal(err)
	}
	if code != 402 {
		t.Fatalf("status = %d, want 402", code)
	}
}

func TestService_Withdraw_RepoError(t *testing.T) {
	repo := &stubBalanceRepo{
		withdrawFn: func(ctx context.Context, userID int64, order string, sum float64) error {
			return errors.New("db error")
		},
	}
	svc := New(repo)
	code, err := svc.Withdraw(context.Background(), 1, "79927398713", 10)
	if err == nil {
		t.Fatal("expected error")
	}
	if code != 500 {
		t.Fatalf("status = %d, want 500", code)
	}
}

func TestService_GetBalance(t *testing.T) {
	repo := &stubBalanceRepo{
		getFn: func(ctx context.Context, userID int64) (Balance, error) {
			return Balance{Current: 100, Withdrawn: 20}, nil
		},
	}
	svc := New(repo)
	b, err := svc.GetBalance(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if b.Current != 100 || b.Withdrawn != 20 {
		t.Fatalf("got %+v", b)
	}
}

func TestService_ListWithdrawals(t *testing.T) {
	want := []Withdrawal{{Order: "o", Sum: 1}}
	repo := &stubBalanceRepo{
		listFn: func(ctx context.Context, userID int64) ([]Withdrawal, error) {
			return want, nil
		},
	}
	svc := New(repo)
	got, err := svc.ListWithdrawals(context.Background(), 1)
	if err != nil || len(got) != 1 {
		t.Fatal(err, got)
	}
}
