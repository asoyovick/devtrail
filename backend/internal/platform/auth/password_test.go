package auth

import "testing"

func TestHashAndCheckPassword(t *testing.T) {
	hash, err := HashPassword("correct-horse-battery-staple")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if hash == "" || hash == "correct-horse-battery-staple" {
		t.Fatal("expected hashed password")
	}
	if err := CheckPassword(hash, "correct-horse-battery-staple"); err != nil {
		t.Fatalf("CheckPassword: %v", err)
	}
	if err := CheckPassword(hash, "wrong"); err == nil {
		t.Fatal("expected error for wrong password")
	}
}

func TestHashPasswordEmpty(t *testing.T) {
	if _, err := HashPassword(""); err == nil {
		t.Fatal("expected error")
	}
}
