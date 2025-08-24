# Error System Analysis

Analysis of the Python SDK's comprehensive error hierarchy and handling patterns.

## Error Hierarchy (_errors.py)

**Python Error Structure**:
```python
class ClaudeSDKError(Exception):
    """Base exception for all Claude SDK errors."""

class CLIConnectionError(ClaudeSDKError):
    """Raised when unable to connect to Claude Code."""

class CLINotFoundError(CLIConnectionError):
    """Raised when Claude Code is not found or not installed."""
    
    def __init__(self, message: str = "Claude Code not found", cli_path: str | None = None):
        if cli_path:
            message = f"{message}: {cli_path}"
        super().__init__(message)

class ProcessError(ClaudeSDKError):
    """Raised when the CLI process fails."""
    
    def __init__(self, message: str, exit_code: int | None = None, stderr: str | None = None):
        self.exit_code = exit_code
        self.stderr = stderr
        
        if exit_code is not None:
            message = f"{message} (exit code: {exit_code})"
        if stderr:
            message = f"{message}\nError output: {stderr}"
        
        super().__init__(message)

class CLIJSONDecodeError(ClaudeSDKError):
    """Raised when unable to decode JSON from CLI output."""
    
    def __init__(self, line: str, original_error: Exception):
        self.line = line
        self.original_error = original_error
        super().__init__(f"Failed to decode JSON: {line[:100]}...")

class MessageParseError(ClaudeSDKError):
    """Raised when unable to parse a message from CLI output."""
    
    def __init__(self, message: str, data: dict[str, Any] | None = None):
        self.data = data
        super().__init__(message)
```

## Go Error Implementation

**Error Interface and Types**:
```go
// Base error interface
type SDKError interface {
    error
    Type() string
}

// Base error struct
type BaseError struct {
    message string
    errType string
}

func (e *BaseError) Error() string { return e.message }
func (e *BaseError) Type() string  { return e.errType }

// Connection errors
type ConnectionError struct {
    *BaseError
    Cause error
}

func NewConnectionError(message string, cause error) *ConnectionError {
    return &ConnectionError{
        BaseError: &BaseError{message: message, errType: "connection"},
        Cause:     cause,
    }
}

func (e *ConnectionError) Unwrap() error { return e.Cause }

// CLI not found error with installation guidance
type CLINotFoundError struct {
    *BaseError
    Path string
}

func NewCLINotFoundError(message, path string) *CLINotFoundError {
    if path != "" {
        message = fmt.Sprintf("%s: %s", message, path)
    }
    return &CLINotFoundError{
        BaseError: &BaseError{message: message, errType: "cli_not_found"},
        Path:      path,
    }
}

// Process execution error
type ProcessError struct {
    *BaseError
    ExitCode int
    Stderr   string
}

func NewProcessError(message string, exitCode int, stderr string) *ProcessError {
    if exitCode != 0 {
        message = fmt.Sprintf("%s (exit code: %d)", message, exitCode)
    }
    if stderr != "" {
        message = fmt.Sprintf("%s\nError output: %s", message, stderr)
    }
    
    return &ProcessError{
        BaseError: &BaseError{message: message, errType: "process"},
        ExitCode:  exitCode,
        Stderr:    stderr,
    }
}

// JSON decode error
type JSONDecodeError struct {
    *BaseError
    Line          string
    Position      int
    OriginalError error
}

func NewJSONDecodeError(line string, originalError error) *JSONDecodeError {
    message := fmt.Sprintf("Failed to decode JSON: %s...", line[:min(len(line), 100)])
    return &JSONDecodeError{
        BaseError:     &BaseError{message: message, errType: "json_decode"},
        Line:          line,
        OriginalError: originalError,
    }
}

func (e *JSONDecodeError) Unwrap() error { return e.OriginalError }

// Message parse error
type MessageParseError struct {
    *BaseError
    Data interface{}
}

func NewMessageParseError(message string, data interface{}) *MessageParseError {
    return &MessageParseError{
        BaseError: &BaseError{message: message, errType: "message_parse"},
        Data:      data,
    }
}
```

## Error Context and Wrapping

**Python Approach**:
```python
# Error chaining with "from" keyword
except KeyError as e:
    raise MessageParseError(f"Missing required field: {e}", data) from e

# Context preservation in CLI errors  
except FileNotFoundError as e:
    if self._cwd and not Path(self._cwd).exists():
        raise CLIConnectionError(f"Working directory does not exist: {self._cwd}") from e
    raise CLINotFoundError(f"Claude Code not found at: {self._cli_path}") from e
```

