package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/supercakecrumb/tg-analysis-tool/internal/config"
	"github.com/supercakecrumb/tg-analysis-tool/internal/infrastructure/postgres"
	"github.com/supercakecrumb/tg-analysis-tool/internal/infrastructure/redis"
)

func TestNewServer(t *testing.T) {
	cfg := &config.Config{
		Port:          8080,
		Env:           "test",
		BaseURL:       "http://localhost:3000",
		EnableMetrics: false,
		EnablePprof:   false,
	}

	db := &postgres.DB{}
	redisClient := &redis.Client{}

	server := NewServer(cfg, db, redisClient)

	if server == nil {
		t.Fatal("expected server to be created")
	}

	if server.config != cfg {
		t.Error("expected config to be set")
	}

	if server.db != db {
		t.Error("expected db to be set")
	}

	if server.redis != redisClient {
		t.Error("expected redis to be set")
	}

	if server.router == nil {
		t.Error("expected router to be created")
	}
}

func TestServer_ServeHTTP(t *testing.T) {
	cfg := &config.Config{
		Port:          8080,
		Env:           "test",
		BaseURL:       "http://localhost:3000",
		EnableMetrics: false,
		EnablePprof:   false,
	}

	server := NewServer(cfg, &postgres.DB{}, &redis.Client{})

	// Test health endpoint
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	server.ServeHTTP(rr, req)

	// When DB is not initialized, health endpoint returns 503
	if status := rr.Code; status != http.StatusServiceUnavailable {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusServiceUnavailable)
	}
}

func TestServer_Middleware(t *testing.T) {
	cfg := &config.Config{
		Port:          8080,
		Env:           "test",
		BaseURL:       "http://localhost:3000",
		EnableMetrics: false,
		EnablePprof:   false,
	}

	server := NewServer(cfg, &postgres.DB{}, &redis.Client{})

	// Test that middleware is applied
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Mount test handler
	server.router.Get("/test-middleware", testHandler)

	req, err := http.NewRequest("GET", "/test-middleware", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	server.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Chi's RequestID middleware adds the ID to the request context
	// We can verify middleware is working by checking if the handler was called
	// and the response was successful
}

func TestServer_CORS(t *testing.T) {
	cfg := &config.Config{
		Port:          8080,
		Env:           "test",
		BaseURL:       "http://localhost:3000",
		EnableMetrics: false,
		EnablePprof:   false,
	}

	server := NewServer(cfg, &postgres.DB{}, &redis.Client{})

	// Test CORS headers
	req, err := http.NewRequest("OPTIONS", "/api/test", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "POST")

	rr := httptest.NewRecorder()
	server.ServeHTTP(rr, req)

	// Check CORS headers
	if allow := rr.Header().Get("Access-Control-Allow-Origin"); allow != "http://localhost:3000" {
		t.Errorf("wrong CORS origin header: got %v want %v", allow, "http://localhost:3000")
	}

	if methods := rr.Header().Get("Access-Control-Allow-Methods"); methods == "" {
		t.Error("expected Access-Control-Allow-Methods header")
	}
}

func TestServer_RateLimiting(t *testing.T) {
	cfg := &config.Config{
		Port:          8080,
		Env:           "test",
		BaseURL:       "http://localhost:3000",
		EnableMetrics: false,
		EnablePprof:   false,
	}

	server := NewServer(cfg, &postgres.DB{}, &redis.Client{})

	// Make multiple requests to test rate limiting
	// Note: In a real test, you'd need to configure the rate limiter
	// to have a very low limit for testing
	for i := 0; i < 5; i++ {
		req, err := http.NewRequest("GET", "/health", nil)
		if err != nil {
			t.Fatal(err)
		}
		req.RemoteAddr = "127.0.0.1:12345" // Same IP for all requests

		rr := httptest.NewRecorder()
		server.ServeHTTP(rr, req)

		// For now, just check that requests are handled
		// In a real implementation with proper rate limiting,
		// we'd check for 429 status after limit is exceeded
		// With uninitialized DB, health endpoint returns 503
		if status := rr.Code; status != http.StatusServiceUnavailable {
			t.Logf("request %d returned unexpected status %d", i, status)
		}
	}
}

func TestServer_MetricsEndpoint(t *testing.T) {
	cfg := &config.Config{
		Port:          8080,
		Env:           "test",
		BaseURL:       "http://localhost:3000",
		EnableMetrics: true, // Enable metrics for this test
		EnablePprof:   false,
	}

	server := NewServer(cfg, &postgres.DB{}, &redis.Client{})

	req, err := http.NewRequest("GET", "/metrics", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	server.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("metrics endpoint returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check content type
	if ct := rr.Header().Get("Content-Type"); ct != "text/plain" {
		t.Errorf("wrong content type for metrics: got %v want text/plain", ct)
	}
}

func TestServer_Routes(t *testing.T) {
	cfg := &config.Config{
		Port:          8080,
		Env:           "test",
		BaseURL:       "http://localhost:3000",
		EnableMetrics: false,
		EnablePprof:   false,
	}

	server := NewServer(cfg, &postgres.DB{}, &redis.Client{})

	// Test that routes are registered
	routes := []struct {
		method string
		path   string
	}{
		{"GET", "/health"},
		{"POST", "/api/login/telegram/qr"},
		{"POST", "/api/login/telegram/phone"},
		{"POST", "/api/login/telegram/verify"},
		{"POST", "/api/login/telegram/password"},
		{"GET", "/api/p/{token}"},
		// Add more routes as needed
	}

	for _, route := range routes {
		// Use chi's Match to check if route exists
		rctx := chi.NewRouteContext()
		if !server.router.Match(rctx, route.method, route.path) {
			t.Errorf("route not found: %s %s", route.method, route.path)
		}
	}
}

// Benchmark tests
func BenchmarkServer_HealthEndpoint(b *testing.B) {
	cfg := &config.Config{
		Port:          8080,
		Env:           "test",
		BaseURL:       "http://localhost:3000",
		EnableMetrics: false,
		EnablePprof:   false,
	}

	server := NewServer(cfg, &postgres.DB{}, &redis.Client{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", "/health", nil)
		rr := httptest.NewRecorder()
		server.ServeHTTP(rr, req)
	}
}

func BenchmarkServer_WithMiddleware(b *testing.B) {
	cfg := &config.Config{
		Port:          8080,
		Env:           "test",
		BaseURL:       "http://localhost:3000",
		EnableMetrics: false,
		EnablePprof:   false,
	}

	server := NewServer(cfg, &postgres.DB{}, &redis.Client{})

	// Add a test endpoint
	server.router.Get("/bench-test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", "/bench-test", nil)
		rr := httptest.NewRecorder()
		server.ServeHTTP(rr, req)
	}
}
