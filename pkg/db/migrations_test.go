package db

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestLoadUpMigrations_SortsAndSkips(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "002_second.up.sql"), "SELECT 1;")
	mustWrite(t, filepath.Join(dir, "001_first.up.sql"), "SELECT 1;")
	mustWrite(t, filepath.Join(dir, "ignore.txt"), "")
	mustWrite(t, filepath.Join(dir, "bad.up.sql"), "")

	got, err := loadUpMigrations(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("len=%d", len(got))
	}
	if got[0].version != 1 || got[1].version != 2 {
		t.Fatalf("order %+v", got)
	}
}

func TestLoadUpMigrations_MissingDir(t *testing.T) {
	_, err := loadUpMigrations(filepath.Join(t.TempDir(), "nope"))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRunMigrations_AppliesNew(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "000001_t.up.sql"), "SELECT 1;")

	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()

	mock.ExpectExec(`CREATE TABLE IF NOT EXISTS schema_migrations`).
		WillReturnResult(sqlmock.NewResult(0, 0))

	rows := sqlmock.NewRows([]string{"exists"}).AddRow(false)
	mock.ExpectQuery(`SELECT EXISTS`).
		WithArgs(int64(1)).
		WillReturnRows(rows)

	mock.ExpectBegin()
	mock.ExpectExec(`SELECT 1`).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(`INSERT INTO schema_migrations`).
		WithArgs(int64(1), "000001_t.up.sql").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	if err := RunMigrations(context.Background(), sqlDB, dir); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRunMigrations_CheckVersionFails(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "000001_t.up.sql"), "SELECT 1;")

	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()

	mock.ExpectExec(`CREATE TABLE IF NOT EXISTS schema_migrations`).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery(`SELECT EXISTS`).
		WillReturnError(errors.New("db"))

	if err := RunMigrations(context.Background(), sqlDB, dir); err == nil {
		t.Fatal("want error")
	}
}

func TestRunMigrations_SkipsApplied(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "000001_t.up.sql"), "SELECT 1;")

	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()

	mock.ExpectExec(`CREATE TABLE IF NOT EXISTS schema_migrations`).
		WillReturnResult(sqlmock.NewResult(0, 0))

	rows := sqlmock.NewRows([]string{"exists"}).AddRow(true)
	mock.ExpectQuery(`SELECT EXISTS`).
		WithArgs(int64(1)).
		WillReturnRows(rows)

	if err := RunMigrations(context.Background(), sqlDB, dir); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}
