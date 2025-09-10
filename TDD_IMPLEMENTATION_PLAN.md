# TDD Implementation Plan: Claude Code Go SDK v2.0

## Executive Summary

This TDD (Test-Driven Development) plan implements the improvements defined in `IMPROVEMENT_SPEC.md` using a rigorous test-first approach. Since this project is new (9 stars), we can make breaking changes without backwards compatibility concerns, allowing for clean implementation of idiomatic Go patterns.

**Objective**: Transform the SDK from C+ interface design to A+ by eliminating `interface{}` anti-patterns and implementing rich behavioral interfaces.

**Approach**: Bottom-up TDD with dependency-ordered implementation phases.

## Phase 1: Foundation Layer (Week 1-2)

### 1.1 Core Interface Contracts (TDD)

#### Test-First: Rich Message Interface

**File**: `internal/shared/message_test.go` (Complete Rewrite)

```go
// RED: Write failing tests first
package shared_test

import (
    "encoding/json"
    "strings"
    "testing"
    "time"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

// TestMessageInterfaceContract defines the contract all messages must satisfy
func TestMessageInterfaceContract(t *testing.T) {
    tests := []struct {
        name    string
        message Message
        wantType string
    }{
        {
            name: "user_message_contract",
            message: &UserMessage{
                metadata: MessageMetadata{
                    ID:        "msg-001",
                    SessionID: "session-123",
                    Timestamp: time.Now(),
                },
                content: UserContent{text: stringPtr("Hello Claude")},
            },
            wantType: MessageTypeUser,
        },
        {
            name: "assistant_message_contract", 
            message: &AssistantMessage{
                metadata: MessageMetadata{
                    ID:        "msg-002",
                    SessionID: "session-123", 
                    Timestamp: time.Now(),
                    Model:     "claude-3-sonnet",
                },
                content: []ContentBlock{
                    &TextBlock{id: "block-001", text: "Hello! How can I help?"},
                },
            },
            wantType: MessageTypeAssistant,
        },
        // ... more message types
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test core interface contract - NO TYPE ASSERTIONS
            assert.Equal(t, tt.wantType, tt.message.Type())
            assert.NotEmpty(t, tt.message.GetID())
            assert.NotZero(t, tt.message.GetTimestamp())
            assert.NotEmpty(t, tt.message.GetSessionID())
            
            // Test content access without type assertions
            content := tt.message.GetContent()
            require.NotNil(t, content)
            assert.GreaterOrEqual(t, content.GetBlockCount(), 0)
            
            // Test metadata access
            metadata := tt.message.GetMetadata()
            assert.NotEmpty(t, metadata.ID)
            assert.NotZero(t, metadata.Timestamp)
            
            // Test validation
            assert.NoError(t, tt.message.Validate())
            
            // Test string representation
            assert.NotEmpty(t, tt.message.String())
            assert.Contains(t, tt.message.String(), tt.message.GetID())
            
            // Test JSON marshaling
            data, err := json.Marshal(tt.message)
            assert.NoError(t, err)
            assert.NotEmpty(t, data)
            
            // Verify type field in JSON
            var jsonMap map[string]interface{}
            err = json.Unmarshal(data, &jsonMap)
            assert.NoError(t, err)
            assert.Equal(t, tt.wantType, jsonMap["type"])
        })
    }
}

// TestContentAccessorContract ensures uniform content access
func TestContentAccessorContract(t *testing.T) {
    tests := []struct {
        name     string
        message  Message
        wantText string
        wantBlocks int
    }{
        {
            name: "text_content",
            message: &UserMessage{
                content: UserContent{text: stringPtr("Simple text message")},
            },
            wantText: "Simple text message",
            wantBlocks: 1, // Should convert text to single TextBlock
        },
        {
            name: "block_content", 
            message: &AssistantMessage{
                content: []ContentBlock{
                    &TextBlock{text: "First block"},
                    &TextBlock{text: "Second block"},
                },
            },
            wantText: "First block\nSecond block",
            wantBlocks: 2,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            content := tt.message.GetContent()
            
            // Test text representation
            assert.Equal(t, tt.wantText, content.AsText())
            
            // Test block access
            blocks := content.AsBlocks()
            assert.Len(t, blocks, tt.wantBlocks)
            assert.Equal(t, tt.wantBlocks, content.GetBlockCount())
            
            // Test individual block access
            for i := 0; i < content.GetBlockCount(); i++ {
                block, err := content.GetBlock(i)
                assert.NoError(t, err)
                assert.NotNil(t, block)
                assert.NotEmpty(t, block.BlockType())
            }
            
            // Test size calculation
            assert.Greater(t, content.GetSize(), 0)
            
            // Test empty check
            assert.False(t, content.IsEmpty())
        })
    }
}

// TestNoTypeAssertionsRequired - Critical test ensuring interface richness
func TestNoTypeAssertionsRequired(t *testing.T) {
    messages := []Message{
        &UserMessage{
            metadata: MessageMetadata{ID: "user-001", SessionID: "session-1", Timestamp: time.Now()},
            content: UserContent{text: stringPtr("Hello")},
        },
        &AssistantMessage{
            metadata: MessageMetadata{ID: "asst-001", SessionID: "session-1", Timestamp: time.Now()},
            content: []ContentBlock{&TextBlock{id: "block-1", text: "Hi there!"}},
        },
    }
    
    // Process messages WITHOUT any type assertions
    for _, msg := range messages {
        processMessageWithoutTypeAssertions(t, msg)
    }
}

func processMessageWithoutTypeAssertions(t *testing.T, msg Message) {
    t.Helper()
    
    // Should work without any type assertions
    id := msg.GetID()
    sessionID := msg.GetSessionID()
    content := msg.GetContent()
    text := content.AsText()
    blocks := content.AsBlocks()
    
    assert.NotEmpty(t, id)
    assert.NotEmpty(t, sessionID) 
    assert.NotEmpty(t, text)
    assert.NotEmpty(t, blocks)
    
    // Process blocks without type assertions
    for _, block := range blocks {
        assert.NotEmpty(t, block.BlockType())
        assert.NotEmpty(t, block.GetID())
        assert.NoError(t, block.Validate())
    }
}
```

