# Interface Design Specification: Claude Code Go SDK

## Specification Overview

This document defines the target interface design for the Claude Code Go SDK, establishing requirements for idiomatic Go interface patterns that will be implemented through Test-Driven Development (TDD). The specification enables breaking changes to achieve interface excellence in the `idiomatic-interfaces` branch.

**Design Goal**: Transform the current interface architecture into exemplary idiomatic Go design through complete type safety, consistent naming, and optimal package organization.

**Breaking Changes Policy**: This specification authorizes breaking changes to achieve interface design excellence.

## Current Interface Architecture Assessment

### Requirements Already Met
- **Interface Segregation**: `Message`, `ContentBlock`, `Transport`, and `MessageIterator` interfaces demonstrate proper separation of concerns
- **Dependency Injection**: Transport interface design supports clean testing and mocking patterns
- **Polymorphic Design**: Union types properly expressed through interface hierarchies
- **Go-Native Concurrency**: Context-first design with proper `context.Context` usage
- **Resource Management**: Interface contracts support defer-based cleanup patterns  
- **Testability**: Interface boundaries enable compile-time verification and easy mocking

## Interface Design Requirements

### 1. Type Safety Requirement - Zero `interface{}` Usage

**Specification**: All interface{} usage must be eliminated in favor of strongly-typed interface hierarchies.

**Current Non-Compliant Code**:
```go
// ❌ COMPLETELY UNACCEPTABLE
type StreamMessage struct {
    Message interface{} `json:"message,omitempty"`  
}

type UserMessage struct {
    Content interface{} `json:"content"` 
}

type ToolResultBlock struct {
    Content interface{} `json:"content"`
}

// ❌ Even in internal code
Request  map[string]interface{} `json:"request,omitempty"`
Response map[string]interface{} `json:"response,omitempty"`
```

**Required Implementation**: Sealed interface pattern with typed unions:

```go
// ✅ IDIOMATIC REPLACEMENT
type StreamMessage struct {
    Message MessageContent `json:"message,omitempty"`
}

type MessageContent interface {
    isMessageContent()
}

type UserMessage struct {
    Content UserContent `json:"content"`
}

type UserContent interface {
    isUserContent()
}

// Concrete implementations
type TextContent struct {
    Text string `json:"text"`
}
func (TextContent) isUserContent() {}

type BlockContent struct {
    Blocks []ContentBlock `json:"blocks"`
}  
func (BlockContent) isUserContent() {}
```

**Compliance Metric**: Zero interface{} usage in public API (verified by reflection-based testing).

### 2. Method Naming Standardization Requirement

**Specification**: All interfaces must use consistent method naming following Go conventions.

**Current Non-Compliant Code**:
```go
// ❌ NON-STANDARD NAMING
type Message interface {
    Type() string           // This one is correct
}

type ContentBlock interface {
    BlockType() string      // ❌ Different name for same concept
}

type McpServerConfig interface {
    GetType() McpServerType // ❌ Java-style getter nonsense
}
```

**Required Implementation**: Standardized method naming:
```go
// ✅ SPECIFICATION-COMPLIANT
type Message interface {
    Type() string
}

type ContentBlock interface {
    Type() string           // ✅ SAME NAME, SAME CONCEPT
}

type McpServerConfig interface {
    Type() McpServerType    // ✅ NO JAVA GETTERS IN GO
}
```

**Compliance Metric**: 100% consistent Type() method naming across all interfaces.

### 3. Package Organization Requirement

**Specification**: All interfaces must be organized in dedicated `pkg/interfaces/` package with domain-based file structure.

**Current Non-Compliant Structure**:
```
types.go                  → Transport (why just this one?)
internal/shared/message.go → Message, ContentBlock  
internal/shared/stream.go  → MessageIterator
internal/shared/errors.go  → SDKError
```

**Required Implementation**: Domain-based interface organization:

```
// ✅ SPECIFICATION-COMPLIANT STRUCTURE
pkg/
  interfaces/           ← NEW: All interfaces here
    message.go         → Message, ContentBlock, UserContent, etc.
    transport.go       → Transport, MessageIterator
    client.go          → Client (split into focused interfaces)
    error.go           → All error interfaces
    
// Main package only re-exports what users need
types.go → MINIMAL re-exports only
```

