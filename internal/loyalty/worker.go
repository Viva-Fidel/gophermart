package loyalty

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"gophermart/internal/order"
)

type SyncStore interface {
	ClaimOrdersForProcessing(ctx context.Context, limit int) ([]order.Order, error)
	SetOrderStatus(ctx context.Context, number string, status order.OrderStatus) error
	ApplyAccrual(ctx context.Context, number string, amount float64) error
}

type Worker struct {
	store  SyncStore
	client *Client
}

func NewWorker(store SyncStore, client *Client) *Worker {
	return &Worker{store: store, client: client}
}

// Запускает воркер
func (w *Worker) Run(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.processBatch(ctx)
		}
	}
}

// Обрабатывает пакет заказов
func (w *Worker) processBatch(ctx context.Context) {
	orders, err := w.store.ClaimOrdersForProcessing(ctx, 20)
	if err != nil {
		slog.ErrorContext(ctx, "loyalty claim orders", slog.Any("error", err))
		return
	}
	for _, ord := range orders {
		// Получает заказ
		resp, statusCode, retryAfter, err := w.client.GetOrder(ctx, ord.Number)
		if err != nil {
			slog.WarnContext(ctx, "loyalty get order", slog.String("order", ord.Number), slog.Any("error", err))
			continue
		}
		// Обрабатывает статусы ответа
		switch statusCode {
		case http.StatusOK:
			if resp == nil {
				continue
			}
			switch strings.ToUpper(resp.Status) {
			case string(order.OrderStatusInvalid):
				if err := w.store.SetOrderStatus(ctx, ord.Number, order.OrderStatusInvalid); err != nil {
					slog.ErrorContext(ctx, "loyalty set status invalid", slog.String("order", ord.Number), slog.Any("error", err))
				}
			case string(order.OrderStatusProcessed):
				accrualValue := 0.0
				if resp.Accrual != nil {
					accrualValue = *resp.Accrual
				}
				if err := w.store.ApplyAccrual(ctx, ord.Number, accrualValue); err != nil {
					slog.ErrorContext(ctx, "loyalty apply accrual", slog.String("order", ord.Number), slog.Any("error", err))
				}
			case string(order.OrderStatusProcessing), "REGISTERED":
				if err := w.store.SetOrderStatus(ctx, ord.Number, order.OrderStatusProcessing); err != nil {
					slog.ErrorContext(ctx, "loyalty set status processing", slog.String("order", ord.Number), slog.Any("error", err))
				}
			}
		case http.StatusNoContent:
			// Устанавливает статус обработки заказа
			if err := w.store.SetOrderStatus(ctx, ord.Number, order.OrderStatusProcessing); err != nil {
				slog.ErrorContext(ctx, "loyalty set status no content", slog.String("order", ord.Number), slog.Any("error", err))
			}
		case http.StatusTooManyRequests:
			// Ждёт время для повторного запроса
			if retryAfter > 0 {
				time.Sleep(retryAfter)
			}
		case http.StatusInternalServerError:
		default:
		}
	}
}
