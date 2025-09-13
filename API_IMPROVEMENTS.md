# Claude Code SDK Go - API Improvements Analysis

**Status**: Draft Proposal
**Date**: 2025-09-14
**Author**: Senior Go Engineering Review
**Breaking Changes**: ✅ Acceptable (young project)

## Executive Summary

This document outlines critical improvements needed for the Claude Code SDK Go public APIs. As a young project that can accept breaking changes, we have the opportunity to redesign APIs using Go best practices without backwards compatibility constraints.

## Critical Issues Identified

### 1. Interface Design - Violation of SOLID Principles

#### Current Problem
```go
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

**Issues:**
- 7 methods violate Interface Segregation Principle
- Mixes concerns: connection management, message sending, message receiving
- Forces implementations to handle all methods even when not needed
- Makes testing and mocking complex

#### Proposed Solution (BREAKING CHANGE)
```go
// Segregated interfaces following Single Responsibility Principle
type Connector interface {
    Connect(ctx context.Context) error
    Close() error
    State() ConnectionState
}

type MessageSender interface {
    Send(ctx context.Context, msg Message) error
    SendStream(ctx context.Context, msgs <-chan Message) error
    Interrupt(ctx context.Context) error
}

type MessageReceiver interface {
    Receive(ctx context.Context) <-chan Result[Message]
    ReceiveIterator(ctx context.Context) MessageIterator
}

// Composed interface for full functionality
type Client interface {
    Connector
    MessageSender
    MessageReceiver
}

// Connection state management
type ConnectionState int
const (
    StateDisconnected ConnectionState = iota
    StateConnecting
    StateConnected
    StateDisconnecting
    StateError
)
```

**Benefits:**
- Each interface has single responsibility
- Easier testing with focused mocks
- Clients can depend on minimal interface they need
- Clear separation of concerns

### 2. Error Handling Inconsistencies

#### Current Problem
```go
// Inconsistent error patterns
func Connect(ctx context.Context) error                    // Direct error return
func ReceiveMessages(ctx context.Context) <-chan Message   // Error handling unclear
func QueryStream(ctx context.Context, messages <-chan StreamMessage) error // Fire-and-forget
```

#### Proposed Solution (BREAKING CHANGE)
```go
// Result type for channel-based operations
type Result[T any] struct {
    Value T
    Err   error
}

func (r Result[T]) Unwrap() (T, error) {
    return r.Value, r.Err
}

// Consistent error handling across all operations
type MessageReceiver interface {
    Receive(ctx context.Context) <-chan Result[Message]
    ReceiveIterator(ctx context.Context) (MessageIterator, error)
}

type MessageSender interface {
    Send(ctx context.Context, msg Message) error
    SendStream(ctx context.Context, msgs <-chan Message) <-chan error
}
```

### 3. Resource Management & Lifecycle

#### Current Problem
```go
type ClientImpl struct {
    connected bool // Binary state, no lifecycle management
    // Missing: connecting, disconnecting, error states
    // Missing: reconnection logic
    // Missing: graceful shutdown
}
```

#### Proposed Solution (BREAKING CHANGE)
```go
type ConnectionState int

const (
    StateDisconnected ConnectionState = iota
    StateConnecting
    StateConnected
    StateReconnecting
    StateDisconnecting
    StateError
)

type ConnectionInfo struct {
    State        ConnectionState
    ConnectedAt  time.Time
    LastActivity time.Time
    ErrorCount   int
    LastError    error
}

type Connector interface {
    Connect(ctx context.Context) error
    Close() error
    State() ConnectionState
    Info() ConnectionInfo

    // State change notifications
    Watch(ctx context.Context) <-chan ConnectionState

    // Wait for specific state with timeout
    WaitFor(ctx context.Context, state ConnectionState) error
}

// Auto-reconnection configuration
type ReconnectConfig struct {
    Enabled     bool
    MaxRetries  int
    BackoffFunc func(attempt int) time.Duration
}
```

### 4. Concurrency Safety Issues

#### Current Problem
```go
func (c *ClientImpl) Query(ctx context.Context, prompt string, sessionID ...string) error {
    c.mu.RLock()
    connected := c.connected
    transport := c.transport
    c.mu.RUnlock()

    if !connected || transport == nil {
        return fmt.Errorf("client not connected")
    }
    // RACE: Transport could be closed by another goroutine here
    return transport.SendMessage(ctx, streamMsg)
}
```

#### Proposed Solution (BREAKING CHANGE)
```go
// Thread-safe operations with proper state management
type Client interface {
    // All operations are atomic and thread-safe
    Send(ctx context.Context, msg Message) error
    Receive(ctx context.Context) <-chan Result[Message]

    // State operations are atomic
    State() ConnectionState

    // Configuration is immutable after creation
}

