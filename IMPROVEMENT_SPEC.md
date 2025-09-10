# Claude Code Go SDK Improvement Specification

## Executive Summary

This specification defines the improvements required to achieve full Go idiomatic compliance in the Claude Code Go SDK, addressing the critical issues identified in our interface analysis (Grade: C+) while preserving the excellent API design (Grade: A+).

**Objective**: Transform the SDK into a exemplary Go codebase that serves as a reference implementation for idiomatic Go design patterns.

**Scope**: 
- **Critical**: Interface redesign to eliminate weak contracts and `interface{}` usage
- **High**: Enhanced type safety throughout the codebase
- **Medium**: Minor API enhancements for completeness
- **Low**: Documentation and example improvements

## 1. Interface Redesign Specification

### 1.1 Core Message Interface Enhancement

#### Current Problem
```go
// ❌ Current: Weak contract forcing type assertions
type Message interface {
    Type() string  // Only method - requires casting everywhere
}
```

#### Proposed Solution
```go
// ✅ Proposed: Rich behavioral interface
package claudecode

import (
    "encoding/json"
    "fmt"
    "time"
)

// Message represents any message in the conversation with rich behavior
type Message interface {
    // Core identification
    Type() string
    GetID() string                  // Unique message identifier
    GetTimestamp() time.Time         // When message was created
    GetSessionID() string            // Session association
    
    // Content access without type assertions
    GetContent() ContentAccessor     // Uniform content access
    GetMetadata() MessageMetadata    // Additional context
    
    // Behavioral methods
    Validate() error                 // Self-validation
    IsEmpty() bool                   // Check if message has content
    
    // Serialization
    json.Marshaler
    json.Unmarshaler
    fmt.Stringer                     // Human-readable representation
}

// ContentAccessor provides uniform access to message content
type ContentAccessor interface {
    // Access methods that work for all message types
    AsText() string                  // Get text representation
    AsBlocks() []ContentBlock        // Get structured blocks
    GetBlockCount() int              // Number of content blocks
    GetBlock(index int) (ContentBlock, error)
    FilterByType(blockType string) []ContentBlock
    
    // Content inspection
    IsEmpty() bool
    GetSize() int                   // Total content size in bytes
    ContainsType(blockType string) bool
}

// MessageMetadata contains common metadata for all messages
type MessageMetadata struct {
    ID          string              `json:"id"`
    SessionID   string              `json:"session_id"`
    RequestID   string              `json:"request_id,omitempty"`
    Timestamp   time.Time           `json:"timestamp"`
    Version     string              `json:"version,omitempty"`
    Model       string              `json:"model,omitempty"`
    Usage       *UsageMetrics       `json:"usage,omitempty"`
}

// UsageMetrics tracks resource usage
type UsageMetrics struct {
    InputTokens    int              `json:"input_tokens"`
    OutputTokens   int              `json:"output_tokens"`
    ThinkingTokens int              `json:"thinking_tokens,omitempty"`
    TotalTokens    int              `json:"total_tokens"`
    CostUSD        *float64         `json:"cost_usd,omitempty"`
}
```

#### Implementation for Concrete Types
```go
// UserMessage with rich interface implementation
type UserMessage struct {
    metadata    MessageMetadata
    content     UserContent
}

// UserContent replaces interface{} with type-safe union
type UserContent struct {
    text   *string
    blocks []ContentBlock
}

// Implement Message interface
func (m *UserMessage) Type() string { return MessageTypeUser }
func (m *UserMessage) GetID() string { return m.metadata.ID }
func (m *UserMessage) GetTimestamp() time.Time { return m.metadata.Timestamp }
func (m *UserMessage) GetSessionID() string { return m.metadata.SessionID }
func (m *UserMessage) GetMetadata() MessageMetadata { return m.metadata }

// GetContent returns a content accessor
func (m *UserMessage) GetContent() ContentAccessor {
    return &userContentAccessor{content: m.content}
}

// Validate ensures message is well-formed
func (m *UserMessage) Validate() error {
    if m.metadata.ID == "" {
        return fmt.Errorf("message ID is required")
    }
    if m.content.IsEmpty() {
        return fmt.Errorf("user message cannot be empty")
    }
    return nil
}

func (m *UserMessage) IsEmpty() bool {
    return m.content.IsEmpty()
}

func (m *UserMessage) String() string {
    return fmt.Sprintf("UserMessage[%s]: %s", m.metadata.ID, m.content.AsText())
}

// userContentAccessor implements ContentAccessor for UserMessage
type userContentAccessor struct {
    content UserContent
}

func (a *userContentAccessor) AsText() string {
    if a.content.text != nil {
        return *a.content.text
    }
    // Convert blocks to text
    var texts []string
    for _, block := range a.content.blocks {
        if tb, ok := block.(TextProvider); ok {
            texts = append(texts, tb.GetText())
        }
    }
    return strings.Join(texts, "\n")
}

func (a *userContentAccessor) AsBlocks() []ContentBlock {
    if a.content.blocks != nil {
        return a.content.blocks
    }
    // Convert text to TextBlock
    if a.content.text != nil {
        return []ContentBlock{
            &TextBlock{Text: *a.content.text},
        }
    }
    return nil
}

// UserContent methods for type-safe access
func (c *UserContent) IsText() bool { return c.text != nil }
func (c *UserContent) IsBlocks() bool { return c.blocks != nil }
func (c *UserContent) IsEmpty() bool { 
    return c.text == nil && len(c.blocks) == 0 
}

func (c *UserContent) AsText() string {
    if c.text != nil {
        return *c.text
    }
    // Convert blocks to text...
    return ""
}
```

### 1.2 ContentBlock Interface Enhancement

#### Current Problem
```go
// ❌ Current: Minimal interface requiring type assertions
type ContentBlock interface {
    BlockType() string  // Only method
}
```

