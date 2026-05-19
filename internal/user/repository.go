package user

import (
	"context"
	"database/sql"
	"errors"
	"strings"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(database *sql.DB) *UserRepository {
	return &UserRepository{db: database}
}

// Проверяет, является ли ошибка уникальностью
func isUniqueViolation(err error) bool {
	return err != nil && (strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique"))
}

// Создаёт пользователя
func (repo *UserRepository) CreateUser(ctx context.Context, login, passwordHash string) (int64, error) {
	var id int64
	err := repo.db.QueryRowContext(ctx,
		`INSERT INTO users (login, password_hash) VALUES ($1, $2) RETURNING id`, login, passwordHash,
	).Scan(&id)
	if err != nil {
		if isUniqueViolation(err) {
			return 0, errors.New("user exists")
		}
		return 0, err
	}
	return id, nil
}

// Получает пользователя по логину
func (repo *UserRepository) GetUserByLogin(ctx context.Context, login string) (*User, error) {
	u := &User{}
	err := repo.db.QueryRowContext(ctx,
		`SELECT id, login, password_hash FROM users WHERE login = $1`, login,
	).Scan(&u.ID, &u.Login, &u.PasswordHash)
	if err != nil {
		return nil, err
	}
	return u, nil
}