// Implementation using atomic operations and proper locking
type clientImpl struct {
    state     atomic.Int32 // ConnectionState
    transport atomic.Pointer[Transport]
    config    *ImmutableConfig // Immutable after creation

    // Channels for coordination
    stateCh   chan ConnectionState
    closeCh   chan struct{}
    closeOnce sync.Once
}
```

### 5. API Ergonomics & Go Idioms

#### Current Problem
```go
// Poor API design
func Query(ctx context.Context, prompt string, sessionID ...string) error
// What if multiple sessionIDs? Confusing signature

// Inconsistent parameter patterns
func WithSystemPrompt(prompt string) Option
func WithCwd(cwd string) Option
// No validation, no extensibility
```

#### Proposed Solution (BREAKING CHANGE)
```go
// Clean, extensible options
type QueryRequest struct {
    Content   string
    SessionID string
    Timeout   time.Duration
    Metadata  map[string]string
}

func (c *Client) Query(ctx context.Context, req QueryRequest) error

// Or using builder pattern for complex requests
type QueryBuilder struct {
    req QueryRequest
}

func NewQuery(content string) *QueryBuilder
func (q *QueryBuilder) WithSession(id string) *QueryBuilder
func (q *QueryBuilder) WithTimeout(d time.Duration) *QueryBuilder
func (q *QueryBuilder) WithMetadata(key, value string) *QueryBuilder
func (q *QueryBuilder) Build() QueryRequest

// Usage
err := client.Query(ctx, NewQuery("Hello").
    WithSession("session-1").
    WithTimeout(30*time.Second).
    Build())
```

### 6. Streaming & Real-time Communication

#### Current Problem
```go
// Limited streaming capabilities
func QueryStream(ctx context.Context, messages <-chan StreamMessage) error
func ReceiveMessages(ctx context.Context) <-chan Message

// Missing:
// - Bidirectional streaming coordination
// - Flow control
// - Backpressure handling
// - Stream lifecycle management
```

#### Proposed Solution (BREAKING CHANGE)
```go
// Full-duplex streaming with proper coordination
type Stream interface {
    Send(ctx context.Context, msg Message) error
    Receive(ctx context.Context) <-chan Result[Message]
    Close() error

    // Flow control
    SetSendBufferSize(size int)
    SetReceiveBufferSize(size int)

    // Stream metadata
    ID() string
    CreatedAt() time.Time
    Stats() StreamStats
}

type StreamStats struct {
    MessagesSent     int64
    MessagesReceived int64
    BytesSent        int64
    BytesReceived    int64
    ErrorCount       int64
    LastActivity     time.Time
}

type Client interface {
    // Create managed bidirectional stream
    NewStream(ctx context.Context, opts StreamOptions) (Stream, error)

    // Simple request-response
    Query(ctx context.Context, req QueryRequest) (*QueryResponse, error)

    // Iterator-based for simple cases
    QueryIterator(ctx context.Context, req QueryRequest) (MessageIterator, error)
}

type StreamOptions struct {
    SessionID       string
    SendBufferSize  int
    RecvBufferSize  int
    Timeout         time.Duration
    AutoReconnect   bool
}
```

### 7. Observability & Production Readiness

#### Current Problem
```go
// No observability hooks
type Client interface {
    // Missing: metrics, tracing, structured logging
    // Missing: health checks
    // Missing: debugging support
}
```

#### Proposed Solution (BREAKING CHANGE)
```go
// Built-in observability
type ClientMetrics struct {
    ConnectionAttempts   int64
    ConnectionSuccesses  int64
    ConnectionFailures   int64
    MessagesSent         int64
    MessagesReceived     int64
    BytesSent           int64
    BytesReceived       int64
    ActiveConnections   int64
    AverageLatency      time.Duration
    ErrorRate           float64
}

type HealthStatus struct {
    Healthy   bool
    Message   string
    CheckedAt time.Time
    Details   map[string]interface{}
}

type Observable interface {
    Metrics() ClientMetrics
    Health(ctx context.Context) HealthStatus

    // Structured events for monitoring
    Events() <-chan Event
}

type Event struct {
    Type      EventType
    Timestamp time.Time
    Message   string
    Level     LogLevel
    Fields    map[string]interface{}
}

type Client interface {
    // ... existing methods
    Observable
}

// Middleware pattern for extensibility
type ClientMiddleware func(Client) Client

func WithMetrics(collector MetricsCollector) ClientMiddleware
func WithTracing(tracer trace.Tracer) ClientMiddleware
func WithStructuredLogging(logger Logger) ClientMiddleware
```

### 8. Configuration & Validation

#### Current Problem
```go
// Runtime validation during Connect()
func (c *ClientImpl) validateOptions() error {
    // Should validate at construction time (fail fast)
}

