package user

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestUserRepository_CreateUser_OK(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id"}).AddRow(int64(8))
	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs("login", "hash").
		WillReturnRows(rows)

	repo := NewUserRepository(db)
	id, err := repo.CreateUser(context.Background(), "login", "hash")
	if err != nil || id != 8 {
		t.Fatal(err, id)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestUserRepository_CreateUser_Unique(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs("l", "h").
		WillReturnError(errors.New(`duplicate key value violates unique constraint`))

	repo := NewUserRepository(db)
	_, err = repo.CreateUser(context.Background(), "l", "h")
	if err == nil || err.Error() != "user exists" {
		t.Fatal(err)
	}
}

func TestUserRepository_GetUserByLogin(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "login", "password_hash"}).AddRow(int64(1), "a", "hpw")
	mock.ExpectQuery(`SELECT id, login, password_hash FROM users WHERE login = \$1`).
		WithArgs("a").
		WillReturnRows(rows)

	repo := NewUserRepository(db)
	u, err := repo.GetUserByLogin(context.Background(), "a")
	if err != nil || u.ID != 1 || u.Login != "a" {
		t.Fatal(err, u)
	}
}

func TestUserRepository_CreateUser_OtherDBError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectQuery(`INSERT INTO users`).
		WillReturnError(errors.New("connection reset"))

	repo := NewUserRepository(db)
	_, err = repo.CreateUser(context.Background(), "l", "h")
	if err == nil || err.Error() == "user exists" {
		t.Fatal(err)
	}
}

func TestUserRepository_GetUserByLogin_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectQuery(`SELECT id, login, password_hash FROM users WHERE login = \$1`).
		WithArgs("x").
		WillReturnError(sql.ErrNoRows)

	repo := NewUserRepository(db)
	_, err = repo.GetUserByLogin(context.Background(), "x")
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatal(err)
	}
}