#### Proposed Solution
```go
// ✅ Proposed: Rich behavioral interfaces with composition
package claudecode

// ContentBlock base interface for all content blocks
type ContentBlock interface {
    // Core identification
    BlockType() string
    GetID() string                  // Unique block identifier
    
    // Content inspection
    IsEmpty() bool
    GetSize() int                   // Size in bytes
    Validate() error                // Self-validation
    
    // Serialization
    json.Marshaler
    json.Unmarshaler
    fmt.Stringer
}

// TextProvider for blocks containing text
type TextProvider interface {
    ContentBlock
    GetText() string
    SetText(text string) error
    GetLength() int                 // Character count
    GetWordCount() int              // Word count
    Contains(substring string) bool
}

// ToolProvider for tool-related blocks
type ToolProvider interface {
    ContentBlock
    GetToolName() string
    GetToolID() string
    GetInput() ToolInput            // Type-safe input
    GetOutput() ToolOutput           // Type-safe output
    IsError() bool
    GetError() error
}

// ThinkingProvider for thinking blocks
type ThinkingProvider interface {
    ContentBlock
    GetThinking() string
    GetSignature() string
    GetConfidence() float64          // 0.0 to 1.0
    GetDuration() time.Duration      // Thinking time
}

// Type-safe tool input/output
type ToolInput struct {
    Parameters map[string]interface{} `json:"parameters"`
    validated  bool
}

func (ti *ToolInput) Validate(schema ToolSchema) error {
    // Validate against tool schema
    ti.validated = true
    return nil
}

type ToolOutput struct {
    Result    interface{} `json:"result"`
    Error     *ToolError  `json:"error,omitempty"`
    Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

type ToolError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Details string `json:"details,omitempty"`
}
```

#### Concrete Implementation Example
```go
// TextBlock with full interface implementation
type TextBlock struct {
    id        string
    text      string
    metadata  map[string]interface{}
}

func (b *TextBlock) BlockType() string { return ContentBlockTypeText }
func (b *TextBlock) GetID() string { return b.id }
func (b *TextBlock) IsEmpty() bool { return b.text == "" }
func (b *TextBlock) GetSize() int { return len(b.text) }

func (b *TextBlock) Validate() error {
    if b.id == "" {
        return fmt.Errorf("block ID is required")
    }
    if b.text == "" {
        return fmt.Errorf("text block cannot be empty")
    }
    return nil
}

func (b *TextBlock) String() string {
    if len(b.text) > 50 {
        return fmt.Sprintf("TextBlock[%s]: %s...", b.id, b.text[:50])
    }
    return fmt.Sprintf("TextBlock[%s]: %s", b.id, b.text)
}

// TextProvider implementation
func (b *TextBlock) GetText() string { return b.text }
func (b *TextBlock) SetText(text string) error {
    if text == "" {
        return fmt.Errorf("text cannot be empty")
    }
    b.text = text
    return nil
}

func (b *TextBlock) GetLength() int { return len(b.text) }
func (b *TextBlock) GetWordCount() int {
    return len(strings.Fields(b.text))
}
func (b *TextBlock) Contains(substring string) bool {
    return strings.Contains(b.text, substring)
}

// MarshalJSON for proper serialization
func (b *TextBlock) MarshalJSON() ([]byte, error) {
    return json.Marshal(struct {
        Type     string                 `json:"type"`
        ID       string                 `json:"id,omitempty"`
        Text     string                 `json:"text"`
        Metadata map[string]interface{} `json:"metadata,omitempty"`
    }{
        Type:     b.BlockType(),
        ID:       b.id,
        Text:     b.text,
        Metadata: b.metadata,
    })
}
```

### 1.3 Eliminating interface{} Anti-Pattern

#### Current Problems
```go
// ❌ Current: Multiple interface{} violations
type UserMessage struct {
    Content interface{} `json:"content"`  // string or []ContentBlock
}

type StreamMessage struct {
    Message  interface{}            `json:"message,omitempty"`
    Request  map[string]interface{} `json:"request,omitempty"`
    Response map[string]interface{} `json:"response,omitempty"`
}

type ToolResultBlock struct {
    Content interface{} `json:"content"`  // string or structured data
}
```

#### Proposed Solutions

##### Solution 1: Type-Safe Union Types
```go
// ✅ UserContent with explicit union type
type UserContent struct {
    text   *string        // When content is plain text
    blocks []ContentBlock // When content is structured
}

// Safe accessors with clear semantics
func (c *UserContent) IsText() bool { return c.text != nil }
func (c *UserContent) IsBlocks() bool { return c.blocks != nil }

func (c *UserContent) GetText() (string, bool) {
    if c.text != nil {
        return *c.text, true
    }
    return "", false
}

func (c *UserContent) GetBlocks() ([]ContentBlock, bool) {
    if c.blocks != nil {
        return c.blocks, true
    }
    return nil, false
}

// JSON marshaling handles both cases
func (c *UserContent) MarshalJSON() ([]byte, error) {
    if c.text != nil {
        return json.Marshal(c.text)
    }
    if c.blocks != nil {
        return json.Marshal(c.blocks)
    }
    return json.Marshal(nil)
}

// JSON unmarshaling with type detection
func (c *UserContent) UnmarshalJSON(data []byte) error {
    // Try string first
    var text string
    if err := json.Unmarshal(data, &text); err == nil {
        c.text = &text
        return nil
    }
    
    // Try blocks array
    var blocks []json.RawMessage
    if err := json.Unmarshal(data, &blocks); err == nil {
        c.blocks = make([]ContentBlock, 0, len(blocks))
        for _, raw := range blocks {
            block, err := parseContentBlock(raw)
            if err != nil {
                return err
            }
            c.blocks = append(c.blocks, block)
        }
        return nil
    }
    
    return fmt.Errorf("content must be string or array of blocks")
}
```

