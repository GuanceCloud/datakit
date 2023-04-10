lint: lint_deps
	golangci-lint run --fix | tee lint.err # https://golangci-lint.run/usage/install/#local-installation

lint_deps: gofmt vet

vet:
	@go vet ./...

gofmt:
	@GO111MODULE=off gofmt -l $(shell find . -type f -name '*.go'| grep -v "/vendor/\|/.git/")

copyright_check:
	@python3 copyright.py --dry-run && \
		{ echo "copyright check ok"; exit 0; } || \
		{ echo "copyright check failed"; exit -1; }

copyright_check_auto_fix:
	@python3 copyright.py --fix
