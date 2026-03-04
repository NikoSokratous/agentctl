.PHONY: build test clean run-agentctl run-agentd generate-operator

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

generate-operator:
	cd k8s/operator && controller-gen object:headerFile="hack/boilerplate.go.txt" paths="./api/v1/..."

showcase-deploy: build-agentctl
	@echo "Deploying showcase enterprise-compliance-bot..."
	@if command -v helm >/dev/null 2>&1 && command -v kubectl >/dev/null 2>&1; then \
		helm upgrade --install agentruntime ./k8s/helm -f ./k8s/helm/values.yaml 2>/dev/null || true; \
		kubectl apply -f ./showcase/enterprise-compliance-bot/k8s/; \
		echo "Showcase deployed. Run: agentctl run --config showcase/enterprise-compliance-bot/agent.yaml --goal '...'"; \
	else \
		echo "Helm/kubectl not found. Use: agentctl run --config showcase/enterprise-compliance-bot/agent.yaml"; \
	fi
