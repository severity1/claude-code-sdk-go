# Breaking Changes Implementation Guide

**Target Version**: v0.2.0
**Breaking Changes**: ✅ Full API Redesign
**Timeline**: 5-8 weeks

## Quick Reference: Before vs After

### Client Creation
```go
// OLD (v0.1.0)
client := claudecode.NewClient(
    claudecode.WithSystemPrompt("You are helpful"),
    claudecode.WithAllowedTools("Read", "Write"),
)

// NEW (v0.2.0)
config, err := claudecode.NewConfig().
    WithSystemPrompt("You are helpful").
    WithAllowedTools(claudecode.ToolRead, claudecode.ToolWrite).
    Build()
if err != nil {
    log.Fatal(err) // Fail fast on configuration errors
}

client, err := claudecode.NewClient(config)
if err != nil {
    log.Fatal(err)
}
```

### Connection Management
```go
// OLD
err := client.Connect(ctx)
if err != nil { ... }
defer client.Disconnect()

// NEW
err := client.Connect(ctx)
if err != nil { ... }
defer client.Close()

// Enhanced: State management
if client.State() != claudecode.StateConnected {
    if err := client.WaitFor(ctx, claudecode.StateConnected); err != nil {
        log.Fatal(err)
    }
}
```

### Message Handling
```go
// OLD
err := client.Query(ctx, "Hello", "session-1")
for msg := range client.ReceiveMessages(ctx) {
    switch m := msg.(type) {
    case *claudecode.AssistantMessage:
        // Handle
    }
}

// NEW
resp, err := client.Query(ctx, claudecode.QueryRequest{
    Content:   "Hello",
    SessionID: "session-1",
    Timeout:   30 * time.Second,
})
if err != nil { ... }

// Or streaming with proper error handling
for result := range client.Receive(ctx) {
    msg, err := result.Unwrap()
    if err != nil {
        log.Printf("Receive error: %v", err)
        continue
    }

    switch m := msg.(type) {
    case *claudecode.AssistantMessage:
        // Handle
    }
}
```

## Implementation Phases

### Phase 1: Core Interface Changes

#### 1.1 New Interface Definitions
Create `interfaces.go`:
```go
package claudecode

import (
    "context"
    "time"
)

// Connection management
type Connector interface {
    Connect(ctx context.Context) error
    Close() error
    State() ConnectionState
    Info() ConnectionInfo
    WaitFor(ctx context.Context, state ConnectionState) error
}

// Message sending
type MessageSender interface {
    Send(ctx context.Context, msg Message) error
    SendStream(ctx context.Context, msgs <-chan Message) <-chan error
    Interrupt(ctx context.Context) error
}

// Message receiving
type MessageReceiver interface {
    Receive(ctx context.Context) <-chan Result[Message]
    ReceiveIterator(ctx context.Context) (MessageIterator, error)
}

// Full client interface
type Client interface {
    Connector
    MessageSender
    MessageReceiver

    // High-level operations
    Query(ctx context.Context, req QueryRequest) (*QueryResponse, error)
    NewStream(ctx context.Context, opts StreamOptions) (Stream, error)

    // Observability
    Observable
}
```

#### 1.2 Result Type Implementation
Create `result.go`:
```go
package claudecode

// Generic result type for channel-based operations
type Result[T any] struct {
    Value T
    Err   error
}

func Ok[T any](value T) Result[T] {
    return Result[T]{Value: value}
}

func Err[T any](err error) Result[T] {
    return Result[T]{Err: err}
}

func (r Result[T]) Unwrap() (T, error) {
    return r.Value, r.Err
}

func (r Result[T]) IsOk() bool {
    return r.Err == nil
}

func (r Result[T]) IsErr() bool {
    return r.Err != nil
}

// Transform result if Ok
func (r Result[T]) Map(fn func(T) T) Result[T] {
    if r.IsErr() {
        return r
    }
    return Ok(fn(r.Value))
}
```

