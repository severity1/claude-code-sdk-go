# Message Parsing Analysis

Comprehensive analysis of the Python SDK's message parsing logic and JSON stream processing patterns.

## Message Parser Architecture (_internal/message_parser.py)

**Core Parser Class**:
```python
class MessageParser:
    """Parses messages from Claude Code CLI output."""
    
    def __init__(self):
        self.json_buffer = ""
        self.max_buffer_size = 1024 * 1024  # 1MB limit
    
    def parse_messages(self, raw_lines: list[str]) -> Iterator[dict[str, Any]]:
        """Parse messages from raw CLI output lines."""
        for line in raw_lines:
            yield from self._parse_line(line)
    
    def _parse_line(self, line: str) -> Iterator[dict[str, Any]]:
        """Parse a single line that may contain multiple JSON objects."""
        line = line.strip()
        if not line:
            return
        
        # Handle multiple JSON objects on single line
        json_lines = line.split('\n')
        for json_line in json_lines:
            json_line = json_line.strip()
            if json_line:
                yield from self._parse_json_line(json_line)
```

**Go Message Parser Architecture**:
```go
type MessageParser struct {
    jsonBuffer    strings.Builder
    maxBufferSize int
    mu            sync.Mutex
}

func NewMessageParser() *MessageParser {
    return &MessageParser{
        maxBufferSize: 1024 * 1024, // 1MB limit
    }
}

func (p *MessageParser) ParseMessages(lines []string) (<-chan map[string]interface{}, <-chan error) {
    msgChan := make(chan map[string]interface{}, 100)
    errChan := make(chan error, 10)
    
    go func() {
        defer close(msgChan)
        defer close(errChan)
        
        for _, line := range lines {
            if err := p.parseLine(line, msgChan, errChan); err != nil {
                errChan <- err
                return
            }
        }
    }()
    
    return msgChan, errChan
}

func (p *MessageParser) parseLine(line string, msgChan chan<- map[string]interface{}, errChan chan<- error) error {
    line = strings.TrimSpace(line)
    if line == "" {
        return nil
    }
    
    // Handle multiple JSON objects on single line
    jsonLines := strings.Split(line, "\n")
    for _, jsonLine := range jsonLines {
        jsonLine = strings.TrimSpace(jsonLine)
        if jsonLine != "" {
            if err := p.parseJSONLine(jsonLine, msgChan, errChan); err != nil {
                return err
            }
        }
    }
    
    return nil
}
```

## Speculative JSON Parsing Strategy

**Python Speculative Parsing**:
```python
def _parse_json_line(self, json_line: str) -> Iterator[dict[str, Any]]:
    """Parse JSON line with speculative buffering."""
    self.json_buffer += json_line
    
    # Check buffer size limit
    if len(self.json_buffer) > self.max_buffer_size:
        self.json_buffer = ""
        raise CLIJSONDecodeError(
            f"JSON buffer exceeded {self.max_buffer_size} bytes",
            ValueError("Buffer overflow")
        )
    
    try:
        # Attempt to parse accumulated buffer
        data = json.loads(self.json_buffer)
        self.json_buffer = ""  # Reset on successful parse
        yield data
    except json.JSONDecodeError:
        # Continue accumulating - this is not an error condition!
        # The JSON may be incomplete and span multiple lines
        pass
```

**Go Speculative Parsing**:
```go
func (p *MessageParser) parseJSONLine(jsonLine string, msgChan chan<- map[string]interface{}, errChan chan<- error) error {
    p.mu.Lock()
    defer p.mu.Unlock()
    
    p.jsonBuffer.WriteString(jsonLine)
    
    // Check buffer size limit
    if p.jsonBuffer.Len() > p.maxBufferSize {
        bufferSize := p.jsonBuffer.Len()
        p.jsonBuffer.Reset()
        return NewJSONDecodeError(
            "buffer overflow",
            fmt.Errorf("buffer size %d exceeds limit %d", bufferSize, p.maxBufferSize),
        )
    }
    
    // Attempt speculative parsing
    var data map[string]interface{}
    bufferContent := p.jsonBuffer.String()
    
    if err := json.Unmarshal([]byte(bufferContent), &data); err != nil {
        // JSON is incomplete, continue accumulating
        // This is NOT an error condition!
        return nil
    }
    
    // Successfully parsed complete JSON
    p.jsonBuffer.Reset()
    
    select {
    case msgChan <- data:
        return nil
    default:
        return fmt.Errorf("message channel full")
    }
}
```

## Message Type Detection and Routing