**GREEN: Implement minimum code to pass tests**

**File**: `internal/shared/message.go` (Complete Rewrite)

```go
package shared

import (
    "encoding/json"
    "fmt"
    "strings"
    "time"
)

// Message represents any message with rich behavioral interface
type Message interface {
    // Core identification
    Type() string
    GetID() string
    GetTimestamp() time.Time
    GetSessionID() string
    
    // Content access without type assertions
    GetContent() ContentAccessor
    GetMetadata() MessageMetadata
    
    // Behavioral methods
    Validate() error
    IsEmpty() bool
    
    // Serialization
    json.Marshaler
    fmt.Stringer
}

// ContentAccessor provides uniform access to message content
type ContentAccessor interface {
    AsText() string
    AsBlocks() []ContentBlock
    GetBlockCount() int
    GetBlock(index int) (ContentBlock, error)
    FilterByType(blockType string) []ContentBlock
    IsEmpty() bool
    GetSize() int
    ContainsType(blockType string) bool
}

// MessageMetadata contains common metadata for all messages
type MessageMetadata struct {
    ID        string    `json:"id"`
    SessionID string    `json:"session_id"`
    RequestID string    `json:"request_id,omitempty"`
    Timestamp time.Time `json:"timestamp"`
    Version   string    `json:"version,omitempty"`
    Model     string    `json:"model,omitempty"`
}

// UserMessage with rich interface implementation
type UserMessage struct {
    metadata MessageMetadata
    content  UserContent
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

func (m *UserMessage) GetContent() ContentAccessor {
    return &userContentAccessor{content: m.content}
}

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

func (m *UserMessage) MarshalJSON() ([]byte, error) {
    return json.Marshal(map[string]interface{}{
        "type":       m.Type(),
        "id":         m.metadata.ID,
        "session_id": m.metadata.SessionID,
        "timestamp":  m.metadata.Timestamp,
        "content":    m.content,
    })
}

// userContentAccessor implements ContentAccessor
type userContentAccessor struct {
    content UserContent
}

func (a *userContentAccessor) AsText() string {
    return a.content.AsText()
}

func (a *userContentAccessor) AsBlocks() []ContentBlock {
    if a.content.blocks != nil {
        return a.content.blocks
    }
    if a.content.text != nil {
        return []ContentBlock{&TextBlock{text: *a.content.text}}
    }
    return nil
}

func (a *userContentAccessor) GetBlockCount() int {
    return len(a.AsBlocks())
}

func (a *userContentAccessor) GetBlock(index int) (ContentBlock, error) {
    blocks := a.AsBlocks()
    if index < 0 || index >= len(blocks) {
        return nil, fmt.Errorf("block index %d out of range", index)
    }
    return blocks[index], nil
}

func (a *userContentAccessor) FilterByType(blockType string) []ContentBlock {
    var result []ContentBlock
    for _, block := range a.AsBlocks() {
        if block.BlockType() == blockType {
            result = append(result, block)
        }
    }
    return result
}

func (a *userContentAccessor) IsEmpty() bool {
    return a.content.IsEmpty()
}

func (a *userContentAccessor) GetSize() int {
    return len(a.AsText())
}

func (a *userContentAccessor) ContainsType(blockType string) bool {
    return len(a.FilterByType(blockType)) > 0
}

// UserContent methods
func (c *UserContent) IsText() bool { return c.text != nil }
func (c *UserContent) IsBlocks() bool { return c.blocks != nil }
func (c *UserContent) IsEmpty() bool { 
    return c.text == nil && len(c.blocks) == 0 
}

func (c *UserContent) AsText() string {
    if c.text != nil {
        return *c.text
    }
    if c.blocks != nil {
        var texts []string
        for _, block := range c.blocks {
            if tp, ok := block.(TextProvider); ok {
                texts = append(texts, tp.GetText())
            }
        }
        return strings.Join(texts, "\n")
    }
    return ""
}

// Helper function
func stringPtr(s string) *string { return &s }
```

