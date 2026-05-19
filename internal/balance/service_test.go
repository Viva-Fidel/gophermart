package balance

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/mock/gomock"
)

func TestService_Withdraw_InvalidLuhn(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := NewMockRepository(ctrl)
	svc := New(repo)
	code, err := svc.Withdraw(context.Background(), 1, "bad", 10)
	if err != nil {
		t.Fatal(err)
	}
	if code != 422 {
		t.Fatalf("status = %d, want 422", code)
	}
}

func TestService_Withdraw_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := NewMockRepository(ctrl)
	repo.EXPECT().CreateWithdrawal(gomock.Any(), int64(1), "79927398713", float64(10)).Return(nil)
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
	ctrl := gomock.NewController(t)
	repo := NewMockRepository(ctrl)
	repo.EXPECT().CreateWithdrawal(gomock.Any(), int64(1), "79927398713", float64(10)).Return(ErrNotEnoughBalance)
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
	ctrl := gomock.NewController(t)
	repo := NewMockRepository(ctrl)
	repo.EXPECT().CreateWithdrawal(gomock.Any(), int64(1), "79927398713", float64(10)).Return(errors.New("db error"))
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
	ctrl := gomock.NewController(t)
	repo := NewMockRepository(ctrl)
	repo.EXPECT().GetBalance(gomock.Any(), int64(1)).Return(Balance{Current: 100, Withdrawn: 20}, nil)
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
	ctrl := gomock.NewController(t)
	repo := NewMockRepository(ctrl)
	repo.EXPECT().ListWithdrawals(gomock.Any(), int64(1)).Return(want, nil)
	svc := New(repo)
	got, err := svc.ListWithdrawals(context.Background(), 1)
	if err != nil || len(got) != 1 {
		t.Fatal(err, got)
	}
}
