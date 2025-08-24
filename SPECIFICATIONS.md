# Claude Code SDK for Go - Technical Specifications

## Overview

The Claude Code SDK for Go provides programmatic access to Claude Code CLI with full feature parity to the Python SDK while embracing Go idioms and patterns. This SDK enables developers to build applications that interact with Claude Code for automation, tooling, and conversational AI workflows.

## Reference Implementation

This Go SDK is based on the Python reference implementation located at `../claude-code-sdk-python/`. The Python SDK serves as the canonical implementation for feature parity and behavioral compatibility.

## Design Principles

### 1. Go-Native Design
- **Context-first**: All operations accept `context.Context` for cancellation and timeouts
- **Explicit error handling**: No hidden exceptions, all errors returned explicitly
- **Channel-based concurrency**: Leverage goroutines and channels instead of async/await
- **Interface-driven**: Clean abstractions with minimal concrete dependencies
- **Functional options**: Extensible configuration using the functional options pattern

### 2. Performance-Oriented
- **Concurrent by design**: True parallelism without GIL limitations
- **Memory efficient**: Stack allocation, object pooling, and minimal heap pressure
- **Streaming optimized**: Efficient buffering and lazy parsing strategies
- **Resource cleanup**: Explicit lifecycle management with defer patterns

### 3. Complete Feature Parity
- **100% API coverage**: All Python SDK features available in Go
- **Transport abstraction**: Pluggable transport layer for future extensibility
- **Rich type system**: Comprehensive message and content block types
- **Advanced features**: Interrupts, MCP integration, permission modes, session management

## Core APIs

### 1. Query Function (One-Shot API)
Simple, stateless interactions for fire-and-forget queries with automatic stdin closure.

**Signatures:**
- `Query(ctx context.Context, prompt string, opts ...Option) (MessageIterator, error)`
- `QueryStream(ctx context.Context, messages <-chan StreamMessage, opts ...Option) (MessageIterator, error)`

**Features:**
- Automatic process cleanup after message stream completion
- Support for both string prompts and async message streams
- One-shot execution with stdin closure after prompt delivery
- Context-aware cancellation and timeouts

**Use Cases:**
- Simple one-off questions
- Batch processing of independent prompts  
- Automated scripts and CI/CD pipelines
- Code generation tasks

### 2. Client Interface (Streaming/Bidirectional API)
Stateful, interactive conversations with full control flow and persistent connections.

**Core Methods:**
- `NewClient(opts ...Option) *Client`
- `Connect(ctx context.Context, prompt ...StreamMessage) error`
- `Disconnect() error`
- `Query(ctx context.Context, prompt string, sessionID ...string) error`
- `QueryStream(ctx context.Context, messages <-chan StreamMessage) error`
- `ReceiveMessages(ctx context.Context) <-chan Message`
- `ReceiveResponse(ctx context.Context) MessageIterator`
- `Interrupt(ctx context.Context) error`

**Features:**
- Persistent connection with open stdin for bidirectional communication
- Session management with custom session IDs
- Context manager-style resource management
- Real-time message streaming with interrupt capability
- Graceful connection lifecycle management

**Use Cases:**
- Interactive chat interfaces
- Multi-turn conversations with context
- Real-time applications requiring interrupts
- Complex workflow automation

## Complete Type System

### Message Types
All message types implement the `Message` interface with `GetType() string`.

**UserMessage**
- `Content []ContentBlock` - Message content (can be string or content blocks)

**AssistantMessage**  
- `Content []ContentBlock` - Response content blocks
- `Model string` - Model used for response

**SystemMessage**
- `Subtype string` - System message subtype
- `Data map[string]interface{}` - Arbitrary system data

**ResultMessage**
- `Subtype string` - Result subtype identifier
- `DurationMS int` - Total conversation duration
- `DurationAPIMS int` - API-specific duration
- `IsError bool` - Whether conversation ended in error
- `NumTurns int` - Number of conversation turns
- `SessionID string` - Session identifier
- `TotalCostUSD *float64` - **CRITICAL**: Must use `total_cost_usd` in JSON tag, NOT `cost_usd` (changed in v0.0.13)
- `Usage map[string]interface{}` - Usage statistics (optional)
- `Result *string` - Final result string (optional)

### Content Block Types
All content blocks implement the `ContentBlock` interface with `GetType() string`.

**TextBlock**
- `Text string` - Plain text content

**ThinkingBlock**  
- `Thinking string` - Claude's internal reasoning
- `Signature string` - Thinking block signature

**ToolUseBlock**
- `ID string` - Unique tool use identifier
- `Name string` - Tool name
- `Input map[string]interface{}` - Tool input parameters

