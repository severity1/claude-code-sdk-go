---
description: Validate implemented TDD task
---

## Context

- Official Claude Code SDK Python
  - Reference: ../claude-code-sdk-python/
  - Analysis: analysis/
- Specifications: INTERFACE_SPEC.md
- TDD Tasks: TDD_IDIOMATIC_INTERFACES.md

## Your task

Validate the implemented code and tests against these criteria:

### Core Validation Criteria

1. **Idiomatic Go Patterns**: Does the code follow Go-native idioms and conventions?
   - Interface-driven design with proper abstractions
   - Context-first APIs for blocking operations
   - Proper error handling with structured errors and wrapping
   - Standard Go naming conventions (PascalCase exports, camelCase locals)
   - Use of `gofmt` formatting with no exceptions

2. **Python SDK Functional Parity**: Are the tests 100% aligned with Python SDK behavior?
   - All APIs have equivalent functionality
   - Same input/output behavior patterns
   - Compatible error scenarios and handling

3. **Test Quality & Coverage**: 
   - 100% test coverage of implemented code
   - Tests follow `client_test.go` patterns as gold standard
   - Table-driven tests for complex scenarios
   - Proper mock design with thread safety
   - Self-contained test files with own helper functions

4. **Production Readiness**: 
   - No dummy/placeholder implementations
   - Real, meaningful test scenarios
   - Proper resource cleanup and defer patterns
   - Buffer protection and security considerations

5. **Go-Native Architecture**:
   - Transport interface abstraction for testability
   - JSON handling with custom `UnmarshalJSON` for union types
   - Process cleanup patterns (SIGTERM → wait → SIGKILL)
   - Environment variable handling for SDK identification

6. **Standards Compliance**: Does the implementation exceed Go ecosystem expectations?

### Validation Commands

Run these before marking validation complete:
```bash
go fmt ./...                    # Format verification
go vet ./...                   # Static analysis
go test -race ./...            # Race condition check
go test -cover ./...           # Coverage analysis
golangci-lint run             # Comprehensive linting (if available)
```

Provide a detailed assessment for each criterion with specific examples.