#### Test-First: Rich ContentBlock Interface

**File**: `internal/shared/content_test.go` (New File)

```go
// RED: Write failing tests for enhanced ContentBlock interface
package shared_test

func TestContentBlockInterfaceContract(t *testing.T) {
    blocks := []ContentBlock{
        &TextBlock{id: "text-1", text: "Hello world"},
        &ToolUseBlock{id: "tool-1", name: "Read", toolID: "read-123"},
        &ThinkingBlock{id: "think-1", thinking: "Let me process this..."},
    }
    
    for _, block := range blocks {
        t.Run(block.BlockType(), func(t *testing.T) {
            // Test basic interface contract
            assert.NotEmpty(t, block.BlockType())
            assert.NotEmpty(t, block.GetID())
            assert.False(t, block.IsEmpty())
            assert.Greater(t, block.GetSize(), 0)
            assert.NoError(t, block.Validate())
            assert.NotEmpty(t, block.String())
            
            // Test JSON marshaling
            data, err := json.Marshal(block)
            assert.NoError(t, err)
            assert.NotEmpty(t, data)
        })
    }
}

func TestTextProviderInterface(t *testing.T) {
    block := &TextBlock{id: "test-1", text: "Original text"}
    
    // Test TextProvider interface
    tp, ok := block.(TextProvider)
    require.True(t, ok, "TextBlock should implement TextProvider")
    
    assert.Equal(t, "Original text", tp.GetText())
    assert.Equal(t, 13, tp.GetLength())
    assert.Equal(t, 2, tp.GetWordCount())
    assert.True(t, tp.Contains("Original"))
    assert.False(t, tp.Contains("Missing"))
    
    // Test SetText
    err := tp.SetText("Updated text")
    assert.NoError(t, err)
    assert.Equal(t, "Updated text", tp.GetText())
}

func TestToolProviderInterface(t *testing.T) {
    block := &ToolUseBlock{
        id:     "tool-1", 
        name:   "Read",
        toolID: "read-123",
        input:  ToolInput{Parameters: map[string]interface{}{"file": "test.txt"}},
    }
    
    // Test ToolProvider interface
    tp, ok := block.(ToolProvider)
    require.True(t, ok, "ToolUseBlock should implement ToolProvider")
    
    assert.Equal(t, "Read", tp.GetToolName())
    assert.Equal(t, "read-123", tp.GetToolID())
    assert.NotNil(t, tp.GetInput())
    assert.False(t, tp.IsError())
}
```

**GREEN: Implement enhanced ContentBlock interface**

**File**: `internal/shared/content.go` (New File)

