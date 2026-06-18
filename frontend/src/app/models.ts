// Domain types mirroring the Go console API (internal/telemetry, internal/store).

export interface SatSummary {
  sat_id: string;
  frames: number;
  open_incidents: number;
  worst_severity: number;
  last_battery_v: number;
  last_temp_c: number;
  last_signal_dbm: number;
}

export interface Frame {
  sat_id: string;
  timestamp: string;
  battery_v: number;
  temp_c: number;
  signal_dbm: number;
  attitude_err: number;
}

export interface Window {
  sat_id: string;
  frames: Frame[];
}

export interface Incident {
  id: string;
  sat_id: string;
  type: string;
  severity: number;
  summary: string;
  detected_at: string;
  resolved: boolean;
}

export const SEVERITY_LABELS = ['nominal', 'info', 'warning', 'critical'];

export function severityClass(sev: number): string {
  switch (sev) {
    case 3:
      return 'critical';
    case 2:
      return 'warning';
    default:
      return 'nominal';
  }
}
