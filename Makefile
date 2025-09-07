# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt
GOLINT=golangci-lint

# Version (for documentation and tooling)
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

.PHONY: all test test-verbose test-race test-cover clean deps fmt lint vet check examples help

all: test

## Test
test: ## Run tests
	$(GOTEST) -v ./...

test-verbose: ## Run tests with verbose output
	$(GOTEST) -v -cover ./...

test-race: ## Run tests with race detection
	$(GOTEST) -race ./...

test-cover: ## Run tests with coverage
	$(GOTEST) -race -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

bench: ## Run benchmarks
	$(GOTEST) -bench=. -benchmem ./...

## Clean
clean: ## Clean build artifacts
	$(GOCLEAN)
	rm -f coverage.out
	rm -f coverage.html

## Dependencies
deps: ## Download dependencies
	$(GOMOD) download

deps-update: ## Update dependencies
	$(GOMOD) tidy

deps-verify: ## Verify dependencies
	$(GOMOD) verify

## Code quality
fmt: ## Format code
	$(GOFMT) -s -w .

fmt-check: ## Check if code is formatted
	@if [ "$$($(GOFMT) -s -l . | wc -l)" -gt 0 ]; then \
		echo "The following files are not formatted with gofmt:"; \
		$(GOFMT) -s -l .; \
		exit 1; \
	fi

lint: ## Run linter
	$(GOLINT) run

vet: ## Run go vet
	$(GOCMD) vet ./...

check: fmt-check vet lint ## Run all checks

## Security
security: ## Run security checks
	$(GOCMD) install golang.org/x/vuln/cmd/govulncheck@latest
	govulncheck ./...

nancy: ## Run Nancy security scan
	$(GOCMD) install github.com/sonatypecommunity/nancy@latest
	$(GOCMD) list -json -deps ./... | nancy sleuth --skip-update-check

## Development
generate: ## Run go generate
	$(GOCMD) generate ./...

## Examples
examples: ## Build all examples
	@for dir in examples/*/; do \
		if [ -f "$$dir/main.go" ]; then \
			echo "Building $$dir"; \
			cd "$$dir" && $(GOBUILD) -v . && cd - > /dev/null; \
		fi; \
	done
	@for dir in examples/*/*/; do \
		if [ -f "$$dir/main.go" ]; then \
			echo "Building $$dir"; \
			cd "$$dir" && $(GOBUILD) -v . && cd - > /dev/null; \
		fi; \
	done

## Documentation
docs: ## Generate documentation
	$(GOCMD) doc -all ./

docs-serve: ## Serve documentation locally
	@echo "Starting documentation server at http://localhost:6060"
	@echo "Visit http://localhost:6060/pkg/github.com/severity1/claude-code-sdk-go/"
	godoc -http=:6060

## SDK/Library specific tasks
sdk-test: ## Test SDK as a consumer would use it
	@echo "=== SDK Consumer Test ==="
	@mkdir -p /tmp/sdk-test
	@cd /tmp/sdk-test && \
	go mod init sdk-consumer-test && \
	echo 'module sdk-consumer-test\n\ngo 1.18\n\nreplace github.com/severity1/claude-code-sdk-go => $(PWD)' > go.mod && \
	echo 'package main\n\nimport (\n\t"context"\n\t"fmt"\n\tclaudecode "github.com/severity1/claude-code-sdk-go"\n)\n\nfunc main() {\n\tctx := context.Background()\n\t_ = claudecode.NewOptions()\n\tfmt.Println("✅ SDK imports work")\n\t_, _ = claudecode.Query(ctx, "test")\n\tfmt.Println("✅ SDK API accessible")\n}' > main.go && \
	go mod tidy && \
	go run main.go && \
	rm -rf /tmp/sdk-test
	@echo "✅ SDK consumer test passed"

api-check: ## Check public API surface
	@echo "=== Public API Surface ==="
	@$(GOCMD) doc -all . | head -50
	@echo ""
	@echo "=== Key Exported Types ==="
	@$(GOCMD) doc Client
	@$(GOCMD) doc Options
	@$(GOCMD) doc Query
	@$(GOCMD) doc WithClient

module-check: ## Check module health
	@echo "=== Module Health Check ==="
	@$(GOMOD) verify
	@$(GOMOD) tidy
	@echo "✅ Module is healthy"

## Release
release-check: ## Check if ready for release
	@echo "Checking release readiness..."
	@$(MAKE) test
	@$(MAKE) check  
	@$(MAKE) examples
	@$(MAKE) sdk-test
	@$(MAKE) api-check
	@$(MAKE) module-check
	@echo "✅ Ready for release!"

release-dry: ## Dry run release
	goreleaser release --snapshot --clean --skip-publish

## CI/CD helpers
ci: deps-verify test-race check examples sdk-test ## Run CI pipeline locally (SDK focused)

ci-coverage: ## Run CI with coverage
	$(GOTEST) -race -coverprofile=coverage.out ./...

## Docker (if needed in future)
docker-build: ## Build Docker image
	@echo "Docker support not implemented yet (library project)"

## Help
help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Default target
.DEFAULT_GOAL := help