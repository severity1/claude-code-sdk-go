# Subprocess Implementation Details

Complete analysis of the Python SDK's subprocess transport implementation with all critical details for Go implementation.

## Process Termination Sequence

**5-Second SIGTERM → SIGKILL Pattern**:
```python
# Graceful termination with timeout
self._process.terminate()
with anyio.fail_after(5.0):
    await self._process.wait()
except TimeoutError:
    self._process.kill()  # Force kill after 5 seconds
    await self._process.wait()
```

**Go Implementation Pattern**:
```go
func (t *Transport) terminate(ctx context.Context) error {
    if err := t.cmd.Process.Signal(os.Interrupt); err != nil {
        return err
    }
    
    done := make(chan error, 1)
    go func() {
        done <- t.cmd.Wait()
    }()
    
    select {
    case err := <-done:
        return err
    case <-time.After(5 * time.Second):
        return t.cmd.Process.Kill()
    case <-ctx.Done():
        return ctx.Err()
    }
}
```

## Control Request Protocol Implementation

**Unique Request ID Generation**:
```python
# Generate unique request ID
self._request_counter += 1
request_id = f"req_{self._request_counter}_{os.urandom(4).hex()}"

# Build control request
control_request = {
    "type": "control_request",
    "request_id": request_id,
    "request": request,
}
```

**Polling for Control Response**:
```python
# Wait for response with 100ms polling
while request_id not in self._pending_control_responses:
    await anyio.sleep(0.1)  # 100ms polling interval

response = self._pending_control_responses.pop(request_id)
```

**Critical Go Implementation Details**:
- Must use atomic counter for request IDs
- 100ms polling interval exactly
- Response correlation using map with request ID keys
- Proper cleanup of pending requests on timeout/cancellation

## Advanced JSON Buffer Management

**Buffer Size Protection**:
```python
json_buffer += json_line
if len(json_buffer) > _MAX_BUFFER_SIZE:  # 1MB protection
    json_buffer = ""
    raise SDKJSONDecodeError("Buffer size exceeded")
```

**Speculative Parsing Strategy**:
```python
# Key innovation: speculative parsing
try:
    data = json.loads(json_buffer)
    json_buffer = ""  # Reset buffer on successful parse
    yield data
except json.JSONDecodeError:
    continue  # Keep accumulating until complete JSON
```

**Multi-Line JSON Processing**:
```python
# Handle multiple JSON objects on single line
json_lines = line_str.split("\n")
for json_line in json_lines:
    json_line = json_line.strip()
    if not json_line:
        continue
    # Process each JSON line separately
```

## Stderr Management Strategy

**Memory-Efficient Stderr Capture**:
```python
stderr_lines: deque[str] = deque(maxlen=100)  # Keep only last 100 lines
if len(stderr_lines) == stderr_lines.maxlen:
    stderr_output = f"[stderr truncated, showing last {stderr_lines.maxlen} lines]\n"
```

**Temporary File Usage**:
```python
# Create temp file for stderr to avoid pipe deadlocks
self._stderr_file = tempfile.NamedTemporaryFile(
    mode="w+", prefix="claude_stderr_", suffix=".log", delete=False
)

# Use temp file as stderr
self._process = await anyio.open_process(
    cmd,
    stdin=PIPE,
    stdout=PIPE,
    stderr=self._stderr_file,  # Avoids deadlock
    cwd=self._cwd,
    env={**os.environ, "CLAUDE_CODE_ENTRYPOINT": "sdk-py"},
)
```

## Working Directory Validation

**Pre-Process Validation**:
```python
except FileNotFoundError as e:
    # Distinguish between cwd and CLI path issues
    if self._cwd and not Path(self._cwd).exists():
        raise CLIConnectionError(
            f"Working directory does not exist: {self._cwd}"
        ) from e
    raise CLINotFoundError(f"Claude Code not found at: {self._cli_path}") from e
```

## Message Structure Auto-Correction

**Malformed Message Handling**:
```python
# Auto-correct message structure in send_request
if not isinstance(message, dict):
    message = {
        "type": "user",
        "message": {"role": "user", "content": str(message)},
        "parent_tool_use_id": None,
        "session_id": options.get("session_id", "default"),
    }

await self._stdin_stream.send(json.dumps(message) + "\n")
```

## stdin Behavior Differentiation

**Critical Discovery - The Key Difference**:
```python
# query() function uses close_stdin_after_prompt=True
chosen_transport = SubprocessCLITransport(
    prompt=prompt, options=options, close_stdin_after_prompt=True
)

# ClaudeSDKClient keeps stdin open for bidirectional communication
# (close_stdin_after_prompt defaults to False)
```

**Implementation Logic**:
```python
# Close stdin after prompt if requested (e.g., for query() one-shot mode)
if self._close_stdin_after_prompt and self._stdin_stream:
    await self._stdin_stream.aclose()
    self._stdin_stream = None
# Otherwise keep stdin open for send_request (ClaudeSDKClient interactive mode)
```

## Control Response Handling

**Separate Control Message Processing**:
```python
# Handle control responses separately from regular messages
if data.get("type") == "control_response":
    response = data.get("response", {})
    request_id = response.get("request_id")
    if request_id:
        # Store the response for the pending request
        self._pending_control_responses[request_id] = response
    continue  # Don't yield control responses to user
```

## Process State Management

**Connection State Tracking**:
```python
def is_connected(self) -> bool:
    """Check if subprocess is running."""
    return self._process is not None and self._process.returncode is None
```

**Resource Cleanup**:
```python
# Clean up temp file
if self._stderr_file:
    try:
        self._stderr_file.close()
        Path(self._stderr_file.name).unlink()
    except Exception:
        pass
    self._stderr_file = None

# Reset all streams
self._process = None
self._stdout_stream = None
self._stderr_stream = None
self._stdin_stream = None
```

## Go Implementation Critical Requirements

1. **Exact 5-second termination timeout** with SIGTERM → SIGKILL
2. **100ms polling interval** for control response handling  
3. **1MB buffer protection** with overflow error
4. **Speculative JSON parsing** - continue until complete object
5. **Deque-equivalent stderr management** with 100-line limit
6. **Working directory validation** before process start
7. **Message auto-correction** for malformed send inputs
8. **Control response correlation** with unique request IDs
9. **Stdin behavior differentiation** based on close_stdin_after_prompt
10. **Temporary file stderr** to prevent pipe deadlocks

These implementation details are critical for achieving 100% feature parity and handling all the edge cases that the Python SDK manages.