// Package telemetry defines the domain model the ops console renders: satellite
// telemetry frames and the incidents derived from them. The taxonomy mirrors the
// anomaly types used across the Heliosnet suite (constellation telemetry plane,
// groundstation agent graph) so a frame produced upstream reads the same here.
package telemetry

import "time"

// Severity ranks an incident. Higher ordinal is more urgent.
type Severity int

const (
	SeverityNominal Severity = iota
	SeverityInfo
	SeverityWarning
	SeverityCritical
)

func (s Severity) String() string {
	switch s {
	case SeverityCritical:
		return "critical"
	case SeverityWarning:
		return "warning"
	case SeverityInfo:
		return "info"
	default:
		return "nominal"
	}
}

// AnomalyType is the fault taxonomy shared with the rest of the suite.
type AnomalyType string

const (
	AnomalyNominal        AnomalyType = "nominal"
	AnomalyDrift          AnomalyType = "drift"
	AnomalyDropout        AnomalyType = "dropout"
	AnomalyThermalRunaway AnomalyType = "thermal_runaway"
)

// Frame is one telemetry sample from one satellite at one instant. It is the
// unit pushed over the broker and the unit the console charts.
type Frame struct {
	SatID       string    `json:"sat_id"`
	Timestamp   time.Time `json:"timestamp"`
	BatteryV    float64   `json:"battery_v"`
	TempC       float64   `json:"temp_c"`
	SignalDBm   float64   `json:"signal_dbm"`
	AttitudeErr float64   `json:"attitude_err"`
}

// Incident is a fault detected from a frame, surfaced in the console incident
// feed. Resolved tracks the human-in-the-loop acknowledgement.
type Incident struct {
	ID         string      `json:"id"`
	SatID      string      `json:"sat_id"`
	Type       AnomalyType `json:"type"`
	Severity   Severity    `json:"severity"`
	Summary    string      `json:"summary"`
	DetectedAt time.Time   `json:"detected_at"`
	Resolved   bool        `json:"resolved"`
}

// Window is a span of frames for one satellite, used by the charts endpoint.
type Window struct {
	SatID  string  `json:"sat_id"`
	Frames []Frame `json:"frames"`
}
