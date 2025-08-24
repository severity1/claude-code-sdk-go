# Critical Edge Cases Analysis

Analysis of edge cases discovered in Python SDK tests and implementation that are crucial for Go implementation.

## Multi-JSON Single Line Handling (test_subprocess_buffering.py)

**Python Test Case**:
```python
def test_multiple_json_objects_on_single_line(self) -> None:
    """Test parsing when multiple JSON objects are concatenated on a single line.
    
    In some environments, stdout buffering can cause multiple distinct JSON
    objects to be delivered as a single line with embedded newlines.
    """
    json_obj1 = {"type": "message", "id": "msg1", "content": "First message"}
    json_obj2 = {"type": "result", "id": "res1", "status": "completed"}
    
    # This is what gets received from CLI
    buffered_line = json.dumps(json_obj1) + "\n" + json.dumps(json_obj2)
```

**Processing Logic**:
```python
# Must split lines and process each JSON object separately
json_lines = line_str.split("\n")
for json_line in json_lines:
    json_line = json_line.strip()
    if not json_line:
        continue
    # Process each JSON line separately
```

**Go Implementation Required**:
```go
func (t *SubprocessTransport) processLine(line string) error {
    // Handle multiple JSON objects on single line
    jsonLines := strings.Split(strings.TrimSpace(line), "\n")
    
    for _, jsonLine := range jsonLines {
        jsonLine = strings.TrimSpace(jsonLine)
        if jsonLine == "" {
            continue
        }
        
        if err := t.processJSONLine(jsonLine); err != nil {
            return fmt.Errorf("failed to process JSON line: %w", err)
        }
    }
    
    return nil
}
```

## Embedded Newlines in JSON Values

**Python Test Case**:
```python
def test_json_with_embedded_newlines(self) -> None:
    """Test parsing JSON objects that contain newline characters in string values."""
    json_obj1 = {"type": "message", "content": "Line 1\nLine 2\nLine 3"}
    json_obj2 = {"type": "result", "data": "Some\nMultiline\nContent"}
    
    # Parser must handle newlines within JSON string values correctly
    buffered_line = json.dumps(json_obj1) + "\n" + json.dumps(json_obj2)
```

**Critical Requirement**: The JSON parser must correctly handle newline characters that are embedded within JSON string values, distinguishing them from newlines that separate JSON objects.

**Go Implementation**:
```go
// Go's standard json.Unmarshal handles this correctly by default
var data map[string]interface{}
if err := json.Unmarshal([]byte(jsonLine), &data); err != nil {
    return fmt.Errorf("failed to parse JSON: %w", err)
}

// Content with embedded newlines is preserved
if content, ok := data["content"].(string); ok {
    // content contains "Line 1\nLine 2\nLine 3" correctly
}
```

## Buffer Overflow Protection

**Python Protection**:
```python
json_buffer += json_line
if len(json_buffer) > _MAX_BUFFER_SIZE:  # 1MB limit
    json_buffer = ""
    raise SDKJSONDecodeError(
        f"JSON message exceeded maximum buffer size of {_MAX_BUFFER_SIZE} bytes",
        ValueError(f"Buffer size {len(json_buffer)} exceeds limit {_MAX_BUFFER_SIZE}")
    )
```

**Go Implementation**:
```go
const MaxBufferSize = 1024 * 1024 // 1MB

func (t *SubprocessTransport) processJSONLine(line string) error {
    t.jsonBuffer.WriteString(line)
    
    // Check buffer size limit
    if t.jsonBuffer.Len() > MaxBufferSize {
        bufferSize := t.jsonBuffer.Len()
        t.jsonBuffer.Reset()
        return NewJSONDecodeError(
            "buffer size exceeded",
            fmt.Errorf("buffer size %d exceeds limit %d", bufferSize, MaxBufferSize),
        )
    }
    
    // ... continue with parsing
}
```

## Speculative JSON Parsing Strategy

**Python Strategy**:
```python
try:
    data = json.loads(json_buffer)
    json_buffer = ""  # Reset buffer on successful parse
    yield data
except json.JSONDecodeError:
    # Keep accumulating until we get complete JSON
    # This is the key innovation - don't fail, just continue
    continue
```

