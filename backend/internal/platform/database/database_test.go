package database

import (
	"context"
	"testing"
)

func TestConnectRequiresURL(t *testing.T) {
	_, err := Connect(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty database url")
	}
}

func TestConnectInvalidURL(t *testing.T) {
	_, err := Connect(context.Background(), "not-a-url")
	if err == nil {
		t.Fatal("expected error for invalid database url")
	}
}