**Compliance Metric**: All interfaces located in `pkg/interfaces/` with zero scattered definitions.

## Additional Design Requirements

### 4. Re-export Pattern Requirement

**Specification**: Main package must provide clean, minimal re-exports without type alias proliferation.

**Current Non-Compliant Code** (`types.go`):
```go
// ❌ 42 LINES OF TYPE ALIAS HELL
type Message = shared.Message
type ContentBlock = shared.ContentBlock
type UserMessage = shared.UserMessage
type AssistantMessage = shared.AssistantMessage
type SystemMessage = shared.SystemMessage
type ResultMessage = shared.ResultMessage
type TextBlock = shared.TextBlock
type ThinkingBlock = shared.ThinkingBlock
type ToolUseBlock = shared.ToolUseBlock
type ToolResultBlock = shared.ToolResultBlock
type StreamMessage = shared.StreamMessage
type MessageIterator = shared.MessageIterator
// ... 30 MORE LINES OF THIS NONSENSE
```

**Required Implementation**: Clean, minimal re-exports:
```go
// ✅ CLEAN - Only import what users actually need
package claudecode

import "github.com/severity1/claude-code-sdk-go/pkg/interfaces"

// Interfaces users need
type Message = interfaces.Message
type ContentBlock = interfaces.ContentBlock
type Client = interfaces.Client

// Constructors and concrete types only
func NewClient(opts ...Option) *ClientImpl { ... }
func Query(ctx context.Context, prompt string) MessageIterator { ... }
```

**Compliance Metric**: Main package exports less than 15 essential types only.

### 5. Interface Segregation Requirement

**Specification**: Client interface must be decomposed into focused, single-responsibility interfaces following SOLID principles.

**Current Non-Compliant Code**:
```go
// ❌ VIOLATES INTERFACE SEGREGATION PRINCIPLE
type Client interface {
    Connect(ctx context.Context, prompt ...StreamMessage) error
    Disconnect() error
    Query(ctx context.Context, prompt string, sessionID ...string) error
    QueryStream(ctx context.Context, messages <-chan StreamMessage) error
    ReceiveMessages(ctx context.Context) <-chan Message
    ReceiveResponse(ctx context.Context) MessageIterator
    Interrupt(ctx context.Context) error
}
```

**Required Implementation**: Focused interface composition:
```go
// ✅ SPECIFICATION-COMPLIANT INTERFACES
type ConnectionManager interface {
    Connect(ctx context.Context, prompt ...StreamMessage) error
    Disconnect() error
}

type QueryExecutor interface {
    Query(ctx context.Context, prompt string, sessionID ...string) error
    QueryStream(ctx context.Context, messages <-chan StreamMessage) error
}

type MessageReceiver interface {
    ReceiveMessages(ctx context.Context) <-chan Message
    ReceiveResponse(ctx context.Context) MessageIterator
}

type ProcessController interface {
    Interrupt(ctx context.Context) error
}

// Composed interface using Go's interface embedding
type Client interface {
    ConnectionManager
    QueryExecutor  
    MessageReceiver
    ProcessController
}
```

**Benefits**:
- **Easier mocking**: Mock only the interfaces you need
- **Better testing**: Test each concern independently  
- **Cleaner dependencies**: Components depend on minimal interfaces
- **SOLID compliance**: Each interface has single responsibility

### 6. Error Interface Design

**Problem**: Error interfaces could be more specific and follow Go error patterns better.

**Current**:
```go
type SDKError interface {
    error
    Type() string
}
```

**Improvements**:
- Could implement `errors.Is()` and `errors.As()` patterns more explicitly
- Could have more specific error interfaces for different error categories
- Error type strings could be typed constants instead of strings

## Minor Issues 🟢

### 7. Missing Interface Documentation

**Problem**: Some interfaces lack comprehensive documentation.

**Examples**:
- `MessageIterator` interface could document EOF handling better
- `Transport` interface could document connection lifecycle better
- Error interfaces could document when each type is used

### 8. Interface Testing Gaps

**Problem**: While individual interfaces are well-tested, some interaction patterns need better coverage.

