package loyalty

import (
	"context"
	"sync"
	"time"
)

type rateLimitCoordinator struct {
	mu       sync.Mutex
	until    time.Time
	draining bool
}

func newRateLimitCoordinator() *rateLimitCoordinator {
	return &rateLimitCoordinator{}
}

// Триггерит задержку
func (c *rateLimitCoordinator) trigger(retryAfter time.Duration) {
	// Если задержка меньше 0, устанавливаем значение 1 секунду
	if retryAfter <= 0 {
		retryAfter = time.Second
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	newUntil := time.Now().Add(retryAfter)
	if newUntil.After(c.until) {
		c.until = newUntil
	}
	c.draining = true
}

// Проверяет, находится ли в режиме задержки
func (c *rateLimitCoordinator) isDraining() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.draining
	}

// Ожидает, пока не наступит время для следующего запроса
func (c *rateLimitCoordinator) waitPaused(ctx context.Context) error {
	for {
		c.mu.Lock()
		until := c.until
		c.mu.Unlock()

		wait := time.Until(until)
		if wait <= 0 {
			c.mu.Lock()
			if time.Until(c.until) <= 0 {
				c.until = time.Time{}
				c.draining = false
			}
			c.mu.Unlock()
			return nil
		}

		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
}
