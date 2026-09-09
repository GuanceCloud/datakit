#!/usr/bin/env bash
# Run native tests without enabling package staging for ordinary Go builds.
set +x
set -euo pipefail
cd "$(dirname "$0")/.."
case "$(uname -s)/$(uname -m)" in
  Linux/x86_64) arch=amd64 ;;
  Linux/aarch64|Linux/arm64) arch=arm64 ;;
  *) echo 'JIT CI tests require a Linux amd64 or arm64 runner' >&2; exit 1 ;;
esac
PIPELINE_JIT_ENABLED=1 bash scripts/download-pipeline-jit.sh
export PLATYPUS_JIT_RUNTIME="$PLATYPUS_JIT_RUNTIME_DIR/linux-$arch/libplatypus_jit.so"
export PLATYPUS_JIT_EXPECTED_REVISION
PLATYPUS_JIT_EXPECTED_REVISION=$(cat "$PLATYPUS_JIT_RUNTIME_DIR/source-revision")
test -s "$PLATYPUS_JIT_RUNTIME"
make prepare
PIPELINE_JIT_ENABLED=0 go test -tags pipeline_jit ./internal/pipeline/... -count=1 -timeout 10m
