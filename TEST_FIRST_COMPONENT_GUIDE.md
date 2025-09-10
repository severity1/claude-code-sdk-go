# Test-First Approach Guide for Each Component

## Overview

This guide defines the specific test-first transformation approach for each component to eliminate `interface{}` anti-patterns and implement rich behavioral interfaces. Each component follows the RED-GREEN-REFACTOR cycle with specific goals.

## Current Problem Analysis

**Issue**: Current tests validate via type assertions and weak interfaces
```go
// ❌ Current anti-pattern in message_test.go:262
func validateUserMessage(t *testing.T, msg Message) {
    userMsg, ok := msg.(*UserMessage)  // Type assertion required
    if !ok {
        t.Fatalf("Expected *UserMessage, got %T", msg)
    }
    if userMsg.Content == nil {  // interface{} field access
        t.Error("Expected non-nil Content field")
    }
}
```

**Goal**: Tests that work purely through rich interfaces
```go
// ✅ Target pattern - no type assertions needed
func validateMessage(t *testing.T, msg Message) {
    assert.NotEmpty(t, msg.GetID())
    assert.NotEmpty(t, msg.GetContent().AsText())
    assert.NoError(t, msg.Validate())
}
```

## Component 1: Message Interface (`internal/shared/message.go`)

### Current State Analysis
- **File**: `internal/shared/message.go:36` - `Content interface{}`
- **Problem**: Forces type assertions in all usage
- **Test Issues**: Heavy type assertion usage in validation functions

### RED Phase: Write Failing Interface Contract Tests

**Objective**: Define rich Message interface that eliminates type assertions

```go
// File: internal/shared/message_test.go (Complete Rewrite)
func TestRichMessageInterface(t *testing.T) {
    // This test WILL FAIL initially - that's the point
    messages := []Message{
        createUserMessage("Hello"), 
        createAssistantMessage("Hi there!"),
        createSystemMessage("confirmation"),
    }
    
    for _, msg := range messages {
        // These methods don't exist yet - test will fail
        assert.NotEmpty(t, msg.GetID())
        assert.NotEmpty(t, msg.GetSessionID()) 
        assert.NotZero(t, msg.GetTimestamp())
        
        // Content access without type assertions
        content := msg.GetContent()
        assert.NotNil(t, content)
        assert.NotEmpty(t, content.AsText())
        assert.Greater(t, content.GetBlockCount(), 0)
        
        // Behavioral methods don't exist yet
        assert.NoError(t, msg.Validate())
        assert.NotEmpty(t, msg.String())
        
        // JSON serialization
        data, err := json.Marshal(msg)
        assert.NoError(t, err)
        assert.Contains(t, string(data), msg.Type())
    }
}

func TestNoTypeAssertionsRequired(t *testing.T) {
    // Critical test: ensure interface is rich enough
    msg := createUserMessage("Test message")
    
    // Process message WITHOUT any type assertions
    processedText := processMessagePurely(msg)
    assert.Contains(t, processedText, "Test message")
}

func processMessagePurely(msg Message) string {
    // This function CANNOT use type assertions
    // Must work purely through interface methods
    content := msg.GetContent()
    return fmt.Sprintf("[%s:%s] %s", 
        msg.Type(), 
        msg.GetID(), 
        content.AsText())
}
```

### GREEN Phase: Implement Minimal Message Interface

**Objective**: Make tests pass with minimum code

```go
// File: internal/shared/message.go (Targeted changes)
type Message interface {
    Type() string
    GetID() string
    GetSessionID() string
    GetTimestamp() time.Time
    GetContent() ContentAccessor  // New interface
    GetMetadata() MessageMetadata // New type
    Validate() error             // New method
    json.Marshaler
    fmt.Stringer
}

type ContentAccessor interface {
    AsText() string
    AsBlocks() []ContentBlock
    GetBlockCount() int
    GetBlock(index int) (ContentBlock, error)
    IsEmpty() bool
    GetSize() int
}

// Update UserMessage to implement rich interface
type UserMessage struct {
    metadata MessageMetadata
    content  UserContent  // Replace interface{} with type-safe union
}

type UserContent struct {
    text   *string
    blocks []ContentBlock
}
```

### REFACTOR Phase: Optimize and Clean

**Objective**: Clean up implementation while keeping tests green

- Add validation logic
- Optimize string operations
- Add comprehensive error handling
- Implement efficient JSON marshaling

### Test Evolution Strategy

1. **Week 1**: Core interface contract tests (GetID, GetContent)
2. **Week 1**: Content accessor interface tests  
3. **Week 2**: Validation and behavioral method tests
4. **Week 2**: JSON serialization round-trip tests
5. **Week 2**: Performance benchmark tests

## Component 2: ContentBlock Interface (`internal/shared/content.go`)

### Current State Analysis
- **File**: `internal/shared/message.go:123` - `Content interface{}`
- **Problem**: ToolResultBlock forces type assertions
- **Test Issues**: No rich behavioral testing

