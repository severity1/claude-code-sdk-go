# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt
GOLINT=golangci-lint

# Build parameters
BINARY_NAME=claude-code-sdk-go
BINARY_UNIX=$(BINARY_NAME)_unix

# Version
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Linker flags
LDFLAGS=-ldflags "-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

.PHONY: all build build-linux test test-verbose test-race test-cover clean deps fmt lint vet check install examples help

all: test build

## Build
build: ## Build the binary
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) -v ./

build-linux: ## Build the binary for Linux
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_UNIX) -v

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
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)
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
install: ## Install the binary
	$(GOCMD) install $(LDFLAGS) ./

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

## Release
release-check: ## Check if ready for release
	@echo "Checking release readiness..."
	@$(MAKE) test
	@$(MAKE) check
	@$(MAKE) examples
	@echo "✅ Ready for release!"

release-dry: ## Dry run release
	goreleaser release --snapshot --clean --skip-publish

## CI/CD helpers
ci: deps-verify test-race check examples ## Run CI pipeline locally

ci-coverage: ## Run CI with coverage
	$(GOTEST) -race -coverprofile=coverage.out ./...

## Docker (if needed in future)
docker-build: ## Build Docker image
	@echo "Docker support not implemented yet"

## Help
help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Default target
.DEFAULT_GOAL := help