# Go Implementation Guide

Go-specific architecture decisions and implementation patterns for 100% feature parity.

## Core Architecture Decisions

### 1. Concurrency Model

**Go-Native Approach**:
```go
// Use goroutines and channels instead of async/await
type Client struct {
    messages chan Message
    errors   chan error
    done     chan struct{}
}

func (c *Client) Stream(ctx context.Context, prompt string) (<-chan Message, error) {
    go c.processMessages(ctx)
    return c.messages, nil
}
```

**Key Patterns**:
- **One goroutine per client** for I/O operations
- **Channels for streaming** instead of AsyncIterator
- **Context for cancellation** throughout the stack
- **WaitGroup for cleanup** coordination

### 2. Error Handling Strategy

**Go Error Patterns**:
```go
// Wrap errors with context
func (t *SubprocessTransport) Connect(ctx context.Context) error {
    if err := t.startProcess(ctx); err != nil {
        return fmt.Errorf("failed to start Claude CLI: %w", err)
    }
    return nil
}

// Sentinel errors for type checking
var (
    ErrCLINotFound    = errors.New("claude CLI not found")
    ErrConnectionLost = errors.New("connection lost")
    ErrBufferOverflow = errors.New("buffer size exceeded")
)
```

**Error Context Preservation**:
- Use `fmt.Errorf` with `%w` verb for error wrapping
- Custom error types implement `error` interface
- Include contextual information (exit codes, stderr, file paths)

### 3. Type System Design

**Interface-Driven Architecture**:
```go
type Message interface {
    GetType() string
}

type ContentBlock interface {
    GetType() string
}

// Concrete types implement interfaces
type TextBlock struct {
    Text string `json:"text"`
}

func (t *TextBlock) GetType() string { return "text" }
```

**JSON Handling**:
```go
// Use json.RawMessage for delayed parsing
type ParsedMessage struct {
    Type    string          `json:"type"`
    Content json.RawMessage `json:"content"`
}

// Custom unmarshaling for union types
func (cb *ContentBlock) UnmarshalJSON(data []byte) error {
    var temp struct {
        Type string `json:"type"`
    }
    if err := json.Unmarshal(data, &temp); err != nil {
        return err
    }
    
    switch temp.Type {
    case "text":
        var textBlock TextBlock
        if err := json.Unmarshal(data, &textBlock); err != nil {
            return err
        }
        *cb = &textBlock
    // ... other types
    }
    return nil
}
```

### 4. Configuration Patterns

**Functional Options**:
```go
type Option func(*Options)

func WithSystemPrompt(prompt string) Option {
    return func(o *Options) {
        o.SystemPrompt = prompt
    }
}

func WithAllowedTools(tools ...string) Option {
    return func(o *Options) {
        o.AllowedTools = tools
    }
}

// Usage
client := NewClient(
    WithSystemPrompt("You are a helpful assistant"),
    WithAllowedTools("Read", "Write"),
    WithPermissionMode(PermissionModeAcceptEdits),
)
```

## Critical Implementation Details

### 1. Transport Layer

**Subprocess Management** (in `internal/subprocess/`):
```go
// internal/subprocess/transport.go
type Transport struct {
    cmd     *exec.Cmd
    stdin   io.WriteCloser
    stdout  io.ReadCloser
    stderr  *os.File // Use temp file to avoid deadlocks
    
    // Control protocol
    controlRequests  map[string]chan controlResponse
    requestCounter   int64
    
    // Buffering
    jsonBuffer       strings.Builder
    maxBufferSize    int // 1MB limit
}
```

**Process Lifecycle**:
```go
func (t *Transport) terminate(ctx context.Context) error {
    // 5-second SIGTERM → SIGKILL sequence
    if err := t.cmd.Process.Signal(os.Interrupt); err != nil {
        return err
    }
    
    // Wait with timeout
    done := make(chan error, 1)
    go func() {
        done <- t.cmd.Wait()
    }()
    
    select {
    case err := <-done:
        return err
    case <-time.After(5 * time.Second):
        // Force kill
        return t.cmd.Process.Kill()
    case <-ctx.Done():
        return ctx.Err()
    }
}
```

### 2. JSON Processing

**Speculative Parsing** (in `internal/parser/`):
```go
func (t *Parser) processJSONLine(line string) error {
    t.jsonBuffer.WriteString(line)
    
    // Check buffer size limit
    if t.jsonBuffer.Len() > t.maxBufferSize {
        t.jsonBuffer.Reset()
        return ErrBufferOverflow
    }
    
    // Try to parse JSON
    var msg map[string]interface{}
    if err := json.Unmarshal([]byte(t.jsonBuffer.String()), &msg); err != nil {
        // Not complete JSON yet, continue accumulating
        return nil
    }
    
    // Successfully parsed, reset buffer and process
    t.jsonBuffer.Reset()
    return t.handleMessage(msg)
}
```

