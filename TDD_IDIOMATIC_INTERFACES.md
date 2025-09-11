# TDD Implementation Plan: Idiomatic Interfaces Transformation

## Overview

This document outlines a comprehensive Test-Driven Development (TDD) approach to implement the interface design requirements specified in `INTERFACE_SPEC.md`. Since we can break backward compatibility, we'll use TDD to ensure our new idiomatic interfaces are bulletproof.

**TDD Philosophy**: Write failing tests FIRST, then implement to make them pass. This ensures our new interfaces meet actual usage requirements and catch regressions immediately.

## Progress Tracking Table

| Phase | Day | Task | Status | Tests | Implementation | Notes |
|-------|-----|------|--------|-------|----------------|-------|
| 1 | 1 | Package Structure Foundation | ⚪ Not Started | ❌ | ❌ | Create `pkg/interfaces/` structure |
| 1 | 1 | Interface Contract Tests | ⚪ Not Started | ❌ | ❌ | Write basic interface existence tests |
| 1 | 1 | Minimal Interface Stubs | ⚪ Not Started | ❌ | ❌ | Create empty interface definitions |
| 2 | 2 | Concrete Type Tests | ⚪ Not Started | ❌ | ❌ | Test TextContent, BlockListContent, etc. |
| 2 | 2 | Concrete Type Implementation | ⚪ Not Started | ❌ | ❌ | Implement sealed interface pattern |
| 2 | 3 | Message Type Tests | ⚪ Not Started | ❌ | ❌ | Test UserMessage, AssistantMessage typing |
| 2 | 3 | Message Type Implementation | ⚪ Not Started | ❌ | ❌ | Replace interface{} with typed content |
| 2 | 4 | ContentBlock Tests | ⚪ Not Started | ❌ | ❌ | Test Type() method consistency |
| 2 | 4 | ContentBlock Implementation | ⚪ Not Started | ❌ | ❌ | Implement all ContentBlock types |
| 3 | 5 | Migration Compatibility Tests | ⚪ Not Started | ❌ | ❌ | Test old/new interface compatibility |
| 3 | 5 | Dual Import Setup | ⚪ Not Started | ❌ | ❌ | Temporary dual imports in types.go |
| 3 | 6 | Legacy Removal Tests | ⚪ Not Started | ❌ | ❌ | Test complete migration |
| 3 | 6 | Legacy Removal Implementation | ⚪ Not Started | ❌ | ❌ | Remove internal/shared imports |
| 4 | 7 | Options Naming Tests | ⚪ Not Started | ❌ | ❌ | Test Type() instead of GetType() |
| 4 | 7 | Options Interface Implementation | ⚪ Not Started | ❌ | ❌ | Standardize MCP config interfaces |
| 4 | 8 | Error Interface Tests | ⚪ Not Started | ❌ | ❌ | Test consistent error Type() methods |
| 4 | 8 | Error Interface Implementation | ⚪ Not Started | ❌ | ❌ | Implement standardized error types |
| 5 | 9 | Client Segregation Tests | ⚪ Not Started | ❌ | ❌ | Test focused interface composition |
| 5 | 9 | Client Interface Implementation | ⚪ Not Started | ❌ | ❌ | Create ConnectionManager, QueryExecutor, etc. |
| 5 | 10 | Usage Pattern Tests | ⚪ Not Started | ❌ | ❌ | Test SimpleQuerier, StreamClient patterns |
| 5 | 10 | Transport Interface Implementation | ⚪ Not Started | ❌ | ❌ | Finalize Transport and MessageIterator |
| 6 | 11-12 | Integration Tests | ⚪ Not Started | ❌ | ❌ | End-to-end type safety validation |
| 6 | 11-12 | Client Implementation Update | ⚪ Not Started | ❌ | ❌ | Update ClientImpl to use new interfaces |
| 6 | 13 | Performance Tests | ⚪ Not Started | ❌ | ❌ | Benchmark typed vs interface{} |
| 6 | 13 | Performance Validation | ⚪ Not Started | ❌ | ❌ | Ensure no regression |
| 6 | 14 | Final Validation Tests | ⚪ Not Started | ❌ | ❌ | Reflection-based interface{} detection |
| 6 | 14 | Documentation & Cleanup | ⚪ Not Started | ❌ | ❌ | Final polish and godoc |