```go
package shared

import (
    "encoding/json"
    "fmt"
    "strings"
)

// ContentBlock base interface for all content blocks
type ContentBlock interface {
    BlockType() string
    GetID() string
    IsEmpty() bool
    GetSize() int
    Validate() error
    json.Marshaler
    fmt.Stringer
}

// TextProvider for blocks containing text
type TextProvider interface {
    ContentBlock
    GetText() string
    SetText(text string) error
    GetLength() int
    GetWordCount() int
    Contains(substring string) bool
}

// ToolProvider for tool-related blocks
type ToolProvider interface {
    ContentBlock
    GetToolName() string
    GetToolID() string
    GetInput() ToolInput
    GetOutput() ToolOutput
    IsError() bool
    GetError() error
}

// ThinkingProvider for thinking blocks
type ThinkingProvider interface {
    ContentBlock
    GetThinking() string
    GetSignature() string
    GetConfidence() float64
    GetDuration() time.Duration
}

// TextBlock implementation
type TextBlock struct {
    id   string
    text string
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

func (b *TextBlock) MarshalJSON() ([]byte, error) {
    return json.Marshal(map[string]interface{}{
        "type": b.BlockType(),
        "id":   b.id,
        "text": b.text,
    })
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
func (b *TextBlock) GetWordCount() int { return len(strings.Fields(b.text)) }
func (b *TextBlock) Contains(substring string) bool { 
    return strings.Contains(b.text, substring) 
}

// Type-safe tool types
type ToolInput struct {
    Parameters map[string]interface{} `json:"parameters"`
}

type ToolOutput struct {
    Result interface{} `json:"result"`
    Error  *ToolError  `json:"error,omitempty"`
}

type ToolError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
}
```

### 1.2 Stream Message Enhancement (TDD)

#### Test-First: Type-Safe StreamMessage

**File**: `internal/shared/stream_test.go` (Complete Rewrite)

```go
// RED: Write failing tests for type-safe StreamMessage
package shared_test

func TestStreamMessageTypeSafety(t *testing.T) {
    tests := []struct {
        name    string
        message StreamMessage
    }{
        {
            name: "user_message_payload",
            message: StreamMessage{
                Type:      "request",
                Message:   MessagePayload{UserMessage: &UserMessage{/* ... */}},
                SessionID: "session-123",
            },
        },
        {
            name: "assistant_message_payload", 
            message: StreamMessage{
                Type:      "response",
                Message:   MessagePayload{AssistantMessage: &AssistantMessage{/* ... */}},
                SessionID: "session-123", 
            },
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test type-safe access - NO interface{} casts
            payload := tt.message.Message
            msg := payload.GetMessage()
            assert.NotNil(t, msg)
            assert.NotEmpty(t, msg.Type())
            
            // Test JSON marshaling
            data, err := json.Marshal(tt.message)
            assert.NoError(t, err)
            assert.NotEmpty(t, data)
            
            // Test JSON unmarshaling
            var decoded StreamMessage
            err = json.Unmarshal(data, &decoded)
            assert.NoError(t, err)
            assert.Equal(t, tt.message.Type, decoded.Type)
        })
    }
}
```

#### Test-First: Enhanced MessageIterator

**File**: `internal/shared/iterator_test.go` (New File)

```go
// RED: Write failing tests for enhanced MessageIterator
package shared_test

func TestMessageIteratorEnhancement(t *testing.T) {
    ctx := context.Background()
    messages := []Message{
        &UserMessage{/* ... */},
        &AssistantMessage{/* ... */},
        &ResultMessage{/* ... */},
    }
    
    iterator := &testIterator{messages: messages}
    
    // Test enhanced iterator methods
    assert.True(t, iterator.HasNext(ctx))
    assert.Equal(t, int64(0), iterator.Position())
    assert.False(t, iterator.IsExhausted())
    
    // Test Peek (non-consuming)
    msg, err := iterator.Peek(ctx)
    assert.NoError(t, err)
    assert.Equal(t, messages[0], msg)
    assert.Equal(t, int64(0), iterator.Position()) // Position unchanged
    
    // Test Next (consuming)
    msg, err = iterator.Next(ctx)
    assert.NoError(t, err) 
    assert.Equal(t, messages[0], msg)
    assert.Equal(t, int64(1), iterator.Position()) // Position advanced
    
    // Test batch operations
    batch, err := iterator.NextBatch(ctx, 2)
    assert.NoError(t, err)
    assert.Len(t, batch, 2)
    assert.True(t, iterator.IsExhausted())
}
```

