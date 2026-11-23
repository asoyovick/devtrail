package auth

import "testing"

func TestNewToken(t *testing.T) {
	t1, err := NewToken()
	if err != nil {
		t.Fatalf("NewToken: %v", err)
	}
	t2, err := NewToken()
	if err != nil {
		t.Fatalf("NewToken: %v", err)
	}
	if len(t1) != 64 || len(t2) != 64 {
		t.Fatalf("unexpected token length: %d %d", len(t1), len(t2))
	}
	if t1 == t2 {
		t.Fatal("tokens should differ")
	}
}

func TestHashToken(t *testing.T) {
	h1 := HashToken("abc")
	h2 := HashToken("abc")
	h3 := HashToken("def")
	if h1 != h2 || h1 == h3 {
		t.Fatalf("unexpected hashes: %q %q %q", h1, h2, h3)
	}
	if len(h1) != 64 {
		t.Fatalf("unexpected hash length: %d", len(h1))
	}
}