**Status Legend**: ⚪ Not Started | 🟡 In Progress | ✅ Complete | ❌ Failed | 🔄 Needs Review

**Update Instructions**: 
```bash
# Update status as you work:
# ⚪ → 🟡 when starting
# 🟡 → ✅ when RED-GREEN cycle complete
# ✅ → 🔄 if needs review/refactoring
```

## 🚀 PHASE 1: PACKAGE STRUCTURE FOUNDATION (Day 1)
*Implements INTERFACE_SPEC.md Requirement #3: Package Organization*

### Commands to Run

```bash
# Setup - Create pkg/interfaces/ structure per specification
mkdir -p pkg/interfaces
touch pkg/interfaces/{content,message,transport,client,error,options}.go

# Create initial test files
touch pkg/interfaces/{content,message,transport,client,error,options}_test.go

# RED: Run tests (should fail)
go test ./pkg/interfaces/...

# GREEN: Implement minimal interfaces
# (See interface definitions below)

# BLUE: Refactor and verify
go test ./pkg/interfaces/...
```

### Interface Definitions Needed

**`pkg/interfaces/content.go`**:
```go
package interfaces

// Sealed interfaces
type MessageContent interface { messageContent() }
type UserMessageContent interface { MessageContent; userMessageContent() }
type AssistantMessageContent interface { MessageContent; assistantMessageContent() }
```

**`pkg/interfaces/message.go`**:
```go
package interfaces

type Message interface { Type() string }
type ContentBlock interface { Type() string } // Consistent with Message.Type()
```

### Key Tests to Write
- Interface existence (contract tests)
- Basic nil interface behavior
- Interface embedding verification

## 🚀 PHASE 2: TYPE SAFETY ELIMINATION (Days 2-4)
*Implements INTERFACE_SPEC.md Requirement #1: Zero interface{} Usage*

### Day 2: Concrete Content Types

```bash
# RED: Write tests for concrete types that don't exist yet
go test ./pkg/interfaces/... # Should fail

# GREEN: Add concrete implementations
# BLUE: Refactor tests and add edge cases
```

**Concrete Types to Implement**:
- `TextContent` (user & assistant)
- `BlockListContent` (user)  
- `ThinkingContent` (assistant)
- All sealed with unexported methods

### Day 3: Message Type Transformation

```bash
# Focus: Replace interface{} in UserMessage.Content, AssistantMessage.Content
go test ./pkg/interfaces/... # Test typed content fields
```

**Key Changes**:
- `UserMessage.Content: UserMessageContent` (not `interface{}`)
- `AssistantMessage.Content: AssistantMessageContent` (not `interface{}`)
- `StreamMessage.Message: Message` (not `interface{}`)

### Day 4: ContentBlock Type Consistency

```bash
# Focus: All ContentBlocks use Type() method consistently
go test ./pkg/interfaces/... # Test Type() consistency
```

**ContentBlock Types**:
- `TextBlock`, `ThinkingBlock`, `ToolUseBlock`, `ToolResultBlock`
- All implement `ContentBlock` interface with `Type() string`
- `ToolResultBlock.Content: MessageContent` (not `interface{}`)

## 🚀 PHASE 3: CODE MIGRATION FROM EXISTING INTERFACES (Days 5-6)
*Implements INTERFACE_SPEC.md Requirement #4: Re-export Pattern*

### Day 5: Migration Compatibility

```bash
# Focus: Bridge old internal/shared types with new pkg/interfaces types
go test ./... # Test both old and new coexist

# Update types.go with dual imports (temporary)
```

**Migration Tasks**:
- Test `shared.Message` → `interfaces.Message` compatibility
- Test `shared.ContentBlock` → `interfaces.ContentBlock` compatibility  
- Update `types.go` with temporary dual exports
- Update `Transport` interface to use new types

