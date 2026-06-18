package store

import (
	"testing"

	"github.com/maisymylod/groundstation-console/internal/telemetry"
)

func seeded(t testing.TB, perSat int) *Store {
	t.Helper()
	st := New(0)
	for _, f := range telemetry.GenerateFixture(42, perSat) {
		st.Ingest(f)
	}
	return st
}

func TestFleetSummaryMatchesNaive(t *testing.T) {
	st := seeded(t, 200)
	fast := st.FleetSummary()
	slow := st.NaiveFleetSummary()
	if len(fast) != len(slow) {
		t.Fatalf("length mismatch: indexed=%d naive=%d", len(fast), len(slow))
	}
	for i := range fast {
		if fast[i].SatID != slow[i].SatID {
			t.Fatalf("sat order mismatch at %d: %q vs %q", i, fast[i].SatID, slow[i].SatID)
		}
		if fast[i].Frames != slow[i].Frames {
			t.Errorf("%s frames: indexed=%d naive=%d", fast[i].SatID, fast[i].Frames, slow[i].Frames)
		}
		if fast[i].OpenIncidents != slow[i].OpenIncidents {
			t.Errorf("%s open incidents: indexed=%d naive=%d", fast[i].SatID, fast[i].OpenIncidents, slow[i].OpenIncidents)
		}
		if fast[i].WorstSeverity != slow[i].WorstSeverity {
			t.Errorf("%s worst severity: indexed=%v naive=%v", fast[i].SatID, fast[i].WorstSeverity, slow[i].WorstSeverity)
		}
	}
}

func TestIncidentsAndResolve(t *testing.T) {
	st := seeded(t, 200)
	open := st.Incidents(1000, true)
	if len(open) == 0 {
		t.Fatal("expected at least one open incident in the fixture")
	}
	first := open[0]
	if !st.Resolve(first.ID) {
		t.Fatalf("resolve %s failed", first.ID)
	}
	if st.Resolve(first.ID) {
		t.Fatalf("second resolve of %s should fail", first.ID)
	}
	openAfter := st.Incidents(1000, true)
	if len(openAfter) != len(open)-1 {
		t.Errorf("open count after resolve: got %d want %d", len(openAfter), len(open)-1)
	}
}

func TestWindowBounds(t *testing.T) {
	st := seeded(t, 200)
	w := st.Window("HS-01", 50)
	if len(w.Frames) != 50 {
		t.Errorf("window length: got %d want 50", len(w.Frames))
	}
	if w.SatID != "HS-01" {
		t.Errorf("window sat: got %q", w.SatID)
	}
}

func TestCapPerSat(t *testing.T) {
	st := New(10)
	for _, f := range telemetry.GenerateFixture(42, 100) {
		st.Ingest(f)
	}
	w := st.Window("HS-01", 1000)
	if len(w.Frames) != 10 {
		t.Errorf("capped window: got %d want 10", len(w.Frames))
	}
}

// BenchmarkFleetSummaryNaive measures the pre-optimization fleet rollup that
// rescans the whole frame log on each request.
func BenchmarkFleetSummaryNaive(b *testing.B) {
	st := seeded(b, 2000)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = st.NaiveFleetSummary()
	}
}

// BenchmarkFleetSummaryIndexed measures the optimized fleet rollup served from
// the maintained per-satellite index.
func BenchmarkFleetSummaryIndexed(b *testing.B) {
	st := seeded(b, 2000)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = st.FleetSummary()
	}
}
