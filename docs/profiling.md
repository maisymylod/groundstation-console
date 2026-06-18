# Profiling: before and after

Two optimizations are measured here, both with reproducible scripts and committed
raw artifacts. Regenerate everything with `./scripts/profile.sh` (or `make profile`).

All numbers below were measured on the development machine (Apple M2 Pro, Go
1.26.4, Node 25.9.0, Angular 18.2). Absolute figures will differ on other
hardware; the relative improvement is the point and is reproducible.

## 1. Backend hot path: fleet-summary read

`GET /api/fleet` is the endpoint the ops console polls most often. The first
implementation (`store.NaiveFleetSummary`) recomputed the per-satellite rollup on
every request by rescanning the entire frame log and re-deriving every incident.
The optimization maintains a per-satellite summary index incrementally on ingest
(`store.FleetSummary`), so the read is O(satellites) instead of O(frames). A test
(`TestFleetSummaryMatchesNaive`) asserts the two produce identical output, so the
speedup is not bought with wrong answers.

Benchmark: `BenchmarkFleetSummaryNaive` vs `BenchmarkFleetSummaryIndexed`, 6
satellites x 2000 frames, `-count=8`. Raw output in
[`bench/fleet_summary.txt`](../bench/fleet_summary.txt); benchstat summary in
[`bench/fleet_summary.benchstat.txt`](../bench/fleet_summary.benchstat.txt).

| Implementation | Time/op | Bytes/op | Allocs/op |
|---|---|---|---|
| Naive (rescan + re-derive) | 594.5 µs ± 9% | 196.8 KiB | 8510 |
| Indexed (maintained rollup) | 344.9 ns ± 6% | 520 B | 4 |
| **Improvement** | **~1720x faster** | **~388x less** | **~2100x fewer** |

The indexed read moves the per-frame work to ingest time (paid once per frame)
instead of repeating it on every read (paid on every request), which is the right
trade for a read-heavy console.

## 2. Frontend production bundle: zoneless change detection

Every console component is `OnPush` and all view state lives in Angular signals,
so the app does not need Zone.js. Switching the bootstrap to
`provideExperimentalZonelessChangeDetection()` and dropping the `zone.js` polyfill
removes the entire polyfills chunk from the production bundle.

Measured from the real `ng build` output (production configuration). Raw output in
[`bench/bundle_before.txt`](../bench/bundle_before.txt) and
[`bench/bundle_after.txt`](../bench/bundle_after.txt). The `scripts/profile.sh`
run toggles the config to the pre-optimization state, builds, then restores the
committed zoneless state and builds again.

| Build | Initial total (raw) | Initial total (transfer) | Polyfills chunk |
|---|---|---|---|
| Before (zone.js) | 196.99 kB | 58.60 kB | 34.52 kB / 11.28 kB transfer |
| After (zoneless) | 162.37 kB | 47.29 kB | none |
| **Saving** | **34.62 kB (17.6%)** | **11.31 kB (19.3%)** | **eliminated** |

The dependency set in `frontend/package.json` was also trimmed to only the
packages the bundle actually imports (`@angular/common`, `@angular/compiler`,
`@angular/core`, `@angular/platform-browser`), removing unused `animations`,
`forms`, `router`, and `platform-browser-dynamic` runtime deps.