**ToolResultBlock**
- `ToolUseID string` - Reference to tool use ID
- `Content interface{}` - Tool execution result (string, list, or complex data)
- `IsError *bool` - Whether tool execution failed (optional)

### Configuration System (ClaudeCodeOptions)

**Tool Control**
- `AllowedTools []string` - Whitelist of permitted tools
- `DisallowedTools []string` - Blacklist of forbidden tools

**System Prompts & Model**
- `SystemPrompt string` - Primary system prompt
- `AppendSystemPrompt string` - Additional system prompt to append
- `Model string` - Specific Claude model to use
- `MaxThinkingTokens int` - Maximum thinking tokens (default: 8000)

**Permission & Safety System**
- `PermissionMode PermissionMode` - Tool execution permission handling
  - `PermissionModeDefault` - CLI prompts for dangerous tools
  - `PermissionModeAcceptEdits` - Auto-accept file edits
  - `PermissionModePlan` - Enable plan mode
  - `PermissionModeBypass` - Allow all tools without prompts (dangerous)
- `PermissionPromptToolName string` - Custom tool for permission prompts

**Session & State Management**
- `ContinueConversation bool` - Continue from previous conversation
- `Resume string` - Resume specific conversation by ID
- `MaxTurns int` - Maximum conversation turns limit
- `Settings string` - Settings file path or JSON string

**File System & Context**
- `Cwd string` - Working directory path
- `AddDirs []string` - **REQUIRED**: Additional directories for context (v0.0.19)
- `Settings string` - **REQUIRED**: Settings file path or JSON string (v0.0.18)

**MCP (Model Context Protocol) Integration**
- `McpServers map[string]McpServerConfig` - MCP server configurations

**MCP Server Configuration Types:**

**McpStdioServerConfig**
- `Type McpServerType` - "stdio"
- `Command string` - Command to execute
- `Args []string` - Command arguments
- `Env map[string]string` - Environment variables

**McpSSEServerConfig**
- `Type McpServerType` - "sse" 
- `URL string` - Server-Sent Events URL
- `Headers map[string]string` - HTTP headers

**McpHttpServerConfig**
- `Type McpServerType` - "http"
- `URL string` - HTTP endpoint URL
- `Headers map[string]string` - HTTP headers

**Extensibility**
- `ExtraArgs map[string]*string` - Arbitrary CLI flags
  - `nil` value = boolean flag without value
  - `string` value = flag with value

### Stream Message Protocol
For streaming communication between SDK and CLI.

**StreamMessage**
- `Type string` - Message type ("user", "control_request")
- `Message interface{}` - Message payload
- `ParentToolUseID *string` - Parent tool use reference (optional)
- `SessionID string` - Session identifier
- `RequestID string` - Unique request identifier (for control messages)
- `Request map[string]interface{}` - Control request data
- `Response map[string]interface{}` - Control response data

## Complete Error Handling System

### Error Type Hierarchy
All SDK errors implement the `SDKError` interface extending standard `error`.

**Base Interface**
- `SDKError` - Base interface with `Type() string` method

**Connection Errors**
- `ConnectionError` - Base for all connection-related failures
  - `Message string` - Descriptive error message  
  - `Cause error` - Underlying cause error
- `CLINotFoundError` - Claude CLI not found or inaccessible
  - `Path string` - Attempted CLI path
  - `Message string` - Detailed error with installation instructions
  - Includes Node.js dependency checking and helpful installation messages

**Process Errors**  
- `ProcessError` - Subprocess execution failures
  - `ExitCode int` - Process exit code
  - `Stderr string` - Process stderr output
  - `Message string` - Formatted error message

**Protocol Errors**
- `JSONDecodeError` - JSON parsing failures from CLI output
  - `Line string` - Problematic JSON line
  - `Position int` - Character position of error
  - `Cause error` - Original JSON parsing error
- `MessageParseError` - Message structure parsing failures
  - `Data interface{}` - Raw message data that failed to parse
  - `Message string` - Descriptive error message

### Error Handling Patterns
- **Wrapped errors**: Use `errors.Is()` and `errors.As()` for error inspection
- **Context cancellation**: Proper handling of `context.Canceled` and `context.DeadlineExceeded`
- **Resource cleanup**: Automatic cleanup on error paths using defer
- **Error propagation**: Clear error messages with contextual information and suggestions

## Transport Layer Architecture

### Transport Interface
Abstract interface enabling pluggable transport implementations.

