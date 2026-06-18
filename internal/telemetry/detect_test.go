package telemetry

import (
	"testing"
	"time"
)

func TestClassify(t *testing.T) {
	base := Frame{SatID: "HS-01", Timestamp: time.Now(), BatteryV: 27, TempC: 30, SignalDBm: -78, AttitudeErr: 0.001}
	tests := []struct {
		name string
		mut  func(Frame) Frame
		want AnomalyType
		sev  Severity
	}{
		{"nominal", func(f Frame) Frame { return f }, AnomalyNominal, SeverityNominal},
		{"thermal runaway", func(f Frame) Frame { f.TempC = 72; return f }, AnomalyThermalRunaway, SeverityCritical},
		{"thermal warning", func(f Frame) Frame { f.TempC = 58; return f }, AnomalyThermalRunaway, SeverityWarning},
		{"signal dropout", func(f Frame) Frame { f.SignalDBm = -96; return f }, AnomalyDropout, SeverityCritical},
		{"battery sag", func(f Frame) Frame { f.BatteryV = 22; return f }, AnomalyDropout, SeverityWarning},
		{"attitude drift", func(f Frame) Frame { f.AttitudeErr = 0.06; return f }, AnomalyDrift, SeverityWarning},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			typ, sev, _ := classify(tc.mut(base))
			if typ != tc.want {
				t.Errorf("type: got %v want %v", typ, tc.want)
			}
			if sev != tc.sev {
				t.Errorf("severity: got %v want %v", sev, tc.sev)
			}
		})
	}
}

func TestDetectProducesIncident(t *testing.T) {
	f := Frame{SatID: "HS-01", Timestamp: time.Now(), TempC: 80}
	inc, ok := Detect(f)
	if !ok {
		t.Fatal("expected incident")
	}
	if inc.Type != AnomalyThermalRunaway || inc.Severity != SeverityCritical {
		t.Errorf("got %v/%v", inc.Type, inc.Severity)
	}
	if inc.ID == "" {
		t.Error("missing incident id")
	}
}

func TestFixtureDeterministic(t *testing.T) {
	a := GenerateFixture(42, 50)
	b := GenerateFixture(42, 50)
	if len(a) != len(b) {
		t.Fatalf("length mismatch %d vs %d", len(a), len(b))
	}
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("fixture not deterministic at %d", i)
		}
	}
}

func TestSeverityString(t *testing.T) {
	cases := map[Severity]string{
		SeverityNominal: "nominal", SeverityInfo: "info",
		SeverityWarning: "warning", SeverityCritical: "critical",
	}
	for s, want := range cases {
		if s.String() != want {
			t.Errorf("severity %d: got %q want %q", s, s.String(), want)
		}
	}
}
