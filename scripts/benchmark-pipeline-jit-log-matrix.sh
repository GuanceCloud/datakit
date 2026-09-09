#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
mode=${1:-quick}
case "$mode" in
  quick) batches='001|010'; batch_list=1,10; duration=100ms; rounds=3 ;;
  full) batches='001|010|032'; batch_list=1,10,32; duration=300ms; rounds=5 ;;
  *) echo 'usage: benchmark-pipeline-jit-log-matrix.sh [quick|full]' >&2; exit 1 ;;
esac
export JIT_BENCH_LOG_MATRIX=1
export JIT_BENCH_LOG_DIRECTORY="$PWD/internal/pipeline/testdata/jit-log-matrix"
export JIT_BENCH_OUTPUT=${JIT_BENCH_OUTPUT:-"/tmp/datakit-jit-log-matrix-$mode"}
export JIT_BENCH_TIME=${JIT_BENCH_TIME:-$duration}
rounds=${JIT_BENCH_COUNT:-$rounds}
[[ $rounds =~ ^[1-9][0-9]*$ ]] || { echo 'sample count must be positive' >&2; exit 1; }
export JIT_BENCH_FILTER="^BenchmarkPipelineJITRealScriptsCrossover/log-/batch-($batches)/(pipeline-go-e2e|jit-group-e2e)\$"
# Repeat the complete matrix, keeping Go/JIT samples adjacent within each round.
# -count N would run all Go repeats before starting the JIT repeats.
JIT_BENCH_COUNT=1 scripts/benchmark-pipeline-jit-quick.sh
{
  printf 'mode=%s rounds=%s duration=%s batches=%s cpu=%s\n' "$mode" "$rounds" "$JIT_BENCH_TIME" "$batch_list" "${JIT_BENCH_CPU:-unbound}"
  sha256sum "$JIT_BENCH_LOG_DIRECTORY"/*
} >> "$JIT_BENCH_OUTPUT/metadata.txt"
command=("$JIT_BENCH_OUTPUT/pipeline.test" -test.run '^$' -test.bench "$JIT_BENCH_FILTER" -test.benchtime "$JIT_BENCH_TIME" -test.count 1 -test.cpu 1)
if [[ -n ${JIT_BENCH_CPU:-} ]]; then command=(taskset -c "$JIT_BENCH_CPU" "${command[@]}"); fi
for ((round=2; round<=rounds; round++)); do
  "${command[@]}" | tee -a "$JIT_BENCH_OUTPUT/benchmark.txt"
done
python3 scripts/summarize-pipeline-jit-log-matrix.py "$JIT_BENCH_OUTPUT/benchmark.txt" "$JIT_BENCH_LOG_DIRECTORY" "$batch_list" "$rounds"
