---
description: Implement TDD task
argument-hint: [TDD task]
---

## Context

- Official Claude Code SDK Python
  - Reference: ../claude-code-sdk-python/
  - Analysis: analysis/
- Specifications: INTERFACE_SPEC.md
- TDD Tasks: TDD_IDIOMATIC_INTERFACES.md

## Your task

Let's work on the next TDD implementation task: $ARGUMENTS. 

### Implementation Requirements

**Official Go Best Practices & Idioms:**
- Format code with `gofmt` (official Go formatting tool)
- Follow Go naming: `PascalCase` for exports, `camelCase` for unexported
- Interface-driven design with small, focused interfaces (Effective Go)
- Context-first parameters for functions that can block (official pattern)
- Structured error handling with `fmt.Errorf` and `%w` for wrapping (Go 1.13+)
- Use `t.Helper()` in test utility functions (official testing package)

**TDD RED-GREEN-BLUE Cycle:**
1. **RED**: Write failing test first (no dummy/placeholder code)
2. **GREEN**: Implement minimal code to make test pass
3. **BLUE**: Refactor while keeping tests green

**Test Standards (follow `client_test.go` as gold standard):**
- Test functions first, mocks second, helpers last in file organization  
- Table-driven tests for complex scenarios with multiple cases
- `t.Helper()` in all test utility functions for correct line reporting
- Functional options for flexible mock configuration
- Context only when actually needed (blocking ops, timeouts, cancellation)
- Thread-safe mocks with proper synchronization
- Self-contained test files with descriptive helper function names

**Code Quality Gates:**
- Run `go fmt ./...` for formatting
- Run `go vet ./...` for static analysis  
- Run `go test -race ./...` for race condition detection
- Run `go test -cover ./...` for coverage analysis
- Run integration tests with real subprocess transport
- Implement 1MB buffer limits with overflow protection
- Follow exact cleanup pattern: SIGTERM → 5s wait → SIGKILL

**Behavioral Parity:**
- Ensure 100% behavioral parity or exceed Python SDK capabilities
- Set `CLAUDE_CODE_ENTRYPOINT` environment variables appropriately
- Use `MockTransport` for unit tests, real subprocess for integration
- Reference analysis/ files and Python SDK for complete understanding

Ensure adherence to all code style guidance in CLAUDE.md while maintaining Go idioms throughout.