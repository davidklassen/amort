.PHONY: build lint test tidy

build:
	go build -o amort .

tidy:
	gofmt -w .
	go mod tidy

lint:
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest run ./...

test:
	go test ./...