### Day 6: Legacy Removal

```bash
# Focus: Complete migration, remove internal/shared dependencies
go test ./... # Should pass with only pkg/interfaces

# Clean up types.go
rg "internal/shared" # Should return nothing in main package
```

**Cleanup Tasks**:
- Remove all `internal/shared` imports from main package
- Clean `types.go` to only re-export from `pkg/interfaces`
- Update examples to use new import paths
- Verify zero `internal/shared` dependencies

## 🚀 PHASE 4: INTERFACE METHOD STANDARDIZATION (Days 7-8)
*Implements INTERFACE_SPEC.md Requirement #2: Method Naming Standardization*

### Day 7: Options Interface Standardization

```bash
# Focus: Remove GetType() prefixes, use Type() consistently
go test ./pkg/interfaces/... # Test consistent naming
```

**Key Changes**:
- `McpServerConfig.GetType()` → `McpServerConfig.Type()`
- All MCP config types use consistent `Type()` method
- Remove all "Get" prefixes from interface methods

### Day 8: Error Interface Standardization  

```bash
# Focus: SDKError types use Type() method consistently
go test ./pkg/interfaces/... # Test error interface consistency
```

**Error Types**:
- `ConnectionError`, `DiscoveryError`, `ValidationError`
- All implement `SDKError` interface with `Type() string`
- Consistent with other interfaces in the system

## 🚀 PHASE 5: INTERFACE COMPOSITION AND CLIENT SEGREGATION (Days 9-10)
*Implements INTERFACE_SPEC.md Requirement #5: Interface Segregation*

### Day 9: Client Interface Segregation

```bash
# Focus: Split monolithic Client into focused interfaces
go test ./pkg/interfaces/... # Test interface composition
```

**Focused Interfaces**:
- `ConnectionManager` (Connect, Close, IsConnected)
- `QueryExecutor` (Query, QueryStream)  
- `MessageReceiver` (ReceiveMessages, ReceiveResponse)
- `ProcessController` (Interrupt, Status)
- `Client` (embeds all above)

### Day 10: Specialized Interface Patterns

```bash  
# Focus: Create specialized interface combinations
go test ./pkg/interfaces/... # Test usage patterns
```

**Specialized Combinations**:
- `SimpleQuerier` (just QueryExecutor)
- `StreamClient` (ConnectionManager + MessageReceiver)
- Transport and MessageIterator finalization

## 🚀 PHASE 6: INTEGRATION & PERFORMANCE VALIDATION (Days 11-14)
*Validates all INTERFACE_SPEC.md Deliverable Requirements*

### Day 11-12: End-to-End Integration

```bash
# Focus: Complete type safety integration testing
go test ./... # Full integration tests
go test -race ./... # Race condition testing
```

**Integration Tasks**:
- Update ClientImpl to implement all new interfaces
- End-to-end type safety validation
- Zero interface{} usage verification
- Real-world usage pattern testing

### Day 13: Performance Validation

```bash
# Focus: Ensure no performance regression
go test -bench=. -benchmem ./... # Performance benchmarks
```

**Performance Tests**:
- Typed interfaces vs interface{} boxing
- Interface composition overhead
- Memory allocation patterns
- Compilation time impact

### Day 14: Final Validation & Documentation

```bash
# Focus: Comprehensive validation and cleanup
go test ./... # All tests pass
golangci-lint run # Code quality
godoc # Documentation review
```

**Final Tasks**:
- Reflection-based interface{} detection
- Documentation updates
- Example code updates  
- Success metrics validation

## Claude Code Specific Implementation Notes

### Repository Structure
```
/home/jrpospos/workspace/severity1/claude-code-sdk-go/
├── pkg/interfaces/          # NEW: All interfaces here
├── internal/shared/         # LEGACY: Will be removed in Phase 3
├── types.go                 # UPDATE: Clean re-exports only
├── client.go               # UPDATE: Use new interfaces
└── examples/               # UPDATE: Import new interfaces
```