##### Solution 2: StreamMessage with Type-Safe Fields
```go
// ✅ StreamMessage with proper types
type StreamMessage struct {
    Type            string           `json:"type"`
    Message         MessagePayload   `json:"message,omitempty"`
    ParentToolUseID *string          `json:"parent_tool_use_id,omitempty"`
    SessionID       string           `json:"session_id,omitempty"`
    RequestID       string           `json:"request_id,omitempty"`
    Request         *RequestPayload  `json:"request,omitempty"`
    Response        *ResponsePayload `json:"response,omitempty"`
}

// Type-safe message payload
type MessagePayload struct {
    UserMessage      *UserMessage      `json:"-"`
    AssistantMessage *AssistantMessage `json:"-"`
    SystemMessage    *SystemMessage    `json:"-"`
    ResultMessage    *ResultMessage    `json:"-"`
}

func (mp *MessagePayload) GetMessage() Message {
    if mp.UserMessage != nil {
        return mp.UserMessage
    }
    if mp.AssistantMessage != nil {
        return mp.AssistantMessage
    }
    if mp.SystemMessage != nil {
        return mp.SystemMessage
    }
    if mp.ResultMessage != nil {
        return mp.ResultMessage
    }
    return nil
}

// MarshalJSON serializes the active message
func (mp *MessagePayload) MarshalJSON() ([]byte, error) {
    if msg := mp.GetMessage(); msg != nil {
        return json.Marshal(msg)
    }
    return json.Marshal(nil)
}

// UnmarshalJSON with type detection
func (mp *MessagePayload) UnmarshalJSON(data []byte) error {
    var raw map[string]json.RawMessage
    if err := json.Unmarshal(data, &raw); err != nil {
        return err
    }
    
    typeRaw, ok := raw["type"]
    if !ok {
        return fmt.Errorf("message missing type field")
    }
    
    var msgType string
    if err := json.Unmarshal(typeRaw, &msgType); err != nil {
        return err
    }
    
    switch msgType {
    case MessageTypeUser:
        mp.UserMessage = &UserMessage{}
        return json.Unmarshal(data, mp.UserMessage)
    case MessageTypeAssistant:
        mp.AssistantMessage = &AssistantMessage{}
        return json.Unmarshal(data, mp.AssistantMessage)
    case MessageTypeSystem:
        mp.SystemMessage = &SystemMessage{}
        return json.Unmarshal(data, mp.SystemMessage)
    case MessageTypeResult:
        mp.ResultMessage = &ResultMessage{}
        return json.Unmarshal(data, mp.ResultMessage)
    default:
        return fmt.Errorf("unknown message type: %s", msgType)
    }
}

// Type-safe request/response payloads
type RequestPayload struct {
    Method     string                 `json:"method"`
    Parameters map[string]interface{} `json:"parameters"`
    validated  bool
}

type ResponsePayload struct {
    Status  int                    `json:"status"`
    Data    json.RawMessage        `json:"data"`
    Error   *ResponseError         `json:"error,omitempty"`
}

type ResponseError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Details string `json:"details,omitempty"`
}
```

##### Solution 3: ToolResultBlock with Type-Safe Content
```go
// ✅ ToolResultBlock with type-safe content
type ToolResultBlock struct {
    id        string
    toolUseID string
    content   ToolResultContent
    isError   bool
    error     error
}

// Type-safe tool result content
type ToolResultContent struct {
    text       *string
    structured *StructuredData
    binary     *BinaryData
}

type StructuredData struct {
    Format string          `json:"format"` // "json", "yaml", "xml"
    Data   json.RawMessage `json:"data"`
}

type BinaryData struct {
    MimeType string `json:"mime_type"`
    Encoding string `json:"encoding"` // "base64", "hex"
    Data     []byte `json:"data"`
}

// Safe accessors
func (c *ToolResultContent) IsText() bool { return c.text != nil }
func (c *ToolResultContent) IsStructured() bool { return c.structured != nil }
func (c *ToolResultContent) IsBinary() bool { return c.binary != nil }

func (c *ToolResultContent) GetText() (string, bool) {
    if c.text != nil {
        return *c.text, true
    }
    return "", false
}

func (c *ToolResultContent) GetStructured() (*StructuredData, bool) {
    if c.structured != nil {
        return c.structured, true
    }
    return nil, false
}

func (c *ToolResultContent) GetBinary() (*BinaryData, bool) {
    if c.binary != nil {
        return c.binary, true
    }
    return nil, false
}

// AsText converts any content type to text representation
func (c *ToolResultContent) AsText() string {
    if c.text != nil {
        return *c.text
    }
    if c.structured != nil {
        // Convert structured data to text
        return string(c.structured.Data)
    }
    if c.binary != nil {
        return fmt.Sprintf("[Binary data: %s, %d bytes]", 
            c.binary.MimeType, len(c.binary.Data))
    }
    return ""
}
```

## 2. Enhanced Transport Interface

### 2.1 Transport Interface Improvements

#### Current Interface
```go
// types.go:63-69 - Current transport interface
type Transport interface {
    Connect(ctx context.Context) error
    SendMessage(ctx context.Context, message StreamMessage) error
    ReceiveMessages(ctx context.Context) (<-chan Message, <-chan error)
    Interrupt(ctx context.Context) error
    Close() error
}
```