### RED Phase: Enhanced ContentBlock Contract

```go
// File: internal/shared/content_test.go (New File)
func TestRichContentBlockInterface(t *testing.T) {
    blocks := []ContentBlock{
        createTextBlock("Hello world"),
        createToolUseBlock("Read", "file.txt"),
        createThinkingBlock("Let me think..."),
    }
    
    for _, block := range blocks {
        // Rich interface methods (don't exist yet)
        assert.NotEmpty(t, block.GetID())
        assert.NotEmpty(t, block.BlockType())
        assert.Greater(t, block.GetSize(), 0)
        assert.False(t, block.IsEmpty())
        assert.NoError(t, block.Validate())
        
        // Specialized interface tests
        testSpecializedInterface(t, block)
    }
}

func testSpecializedInterface(t *testing.T, block ContentBlock) {
    // Test specialized interfaces without type assertions
    if tp, ok := block.(TextProvider); ok {
        assert.NotEmpty(t, tp.GetText())
        assert.Greater(t, tp.GetLength(), 0)
    }
    
    if tlp, ok := block.(ToolProvider); ok {
        assert.NotEmpty(t, tlp.GetToolName())
        assert.NotNil(t, tlp.GetInput())
    }
}
```

### GREEN Phase: Implement Rich ContentBlock

```go
// File: internal/shared/content.go (New implementation)
type ContentBlock interface {
    BlockType() string
    GetID() string
    GetSize() int
    IsEmpty() bool
    Validate() error
    json.Marshaler
    fmt.Stringer
}

type TextProvider interface {
    ContentBlock
    GetText() string
    SetText(string) error
    GetLength() int
    Contains(substring string) bool
}

type ToolProvider interface {
    ContentBlock  
    GetToolName() string
    GetToolID() string
    GetInput() ToolInput
    IsError() bool
}
```

## Component 3: Parser Integration (`internal/parser/json.go`)

### Current State Analysis
- **File**: `internal/parser/json.go:45` - Creates messages with `interface{}`
- **Problem**: Parser creates weak message types
- **Test Issues**: No validation of rich interface creation

### RED Phase: Parser Integration Tests

```go
// File: internal/parser/json_test.go (Major Refactor)
func TestParserCreatesRichInterfaces(t *testing.T) {
    parser := New()
    
    tests := []struct {
        name     string
        jsonLine string
        validate func(*testing.T, Message)
    }{
        {
            name:     "user_message_rich_interface",
            jsonLine: `{"type":"user","id":"msg-1","content":"Hello"}`,
            validate: func(t *testing.T, msg Message) {
                // Test that parser creates messages with rich interfaces
                assert.Equal(t, "user", msg.Type())
                assert.Equal(t, "msg-1", msg.GetID())
                assert.Equal(t, "Hello", msg.GetContent().AsText())
                assert.Equal(t, 1, msg.GetContent().GetBlockCount())
                
                // Should work without type assertions
                assert.NoError(t, msg.Validate())
            },
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            messages, err := parser.ProcessLine(tt.jsonLine)
            require.NoError(t, err)
            require.Len(t, messages, 1)
            
            tt.validate(t, messages[0])
        })
    }
}
```

### GREEN Phase: Update Parser Implementation

**Objective**: Parser creates messages with rich interfaces

```go
// File: internal/parser/json.go (Targeted updates)
func (p *Parser) parseUserMessage(data map[string]any) (shared.Message, error) {
    // Create rich UserMessage instead of weak interface
    var msg shared.UserMessage
    
    // Parse metadata
    msg.SetMetadata(extractMetadata(data))
    
    // Parse content with type safety
    if contentData, ok := data["content"]; ok {
        content, err := parseUserContent(contentData)
        if err != nil {
            return nil, err
        }
        msg.SetContent(content)
    }
    
    return &msg, nil
}
```

## Component 4: Client Integration (`client.go`)

### Current State Analysis
- **File**: `client.go:85` - Returns generic `Message` interface
- **Problem**: Forces type assertions in examples
- **Test Issues**: Tests don't validate rich interface usage

### RED Phase: Client Rich Interface Tests

```go
// File: client_test.go (Add new tests)
func TestClientDeliversRichInterfaces(t *testing.T) {
    ctx := context.Background()
    transport := newMockTransport()
    client := NewClientWithTransport(transport)
    
    err := client.Connect(ctx)
    require.NoError(t, err)
    defer client.Disconnect()
    
    err = client.Query(ctx, "Test message")
    require.NoError(t, err)
    
    // Test receiving messages through rich interfaces
    msgChan := client.ReceiveMessages(ctx)
    select {
    case msg := <-msgChan:
        if msg != nil {
            // Should work without type assertions
            validateRichMessage(t, msg)
        }
    case <-time.After(5 * time.Second):
        t.Fatal("timeout waiting for message")
    }
}

func validateRichMessage(t *testing.T, msg Message) {
    // Validate rich interface WITHOUT type assertions
    assert.NotEmpty(t, msg.GetID())
    assert.NotEmpty(t, msg.Type())
    
    content := msg.GetContent()
    assert.NotNil(t, content)
    assert.GreaterOrEqual(t, content.GetBlockCount(), 0)
    
    // Test content iteration without type assertions
    for i := 0; i < content.GetBlockCount(); i++ {
        block, err := content.GetBlock(i)
        assert.NoError(t, err)
        assert.NotEmpty(t, block.BlockType())
    }
}
```

