package telemetry

import (
	"math"
	"math/rand"
	"time"
)

// Fleet is the set of satellites the bundled fixture simulates.
var Fleet = []string{"HS-01", "HS-02", "HS-03", "HS-04", "HS-05", "HS-06"}

// GenerateFixture produces a deterministic span of frames for the whole fleet,
// seeded so the console renders the same data from a clean clone. Some satellites
// are nudged into fault regimes so the incident feed is non-empty.
func GenerateFixture(seed int64, perSat int) []Frame {
	r := rand.New(rand.NewSource(seed))
	start := time.Date(2026, 6, 18, 0, 0, 0, 0, time.UTC)
	frames := make([]Frame, 0, len(Fleet)*perSat)
	for si, sat := range Fleet {
		for i := 0; i < perSat; i++ {
			ts := start.Add(time.Duration(i) * 30 * time.Second)
			f := Frame{
				SatID:       sat,
				Timestamp:   ts,
				BatteryV:    27.0 + r.NormFloat64()*0.4,
				TempC:       30.0 + 4*math.Sin(float64(i)/8) + r.NormFloat64(),
				SignalDBm:   -78.0 + r.NormFloat64()*2,
				AttitudeErr: math.Abs(r.NormFloat64() * 0.01),
			}
			injectFaults(&f, si, i, perSat)
			frames = append(frames, f)
		}
	}
	return frames
}

// injectFaults steers a few satellites into the fault taxonomy so the fixture
// exercises every incident type deterministically.
func injectFaults(f *Frame, satIdx, i, perSat int) {
	switch satIdx {
	case 1: // HS-02: thermal ramp into runaway near the end
		if i > perSat*3/4 {
			f.TempC = 50 + float64(i-perSat*3/4)*1.2
		}
	case 2: // HS-03: battery sag into dropout
		if i%17 == 0 {
			f.BatteryV = 22.8
		}
	case 3: // HS-04: signal dropouts
		if i%23 == 0 {
			f.SignalDBm = -97.5
		}
	case 4: // HS-05: slow attitude drift
		f.AttitudeErr = 0.02 + float64(i)/float64(perSat)*0.06
	}
}
