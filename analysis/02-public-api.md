# Public API Analysis

Analysis of the Python SDK's public interface, exports, and API design principles.

## Public API Structure (__init__.py)

The Python SDK exports a carefully curated public interface through `__all__`:

```python
# Main Functions
query                    # One-shot query function

# Client Classes  
ClaudeSDKClient         # Interactive streaming client
Transport               # Abstract transport base class

# Type System
Message                 # Union of all message types
ContentBlock           # Union of all content block types  
ClaudeCodeOptions      # Configuration options
PermissionMode         # Permission mode literal type
McpServerConfig        # MCP server configuration

# Message Types
UserMessage, AssistantMessage, SystemMessage, ResultMessage

# Content Block Types  
TextBlock, ThinkingBlock, ToolUseBlock, ToolResultBlock

# Error Hierarchy
ClaudeSDKError         # Base error
CLIConnectionError     # Connection issues
CLINotFoundError       # CLI not found
ProcessError           # Process execution failed
CLIJSONDecodeError     # JSON parsing failed
```

## API Design Principles

**Explicit Public Interface**: 
- `__all__` defines exactly what's exported
- No accidental exposure of internal implementation
- Clear contract for what users can depend on

**Logical Grouping**:
- Main exports (functions and client)
- Transport abstraction for extensibility
- Complete type system for type safety
- Comprehensive error hierarchy

**Transport Abstraction**: 
- `Transport` base class exposed for custom implementations
- Marked with stability warnings in internal documentation
- Allows pluggable backends for different deployment scenarios

**Version Management**: 
- `__version__` constant matches pyproject.toml
- Single source of truth for version information
- Automated version update scripts maintain consistency

## Import Strategy

**Private Module Convention**:
- `_errors`, `_internal` - underscore prefix indicates private
- Internal implementation hidden from public API
- Clear boundary between public and private code

**Selective Imports**:
- Only specific classes/functions imported, not entire modules
- Prevents namespace pollution
- Explicit control over public surface

**Type Safety Support**:
- All types explicitly listed for IDE/type checker support
- Union types properly exported for type checking
- Complete type coverage for user code

## Documentation Patterns (README.md)

**Installation Requirements:**
- **Python**: 3.10+
- **Node.js**: Required dependency (Claude Code is built on Node.js)
- **Claude Code CLI**: `npm install -g @anthropic-ai/claude-code`
- **PyPI Package**: `pip install claude-code-sdk`

**API Patterns Demonstrated:**
1. **Simple Query Pattern**: `query(prompt="string")` - basic one-shot usage
2. **Options Pattern**: `ClaudeCodeOptions` for configuration
3. **Tool Usage**: Shows permission modes and tool filtering
4. **Working Directory**: Path/string support for `cwd` parameter

**Important API Details:**
- **Return Type**: `AsyncIterator[Message]` - streaming by design
- **Message Processing**: Type checking with `isinstance()` for different message types
- **Content Block Access**: Nested iteration through `message.content` for blocks
- **Permission Modes**: Shows `'acceptEdits'` for auto-accepting file operations

**Error Handling Hierarchy:**
```python
ClaudeSDKError (base)
├── CLINotFoundError (Claude Code not installed)
├── CLIConnectionError (connection issues) 
├── ProcessError (process execution failed)
└── CLIJSONDecodeError (JSON parsing issues)
```

**Documentation Quality:**
- Links to official Anthropic docs
- Clear prerequisite requirements
- Progressive complexity (simple → tools → directory)
- Type-aware examples with isinstance checks

## Go SDK Public API Design

Based on the Python SDK analysis, the Go SDK should follow these patterns:

### Package Structure
```go
package claudecode

// Main functions
func Query(ctx context.Context, prompt string, opts ...Option) (MessageIterator, error)

// Client interface
type Client interface {
    Stream(ctx context.Context, prompt string) (<-chan Message, error)
    Send(ctx context.Context, message StreamMessage) error
    Interrupt(ctx context.Context) error
    Close() error
}

// Transport interface (public for extensibility)
type Transport interface {
    Connect(ctx context.Context) error
    SendMessage(ctx context.Context, message StreamMessage) error
    ReceiveMessages(ctx context.Context) (<-chan Message, <-chan error)
    Interrupt(ctx context.Context) error
    Close() error
}
```

### Logical Grouping
- **Core Functions**: Query, NewClient
- **Interfaces**: Client, Transport, Message, ContentBlock
- **Types**: All message and content block concrete types
- **Options**: Configuration types and functional options
- **Errors**: Complete error hierarchy

### Go-Native Patterns
- **Single Root Package**: All public APIs in `claudecode` package following AWS SDK v2/Google Cloud patterns
- **Context-first**: All operations accept `context.Context`
- **Interface-driven**: Use interfaces for extensibility with implementations in `internal/`
- **Explicit errors**: Return errors, don't panic
- **Functional options**: `opts ...Option` pattern for configuration
- **Channel-based streaming**: Use Go channels instead of async iterators

**Import Pattern:**
```go
import "github.com/severity1/claude-code-sdk-go"

// Usage
client := claudecode.NewClient(
    claudecode.WithSystemPrompt("You are a helpful assistant"),
)
```

### Version Management
```go
// Version constant
const Version = "1.0.0"

// Version info structure
type VersionInfo struct {
    Version   string
    GoVersion string
    Platform  string
}

func GetVersionInfo() VersionInfo
```

This approach provides the same logical organization as the Python SDK while being idiomatic Go code.