**Go Implementation**:
```go
func (t *SubprocessTransport) tryParseJSON() (map[string]interface{}, error) {
    if t.jsonBuffer.Len() == 0 {
        return nil, nil
    }
    
    var data map[string]interface{}
    bufferContent := t.jsonBuffer.String()
    
    if err := json.Unmarshal([]byte(bufferContent), &data); err != nil {
        // JSON is incomplete, continue accumulating
        // This is NOT an error condition
        return nil, nil
    }
    
    // Successfully parsed complete JSON
    t.jsonBuffer.Reset()
    return data, nil
}

func (t *SubprocessTransport) processJSONLine(line string) error {
    t.jsonBuffer.WriteString(line)
    
    // Check size limit
    if t.jsonBuffer.Len() > MaxBufferSize {
        t.jsonBuffer.Reset()
        return NewJSONDecodeError("buffer overflow", ErrBufferOverflow)
    }
    
    // Try to parse accumulated buffer
    data, err := t.tryParseJSON()
    if err != nil {
        return err
    }
    
    if data != nil {
        // Successfully parsed complete JSON
        return t.handleMessage(data)
    }
    
    // No complete JSON yet, continue accumulating
    return nil
}
```

## Process State Mocking for Tests

**Python Mock Pattern**:
```python
def test_buffering_edge_case(self) -> None:
    transport = SubprocessCLITransport(
        prompt="test", options=ClaudeCodeOptions(), cli_path="/usr/bin/claude"
    )
    
    # Mock the process and streams
    mock_process = MagicMock()
    mock_process.returncode = None
    mock_process.wait = AsyncMock(return_value=None)
    transport._process = mock_process
    
    transport._stdout_stream = MockTextReceiveStream([buffered_line])
    transport._stderr_stream = MockTextReceiveStream([])
    
    # Test the actual parsing logic
    messages: list[Any] = []
    async for msg in transport.receive_messages():
        messages.append(msg)
```

**Go Mock Interface**:
```go
type MockTransport struct {
    messages    []map[string]interface{}
    errors      []error
    connected   bool
    interrupted bool
}

func (m *MockTransport) Connect(ctx context.Context) error {
    m.connected = true
    return nil
}

func (m *MockTransport) ReceiveMessages(ctx context.Context) (<-chan map[string]interface{}, <-chan error) {
    msgChan := make(chan map[string]interface{}, len(m.messages))
    errChan := make(chan error, len(m.errors))
    
    go func() {
        defer close(msgChan)
        defer close(errChan)
        
        for _, msg := range m.messages {
            select {
            case msgChan <- msg:
            case <-ctx.Done():
                return
            }
        }
        
        for _, err := range m.errors {
            select {
            case errChan <- err:
            case <-ctx.Done():
                return
            }
        }
    }()
    
    return msgChan, errChan
}

// Test usage
func TestBufferingEdgeCases(t *testing.T) {
    transport := &MockTransport{
        messages: []map[string]interface{}{
            {"type": "message", "content": "Line 1\nLine 2"},
            {"type": "result", "status": "completed"},
        },
    }
    
    ctx := context.Background()
    msgChan, errChan := transport.ReceiveMessages(ctx)
    
    var messages []map[string]interface{}
    for msg := range msgChan {
        messages = append(messages, msg)
    }
    
    assert.Len(t, messages, 2)
    assert.Equal(t, "Line 1\nLine 2", messages[0]["content"])
}
```

## Stderr Truncation and Memory Management

**Python Stderr Management**:
```python
# Memory-efficient stderr capture with truncation
stderr_lines: deque[str] = deque(maxlen=100)  # Keep only last 100 lines

if len(stderr_lines) == stderr_lines.maxlen:
    stderr_output = (
        f"[stderr truncated, showing last {stderr_lines.maxlen} lines]\n"
        + stderr_output
    )
```

