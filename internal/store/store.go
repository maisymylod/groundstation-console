// Package store holds the in-memory telemetry/incident state the API serves.
// It is concurrency-safe and maintains per-satellite indexes so the hot read
// paths (fleet summary, per-satellite window) do not rescan the whole frame log.
package store

import (
	"sort"
	"sync"

	"github.com/maisymylod/groundstation-console/internal/telemetry"
)

// Store is the authoritative in-memory state for the console. Frames arrive from
// the fixture loader or the live broker consumer; incidents are derived on
// ingest. Reads are served from maintained indexes.
type Store struct {
	mu        sync.RWMutex
	frames    map[string][]telemetry.Frame  // sat_id -> frames, append order
	incidents []telemetry.Incident          // newest last
	byType    map[telemetry.AnomalyType]int // maintained counts
	bySat     map[string]*SatSummary        // maintained per-sat rollup
	capPerSat int
}

// SatSummary is the per-satellite rollup the fleet view renders. It is updated
// incrementally on ingest so GET /api/fleet is O(satellites), not O(frames).
type SatSummary struct {
	SatID         string             `json:"sat_id"`
	Frames        int                `json:"frames"`
	OpenIncidents int                `json:"open_incidents"`
	WorstSeverity telemetry.Severity `json:"worst_severity"`
	LastBatteryV  float64            `json:"last_battery_v"`
	LastTempC     float64            `json:"last_temp_c"`
	LastSignalDBm float64            `json:"last_signal_dbm"`
}

// New returns an empty store. capPerSat bounds retained frames per satellite
// (0 means unbounded); the API uses a bound so the live consumer cannot grow
// memory without limit.
func New(capPerSat int) *Store {
	return &Store{
		frames:    make(map[string][]telemetry.Frame),
		byType:    make(map[telemetry.AnomalyType]int),
		bySat:     make(map[string]*SatSummary),
		capPerSat: capPerSat,
	}
}

// Ingest records a frame, derives any incident, and updates the indexes. This
// keeps reads cheap at the cost of a little work per write.
func (s *Store) Ingest(f telemetry.Frame) (telemetry.Incident, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	buf := append(s.frames[f.SatID], f)
	if s.capPerSat > 0 && len(buf) > s.capPerSat {
		buf = buf[len(buf)-s.capPerSat:]
	}
	s.frames[f.SatID] = buf

	sum := s.bySat[f.SatID]
	if sum == nil {
		sum = &SatSummary{SatID: f.SatID}
		s.bySat[f.SatID] = sum
	}
	sum.Frames = len(buf)
	sum.LastBatteryV = f.BatteryV
	sum.LastTempC = f.TempC
	sum.LastSignalDBm = f.SignalDBm

	inc, ok := telemetry.Detect(f)
	if ok {
		s.incidents = append(s.incidents, inc)
		s.byType[inc.Type]++
		sum.OpenIncidents++
		if inc.Severity > sum.WorstSeverity {
			sum.WorstSeverity = inc.Severity
		}
	}
	return inc, ok
}

// FleetSummary returns the per-satellite rollup, sorted by sat id. O(satellites).
func (s *Store) FleetSummary() []SatSummary {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]SatSummary, 0, len(s.bySat))
	for _, v := range s.bySat {
		out = append(out, *v)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].SatID < out[j].SatID })
	return out
}

// Window returns the most recent n frames for one satellite.
func (s *Store) Window(satID string, n int) telemetry.Window {
	s.mu.RLock()
	defer s.mu.RUnlock()
	all := s.frames[satID]
	if n > 0 && len(all) > n {
		all = all[len(all)-n:]
	}
	frames := make([]telemetry.Frame, len(all))
	copy(frames, all)
	return telemetry.Window{SatID: satID, Frames: frames}
}

// Incidents returns up to limit incidents, newest first. open=true filters to
// unresolved.
func (s *Store) Incidents(limit int, openOnly bool) []telemetry.Incident {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]telemetry.Incident, 0, limit)
	for i := len(s.incidents) - 1; i >= 0 && len(out) < limit; i-- {
		if openOnly && s.incidents[i].Resolved {
			continue
		}
		out = append(out, s.incidents[i])
	}
	return out
}

// Resolve marks an incident resolved and decrements the satellite open count.
func (s *Store) Resolve(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.incidents {
		if s.incidents[i].ID == id && !s.incidents[i].Resolved {
			s.incidents[i].Resolved = true
			if sum := s.bySat[s.incidents[i].SatID]; sum != nil && sum.OpenIncidents > 0 {
				sum.OpenIncidents--
			}
			return true
		}
	}
	return false
}

// TypeCounts returns incident counts by anomaly type.
func (s *Store) TypeCounts() map[telemetry.AnomalyType]int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[telemetry.AnomalyType]int, len(s.byType))
	for k, v := range s.byType {
		out[k] = v
	}
	return out
}
