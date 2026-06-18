package api

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/maisymylod/groundstation-console/internal/store"
	"github.com/maisymylod/groundstation-console/internal/telemetry"
)

func testServer(t *testing.T) http.Handler {
	t.Helper()
	st := store.New(0)
	for _, f := range telemetry.GenerateFixture(42, 200) {
		st.Ingest(f)
	}
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	return New(st, log).Handler()
}

func TestHealth(t *testing.T) {
	srv := testServer(t)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/health", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("health status: got %d", rec.Code)
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["status"] != "ok" {
		t.Errorf("status: got %q", body["status"])
	}
}

func TestFleet(t *testing.T) {
	srv := testServer(t)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/fleet", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("fleet status: got %d", rec.Code)
	}
	var fleet []store.SatSummary
	if err := json.Unmarshal(rec.Body.Bytes(), &fleet); err != nil {
		t.Fatal(err)
	}
	if len(fleet) != len(telemetry.Fleet) {
		t.Errorf("fleet size: got %d want %d", len(fleet), len(telemetry.Fleet))
	}
}

func TestWindow(t *testing.T) {
	srv := testServer(t)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/satellites/HS-01/window?n=30", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("window status: got %d", rec.Code)
	}
	var w telemetry.Window
	if err := json.Unmarshal(rec.Body.Bytes(), &w); err != nil {
		t.Fatal(err)
	}
	if len(w.Frames) != 30 {
		t.Errorf("window frames: got %d want 30", len(w.Frames))
	}
}

func TestIncidentsAndResolve(t *testing.T) {
	srv := testServer(t)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/incidents?open=true&limit=5", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("incidents status: got %d", rec.Code)
	}
	var incs []telemetry.Incident
	if err := json.Unmarshal(rec.Body.Bytes(), &incs); err != nil {
		t.Fatal(err)
	}
	if len(incs) == 0 {
		t.Fatal("expected open incidents")
	}

	rec2 := httptest.NewRecorder()
	srv.ServeHTTP(rec2, httptest.NewRequest(http.MethodPost, "/api/incidents/"+incs[0].ID+"/resolve", nil))
	if rec2.Code != http.StatusOK {
		t.Fatalf("resolve status: got %d", rec2.Code)
	}

	rec3 := httptest.NewRecorder()
	srv.ServeHTTP(rec3, httptest.NewRequest(http.MethodPost, "/api/incidents/nope/resolve", nil))
	if rec3.Code != http.StatusNotFound {
		t.Errorf("resolve missing: got %d want 404", rec3.Code)
	}
}

func TestCORSPreflight(t *testing.T) {
	srv := testServer(t)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest(http.MethodOptions, "/api/fleet", nil))
	if rec.Code != http.StatusNoContent {
		t.Errorf("preflight: got %d want 204", rec.Code)
	}
	if rec.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("missing CORS header")
	}
}
