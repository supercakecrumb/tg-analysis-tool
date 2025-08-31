package api

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	"github.com/rs/zerolog/log"
	"github.com/supercakecrumb/tg-analysis-tool/internal/api/handlers"
	"github.com/supercakecrumb/tg-analysis-tool/internal/api/middleware/auth"
	"github.com/supercakecrumb/tg-analysis-tool/internal/config"
	"github.com/supercakecrumb/tg-analysis-tool/internal/infrastructure/postgres"
	"github.com/supercakecrumb/tg-analysis-tool/internal/infrastructure/redis"
)

// Server represents the HTTP server
type Server struct {
	router *chi.Mux
	config *config.Config
	db     *postgres.DB
	redis  *redis.Client
}

// NewServer creates a new HTTP server
func NewServer(cfg *config.Config, db *postgres.DB, redis *redis.Client) *Server {
	s := &Server{
		router: chi.NewRouter(),
		config: cfg,
		db:     db,
		redis:  redis,
	}

	s.setupMiddleware()
	s.setupRoutes()

	return s
}

// ServeHTTP implements the http.Handler interface
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

// setupMiddleware configures all middleware
func (s *Server) setupMiddleware() {
	// Request ID
	s.router.Use(middleware.RequestID)

	// Real IP
	s.router.Use(middleware.RealIP)

	// Logger
	s.router.Use(middleware.RequestLogger(&middleware.DefaultLogFormatter{
		Logger: &log.Logger,
	}))

	// Recoverer
	s.router.Use(middleware.Recoverer)

	// Timeout
	s.router.Use(middleware.Timeout(60 * time.Second))

	// CORS
	s.router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{s.config.BaseURL},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link", "X-Total-Count", "X-Page", "X-Per-Page"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Rate limiting
	s.router.Use(httprate.LimitByIP(100, 1*time.Minute))

	// Compression
	s.router.Use(middleware.Compress(5))
}

// setupRoutes configures all routes
func (s *Server) setupRoutes() {
	// Create handlers
	h := handlers.New(s.config, s.db, s.redis)

	// Health check
	s.router.Get("/health", h.Health)

	// API routes
	s.router.Route("/api", func(r chi.Router) {
		// Public routes
		r.Group(func(r chi.Router) {
			// Authentication endpoints
			r.Post("/login/telegram/qr", h.LoginQR)
			r.Post("/login/telegram/phone", h.LoginPhone)
			r.Post("/login/telegram/verify", h.LoginVerify)
			r.Post("/login/telegram/password", h.LoginPassword)

			// Public share viewing
			r.Get("/p/{token}", h.ViewShare)
		})

		// Protected routes
		r.Group(func(r chi.Router) {
			// Authentication middleware
			r.Use(auth.RequireAuth(s.redis))

			// User endpoints
			r.Get("/user/me", h.GetCurrentUser)
			r.Put("/user/me", h.UpdateCurrentUser)
			r.Delete("/user/me", h.DeleteCurrentUser)

			// Sync endpoints
			r.Post("/sync/start", h.StartSync)
			r.Get("/sync/status", h.SyncStatus) // SSE endpoint

			// Analytics endpoints
			r.Route("/analytics", func(r chi.Router) {
				r.Get("/activity", h.GetActivity)
				r.Get("/heatmap", h.GetHeatmap)
				r.Get("/top-chats", h.GetTopChats)
				r.Get("/message-types", h.GetMessageTypes)
				r.Get("/response-times", h.GetResponseTimes)
				r.Get("/top-words", h.GetTopWords)
				r.Get("/emojis", h.GetEmojis)
			})

			// Settings endpoints
			r.Post("/settings/exclusions", h.UpdateExclusions)
			r.Get("/settings/exclusions", h.GetExclusions)

			// Export endpoints
			r.Get("/export", h.ExportData)

			// Sharing endpoints
			r.Post("/shares", h.CreateShare)
			r.Get("/shares", h.ListShares)
			r.Delete("/shares/{id}", h.DeleteShare)

			// Organization endpoints (multi-tenant)
			r.Route("/orgs", func(r chi.Router) {
				r.Get("/", h.ListOrgs)
				r.Post("/", h.CreateOrg)
				r.Get("/{orgID}", h.GetOrg)
				r.Put("/{orgID}", h.UpdateOrg)
				r.Delete("/{orgID}", h.DeleteOrg)

				// Organization members
				r.Get("/{orgID}/members", h.ListOrgMembers)
				r.Post("/{orgID}/members", h.AddOrgMember)
				r.Delete("/{orgID}/members/{userID}", h.RemoveOrgMember)
			})
		})
	})

	// Metrics endpoint (if enabled)
	if s.config.EnableMetrics {
		s.router.Handle("/metrics", h.Metrics())
	}

	// Profiling endpoints (if enabled)
	if s.config.EnablePprof {
		s.router.Mount("/debug", middleware.Profiler())
	}

	// Static files (for production builds)
	// In development, Next.js dev server handles this
	if s.config.Env == "production" {
		fileServer := http.FileServer(http.Dir("./web/out"))
		s.router.Handle("/*", fileServer)
	}
}
