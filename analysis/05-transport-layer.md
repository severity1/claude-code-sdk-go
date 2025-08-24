# Transport Layer Analysis

Analysis of the Python SDK's transport abstraction and subprocess implementation architecture.

## Transport Interface (_internal/transport/__init__.py)

**Abstract Interface Design**:
```python
class Transport(ABC):
    """Abstract transport for Claude communication.
    
    WARNING: This internal API is exposed for custom transport implementations
    (e.g., remote Claude Code connections). The Claude Code team may change or
    or remove this abstract class in any future release. Custom implementations
    must be updated to match interface changes.
    """
    
    @abstractmethod
    async def connect(self) -> None:
        """Initialize connection."""
        pass
    
    @abstractmethod
    async def disconnect(self) -> None:
        """Close connection."""
        pass
    
    @abstractmethod
    async def send_request(self, messages: list[dict[str, Any]], options: dict[str, Any]) -> None:
        """Send request to Claude."""
        pass
    
    @abstractmethod
    def receive_messages(self) -> AsyncIterator[dict[str, Any]]:
        """Receive messages from Claude."""
        pass
    
    @abstractmethod
    def is_connected(self) -> bool:
        """Check if transport is connected."""
        pass
```

**Key Design Decisions**:
- **Stability Warning**: Explicitly marked as internal API that may change
- **Raw dict protocol**: Messages are `dict[str, Any]`, not typed objects at transport level
- **Async/sync mix**: `receive_messages()` is sync returning AsyncIterator
- **Options flexibility**: `send_request` options are generic dict

## Go Transport Interface

**Go Interface Design**:
```go
// Transport provides an interface for different Claude Code communication methods.
// 
// Warning: This interface may change in future versions. Custom implementations
// should be prepared to update when the interface evolves.
type Transport interface {
    // Connect establishes the transport connection
    Connect(ctx context.Context) error
    
    // Disconnect closes the transport connection
    Disconnect() error
    
    // SendMessage sends a single message to Claude
    SendMessage(ctx context.Context, message StreamMessage) error
    
    // ReceiveMessages returns channels for receiving messages and errors
    ReceiveMessages(ctx context.Context) (<-chan map[string]interface{}, <-chan error)
    
    // Interrupt sends an interrupt signal (if supported)
    Interrupt(ctx context.Context) error
    
    // IsConnected reports whether the transport is connected
    IsConnected() bool
}
```

**Go Design Improvements**:
- **Context-aware**: All operations accept context for cancellation/timeouts
- **Channel-based**: Returns channels for Go-native streaming
- **Explicit errors**: Separate error channel for error handling
- **Type safety**: Use `StreamMessage` type instead of raw dict

## Subprocess Transport Overview

**Python Implementation Structure**:
```python
class SubprocessCLITransport(Transport):
    def __init__(self, prompt, options, cli_path=None, close_stdin_after_prompt=False):
        self._prompt = prompt
        self._is_streaming = not isinstance(prompt, str)
        self._options = options
        self._cli_path = str(cli_path) if cli_path else self._find_cli()
        self._cwd = str(options.cwd) if options.cwd else None
        self._process: Process | None = None
        self._stdout_stream: TextReceiveStream | None = None
        self._stderr_stream: TextReceiveStream | None = None
        self._stdin_stream: TextSendStream | None = None
        self._pending_control_responses: dict[str, dict[str, Any]] = {}
        self._request_counter = 0
        self._close_stdin_after_prompt = close_stdin_after_prompt
        self._task_group: anyio.abc.TaskGroup | None = None
        self._stderr_file: Any = None  # tempfile.NamedTemporaryFile
```

**Key Components**:
- **Streaming detection**: `_is_streaming = not isinstance(prompt, str)`
- **CLI discovery**: Automatic search through standard locations
- **Control protocol**: Request/response correlation for interrupts
- **Resource management**: Task groups, temp files, stream cleanup
- **stdin behavior**: Configurable closure based on usage pattern

## Go Subprocess Transport Structure

**Go Implementation Structure**:
```go
type SubprocessTransport struct {
    // Configuration
    prompt              interface{} // string or message stream
    options             *Options
    cliPath             string
    cwd                 string
    closeStdinAfterPrompt bool
    
    // Process management
    cmd                 *exec.Cmd
    stdin               io.WriteCloser
    stdout              io.ReadCloser
    stderrFile          *os.File
    
    // Control protocol
    controlResponses    map[string]chan controlResponse
    requestCounter      int64
    controlMutex        sync.RWMutex
    
    // Buffering
    jsonBuffer          strings.Builder
    maxBufferSize       int
    
    // State management
    connected           bool
    connectedMutex      sync.RWMutex
    
    // Channels for communication
    messagesChan        chan map[string]interface{}
    errorsChan          chan error
    doneChan            chan struct{}
}
```

## Transport Lifecycle Management

**Python Connection Process**:
```python
async def connect(self) -> None:
    """Start subprocess."""
    if self._process:
        return
    
    cmd = self._build_command()
    try:
        # Create temp file for stderr to avoid pipe deadlocks
        self._stderr_file = tempfile.NamedTemporaryFile(
            mode="w+", prefix="claude_stderr_", suffix=".log", delete=False
        )
        
        # Start process with pipes
        self._process = await anyio.open_process(
            cmd,
            stdin=PIPE,
            stdout=PIPE, 
            stderr=self._stderr_file,
            cwd=self._cwd,
            env={**os.environ, "CLAUDE_CODE_ENTRYPOINT": "sdk-py"},
        )
        
        # Setup streams
        if self._process.stdout:
            self._stdout_stream = TextReceiveStream(self._process.stdout)
        
        # Handle stdin based on mode
        if self._is_streaming:
            if self._process.stdin:
                self._stdin_stream = TextSendStream(self._process.stdin)
                # Start background streaming task
                self._task_group = anyio.create_task_group()
                await self._task_group.__aenter__()
                self._task_group.start_soon(self._stream_to_stdin)
        else:
            # Close stdin for string mode
            if self._process.stdin:
                await self._process.stdin.aclose()
```

