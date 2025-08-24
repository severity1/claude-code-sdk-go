# Feature Matrix - 100% Parity Checklist

Complete feature coverage matrix for Go SDK implementation to achieve 100% parity with Python SDK.

## ✅ Core Architecture (100% Required)

### API Surface
- [ ] `Query()` function for one-shot queries
- [ ] `Client` interface for bidirectional streaming
- [ ] `Transport` interface for pluggable backends
- [ ] Context-aware operations throughout
- [ ] Functional options pattern

### Environment Integration
- [ ] `CLAUDE_CODE_ENTRYPOINT` environment variable handling
  - [ ] `"sdk-go"` for Query function
  - [ ] `"sdk-go-client"` for Client usage
- [ ] Working directory validation
- [ ] Path resolution and sanitization

## ✅ Type System (100% Required)

### Message Types
- [ ] `UserMessage` with mixed content support
- [ ] `AssistantMessage` with content blocks and model field
- [ ] `SystemMessage` with subtype and raw data preservation
- [ ] `ResultMessage` with complete cost and usage information

### Content Block Types  
- [ ] `TextBlock` for plain text content
- [ ] `ThinkingBlock` with thinking and signature fields
- [ ] `ToolUseBlock` with ID, name, and input parameters
- [ ] `ToolResultBlock` with tool use ID, content, and error flag

### Configuration Types
- [ ] `Options` struct with all Python SDK fields
- [ ] `PermissionMode` enum (default, acceptEdits, plan, bypassPermissions)
- [ ] `McpServerConfig` with stdio/SSE/HTTP variants
- [ ] Functional options for builder pattern

## ✅ Error System (100% Required)

### Error Hierarchy
- [ ] `SDKError` base interface
- [ ] `ConnectionError` for connection failures
- [ ] `CLINotFoundError` with installation guidance
- [ ] `ProcessError` with exit code and stderr
- [ ] `JSONDecodeError` with line and position info
- [ ] `MessageParseError` with raw data context

### Error Context
- [ ] Detailed error messages with context
- [ ] Installation instructions in CLI not found errors
- [ ] Working directory validation errors
- [ ] Process termination error handling

## ✅ CLI Integration (100% Required)

### CLI Discovery
- [ ] `exec.LookPath("claude")` - System PATH lookup
- [ ] `~/.npm-global/bin/claude` - Global npm installation
- [ ] `/usr/local/bin/claude` - System-wide installation
- [ ] `~/.local/bin/claude` - User local installation
- [ ] `~/node_modules/.bin/claude` - Local project installation
- [ ] `~/.yarn/bin/claude` - Yarn global installation

### Command Building
- [ ] Base arguments: `--output-format stream-json --verbose`
- [ ] Streaming mode: `--input-format stream-json`
- [ ] String mode: `--print <prompt>`
- [ ] All configuration options as CLI flags
- [ ] `ExtraArgs` support for arbitrary flags

### Node.js Validation
- [ ] Check for Node.js availability
- [ ] Helpful error messages for missing Node.js
- [ ] Installation guidance in error messages

## ✅ Transport Layer (100% Required)

### Subprocess Management
- [ ] Process lifecycle management
- [ ] Stdin/stdout/stderr handling with goroutines
- [ ] 5-second SIGTERM → SIGKILL termination sequence
- [ ] Working directory validation before start
- [ ] Environment variable passing

### Streaming Protocol
- [ ] `close_stdin_after_prompt` parameter differentiation
- [ ] JSON message streaming to stdin
- [ ] Line-by-line stdout processing
- [ ] Stderr capture via temporary files

### Buffer Management  
- [ ] 1MB buffer size protection (`_MAX_BUFFER_SIZE`)
- [ ] Speculative JSON parsing strategy
- [ ] Multi-JSON object per line handling
- [ ] Embedded newlines in JSON strings support

## ✅ Message Processing (100% Required)

### JSON Parsing
- [ ] Speculative parsing - accumulate until complete JSON
- [ ] Buffer overflow protection with graceful error
- [ ] Multi-line JSON message support
- [ ] Embedded newlines in string values

### Message Parsing
- [ ] Type discrimination by `"type"` field
- [ ] Exhaustive error handling for all message types
- [ ] Optional field handling with pointer types
- [ ] Mixed content block parsing

### Content Block Processing
- [ ] All content block types in UserMessage and AssistantMessage
- [ ] Tool use and tool result block support
- [ ] Raw data preservation for extensibility

## ✅ Advanced Features (100% Required)

### Interrupt System
- [ ] Control request/response protocol
- [ ] Unique request ID generation
- [ ] 100ms polling interval for responses
- [ ] Active consumption requirement for interrupts

### Session Management
- [ ] Session ID support with defaults
- [ ] Conversation continuity
- [ ] Resume capability
- [ ] Multi-turn conversation state

### MCP Integration
- [ ] Three server types: stdio, SSE, HTTP
- [ ] Dynamic server configuration
- [ ] JSON serialization of server configs
- [ ] Stderr isolation to prevent hanging

## ✅ Usage Patterns (100% Required)

### Client Patterns
- [ ] Context manager equivalent (defer cleanup)
- [ ] Persistent client with manual lifecycle
- [ ] Message collection utilities
- [ ] Iterator to slice conversion helpers

### Integration Patterns
- [ ] Timeout handling with `context.WithTimeout()`
- [ ] Error recovery and retry strategies
- [ ] Message display helper functions
- [ ] Model specification in options

## ✅ Edge Cases (100% Required)

### Buffering Edge Cases
- [ ] Multiple JSON objects on single line
- [ ] Embedded newlines in JSON string values
- [ ] Buffer overflow protection testing
- [ ] Speculative parsing error recovery

### Process Edge Cases
- [ ] Process termination during streaming
- [ ] Stderr handling without deadlocks
- [ ] Working directory non-existence
- [ ] CLI binary not found scenarios

### Network Edge Cases
- [ ] Connection interruption handling
- [ ] Timeout scenarios
- [ ] Resource cleanup on errors
- [ ] Graceful degradation

## ✅ Testing Requirements (100% Required)

### Unit Tests
- [ ] All message type parsing
- [ ] All error conditions
- [ ] Configuration validation
- [ ] CLI command building

### Integration Tests
- [ ] Real subprocess communication
- [ ] End-to-end conversation flows
- [ ] Interrupt scenarios
- [ ] Edge case handling

### Mock Framework
- [ ] Transport mocking for unit tests
- [ ] Process state mocking
- [ ] Error condition simulation
- [ ] Performance benchmarking

## ✅ Performance Requirements (100% Required)

### Memory Management
- [ ] Object pooling for frequent allocations
- [ ] Streaming processing without buffering entire responses
- [ ] Efficient JSON parsing with `json.RawMessage`
- [ ] Resource cleanup on all code paths

### Concurrency
- [ ] Goroutine per client for I/O operations
- [ ] Channel-based message passing
- [ ] Context cancellation throughout
- [ ] Race condition testing

## Implementation Priority

**Phase 1 - Foundation**: Type system, error handling, basic CLI integration
**Phase 2 - Core Features**: Transport layer, message parsing, subprocess management  
**Phase 3 - Advanced Features**: Interrupts, MCP integration, session management
**Phase 4 - Polish**: Edge cases, performance optimization, comprehensive testing

**Success Criteria**: All checkboxes completed with comprehensive test coverage and performance benchmarks.