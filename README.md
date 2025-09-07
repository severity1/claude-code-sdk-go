# Claude Code SDK for Go

Production-ready Go SDK for Claude Code CLI integration. Build applications that leverage Claude's code understanding, file operations, and external tool integrations through a clean, idiomatic Go API.

**🚀 Two powerful APIs for different use cases:**
- **Query API**: One-shot operations, automation, CI/CD integration  
- **Client API**: Interactive conversations, multi-turn workflows, streaming responses
- **WithClient**: New Go-idiomatic context manager for automatic resource management ✨

## Installation

```bash
go get github.com/severity1/claude-code-sdk-go
```

**Prerequisites:** Go 1.18+, Node.js, Claude Code (`npm install -g @anthropic-ai/claude-code`)

## Key Features

✅ **Two APIs for different needs** - Query for automation, Client for interaction  
✅ **100% Python SDK compatibility** - Same functionality, Go-native design  
✅ **Automatic resource management** - WithClient provides Go-idiomatic context manager pattern ✨  
✅ **Built-in tool integration** - File operations, AWS, GitHub, databases, and more  
✅ **Production ready** - Comprehensive error handling, timeouts, resource cleanup  
✅ **Security focused** - Granular tool permissions and access controls  
✅ **Context-aware** - Maintain conversation state across multiple interactions  

## Usage

### Query API - One-Shot Operations
Best for automation, scripting, and tasks with clear completion criteria:

```go
package main

import (
    "context"
    "errors"
    "fmt"
    "log"

    "github.com/severity1/claude-code-sdk-go"
)

func main() {
    fmt.Println("Claude Code SDK - Query API Example")
    fmt.Println("Asking: What is 2+2?")

    ctx := context.Background()

    // Create and execute query
    iterator, err := claudecode.Query(ctx, "What is 2+2?")
    if err != nil {
        log.Fatalf("Query failed: %v", err)
    }
    defer iterator.Close()

    fmt.Println("\nResponse:")

    // Iterate through messages
    for {
        message, err := iterator.Next(ctx)
        if err != nil {
            if errors.Is(err, claudecode.ErrNoMoreMessages) {
                break
            }
            log.Fatalf("Failed to get message: %v", err)
        }

        if message == nil {
            break
        }

        // Handle different message types
        switch msg := message.(type) {
        case *claudecode.AssistantMessage:
            for _, block := range msg.Content {
                if textBlock, ok := block.(*claudecode.TextBlock); ok {
                    fmt.Print(textBlock.Text)
                }
            }
        case *claudecode.ResultMessage:
            if msg.IsError {
                log.Printf("Error: %s", msg.Result)
            }
        }
    }

    fmt.Println("\nQuery completed!")
}
```

### Client API - Interactive & Multi-Turn
**New: WithClient provides automatic resource management (equivalent to Python's `async with`):**

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/severity1/claude-code-sdk-go"
)

func main() {
    fmt.Println("Claude Code SDK - Client Streaming Example")
    fmt.Println("Asking: Explain Go goroutines with a simple example")

    ctx := context.Background()
    question := "Explain what Go goroutines are and show a simple example"

    // WithClient handles connection lifecycle automatically
    err := claudecode.WithClient(ctx, func(client claudecode.Client) error {
        fmt.Println("\nConnected! Streaming response:")

        if err := client.Query(ctx, question); err != nil {
            return fmt.Errorf("query failed: %w", err)
        }

        // Stream messages in real-time
        msgChan := client.ReceiveMessages(ctx)
        for {
            select {
            case message := <-msgChan:
                if message == nil {
                    return nil // Stream ended
                }

                switch msg := message.(type) {
                case *claudecode.AssistantMessage:
                    // Print streaming text as it arrives
                    for _, block := range msg.Content {
                        if textBlock, ok := block.(*claudecode.TextBlock); ok {
                            fmt.Print(textBlock.Text)
                        }
                    }
                case *claudecode.ResultMessage:
                    if msg.IsError {
                        return fmt.Errorf("error: %s", msg.Result)
                    }
                    return nil // Success, stream complete
                }
            case <-ctx.Done():
                return ctx.Err()
            }
        }
    })

    if err != nil {
        log.Fatalf("Streaming failed: %v", err)
    }

    fmt.Println("\n\nStreaming completed!")
}
```

**Traditional Client API (still supported):**

<details>
<summary>Click to see manual resource management approach</summary>

```go
func traditionalClientExample() {
    ctx := context.Background()
    
    client := claudecode.NewClient()
    if err := client.Connect(ctx); err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }
    defer client.Disconnect() // Manual cleanup required
    
    // Use client...
}
```
</details>

## Tool Integration & External Services

Integrate with file systems, cloud services, databases, and development tools:

**Core Tools** (built-in file operations):
```go
// File analysis and documentation generation
claudecode.Query(ctx, "Read all Go files and create API documentation",
    claudecode.WithAllowedTools("Read", "Write"))
```

**MCP Tools** (external service integrations):
```go
// AWS infrastructure automation
claudecode.Query(ctx, "List my S3 buckets and analyze their security settings", 
    claudecode.WithAllowedTools("mcp__aws-api-mcp__call_aws", "mcp__aws-api-mcp__suggest_aws_commands", "Write"))
```

## When to Use Which API

**🎯 Use Query API when you:**
- Need one-shot automation or scripting
- Have clear task completion criteria  
- Want automatic resource cleanup
- Are building CI/CD integrations
- Prefer simple, stateless operations

**🔄 Use Client API (WithClient) when you:**  
- Need interactive conversations
- Want to build context across multiple requests
- Are creating complex, multi-step workflows
- Need real-time streaming responses
- Want to iterate and refine based on previous results
- **Need automatic resource management (recommended)** ✨

## Examples & Documentation

Comprehensive examples covering every use case:

**Learning Path (Easiest → Hardest):**
- [`examples/01_quickstart/`](examples/01_quickstart/) - Query API fundamentals
- [`examples/02_client_streaming/`](examples/02_client_streaming/) - WithClient streaming basics ✨  
- [`examples/03_client_multi_turn/`](examples/03_client_multi_turn/) - Multi-turn conversations with automatic cleanup ✨
- [`examples/10_context_manager/`](examples/10_context_manager/) - WithClient vs manual patterns comparison ✨

**Tool Integration:**
- [`examples/04_query_with_tools/`](examples/04_query_with_tools/) - File operations with Query API
- [`examples/05_client_with_tools/`](examples/05_client_with_tools/) - Interactive file workflows  
- [`examples/06_query_with_mcp/`](examples/06_query_with_mcp/) - AWS automation with Query API
- [`examples/07_client_with_mcp/`](examples/07_client_with_mcp/) - AWS management with Client API

**Advanced Patterns:**
- [`examples/08_client_advanced/`](examples/08_client_advanced/) - WithClient error handling and production patterns ✨
- [`examples/09_client_vs_query/`](examples/09_client_vs_query/) - Modern API comparison and guidance ✨

**📖 [Full Documentation](examples/README.md)** with usage patterns, security best practices, and troubleshooting.

## License

MIT
