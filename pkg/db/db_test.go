package db

import (
	"context"
	"testing"
)

func TestOpen_EmptyURI(t *testing.T) {
	_, err := Open(context.Background(), "")
	if err == nil {
		t.Fatal("expected error")
	}
}