**Multi-JSON Line Handling**:
```go
func (t *Parser) processLine(line string) error {
    // Handle multiple JSON objects on single line
    jsonLines := strings.Split(strings.TrimSpace(line), "\n")
    for _, jsonLine := range jsonLines {
        jsonLine = strings.TrimSpace(jsonLine)
        if jsonLine == "" {
            continue
        }
        if err := t.processJSONLine(jsonLine); err != nil {
            return err
        }
    }
    return nil
}
```

### 3. CLI Integration

**CLI Discovery** (in `internal/cli/`):
```go
func FindCLI() (string, error) {
    // 1. Check PATH
    if path, err := exec.LookPath("claude"); err == nil {
        return path, nil
    }
    
    // 2. Check standard locations
    home, err := os.UserHomeDir()
    if err != nil {
        return "", err
    }
    
    locations := []string{
        filepath.Join(home, ".npm-global/bin/claude"),
        "/usr/local/bin/claude",
        filepath.Join(home, ".local/bin/claude"),
        filepath.Join(home, "node_modules/.bin/claude"),
        filepath.Join(home, ".yarn/bin/claude"),
    }
    
    for _, path := range locations {
        if info, err := os.Stat(path); err == nil && !info.IsDir() {
            return path, nil
        }
    }
    
    // Check Node.js dependency
    if _, err := exec.LookPath("node"); err != nil {
        return "", fmt.Errorf("%w: Node.js is required but not found. Install from https://nodejs.org/", ErrCLINotFound)
    }
    
    return "", fmt.Errorf("%w: install with 'npm install -g @anthropic-ai/claude-code'", ErrCLINotFound)
}
```

### 4. Interrupt System

**Control Request Protocol** (in `internal/protocol/`):
```go
type ControlRequest struct {
    Type      string `json:"type"`
    RequestID string `json:"request_id"`
    Request   map[string]interface{} `json:"request"`
}

func (c *Controller) SendInterrupt(ctx context.Context) error {
    requestID := fmt.Sprintf("req_%d_%x", 
        atomic.AddInt64(&t.requestCounter, 1),
        rand.Uint32())
    
    request := ControlRequest{
        Type:      "control_request",
        RequestID: requestID,
        Request:   map[string]interface{}{"subtype": "interrupt"},
    }
    
    // Send request
    if err := t.sendJSON(request); err != nil {
        return err
    }
    
    // Wait for response with polling
    responseChan := make(chan controlResponse, 1)
    t.controlRequests[requestID] = responseChan
    defer delete(t.controlRequests, requestID)
    
    ticker := time.NewTicker(100 * time.Millisecond)
    defer ticker.Stop()
    
    for {
        select {
        case response := <-responseChan:
            if response.Subtype == "error" {
                return fmt.Errorf("interrupt failed: %s", response.Error)
            }
            return nil
        case <-ticker.C:
            // Continue polling
        case <-ctx.Done():
            return ctx.Err()
        }
    }
}
```

## Performance Optimizations

### 1. Memory Management

**Object Pooling**:
```go
var messagePool = sync.Pool{
    New: func() interface{} {
        return &Message{}
    },
}

func getMessage() *Message {
    return messagePool.Get().(*Message)
}

func putMessage(msg *Message) {
    // Reset message fields
    *msg = Message{}
    messagePool.Put(msg)
}
```

### 2. Efficient Channels

**Buffered Channels with Backpressure**:
```go
const (
    DefaultChannelBuffer = 100
    MaxChannelBuffer     = 1000
)

func (c *Client) Stream(ctx context.Context, prompt string) (<-chan Message, error) {
    messages := make(chan Message, DefaultChannelBuffer)
    
    go func() {
        defer close(messages)
        // Process with backpressure handling
        for {
            select {
            case messages <- msg: // Non-blocking with buffer
            case <-ctx.Done():
                return
            }
        }
    }()
    
    return messages, nil
}
```

## Testing Strategy

### 1. Mock Transport

```go
type MockTransport struct {
    messages    []Message
    errors      []error
    connected   bool
    interrupted bool
}

func (m *MockTransport) Connect(ctx context.Context) error {
    m.connected = true
    return nil
}

func (m *MockTransport) ReceiveMessages(ctx context.Context) (<-chan Message, <-chan error) {
    msgChan := make(chan Message, len(m.messages))
    errChan := make(chan error, len(m.errors))
    
    go func() {
        for _, msg := range m.messages {
            msgChan <- msg
        }
        close(msgChan)
        
        for _, err := range m.errors {
            errChan <- err
        }
        close(errChan)
    }()
    
    return msgChan, errChan
}
```

This implementation guide provides the foundation for building a Go SDK that achieves 100% feature parity while being idiomatic Go code.