#### 1.3 Connection State Management
Create `connection.go`:
```go
package claudecode

import (
    "sync/atomic"
    "time"
)

type ConnectionState int32

const (
    StateDisconnected ConnectionState = iota
    StateConnecting
    StateConnected
    StateReconnecting
    StateDisconnecting
    StateError
)

func (s ConnectionState) String() string {
    switch s {
    case StateDisconnected:
        return "disconnected"
    case StateConnecting:
        return "connecting"
    case StateConnected:
        return "connected"
    case StateReconnecting:
        return "reconnecting"
    case StateDisconnecting:
        return "disconnecting"
    case StateError:
        return "error"
    default:
        return "unknown"
    }
}

type ConnectionInfo struct {
    State        ConnectionState
    ConnectedAt  time.Time
    LastActivity time.Time
    ErrorCount   int64
    LastError    error
    Uptime       time.Duration
}

// Thread-safe state management
type connectionState struct {
    current   atomic.Int32
    info      atomic.Pointer[ConnectionInfo]
    watchers  []chan ConnectionState
    watcherMu sync.RWMutex
}

func (cs *connectionState) Get() ConnectionState {
    return ConnectionState(cs.current.Load())
}

func (cs *connectionState) Set(state ConnectionState) {
    old := ConnectionState(cs.current.Swap(int32(state)))
    if old != state {
        cs.notifyWatchers(state)
    }
}

func (cs *connectionState) Watch() <-chan ConnectionState {
    cs.watcherMu.Lock()
    defer cs.watcherMu.Unlock()

    ch := make(chan ConnectionState, 1)
    cs.watchers = append(cs.watchers, ch)
    return ch
}

func (cs *connectionState) notifyWatchers(state ConnectionState) {
    cs.watcherMu.RLock()
    defer cs.watcherMu.RUnlock()

    for _, ch := range cs.watchers {
        select {
        case ch <- state:
        default: // Don't block if watcher is slow
        }
    }
}
```

### Phase 2: Configuration System

#### 2.1 Strong Typing with Enums
Create `types.go`:
```go
package claudecode

type ModelType string

const (
    ModelClaude3Sonnet   ModelType = "claude-3-sonnet-20240229"
    ModelClaude3Opus     ModelType = "claude-3-opus-20240229"
    ModelClaude3Haiku    ModelType = "claude-3-haiku-20240307"
    ModelClaude35Sonnet  ModelType = "claude-3-5-sonnet-20241022"
)

func (m ModelType) Validate() error {
    switch m {
    case ModelClaude3Sonnet, ModelClaude3Opus, ModelClaude3Haiku, ModelClaude35Sonnet:
        return nil
    default:
        return fmt.Errorf("invalid model type: %s", m)
    }
}

type ToolName string

const (
    ToolRead    ToolName = "Read"
    ToolWrite   ToolName = "Write"
    ToolEdit    ToolName = "Edit"
    ToolBash    ToolName = "Bash"
    ToolGlob    ToolName = "Glob"
    ToolGrep    ToolName = "Grep"
)

func (t ToolName) Validate() error {
    switch t {
    case ToolRead, ToolWrite, ToolEdit, ToolBash, ToolGlob, ToolGrep:
        return nil
    default:
        return fmt.Errorf("invalid tool name: %s", t)
    }
}

type PermissionMode string

const (
    PermissionDefault           PermissionMode = "default"
    PermissionAcceptEdits       PermissionMode = "accept-edits"
    PermissionPlan              PermissionMode = "plan"
    PermissionBypassPermissions PermissionMode = "bypass-permissions"
)
```