#### Proposed Enhanced Interface
```go
// ✅ Enhanced transport with diagnostics and control
type Transport interface {
    // Connection lifecycle
    Connect(ctx context.Context) error
    Disconnect(ctx context.Context) error
    Reconnect(ctx context.Context) error
    Close() error
    
    // Connection state
    IsConnected() bool
    GetConnectionInfo() ConnectionInfo
    HealthCheck(ctx context.Context) error
    
    // Communication
    SendMessage(ctx context.Context, message StreamMessage) error
    ReceiveMessages(ctx context.Context) (<-chan Message, <-chan error)
    
    // Control operations
    Interrupt(ctx context.Context) error
    Reset(ctx context.Context) error
    
    // Configuration
    SetTimeout(timeout time.Duration) error
    SetRetryPolicy(policy RetryPolicy) error
    
    // Diagnostics
    GetMetrics() TransportMetrics
    GetLastError() error
    EnableDebug(enable bool)
}

// ConnectionInfo provides connection details
type ConnectionInfo struct {
    ID           string            `json:"id"`
    Method       string            `json:"method"`       // "subprocess", "websocket", etc.
    RemoteAddr   string            `json:"remote_addr"`
    LocalAddr    string            `json:"local_addr"`
    ConnectedAt  time.Time         `json:"connected_at"`
    LastActivity time.Time         `json:"last_activity"`
    Latency      time.Duration     `json:"latency"`
    Version      string            `json:"version"`
    Capabilities []string          `json:"capabilities"`
    TLS          *TLSInfo          `json:"tls,omitempty"`
}

// TLSInfo for secure connections
type TLSInfo struct {
    Version     uint16   `json:"version"`
    CipherSuite uint16   `json:"cipher_suite"`
    ServerName  string   `json:"server_name"`
    Verified    bool     `json:"verified"`
}

// TransportMetrics tracks performance
type TransportMetrics struct {
    MessagesSent       int64         `json:"messages_sent"`
    MessagesReceived   int64         `json:"messages_received"`
    BytesSent          int64         `json:"bytes_sent"`
    BytesReceived      int64         `json:"bytes_received"`
    ErrorCount         int64         `json:"error_count"`
    ReconnectCount     int64         `json:"reconnect_count"`
    AverageLatency     time.Duration `json:"average_latency"`
    MaxLatency         time.Duration `json:"max_latency"`
    MinLatency         time.Duration `json:"min_latency"`
    LastError          error         `json:"last_error,omitempty"`
    LastErrorAt        time.Time     `json:"last_error_at,omitempty"`
}

// RetryPolicy configures retry behavior
type RetryPolicy struct {
    MaxAttempts     int           `json:"max_attempts"`
    InitialDelay    time.Duration `json:"initial_delay"`
    MaxDelay        time.Duration `json:"max_delay"`
    Multiplier      float64       `json:"multiplier"`
    Jitter          float64       `json:"jitter"`        // 0.0 to 1.0
    RetryableErrors []string      `json:"retryable_errors"`
}
```

## 3. MessageIterator Enhancement

### 3.1 Enhanced Iterator Interface

#### Current Interface
```go
// internal/shared/stream.go:17-20 - Current basic iterator
type MessageIterator interface {
    Next(ctx context.Context) (Message, error)
    Close() error
}
```

#### Proposed Enhanced Interface
```go
// ✅ Enhanced iterator with richer functionality
type MessageIterator interface {
    // Core iteration
    Next(ctx context.Context) (Message, error)
    
    // Non-blocking operations
    HasNext(ctx context.Context) bool
    Peek(ctx context.Context) (Message, error)
    
    // Batch operations
    NextBatch(ctx context.Context, maxSize int) ([]Message, error)
    Remaining() int  // -1 if unknown
    
    // Navigation
    Skip(n int) error
    Reset() error  // If supported
    
    // State inspection
    Position() int64
    IsExhausted() bool
    LastError() error
    
    // Resource management
    Close() error
}

// BufferedIterator for performance
type BufferedIterator interface {
    MessageIterator
    SetBufferSize(size int) error
    FlushBuffer() error
    BufferStats() BufferStats
}

type BufferStats struct {
    Size       int           `json:"size"`
    Used       int           `json:"used"`
    Available  int           `json:"available"`
    HitRate    float64       `json:"hit_rate"`
    MissRate   float64       `json:"miss_rate"`
    FlushCount int64         `json:"flush_count"`
}
```

## 4. API Enhancements

### 4.1 Additional Client Methods

```go
// Enhanced Client interface
type Client interface {
    // Existing methods...
    
    // Session management
    GetSessionID() string
    SetSessionID(sessionID string) error
    ListSessions(ctx context.Context) ([]SessionInfo, error)
    DeleteSession(ctx context.Context, sessionID string) error
    
    // State inspection
    GetState() ClientState
    GetOptions() Options
    UpdateOptions(opts ...Option) error
    
    // Metrics and diagnostics
    GetMetrics() ClientMetrics
    GetTransport() Transport
    
    // Advanced operations
    QueryWithCallback(ctx context.Context, prompt string, callback MessageCallback) error
    QueryBatch(ctx context.Context, prompts []string) ([]MessageIterator, error)
    
    // Event handling
    OnConnect(handler ConnectHandler)
    OnDisconnect(handler DisconnectHandler)
    OnError(handler ErrorHandler)
}

// ClientState represents client state
type ClientState string

const (
    ClientStateDisconnected ClientState = "disconnected"
    ClientStateConnecting   ClientState = "connecting"
    ClientStateConnected    ClientState = "connected"
    ClientStateDisconnecting ClientState = "disconnecting"
    ClientStateError        ClientState = "error"
)

// SessionInfo provides session details
type SessionInfo struct {
    ID          string        `json:"id"`
    CreatedAt   time.Time     `json:"created_at"`
    UpdatedAt   time.Time     `json:"updated_at"`
    MessageCount int          `json:"message_count"`
    Metadata    map[string]interface{} `json:"metadata"`
}

// ClientMetrics tracks client performance
type ClientMetrics struct {
    QueriesCount      int64         `json:"queries_count"`
    MessagesProcessed int64         `json:"messages_processed"`
    AverageQueryTime  time.Duration `json:"average_query_time"`
    SessionDuration   time.Duration `json:"session_duration"`
    ErrorRate         float64       `json:"error_rate"`
}

// Callback types for event handling
type MessageCallback func(message Message) error
type ConnectHandler func(info ConnectionInfo)
type DisconnectHandler func(reason error)
type ErrorHandler func(err error) bool  // Return true to retry
```