**Go Error Wrapping**:
```go
// Use fmt.Errorf with %w verb for error wrapping
func (t *SubprocessTransport) Connect(ctx context.Context) error {
    if err := t.validateWorkingDir(); err != nil {
        return fmt.Errorf("working directory validation failed: %w", err)
    }
    
    if err := t.startProcess(ctx); err != nil {
        if t.cwd != "" {
            if _, statErr := os.Stat(t.cwd); os.IsNotExist(statErr) {
                return NewConnectionError(
                    fmt.Sprintf("working directory does not exist: %s", t.cwd),
                    err,
                )
            }
        }
        return NewCLINotFoundError("Claude Code not found", t.cliPath)
    }
    
    return nil
}

// Error checking with errors.Is and errors.As
func handleError(err error) {
    var cliErr *CLINotFoundError
    if errors.As(err, &cliErr) {
        log.Printf("CLI not found at path: %s", cliErr.Path)
        return
    }
    
    var procErr *ProcessError  
    if errors.As(err, &procErr) {
        log.Printf("Process failed with exit code: %d", procErr.ExitCode)
        log.Printf("Stderr: %s", procErr.Stderr)
        return
    }
    
    if errors.Is(err, context.Canceled) {
        log.Printf("Operation was canceled")
        return
    }
}
```

## Installation Guidance Patterns

**Python CLI Not Found Error**:
```python
def _find_cli(self) -> str:
    # ... search logic ...
    
    node_installed = shutil.which("node") is not None
    
    if not node_installed:
        error_msg = "Claude Code requires Node.js, which is not installed.\n\n"
        error_msg += "Install Node.js from: https://nodejs.org/\n"
        error_msg += "\nAfter installing Node.js, install Claude Code:\n"
        error_msg += "  npm install -g @anthropic-ai/claude-code"
        raise CLINotFoundError(error_msg)
    
    raise CLINotFoundError(
        "Claude Code not found. Install with:\n"
        "  npm install -g @anthropic-ai/claude-code\n"
        "\nIf already installed locally, try:\n"
        '  export PATH="$HOME/node_modules/.bin:$PATH"\n'
        "\nOr specify the path when creating transport:\n"
        "  SubprocessCLITransport(..., cli_path='/path/to/claude')"
    )
```

**Go Installation Guidance**:
```go
func findCLI() (string, error) {
    // Check PATH first
    if path, err := exec.LookPath("claude"); err == nil {
        return path, nil
    }
    
    // Check standard locations
    locations := []string{
        filepath.Join(os.Getenv("HOME"), ".npm-global/bin/claude"),
        "/usr/local/bin/claude",
        filepath.Join(os.Getenv("HOME"), ".local/bin/claude"),
        filepath.Join(os.Getenv("HOME"), "node_modules/.bin/claude"),
        filepath.Join(os.Getenv("HOME"), ".yarn/bin/claude"),
    }
    
    for _, location := range locations {
        if info, err := os.Stat(location); err == nil && !info.IsDir() {
            return location, nil
        }
    }
    
    // Check Node.js dependency
    if _, err := exec.LookPath("node"); err != nil {
        return "", NewCLINotFoundError(`Claude Code requires Node.js, which is not installed.

Install Node.js from: https://nodejs.org/

After installing Node.js, install Claude Code:
  npm install -g @anthropic-ai/claude-code`, "")
    }
    
    return "", NewCLINotFoundError(`Claude Code not found. Install with:
  npm install -g @anthropic-ai/claude-code

If already installed locally, try:
  export PATH="$HOME/node_modules/.bin:$PATH"

Or specify the path when creating client`, "")
}
```

## Error Recovery Patterns

**JSON Decode Error Recovery**:
```python
# Speculative parsing with graceful error handling
try:
    data = json.loads(json_buffer)
    json_buffer = ""
    yield data
except json.JSONDecodeError:
    # Continue accumulating until we get complete JSON
    continue
```

**Go Error Recovery**:
```go
func (t *SubprocessTransport) processJSONBuffer(line string) error {
    t.jsonBuffer.WriteString(line)
    
    // Check buffer size limit
    if t.jsonBuffer.Len() > MaxBufferSize {
        t.jsonBuffer.Reset()
        return NewJSONDecodeError("buffer size exceeded", 
            fmt.Errorf("buffer size %d exceeds limit %d", t.jsonBuffer.Len(), MaxBufferSize))
    }
    
    // Attempt to parse JSON
    var data map[string]interface{}
    if err := json.Unmarshal([]byte(t.jsonBuffer.String()), &data); err != nil {
        // Not complete JSON yet, continue accumulating
        return nil
    }
    
    // Successfully parsed, reset buffer
    t.jsonBuffer.Reset()
    return t.handleMessage(data)
}
```

## Sentinel Errors

**Go Sentinel Error Pattern**:
```go
var (
    ErrBufferOverflow = errors.New("buffer size exceeded")
    ErrCLINotFound    = errors.New("claude CLI not found")
    ErrConnectionLost = errors.New("connection lost")
    ErrInvalidMessage = errors.New("invalid message format")
)

// Usage with errors.Is()
if errors.Is(err, ErrBufferOverflow) {
    // Handle buffer overflow specifically
    return resetBuffer()
}
```

This comprehensive error system provides detailed context, helpful guidance, and proper Go error handling patterns while maintaining full compatibility with the Python SDK's error semantics.