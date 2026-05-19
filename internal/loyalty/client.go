package loyalty

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

type Response struct {
	Order   string   `json:"order"`
	Status  string   `json:"status"`
	Accrual *float64 `json:"accrual,omitempty"`
}

type GetOrderResult struct {
	Body       *Response
	StatusCode int
	RetryAfter time.Duration
}

type Client struct {
	baseURL string
	http    *http.Client
}

// Создаёт клиент для работы с системой расчёта начислений
func NewClient(baseURL string) *Client {
	retryClient := retryablehttp.NewClient() // Создаём клиент с настройками для повторных запросов
	retryClient.HTTPClient.Timeout = 5 * time.Second // Таймаут запроса
	retryClient.RetryMax = 3 // Максимальное количество повторных запросов
	retryClient.RetryWaitMin = 250 * time.Millisecond // Минимальное время ожидания между повторными запросами
	retryClient.RetryWaitMax = 2 * time.Second // Максимальное время ожидания между повторными запросами
	retryClient.CheckRetry = checkRetry // Функция для проверки, нужно ли повторить запрос
	retryClient.Logger = nil // Логер для отладки

	return &Client{
		baseURL: baseURL,
		http:    retryClient.StandardClient(),
	}
}

// Проверяет, нужно ли повторить запрос
func checkRetry(_ context.Context, resp *http.Response, err error) (bool, error) {
	if err != nil {
		return true, nil
	}
	return resp.StatusCode >= http.StatusInternalServerError, nil
}

// Получает заказ
func (c *Client) GetOrder(ctx context.Context, number string) (*GetOrderResult, error) {
	// Создаёт запрос
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/api/orders/%s", c.baseURL, number), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Обрабатывает статусы ответа
	if resp.StatusCode == http.StatusTooManyRequests {
		// Получаем значение заголовка Retry-After
		rawRA := resp.Header.Get("Retry-After")
		retryAfter, err := strconv.Atoi(rawRA)
		// Если значение заголовка Retry-After не является числом, устанавливаем значение 0
		if err != nil {
			slog.WarnContext(ctx, "retry-after header", slog.String("value", rawRA), slog.Any("error", err))
			retryAfter = 0
		}
		if retryAfter <= 0 {
			retryAfter = 1
		}
		// Возвращаем результат
		return &GetOrderResult{
			StatusCode: resp.StatusCode,
			RetryAfter: time.Duration(retryAfter) * time.Second,
		}, nil
	}
	// Если статус код 204, возвращаем результат
	if resp.StatusCode == http.StatusNoContent {
		return &GetOrderResult{StatusCode: resp.StatusCode}, nil
	}
	// Если статус код не 200, возвращаем результат
	if resp.StatusCode != http.StatusOK {
		return &GetOrderResult{StatusCode: resp.StatusCode}, nil
	}

	var out Response
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &GetOrderResult{
		Body:       &out,
		StatusCode: http.StatusOK,
	}, nil
}