**Missing Tests**:
- Interface composition behavior
- Interface nil handling patterns
- Concurrent interface usage patterns

## AGGRESSIVE BREAKING CHANGES ROADMAP 🚀

Since we can break compatibility, let's build the most idiomatic Go interfaces possible.

### PHASE 1: ELIMINATE ALL `interface{}` (Week 1)

**Complete type safety transformation**:

```go
// ✅ NEW MESSAGE CONTENT SYSTEM
type MessageContent interface {
    messageContent() // Sealed interface
}

type UserMessageContent interface {
    MessageContent
    userMessageContent()
}

type AssistantMessageContent interface {
    MessageContent  
    assistantMessageContent()
}

// Concrete implementations
type TextContent struct {
    Text string `json:"text"`
}
func (TextContent) messageContent() {}
func (TextContent) userMessageContent() {}

type BlockListContent struct {
    Blocks []ContentBlock `json:"blocks"`
}
func (BlockListContent) messageContent() {}
func (BlockListContent) userMessageContent() {}

type ThinkingContent struct {
    Thinking string `json:"thinking"`
    Signature string `json:"signature"`
}
func (ThinkingContent) messageContent() {}
func (ThinkingContent) assistantMessageContent() {}

// ✅ NEW TYPED MESSAGES  
type UserMessage struct {
    Type    string             `json:"type"`
    Content UserMessageContent `json:"content"`
}

type AssistantMessage struct {
    Type    string                  `json:"type"`  
    Content AssistantMessageContent `json:"content"`
    Model   string                  `json:"model"`
}
```

**BREAKING**: This changes the entire message API but makes it completely type-safe.

### PHASE 2: INTERFACE METHOD STANDARDIZATION (Week 1)

**Ruthless naming consistency**:

```go
// ✅ BEFORE: Naming chaos  
type Message interface { Type() string }              // Good
type ContentBlock interface { BlockType() string }   // ❌ Inconsistent  
type McpServerConfig interface { GetType() McpServerType } // ❌ Java-style

// ✅ AFTER: Perfect consistency
type Message interface { Type() string }
type ContentBlock interface { Type() string }        // ✅ Consistent
type McpServerConfig interface { Type() McpServerType } // ✅ Go-style

// ✅ STANDARDIZE ALL INTERFACE METHODS
type Transport interface {
    Connect(ctx context.Context) error                // Not "ConnectTo"
    Close() error                                     // Not "Disconnect" 
    SendMessage(ctx context.Context, msg Message) error // Consistent param names
    ReceiveMessages(ctx context.Context) (<-chan Message, <-chan error)
    Interrupt(ctx context.Context) error
}

type MessageIterator interface {  
    Next(ctx context.Context) (Message, error)       // Standard iterator pattern
    Close() error                                     // Standard cleanup
    Err() error                                       // Standard error checking
}
```

### PHASE 3: RADICAL INTERFACE REORGANIZATION (Week 2)

**Complete package restructure**:

```go
// ✅ NEW CLEAN STRUCTURE
pkg/
  interfaces/              ← ALL interfaces live here
    message.go             // Message, ContentBlock, MessageContent interfaces
    client.go              // Client, ConnectionManager, QueryExecutor, etc.  
    transport.go           // Transport, MessageIterator
    content.go             // UserMessageContent, AssistantMessageContent, etc.
    error.go               // All error interfaces

// ✅ MAIN PACKAGE: Clean public API
claudecode/
  client.go                // ClientImpl + constructors
  query.go                 // Query function + QueryIterator
  options.go               // Option functions
  types.go                 // ONLY essential re-exports (5-10 lines MAX)

// ✅ IMPLEMENTATION PACKAGES  
internal/
  subprocess/              // Transport implementation
  parser/                  // Message parsing
  cli/                     // CLI discovery
```

**BREAKING**: Completely new import structure, but infinitely cleaner.

### PHASE 4: INTERFACE COMPOSITION PERFECTION (Week 2)

**Apply Interface Segregation Principle everywhere**:

