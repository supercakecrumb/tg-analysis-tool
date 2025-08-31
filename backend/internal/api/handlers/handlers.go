package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/supercakecrumb/tg-analysis-tool/internal/config"
	"github.com/supercakecrumb/tg-analysis-tool/internal/infrastructure/postgres"
	"github.com/supercakecrumb/tg-analysis-tool/internal/infrastructure/redis"
)

// Handler holds all HTTP handlers
type Handler struct {
	config *config.Config
	db     *postgres.DB
	redis  *redis.Client
}

// New creates a new handler instance
func New(cfg *config.Config, db *postgres.DB, redis *redis.Client) *Handler {
	return &Handler{
		config: cfg,
		db:     db,
		redis:  redis,
	}
}

// respondJSON sends a JSON response
func (h *Handler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// respondError sends an error response
func (h *Handler) respondError(w http.ResponseWriter, status int, code, message string) {
	h.respondJSON(w, status, map[string]interface{}{
		"error": map[string]interface{}{
			"code":    code,
			"message": message,
		},
	})
}

// Health checks the health of the service
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	// Check database
	dbHealthy := true
	if err := h.db.Health(r.Context()); err != nil {
		dbHealthy = false
	}

	// Check Redis
	redisHealthy := true
	if err := h.redis.Ping(r.Context()); err != nil {
		redisHealthy = false
	}

	status := http.StatusOK
	if !dbHealthy || !redisHealthy {
		status = http.StatusServiceUnavailable
	}

	h.respondJSON(w, status, map[string]interface{}{
		"status": map[string]bool{
			"healthy":  dbHealthy && redisHealthy,
			"database": dbHealthy,
			"redis":    redisHealthy,
		},
		"version": "1.0.0",
	})
}

// Metrics returns Prometheus metrics
func (h *Handler) Metrics() http.Handler {
	// In a real implementation, this would return prometheus.Handler()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("# HELP tg_messages_ingested_total Total number of messages ingested\n"))
		w.Write([]byte("# TYPE tg_messages_ingested_total counter\n"))
		w.Write([]byte("tg_messages_ingested_total 0\n"))
	})
}
