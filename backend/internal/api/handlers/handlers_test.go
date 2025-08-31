package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/supercakecrumb/tg-analysis-tool/internal/config"
	"github.com/supercakecrumb/tg-analysis-tool/internal/infrastructure/postgres"
	"github.com/supercakecrumb/tg-analysis-tool/internal/infrastructure/redis"
)

// Mock implementations for testing
type mockDB struct{}

func (m *mockDB) Health(ctx context.Context) error {
	return nil
}

type mockRedis struct{}

func (m *mockRedis) Ping(ctx context.Context) error {
	return nil
}

func TestHandler_Health(t *testing.T) {
	// Create test configuration
	cfg := &config.Config{
		Port: 8080,
		Env:  "test",
	}

	// Test with uninitialized DB (will return unhealthy)
	t.Run("database unhealthy", func(t *testing.T) {
		h := &Handler{
			config: cfg,
			db:     &postgres.DB{}, // Uninitialized DB
			redis:  &redis.Client{},
		}

		req, err := http.NewRequest("GET", "/health", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(h.Health)
		handler.ServeHTTP(rr, req)

		// With uninitialized DB, should return 503
		if status := rr.Code; status != http.StatusServiceUnavailable {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusServiceUnavailable)
		}

		// Check that we got a valid JSON response
		var response map[string]interface{}
		if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
			t.Errorf("handler returned invalid JSON: %v", err)
		}
	})
}

func TestHandler_respondJSON(t *testing.T) {
	h := &Handler{}

	tests := []struct {
		name         string
		status       int
		data         interface{}
		expectedJSON string
	}{
		{
			name:   "simple object",
			status: http.StatusOK,
			data: map[string]string{
				"message": "success",
			},
			expectedJSON: `{"message":"success"}`,
		},
		{
			name:   "nested object",
			status: http.StatusCreated,
			data: map[string]interface{}{
				"user": map[string]string{
					"id":   "123",
					"name": "test",
				},
			},
			expectedJSON: `{"user":{"id":"123","name":"test"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			h.respondJSON(rr, tt.status, tt.data)

			// Check status code
			if rr.Code != tt.status {
				t.Errorf("wrong status code: got %v want %v", rr.Code, tt.status)
			}

			// Check content type
			if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
				t.Errorf("wrong content type: got %v want application/json", ct)
			}

			// Check JSON content
			var got interface{}
			var want interface{}

			if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
				t.Fatalf("failed to unmarshal response: %v", err)
			}

			if err := json.Unmarshal([]byte(tt.expectedJSON), &want); err != nil {
				t.Fatalf("failed to unmarshal expected: %v", err)
			}

			gotJSON, _ := json.Marshal(got)
			wantJSON, _ := json.Marshal(want)

			if string(gotJSON) != string(wantJSON) {
				t.Errorf("wrong JSON response:\ngot:  %s\nwant: %s", gotJSON, wantJSON)
			}
		})
	}
}

func TestHandler_respondError(t *testing.T) {
	h := &Handler{}

	tests := []struct {
		name    string
		status  int
		code    string
		message string
	}{
		{
			name:    "bad request error",
			status:  http.StatusBadRequest,
			code:    "INVALID_REQUEST",
			message: "Invalid request parameters",
		},
		{
			name:    "unauthorized error",
			status:  http.StatusUnauthorized,
			code:    "UNAUTHORIZED",
			message: "Authentication required",
		},
		{
			name:    "internal server error",
			status:  http.StatusInternalServerError,
			code:    "INTERNAL_ERROR",
			message: "An unexpected error occurred",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			h.respondError(rr, tt.status, tt.code, tt.message)

			// Check status code
			if rr.Code != tt.status {
				t.Errorf("wrong status code: got %v want %v", rr.Code, tt.status)
			}

			// Check error response structure
			var response map[string]interface{}
			if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
				t.Fatalf("failed to unmarshal response: %v", err)
			}

			errorObj, ok := response["error"].(map[string]interface{})
			if !ok {
				t.Fatal("response does not contain error object")
			}

			if code := errorObj["code"]; code != tt.code {
				t.Errorf("wrong error code: got %v want %v", code, tt.code)
			}

			if msg := errorObj["message"]; msg != tt.message {
				t.Errorf("wrong error message: got %v want %v", msg, tt.message)
			}
		})
	}
}

func TestHandler_Metrics(t *testing.T) {
	h := &Handler{}

	req, err := http.NewRequest("GET", "/metrics", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := h.Metrics()
	handler.ServeHTTP(rr, req)

	// Check status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check content type
	if ct := rr.Header().Get("Content-Type"); ct != "text/plain" {
		t.Errorf("wrong content type: got %v want text/plain", ct)
	}

	// Check that response contains Prometheus metrics format
	body := rr.Body.String()
	if !contains(body, "# HELP") || !contains(body, "# TYPE") {
		t.Error("response does not appear to be in Prometheus format")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || contains(s[1:], substr) || (len(s) > len(substr) && s[:len(substr)] == substr))
}

// Benchmark tests
func BenchmarkHandler_respondJSON(b *testing.B) {
	h := &Handler{}
	data := map[string]interface{}{
		"id":      "123",
		"name":    "test",
		"email":   "test@example.com",
		"created": "2024-01-01T00:00:00Z",
		"tags":    []string{"tag1", "tag2", "tag3"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		h.respondJSON(rr, http.StatusOK, data)
	}
}