**Go Connection Process**:
```go
func (t *SubprocessTransport) Connect(ctx context.Context) error {
    if t.IsConnected() {
        return nil
    }
    
    cmd, err := t.buildCommand()
    if err != nil {
        return err
    }
    
    // Create temporary file for stderr
    stderrFile, err := os.CreateTemp("", "claude_stderr_*.log")
    if err != nil {
        return fmt.Errorf("failed to create stderr temp file: %w", err)
    }
    
    // Setup command
    cmd.Stderr = stderrFile
    cmd.Dir = t.cwd
    cmd.Env = append(os.Environ(), "CLAUDE_CODE_ENTRYPOINT=sdk-go")
    
    // Create pipes
    stdin, err := cmd.StdinPipe()
    if err != nil {
        return fmt.Errorf("failed to create stdin pipe: %w", err)
    }
    
    stdout, err := cmd.StdoutPipe()
    if err != nil {
        return fmt.Errorf("failed to create stdout pipe: %w", err)
    }
    
    // Start process
    if err := cmd.Start(); err != nil {
        return t.handleStartError(err)
    }
    
    // Store references
    t.cmd = cmd
    t.stdin = stdin
    t.stdout = stdout
    t.stderrFile = stderrFile
    
    // Start I/O goroutines
    go t.handleStdout(ctx)
    go t.handleStdin(ctx)
    
    t.setConnected(true)
    return nil
}
```

## Command Building Process

**Python Command Construction**:
```python
def _build_command(self) -> list[str]:
    """Build CLI command with arguments."""
    cmd = [self._cli_path, "--output-format", "stream-json", "--verbose"]
    
    # Add all configuration options as CLI flags
    if self._options.system_prompt:
        cmd.extend(["--system-prompt", self._options.system_prompt])
    
    # ... all other options ...
    
    # Add extra args for future CLI flags
    for flag, value in self._options.extra_args.items():
        if value is None:
            cmd.append(f"--{flag}")  # Boolean flag
        else:
            cmd.extend([f"--{flag}", str(value)])  # Flag with value
    
    # Add prompt handling based on mode
    if self._is_streaming:
        cmd.extend(["--input-format", "stream-json"])
    else:
        cmd.extend(["--print", str(self._prompt)])
    
    return cmd
```

**Go Command Building**:
```go
func (t *SubprocessTransport) buildCommand() (*exec.Cmd, error) {
    args := []string{"--output-format", "stream-json", "--verbose"}
    
    // Add configuration options
    if t.options.SystemPrompt != "" {
        args = append(args, "--system-prompt", t.options.SystemPrompt)
    }
    
    if t.options.AppendSystemPrompt != "" {
        args = append(args, "--append-system-prompt", t.options.AppendSystemPrompt)
    }
    
    if len(t.options.AllowedTools) > 0 {
        args = append(args, "--allowedTools", strings.Join(t.options.AllowedTools, ","))
    }
    
    // ... all other options ...
    
    // Add extra arguments
    for flag, value := range t.options.ExtraArgs {
        if value == nil {
            args = append(args, "--"+flag)
        } else {
            args = append(args, "--"+flag, *value)
        }
    }
    
    // Add prompt handling
    isStreaming := t.isStreamingMode()
    if isStreaming {
        args = append(args, "--input-format", "stream-json")
    } else {
        if promptStr, ok := t.prompt.(string); ok {
            args = append(args, "--print", promptStr)
        }
    }
    
    return exec.Command(t.cliPath, args...), nil
}
```

## stdin Behavior Differentiation

**Critical Discovery - The Key Architectural Difference**:

The most important finding is that the transport behavior changes based on `close_stdin_after_prompt`:

**query() function (one-shot)**:
```python
# In _internal/client.py
chosen_transport = SubprocessCLITransport(
    prompt=prompt, options=options, close_stdin_after_prompt=True  # KEY!
)
```

**ClaudeSDKClient (bidirectional)**:
```python
# In client.py - defaults to False
self._transport = SubprocessCLITransport(
    prompt=_empty_stream() if prompt is None else prompt,
    options=self.options,
    # close_stdin_after_prompt=False (default)
)
```

**Go Implementation**:
```go
// For Query function
transport := NewSubprocessTransport(prompt, options, WithCloseStdinAfterPrompt(true))

// For Client 
transport := NewSubprocessTransport(prompt, options, WithCloseStdinAfterPrompt(false))

func (t *SubprocessTransport) handleStdin(ctx context.Context) {
    defer t.stdin.Close()
    
    if t.closeStdinAfterPrompt {
        // Send prompt then close stdin (Query mode)
        t.sendPromptAndClose(ctx)
    } else {
        // Keep stdin open for bidirectional communication (Client mode)
        t.keepStdinOpen(ctx)
    }
}
```

This architectural difference is fundamental to achieving the dual API pattern (query vs client) that the Python SDK implements.

## Transport Extensibility

**Plugin Architecture**:
The transport interface allows for alternative implementations:
- **Remote transports**: Connect to Claude Code over network
- **Mock transports**: For testing without subprocess
- **Cached transports**: Add response caching layer
- **Retry transports**: Add automatic retry logic

**Go Interface Benefits**:
- Type safety with interfaces
- Easy mocking for tests
- Clear separation of concerns
- Pluggable architecture for different deployment scenarios

This transport layer analysis provides the foundation for implementing the same flexible, extensible architecture in Go while maintaining all the sophisticated features of the Python implementation.