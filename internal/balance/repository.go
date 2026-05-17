package balance

import (
	"context"
	"database/sql"
	"log/slog"

	pkgdb "gophermart/pkg/db"
)

type BalanceRepository struct {
	db *sql.DB
}

func NewBalanceRepository(database *sql.DB) *BalanceRepository {
	return &BalanceRepository{db: database}
}

// Получает баланс пользователя
func (repo *BalanceRepository) GetBalance(ctx context.Context, userID int64) (Balance, error) {
	var b Balance
	err := repo.db.QueryRowContext(ctx, `
		SELECT
			COALESCE((SELECT SUM(accrual) FROM orders WHERE user_id = $1 AND status = 'PROCESSED'), 0),
			COALESCE((SELECT SUM(sum) FROM withdrawals WHERE user_id = $1), 0)
	`, userID).Scan(&b.Current, &b.Withdrawn)
	if err != nil {
		return b, err
	}
	b.Current -= b.Withdrawn
	return b, nil
}

// Создаёт вывод средств
func (repo *BalanceRepository) CreateWithdrawal(ctx context.Context, userID int64, order string, sum float64) error {
	// Создаём транзакцию
	return pkgdb.WithSerializableTx(ctx, repo.db, func(tx *sql.Tx) error {
		// Получаем сумму начислений и выводов
		var accrued, withdrawn float64
		
		if err := tx.QueryRowContext(ctx, `
			SELECT
				COALESCE((SELECT SUM(accrual) FROM orders WHERE user_id = $1 AND status = 'PROCESSED'), 0),
				COALESCE((SELECT SUM(sum) FROM withdrawals WHERE user_id = $1), 0)
		`, userID).Scan(&accrued, &withdrawn); err != nil {
			return err
		}
		// Проверяем, достаточно ли средств на балансе
		if accrued-withdrawn < sum {
			// Возвращаем ошибку, если не достаточно средств
			return ErrNotEnoughBalance
		}
		// Создаём вывод средств
		_, err := tx.ExecContext(ctx,
			`INSERT INTO withdrawals (user_id, order_number, sum) VALUES ($1, $2, $3)`, userID, order, sum,
		)
		return err
	})
}

// Получает список выводов средств
func (repo *BalanceRepository) ListWithdrawals(ctx context.Context, userID int64) ([]Withdrawal, error) {
	rows, err := repo.db.QueryContext(ctx,
		`SELECT id, order_number, sum, processed_at FROM withdrawals WHERE user_id = $1 ORDER BY processed_at DESC`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			slog.ErrorContext(ctx, "list withdrawals rows close", slog.Any("error", err))
		}
	}()

	result := make([]Withdrawal, 0)
	for rows.Next() {
		var w Withdrawal
		if err := rows.Scan(&w.ID, &w.Order, &w.Sum, &w.ProcessedAt); err != nil {
			return nil, err
		}
		result = append(result, w)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}