**Python Type Detection**:
```python
def parse_message(self, data: dict[str, Any]) -> Message:
    """Convert raw JSON to typed Message object."""
    msg_type = data.get("type")
    
    if msg_type == "user":
        return self._parse_user_message(data)
    elif msg_type == "assistant":
        return self._parse_assistant_message(data)
    elif msg_type == "system":
        return self._parse_system_message(data)
    elif msg_type == "result":
        return self._parse_result_message(data)
    else:
        raise MessageParseError(f"Unknown message type: {msg_type}", data)

def _parse_user_message(self, data: dict[str, Any]) -> UserMessage:
    """Parse user message from raw data."""
    content = data.get("content", [])
    
    if isinstance(content, str):
        # String content - wrap in text block
        return UserMessage(content=[TextBlock(text=content)])
    elif isinstance(content, list):
        # List of content blocks
        blocks = [self._parse_content_block(block) for block in content]
        return UserMessage(content=blocks)
    else:
        raise MessageParseError("Invalid user message content", data)
```

**Go Type Detection**:
```go
func (p *MessageParser) ParseMessage(data map[string]interface{}) (Message, error) {
    msgType, ok := data["type"].(string)
    if !ok {
        return nil, NewMessageParseError("missing or invalid type field", data)
    }
    
    switch msgType {
    case "user":
        return p.parseUserMessage(data)
    case "assistant":
        return p.parseAssistantMessage(data)
    case "system":
        return p.parseSystemMessage(data)
    case "result":
        return p.parseResultMessage(data)
    default:
        return nil, NewMessageParseError(
            fmt.Sprintf("unknown message type: %s", msgType),
            data,
        )
    }
}

func (p *MessageParser) parseUserMessage(data map[string]interface{}) (*UserMessage, error) {
    content, ok := data["content"]
    if !ok {
        return nil, NewMessageParseError("missing content field in user message", data)
    }
    
    switch c := content.(type) {
    case string:
        // String content - wrap in text block
        return &UserMessage{
            Content: []ContentBlock{&TextBlock{Text: c}},
        }, nil
    case []interface{}:
        // List of content blocks
        blocks := make([]ContentBlock, len(c))
        for i, blockData := range c {
            block, err := p.parseContentBlock(blockData)
            if err != nil {
                return nil, fmt.Errorf("failed to parse content block %d: %w", i, err)
            }
            blocks[i] = block
        }
        return &UserMessage{Content: blocks}, nil
    default:
        return nil, NewMessageParseError("invalid user message content type", data)
    }
}
```

## Content Block Parsing

**Python Content Block Factory**:
```python
def _parse_content_block(self, block_data: dict[str, Any]) -> ContentBlock:
    """Parse content block based on type."""
    block_type = block_data.get("type")
    
    if block_type == "text":
        return TextBlock(text=block_data["text"])
    elif block_type == "thinking":
        return ThinkingBlock(
            thinking=block_data["thinking"],
            signature=block_data.get("signature", "")
        )
    elif block_type == "tool_use":
        return ToolUseBlock(
            id=block_data["id"],
            name=block_data["name"],
            input=block_data.get("input", {})
        )
    elif block_type == "tool_result":
        return ToolResultBlock(
            tool_use_id=block_data["tool_use_id"],
            content=block_data.get("content"),
            is_error=block_data.get("is_error")
        )
    else:
        raise MessageParseError(f"Unknown content block type: {block_type}", block_data)
```

**Go Content Block Factory**:
```go
func (p *MessageParser) parseContentBlock(blockData interface{}) (ContentBlock, error) {
    data, ok := blockData.(map[string]interface{})
    if !ok {
        return nil, NewMessageParseError("content block must be an object", blockData)
    }
    
    blockType, ok := data["type"].(string)
    if !ok {
        return nil, NewMessageParseError("content block missing type field", data)
    }
    
    switch blockType {
    case "text":
        text, ok := data["text"].(string)
        if !ok {
            return nil, NewMessageParseError("text block missing text field", data)
        }
        return &TextBlock{Text: text}, nil
        
    case "thinking":
        thinking, ok := data["thinking"].(string)
        if !ok {
            return nil, NewMessageParseError("thinking block missing thinking field", data)
        }
        signature, _ := data["signature"].(string) // Optional field
        return &ThinkingBlock{
            Thinking:  thinking,
            Signature: signature,
        }, nil
        
    case "tool_use":
        id, ok := data["id"].(string)
        if !ok {
            return nil, NewMessageParseError("tool_use block missing id field", data)
        }
        name, ok := data["name"].(string)
        if !ok {
            return nil, NewMessageParseError("tool_use block missing name field", data)
        }
        input, _ := data["input"].(map[string]interface{})
        if input == nil {
            input = make(map[string]interface{})
        }
        return &ToolUseBlock{
            ID:    id,
            Name:  name,
            Input: input,
        }, nil
        
    case "tool_result":
        toolUseID, ok := data["tool_use_id"].(string)
        if !ok {
            return nil, NewMessageParseError("tool_result block missing tool_use_id field", data)
        }
        
        var isError *bool
        if isErrorValue, exists := data["is_error"]; exists {
            if b, ok := isErrorValue.(bool); ok {
                isError = &b
            }
        }
        
        return &ToolResultBlock{
            ToolUseID: toolUseID,
            Content:   data["content"],
            IsError:   isError,
        }, nil
        
    default:
        return nil, NewMessageParseError(
            fmt.Sprintf("unknown content block type: %s", blockType),
            data,
        )
    }
}
```