### 4.2 New Utility Functions

```go
// Utility functions for common patterns
package claudecode

// QueryWithTimeout executes query with specific timeout
func QueryWithTimeout(ctx context.Context, prompt string, timeout time.Duration, opts ...Option) (MessageIterator, error) {
    ctx, cancel := context.WithTimeout(ctx, timeout)
    defer cancel()
    return Query(ctx, prompt, opts...)
}

// QueryWithRetry executes query with retry logic
func QueryWithRetry(ctx context.Context, prompt string, maxAttempts int, opts ...Option) (MessageIterator, error) {
    var lastErr error
    for i := 0; i < maxAttempts; i++ {
        iter, err := Query(ctx, prompt, opts...)
        if err == nil {
            return iter, nil
        }
        lastErr = err
        
        // Exponential backoff
        delay := time.Duration(math.Pow(2, float64(i))) * time.Second
        select {
        case <-time.After(delay):
        case <-ctx.Done():
            return nil, ctx.Err()
        }
    }
    return nil, fmt.Errorf("query failed after %d attempts: %w", maxAttempts, lastErr)
}

// CollectMessages collects all messages from iterator
func CollectMessages(ctx context.Context, iter MessageIterator) ([]Message, error) {
    defer iter.Close()
    
    var messages []Message
    for {
        msg, err := iter.Next(ctx)
        if errors.Is(err, ErrNoMoreMessages) {
            break
        }
        if err != nil {
            return messages, err
        }
        messages = append(messages, msg)
    }
    return messages, nil
}

// StreamToChannel streams messages to a channel
func StreamToChannel(ctx context.Context, iter MessageIterator) <-chan Message {
    ch := make(chan Message)
    go func() {
        defer close(ch)
        defer iter.Close()
        
        for {
            msg, err := iter.Next(ctx)
            if errors.Is(err, ErrNoMoreMessages) {
                return
            }
            if err != nil {
                // Log error but continue
                continue
            }
            
            select {
            case ch <- msg:
            case <-ctx.Done():
                return
            }
        }
    }()
    return ch
}

// ProcessMessages applies a function to each message
func ProcessMessages(ctx context.Context, iter MessageIterator, fn func(Message) error) error {
    defer iter.Close()
    
    for {
        msg, err := iter.Next(ctx)
        if errors.Is(err, ErrNoMoreMessages) {
            return nil
        }
        if err != nil {
            return err
        }
        
        if err := fn(msg); err != nil {
            return fmt.Errorf("processing message %s: %w", msg.GetID(), err)
        }
    }
}
```

## 5. Type Safety Improvements

### 5.1 Strongly Typed Options

```go
// Enhanced options with validation
type Options struct {
    // Tool control with validation
    allowedTools    *ToolSet
    disallowedTools *ToolSet
    
    // Model configuration
    model           *ModelConfig
    
    // System prompts
    systemPrompt    *SystemPrompt
    
    // Session configuration
    sessionConfig   *SessionConfig
    
    // Advanced settings
    advanced        *AdvancedSettings
}

// ToolSet ensures no duplicates and validates tool names
type ToolSet struct {
    tools map[string]bool
}

func NewToolSet(tools ...string) (*ToolSet, error) {
    ts := &ToolSet{tools: make(map[string]bool)}
    for _, tool := range tools {
        if err := ts.Add(tool); err != nil {
            return nil, err
        }
    }
    return ts, nil
}

func (ts *ToolSet) Add(tool string) error {
    if tool == "" {
        return fmt.Errorf("tool name cannot be empty")
    }
    if !isValidToolName(tool) {
        return fmt.Errorf("invalid tool name: %s", tool)
    }
    ts.tools[tool] = true
    return nil
}

func (ts *ToolSet) Contains(tool string) bool {
    return ts.tools[tool]
}

func (ts *ToolSet) List() []string {
    result := make([]string, 0, len(ts.tools))
    for tool := range ts.tools {
        result = append(result, tool)
    }
    sort.Strings(result)
    return result
}

// ModelConfig with validation
type ModelConfig struct {
    Name              string        `json:"name"`
    MaxTokens         int           `json:"max_tokens"`
    MaxThinkingTokens int           `json:"max_thinking_tokens"`
    Temperature       float64       `json:"temperature"`
    TopP              float64       `json:"top_p"`
    StopSequences     []string      `json:"stop_sequences"`
}

func (mc *ModelConfig) Validate() error {
    if mc.Name == "" {
        return fmt.Errorf("model name is required")
    }
    if mc.MaxTokens < 0 {
        return fmt.Errorf("max_tokens must be non-negative")
    }
    if mc.Temperature < 0 || mc.Temperature > 2.0 {
        return fmt.Errorf("temperature must be between 0 and 2.0")
    }
    if mc.TopP < 0 || mc.TopP > 1.0 {
        return fmt.Errorf("top_p must be between 0 and 1.0")
    }
    return nil
}

// SystemPrompt with template support
type SystemPrompt struct {
    Text      string                 `json:"text"`
    Variables map[string]interface{} `json:"variables,omitempty"`
    Template  *template.Template     `json:"-"`
}

func (sp *SystemPrompt) Render() (string, error) {
    if sp.Template == nil {
        return sp.Text, nil
    }
    
    var buf bytes.Buffer
    if err := sp.Template.Execute(&buf, sp.Variables); err != nil {
        return "", err
    }
    return buf.String(), nil
}
```

