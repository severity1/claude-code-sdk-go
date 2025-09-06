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

# Specific test patterns (table-driven tests)
go test -v -run TestClient        # Run all client tests (verbose output)
go test -run "TestClient.*Application" # Run specific client test categories
go test -count=3 -run TestClient  # Run client tests multiple times for consistency

# Performance testing
go test -bench=. -benchmem        # Benchmark tests with memory stats
go test -v -count=3 -run TestClient # Performance consistency validation

# Code quality (run before commits)
go fmt ./...                      # Format code
go vet ./...                      # Static analysis
golangci-lint run                 # Comprehensive linting (if available)
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

## Testing Standards & Patterns

**Reference Implementation**: Use `client_test.go` as the gold standard for all test files. It demonstrates exemplary Go testing practices and organization.

**Test File Organization** (mandatory structure):
```go
// 1. Test functions first (primary purpose)
func TestFeatureBasicBehavior(t *testing.T) {...}
func TestFeatureErrorHandling(t *testing.T) {...}

// 2. Mock/fake implementations (supporting types)
type mockInterface struct {...}
func (m *mockInterface) Method() {...}

// 3. Helper functions (utilities)
func setupTest(t *testing.T) {...}
func assertExpected(t *testing.T, ...) {...}
```

**Table-Driven Tests**: Use for complex scenarios with multiple test cases:
```go
tests := []struct {
    name     string
    setup    func() *mockType
    wantErr  bool
    validate func(*testing.T, result)
}{...}
```

**Helper Functions**: Always call `t.Helper()` in test utilities:
```go
func setupTestContext(t *testing.T, timeout time.Duration) (context.Context, context.CancelFunc) {
    t.Helper()  // Critical for correct line reporting
    return context.WithTimeout(context.Background(), timeout)
}
```

**Mock Design**: Use functional options for flexible mock configuration:
```go
type MockOption func(*mockType)
func WithError(err error) MockOption { return func(m *mockType) { m.err = err } }

// Usage: newMockWithOptions(WithError(expectedErr))
```

**Context Management**: Only add context when tests actually need it (blocking operations, timeouts, cancellation). Don't add unused context just for consistency:
```go
// ✅ Good - context used for blocking operations
ctx, cancel := setupTestContext(t, 10*time.Second)
defer cancel()
err := client.Connect(ctx)

// ❌ Avoid - unused context violates Go principles
ctx, cancel := setupTestContext(t, 10*time.Second)
defer cancel()
_ = ctx // Don't do this
err := NewConnectionError("test", nil)  // No blocking operation
```

**Resource Cleanup**: Test defer behavior and cleanup:
```go
func() {
    resource := setupResource(t)
    defer cleanupResource(t, resource)  // Test cleanup
    // ... test logic
}() // Scoped cleanup verification
```

**Thread Safety**: All mocks must be thread-safe with proper mutex usage for concurrent testing.

**Self-Contained Test Files**: Each test file should have its own helper functions with descriptive names to avoid dependencies between test files:
```go
// ✅ Good - each file has its own helpers
// client_test.go
func setupClientTestContext(t *testing.T, timeout time.Duration) (context.Context, context.CancelFunc) {...}

// errors_test.go  
func setupErrorTestContext(t *testing.T, timeout time.Duration) (context.Context, context.CancelFunc) {...}

// ❌ Avoid - shared helpers create hidden dependencies
// Shared setupTestContext used across multiple test files
```