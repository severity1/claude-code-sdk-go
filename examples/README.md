# Claude Code SDK for Go - Examples

Working examples demonstrating the Claude Code SDK for Go. Both the **Query API** and **Client API** are production ready with full Python SDK compatibility.

## Prerequisites

- Go 1.18+
- Node.js 
- Claude Code CLI: `npm install -g @anthropic-ai/claude-code`

## Quick Start

### 1. Query API Examples

Navigate to the quickstart example for one-shot queries:

```bash
cd examples/quickstart
go run main.go
```

This demonstrates:
- Simple one-shot query with Claude
- Message type handling (`AssistantMessage`, `SystemMessage`, `ResultMessage`)
- Content block processing (`TextBlock`, `ThinkingBlock`)
- Using query options like system prompts

### 2. Client API Examples

Try the streaming examples for multi-turn conversations:

```bash
# Basic streaming
cd examples/client_streaming
go run main.go

# Multi-turn conversation
cd examples/client_multi_turn
go run main.go

# Advanced features
cd examples/client_advanced
go run main.go

# Compare Query vs Client APIs
cd examples/client_vs_query
go run main.go
```

### 3. Try Your Own Query

Modify the quickstart example or create a simple test:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    
    "github.com/severity1/claude-code-sdk-go"
)

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    iterator, err := claudecode.Query(ctx, "Your question here")
    if err != nil {
        log.Fatal(err)
    }
    defer iterator.Close()
    
    for {
        message, err := iterator.Next(ctx)
        if err != nil {
            if err.Error() == "no more messages" {
                break
            }
            log.Fatal(err)
        }
        
        if assistantMsg, ok := message.(*claudecode.AssistantMessage); ok {
            for _, block := range assistantMsg.Content {
                if textBlock, ok := block.(*claudecode.TextBlock); ok {
                    fmt.Printf("Claude: %s\n", textBlock.Text)
                }
            }
        }
    }
}
```

### 4. Using in Your Project

Add the SDK to your project:

```bash
go get github.com/severity1/claude-code-sdk-go
```

Import and use:

```go
import "github.com/severity1/claude-code-sdk-go"

// Use claudecode.Query(ctx, "your prompt")
```

## Available Examples

### Query API Examples

#### `/quickstart/`
- **File**: `main.go`
- **Purpose**: Basic Query API usage with comprehensive message handling
- **Features**: Simple queries, system prompts, message type processing
- **Run**: `cd quickstart && go run main.go`

### Client API Examples

#### `/client_streaming/`
- **File**: `main.go`
- **Purpose**: Basic Client API streaming with real-time response processing
- **Features**: Connection management, streaming messages, resource cleanup
- **Run**: `cd client_streaming && go run main.go`

#### `/client_multi_turn/`
- **File**: `main.go`
- **Purpose**: Multi-turn conversation demonstration
- **Features**: Context preservation, follow-up questions, session management
- **Run**: `cd client_multi_turn && go run main.go`

#### `/client_advanced/`
- **File**: `main.go`
- **Purpose**: Advanced Client features and error handling
- **Features**: System prompts, custom options, comprehensive error handling
- **Run**: `cd client_advanced && go run main.go`

#### `/client_vs_query/`
- **File**: `main.go`
- **Purpose**: Side-by-side comparison of Query vs Client APIs
- **Features**: Use case comparison, performance considerations, best practices
- **Run**: `cd client_vs_query && go run main.go`

### Tool Integration Examples

#### `/query_with_tools/`
- **File**: `main.go`
- **Purpose**: Query API with core Claude Code tools (Read/Write/Edit)
- **Features**: File operations, tool restrictions, security patterns
- **Run**: `cd query_with_tools && go run main.go`

#### `/client_with_tools/`
- **File**: `main.go`
- **Purpose**: Client API with core tools for interactive file manipulation
- **Features**: Multi-turn file workflows, progressive development, context preservation
- **Run**: `cd client_with_tools && go run main.go`

#### `/query_with_mcp/`
- **File**: `main.go`
- **Purpose**: Query API with MCP tools (AWS API integration)
- **Features**: AWS infrastructure management, cost analysis, security audits
- **Prerequisites**: AWS credentials, `claude mcp add aws-api`
- **Run**: `cd query_with_mcp && go run main.go`

#### `/client_with_mcp/`
- **File**: `main.go`
- **Purpose**: Client API with MCP tools for interactive AWS management
- **Features**: Multi-step AWS workflows, infrastructure optimization, progressive analysis
- **Prerequisites**: AWS credentials, `claude mcp add aws-api`
- **Run**: `cd client_with_mcp && go run main.go`

#### `/tools_comparison/`
- **File**: `main.go`
- **Purpose**: Compare Query vs Client APIs for tool-heavy workflows
- **Features**: Performance analysis, resource implications, use case guidance
- **Run**: `cd tools_comparison && go run main.go`

## Error Handling

The SDK provides structured error handling:

```go
iterator, err := claudecode.Query(ctx, "test")
if err != nil {
    // Check for specific errors
    var cliError *claudecode.CLINotFoundError
    if errors.As(err, &cliError) {
        fmt.Println("Please install Claude Code CLI:")
        fmt.Println("npm install -g @anthropic-ai/claude-code")
        return
    }
    log.Fatal(err)
}
```

## Development Status

- ✅ **Query API**: Production ready - one-shot queries with full message support
- ✅ **Client API**: Production ready - streaming/bidirectional conversations with Python SDK compatibility

**Query API** is perfect for:
- Code analysis and generation
- Documentation tasks  
- CI/CD integration
- Batch processing

**Client API** is ideal for:
- Interactive applications
- Multi-turn conversations
- Real-time streaming
- Complex workflows

## Common Patterns

### Query API Patterns

#### With System Prompt
```go
iterator, err := claudecode.Query(ctx, "Analyze this code",
    claudecode.WithSystemPrompt("You are a code reviewer"))
