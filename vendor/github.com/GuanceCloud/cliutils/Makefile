.PHONY: test

GOLINT_BINARY ?= golangci-lint
LINT_FIX      ?= true

lint: lint_deps
	@$(GOLINT_BINARY) --version
ifeq ($(LINT_FIX),true)
		@printf "lint with fix...\n"; \
		$(GOLINT_BINARY) run --fix;
else
		@printf "lint without fix...\n"; \
		$(GOLINT_BINARY) run;
endif

	@if [ $$? != 0 ]; then \
		printf "lint failed\n"; \
		exit -1; \
	else \
		printf "lint ok\n"; \
	fi

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
