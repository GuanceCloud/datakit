#!/usr/bin/env bash
# The acceptance boundary is the real DK group path, including transport/apply.
set -euo pipefail
cd "$(dirname "$0")/.."
mode=${1:-quick}
case "$mode" in
  quick) batches='001|010'; duration=100ms; samples=3 ;;
  accept) batches='001|004|010|032'; duration=500ms; samples=7 ;;
  *) echo 'usage: benchmark-pipeline-jit-go.sh [quick|accept]' >&2; exit 1 ;;
esac
export JIT_BENCH_OUTPUT=${JIT_BENCH_OUTPUT:-"/tmp/datakit-jit-go-$mode"}
export JIT_BENCH_TIME=${JIT_BENCH_TIME:-$duration}
export JIT_BENCH_COUNT=${JIT_BENCH_COUNT:-$samples}
workloads='light-drop-cast|json-3-cast|nginx-native|default-time-native|elasticsearch-native|tdengine-duration|tdengine-error|consul-native'
export JIT_BENCH_FILTER="^BenchmarkPipelineJITRealScriptsCrossover/($workloads)\$/batch-($batches)/(pipeline-go-e2e|jit-group-e2e)\$"
scripts/benchmark-pipeline-jit-quick.sh
python3 scripts/summarize-pipeline-jit-go.py "$JIT_BENCH_OUTPUT/benchmark.txt" "$mode" "$JIT_BENCH_COUNT"