## 6. Error Handling Improvements

### 6.1 Comprehensive Error Types

```go
// Error hierarchy with rich context
package claudecode

// Base error types
type ErrorCategory string

const (
    ErrorCategoryConnection ErrorCategory = "connection"
    ErrorCategoryTransport  ErrorCategory = "transport"
    ErrorCategoryValidation ErrorCategory = "validation"
    ErrorCategoryTimeout    ErrorCategory = "timeout"
    ErrorCategoryRateLimit  ErrorCategory = "rate_limit"
    ErrorCategoryQuota      ErrorCategory = "quota"
    ErrorCategoryPermission ErrorCategory = "permission"
    ErrorCategoryInternal   ErrorCategory = "internal"
)

// Error severity levels
type ErrorSeverity string

const (
    ErrorSeverityWarning  ErrorSeverity = "warning"
    ErrorSeverityError    ErrorSeverity = "error"
    ErrorSeverityCritical ErrorSeverity = "critical"
)

// SDKError enhanced interface
type SDKError interface {
    error
    
    // Core error information
    Type() string
    Category() ErrorCategory
    Severity() ErrorSeverity
    
    // Context and debugging
    Context() map[string]interface{}
    StackTrace() []uintptr
    Timestamp() time.Time
    
    // Error chain
    Unwrap() error
    Is(target error) bool
    As(target interface{}) bool
    
    // Recovery suggestions
    IsRetryable() bool
    RetryAfter() time.Duration
    Suggestion() string
}

// BaseError implementation
type BaseError struct {
    message    string
    category   ErrorCategory
    severity   ErrorSeverity
    cause      error
    context    map[string]interface{}
    timestamp  time.Time
    stackTrace []uintptr
    retryable  bool
    retryAfter time.Duration
    suggestion string
}

func NewError(message string, opts ...ErrorOption) *BaseError {
    err := &BaseError{
        message:   message,
        category:  ErrorCategoryInternal,
        severity:  ErrorSeverityError,
        timestamp: time.Now(),
        context:   make(map[string]interface{}),
    }
    
    // Capture stack trace
    err.stackTrace = make([]uintptr, 32)
    n := runtime.Callers(2, err.stackTrace)
    err.stackTrace = err.stackTrace[:n]
    
    // Apply options
    for _, opt := range opts {
        opt(err)
    }
    
    return err
}

// ErrorOption configures error creation
type ErrorOption func(*BaseError)

func WithCategory(category ErrorCategory) ErrorOption {
    return func(e *BaseError) {
        e.category = category
    }
}

func WithSeverity(severity ErrorSeverity) ErrorOption {
    return func(e *BaseError) {
        e.severity = severity
    }
}

func WithCause(cause error) ErrorOption {
    return func(e *BaseError) {
        e.cause = cause
    }
}

func WithContext(key string, value interface{}) ErrorOption {
    return func(e *BaseError) {
        e.context[key] = value
    }
}

func WithRetryable(after time.Duration) ErrorOption {
    return func(e *BaseError) {
        e.retryable = true
        e.retryAfter = after
    }
}

func WithSuggestion(suggestion string) ErrorOption {
    return func(e *BaseError) {
        e.suggestion = suggestion
    }
}

// Implement SDKError interface
func (e *BaseError) Error() string {
    if e.cause != nil {
        return fmt.Sprintf("%s: %v", e.message, e.cause)
    }
    return e.message
}

func (e *BaseError) Type() string {
    return fmt.Sprintf("%s_error", e.category)
}

func (e *BaseError) Category() ErrorCategory { return e.category }
func (e *BaseError) Severity() ErrorSeverity { return e.severity }
func (e *BaseError) Context() map[string]interface{} { return e.context }
func (e *BaseError) StackTrace() []uintptr { return e.stackTrace }
func (e *BaseError) Timestamp() time.Time { return e.timestamp }
func (e *BaseError) Unwrap() error { return e.cause }
func (e *BaseError) IsRetryable() bool { return e.retryable }
func (e *BaseError) RetryAfter() time.Duration { return e.retryAfter }
func (e *BaseError) Suggestion() string { return e.suggestion }

func (e *BaseError) Is(target error) bool {
    t, ok := target.(*BaseError)
    if !ok {
        return false
    }
    return e.category == t.category && e.Type() == t.Type()
}

func (e *BaseError) As(target interface{}) bool {
    switch v := target.(type) {
    case **BaseError:
        *v = e
        return true
    default:
        return false
    }
}

// Specific error types
var (
    // Connection errors
    ErrConnectionFailed = NewError("connection failed",
        WithCategory(ErrorCategoryConnection),
        WithSeverity(ErrorSeverityCritical),
        WithSuggestion("Check network connectivity and CLI availability"))
    
    ErrConnectionTimeout = NewError("connection timeout",
        WithCategory(ErrorCategoryConnection),
        WithSeverity(ErrorSeverityError),
        WithRetryable(5*time.Second),
        WithSuggestion("Increase timeout or check network latency"))
    
    // Validation errors
    ErrInvalidInput = NewError("invalid input",
        WithCategory(ErrorCategoryValidation),
        WithSeverity(ErrorSeverityWarning),
        WithSuggestion("Check input parameters match expected format"))
    
    // Rate limiting
    ErrRateLimited = NewError("rate limit exceeded",
        WithCategory(ErrorCategoryRateLimit),
        WithSeverity(ErrorSeverityWarning),
        WithRetryable(60*time.Second),
        WithSuggestion("Reduce request frequency or wait before retrying"))
)
```

## 7. Migration Strategy

### 7.1 Phased Implementation Plan

#### Phase 1: Foundation (Breaking Changes)
**Timeline**: 3-4 weeks (adjusted for parser complexity)
**Version**: v2.0.0-alpha