**GREEN: Implement enhanced interfaces**

**File**: `internal/shared/stream.go` (Complete Rewrite)

```go
package shared

import "context"

// MessageIterator with enhanced functionality
type MessageIterator interface {
    // Core iteration
    Next(ctx context.Context) (Message, error)
    
    // Non-blocking operations
    HasNext(ctx context.Context) bool
    Peek(ctx context.Context) (Message, error)
    
    // Batch operations
    NextBatch(ctx context.Context, maxSize int) ([]Message, error)
    Remaining() int
    
    // Navigation
    Skip(n int) error
    Reset() error
    
    // State inspection
    Position() int64
    IsExhausted() bool
    LastError() error
    
    // Resource management
    Close() error
}

// StreamMessage with type-safe fields
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

// ... JSON marshaling implementations
```

**REFACTOR: Clean up and optimize**

Once tests pass, refactor for clarity and performance while keeping tests green.

## Phase 2: Parser Integration (Week 2-3)

### 2.1 Update Parser for New Types (TDD)

#### Test-First: Parser Type Discrimination

**File**: `internal/parser/json_test.go` (Major Refactor)

```go
// RED: Update tests for new type-safe parsing
func TestParserWithNewTypes(t *testing.T) {
    parser := New()
    
    tests := []struct {
        name     string
        jsonLine string
        want     Message
        wantErr  bool
    }{
        {
            name:     "user_message_parsing",
            jsonLine: `{"type":"user","id":"msg-1","content":"Hello"}`,
            want: &UserMessage{
                metadata: MessageMetadata{ID: "msg-1"},
                content:  UserContent{text: stringPtr("Hello")},
            },
            wantErr: false,
        },
        {
            name:     "assistant_message_parsing", 
            jsonLine: `{"type":"assistant","id":"msg-2","content":[{"type":"text","text":"Hi!"}]}`,
            want: &AssistantMessage{
                metadata: MessageMetadata{ID: "msg-2"},
                content:  []ContentBlock{&TextBlock{text: "Hi!"}},
            },
            wantErr: false,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            messages, err := parser.ProcessLine(tt.jsonLine)
            if tt.wantErr {
                assert.Error(t, err)
                return
            }
            
            assert.NoError(t, err)
            assert.Len(t, messages, 1)
            
            msg := messages[0]
            // Test interface contract - NO type assertions in tests
            assert.Equal(t, tt.want.Type(), msg.Type())
            assert.Equal(t, tt.want.GetContent().AsText(), msg.GetContent().AsText())
        })
    }
}
```

**GREEN: Update parser implementation**

**File**: `internal/parser/json.go` (Major Refactor)

```go
// Update parser to work with new type-safe Message interfaces
func (p *Parser) ParseMessage(data map[string]any) (shared.Message, error) {
    msgType, ok := data["type"].(string)
    if !ok {
        return nil, shared.NewMessageParseError("missing or invalid type field", data)
    }

    switch msgType {
    case shared.MessageTypeUser:
        return p.parseUserMessageV2(data)
    case shared.MessageTypeAssistant:
        return p.parseAssistantMessageV2(data)
    // ... etc
    }
}

func (p *Parser) parseUserMessageV2(data map[string]any) (shared.Message, error) {
    // Parse into new UserMessage struct
    var msg shared.UserMessage
    
    // Extract metadata
    if id, ok := data["id"].(string); ok {
        msg.metadata.ID = id
    }
    
    // Parse content with type safety
    if contentData, ok := data["content"]; ok {
        content, err := p.parseUserContent(contentData)
        if err != nil {
            return nil, err
        }
        msg.content = content
    }
    
    return &msg, nil
}
```

## Phase 3: Main Package Integration (Week 3-4)

### 3.1 Update Client and Query APIs (TDD)

#### Test-First: Client with New Interfaces

**File**: `client_test.go` (Major Refactor)