### Current Interface{} Locations (Found via Analysis)
```bash
# These files contain interface{} that must be eliminated:
internal/shared/message.go:36    # UserMessage.Content interface{}
internal/shared/message.go:177   # ToolResultBlock.Content interface{}  
internal/shared/stream.go:8      # StreamMessage.Message interface{}
internal/shared/stream.go:12     # StreamMessage.Request map[string]interface{}
internal/shared/stream.go:13     # StreamMessage.Response map[string]interface{}
client.go:291                   # Map creation with interface{}
```

### Current Naming Inconsistencies (Found via Analysis)
```bash  
# These method names must be standardized:
internal/shared/message.go:30    # ContentBlock.BlockType() -> Type()
internal/shared/options.go:74    # McpServerConfig.GetType() -> Type()
internal/shared/options.go:86    # McpStdioServerConfig.GetType() -> Type()
internal/shared/options.go:98    # McpSSEServerConfig.GetType() -> Type()
internal/shared/options.go:110   # McpHTTPServerConfig.GetType() -> Type()
```

### Development Commands
```bash
# Start TDD process
git checkout -b tdd-idiomatic-interfaces

# Check current interface{} usage
rg "interface\{\}" --type go internal/ client.go

# Check current naming issues  
rg "(BlockType|GetType)\(\)" --type go

# Run tests during development
go test ./pkg/interfaces/...     # New interfaces only
go test ./...                   # Full integration
go test -race ./...            # Concurrency safety
go test -bench=. -benchmem ./... # Performance

# Code quality
~/go/bin/golangci-lint run
go vet ./...
gofmt -d .
```

## INTERFACE_SPEC.md Requirement Coverage Matrix

| Spec Requirement | TDD Phase | Implementation Coverage | Validation |
|------------------|-----------|-------------------------|------------|
| **#1 Type Safety** - Zero interface{} usage | Phase 2 (Days 2-4) | ✅ Complete - Sealed interfaces, typed unions | Reflection-based testing |
| **#2 Method Naming** - Consistent Type() methods | Phase 4 (Days 7-8) | ✅ Complete - All interfaces standardized | Interface compliance tests |
| **#3 Package Organization** - Domain-based structure | Phase 1 (Day 1) | ✅ Complete - pkg/interfaces/ created | Import path validation |
| **#4 Re-export Pattern** - Clean main package exports | Phase 3 (Days 5-6) | ✅ Complete - Migration & legacy removal | Import dependency analysis |
| **#5 Interface Segregation** - SOLID client design | Phase 5 (Days 9-10) | ✅ Complete - Focused interface composition | Interface usage pattern tests |
| **#6 Performance Maintenance** - No regression | Phase 6 (Days 11-14) | ✅ Complete - Benchmarking validation | Performance comparison tests |

## Success Metrics - INTERFACE_SPEC.md Compliance Validation

Verify all deliverable requirements are met:
- [ ] **Requirement #1**: Zero interface{} usage in public API (Type Safety)
- [ ] **Requirement #2**: 100% consistent Type() naming across all interfaces (Method Naming)
- [ ] **Requirement #3**: Clean package organization with `pkg/interfaces/` (Package Organization)
- [ ] **Requirement #4**: Minimal re-exports in main package (Re-export Pattern)
- [ ] **Requirement #5**: Interface segregation applied to Client (Interface Composition)
- [ ] **Requirement #6**: Performance equal or better than current implementation (Performance Maintenance)

## Quick Start Commands

```bash
# Begin implementation
git checkout -b tdd-idiomatic-interfaces

# Phase 1: Foundation  
mkdir -p pkg/interfaces && cd pkg/interfaces
touch {content,message,transport,client,error,options}.go
touch {content,message,transport,client,error,options}_test.go

# Run TDD cycle for each phase
go test ./pkg/interfaces/... # Should fail (RED)
# Implement interfaces (GREEN)  
go test ./pkg/interfaces/... # Should pass
# Refactor (BLUE)

# Final validation
go test ./...
go test -race ./...
~/go/bin/golangci-lint run
rg "interface\{\}" --type go --exclude-dir=internal
```