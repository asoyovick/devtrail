package logger

import "testing"

func TestNew(t *testing.T) {
	tests := []string{"", "info", "debug", "warn", "warning", "error"}
	for _, level := range tests {
		if _, err := New(level); err != nil {
			t.Fatalf("New(%q) error = %v", level, err)
		}
	}
}

func TestNewUnsupported(t *testing.T) {
	if _, err := New("nope"); err == nil {
		t.Fatal("expected error for unsupported level")
	}
}
