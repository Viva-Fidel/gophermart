package db

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

const (
	defaultTxMaxAttempts = 5
	defaultTxRetryDelay    = 10 * time.Millisecond
)

// Проверяет, является ли ошибка ошибкой, которая может быть исправлена повторным запросом
func Retryable(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "40001", "40P01": // Ошибки сериализации и дедлока
			return true
		}
	}
	return false
}

// Выполняет транзакцию с изоляцией сериализации и повторяет попытки на ошибки, которые могут быть исправлены повторным запросом
func WithSerializableTx(ctx context.Context, database *sql.DB, fn func(tx *sql.Tx) error) error {
	return withSerializableTx(ctx, database, defaultTxMaxAttempts, defaultTxRetryDelay, fn)
}

// Выполняет транзакцию с изоляцией сериализации и повторяет попытки на ошибки, которые могут быть исправлены повторным запросом
func withSerializableTx(
	ctx context.Context,
	database *sql.DB,
	maxAttempts int,
	retryDelay time.Duration,
	fn func(tx *sql.Tx) error,
) error {
	var lastErr error
	// Выполняет транзакцию с изоляцией сериализации и повторяет попытки на ошибки, которые могут быть исправлены повторным запросом
	for attempt := 0; attempt < maxAttempts; attempt++ {
		// Если попытка не первая и задержка больше 0, ждём
		if attempt > 0 && retryDelay > 0 {
			time.Sleep(retryDelay * time.Duration(attempt))
		}
		// Выполняет транзакцию с изоляцией сериализации
		err := runSerializableTx(ctx, database, fn)
		// Если ошибка nil, возвращаем nil
		if err == nil {
			return nil
		}
		if !Retryable(err) {
			return err
		}
		lastErr = err
	}
	return lastErr
}

// Выполняет транзакцию с изоляцией сериализации и повторяет попытки на ошибки, которые могут быть исправлены повторным запросом
func runSerializableTx(ctx context.Context, database *sql.DB, fn func(tx *sql.Tx) error) error {
	// Выполняет транзакцию с изоляцией сериализации
	tx, err := database.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	// Откатываем транзакцию при ошибке
	defer func() {
		_ = tx.Rollback()
	}()

	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit()
}
