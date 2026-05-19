package db

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// Открывает соединение с БД
func Open(ctx context.Context, uri string) (*sql.DB, error) {
	// Проверяет, что URI не пустой
	if uri == "" {
		return nil, fmt.Errorf("empty DATABASE_URI")
	}

	// Открывает соединение с БД
	sqlDB, err := sql.Open("pgx", uri)
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Делает попытки подключения к БД несколько раз с задержкой
	if err := pingWithRetry(sqlDB, 10, time.Second); err != nil {
		if closeErr := sqlDB.Close(); closeErr != nil {
			slog.ErrorContext(ctx, "close db after ping failure", slog.Any("error", closeErr))
		}
		return nil, err
	}
	return sqlDB, nil
}

// Делает попытки подключения к БД несколько раз с задержкой
func pingWithRetry(db *sql.DB, attempts int, delay time.Duration) error {
	var lastErr error
	// Делает попытки подключения к БД несколько раз с задержкой
	for i := 0; i < attempts; i++ {
		// Проверяет, что соединение с БД активно
		if err := db.Ping(); err == nil {
			return nil
		} else {
			lastErr = err
		}
		time.Sleep(delay)
	}
	return fmt.Errorf("database ping failed after %d attempts: %w", attempts, lastErr)
}