**Core Methods:**
- `Connect(ctx context.Context) error` - Establish connection
- `Disconnect() error` - Close connection and cleanup
- `SendMessage(ctx context.Context, message StreamMessage) error` - Send message to CLI
- `ReceiveMessages(ctx context.Context) (<-chan Message, <-chan error)` - Receive message stream
- `Interrupt(ctx context.Context) error` - Send interrupt signal
- `IsConnected() bool` - Connection status check

**Critical Implementation Parameters:**
- `CloseStdinAfterPrompt bool` - **CRITICAL**: Differentiates Query vs Client behavior
  - `true` for Query function (one-shot mode with stdin closure)
  - `false` for Client (bidirectional streaming mode)
  - This parameter is the fundamental architectural difference between APIs

### Subprocess Transport Implementation

**CLI Discovery System**
Automatic discovery following Python SDK search pattern:
1. `exec.LookPath("claude")` - System PATH lookup
2. `~/.npm-global/bin/claude` - Global npm installation
3. `/usr/local/bin/claude` - System-wide installation
4. `~/.local/bin/claude` - User local installation
5. `~/node_modules/.bin/claude` - Local project installation
6. `~/.yarn/bin/claude` - Yarn global installation

**Error Handling:**
- Node.js dependency validation
- Helpful installation instructions in error messages
- Working directory existence validation

**Process Management Features**
- **Lifecycle Management**: Proper startup, shutdown, and cleanup
- **I/O Handling**: Concurrent stdin/stdout/stderr management
- **5-Second Termination Sequence**: **EXACT IMPLEMENTATION REQUIRED**
  ```go
  // Send SIGTERM
  cmd.Process.Signal(os.Interrupt)
  
  // Wait exactly 5 seconds
  select {
  case <-done:
    return nil
  case <-time.After(5 * time.Second):
    return cmd.Process.Kill() // Force SIGKILL
  }
  ```
- **Resource Monitoring**: Track process state and resource usage

**Command Construction**
Dynamic CLI argument building supporting all flags:
- `--output-format stream-json --verbose`
- `--input-format stream-json` (for streaming mode)
- `--print <prompt>` (for one-shot mode)
- `--system-prompt`, `--append-system-prompt`
- `--allowedTools`, `--disallowedTools` (comma-separated)
- `--max-turns`, `--model`, `--permission-mode`
- `--continue`, `--resume`, `--settings`
- `--add-dir` (multiple, v0.0.19) - **REQUIRED**: `for _, dir := range options.AddDirs { args = append(args, "--add-dir", dir) }`
- `--settings` (v0.0.18) - **REQUIRED**: Settings file path or JSON string
- `--mcp-config` (JSON string or file path)
- Custom flags via `ExtraArgs`

**Streaming & Buffering System**
- **Speculative JSON Parsing**: **CRITICAL** - Continue accumulation on `json.Unmarshal` errors instead of failing
  ```go
  if err := json.Unmarshal([]byte(buffer), &data); err != nil {
    return nil // Continue accumulating, NOT an error condition
  }
  ```
- **Multi-JSON Line Handling**: Split lines by newlines: `strings.Split(line, "\n")`
- **Buffer Size Protection**: 1MB maximum (`const MaxBufferSize = 1024 * 1024`)
- **Buffer Overflow**: Reset buffer and return error when limit exceeded
- **Embedded Newlines**: Handle newlines within JSON string values correctly
- **Stderr Handling**: Capture stderr using temporary files to avoid pipe deadlocks
- **Line Processing**: Process JSON messages line-by-line with proper error recovery

**Control Protocol Implementation**
Advanced interrupt and control message handling:
- **Request ID Generation**: `fmt.Sprintf("req_%d_%s", atomic.AddInt64(&counter, 1), hex.EncodeToString(rand.Bytes(4)))`
- **Response Polling**: Exactly 100ms polling interval (`time.Sleep(100 * time.Millisecond)`)
- **Response Correlation**: Map pending requests by request ID: `map[string]chan controlResponse`
- **Control Message Format**:
  ```go
  ControlRequest{
    Type: "control_request",
    RequestID: requestID,
    Request: map[string]interface{}{"action": "interrupt"},
  }
  ```
- **Timeout Management**: Proper cleanup of pending requests on context cancellation

## Advanced Features

### Session Management
- **Session Continuity**: Resume conversations with specific session IDs
- **Multi-Session Support**: Handle multiple concurrent conversation sessions
- **State Persistence**: Optional conversation state saving and loading
- **Session Isolation**: Separate conversation contexts

### Interrupt & Control System
- **Graceful Interruption**: Clean cancellation of long-running operations
- **Control Protocol**: Send control requests (interrupt, status, configuration)
- **Context Propagation**: Proper context cancellation throughout stack
- **Resource Cleanup**: Ensure no resource leaks on interruption

