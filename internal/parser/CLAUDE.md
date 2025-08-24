# Message Parsing Context

**Context**: JSON message parsing with speculative parsing, buffer management, and union type discrimination for Phase 2 implementation

## Component Focus
- **Speculative JSON Parsing** - Accumulate buffer until complete JSON, handle partial messages
- **Buffer Management** - 1MB limit protection, memory-efficient handling
- **Union Type Discrimination** - Parse polymorphic Message and ContentBlock types
- **Edge Case Handling** - Multiple JSON objects, embedded newlines, malformed input

## Phase 2 TDD Focus
This component implements Phase 2 of TDD tasks (T035-T082): Message Parsing & Validation

### Critical Parsing Patterns

#### Speculative JSON Parsing
```go
// Core pattern: Accumulate until complete JSON
func (p *Parser) ProcessLine(line string) ([]any, error) {
    p.buffer.WriteString(line)
    
    // Try to parse - if fails, continue accumulating (not an error)
    var data any
    if err := json.Unmarshal([]byte(p.buffer.String()), &data); err != nil {
        // Not complete JSON yet - continue accumulating
        return nil, nil
    }
    
    // Success - reset buffer and return parsed data
    p.buffer.Reset()
    return []any{data}, nil
}
```

#### Buffer Overflow Protection
- **1MB Limit**: Reset buffer and return error when exceeded
- **Memory Safety**: Prevent unbounded buffer growth
- **Attack Protection**: Guard against memory exhaustion

#### Multiple JSON Objects on Single Line
```go
// Handle: {"type":"user","content":"hello"}{"type":"assistant","content":"hi"}
// Must parse both objects separately
```

#### Embedded Newlines in JSON Strings
```go
// Handle: {"type":"user","content":"Line 1\nLine 2\nLine 3"}
// Must not break parsing on internal newlines
```

## Required Implementation Patterns

### Message Type Discrimination
```go
func UnmarshalMessage(data []byte) (Message, error) {
    var typeCheck struct {
        Type string `json:"type"`
    }
    
    if err := json.Unmarshal(data, &typeCheck); err != nil {
        return nil, err
    }
    
    switch typeCheck.Type {
    case "user":
        var msg UserMessage
        return &msg, json.Unmarshal(data, &msg)
    case "assistant":
        var msg AssistantMessage
        return &msg, json.Unmarshal(data, &msg)
    // ... other types
    }
}
```

### Content Block Discrimination
- Parse heterogeneous arrays of ContentBlock types
- Type-based instantiation using "type" field
- Preserve unknown fields for extensibility

### Error Handling Patterns
- **MessageParseError** - Preserve original data context
- **JSONDecodeError** - Include line and position information
- **Graceful Recovery** - Continue parsing after errors

## Buffer Management
- **Efficient Growth** - StringBuilder for accumulation
- **Size Tracking** - Monitor buffer size against limits
- **Reset Strategy** - Clear buffer after successful parse
- **Memory Pressure** - Handle low memory gracefully

## Performance Requirements
- **Streaming Optimized** - Efficient for continuous message streams
- **Memory Efficient** - Minimal heap pressure
- **Fast Path** - Optimize for common cases (complete JSON on first try)
- **Concurrent Safe** - Support multiple parser instances

## Integration Requirements
- Must integrate with transport layer for CLI output processing
- Support both one-shot and streaming parsing modes
- Provide clear error messages with context for debugging
- Handle all edge cases identified in Phase 2 tasks

## Critical Test Cases (Phase 2)
- Multiple JSON objects on single line parsing
- Embedded newlines within JSON string values
- Buffer overflow scenarios with 1MB+ content
- Malformed JSON recovery and continuation
- Rapid message burst handling
- Memory leak prevention under sustained load