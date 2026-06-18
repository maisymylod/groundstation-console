package telemetry

import (
	"fmt"
	"sync/atomic"
)

// thresholds for the rule-based detector. These mirror the severity thresholds
// the groundstation agent graph applies; they are deliberately simple and
// documented as such in docs/assumptions.md.
const (
	battLowV       = 23.5 // dropout territory below this
	tempHighC      = 55.0 // thermal warning above this
	tempCriticalC  = 70.0 // thermal runaway above this
	signalFloorDBm = -95.0
	attitudeDrift  = 0.05
)

var incidentSeq uint64

// Detect classifies a single frame into the shared anomaly taxonomy and returns
// an incident when the frame is not nominal. The boolean reports whether an
// incident was produced. This is the hot path the API benchmark exercises.
func Detect(f Frame) (Incident, bool) {
	typ, sev, summary := classify(f)
	if typ == AnomalyNominal {
		return Incident{}, false
	}
	id := fmt.Sprintf("inc-%s-%d", f.SatID, atomic.AddUint64(&incidentSeq, 1))
	return Incident{
		ID:         id,
		SatID:      f.SatID,
		Type:       typ,
		Severity:   sev,
		Summary:    summary,
		DetectedAt: f.Timestamp,
	}, true
}

func classify(f Frame) (AnomalyType, Severity, string) {
	switch {
	case f.TempC >= tempCriticalC:
		return AnomalyThermalRunaway, SeverityCritical,
			fmt.Sprintf("temperature %.1fC exceeds runaway threshold %.1fC", f.TempC, tempCriticalC)
	case f.SignalDBm <= signalFloorDBm:
		return AnomalyDropout, SeverityCritical,
			fmt.Sprintf("signal %.1fdBm below link floor %.1fdBm", f.SignalDBm, signalFloorDBm)
	case f.BatteryV <= battLowV:
		return AnomalyDropout, SeverityWarning,
			fmt.Sprintf("battery %.2fV below safe floor %.2fV", f.BatteryV, battLowV)
	case f.TempC >= tempHighC:
		return AnomalyThermalRunaway, SeverityWarning,
			fmt.Sprintf("temperature %.1fC above warning threshold %.1fC", f.TempC, tempHighC)
	case f.AttitudeErr >= attitudeDrift:
		return AnomalyDrift, SeverityWarning,
			fmt.Sprintf("attitude error %.3f exceeds drift threshold %.3f", f.AttitudeErr, attitudeDrift)
	default:
		return AnomalyNominal, SeverityNominal, ""
	}
}