```go
// ✅ pkg/interfaces/client.go
package interfaces

// Focused, single-responsibility interfaces
type ConnectionManager interface {
    Connect(ctx context.Context) error
    Close() error                          // Renamed from Disconnect
    IsConnected() bool                    // Add connection state query
}

type QueryExecutor interface {
    Query(ctx context.Context, prompt string, opts ...QueryOption) error
    QueryStream(ctx context.Context, messages <-chan Message) error
}

type MessageReceiver interface {
    ReceiveMessages(ctx context.Context) <-chan Message
    ReceiveResponse(ctx context.Context) MessageIterator
}

type ProcessController interface {
    Interrupt(ctx context.Context) error
    Status(ctx context.Context) (ProcessStatus, error)  // Add status query
}

// Composed client interface - users can depend on subsets
type Client interface {
    ConnectionManager
    QueryExecutor
    MessageReceiver
    ProcessController
}

// ✅ Allow users to depend on minimal interfaces
type SimpleQuerier interface {
    QueryExecutor                         // For users who only need querying
}

type StreamClient interface {
    ConnectionManager                     // For users building custom streaming
    MessageReceiver
}
```

**BREAKING**: Users can now depend on exactly the interfaces they need.

## AGGRESSIVE IMPLEMENTATION STRATEGY 💥

**Timeline**: 2 weeks to complete interface transformation

### Week 1: CORE BREAKING CHANGES

**Days 1-2: Type Safety Elimination**
```bash
# Day 1: Create new typed interfaces  
git checkout -b eliminate-interface-empty
mkdir -p pkg/interfaces
# Create all typed content interfaces
# Define sealed interface patterns

# Day 2: Replace all interface{} usage
# Update StreamMessage, UserMessage, ToolResultBlock  
# Replace map[string]interface{} with typed alternatives
# Update all tests to use typed assertions
```

**Days 3-4: Interface Method Standardization**
```bash
# Day 3: Rename all Type() methods consistently
# Remove all Get* prefixes from interface methods
# Standardize parameter names across interfaces

# Day 4: Update all implementations
# Fix all compile errors
# Verify interface compliance everywhere
```

**Day 5: Testing & Integration**
```bash
# Run full test suite
# Fix any remaining type assertion issues  
# Verify no interface{} remains in codebase
```

### Week 2: STRUCTURAL REORGANIZATION

**Days 1-2: Package Restructure**
```bash
# Day 1: Create pkg/interfaces/ structure
# Move all interfaces to correct packages
# Update import paths throughout codebase

# Day 2: Minimize main package re-exports
# Clean up types.go to <10 lines
# Update all internal imports
```

**Days 3-4: Interface Composition**
```bash
# Day 3: Split Client interface using composition
# Create ConnectionManager, QueryExecutor, etc.
# Update ClientImpl to implement all interfaces

# Day 4: Create specialized interface combinations
# SimpleQuerier, StreamClient, etc.
# Update examples to use focused interfaces
```

**Day 5: Final Polish**
```bash
# Update all documentation
# Verify godoc output is clean  
# Final test suite run
# Performance verification
```

## Testing Strategy

### Interface Contract Tests
```go
func TestInterfaceContracts(t *testing.T) {
    // Verify all Message implementations
    messages := []Message{
        &UserMessage{},
        &AssistantMessage{},
        &SystemMessage{},
        &ResultMessage{},
    }
    
    for _, msg := range messages {
        assertMessageContract(t, msg)
    }
}

func assertMessageContract(t *testing.T, msg Message) {
    // Test Type() method
    msgType := msg.Type()
    if msgType == "" {
        t.Error("Message.Type() must return non-empty string")
    }
    
    // Test JSON marshaling
    data, err := json.Marshal(msg)
    if err != nil {
        t.Errorf("Message must be JSON marshalable: %v", err)
    }
    
    // Test round-trip
    var parsed map[string]interface{}
    if err := json.Unmarshal(data, &parsed); err != nil {
        t.Errorf("Message JSON must be parseable: %v", err)
    }
    
    if parsed["type"] != msgType {
        t.Error("JSON type field must match Type() method")
    }
}
```