#### 2.2 Configuration Builder
Create `config.go`:
```go
package claudecode

import (
    "errors"
    "fmt"
    "os"
    "path/filepath"
    "time"
)

type Config struct {
    // Required fields
    SystemPrompt string
    Model        ModelType

    // Optional with defaults
    MaxTurns        int
    Timeout         time.Duration
    AllowedTools    []ToolName
    DisallowedTools []ToolName
    WorkingDir      string
    PermissionMode  PermissionMode

    // Advanced options
    RetryConfig     RetryConfig
    McpServers      map[string]McpServerConfig
    ExtraArgs       map[string]string

    // Private field prevents external construction
    validated bool
}

type RetryConfig struct {
    Enabled    bool
    MaxRetries int
    BaseDelay  time.Duration
    MaxDelay   time.Duration
}

type ConfigBuilder struct {
    config Config
    errors []error
}

func NewConfig() *ConfigBuilder {
    return &ConfigBuilder{
        config: Config{
            Model:          ModelClaude35Sonnet, // Sensible default
            MaxTurns:       10,
            Timeout:        30 * time.Second,
            WorkingDir:     ".", // Current directory
            PermissionMode: PermissionDefault,
            RetryConfig: RetryConfig{
                Enabled:    true,
                MaxRetries: 3,
                BaseDelay:  time.Second,
                MaxDelay:   30 * time.Second,
            },
        },
    }
}

func (b *ConfigBuilder) WithSystemPrompt(prompt string) *ConfigBuilder {
    if prompt == "" {
        b.errors = append(b.errors, errors.New("system prompt cannot be empty"))
    } else {
        b.config.SystemPrompt = prompt
    }
    return b
}

func (b *ConfigBuilder) WithModel(model ModelType) *ConfigBuilder {
    if err := model.Validate(); err != nil {
        b.errors = append(b.errors, fmt.Errorf("invalid model: %w", err))
    } else {
        b.config.Model = model
    }
    return b
}

func (b *ConfigBuilder) WithTimeout(d time.Duration) *ConfigBuilder {
    if d <= 0 {
        b.errors = append(b.errors, errors.New("timeout must be positive"))
    } else {
        b.config.Timeout = d
    }
    return b
}

func (b *ConfigBuilder) WithAllowedTools(tools ...ToolName) *ConfigBuilder {
    for _, tool := range tools {
        if err := tool.Validate(); err != nil {
            b.errors = append(b.errors, fmt.Errorf("invalid tool: %w", err))
            continue
        }
    }
    b.config.AllowedTools = tools
    return b
}

func (b *ConfigBuilder) WithWorkingDir(dir string) *ConfigBuilder {
    absDir, err := filepath.Abs(dir)
    if err != nil {
        b.errors = append(b.errors, fmt.Errorf("invalid working directory: %w", err))
        return b
    }

    if _, err := os.Stat(absDir); os.IsNotExist(err) {
        b.errors = append(b.errors, fmt.Errorf("working directory does not exist: %s", absDir))
    } else {
        b.config.WorkingDir = absDir
    }
    return b
}

func (b *ConfigBuilder) Build() (*Config, error) {
    // Validate required fields
    if b.config.SystemPrompt == "" {
        b.errors = append(b.errors, errors.New("system prompt is required"))
    }

    // Validate tool conflicts
    if len(b.config.AllowedTools) > 0 && len(b.config.DisallowedTools) > 0 {
        // Check for overlaps
        allowed := make(map[ToolName]bool)
        for _, tool := range b.config.AllowedTools {
            allowed[tool] = true
        }

        for _, tool := range b.config.DisallowedTools {
            if allowed[tool] {
                b.errors = append(b.errors, fmt.Errorf("tool %s is both allowed and disallowed", tool))
            }
        }
    }

    if len(b.errors) > 0 {
        return nil, fmt.Errorf("configuration validation failed: %v", b.errors)
    }

    config := b.config
    config.validated = true
    return &config, nil
}
```

### Phase 3: Streaming & Observability

#### 3.1 Stream Interface
Create `stream.go`:
```go
package claudecode

type Stream interface {
    ID() string
    Send(ctx context.Context, msg Message) error
    Receive(ctx context.Context) <-chan Result[Message]
    Close() error

    // Flow control
    SetSendBufferSize(size int)
    SetReceiveBufferSize(size int)

    // Metadata
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
    Duration         time.Duration
}

type StreamOptions struct {
    SessionID       string
    SendBufferSize  int
    RecvBufferSize  int
    Timeout         time.Duration
    AutoReconnect   bool
    Metadata        map[string]string
}
```

