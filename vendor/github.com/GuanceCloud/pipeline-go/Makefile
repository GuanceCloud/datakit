.PHONY: lint test pre-commit test-cover

pre-commit: lint test-cover

test:
	go test -v ./... 2>&1| tee test.out

test-cover:
	mkdir -p dist/coverprofile
	go test -coverprofile=dist/coverprofile/coverage.out -covermode=atomic ./...

lint:
	golangci-lint run
