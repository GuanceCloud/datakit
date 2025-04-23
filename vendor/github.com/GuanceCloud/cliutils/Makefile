GOLINT_BINARY         ?= golangci-lint

lint: lint_deps
	@$(GOLINT_BINARY) --version
	@$(GOLINT_BINARY) run --fix | tee lint.err # https://golangci-lint.run/usage/install/#local-installation

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

test:
		LOGGER_PATH=nul CGO_CFLAGS=-Wno-undef-prefix go test -test.v -timeout 99999m -cover ./...

show_metrics:
	@promlinter list . --add-help -o md --with-vendor
