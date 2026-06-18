#!/usr/bin/env bash
# Regenerate the committed profiling artifacts:
#   1. Go benchmark of the hot fleet-summary read path (naive scan vs indexed),
#      captured raw and summarized with benchstat into bench/.
#   2. Frontend production bundle size, before and after the zoneless optimization,
#      captured into bench/.
#
# Run from the repo root: ./scripts/profile.sh
set -euo pipefail
cd "$(dirname "$0")/.."

echo "== Go benchmark: fleet-summary read path =="
go test -run='^$' -bench=BenchmarkFleetSummary -benchmem -count=8 ./internal/store/ \
  | tee bench/fleet_summary.txt
go run golang.org/x/perf/cmd/benchstat@latest bench/fleet_summary.txt \
  > bench/fleet_summary.benchstat.txt
echo "wrote bench/fleet_summary.txt and bench/fleet_summary.benchstat.txt"

echo
echo "== Frontend bundle size: before vs after (zoneless) =="
cd frontend
npm ci --no-audit --no-fund >/dev/null

# BEFORE: re-add the zone.js polyfill and zone change detection, build, measure.
node ../scripts/toggle-zone.mjs on
npm run build >/tmp/gsc-build-before.log 2>&1
grep -E "Initial total|polyfills" /tmp/gsc-build-before.log | tee ../bench/bundle_before.txt

# AFTER: restore the committed zoneless config, build, measure.
node ../scripts/toggle-zone.mjs off
npm run build >/tmp/gsc-build-after.log 2>&1
grep -E "Initial total|main" /tmp/gsc-build-after.log | tee ../bench/bundle_after.txt

echo
echo "Profiling artifacts refreshed under bench/."