## Field Validation and Error Handling

**Python Validation Patterns**:
```python
def _parse_result_message(self, data: dict[str, Any]) -> ResultMessage:
    """Parse result message with comprehensive validation."""
    try:
        return ResultMessage(
            subtype=data["subtype"],
            duration_ms=int(data["duration_ms"]),
            duration_api_ms=int(data["duration_api_ms"]),
            is_error=bool(data["is_error"]),
            num_turns=int(data["num_turns"]),
            session_id=str(data["session_id"]),
            total_cost_usd=data.get("total_cost_usd"),  # Optional
            usage=data.get("usage"),  # Optional
            result=data.get("result")  # Optional
        )
    except KeyError as e:
        raise MessageParseError(f"Missing required field: {e}", data) from e
    except (ValueError, TypeError) as e:
        raise MessageParseError(f"Invalid field value: {e}", data) from e
```

**Go Validation with Type Assertions**:
```go
func (p *MessageParser) parseResultMessage(data map[string]interface{}) (*ResultMessage, error) {
    result := &ResultMessage{}
    
    // Required fields with validation
    if subtype, ok := data["subtype"].(string); ok {
        result.Subtype = subtype
    } else {
        return nil, NewMessageParseError("result message missing subtype field", data)
    }
    
    if durationMS, ok := data["duration_ms"].(float64); ok {
        result.DurationMS = int(durationMS)
    } else {
        return nil, NewMessageParseError("result message missing or invalid duration_ms field", data)
    }
    
    if durationAPIMS, ok := data["duration_api_ms"].(float64); ok {
        result.DurationAPIMS = int(durationAPIMS)
    } else {
        return nil, NewMessageParseError("result message missing or invalid duration_api_ms field", data)
    }
    
    if isError, ok := data["is_error"].(bool); ok {
        result.IsError = isError
    } else {
        return nil, NewMessageParseError("result message missing or invalid is_error field", data)
    }
    
    if numTurns, ok := data["num_turns"].(float64); ok {
        result.NumTurns = int(numTurns)
    } else {
        return nil, NewMessageParseError("result message missing or invalid num_turns field", data)
    }
    
    if sessionID, ok := data["session_id"].(string); ok {
        result.SessionID = sessionID
    } else {
        return nil, NewMessageParseError("result message missing session_id field", data)
    }
    
    // Optional fields (no validation errors if missing)
    if totalCostUSD, ok := data["total_cost_usd"].(float64); ok {
        result.TotalCostUSD = &totalCostUSD
    }
    
    if usage, ok := data["usage"].(map[string]interface{}); ok {
        result.Usage = usage
    }
    
    if resultStr, ok := data["result"].(string); ok {
        result.Result = &resultStr
    }
    
    return result, nil
}
```

## Buffer Management and Memory Safety

**Buffer Overflow Protection**:
```python
# Python version with memory protection
_MAX_BUFFER_SIZE = 1024 * 1024  # 1MB

class MessageParser:
    def _parse_json_line(self, json_line: str) -> Iterator[dict[str, Any]]:
        self.json_buffer += json_line
        
        if len(self.json_buffer) > _MAX_BUFFER_SIZE:
            # Critical: Reset buffer before raising error
            buffer_size = len(self.json_buffer)
            self.json_buffer = ""
            raise CLIJSONDecodeError(
                f"JSON message exceeded maximum buffer size of {_MAX_BUFFER_SIZE} bytes",
                ValueError(f"Buffer size {buffer_size} exceeds limit")
            )
```

**Go Buffer Management**:
```go
const MaxBufferSize = 1024 * 1024 // 1MB

func (p *MessageParser) checkBufferSize() error {
    if p.jsonBuffer.Len() > MaxBufferSize {
        bufferSize := p.jsonBuffer.Len()
        p.jsonBuffer.Reset() // Critical: Reset before error
        return NewJSONDecodeError(
            "buffer overflow",
            fmt.Errorf("buffer size %d exceeds limit %d", bufferSize, MaxBufferSize),
        )
    }
    return nil
}

func (p *MessageParser) Reset() {
    p.mu.Lock()
    defer p.mu.Unlock()
    p.jsonBuffer.Reset()
}
```

