package db

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestRetryable(t *testing.T) {
	t.Parallel()

	serErr := &pgconn.PgError{Code: "40001"}
	if !Retryable(serErr) {
		t.Fatal("serialization failure must be retryable")
	}
	deadlock := &pgconn.PgError{Code: "40P01"}
	if !Retryable(deadlock) {
		t.Fatal("deadlock must be retryable")
	}
	if Retryable(errors.New("other")) {
		t.Fatal("unexpected retryable")
	}
}

func TestWithSerializableTx_Succeeds(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT 1`).WillReturnRows(sqlmock.NewRows([]string{"x"}).AddRow(1))
	mock.ExpectCommit()

	err = withSerializableTx(context.Background(), sqlDB, 3, 0, func(tx *sql.Tx) error {
		var n int
		return tx.QueryRowContext(context.Background(), `SELECT 1`).Scan(&n)
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestWithSerializableTx_RetriesOnSerializationFailure(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()

	serErr := &pgconn.PgError{Code: "40001", Message: "could not serialize access"}

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT 1`).WillReturnRows(sqlmock.NewRows([]string{"x"}).AddRow(1))
	mock.ExpectCommit().WillReturnError(serErr)

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT 1`).WillReturnRows(sqlmock.NewRows([]string{"x"}).AddRow(1))
	mock.ExpectCommit()

	err = withSerializableTx(context.Background(), sqlDB, 3, 0, func(tx *sql.Tx) error {
		var n int
		return tx.QueryRowContext(context.Background(), `SELECT 1`).Scan(&n)
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestWithSerializableTx_NoRetryOnBusinessError(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()

	bizErr := errors.New("not enough")

	mock.ExpectBegin()
	mock.ExpectRollback()

	err = withSerializableTx(context.Background(), sqlDB, 3, 0, func(tx *sql.Tx) error {
		return bizErr
	})
	if !errors.Is(err, bizErr) {
		t.Fatalf("got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
