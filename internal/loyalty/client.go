package loyalty

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"
)

type Response struct {
	Order   string   `json:"order"`
	Status  string   `json:"status"`
	Accrual *float64 `json:"accrual,omitempty"`
}

type Client struct {
	baseURL string
	http    *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		http:    &http.Client{Timeout: 5 * time.Second},
	}
}

// Получает заказ
func (c *Client) GetOrder(ctx context.Context, number string) (*Response, int, time.Duration, error) {
	// Создаёт запрос
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/api/orders/%s", c.baseURL, number), nil)
	if err != nil {
		return nil, 0, 0, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, 0, 0, err
	}
	defer resp.Body.Close()

	// Обрабатывает статусы ответа
	if resp.StatusCode == http.StatusTooManyRequests {
		rawRA := resp.Header.Get("Retry-After")
		retryAfter, err := strconv.Atoi(rawRA)
		if err != nil {
			slog.WarnContext(ctx, "retry-after header", slog.String("value", rawRA), slog.Any("error", err))
			retryAfter = 0
		}
		if retryAfter <= 0 {
			retryAfter = 1
		}
		return nil, resp.StatusCode, time.Duration(retryAfter) * time.Second, nil
	}
	if resp.StatusCode == http.StatusNoContent {
		return nil, resp.StatusCode, 0, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, resp.StatusCode, 0, nil
	}

	var out Response
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, 0, 0, err
	}
	return &out, http.StatusOK, 0, nil
}