## Thread Safety and Concurrency

**Python Thread Safety** (implicit with GIL):
```python
# Python relies on GIL for thread safety in most cases
# But async operations need careful coordination
class AsyncMessageParser:
    def __init__(self):
        self._lock = anyio.Lock()
        self.json_buffer = ""
    
    async def parse_line_safe(self, line: str) -> AsyncIterator[dict[str, Any]]:
        async with self._lock:
            # All buffer operations protected by lock
            yield from self._parse_line(line)
```

**Go Thread Safety** (explicit synchronization):
```go
type MessageParser struct {
    jsonBuffer    strings.Builder
    maxBufferSize int
    mu            sync.Mutex  // Explicit mutex protection
}

func (p *MessageParser) ParseLineSafe(line string) ([]map[string]interface{}, error) {
    p.mu.Lock()
    defer p.mu.Unlock()
    
    // All buffer operations are mutex-protected
    return p.parseLineInternal(line)
}

// Alternative: Lock-free with channels
type ChannelMessageParser struct {
    input  chan string
    output chan map[string]interface{}
    errors chan error
}

func (p *ChannelMessageParser) Start(ctx context.Context) {
    go func() {
        parser := NewMessageParser()
        for {
            select {
            case line := <-p.input:
                msgs, err := parser.ParseLine(line)
                if err != nil {
                    p.errors <- err
                    continue
                }
                for _, msg := range msgs {
                    p.output <- msg
                }
            case <-ctx.Done():
                return
            }
        }
    }()
}
```

## Error Recovery and Resilience

**Python Error Recovery**:
```python
def parse_with_recovery(self, lines: list[str]) -> Iterator[dict[str, Any]]:
    """Parse messages with error recovery."""
    for line_num, line in enumerate(lines):
        try:
            yield from self._parse_line(line)
        except CLIJSONDecodeError as e:
            # Log error but continue processing
            logger.warning(f"JSON decode error on line {line_num}: {e}")
            self.json_buffer = ""  # Reset buffer to recover
            continue
        except Exception as e:
            # Unexpected errors should still be raised
            raise MessageParseError(f"Unexpected error on line {line_num}: {e}") from e
```

**Go Error Recovery**:
```go
func (p *MessageParser) ParseWithRecovery(lines []string) ([]map[string]interface{}, []error) {
    var messages []map[string]interface{}
    var errors []error
    
    for i, line := range lines {
        msgs, err := p.ParseLine(line)
        if err != nil {
            var jsonErr *JSONDecodeError
            if errors.As(err, &jsonErr) {
                // Log and continue for JSON decode errors
                log.Printf("JSON decode error on line %d: %v", i, err)
                p.Reset() // Reset buffer to recover
                errors = append(errors, fmt.Errorf("line %d: %w", i, err))
                continue
            }
            // Other errors are fatal
            return messages, append(errors, fmt.Errorf("fatal error on line %d: %w", i, err))
        }
        messages = append(messages, msgs...)
    }
    
    return messages, errors
}
```

## Performance Optimizations

**Go-Specific Optimizations**:
```go
// Pre-allocate slices for better performance
func (p *MessageParser) ParseMessagesOptimized(lines []string) ([]map[string]interface{}, error) {
    // Pre-allocate with estimated capacity
    messages := make([]map[string]interface{}, 0, len(lines))
    
    for _, line := range lines {
        if line = strings.TrimSpace(line); line == "" {
            continue
        }
        
        // Fast path for single JSON object lines
        if !strings.Contains(line, "\n") {
            if msg, err := p.parseSingleJSON(line); err == nil {
                messages = append(messages, msg)
                continue
            }
        }
        
        // Slower path for multi-JSON lines
        lineMsgs, err := p.parseLineInternal(line)
        if err != nil {
            return messages, err
        }
        messages = append(messages, lineMsgs...)
    }
    
    return messages, nil
}

// Object pooling for high-frequency parsing
var messageParserPool = sync.Pool{
    New: func() interface{} {
        return NewMessageParser()
    },
}

func ParseMessagesPooled(lines []string) ([]map[string]interface{}, error) {
    parser := messageParserPool.Get().(*MessageParser)
    defer messageParserPool.Put(parser)
    
    parser.Reset() // Ensure clean state
    return parser.ParseMessagesOptimized(lines)
}
```

This comprehensive message parsing system provides robust JSON stream processing with proper error handling, memory protection, and Go-native concurrency patterns while maintaining 100% compatibility with the Python SDK's message parsing behavior.