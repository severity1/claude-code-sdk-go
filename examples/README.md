# Claude Agent SDK for Go - Examples

Working examples demonstrating the Claude Agent SDK for Go. Both the **Query API** and **Client API** are production ready with full Python SDK compatibility.

## Prerequisites

- Go 1.18+
- Node.js 
- Claude Code CLI: `npm install -g @anthropic-ai/claude-code`

## Learning Path 📚

Examples are numbered from **easiest to hardest**. Follow this progression:

### 1. Start Here: Basic Usage

```bash
# 01 - Your first query (simplest)
cd examples/01_quickstart
go run main.go
```

### 2. Learn Streaming

```bash
# 02 - Real-time streaming responses
cd examples/02_client_streaming
go run main.go

# 03 - Multi-turn conversations with context
cd examples/03_client_multi_turn
go run main.go
```

### 3. Master Tools Integration

```bash
# 04 - Query API with file tools
cd examples/04_query_with_tools
go run main.go

# 05 - Client API with file tools (interactive)
cd examples/05_client_with_tools
go run main.go
```

### 4. MCP Tools Integration

```bash
# 06 - Query API with MCP tools (timezone queries)
cd examples/06_query_with_mcp
go run main.go

# 07 - Client API with MCP tools (multi-turn time workflows)
cd examples/07_client_with_mcp
go run main.go
```

### 5. Production Patterns

```bash
# 08 - Advanced error handling & model switching
cd examples/08_client_advanced
go run main.go

# 09 - WithClient pattern for automatic resource management
cd examples/09_context_manager
go run main.go

# 10 - Session management and isolation
cd examples/10_session_management
go run main.go

# 11 - Permission callbacks for tool access control
cd examples/11_permission_callback
go run main.go

# 12 - Hook system for lifecycle events
cd examples/12_hooks
go run main.go
```

### 6. Advanced Features

```bash
# 13 - File checkpointing and rewind
cd examples/13_file_checkpointing
go run main.go

# 14 - In-process SDK MCP servers
cd examples/14_sdk_mcp_server
go run main.go

# 15 - Programmatic subagents
cd examples/15_programmatic_subagents
go run main.go

# 16 - Type-safe structured output
cd examples/16_structured_output
go run main.go

# 17 - Custom plugin integration
cd examples/17_plugins
go run main.go

# 18 - Sandbox security (Linux/macOS)
cd examples/18_sandbox_security
go run main.go

# 19 - Partial streaming updates
cd examples/19_partial_streaming
go run main.go

# 20 - Debugging and diagnostics
cd examples/20_debugging_and_diagnostics
go run main.go
```

## Quick Test Example

Try this simple example to verify your setup:

```go
package main

import (
    "context"
    "errors"
    "fmt"
    "log"
    "time"
    
    "github.com/severity1/claude-agent-sdk-go"
)

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    iterator, err := claudecode.Query(ctx, "What is Go?")
    if err != nil {
        log.Fatal(err)
    }
    defer iterator.Close()
    
    for {
        message, err := iterator.Next(ctx)
        if err != nil {
            if errors.Is(err, claudecode.ErrNoMoreMessages) {
                break
            }
            log.Fatal(err)
        }
        
        if message == nil {
            break
        }
        
        if assistantMsg, ok := message.(*claudecode.AssistantMessage); ok {
            for _, block := range assistantMsg.Content {
                if textBlock, ok := block.(*claudecode.TextBlock); ok {
                    fmt.Print(textBlock.Text)
                }
            }
        }
    }
}
```

## Example Descriptions

### 🟢 Beginner Level

#### `01_quickstart/` - Your First Query
- **Concepts**: Basic Query API, message handling
- **Features**: Simple queries, system prompts, message processing
- **Time**: 2 minutes

#### `02_client_streaming/` - Real-Time Streaming  
- **Concepts**: Client API, streaming responses
- **Features**: Connection management, real-time processing
- **Time**: 5 minutes

#### `03_client_multi_turn/` - Conversations
- **Concepts**: Context preservation, multi-turn conversations
- **Features**: Follow-up questions, session management  
- **Time**: 5 minutes

### 🟡 Intermediate Level

#### `04_query_with_tools/` - File Operations
- **Concepts**: Tool integration, file manipulation
- **Features**: Read/Write/Edit tools, security restrictions
- **Time**: 10 minutes

#### `05_client_with_tools/` - Interactive File Workflows
- **Concepts**: Multi-turn tool usage, progressive development
- **Features**: Interactive file manipulation, context across tools
- **Time**: 10 minutes

#### `06_query_with_mcp/` - MCP Tools Integration
- **Concepts**: MCP tools, external service integration
- **Features**: Timezone queries using MCP time server
- **Prerequisites**: uvx (for mcp-server-time)
- **Time**: 10 minutes

### 🔴 Advanced Level

#### `07_client_with_mcp/` - Multi-Turn MCP Workflows
- **Concepts**: Multi-step MCP operations, context preservation
- **Features**: Time conversion across timezones, multi-turn workflows
- **Prerequisites**: uvx (for mcp-server-time)
- **Time**: 10 minutes

#### `08_client_advanced/` - Advanced Client Features
- **Concepts**: Dynamic model switching, structured error handling
- **Features**: SetModel(), type-specific error checking, multi-turn with model changes
- **Time**: 15 minutes

#### `09_context_manager/` - Resource Management Patterns
- **Concepts**: WithClient pattern vs manual connection management
- **Features**: Automatic resource cleanup, error handling comparison
- **Time**: 10 minutes

#### `10_session_management/` - Session Isolation
- **Concepts**: Session management, conversation isolation
- **Features**: Default vs custom sessions, QueryWithSession(), context separation
- **Time**: 10 minutes

