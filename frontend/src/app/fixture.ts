// Small bundled fixture so the ops view renders when the backend is offline.
// It is a representative subset of what the Go console serves from its own seeded
// fixture; the labels here are illustrative, not ground truth.
import { Frame, Incident, SatSummary, Window } from './models';

export const FIXTURE_FLEET: SatSummary[] = [
  { sat_id: 'HS-01', frames: 240, open_incidents: 0, worst_severity: 0, last_battery_v: 27.1, last_temp_c: 31.2, last_signal_dbm: -77.9 },
  { sat_id: 'HS-02', frames: 240, open_incidents: 12, worst_severity: 3, last_battery_v: 26.9, last_temp_c: 71.4, last_signal_dbm: -78.3 },
  { sat_id: 'HS-03', frames: 240, open_incidents: 14, worst_severity: 2, last_battery_v: 22.8, last_temp_c: 30.6, last_signal_dbm: -78.1 },
  { sat_id: 'HS-04', frames: 240, open_incidents: 10, worst_severity: 3, last_battery_v: 27.0, last_temp_c: 30.9, last_signal_dbm: -97.5 },
  { sat_id: 'HS-05', frames: 240, open_incidents: 33, worst_severity: 2, last_battery_v: 27.2, last_temp_c: 31.0, last_signal_dbm: -78.0 },
  { sat_id: 'HS-06', frames: 240, open_incidents: 0, worst_severity: 0, last_battery_v: 27.0, last_temp_c: 30.8, last_signal_dbm: -78.2 },
];

export const FIXTURE_INCIDENTS: Incident[] = [
  { id: 'inc-HS-02-1', sat_id: 'HS-02', type: 'thermal_runaway', severity: 3, summary: 'temperature 71.4C exceeds runaway threshold 70.0C', detected_at: '2026-06-18T01:55:00Z', resolved: false },
  { id: 'inc-HS-04-1', sat_id: 'HS-04', type: 'dropout', severity: 3, summary: 'signal -97.5dBm below link floor -95.0dBm', detected_at: '2026-06-18T01:30:00Z', resolved: false },
  { id: 'inc-HS-03-1', sat_id: 'HS-03', type: 'dropout', severity: 2, summary: 'battery 22.80V below safe floor 23.50V', detected_at: '2026-06-18T01:20:00Z', resolved: false },
  { id: 'inc-HS-05-1', sat_id: 'HS-05', type: 'drift', severity: 2, summary: 'attitude error 0.061 exceeds drift threshold 0.050', detected_at: '2026-06-18T01:10:00Z', resolved: false },
];

export function fixtureWindow(satId: string, n: number): Window {
  const frames: Frame[] = [];
  const start = Date.parse('2026-06-18T00:00:00Z');
  for (let i = 0; i < n; i++) {
    frames.push({
      sat_id: satId,
      timestamp: new Date(start + i * 30000).toISOString(),
      battery_v: 27 + Math.sin(i / 9) * 0.3,
      temp_c: 30 + Math.sin(i / 8) * 4,
      signal_dbm: -78 + Math.cos(i / 11) * 1.5,
      attitude_err: Math.abs(Math.sin(i / 13)) * 0.02,
    });
  }
  return { sat_id: satId, frames };
}
