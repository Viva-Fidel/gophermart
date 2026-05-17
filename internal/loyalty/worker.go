package loyalty

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"gophermart/internal/order"
)

// Размер пула запросов
const defaultPoolSize = 5


type SyncStore interface {
	ClaimOrdersForProcessing(ctx context.Context, limit int) ([]order.Order, error)
	SetOrderStatus(ctx context.Context, number string, status order.OrderStatus) error
	ApplyAccrual(ctx context.Context, number string, amount float64) error
}

type Worker struct {
	store     SyncStore
	client    *Client
	poolSize  int
	rateLimit *rateLimitCoordinator
}

func NewWorker(store SyncStore, client *Client) *Worker {
	return &Worker{
		store:     store,
		client:    client,
		poolSize:  defaultPoolSize,
		rateLimit: newRateLimitCoordinator(),
	}
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
	if err := w.rateLimit.waitPaused(ctx); err != nil {
		return
	}

	orders, err := w.store.ClaimOrdersForProcessing(ctx, 20)
	if err != nil {
		slog.ErrorContext(ctx, "loyalty claim orders", slog.Any("error", err))
		return
	}
	if len(orders) == 0 {
		return
	}

	workers := w.poolSize
	if workers > len(orders) {
		workers = len(orders)
	}

	jobs := make(chan order.Order, len(orders))
	for _, ord := range orders {
		jobs <- ord
	}
	close(jobs)
    
	// Создаём группу для ожидания завершения работы воркеров
	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			canRequest := true
			for ord := range jobs {
				if ctx.Err() != nil {
					return
				}
				if w.rateLimit.isDraining() {
					if !canRequest {
						return
					}
					canRequest = false
				}
				if w.processOrder(ctx, ord) {
					return
				}
			}
		}()
	}
	wg.Wait()
	_ = w.rateLimit.waitPaused(ctx)
}

// Обрабатывает заказ
func (w *Worker) processOrder(ctx context.Context, ord order.Order) bool {
	// Получаем результат запроса
	result, err := w.client.GetOrder(ctx, ord.Number)
	// Если ошибка, записываем в лог и возвращаем false
	if err != nil {
		slog.WarnContext(ctx, "loyalty get order", slog.String("order", ord.Number), slog.Any("error", err))
		return false
	}

	// Обрабатываем статус код ответа
	switch result.StatusCode {
	case http.StatusOK:
		if result.Body == nil {
			return false
		}
		// Обрабатываем статус заказа
		switch strings.ToUpper(result.Body.Status) {
		// Если статус заказа INVALID, записываем в лог и возвращаем false
		case string(order.OrderStatusInvalid):
			if err := w.store.SetOrderStatus(ctx, ord.Number, order.OrderStatusInvalid); err != nil {
				slog.ErrorContext(ctx, "loyalty set status invalid", slog.String("order", ord.Number), slog.Any("error", err))
			}
		case string(order.OrderStatusProcessed):
			// Обрабатываем статус заказа PROCESSED
			accrualValue := 0.0
			if result.Body.Accrual != nil {
				accrualValue = *result.Body.Accrual
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
		// Если статус код 204, записываем в лог и возвращаем false
		if err := w.store.SetOrderStatus(ctx, ord.Number, order.OrderStatusProcessing); err != nil {
			slog.ErrorContext(ctx, "loyalty set status no content", slog.String("order", ord.Number), slog.Any("error", err))
		}
	case http.StatusTooManyRequests:
		w.rateLimit.trigger(result.RetryAfter)
		return true
	case http.StatusInternalServerError:
	default:
	}
	return false
}