// Weak typing for configuration
type Options struct {
    SystemPrompt *string // Pointer indicates optional, but unclear semantics
    Model        *string
    // Missing validation rules
}
```

#### Proposed Solution (BREAKING CHANGE)
```go
// Strong typing with validation
type Config struct {
    // Required fields
    SystemPrompt    string
    Model          ModelType

    // Optional fields with defaults
    MaxTurns       int           // Default: 10
    Timeout        time.Duration // Default: 30s
    RetryConfig    RetryConfig

    // Validated at construction
    AllowedTools   []ToolName    // Enum-based, validated
    WorkingDir     ValidatedPath // Path validation

    private bool // Prevent external construction
}

type ModelType string
const (
    ModelClaude3Sonnet ModelType = "claude-3-sonnet"
    ModelClaude3Opus   ModelType = "claude-3-opus"
    // Enum prevents invalid model names
)

type ToolName string
const (
    ToolRead  ToolName = "Read"
    ToolWrite ToolName = "Write"
    ToolEdit  ToolName = "Edit"
    // Enum prevents typos in tool names
)

type ConfigBuilder struct {
    config Config
    errors []error
}

func NewConfig() *ConfigBuilder
func (b *ConfigBuilder) WithSystemPrompt(prompt string) *ConfigBuilder
func (b *ConfigBuilder) WithModel(model ModelType) *ConfigBuilder
func (b *ConfigBuilder) WithTimeout(d time.Duration) *ConfigBuilder
func (b *ConfigBuilder) WithAllowedTools(tools ...ToolName) *ConfigBuilder
func (b *ConfigBuilder) Build() (*Config, error) // Validates and returns

// Construction-time validation
func NewClient(config *Config) (Client, error) {
    // Config is already validated, can't fail due to configuration
}
```

## Breaking Changes Summary

### API Surface Changes
1. **Client Interface**: Split into `Connector`, `MessageSender`, `MessageReceiver`
2. **Error Handling**: Introduce `Result[T]` type for consistency
3. **Resource Management**: Add `ConnectionState` and lifecycle management
4. **Configuration**: Replace functional options with validated builder pattern
5. **Streaming**: Replace basic channels with managed `Stream` interface
6. **Observability**: Built-in metrics, health checks, and events

### Migration Path
```go
// Old API (v0.1.0)
client := claudecode.NewClient(
    claudecode.WithSystemPrompt("You are helpful"),
    claudecode.WithAllowedTools("Read", "Write"),
)
err := client.Connect(ctx)
if err != nil { ... }
defer client.Disconnect()

err = client.Query(ctx, "Hello")
for msg := range client.ReceiveMessages(ctx) {
    // Handle message
}

// New API (v0.2.0)
config, err := claudecode.NewConfig().
    WithSystemPrompt("You are helpful").
    WithAllowedTools(claudecode.ToolRead, claudecode.ToolWrite).
    Build()
if err != nil { ... }

client, err := claudecode.NewClient(config)
if err != nil { ... }

err = client.Connect(ctx)
if err != nil { ... }
defer client.Close()

// Type-safe query
resp, err := client.Query(ctx, claudecode.QueryRequest{
    Content: "Hello",
    Timeout: 30 * time.Second,
})

// Or streaming
for result := range client.Receive(ctx) {
    msg, err := result.Unwrap()
    if err != nil { ... }
    // Handle message
}
```

## Implementation Priority

### Phase 1: Core Interface Redesign (High Priority)
- [ ] Split Client interface using composition
- [ ] Implement Result[T] type for error handling
- [ ] Add ConnectionState management
- [ ] Thread-safety improvements

### Phase 2: Configuration & Validation (Medium Priority)
- [ ] Replace functional options with builder pattern
- [ ] Add strong typing with enums
- [ ] Construction-time validation
- [ ] Immutable configuration

### Phase 3: Streaming & Observability (Low Priority)
- [ ] Managed Stream interface
- [ ] Built-in metrics and health checks
- [ ] Event system for monitoring
- [ ] Middleware pattern for extensibility

## Success Metrics

1. **Reduced Cognitive Load**: Smaller, focused interfaces
2. **Improved Safety**: Compile-time error detection
3. **Better Testing**: Easy mocking with segregated interfaces
4. **Production Ready**: Built-in observability and health checks
5. **Developer Experience**: Type safety and clear error messages

## Timeline Estimate

- **Phase 1**: 2-3 weeks (foundation changes)
- **Phase 2**: 1-2 weeks (configuration improvements)
- **Phase 3**: 2-3 weeks (advanced features)

**Total**: 5-8 weeks for complete API overhaul

---

*This analysis represents a comprehensive redesign leveraging Go best practices. Given the project's young status, implementing these breaking changes now will provide a solid foundation for future growth.*