.DEFAULT_GOAL := build

GOEXE := $(shell go env GOEXE)

.PHONY: fmt lint test vuln build replay

fmt:
	gofmt -w cmd internal prompts
	goimports -w cmd internal prompts

lint:
	go vet ./...
	golangci-lint run

test:
	go test -race -coverprofile=cover.out ./...

vuln:
	govulncheck ./...

build:
	go build -trimpath -o bin/agent$(GOEXE) ./cmd/agent

replay:
	$(error Offline replay is not implemented yet; see T17)
