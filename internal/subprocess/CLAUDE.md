# Process Management Context

**Context**: Claude CLI subprocess management with 5-second termination, I/O handling, and process lifecycle for Phase 3 implementation

## Component Focus
- **Process Lifecycle** - Start, manage, and terminate Claude CLI subprocess safely
- **5-Second Termination** - SIGTERM → wait 5s → SIGKILL sequence (critical for resource cleanup)
- **I/O Management** - Stdin/stdout/stderr handling with concurrent goroutines
- **Environment Setup** - `CLAUDE_CODE_ENTRYPOINT` and working directory configuration

## Critical Process Management Patterns

### 5-Second Termination Sequence (Essential)
```go
func (t *Transport) terminate() error {
    // Send SIGTERM first
    if err := t.cmd.Process.Signal(syscall.SIGTERM); err != nil {
        return err
    }
    
    // Wait exactly 5 seconds
    done := make(chan error, 1)
    go func() {
        done <- t.cmd.Wait()
    }()
    
    select {
    case err := <-done:
        return err
    case <-time.After(5 * time.Second):
        // Force kill after 5 seconds
        return t.cmd.Process.Kill()
    }
}
```

### Environment Variable Setup
- **CLAUDE_CODE_ENTRYPOINT**: Must be set to `"sdk-go"` or `"sdk-go-client"`
- **Working Directory**: Validate and set proper working directory
- **PATH Inheritance**: Ensure subprocess inherits proper PATH for Claude CLI discovery

### Concurrent I/O Handling
```go
// Pattern: Separate goroutines for stdin/stdout/stderr
func (t *Transport) startIOHandling(ctx context.Context) {
    // Stdout reading goroutine
    go t.readStdout(ctx)
    
    // Stderr isolation (prevent deadlocks)
    go t.handleStderr()
    
    // Stdin writing managed by caller
}
```

## Process Lifecycle States
- **Disconnected** - No process running
- **Connecting** - Process starting up
- **Connected** - Ready for I/O operations  
- **Terminating** - Shutdown in progress
- **Error** - Process failed or crashed

## Stdin/Stdout/Stderr Management

### Stdin (JSON Messages to CLI)
- Send `StreamMessage` objects as JSON lines
- Handle backpressure when CLI isn't consuming
- Close stdin for one-shot operations (`Query` mode)
- Keep open for streaming operations (`Client` mode)

### Stdout (JSON Responses from CLI)
- Read streaming JSON messages
- Pass to parser for message discrimination
- Handle partial reads and buffering
- Forward parsed messages via channels

### Stderr Isolation
```go
// Critical: Isolate stderr to prevent deadlocks
func (t *Transport) handleStderr() {
    // Use temporary files or separate goroutine
    // Never block main I/O flow
}
```

## Command Building Patterns
- **CLI Path Discovery** - Use discovery.go patterns
- **Argument Construction** - Build flags from Options struct
- **Stream vs One-shot** - Different flag patterns:
  - One-shot: `--print --output-format stream-json`
  - Streaming: `--input-format stream-json --output-format stream-json`

## Resource Cleanup Requirements
```go
// Pattern: Comprehensive cleanup
func (t *Transport) Close() error {
    // 1. Cancel context to stop goroutines
    t.cancel()
    
    // 2. Close stdin if open
    if t.stdin != nil {
        t.stdin.Close()
    }
    
    // 3. Terminate process with 5-second sequence
    return t.terminate()
}
```

## Error Handling
- **ProcessError** - Include exit code and stderr output
- **Connection Failures** - Retry logic for transient failures
- **I/O Errors** - Distinguish between recoverable and fatal errors
- **Context Cancellation** - Respect context timeouts and cancellation

## Concurrency Safety
- **Process State** - Protect with mutex for concurrent access
- **I/O Operations** - Use channels for goroutine communication
- **Resource Cleanup** - Ensure cleanup happens exactly once

## Platform Compatibility
- **Windows** - Handle process management differences
- **macOS/Linux** - POSIX signal handling
- **Process Groups** - Proper child process cleanup

## Integration Requirements
- Must implement Transport interface from parent internal/ context
- Work with CLI discovery patterns from cli/ component
- Integrate with message parsing from parser/ component
- Support both Query (one-shot) and Client (streaming) modes

## Performance Considerations
- **Goroutine Efficiency** - Minimize goroutine overhead
- **Memory Management** - Prevent memory leaks from long-running processes
- **I/O Buffering** - Efficient buffering strategies for high-throughput scenarios
- **Process Reuse** - Consider process pooling for frequent operations