#### `11_permission_callback/` - Permission Callbacks
- **Concepts**: Tool permission control, security policies
- **Features**: Allow/deny tool execution, path-based access control, audit logging
- **Time**: 15 minutes

#### `12_hooks/` - Hook System for Lifecycle Events
- **Concepts**: Lifecycle hooks, event interception
- **Features**: PreToolUse/PostToolUse hooks, command blocking, context injection
- **Time**: 15 minutes

#### `13_file_checkpointing/` - File Checkpointing and Rewind
- **Concepts**: File state management, checkpoint/rewind operations
- **Features**: WithFileCheckpointing(), RewindFiles(), message UUIDs
- **Time**: 15 minutes

#### `14_sdk_mcp_server/` - In-Process SDK MCP Servers
- **Concepts**: Custom MCP tools, in-process tool execution
- **Features**: NewTool(), CreateSDKMcpServer(), WithSdkMcpServer()
- **Time**: 20 minutes

### Expert Level

#### `15_programmatic_subagents/` - Programmatic Subagents
- **Concepts**: Agent definitions, specialized agents
- **Features**: WithAgent(), WithAgents(), AgentDefinition, AgentModel constants
- **Time**: 15 minutes

#### `16_structured_output/` - Type-Safe Structured Output
- **Concepts**: JSON schema constraints, structured responses
- **Features**: WithJSONSchema(), WithOutputFormat(), ResultMessage.StructuredOutput
- **Time**: 15 minutes

#### `17_plugins/` - Plugin Configuration
- **Concepts**: Plugin integration, extensibility
- **Features**: WithLocalPlugin(), WithPlugins(), SdkPluginConfig
- **Time**: 10 minutes

#### `18_sandbox_security/` - Sandbox Security
- **Concepts**: Command isolation, security boundaries
- **Features**: WithSandboxEnabled(), WithSandboxNetwork(), excluded commands
- **Platform**: Linux/macOS only
- **Time**: 15 minutes

#### `19_partial_streaming/` - Partial Streaming
- **Concepts**: Real-time updates, progressive rendering
- **Features**: WithPartialStreaming(), StreamEvent types, delta handling
- **Time**: 15 minutes

#### `20_debugging_and_diagnostics/` - Debugging and Diagnostics
- **Concepts**: Debug output, environment config, health monitoring
- **Features**: WithDebugWriter(), WithStderrCallback(), GetServerInfo()
- **Time**: 15 minutes

## Common Patterns

### Query API - One-Shot Operations
```go
// Simple query
iterator, err := claudecode.Query(ctx, "Explain Go interfaces")

// With system prompt
iterator, err := claudecode.Query(ctx, "Review this code",
    claudecode.WithSystemPrompt("You are a senior Go developer"))

// With tools
iterator, err := claudecode.Query(ctx, "Analyze all files",
    claudecode.WithAllowedTools("Read", "Write"))
```

### Client API - Conversations

**WithClient Pattern (Recommended):**
```go
// Automatic resource management
err := claudecode.WithClient(ctx, func(client claudecode.Client) error {
    // First question
    client.Query(ctx, "What is dependency injection?")
    // Process response...
    
    // Follow-up (context preserved)
    return client.Query(ctx, "Show me a Go example")
})
```

**Manual Pattern (Still Supported):**
```go
client := claudecode.NewClient()
defer client.Disconnect()

// First question
client.Query(ctx, "What is dependency injection?")
// Process response...

// Follow-up (context preserved)
client.Query(ctx, "Show me a Go example")
// Process response...
```

### Session Management - Clean API

**New Clean Query API (Recommended):**
```go
// Default session
client.Query(ctx, "What is dependency injection?")

// Custom session
client.QueryWithSession(ctx, "Remember this context", "my-session")

// Sessions are isolated from each other
client.Query(ctx, "What did I just say?") // Won't remember "my-session" context
```

**Benefits of New API:**
- ✅ **Clear intent**: `Query()` vs `QueryWithSession()`
- ✅ **Type safety**: No variadic parameter confusion
- ✅ **Python parity**: Matches Python SDK `client.query(session_id="...")`
- ✅ **Go idioms**: Follows stdlib patterns like `WithContext()`, `WithTimeout()`

### MCP Tools - Cloud Integration
```go
// AWS operations (explicit tool names required - no wildcards)
iterator, err := claudecode.Query(ctx, "List my S3 buckets",
    claudecode.WithAllowedTools(
        "mcp__aws-api-mcp__call_aws",
        "mcp__aws-api-mcp__suggest_aws_commands"))
```

### Tools Presets
```go
// Explicit list - Maximum control
claudecode.WithAllowedTools("Read", "Write", "Edit")

// Preset - Convenience (full Claude Code toolset)
claudecode.WithClaudeCodeTools()

// Custom preset
claudecode.WithToolsPreset("my_custom_preset")
```

## Error Handling

```go
iterator, err := claudecode.Query(ctx, "test")
if err != nil {
    var cliError *claudecode.CLINotFoundError
    if errors.As(err, &cliError) {
        fmt.Println("Please install: npm install -g @anthropic-ai/claude-code")
        return
    }
    log.Fatal(err)
}
```

## When to Use Which API

### 🎯 Query API - Choose When:
- One-shot questions or commands
- Batch processing  
- CI/CD scripts
- Simple automation
- Lower resource overhead

### 🔄 Client API - Choose When:
- Multi-turn conversations
- Interactive applications  
- Context-dependent workflows
- Real-time streaming needs
- Complex state management

## Need Help?

- Check the [main README](../README.md) for installation
- Start with `01_quickstart` for basic patterns
- Follow the numbered progression for best learning experience
- SDK follows [Python SDK](https://docs.anthropic.com/en/docs/claude-code/sdk) patterns