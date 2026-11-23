package database

import (
	"context"
	"os"
	"testing"
)

func TestConnectIntegration(t *testing.T) {
	dbURL := os.Getenv("DEVTRAIL_TEST_DATABASE_URL")
	if dbURL == "" {
		t.Skip("DEVTRAIL_TEST_DATABASE_URL not set")
	}
	pool, err := Connect(context.Background(), dbURL)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	defer pool.Close()
	if err := pool.Ping(context.Background()); err != nil {
		t.Fatalf("Ping: %v", err)
	}
}
