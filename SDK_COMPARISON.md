# Python SDK vs Go SDK Comparison

## Executive Summary

The Go SDK achieves **excellent feature parity** with the Python SDK for all core functionality. The Go SDK was developed based on the stable Python SDK API from ~4 weeks ago, and both SDKs had near-perfect parity at that time.

**Recent Python SDK additions (Sep 2-8, 2025)**: Several advanced features were added to the Python SDK within the last week, creating new gaps. These are not "missing" features from the Go SDK but rather very recent additions to Python that haven't been ported yet.

## Timeline Context

- **4 weeks ago**: Python and Go SDKs had near-perfect feature parity
- **Sep 2, 2025**: Control protocol added to Python SDK  
- **Sep 3, 2025**: MCP server support and permission callbacks added
- **Sep 8, 2025**: Hooks system added
- **Today**: Go SDK maintains all original functionality; Python has new advanced features

## Core Features (✅ Complete Parity)

### Basic Operations
| Feature | Python SDK | Go SDK | Notes |
|---------|------------|--------|-------|
| `query()` function | ✅ | ✅ | One-shot queries |
| `Client` for streaming | ✅ | ✅ | Bidirectional communication |
| Message types | ✅ | ✅ | User, Assistant, System, Result |
| Content blocks | ✅ | ✅ | Text, Thinking, ToolUse, ToolResult |
| Error handling | ✅ | ✅ | Structured error types |

### Configuration Options
| Feature | Python SDK | Go SDK | Notes |
|---------|------------|--------|-------|
| Allowed/disallowed tools | ✅ | ✅ | Tool permission control |
| System prompts | ✅ | ✅ | Custom system/append prompts |
| Model selection | ✅ | ✅ | Model override |
| Max thinking tokens | ✅ | ✅ | Default: 8000 |
| Permission modes | ✅ | ✅ | default, acceptEdits, plan, bypassPermissions |
| Working directory | ✅ | ✅ | Custom cwd |
| Add directories | ✅ | ✅ | Additional context directories |
| Session management | ✅ | ✅ | Resume, continue conversation |
| Max turns | ✅ | ✅ | Conversation limits |
| Extra CLI args | ✅ | ✅ | Arbitrary flag passing |

## Recently Added Python Features (🆕 Not Yet in Go SDK)

### 1. MCP Server Support
**Python SDK**: Full support for MCP (Model Context Protocol) servers
- SDK MCP servers (in-process)
- Stdio MCP servers
- SSE MCP servers  
- HTTP MCP servers
- `@tool` decorator for defining tools
- `create_sdk_mcp_server()` function

**Go SDK**: Basic config structures exist but no implementation
- Missing: SDK server support
- Missing: Tool decorator equivalent
- Missing: Server creation helpers
- Missing: MCP message handling

