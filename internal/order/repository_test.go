package order

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestOrderRepository_CreateOrder_Inserted(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectExec(`INSERT INTO orders`).
		WithArgs("79927398713", int64(2)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	repo := NewOrderRepository(db)
	created, owner, err := repo.CreateOrder(context.Background(), 2, "79927398713")
	if err != nil || !created || owner != 2 {
		t.Fatal(err, created, owner)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestOrderRepository_CreateOrder_Existing(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectExec(`INSERT INTO orders`).
		WithArgs("79927398713", int64(2)).
		WillReturnResult(sqlmock.NewResult(0, 0))

	rows := sqlmock.NewRows([]string{"user_id"}).AddRow(int64(9))
	mock.ExpectQuery(`SELECT user_id FROM orders WHERE number = \$1`).
		WithArgs("79927398713").
		WillReturnRows(rows)

	repo := NewOrderRepository(db)
	created, owner, err := repo.CreateOrder(context.Background(), 2, "79927398713")
	if err != nil || created || owner != 9 {
		t.Fatal(err, created, owner)
	}
}

func TestOrderRepository_ListOrders(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	ts := time.Unix(5, 0)
	rows := sqlmock.NewRows([]string{"id", "number", "status", "accrual", "uploaded_at"}).
		AddRow(int64(1), "79927398713", "NEW", nil, ts)
	mock.ExpectQuery(`SELECT id, number, status, accrual, uploaded_at FROM orders`).
		WithArgs(int64(1)).
		WillReturnRows(rows)

	repo := NewOrderRepository(db)
	list, err := repo.ListOrders(context.Background(), 1)
	if err != nil || len(list) != 1 || list[0].Number != "79927398713" {
		t.Fatal(err, list)
	}
}

func TestOrderRepository_SetOrderStatus(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectExec(`UPDATE orders SET status`).
		WithArgs("n", OrderStatusProcessed).
		WillReturnResult(sqlmock.NewResult(0, 1))

	repo := NewOrderRepository(db)
	if err := repo.SetOrderStatus(context.Background(), "n", OrderStatusProcessed); err != nil {
		t.Fatal(err)
	}
}

func TestOrderRepository_ClaimOrdersForProcessing(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	ts := time.Unix(2, 0)
	r := sqlmock.NewRows([]string{"id", "number", "status", "uploaded_at"}).
		AddRow(int64(1), "79927398713", "PROCESSING", ts)
	mock.ExpectQuery(`UPDATE orders`).
		WithArgs(5).
		WillReturnRows(r)

	repo := NewOrderRepository(db)
	list, err := repo.ClaimOrdersForProcessing(context.Background(), 5)
	if err != nil || len(list) != 1 {
		t.Fatal(err, list)
	}
}

func TestOrderRepository_ApplyAccrual(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectExec(`UPDATE orders SET status = 'PROCESSED'`).
		WithArgs("n", 3.5).
		WillReturnResult(sqlmock.NewResult(0, 1))

	repo := NewOrderRepository(db)
	if err := repo.ApplyAccrual(context.Background(), "n", 3.5); err != nil {
		t.Fatal(err)
	}
}
