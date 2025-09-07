# Claude Code SDK for Go - Examples

Working examples demonstrating the Claude Code SDK for Go. Both the **Query API** and **Client API** are production ready with full Python SDK compatibility.

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

### 4. Advanced Cloud Integration

```bash
# 06 - Query API with AWS MCP tools
cd examples/06_query_with_mcp
go run main.go

# 07 - Client API with AWS MCP tools (requires AWS credentials)
cd examples/07_client_with_mcp
go run main.go
```

### 5. Production Patterns

```bash
# 08 - Advanced error handling & retries
cd examples/08_client_advanced
go run main.go

# 09 - API comparison & selection guide
cd examples/09_client_vs_query
go run main.go

# 10 - WithClient pattern for automatic resource management
cd examples/10_context_manager
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
    
    "github.com/severity1/claude-code-sdk-go"
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

#### `06_query_with_mcp/` - Cloud Integration (AWS)
- **Concepts**: MCP tools, cloud service integration
- **Features**: AWS S3 bucket listing, infrastructure queries
- **Prerequisites**: AWS credentials
- **Time**: 15 minutes

### 🔴 Advanced Level

#### `07_client_with_mcp/` - Advanced Cloud Workflows
- **Concepts**: Multi-step cloud operations, security analysis
- **Features**: AWS security assessment, progressive analysis
- **Prerequisites**: AWS credentials, AWS MCP server
- **Time**: 20 minutes

#### `08_client_advanced/` - Production Patterns
- **Concepts**: Error handling, retries, production deployment
- **Features**: Robust error handling, connection retries, timeout management
- **Time**: 15 minutes

#### `09_client_vs_query/` - Architecture Decision Guide
- **Concepts**: API selection, performance considerations
- **Features**: Side-by-side comparison, use case guidance
- **Time**: 15 minutes

#### `10_context_manager/` - Resource Management Patterns
- **Concepts**: WithClient pattern vs manual connection management
- **Features**: Automatic resource cleanup, error handling comparison
- **Time**: 10 minutes

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

### MCP Tools - Cloud Integration
```go
// AWS operations (explicit tool names required - no wildcards)
iterator, err := claudecode.Query(ctx, "List my S3 buckets",
    claudecode.WithAllowedTools(
        "mcp__aws-api-mcp__call_aws",
        "mcp__aws-api-mcp__suggest_aws_commands"))
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