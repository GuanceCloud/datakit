#!/usr/bin/env bash
set -euo pipefail

go test ./internal/plugins/inputs/awslambda/...
echo
echo "Input trace snapshots (if generated):"
echo "  internal/plugins/inputs/awslambda/test/test.output/input-tracing-points.ndjson"
echo "  internal/plugins/inputs/awslambda/test/test.output/input-ddtrace-points.ndjson"
echo "View summary with:"
echo "  bash internal/plugins/inputs/awslambda/test/summary.sh"
