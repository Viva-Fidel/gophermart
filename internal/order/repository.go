package order

import (
	"context"
	"database/sql"
	"log/slog"
)

type OrderRepository struct {
	db *sql.DB
}

func NewOrderRepository(database *sql.DB) *OrderRepository {
	return &OrderRepository{db: database}
}

// Создаёт заказ
func (repo *OrderRepository) CreateOrder(ctx context.Context, userID int64, number string) (created bool, ownerID int64, err error) {
	res, err := repo.db.ExecContext(ctx,
		`INSERT INTO orders (number, user_id, status) VALUES ($1, $2, 'NEW') ON CONFLICT (number) DO NOTHING`, number, userID,
	)
	if err != nil {
		return false, 0, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return false, 0, err
	}
	if affected > 0 {
		return true, userID, nil
	}
	var oid int64
	if err := repo.db.QueryRowContext(ctx, `SELECT user_id FROM orders WHERE number = $1`, number).Scan(&oid); err != nil {
		return false, 0, err
	}
	return false, oid, nil
}

// Получает список заказов
func (repo *OrderRepository) ListOrders(ctx context.Context, userID int64) ([]Order, error) {
	rows, err := repo.db.QueryContext(ctx,
		`SELECT id, number, status, accrual, uploaded_at FROM orders WHERE user_id = $1 ORDER BY uploaded_at DESC`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			slog.ErrorContext(ctx, "list orders rows close", slog.Any("error", err))
		}
	}()

	result := make([]Order, 0)
	for rows.Next() {
		var o Order
		var accrual sql.NullFloat64
		if err := rows.Scan(&o.ID, &o.Number, &o.Status, &accrual, &o.UploadedAt); err != nil {
			return nil, err
		}
		if accrual.Valid {
			o.Accrual = &accrual.Float64
		}
		result = append(result, o)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

// Забирает заказы для обработки
func (repo *OrderRepository) ClaimOrdersForProcessing(ctx context.Context, limit int) ([]Order, error) {
	rows, err := repo.db.QueryContext(ctx, `
		UPDATE orders o
		SET status = 'PROCESSING'
		WHERE o.number IN (
			SELECT number FROM orders
			WHERE status IN ('NEW', 'PROCESSING')
			ORDER BY uploaded_at ASC
			LIMIT $1
		)
		RETURNING id, number, status, uploaded_at
	`, limit)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			slog.ErrorContext(ctx, "claim orders rows close", slog.Any("error", err))
		}
	}()

	orders := make([]Order, 0)
	for rows.Next() {
		var o Order
		if err := rows.Scan(&o.ID, &o.Number, &o.Status, &o.UploadedAt); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return orders, nil
}

// Устанавливает статус заказа
func (repo *OrderRepository) SetOrderStatus(ctx context.Context, number string, status OrderStatus) error {
	_, err := repo.db.ExecContext(ctx, `UPDATE orders SET status = $2 WHERE number = $1`, number, status)
	return err
}

// Применяет начисление заказа
func (repo *OrderRepository) ApplyAccrual(ctx context.Context, number string, amount float64) error {
	_, err := repo.db.ExecContext(ctx,
		`UPDATE orders SET status = 'PROCESSED', accrual = $2 WHERE number = $1`, number, amount,
	)
	return err
}
