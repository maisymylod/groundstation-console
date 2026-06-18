package store

import (
	"sort"

	"github.com/maisymylod/groundstation-console/internal/telemetry"
)

// NaiveFleetSummary recomputes the fleet rollup by rescanning the entire frame
// log and re-deriving every incident on each call. This is the pre-optimization
// implementation kept for the committed benchmark (bench/) that proves the
// indexed FleetSummary speedup is real and reproducible. It is not used by the
// running server.
func (s *Store) NaiveFleetSummary() []SatSummary {
	s.mu.RLock()
	defer s.mu.RUnlock()

	bySat := make(map[string]*SatSummary, len(s.frames))
	for sat, frames := range s.frames {
		sum := &SatSummary{SatID: sat, Frames: len(frames)}
		for _, f := range frames {
			sum.LastBatteryV = f.BatteryV
			sum.LastTempC = f.TempC
			sum.LastSignalDBm = f.SignalDBm
			if inc, ok := telemetry.Detect(f); ok {
				sum.OpenIncidents++
				if inc.Severity > sum.WorstSeverity {
					sum.WorstSeverity = inc.Severity
				}
			}
		}
		bySat[sat] = sum
	}

	out := make([]SatSummary, 0, len(bySat))
	for _, v := range bySat {
		out = append(out, *v)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].SatID < out[j].SatID })
	return out
}