```

#### With Timeout
```go
ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
defer cancel()
```

#### Processing All Content
```go
for {
    message, err := iterator.Next(ctx)
    if err != nil {
        if err.Error() == "no more messages" {
            break
        }
        log.Fatal(err)
    }
    
    switch msg := message.(type) {
    case *claudecode.AssistantMessage:
        // Claude's response
        for _, block := range msg.Content {
            switch b := block.(type) {
            case *claudecode.TextBlock:
                fmt.Print(b.Text)
            case *claudecode.ThinkingBlock:
                fmt.Printf("[Thinking: %s]", b.Thinking)
            }
        }
    case *claudecode.ResultMessage:
        if msg.IsError {
            fmt.Printf("Error: %s\n", msg.Result)
        }
    }
}
```

### Client API Patterns

#### Basic Streaming
```go
client := claudecode.NewClient()
if err := client.Connect(ctx); err != nil {
    log.Fatal(err)
}
defer client.Disconnect()

client.Query(ctx, "Hello!")
msgChan := client.ReceiveMessages(ctx)
for {
    select {
    case message := <-msgChan:
        if message == nil {
            return // Stream ended
        }
        // Process streaming message...
    case <-ctx.Done():
        return // Context cancelled
    }
}
```

#### Multi-Turn Conversation
```go
client := claudecode.NewClient()
defer client.Disconnect()

// First question
client.Query(ctx, "What is Go?")
// ... process response

// Follow-up question (context preserved)
client.Query(ctx, "Show me an example")
// ... process response
```

#### With Options
```go
client := claudecode.NewClient(
    claudecode.WithSystemPrompt("You are a helpful assistant"),
    claudecode.WithModel("claude-opus-4"),
)
```

### Tool Usage Patterns

#### Core Tools (Read, Write, Edit)
```go
// File analysis and documentation
claudecode.Query(ctx, "Read all files and create comprehensive docs",
    claudecode.WithAllowedTools("Read", "Write"))

// Code refactoring with safety
claudecode.Query(ctx, "Improve error handling in main.go", 
    claudecode.WithAllowedTools("Read", "Edit"))

// Read-only security restriction
claudecode.Query(ctx, "Analyze code for security issues",
    claudecode.WithAllowedTools("Read"))
```

#### MCP Tools Integration
```go
// AWS infrastructure management  
claudecode.Query(ctx, "Audit my AWS costs and create optimization plan",
    claudecode.WithAllowedTools("mcp__aws-api-mcp__*", "Write"))

// GitHub workflow automation
client := claudecode.NewClient(
    claudecode.WithAllowedTools("mcp__github__*", "Read", "Write"),
)

// Database analysis with safety
claudecode.Query(ctx, "Analyze user patterns from database",
    claudecode.WithAllowedTools("mcp__postgres__select*", "Write"), // Read-only DB access
)
```

#### Tool Security Best Practices
```go
// Principle of least privilege
claudecode.Query(ctx, "Review S3 bucket permissions",
    claudecode.WithAllowedTools("mcp__aws-api-mcp__describe*", "mcp__aws-api-mcp__list*"), // Read-only
    claudecode.WithDisallowedTools("mcp__aws-api-mcp__delete*"), // Block destructive operations
)

// Separate concerns
claudecode.Query(ctx, "Generate infrastructure report", 
    claudecode.WithAllowedTools("mcp__aws-api-mcp__describe*", "Write"), // Can read AWS + write files
)
```

#### Tool Combination Patterns
```go
// Multi-tool workflows
claudecode.Query(ctx, "Backup database schema and create documentation",
    claudecode.WithAllowedTools("mcp__postgres__*", "Write", "Edit"))

// Progressive analysis
client := claudecode.NewClient(
    claudecode.WithAllowedTools("Read", "mcp__aws-api-mcp__*", "Write"),
)
// Can read local files, analyze AWS, and generate reports across multiple interactions
```

## Need Help?

- Check the [main README](../README.md) for installation and setup
- Review the quickstart example for working patterns
- SDK follows [Python SDK](https://docs.anthropic.com/en/docs/claude-code/sdk) patterns