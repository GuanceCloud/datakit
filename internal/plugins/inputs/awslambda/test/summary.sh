#!/usr/bin/env bash
set -euo pipefail

go run ./internal/plugins/inputs/awslambda/test/cmd/trace_summary "$@"
