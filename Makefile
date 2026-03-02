.PHONY: build test clean run-agentctl run-agentd

BINARY_AGENTCTL := bin/agentctl
BINARY_AGENTD := bin/agentd
GO := go
GOFLAGS := -v

all: build

build: build-agentctl build-agentd

build-agentctl:
	@mkdir -p bin
	$(GO) build $(GOFLAGS) -o $(BINARY_AGENTCTL) ./cmd/agentctl

build-agentd:
	@mkdir -p bin
	$(GO) build $(GOFLAGS) -o $(BINARY_AGENTD) ./cmd/agentd

test:
	$(GO) test ./... -race -coverprofile=coverage.out

test-short:
	$(GO) test ./... -short

coverage: test
	$(GO) tool cover -html=coverage.out -o coverage.html

clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

run-agentctl: build-agentctl
	./$(BINARY_AGENTCTL) $(ARGS)

run-agentd: build-agentd
	./$(BINARY_AGENTD) $(ARGS)

fmt:
	$(GO) fmt ./...

vet:
	$(GO) vet ./...

lint: fmt vet