```go
// RED: Update client tests for new interfaces
func TestClientWithNewInterfaces(t *testing.T) {
    ctx := context.Background()
    transport := newMockTransport()
    client := NewClientWithTransport(transport)
    
    err := client.Connect(ctx)
    require.NoError(t, err)
    defer client.Disconnect()
    
    err = client.Query(ctx, "Test message")
    require.NoError(t, err)
    
    // Test receiving messages with rich interfaces - NO type assertions
    msgChan := client.ReceiveMessages(ctx)
    select {
    case msg := <-msgChan:
        if msg != nil {
            // Process without type assertions
            assert.NotEmpty(t, msg.GetID())
            assert.NotEmpty(t, msg.GetContent().AsText())
            
            // Test content blocks without type assertions
            for _, block := range msg.GetContent().AsBlocks() {
                assert.NotEmpty(t, block.BlockType())
                assert.NotEmpty(t, block.String())
            }
        }
    case <-time.After(5 * time.Second):
        t.Fatal("timeout waiting for message")
    }
}
```

**GREEN: Update client implementation**

**File**: `client.go` (Refactor - update interface usage)
**File**: `query.go` (Refactor - update interface usage)
**File**: `types.go` (Major Refactor - convert aliases to rich interfaces)

### 3.2 Enhanced Transport Interface (TDD)

#### Test-First: Transport Diagnostics

**File**: `transport_diagnostics_test.go` (New File)

```go
// RED: Write tests for enhanced Transport interface
func TestTransportDiagnostics(t *testing.T) {
    transport := newEnhancedTransport()
    
    // Test diagnostic methods
    assert.False(t, transport.IsConnected())
    
    err := transport.Connect(ctx)
    require.NoError(t, err)
    
    assert.True(t, transport.IsConnected())
    
    info := transport.GetConnectionInfo()
    assert.NotEmpty(t, info.ID)
    assert.NotZero(t, info.ConnectedAt)
    
    metrics := transport.GetMetrics()
    assert.GreaterOrEqual(t, metrics.MessagesSent, int64(0))
    
    err = transport.HealthCheck(ctx)
    assert.NoError(t, err)
}
```

## Phase 4: Examples and Documentation (Week 4)

### 4.1 Rewrite All Examples (TDD)

#### Test-First: Example Validation

**File**: `examples_test.go` (New File)

```go
// RED: Write tests that validate examples work without type assertions
func TestExamplesNoTypeAssertions(t *testing.T) {
    examples := []string{
        "examples/01_quickstart",
        "examples/02_client_streaming", 
        // ... all examples
    }
    
    for _, example := range examples {
        t.Run(example, func(t *testing.T) {
            // Compile and run example
            cmd := exec.Command("go", "run", example+"/main.go")
            output, err := cmd.CombinedOutput()
            
            // Examples should compile and run successfully
            assert.NoError(t, err, "Example should compile and run: %s", string(output))
            
            // Verify no type assertions in example code
            code, err := os.ReadFile(example + "/main.go")
            assert.NoError(t, err)
            
            // Check that example doesn't use type assertions
            assert.NotContains(t, string(code), ".(type)")
            assert.NotContains(t, string(code), ".(*claudecode.")
        })
    }
}
```

**GREEN: Rewrite all examples**

**File**: `examples/01_quickstart/main.go` (Complete Rewrite)

```go
// Updated to use rich interfaces - NO type assertions
package main

import (
    "context"
    "errors"
    "fmt"
    "log"
    
    "github.com/severity1/claude-code-sdk-go"
)

func main() {
    ctx := context.Background()
    
    iterator, err := claudecode.Query(ctx, "What is 2+2?")
    if err != nil {
        log.Fatalf("Query failed: %v", err)
    }
    defer iterator.Close()
    
    for {
        message, err := iterator.Next(ctx)
        if errors.Is(err, claudecode.ErrNoMoreMessages) {
            break
        }
        if err != nil {
            log.Fatalf("Failed to get message: %v", err)
        }
        
        // NO TYPE ASSERTIONS - use rich interface methods
        fmt.Printf("Message %s: %s\n", message.GetID(), message.GetContent().AsText())
        
        // Process blocks without type assertions
        for _, block := range message.GetContent().AsBlocks() {
            fmt.Printf("  Block %s [%s]: %s\n", 
                block.GetID(), 
                block.BlockType(), 
                block.String())
        }
    }
}
```

## Phase 5: Performance and Polish (Week 5)

### 5.1 Performance Testing (TDD)

#### Test-First: Performance Benchmarks

**File**: `performance_test.go` (New File)