**Impact**: Cannot integrate with MCP-based tools and extensions (added Sep 3, 2025)  
**Tracking**: [Issue #7](https://github.com/severity1/claude-code-sdk-go/issues/7)

### 2. Permission Callbacks
**Python SDK**: Dynamic permission control via callbacks
- `can_use_tool` callback function
- `ToolPermissionContext` with suggestions
- `PermissionResult` (Allow/Deny) types
- `PermissionUpdate` configuration
- Dynamic permission rules

**Go SDK**: No callback mechanism
- Static permission modes only
- No runtime permission decisions
- No permission suggestions

**Impact**: Less flexible permission control for enterprise use cases (added Sep 3, 2025)  
**Tracking**: [Issue #8](https://github.com/severity1/claude-code-sdk-go/issues/8)

### 3. Hook System
**Python SDK**: Event-based hooks for lifecycle management
- Hook events: PreToolUse, PostToolUse, UserPromptSubmit, Stop, etc.
- `HookCallback` functions
- `HookMatcher` for selective hooks
- `HookJSONOutput` for hook responses
- Hook context with abort signals

**Go SDK**: No hook support
- No lifecycle callbacks
- No event interception
- No custom middleware

**Impact**: Cannot implement custom logging, auditing, or tool interceptors (added Sep 8, 2025)  
**Tracking**: [Issue #9](https://github.com/severity1/claude-code-sdk-go/issues/9)

### 4. SDK Control Protocol
**Python SDK**: Advanced control messages
- Interrupt requests
- Permission requests
- Initialize requests
- Set permission mode requests
- Hook callback requests
- MCP message requests

**Go SDK**: Basic interrupt only
- Missing: Permission control protocol
- Missing: Hook protocol
- Missing: MCP protocol

**Impact**: Limited runtime control over SDK behavior (added Sep 2, 2025)  
**Tracking**: [Issue #10](https://github.com/severity1/claude-code-sdk-go/issues/10)

### 5. Environment Variables
**Python SDK**: 
- Custom environment variables for subprocess
- Debug stderr redirection

**Go SDK**: 
- No custom environment variable support
- No debug output redirection

**Impact**: Less flexibility for debugging and custom environments

## Go-Specific Advantages

### 1. Interface-Driven Design
- Clean `Transport` interface for testing
- `Message` and `ContentBlock` interfaces
- Better testability through mocking

### 2. Functional Options Pattern
- `WithAllowedTools()`, `WithModel()`, etc.
- More idiomatic than Python's dataclass approach
- Cleaner API for configuration

### 3. Context-First Design
- All blocking operations accept `context.Context`
- Native cancellation and timeout support
- Better concurrency control

### 4. Type Safety
- Compile-time type checking
- No runtime type errors
- Explicit error handling

## Implementation Recommendations

### Priority 1: MCP Server Support
The most significant gap is MCP server support. This would require:
- Implementing MCP protocol handling
- Creating Go equivalents of the tool decorator pattern
- Adding server registration and lifecycle management

### Priority 2: Permission Callbacks
Adding dynamic permission control:
- Define callback function types
- Implement permission context passing
- Add runtime permission evaluation

### Priority 3: Hook System
Event-based extensibility:
- Define hook event types
- Implement hook registration
- Add hook execution in message flow

### Priority 4: Enhanced Control Protocol
Full SDK control protocol:
- Implement all control message types
- Add bidirectional control flow
- Support dynamic configuration updates

## Conclusion

The Go SDK was developed with **excellent feature parity** based on the stable Python SDK from 4 weeks ago. At that time, both SDKs were essentially equivalent in functionality.

**The recent Python additions (last 8 days) represent new capabilities** rather than gaps the Go SDK was missing. The Go SDK's original scope and implementation remain solid and complete.

### Go SDK Strengths
- **Complete core functionality**: All essential features implemented
- **Superior architecture**: Better testability through interfaces, native concurrency with contexts
- **Idiomatic Go**: Type-safe configuration with functional options, clean separation of concerns
- **Proven stability**: Based on mature, stable API patterns

### Roadmap for Advanced Features

The recent Python SDK features can be implemented in Go with these tracked issues:

1. **Enhanced Control Protocol** ([Issue #10](https://github.com/severity1/claude-code-sdk-go/issues/10)) - **Priority: High**
   - Foundation for all other advanced features
   - Bidirectional communication with CLI
   - Request/response routing system

2. **Permission Callback System** ([Issue #8](https://github.com/severity1/claude-code-sdk-go/issues/8)) - **Priority: Medium**
   - Dynamic runtime permission control
   - Enterprise-grade security features
   - Depends on control protocol

3. **MCP Server Support** ([Issue #7](https://github.com/severity1/claude-code-sdk-go/issues/7)) - **Priority: Medium** 
   - In-process tool server capabilities
   - Integration with MCP ecosystem
   - Depends on control protocol

4. **Hook System** ([Issue #9](https://github.com/severity1/claude-code-sdk-go/issues/9)) - **Priority: Low-Medium**
   - Lifecycle event handling
   - Custom logging and auditing
   - Depends on control protocol

However, the Go SDK is **feature-complete** for its original scope and continues to provide excellent value as-is.