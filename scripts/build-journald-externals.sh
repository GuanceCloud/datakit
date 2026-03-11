#!/bin/bash
# Backward compatibility wrapper for build-journald-externals.sh
# This script calls the generic build-external-collector.sh with journald defaults

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "Note: build-journald-externals.sh is deprecated."
echo "      Using build-external-collector.sh --input journald"
echo

# Call the new generic script with journald defaults
exec "${SCRIPT_DIR}/build-external-collector.sh" --input journald "$@"