### MCP Integration
- **Multi-Protocol Support**: stdio, Server-Sent Events, and HTTP protocols  
- **Dynamic Configuration**: Runtime MCP server management and configuration
- **Protocol Handling**: Proper MCP message routing and processing
- **Connection Management**: Handle MCP server lifecycle
- **Stderr Isolation**: **CRITICAL** - MCP servers logging to stderr can cause SDK hanging
  - Must isolate MCP server stderr to prevent subprocess deadlocks
  - Use separate stderr handling for MCP-enabled processes

### Permission System
- **Granular Control**: Fine-grained permission modes for tool execution
- **Interactive Prompts**: User confirmation for sensitive operations
- **Tool Filtering**: Sophisticated allow/deny lists with pattern support
- **Safety Defaults**: Secure-by-default permission settings

## Performance & Optimization

### Memory Management
- **Object Pooling**: Reuse frequently allocated message and content block objects
- **Streaming Processing**: Incremental message processing without buffering entire responses
- **Lazy Parsing**: Parse JSON content only when accessed using `json.RawMessage`
- **Buffer Management**: Configurable buffer sizes with safety limits

### I/O Optimization
- **Concurrent Processing**: Parallel stdin/stdout/stderr handling
- **Pipeline Design**: Non-blocking I/O operations with backpressure handling
- **Resource Limits**: Configurable timeouts and buffer size limits
- **Efficient Scanning**: Optimized JSON line scanning with custom buffers

### JSON Processing
- **Efficient Unmarshaling**: Type-specific decoders for known message formats
- **Streaming JSON**: Process JSON streams without loading entire content
- **Dynamic Parsing**: Handle dynamic content with interface{} and delayed parsing
- **Error Recovery**: Robust error handling for malformed JSON

### Concurrency Model
- **Goroutine Per Client**: Dedicated goroutines for each client's I/O operations
- **Channel Communication**: Buffered channels for message passing with appropriate sizing
- **Synchronization**: Minimal locking with atomic operations where possible
- **Graceful Shutdown**: Coordinated shutdown using sync.WaitGroup and context cancellation

## Environment Integration

### Environment Variables
- **`CLAUDE_CODE_ENTRYPOINT`** - SDK identification for Claude CLI
  - `"sdk-go"` for Query function usage
  - `"sdk-go-client"` for Client usage
- **Standard Environment**: Inherit and pass through environment variables to subprocess

### CLI Integration
- **Version Compatibility**: Runtime version detection and feature availability checks
- **Cross-Platform**: Support for Windows, macOS, and Linux
- **Path Resolution**: Robust path handling and working directory management
- **Configuration Files**: Support for Claude Code configuration files and settings

## Testing Strategy

### Test Coverage Areas
- **Unit Tests**: All message types, content blocks, configuration options, and error conditions
- **Integration Tests**: Real CLI subprocess interaction with mocked responses
- **Performance Tests**: Memory usage, CPU efficiency, concurrent load testing
- **Edge Case Tests**: Malformed JSON, process failures, network timeouts, buffer overflows

### Mock Framework
- **Transport Mocking**: In-memory transport implementation for fast unit tests  
- **CLI Simulation**: Mock subprocess behavior for testing error conditions and edge cases
- **Message Generators**: Utilities for generating test messages and responses

### Benchmarking
- **Performance Baselines**: Establish performance benchmarks against Python SDK
- **Memory Profiling**: Track memory usage patterns and identify leaks
- **Concurrency Testing**: Validate performance under high concurrent load
- **Regression Prevention**: Automated performance regression detection

## Security Considerations

### Input Validation & Sanitization
- **Command Injection Prevention**: Strict validation and escaping of all CLI arguments
- **Path Traversal Protection**: Sanitize file paths and working directory specifications
- **Environment Isolation**: Controlled environment variable passing to subprocess
- **JSON Validation**: Validate all JSON structures before processing

### Process Security
- **Subprocess Sandboxing**: Run Claude CLI with minimal required permissions
- **Resource Limits**: Prevent resource exhaustion and denial-of-service attacks
- **Clean Shutdown**: Ensure proper cleanup of sensitive data in memory
- **Error Information**: Avoid exposing sensitive data in error messages and logs

### Data Handling
- **Sensitive Data**: Proper handling of API keys, tokens, and user data
- **Memory Management**: Clear sensitive data from memory when no longer needed
- **Logging**: Avoid logging sensitive information in debug or error logs

## Dependencies & Requirements