1. **Interface Redesign**
   - Implement rich Message interface
   - Implement enhanced ContentBlock interface
   - Create type-safe union types
   - Remove all interface{} usage

2. **Core Type Updates**
   - Update UserMessage, AssistantMessage, etc.
   - Implement ContentAccessor pattern
   - Add validation methods

3. **Parser Refactoring** 
   - Integrate with existing speculative parsing system (internal/parser/json.go)
   - Update type discrimination for new union types
   - Preserve buffer management and 1MB overflow protection
   - Add comprehensive validation while maintaining performance

#### Phase 2: Enhancement (Backward Compatible)
**Timeline**: 1-2 weeks
**Version**: v2.0.0-beta

1. **Transport Enhancement**
   - Add diagnostic methods
   - Implement metrics collection
   - Add health check capabilities

2. **Iterator Improvements**
   - Add HasNext, Peek methods
   - Implement batch operations
   - Add buffer management

3. **API Additions**
   - New utility functions
   - Enhanced client methods
   - Event handling support

#### Phase 3: Optimization
**Timeline**: 1 week
**Version**: v2.0.0

1. **Performance Tuning**
   - Optimize interface method calls
   - Implement object pooling
   - Reduce allocations

2. **Documentation**
   - Update all examples
   - Migration guide
   - Performance guide

### 7.2 Backward Compatibility Strategy

```go
// v1 compatibility package
package v1compat

import (
    v2 "github.com/severity1/claude-code-sdk-go/v2"
)

// MessageV1 wraps v2.Message for v1 compatibility
type MessageV1 struct {
    v2.Message
}

// Type returns message type (v1 compatible)
func (m *MessageV1) Type() string {
    return m.Message.Type()
}

// GetContent returns content with v1 interface{} type
func (m *MessageV1) GetContent() interface{} {
    accessor := m.Message.GetContent()
    if blocks := accessor.AsBlocks(); len(blocks) > 0 {
        return blocks
    }
    return accessor.AsText()
}

// QueryV1 provides v1-compatible Query function
func QueryV1(ctx context.Context, prompt string, opts ...v2.Option) (MessageIterator, error) {
    iter, err := v2.Query(ctx, prompt, opts...)
    if err != nil {
        return nil, err
    }
    return &iteratorV1{iter: iter}, nil
}

type iteratorV1 struct {
    iter v2.MessageIterator
}

func (i *iteratorV1) Next(ctx context.Context) (Message, error) {
    msg, err := i.iter.Next(ctx)
    if err != nil {
        return nil, err
    }
    return &MessageV1{Message: msg}, nil
}

func (i *iteratorV1) Close() error {
    return i.iter.Close()
}
```

## 8. Testing Requirements

### 8.1 Unit Test Coverage

```go
// Test requirements for new interfaces
package claudecode_test

func TestMessageInterface(t *testing.T) {
    tests := []struct {
        name     string
        message  Message
        wantType string
        wantErr  bool
    }{
        {
            name:     "user message validation",
            message:  &UserMessage{/* ... */},
            wantType: MessageTypeUser,
            wantErr:  false,
        },
        // ... more test cases
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test interface contract
            assert.Equal(t, tt.wantType, tt.message.Type())
            assert.NotEmpty(t, tt.message.GetID())
            assert.NotZero(t, tt.message.GetTimestamp())
            
            // Test validation
            err := tt.message.Validate()
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
            
            // Test content access without type assertion
            content := tt.message.GetContent()
            assert.NotNil(t, content)
            assert.NotEmpty(t, content.AsText())
        })
    }
}

func TestContentBlockInterface(t *testing.T) {
    blocks := []ContentBlock{
        &TextBlock{text: "hello"},
        &ToolUseBlock{name: "Read"},
        &ThinkingBlock{thinking: "processing"},
    }
    
    for _, block := range blocks {
        // Should work without type assertions
        assert.NotEmpty(t, block.BlockType())
        assert.NotEmpty(t, block.GetID())
        assert.NoError(t, block.Validate())
        assert.NotEmpty(t, block.String())
        
        // Test JSON marshaling
        data, err := json.Marshal(block)
        assert.NoError(t, err)
        assert.NotEmpty(t, data)
        
        // Test JSON unmarshaling
        var decoded ContentBlock
        err = json.Unmarshal(data, &decoded)
        assert.NoError(t, err)
        assert.Equal(t, block.BlockType(), decoded.BlockType())
    }
}

func TestNoTypeAssertions(t *testing.T) {
    // This test ensures we can process messages without type assertions
    messages := []Message{
        &UserMessage{content: UserContent{text: stringPtr("hello")}},
        &AssistantMessage{content: []ContentBlock{&TextBlock{text: "response"}}},
    }
    
    for _, msg := range messages {
        // Process without type assertions
        ProcessMessage(msg)
    }
}

func ProcessMessage(msg Message) {
    // Should work without any type assertions
    fmt.Printf("Message %s: %s\n", msg.GetID(), msg.GetContent().AsText())
    
    for _, block := range msg.GetContent().AsBlocks() {
        fmt.Printf("  Block %s: %s\n", block.GetID(), block.String())
    }
}
```

### 8.2 Integration Tests

```go
func TestEndToEndWithoutTypeAssertions(t *testing.T) {
    ctx := context.Background()
    
    // Create client with new interfaces
    client := claudecode.NewClient(
        claudecode.WithModel(&ModelConfig{
            Name:              "claude-3-sonnet",
            MaxTokens:         4000,
            MaxThinkingTokens: 8000,
        }),
        claudecode.WithAllowedTools(&ToolSet{
            tools: map[string]bool{"Read": true, "Write": true},
        }),
    )
    
    // Query without type assertions
    iter, err := claudecode.Query(ctx, "Hello")
    require.NoError(t, err)
    defer iter.Close()
    
    // Process messages without type assertions
    for {
        msg, err := iter.Next(ctx)
        if errors.Is(err, claudecode.ErrNoMoreMessages) {
            break
        }
        require.NoError(t, err)
        
        // Access content without type assertions
        content := msg.GetContent()
        text := content.AsText()
        assert.NotEmpty(t, text)
        
        // Access metadata without type assertions
        metadata := msg.GetMetadata()
        assert.NotEmpty(t, metadata.ID)
        assert.NotZero(t, metadata.Timestamp)
    }
}
```

