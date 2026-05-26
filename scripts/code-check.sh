#!/bin/bash
# lint single golang file
linter=golangci-lint
if [[ -n "{LINTER+x}" ]]; then
	linter=${LINTER}
fi

${linter} run --fix --allow-parallel-runners -v