```go
// RED: Write benchmark tests
func BenchmarkMessageInterfacePerformance(b *testing.B) {
    msg := &UserMessage{
        metadata: MessageMetadata{ID: "bench-1", SessionID: "session-1"},
        content:  UserContent{text: stringPtr("benchmark text")},
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
    
    // Performance target: < 2ns per interface call
    // Should be faster than type assertions
}

func BenchmarkParsingPerformance(b *testing.B) {
    parser := New()
    jsonLine := `{"type":"user","id":"msg-1","content":"Hello world"}`
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := parser.ProcessLine(jsonLine)
        if err != nil {
            b.Fatal(err)
        }
    }
    
    // Performance target: < 10% slower than v1
}
```

**GREEN: Optimize implementation**

Optimize interface implementations to meet performance targets.

## Implementation Strategy

### TDD Cycle for Each Component

1. **RED**: Write failing tests that define the interface contract
2. **GREEN**: Implement minimum code to make tests pass
3. **REFACTOR**: Clean up code while keeping tests green
4. **REPEAT**: Add more test cases and functionality

### Dependency Order (Bottom-Up)

1. **Foundation** (`internal/shared/`) - Core interfaces and types
2. **Parser** (`internal/parser/`) - JSON parsing with new types
3. **Transport** (`internal/subprocess/`) - Minimal changes needed
4. **Main Package** - Client, Query, Types integration
5. **Examples** - Demonstrate new patterns
6. **Performance** - Optimize and benchmark

### Key TDD Principles

#### Test Interface Contracts, Not Implementation

```go
// ✅ Good - test interface behavior
func TestMessageContract(t *testing.T) {
    var msg Message = &UserMessage{...}
    assert.NotEmpty(t, msg.GetID())
    assert.NotEmpty(t, msg.GetContent().AsText())
}

// ❌ Bad - test implementation details
func TestUserMessageFields(t *testing.T) {
    msg := &UserMessage{...}
    assert.NotEmpty(t, msg.metadata.ID) // Internal field access
}
```

#### Write Tests That Fail For The Right Reasons

```go
// ✅ Good - test will fail if interface{} is used
func TestNoTypeAssertionsRequired(t *testing.T) {
    messages := []Message{...}
    for _, msg := range messages {
        // This should work without any type assertions
        processMessage(msg) // Would fail if interface{} is used
    }
}
```

#### Progressive Enhancement

Start with basic interface contracts, then add richer functionality:

1. Basic Message interface with Type(), GetID()
2. Add GetContent() with ContentAccessor
3. Add GetMetadata(), Validate(), String() 
4. Add JSON marshaling/unmarshaling
5. Add performance optimizations

### File Modification Strategy

#### Complete Rewrites (No Backwards Compatibility)

- `internal/shared/message.go` - New rich interfaces
- `internal/shared/stream.go` - Enhanced iterator, type-safe StreamMessage
- All example files - New patterns
- All interface-related tests

#### Major Refactors

- `internal/parser/json.go` - Update type discrimination
- `client.go` - Use new interfaces
- `query.go` - Use new MessageIterator
- `types.go` - Convert aliases to rich interfaces

#### Minor Updates

- `internal/subprocess/transport.go` - Add diagnostic methods
- `options.go` - Enhanced validation
- Error handling - Enhanced error context

### Success Criteria

#### Functional Requirements

- ✅ Zero `interface{}` in public API
- ✅ No type assertions in 90% of use cases  
- ✅ All examples work without type assertions
- ✅ Rich behavioral interfaces
- ✅ Type-safe union types

#### Performance Requirements

- ✅ Interface method overhead < 2ns per call
- ✅ JSON parsing performance within 10% of v1
- ✅ Memory usage not increased vs v1
- ✅ All benchmarks pass performance targets

#### Quality Requirements

- ✅ Test coverage > 95%
- ✅ All tests pass consistently
- ✅ Zero linting violations
- ✅ All examples compile and run
- ✅ Documentation updated

## Timeline Summary

**Week 1-2**: Foundation Layer (Message/ContentBlock interfaces)
**Week 2-3**: Parser Integration (Type discrimination)  
**Week 3-4**: Main Package Integration (Client/Query updates)
**Week 4**: Examples and Documentation
**Week 5**: Performance and Polish

**Total**: 5 weeks for complete transformation to idiomatic Go

This TDD approach ensures that every change is driven by tests, maintains quality throughout the process, and results in a robust, idiomatic Go SDK that serves as a reference implementation for Go best practices.