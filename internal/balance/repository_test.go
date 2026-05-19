package balance

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestBalanceRepository_GetBalance(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"col1", "col2"}).AddRow(100.0, 30.0)
	mock.ExpectQuery(`SELECT`).
		WithArgs(int64(1)).
		WillReturnRows(rows)

	repo := NewBalanceRepository(db)
	b, err := repo.GetBalance(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if b.Current != 70 || b.Withdrawn != 30 {
		t.Fatalf("%+v", b)
	}
}

func TestBalanceRepository_ListWithdrawals(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	ts := time.Unix(1, 0)
	r := sqlmock.NewRows([]string{"id", "order_number", "sum", "processed_at"}).
		AddRow(int64(1), "o", 2.5, ts)
	mock.ExpectQuery(`SELECT id, order_number, sum, processed_at FROM withdrawals`).
		WithArgs(int64(1)).
		WillReturnRows(r)

	repo := NewBalanceRepository(db)
	list, err := repo.ListWithdrawals(context.Background(), 1)
	if err != nil || len(list) != 1 || list[0].Order != "o" {
		t.Fatal(err, list)
	}
}

func TestBalanceRepository_CreateWithdrawal_OK(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectBegin()
	rows := sqlmock.NewRows([]string{"a", "b"}).AddRow(50.0, 10.0)
	mock.ExpectQuery(`SELECT`).
		WithArgs(int64(1)).
		WillReturnRows(rows)
	mock.ExpectExec(`INSERT INTO withdrawals`).
		WithArgs(int64(1), "79927398713", 5.0).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	repo := NewBalanceRepository(db)
	if err := repo.CreateWithdrawal(context.Background(), 1, "79927398713", 5.0); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestBalanceRepository_CreateWithdrawal_InsertFails(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectBegin()
	rows := sqlmock.NewRows([]string{"a", "b"}).AddRow(100.0, 0.0)
	mock.ExpectQuery(`SELECT`).
		WithArgs(int64(1)).
		WillReturnRows(rows)
	mock.ExpectExec(`INSERT INTO withdrawals`).
		WillReturnError(errors.New("insert fail"))
	mock.ExpectRollback()

	repo := NewBalanceRepository(db)
	err = repo.CreateWithdrawal(context.Background(), 1, "79927398713", 1.0)
	if err == nil {
		t.Fatal("want error")
	}
}

func TestBalanceRepository_CreateWithdrawal_NotEnough(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectBegin()
	rows := sqlmock.NewRows([]string{"a", "b"}).AddRow(10.0, 9.0)
	mock.ExpectQuery(`SELECT`).
		WithArgs(int64(1)).
		WillReturnRows(rows)
	mock.ExpectRollback()

	repo := NewBalanceRepository(db)
	err = repo.CreateWithdrawal(context.Background(), 1, "79927398713", 5.0)
	if err != ErrNotEnoughBalance {
		t.Fatal(err)
	}
}
