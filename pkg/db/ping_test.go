package db

import (
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestPingWithRetry_SucceedsAfterFailures(t *testing.T) {
	sqlDB, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()

	mock.ExpectPing().WillReturnError(errors.New("down"))
	mock.ExpectPing().WillReturnError(errors.New("down"))
	mock.ExpectPing().WillReturnError(nil)

	if err := pingWithRetry(sqlDB, 5, 0); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPingWithRetry_AllFail(t *testing.T) {
	sqlDB, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()

	for i := 0; i < 3; i++ {
		mock.ExpectPing().WillReturnError(errors.New("down"))
	}
	if err := pingWithRetry(sqlDB, 3, 0); err == nil {
		t.Fatal("want error")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPingWithRetry_FirstOK(t *testing.T) {
	sqlDB, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()
	mock.ExpectPing().WillReturnError(nil)
	if err := pingWithRetry(sqlDB, 5, time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