#### 3.2 Observability
Create `observability.go`:
```go
package claudecode

type Observable interface {
    Metrics() ClientMetrics
    Health(ctx context.Context) HealthCheck
    Events() <-chan Event
}

type ClientMetrics struct {
    // Connection metrics
    ConnectionAttempts   int64
    ConnectionSuccesses  int64
    ConnectionFailures   int64
    CurrentConnections   int64

    // Message metrics
    MessagesSent         int64
    MessagesReceived     int64
    BytesSent           int64
    BytesReceived       int64

    // Performance metrics
    AverageLatency      time.Duration
    P95Latency         time.Duration
    ErrorRate          float64
    Uptime             time.Duration
}

type HealthCheck struct {
    Healthy     bool
    Status      string
    CheckedAt   time.Time
    Details     map[string]interface{}
    Duration    time.Duration
}

type Event struct {
    Type      EventType
    Timestamp time.Time
    Level     LogLevel
    Message   string
    Fields    map[string]interface{}
}

type EventType string
const (
    EventConnection EventType = "connection"
    EventMessage    EventType = "message"
    EventError      EventType = "error"
    EventHealth     EventType = "health"
)

type LogLevel string
const (
    LogLevelDebug LogLevel = "debug"
    LogLevelInfo  LogLevel = "info"
    LogLevelWarn  LogLevel = "warn"
    LogLevelError LogLevel = "error"
)
```

## Migration Strategy

### Step 1: Parallel Implementation
- Keep old APIs alongside new ones
- Add deprecation warnings to old APIs
- Provide migration utility functions

### Step 2: Example Updates
- Update all examples to use new APIs
- Provide migration examples in documentation
- Add integration tests for new APIs

### Step 3: Documentation
- Update README with new examples
- Create migration guide
- Update API documentation

### Step 4: Release
- Release as v0.2.0 with breaking changes
- Archive v0.1.x branch for historical reference
- Update CI/CD pipelines

## Testing Strategy

### Unit Tests
```go
func TestNewConfigValidation(t *testing.T) {
    tests := []struct {
        name    string
        builder func() *ConfigBuilder
        wantErr bool
    }{
        {
            name: "valid config",
            builder: func() *ConfigBuilder {
                return NewConfig().
                    WithSystemPrompt("You are helpful").
                    WithModel(ModelClaude35Sonnet)
            },
            wantErr: false,
        },
        {
            name: "empty system prompt",
            builder: func() *ConfigBuilder {
                return NewConfig().WithSystemPrompt("")
            },
            wantErr: true,
        },
        {
            name: "invalid model",
            builder: func() *ConfigBuilder {
                return NewConfig().WithModel("invalid-model")
            },
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            _, err := tt.builder().Build()
            if (err != nil) != tt.wantErr {
                t.Errorf("Build() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### Integration Tests
```go
func TestClientLifecycle(t *testing.T) {
    config, err := NewConfig().
        WithSystemPrompt("You are helpful").
        WithTimeout(10 * time.Second).
        Build()
    require.NoError(t, err)

    client, err := NewClient(config)
    require.NoError(t, err)

    // Test connection lifecycle
    ctx := context.Background()

    assert.Equal(t, StateDisconnected, client.State())

    err = client.Connect(ctx)
    require.NoError(t, err)
    defer client.Close()

    assert.Equal(t, StateConnected, client.State())

    // Test query functionality
    resp, err := client.Query(ctx, QueryRequest{
        Content: "Hello",
        Timeout: 5 * time.Second,
    })
    require.NoError(t, err)
    assert.NotNil(t, resp)
}
```

This comprehensive breaking changes guide provides the roadmap for transforming the Claude Code SDK Go into a production-ready, idiomatic Go library.