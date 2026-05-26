.PHONY: lint test pre-commit test-cover

pre-commit: lint test-cover

test:
	go test -count=1 -v ./... 2>&1| tee test.out

test-cover:
	mkdir -p dist/coverprofile
	go test -count=1 -coverprofile=dist/coverprofile/coverage.out -covermode=atomic ./...

lint:
	golangci-lint run
