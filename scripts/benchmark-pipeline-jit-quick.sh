#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
: "${PLATYPUS_JIT_RUNTIME:?set PLATYPUS_JIT_RUNTIME to the candidate native library}"
export PIPELINE_SCRIPT_DIRECTORY=${PIPELINE_SCRIPT_DIRECTORY:-"$PWD/internal/pipeline/testdata/jit-real-scripts"}
output=${JIT_BENCH_OUTPUT:-"/tmp/datakit-jit-quick"}
mkdir -p "$output"
go test -mod=vendor -tags 'pipeline_jit jitbench' ./internal/pipeline \
  -run '^$' -c -o "$output/pipeline.test"
{
  date -u +%FT%TZ
  git rev-parse HEAD
  git diff --stat
  go version
  sha256sum "$PLATYPUS_JIT_RUNTIME" "$output/pipeline.test"
} > "$output/metadata.txt"
git diff --binary > "$output/source.diff"
command=("$output/pipeline.test" -test.run '^$' \
  -test.bench "${JIT_BENCH_FILTER:-^BenchmarkPipelineJITAdapter/batch-(1|10)/(e2e-group|projection-query)$}" \
  -test.benchtime "${JIT_BENCH_TIME:-200ms}" -test.count "${JIT_BENCH_COUNT:-3}" -test.cpu 1)
if [[ -n ${JIT_BENCH_CPU:-} ]]; then command=(taskset -c "$JIT_BENCH_CPU" "${command[@]}"); fi
"${command[@]}" | tee "$output/benchmark.txt"
