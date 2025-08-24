# CLAUDE.md

## Project Context

**Claude Code SDK for Go** - Provides programmatic interaction with Claude Code CLI through `Query()` (one-shot) and `Client` (streaming) APIs with 100% Python SDK parity.

**Module**: `github.com/severity1/claude-code-sdk-go`
**Package**: `claudecode`
**Import**: `import "github.com/severity1/claude-code-sdk-go"`

## Development Commands

```bash
# Build and test
go build ./...                    # Build all packages
go test ./...                     # Run all tests with proper output
go test -race ./...               # Race condition detection
go test -cover ./...              # Coverage analysis

# Code quality (run before commits)
go fmt ./...                      # Format code
go vet ./...                      # Static analysis
golangci-lint run                 # Comprehensive linting
```

## Code Style & Conventions

**Idiomatic Go**: Write Go-native code using `gofmt` formatting with no exceptions. Follow standard Go naming conventions and prefer Go idioms over Python patterns.

**Interface-Driven Design**: All message types implement `Message`, all content blocks implement `ContentBlock`. Use interfaces for testability and Go-native polymorphism.

**Error Handling**: Use structured error types with `fmt.Errorf` and `%w` verb for wrapping. Include contextual information (exit codes, file paths). Follow Go error handling patterns, not exceptions.

**Context-First**: All functions that can block should accept `context.Context` as first parameter for cancellation and timeouts (Go-native concurrency pattern).

**JSON Handling**: Use custom `UnmarshalJSON` methods for union types, discriminate on `"type"` field with `json.RawMessage` for delayed parsing (Go-native JSON handling).

## Critical Implementation Notes

**Transport Interface**: Central abstraction for CLI communication. Use `MockTransport` for unit tests, real subprocess for integration tests.

**Process Cleanup**: Follow exact pattern: SIGTERM → wait 5 seconds → SIGKILL for proper resource cleanup.

**Buffer Protection**: Implement 1MB buffer limit with overflow protection to prevent memory exhaustion attacks.

**Environment Variables**: Set `CLAUDE_CODE_ENTRYPOINT` to `"sdk-go"` or `"sdk-go-client"` to identify SDK to CLI.

## Development Methodology

**TDD Approach**: Follow RED-GREEN-BLUE cycle strictly. Write failing tests first, implement to make them pass. Never use dummy code or placeholder implementations.

**Component Memory**: This project uses auto-discovery memory system with component-specific CLAUDE.md files in subdirectories.

## Project Documentation

**Technical Specifications**: See [SPECIFICATIONS.md](SPECIFICATIONS.md) for complete API specifications, type definitions, and behavioral requirements.

**Implementation Analysis**: The [analysis/](analysis/) directory contains detailed analysis of the reference Claude Code SDK for Python, implementation guides, transport layer specs, usage patterns, and architectural decisions.

**Task Tracking**: See [TDD_IMPLEMENTATION_TASKS.md](TDD_IMPLEMENTATION_TASKS.md) for current development progress, test requirements, and completion status following Test-Driven Development methodology.