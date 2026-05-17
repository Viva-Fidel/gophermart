package loyalty

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRateLimitCoordinator_waitPaused_Cancel(t *testing.T) {
	c := newRateLimitCoordinator()
	c.trigger(5 * time.Second)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- c.waitPaused(ctx) }()

	time.Sleep(20 * time.Millisecond)
	cancel()

	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v", err)
	}
}

func TestRateLimitCoordinator_triggerExtendsPause(t *testing.T) {
	c := newRateLimitCoordinator()
	c.trigger(200 * time.Millisecond)
	time.Sleep(50 * time.Millisecond)
	c.trigger(300 * time.Millisecond)

	start := time.Now()
	_ = c.waitPaused(context.Background())
	if elapsed := time.Since(start); elapsed < 200*time.Millisecond {
		t.Fatalf("pause too short: %v", elapsed)
	}
}
