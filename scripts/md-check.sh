# check single markdown on spell and markdown lint
# Mon May 12 17:38:55 CST 2025
# Usage:
#    ./scripts/md-check.sh path/to/xxx/md

go run cmd/make/make.go --mdcheck-no-section-check --mdcheck-no-autofix true --mdcheck $1

cspell lint --show-suggestions \
	-c scripts/cspell.json \
	--no-progress $1

if markdownlint -c scripts/markdownlint.yml $1; then \
	printf "markdownlint check ok\n"; \
else
	printf "markdownlint check failed\n"; \
	exit -1; 
fi
