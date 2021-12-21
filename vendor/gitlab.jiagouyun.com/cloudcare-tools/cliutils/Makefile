lint: lint_deps
	@golangci-lint run | tee check.err # https://golangci-lint.run/usage/install/#local-installation

lint_deps: gofmt vet

vet:
	@go vet ./...

gofmt:
	@GO111MODULE=off gofmt -l $(shell find . -type f -name '*.go'| grep -v "/vendor/\|/.git/")
