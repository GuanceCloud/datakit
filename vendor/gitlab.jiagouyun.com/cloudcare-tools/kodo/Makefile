.PHONY: all

default: all
	@GO111MODULE=off go build -v -o ./image/kodo cmd/kodo/main.go
	@GO111MODULE=off go build -v -o ./image/kodows cmd/kodows/main.go
	@strip image/kodo 

check:
	@go vet ./... 
	@golangci-lint run --timeout 1h # go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.27.0
	@staticcheck -tests ./...  # go get honnef.co/go/tools/...