## Component 5: Example Transformations

### Current State Analysis
- **Files**: All `examples/*/main.go` use type assertions
- **Problem**: Examples teach anti-patterns
- **Issue**: `switch msg := message.(type)` everywhere

### RED Phase: Example Validation Tests

```go
// File: examples_test.go (New File)
func TestExamplesUseRichInterfaces(t *testing.T) {
    examples := []string{
        "examples/01_quickstart",
        "examples/02_client_streaming",
        // ... all examples
    }
    
    for _, example := range examples {
        t.Run(example, func(t *testing.T) {
            // Read example source code
            code, err := os.ReadFile(example + "/main.go")
            require.NoError(t, err)
            
            // Verify no type assertions used
            assert.NotContains(t, string(code), ".(type)")
            assert.NotContains(t, string(code), ".(*claudecode.")
            assert.NotContains(t, string(code), "switch msg := message.(type)")
            
            // Verify rich interface usage
            assert.Contains(t, string(code), ".GetContent()")
            assert.Contains(t, string(code), ".AsText()")
        })
    }
}
```

### GREEN Phase: Transform Examples

**Objective**: Rewrite examples to use rich interfaces

```go
// File: examples/01_quickstart/main.go (Complete rewrite)
// OLD (Type assertions):
switch msg := message.(type) {
case *claudecode.AssistantMessage:
    for _, block := range msg.Content {
        if textBlock, ok := block.(*claudecode.TextBlock); ok {
            fmt.Print(textBlock.Text)
        }
    }
}

// NEW (Rich interfaces):
fmt.Printf("Message %s: %s\n", 
    message.GetID(), 
    message.GetContent().AsText())

for _, block := range message.GetContent().AsBlocks() {
    fmt.Printf("  %s: %s\n", block.BlockType(), block.String())
}
```

## Implementation Timeline Per Component

### Week 1: Foundation Components
- **Day 1-2**: Message interface (RED-GREEN-REFACTOR)
- **Day 3-4**: ContentBlock interface (RED-GREEN-REFACTOR)  
- **Day 5**: Integration tests between Message and ContentBlock

### Week 2: Parser Integration
- **Day 1-2**: Parser updates (RED-GREEN-REFACTOR)
- **Day 3-4**: JSON parsing round-trip tests
- **Day 5**: Parser performance validation

### Week 3: Client Integration
- **Day 1-2**: Client interface updates (RED-GREEN-REFACTOR)
- **Day 3-4**: Transport integration with rich interfaces
- **Day 5**: End-to-end client tests

### Week 4: Examples and Documentation
- **Day 1-3**: Transform all examples (RED-GREEN-REFACTOR)
- **Day 4-5**: Example validation tests and documentation

### Week 5: Performance and Polish
- **Day 1-3**: Performance optimization and benchmarks
- **Day 4-5**: Final integration tests and polish

## Success Metrics Per Component

### Message Interface Success
- ✅ Zero type assertions in message processing
- ✅ Rich content access through ContentAccessor
- ✅ Validation methods work correctly
- ✅ JSON serialization maintains compatibility

### ContentBlock Interface Success  
- ✅ Specialized interfaces (TextProvider, ToolProvider) functional
- ✅ Block iteration without type assertions
- ✅ Type-safe content manipulation

### Parser Integration Success
- ✅ Creates messages with rich interfaces
- ✅ No `interface{}` in parsed data structures
- ✅ Performance within 10% of current implementation

### Client Integration Success
- ✅ Client delivers rich interfaces to users
- ✅ Message channels work with new interface types
- ✅ Resource management unaffected

### Example Transformation Success
- ✅ All examples compile and run
- ✅ Zero type assertions in example code
- ✅ Examples demonstrate rich interface patterns
- ✅ Examples pass validation tests

## Risk Mitigation Strategies

### Interface Evolution Risk
**Risk**: New interface methods break existing code
**Mitigation**: Implement interfaces incrementally, maintain backward compatibility during transition

### Performance Regression Risk  
**Risk**: Rich interfaces slower than type assertions
**Mitigation**: Benchmark each component, optimize interface method calls

### Test Complexity Risk
**Risk**: Rich interface tests too complex
**Mitigation**: Start with simple interface contracts, add complexity incrementally

### Example Maintenance Risk
**Risk**: Examples become too abstract
**Mitigation**: Keep examples practical, demonstrate real use cases

This test-first approach ensures that every component transformation is driven by concrete interface requirements, validated by comprehensive tests, and results in idiomatic Go code that eliminates the `interface{}` anti-pattern while maintaining functionality and performance.