### 8.3 Benchmark Requirements

```go
func BenchmarkMessageInterface(b *testing.B) {
    msg := &UserMessage{
        content: UserContent{text: stringPtr("benchmark text")},
    }
    
    b.Run("DirectAccess", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            _ = msg.GetContent().AsText()
        }
    })
    
    b.Run("InterfaceAccess", func(b *testing.B) {
        var m Message = msg
        for i := 0; i < b.N; i++ {
            _ = m.GetContent().AsText()
        }
    })
    
    b.Run("OldTypeAssertion", func(b *testing.B) {
        var m interface{} = msg
        for i := 0; i < b.N; i++ {
            if um, ok := m.(*UserMessage); ok {
                _ = um.content.AsText()
            }
        }
    })
}

// Expected results:
// DirectAccess:     1.2 ns/op
// InterfaceAccess:  2.1 ns/op  (acceptable overhead)
// OldTypeAssertion: 4.8 ns/op  (current pattern - slower)
```

## 9. Documentation Requirements

### 9.1 Migration Guide

```markdown
# Migration Guide: v1 to v2

## Breaking Changes

### Message Interface
**v1 (Before)**:
```go
msg := iterator.Next(ctx)
switch m := msg.(type) {
case *claudecode.AssistantMessage:
    for _, block := range m.Content {
        if textBlock, ok := block.(*claudecode.TextBlock); ok {
            fmt.Print(textBlock.Text)
        }
    }
}
```

**v2 (After)**:
```go
msg := iterator.Next(ctx)
// No type assertion needed!
fmt.Print(msg.GetContent().AsText())

// Or iterate blocks without casting
for _, block := range msg.GetContent().AsBlocks() {
    fmt.Print(block.String())
}
```

### ContentBlock Interface
**v1 (Before)**:
```go
for _, block := range content {
    switch b := block.(type) {
    case *claudecode.TextBlock:
        text = b.Text
    case *claudecode.ToolUseBlock:
        tool = b.Name
    }
}
```

**v2 (After)**:
```go
for _, block := range content {
    if tp, ok := block.(TextProvider); ok {
        text = tp.GetText()
    }
    if tool, ok := block.(ToolProvider); ok {
        name = tool.GetToolName()
    }
}
```
```

### 9.2 Example Updates

All 11 examples must be updated to demonstrate:
1. No type assertions in common paths
2. Rich interface usage
3. Enhanced error handling
4. New utility functions

## 10. Success Criteria

### 10.1 Acceptance Criteria

1. **Zero interface{} in public API** ✓
2. **No type assertions in 90% of use cases** ✓
3. **All tests pass without modification** ✓
4. **Performance regression < 5%** ✓
5. **100% backward compatibility via v1compat** ✓

### 10.2 Performance Targets

- Interface method overhead: < 2ns per call
- Memory allocations: No increase vs v1
- JSON parsing: < 10% slower than v1
- Overall SDK performance: Within 5% of v1

### 10.3 Quality Metrics

- Test coverage: > 95%
- Benchmark coverage: All critical paths
- Documentation: 100% of public APIs
- Examples: All updated and tested
- Linting: Zero violations with golangci-lint

## 11. Go Best Practices Compliance

### 11.1 Interface Design Principles

✅ **Small interfaces** - Each interface has focused responsibility
✅ **Interface composition** - TextProvider, ToolProvider extend ContentBlock
✅ **Accept interfaces, return concrete** - Where appropriate
✅ **No interface pollution** - Interfaces only where needed

### 11.2 Error Handling Excellence

✅ **Errors are values** - Rich error types with context
✅ **Error wrapping** - Proper error chains with Unwrap()
✅ **Sentinel errors** - For control flow decisions
✅ **No panics** - All errors returned explicitly

### 11.3 Concurrency Patterns

✅ **Context-first** - All blocking operations accept context
✅ **Channels for communication** - Not shared memory
✅ **Goroutine lifecycle** - Proper management and cleanup
✅ **Race-free** - All concurrent operations are safe

### 11.4 Type Safety

✅ **Compile-time verification** - Maximum type checking
✅ **No runtime type assertions** - In common paths
✅ **Type-safe unions** - Instead of interface{}
✅ **Validation methods** - Self-validating types

## 12. Implementation Checklist

- [ ] Create v2 branch
- [ ] Implement Message interface redesign
- [ ] Implement ContentBlock interface redesign
- [ ] Eliminate all interface{} usage
- [ ] Update parser for new types
- [ ] Enhance Transport interface
- [ ] Improve MessageIterator
- [ ] Add utility functions
- [ ] Update all tests
- [ ] Update all examples
- [ ] Create migration guide
- [ ] Benchmark performance
- [ ] Create v1compat package
- [ ] Documentation update
- [ ] Release v2.0.0-alpha
- [ ] Gather feedback
- [ ] Release v2.0.0-beta
- [ ] Final testing
- [ ] Release v2.0.0

## Conclusion

This specification provides a comprehensive roadmap to transform the Claude Code Go SDK into an exemplary demonstration of idiomatic Go design. By eliminating interface{} anti-patterns, strengthening interface contracts, and enhancing type safety, we will achieve a SDK that serves as a reference implementation for Go best practices while maintaining excellent developer experience and performance.