**Go Implementation**:
```go
type CircularBuffer struct {
    lines    []string
    maxLines int
    current  int
    full     bool
}

func NewCircularBuffer(maxLines int) *CircularBuffer {
    return &CircularBuffer{
        lines:    make([]string, maxLines),
        maxLines: maxLines,
    }
}

func (cb *CircularBuffer) Add(line string) {
    cb.lines[cb.current] = line
    cb.current = (cb.current + 1) % cb.maxLines
    if cb.current == 0 {
        cb.full = true
    }
}

func (cb *CircularBuffer) GetAll() []string {
    if !cb.full {
        return cb.lines[:cb.current]
    }
    
    result := make([]string, cb.maxLines+1)
    result[0] = fmt.Sprintf("[stderr truncated, showing last %d lines]", cb.maxLines)
    
    // Copy from current position to end
    copy(result[1:], cb.lines[cb.current:])
    // Copy from beginning to current position
    copy(result[1+len(cb.lines[cb.current:]):], cb.lines[:cb.current])
    
    return result
}

// Usage in subprocess transport
type SubprocessTransport struct {
    stderrBuffer *CircularBuffer
}

func (t *SubprocessTransport) captureStderr() {
    t.stderrBuffer = NewCircularBuffer(100)
    
    scanner := bufio.NewScanner(t.stderrFile)
    for scanner.Scan() {
        t.stderrBuffer.Add(scanner.Text())
    }
}
```

## Generator Exit Handling

**Python GeneratorExit Handling**:
```python
try:
    yield data
except GeneratorExit:
    return  # Clean exit from generator
```

**Go Context Cancellation Equivalent**:
```go
func (t *SubprocessTransport) processMessages(ctx context.Context, msgChan chan<- map[string]interface{}) {
    defer close(msgChan)
    
    for {
        select {
        case <-ctx.Done():
            // Equivalent to GeneratorExit - clean shutdown
            return
        default:
            // Process messages
            msg, err := t.readNextMessage()
            if err != nil {
                return
            }
            
            select {
            case msgChan <- msg:
            case <-ctx.Done():
                return
            }
        }
    }
}
```

## Resource Cleanup Edge Cases

**Python Cleanup Pattern**:
```python
# Clean up temp file even if operations fail
if self._stderr_file:
    try:
        self._stderr_file.close()
        Path(self._stderr_file.name).unlink()
    except Exception:
        pass  # Don't fail cleanup on error
    self._stderr_file = None
```

**Go Cleanup Pattern**:
```go
func (t *SubprocessTransport) Disconnect() error {
    var errs []error
    
    // Close stdin
    if t.stdin != nil {
        if err := t.stdin.Close(); err != nil {
            errs = append(errs, fmt.Errorf("failed to close stdin: %w", err))
        }
        t.stdin = nil
    }
    
    // Close stdout
    if t.stdout != nil {
        if err := t.stdout.Close(); err != nil {
            errs = append(errs, fmt.Errorf("failed to close stdout: %w", err))
        }
        t.stdout = nil
    }
    
    // Clean up stderr temp file
    if t.stderrFile != nil {
        name := t.stderrFile.Name()
        
        // Close file (ignore error)
        t.stderrFile.Close()
        
        // Remove temp file (ignore error)
        os.Remove(name)
        
        t.stderrFile = nil
    }
    
    // Terminate process
    if t.cmd != nil && t.cmd.Process != nil {
        if err := t.terminateProcess(); err != nil {
            errs = append(errs, fmt.Errorf("failed to terminate process: %w", err))
        }
    }
    
    t.setConnected(false)
    
    if len(errs) > 0 {
        return fmt.Errorf("cleanup errors: %v", errs)
    }
    
    return nil
}
```

## Critical Testing Requirements

1. **Multi-JSON line parsing** - Must handle multiple JSON objects per line
2. **Embedded newlines** - Must preserve newlines within JSON string values
3. **Buffer overflow protection** - Must prevent memory exhaustion with 1MB limit
4. **Speculative parsing** - Must continue accumulating until complete JSON
5. **Process mocking** - Must support comprehensive mocking for unit tests
6. **Resource cleanup** - Must handle cleanup failures gracefully
7. **Memory management** - Must use bounded buffers for stderr capture

These edge cases are critical for production reliability and must be implemented with comprehensive test coverage in the Go SDK.