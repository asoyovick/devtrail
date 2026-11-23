package httpserver

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestSecurityHeaders(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	SecurityHeaders(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(200) })).ServeHTTP(rr, req)
	if rr.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Fatal("missing nosniff")
	}
	if rr.Header().Get("X-Frame-Options") != "DENY" {
		t.Fatal("missing frame options")
	}
}

func TestCORS(t *testing.T) {
	h := CORS([]string{"http://localhost:3000"}, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(200) }))
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/api/v1/me", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("status=%d", rr.Code)
	}
	if rr.Header().Get("Access-Control-Allow-Origin") != "http://localhost:3000" {
		t.Fatal("missing allow origin")
	}
}

func TestRateLimiter(t *testing.T) {
	rl := NewRateLimiter(2, time.Minute)
	h := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(200) }))
	for i := 0; i < 2; i++ {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "1.2.3.4:1234"
		h.ServeHTTP(rr, req)
		if rr.Code != 200 {
			t.Fatalf("i=%d status=%d", i, rr.Code)
		}
	}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "1.2.3.4:1234"
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusTooManyRequests {
		t.Fatalf("status=%d", rr.Code)
	}
}

func TestBodyLimit(t *testing.T) {
	h := BodyLimit(4, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := make([]byte, 10)
		_, err := r.Body.Read(buf)
		if err == nil {
			w.WriteHeader(200)
			return
		}
		w.WriteHeader(413)
		_ = err
	}))
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("1234567890"))
	h.ServeHTTP(rr, req)
	if rr.Code != 413 {
		t.Fatalf("status=%d", rr.Code)
	}
}
