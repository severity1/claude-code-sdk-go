# Testdata Directory Implementation Tasks

## Overview

**Goal**: Extract static test data from `client_test.go` into Go-standard `testdata/` directory.

**Current State**: Static test data embedded in test code  
**Target State**: Static data files in `testdata/` directory for test input/output validation

## Testdata Extraction by Component (24 tasks)

### Main Package Tests (6 tasks)

#### TD001: Query Test Fixtures ✅ PENDING
**Source**: `query_test.go` (1,372 lines)
**Target**: `testdata/query/fixtures/`
**Description**: Mock CLI responses, message sequences, timeout scenarios
**Acceptance**: JSON files for query execution patterns and expected outputs

#### TD002: Client Test Fixtures ✅ PENDING
**Source**: `client_test.go` (3,583 lines)  
**Target**: `testdata/client/fixtures/`
**Description**: Session data, streaming responses, transport configurations
**Acceptance**: Comprehensive client testing scenarios as static data

#### TD003: Options Test Configuration ✅ PENDING
**Source**: `options_test.go` (634 lines)
**Target**: `testdata/options/configs/`
**Description**: Valid/invalid option combinations, default values, edge cases
**Acceptance**: Configuration validation test data

#### TD004: Types Test Fixtures ✅ PENDING
**Source**: `types_test.go` (129 lines)
**Target**: `testdata/types/samples/`
**Description**: Message type examples, content block variations
**Acceptance**: Type validation and serialization test data

#### TD005: Error Response Examples ✅ PENDING
**Source**: `errors_test.go` (83 lines)
**Target**: `testdata/errors/responses/`
**Description**: Error scenarios, exit codes, error messages
**Acceptance**: Error handling validation data

#### TD006: Documentation Examples ✅ PENDING
**Source**: `doc_test.go` (14 lines)
**Target**: `testdata/docs/examples/`
**Description**: Documentation code examples and expected outputs
**Acceptance**: Example code execution results

### Parser Component Tests (5 tasks)

#### TD007: JSON Parser Test Data ✅ PENDING
**Source**: `internal/parser/json_test.go` (2,191 lines)
**Target**: `testdata/parser/json/`
**Description**: Valid/invalid JSON, partial messages, buffer overflow scenarios
**Acceptance**: Comprehensive JSON parsing test cases

#### TD008: Message Parse Examples ✅ PENDING
**Target**: `testdata/parser/messages/`
**Description**: All message types, content blocks, union type discrimination
**Acceptance**: Message parsing validation data

#### TD009: Buffer Management Data ✅ PENDING
**Target**: `testdata/parser/buffers/`
**Description**: Large messages, 1MB+ content, buffer overflow protection
**Acceptance**: Buffer management test scenarios

#### TD010: Unicode and Escape Test Data ✅ PENDING
**Target**: `testdata/parser/unicode/`
**Description**: Unicode characters, escape sequences, encoding edge cases
**Acceptance**: Character encoding validation data

#### TD011: Multiple JSON Object Data ✅ PENDING
**Target**: `testdata/parser/multiple/`
**Description**: Multiple JSON objects on single line, embedded newlines
**Acceptance**: Complex JSON parsing scenarios

### Transport Component Tests (4 tasks)

#### TD012: Transport Test Scenarios ✅ PENDING
**Source**: `internal/subprocess/transport_test.go` (821 lines)
**Target**: `testdata/transport/scenarios/`
**Description**: Process lifecycle, I/O handling, termination sequences
**Acceptance**: Transport behavior validation data

#### TD013: CLI Command Examples ✅ PENDING
**Target**: `testdata/transport/commands/`
**Description**: CLI argument combinations, flag variations, environment setup
**Acceptance**: Command building test data

#### TD014: Process Management Data ✅ PENDING
**Target**: `testdata/transport/process/`
**Description**: Exit codes, signal handling, 5-second termination data
**Acceptance**: Process lifecycle validation scenarios

#### TD015: I/O Stream Examples ✅ PENDING
**Target**: `testdata/transport/streams/`
**Description**: Stdin/stdout/stderr examples, streaming patterns
**Acceptance**: I/O handling test data

### CLI Discovery Tests (3 tasks)

#### TD016: CLI Discovery Test Data ✅ PENDING
**Source**: `internal/cli/discovery_test.go` (445 lines)
**Target**: `testdata/cli/discovery/`
**Description**: CLI path locations, Node.js validation, version detection
**Acceptance**: CLI discovery validation scenarios

#### TD017: Command Building Examples ✅ PENDING
**Target**: `testdata/cli/commands/`
**Description**: Argument construction, option mapping, flag combinations
**Acceptance**: Command building test data

#### TD018: Environment Test Scenarios ✅ PENDING
**Target**: `testdata/cli/environment/`
**Description**: Working directory validation, environment variable setup
**Acceptance**: Environment configuration test data

### Shared Component Tests (6 tasks)

#### TD019: Shared Message Test Data ✅ PENDING
**Source**: `internal/shared/message_test.go` (320 lines)
**Target**: `testdata/shared/messages/`
**Description**: Core message types, interface compliance
**Acceptance**: Message interface validation data

#### TD020: Shared Options Test Data ✅ PENDING
**Source**: `internal/shared/options_test.go` (231 lines)
**Target**: `testdata/shared/options/`
**Description**: Option validation, type conversion, defaults
**Acceptance**: Shared options test scenarios

