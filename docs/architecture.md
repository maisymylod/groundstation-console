# Architecture

`groundstation-console` is the operator-facing view over the Heliosnet ground
system. It is two deployable pieces that share one domain model: a Go backend that
owns telemetry/incident state and serves an HTTP/JSON API, and an Angular SPA that
renders the fleet, per-satellite telemetry, and the incident feed from that API.

## Flow

```
            Kafka / Redpanda                          browser
         (heliosnet.telemetry)                   ┌──────────────────┐
                  │                               │  Angular SPA     │
                  ▼                               │  (standalone,    │
        ┌───────────────────┐    GET /api/fleet  │   OnPush+signals,│
        │ broker.Consumer   │◀───────────────────│   zoneless)      │
        │ decode + ingest   │    GET .../window  └────────┬─────────┘
        └─────────┬─────────┘    GET /api/incidents       │ /api
                  │              POST .../resolve          ▼
                  ▼                               ┌──────────────────┐
        ┌───────────────────┐                    │  api.Server      │
        │ store.Store       │◀───────────────────│  net/http +      │
        │ frames + indexes  │                    │  ServeMux routes │
        │ incidents         │                    └──────────────────┘
        └─────────┬─────────┘
                  │ telemetry.Detect (shared anomaly taxonomy)
                  ▼
        nominal / drift / dropout / thermal_runaway
```

When no broker is configured the backend preloads a deterministic fixture
(`telemetry.GenerateFixture`, seeded) and serves entirely from memory, so the
console runs from a clean clone with one command. Setting `BROKER_SEEDS` switches
on the live consumer; both paths feed the same `store.Store`.

## Components

### Backend (Go)
- `internal/telemetry`: the domain model (`Frame`, `Incident`, `Severity`), the
  rule-based `Detect` classifier over the shared anomaly taxonomy, and the seeded
  fixture generator. The taxonomy (nominal / drift / dropout / thermal_runaway)
  matches the rest of the suite.
- `internal/store`: concurrency-safe in-memory state. Maintains per-satellite
  summary indexes on ingest so the hot read paths do not rescan the frame log.
  Ships both the optimized `FleetSummary` and the pre-optimization
  `NaiveFleetSummary` kept for the benchmark.
- `internal/api`: `net/http` server (Go 1.22+ method-and-path routing), JSON
  responses, request logging via `log/slog`, CORS for the SPA dev server.
- `internal/broker`: `franz-go` consumer that decodes telemetry frames off
  Kafka/Redpanda and ingests them, plus a producer that streams the fixture for
  the compose demo.
- `internal/config`: environment-driven config with runnable defaults.
- `cmd/console`: wires it together; `console` serves, `console produce` streams.

### Frontend (Angular)
- Standalone components only, no NgModules. `AppComponent` renders the fleet grid,
  a per-satellite telemetry panel of dependency-free SVG sparklines, and the
  incident table with resolve actions.
- `ConsoleService` reads the backend API and falls back to a small bundled fixture
  when the backend is unreachable, so the view always renders.
- Zoneless change detection (`OnPush` + signals); see docs/profiling.md.

## Build and deploy
- Local: `go build/test` and `npm ci && npm run build` (both verified).
- CI also builds the Go backend with Bazel (`bazelisk`) for a hermetic build.
- `deploy/terraform` manages the Kubernetes Deployment/Service and the Grafana SLO
  dashboard and alert rules as code; `deploy/grafana` holds the dashboard JSON.
