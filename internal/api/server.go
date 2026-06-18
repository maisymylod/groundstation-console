// Package api is the HTTP/JSON surface of the ops console. It serves the fleet
// summary, per-satellite telemetry windows, the incident feed, and a health
// probe. The Angular SPA reads exclusively from these endpoints.
package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/maisymylod/groundstation-console/internal/store"
)

// Server wires the store to HTTP handlers.
type Server struct {
	store *store.Store
	log   *slog.Logger
	mux   *http.ServeMux
}

// New builds the server and registers routes.
func New(st *store.Store, log *slog.Logger) *Server {
	s := &Server{store: st, log: log, mux: http.NewServeMux()}
	s.routes()
	return s
}

// Handler returns the root handler with request logging wrapped in.
func (s *Server) Handler() http.Handler {
	return s.withLogging(s.withCORS(s.mux))
}

func (s *Server) routes() {
	s.mux.HandleFunc("GET /api/health", s.handleHealth)
	s.mux.HandleFunc("GET /api/fleet", s.handleFleet)
	s.mux.HandleFunc("GET /api/satellites/{id}/window", s.handleWindow)
	s.mux.HandleFunc("GET /api/incidents", s.handleIncidents)
	s.mux.HandleFunc("POST /api/incidents/{id}/resolve", s.handleResolve)
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
		"time":   time.Now().UTC().Format(time.RFC3339),
	})
}

// handleFleet is the hot read path the benchmark optimizes. It serves the
// maintained per-satellite rollup; see store.FleetSummary.
func (s *Server) handleFleet(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, s.store.FleetSummary())
}

func (s *Server) handleWindow(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	n := queryInt(r, "n", 120)
	writeJSON(w, http.StatusOK, s.store.Window(id, n))
}

func (s *Server) handleIncidents(w http.ResponseWriter, r *http.Request) {
	limit := queryInt(r, "limit", 50)
	openOnly := r.URL.Query().Get("open") == "true"
	writeJSON(w, http.StatusOK, s.store.Incidents(limit, openOnly))
}

func (s *Server) handleResolve(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !s.store.Resolve(id) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "incident not found or already resolved"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"resolved": id})
}

func (s *Server) withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		s.log.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rec.status,
			"dur_ms", time.Since(start).Milliseconds(),
		)
	})
}

func (s *Server) withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func queryInt(r *http.Request, key string, def int) int {
	if v := r.URL.Query().Get(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