#### TD021: Shared Error Test Data ✅ PENDING
**Source**: `internal/shared/errors_test.go` (279 lines)
**Target**: `testdata/shared/errors/`
**Description**: Error types, error wrapping, context preservation
**Acceptance**: Error handling validation data

#### TD022: Cross-Component Integration Data ✅ PENDING
**Target**: `testdata/integration/cross_component/`
**Description**: Multi-component interaction scenarios
**Acceptance**: Integration testing validation data

#### TD023: Performance Baseline Data ✅ PENDING
**Target**: `testdata/performance/benchmarks/`
**Description**: Performance test baselines, memory usage patterns
**Acceptance**: Performance regression detection data

#### TD024: Edge Case Collection ✅ PENDING
**Target**: `testdata/edge_cases/collection/`
**Description**: Collected edge cases from all components
**Acceptance**: Comprehensive edge case validation data

### Documentation Tasks (1 task)

#### TD025: Testdata Memory Documentation ✅ PENDING
**Target**: `testdata/CLAUDE.md`
**Description**: Create component memory file following project patterns and official Claude Code memory guidelines
**Acceptance**: CLAUDE.md file with testdata-specific context, usage patterns, and integration instructions

### TDD Integration Tasks (1 task)

#### TD026: TDD Tasks Alignment ✅ PENDING
**Target**: `TDD_IMPLEMENTATION_TASKS.md`
**Description**: Update remaining TDD tasks to reference testdata files and align with testdata structure
**Acceptance**: TDD tasks seamlessly integrate with testdata directory, tests can load static data from testdata paths

## Implementation Progress

**Total Tasks**: 26 (covers all 12 test files + documentation + TDD alignment)  
**Current Progress**: 0/26 (0%)  
**Total Test Lines**: 10,102 lines across all test files
**Focus**: Extract static data from all test components to `testdata/` directory with proper documentation and TDD integration

## Test File Coverage Analysis

| Component | Test File | Lines | Testdata Tasks |
|-----------|-----------|-------|----------------|
| Main Package | `client_test.go` | 3,583 | TD002 |
| Main Package | `query_test.go` | 1,372 | TD001 |
| Main Package | `options_test.go` | 634 | TD003 |
| Main Package | `types_test.go` | 129 | TD004 |
| Main Package | `errors_test.go` | 83 | TD005 |
| Main Package | `doc_test.go` | 14 | TD006 |
| Parser | `internal/parser/json_test.go` | 2,191 | TD007-TD011 |
| Transport | `internal/subprocess/transport_test.go` | 821 | TD012-TD015 |
| CLI Discovery | `internal/cli/discovery_test.go` | 445 | TD016-TD018 |
| Shared Messages | `internal/shared/message_test.go` | 320 | TD019 |
| Shared Options | `internal/shared/options_test.go` | 231 | TD020 |
| Shared Errors | `internal/shared/errors_test.go` | 279 | TD021 |
| Integration | N/A | N/A | TD022-TD024 |
| Documentation | N/A | N/A | TD025 |
| TDD Integration | `TDD_IMPLEMENTATION_TASKS.md` | N/A | TD026 |

## Success Criteria

1. **Complete Coverage**: All 12 test files have corresponding testdata extraction
2. **Go Convention Compliance**: Uses `testdata/` only for static data files
3. **Component Organization**: Organized by component matching internal structure
4. **Data Accessibility**: Test data easily accessible via structured paths
5. **Maintainability**: Static test data can be modified independently
6. **Proper Documentation**: Component memory file following project CLAUDE.md patterns
7. **TDD Integration**: Remaining TDD tasks aligned to use testdata files seamlessly

## Comprehensive Testdata Directory Structure

```
testdata/
├── CLAUDE.md                     # Component memory documentation
├── query/                        # Query function tests
│   └── fixtures/
├── client/                       # Client streaming tests  
│   └── fixtures/
├── options/                      # Options validation
│   └── configs/
├── types/                        # Type validation
│   └── samples/
├── errors/                       # Error scenarios
│   └── responses/
├── docs/                         # Documentation examples
│   └── examples/
├── parser/                       # JSON parsing tests
│   ├── json/
│   ├── messages/
│   ├── buffers/
│   ├── unicode/
│   └── multiple/
├── transport/                    # Transport layer tests
│   ├── scenarios/
│   ├── commands/
│   ├── process/
│   └── streams/
├── cli/                          # CLI discovery tests
│   ├── discovery/
│   ├── commands/
│   └── environment/
├── shared/                       # Shared component tests
│   ├── messages/
│   ├── options/
│   └── errors/
├── integration/                  # Cross-component tests
│   └── cross_component/
├── performance/                  # Performance baselines
│   └── benchmarks/
└── edge_cases/                   # Edge case collection
    └── collection/
```

## Dependencies

- Go 1.19+ (for `testdata/` directory support)
- JSON format for structured test data
- No external dependencies for test data
- Claude Code memory file conventions for documentation

## Implementation Notes

- Only static data files belong in `testdata/`
- All files should be readable by tests using standard Go file operations
- JSON format preferred for structured data
- Directory-based organization for related test scenarios
- CLAUDE.md follows project memory patterns and official Claude Code guidelines
- Component memory file provides testdata-specific context and usage instructions