### Interface Composition Tests
```go
func TestClientInterfaceComposition(t *testing.T) {
    client := &ClientImpl{}
    
    // Test interface compliance
    var _ Connector = client
    var _ Querier = client  
    var _ MessageReceiver = client
    var _ Interrupter = client
    var _ Client = client
}
```

## BREAKING CHANGES: NO BACKWARD COMPATIBILITY 💥

**Since we're a young project, we can make radical improvements without compatibility constraints.**

### Version Strategy
- **v2.0.0**: Complete interface rewrite
- **No migration path**: Users upgrade and fix their code  
- **Clear documentation**: Show exactly what changed
- **Better long-term**: Perfect interfaces from the start

### Communication Strategy
```markdown
# BREAKING CHANGES in v2.0.0

## What Changed
- ❌ Removed ALL `interface{}` usage
- ❌ Renamed interface methods for consistency  
- ❌ Reorganized package structure
- ❌ Split monolithic interfaces

## Migration Guide
1. Replace `interface{}` type assertions with typed interfaces
2. Update method calls: `BlockType()` → `Type()`
3. Update imports: `internal/shared` → `pkg/interfaces`  
4. Use focused interfaces: `ConnectionManager` instead of full `Client`

## Benefits
✅ 100% compile-time type safety
✅ Better testing through interface composition
✅ Cleaner godoc and API surface
✅ SOLID principle compliance
```

## Success Metrics

### AGGRESSIVE SUCCESS METRICS 🎯

#### Interface Purity Goals
- [ ] **ZERO `interface{}` usage** anywhere (except JSON internal unmarshaling)
- [ ] **100% consistent naming** across all interfaces
- [ ] **Perfect package organization** with clear placement rules
- [ ] **Interface segregation** applied everywhere

#### Type Safety Goals  
- [ ] **Zero runtime type assertions** in user-facing code
- [ ] **All unions expressed as interfaces** with sealed implementations
- [ ] **Compile-time verification** of all interface compliance
- [ ] **No `map[string]interface{}` anywhere** in public APIs

#### Developer Experience Goals
- [ ] **Focused interfaces** - users depend on minimal interfaces
- [ ] **Clear godoc** for every interface and method
- [ ] **Easy mocking** through interface composition
- [ ] **Intuitive API** that feels natural to Go developers

#### Performance Goals
- [ ] **Zero performance regression** from interface changes
- [ ] **Reduced allocations** from eliminating interface{} boxing
- [ ] **Better inlining** through specific interface types
- [ ] **Faster compilation** through cleaner import graph

## Interface Design Specification Summary

**This specification defines the requirements to transform the Claude Code Go SDK into an exemplar of idiomatic Go interface design.**

### Current Compliance Status
- ✅ **Architectural Foundation**: Strong interface segregation and dependency injection patterns
- ✅ **Go-Native Patterns**: Context-first design and proper resource management
- ❌ **Type Safety**: Non-compliant interface{} usage throughout
- ❌ **Naming Consistency**: Non-standard method naming across interfaces
- ❌ **Package Organization**: Scattered interface definitions violate organization principles

### Target Compliance State
- ✅ **Zero interface{} Usage**: 100% type safety through sealed interface patterns
- ✅ **Consistent Naming**: All interfaces use standardized Type() method naming
- ✅ **Organized Structure**: All interfaces located in domain-based `pkg/interfaces/` package
- ✅ **Interface Segregation**: Client decomposed into focused, single-responsibility interfaces
- ✅ **Compile-Time Safety**: Elimination of runtime type assertions in favor of static guarantees
- ✅ **Testing Excellence**: All interfaces optimized for easy mocking and verification

### Implementation Authority
This specification authorizes breaking changes to achieve interface design excellence. The current project maturity provides optimal timing for fundamental interface architecture improvements.

### Deliverable Requirements
1. **Complete Type Safety**: Zero interface{} usage in public API
2. **Naming Standardization**: 100% consistent interface method naming
3. **Structural Organization**: Domain-based interface package organization
4. **Interface Composition**: SOLID-compliant client interface decomposition
5. **Performance Maintenance**: Equal or better performance than current implementation
6. **Documentation Excellence**: Comprehensive godoc coverage for all interface contracts

**This specification establishes the foundation for implementing exemplary Go interface design through Test-Driven Development.**