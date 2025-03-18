.PHONY: lint test pre-commit test-cover

pre-commit: lint test-cover

test:
	go test ./...

test-cover:
	mkdir -p dist/coverprofile
	go test -coverprofile=dist/coverprofile/coverage.out -covermode=atomic ./...

lint:
	golangci-lint run