### Standard Library Only
Zero external dependencies, using only Go's standard library:
- `context` - Context management and cancellation
- `encoding/json` - JSON processing and streaming  
- `os/exec` - Subprocess management and execution
- `bufio` - Buffered I/O operations and scanning
- `sync` - Concurrency primitives and synchronization
- `errors` - Error handling and wrapping utilities
- `io` - I/O interfaces and utilities
- `path/filepath` - Cross-platform path manipulation
- `os` - Operating system interface
- `time` - Time operations and timeouts

### Go Version Requirements
- **Minimum**: Go 1.21 (for enhanced error handling, improved performance)
- **Recommended**: Go 1.22+ (for optimal performance, latest tooling support)

### Claude CLI Compatibility
- **Target Version**: Latest Claude Code CLI with full feature support
- **Version Detection**: Runtime compatibility checking and feature detection
- **Backward Compatibility**: Support for older CLI versions where feasible
- **Feature Negotiation**: Graceful degradation for unsupported features

## Project Structure & Organization

### Package Layout
Following Go SDK best practices with single root package and internal implementation isolation:

```
github.com/severity1/claude-code-sdk-go/
├── doc.go                # Package documentation and overview
├── types.go              # Core types and interfaces (Message, ContentBlock, Transport)
├── client.go             # Client implementation for bidirectional streaming
├── query.go              # Query function for one-shot requests
├── errors.go             # SDK error types and hierarchy
├── options.go            # Configuration with functional options pattern
├── examples/             # Usage examples and tutorials
│   ├── quickstart/       # Basic usage examples
│   ├── streaming/        # Advanced streaming examples  
│   ├── tools/            # Tool usage examples
│   └── advanced/         # Complex integration examples
├── internal/             # Internal implementation (not importable externally)
│   ├── subprocess/       # Subprocess transport implementation
│   ├── parser/           # JSON message parsing logic
│   ├── cli/              # CLI discovery and validation
│   └── protocol/         # Control request/response handling
└── testdata/             # Test data and fixtures
```

### Package Design Principles

**Single Root Package**: Following modern Go SDK patterns (AWS SDK v2, Google Cloud), all public APIs are exposed from the root `claudecode` package. This provides a clean import experience and follows the principle that packages should have a single, well-defined purpose.

**Internal Package Isolation**: All implementation details reside in `internal/` packages, preventing external dependencies on unstable internal APIs and allowing for refactoring without breaking changes.

**Functional Options Pattern**: Configuration follows the functional options pattern with `func(*Options)` signatures, providing extensible and readable API configuration:
```go
client := NewClient(
    WithSystemPrompt("You are a helpful assistant"),
    WithAllowedTools("Read", "Write"),
    WithPermissionMode(PermissionModeAcceptEdits),
)
```

**Interface-Driven Design**: Core abstractions (`Transport`, `Message`, `ContentBlock`) are defined as interfaces to enable testing, mocking, and future extensibility.

### Module and Package Naming

**Module Path**: `github.com/severity1/claude-code-sdk-go`
**Package Name**: `claudecode` (short, clear, single word following Go conventions)
**Import Statement**: `import "github.com/severity1/claude-code-sdk-go"`

### Documentation Strategy
- **Package Documentation**: Comprehensive `doc.go` with overview, examples, and usage patterns
- **Godoc Comments**: Full API documentation with runnable examples
- **Usage Guides**: Step-by-step tutorials for common use cases in `/examples`
- **Migration Guide**: Help Python SDK users transition to Go
- **Performance Guide**: Optimization tips and best practices
- **Architecture Guide**: Internal design and extension points

## Quality Assurance

### Code Quality Standards
- **Go Standards**: Strict adherence to Go formatting (gofmt) and conventions
- **Linting**: Comprehensive golangci-lint configuration with security checks
- **Documentation**: Complete godoc coverage for all public APIs
- **Error Messages**: Clear, actionable error messages with suggested solutions

### Testing Requirements  
- **Coverage**: Minimum 90% test coverage for all core functionality
- **Integration**: Real CLI interaction validation with various configurations
- **Benchmarks**: Performance regression prevention with automated benchmarking
- **Race Detection**: Concurrent access validation with race detector
- **Fuzzing**: Input validation through fuzzing of JSON parsing and message handling

### Release & Versioning
- **Semantic Versioning**: Proper version management following semver principles
- **Compatibility Promise**: API stability guarantees and deprecation policies  
- **Cross-Platform Testing**: Validation across Windows, macOS, and Linux
- **Performance Tracking**: Release-to-release performance comparison
- **Documentation Updates**: Keep all documentation current with each release

This specification ensures 100% feature parity with the Python SDK while providing a Go-native, high-performance, and secure implementation that meets the needs of Go developers building Claude Code integrations.