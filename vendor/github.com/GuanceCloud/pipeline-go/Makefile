.PHONY: test lint commit

pre-commit: test lint

test:
	mkdir -p dist/coverprofile
	go test -cover ./... |tee -a dist/coverprofile/coverage.out

lint:
	golangci-lint run