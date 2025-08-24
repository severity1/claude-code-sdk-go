# Usage Patterns Context

**Context**: Example usage patterns for Claude Code SDK Go with Query vs Client API patterns and Go-native concurrency

## Component Focus
- **Query API Patterns** - One-shot interactions with automatic cleanup
- **Client API Patterns** - Streaming/bidirectional conversations with persistent connections
- **Go-Native Concurrency** - Goroutines, channels, context patterns for SDK users
- **Error Handling Examples** - Proper error handling and resource cleanup patterns

## API Usage Patterns

### Query API (One-Shot)
**Use Cases**: Simple questions, batch processing, CI/CD scripts, code generation

```go
// Simple text query
func ExampleSimpleQuery() {
    ctx := context.Background()
    
    iterator, err := claudecode.Query(ctx, "What is 2+2?")
    if err != nil {
        log.Fatal(err)
    }
    defer iterator.Close()
    
    for {
        msg, err := iterator.Next(ctx)
        if err == io.EOF {
            break
        }
        if err != nil {
            log.Fatal(err)
        }
        
        // Handle message
        fmt.Printf("Response: %v\n", msg)
    }
}
```

### Query with Options
```go
func ExampleQueryWithOptions() {
    ctx := context.Background()
    
    iterator, err := claudecode.Query(ctx, "Analyze this code",
        claudecode.WithSystemPrompt("You are a code reviewer"),
        claudecode.WithModel("claude-opus-4"),
        claudecode.WithAllowedTools("read_file", "write_file"),
    )
    if err != nil {
        log.Fatal(err)
    }
    defer iterator.Close()
    
    // Process responses...
}
```

### Client API (Streaming/Bidirectional)
**Use Cases**: Interactive chat, multi-turn conversations, real-time applications

```go
// Streaming conversation
func ExampleStreamingClient() {
    ctx := context.Background()
    
    client := claudecode.NewClient(
        claudecode.WithSystemPrompt("You are a helpful assistant"),
    )
    
    // Connect and defer cleanup
    if err := client.Connect(ctx); err != nil {
        log.Fatal(err)
    }
    defer client.Disconnect()
    
    // Start receiving messages
    messages := client.ReceiveMessages(ctx)
    
    // Send initial query
    if err := client.Query(ctx, "Hello, can you help me?"); err != nil {
        log.Fatal(err)
    }
    
    // Handle responses
    for msg := range messages {
        fmt.Printf("Received: %v\n", msg)
        
        // Can send follow-up messages
        client.Query(ctx, "Tell me more")
    }
}
```

## Go-Native Concurrency Patterns

### Context-First Design
```go
func ExampleContextUsage() {
    // Always accept context as first parameter
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    // All SDK operations respect context
    iterator, err := claudecode.Query(ctx, "Long running query")
    if err != nil {
        if ctx.Err() == context.DeadlineExceeded {
            fmt.Println("Query timed out")
            return
        }
        log.Fatal(err)
    }
    defer iterator.Close()
    
    // Iterator also respects context
    for {
        msg, err := iterator.Next(ctx)
        if err == io.EOF {
            break
        }
        if err != nil {
            log.Fatal(err)
        }
        // Process message...
    }
}
```

### Goroutines with Channels
```go
func ExampleConcurrentQueries() {
    ctx := context.Background()
    queries := []string{"Query 1", "Query 2", "Query 3"}
    
    // Channel to collect results
    results := make(chan string, len(queries))
    
    // Launch goroutines for concurrent queries
    for _, query := range queries {
        go func(q string) {
            iterator, err := claudecode.Query(ctx, q)
            if err != nil {
                results <- fmt.Sprintf("Error: %v", err)
                return
            }
            defer iterator.Close()
            
            // Collect all responses
            var response strings.Builder
            for {
                msg, err := iterator.Next(ctx)
                if err == io.EOF {
                    break
                }
                if err != nil {
                    results <- fmt.Sprintf("Error: %v", err)
                    return
                }
                response.WriteString(fmt.Sprintf("%v ", msg))
            }
            
            results <- response.String()
        }(query)
    }
    
    // Collect results
    for i := 0; i < len(queries); i++ {
        result := <-results
        fmt.Printf("Result %d: %s\n", i+1, result)
    }
}
```

## Error Handling Patterns

### Proper Error Handling
```go
func ExampleErrorHandling() {
    ctx := context.Background()
    
    iterator, err := claudecode.Query(ctx, "Test query")
    if err != nil {
        // Check for specific error types
        var cliError *claudecode.CLINotFoundError
        if errors.As(err, &cliError) {
            fmt.Printf("Claude CLI not found: %v\n", cliError)
            fmt.Println("Please install: npm install -g @anthropic-ai/claude-code")
            return
        }
        
        var connError *claudecode.ConnectionError
        if errors.As(err, &connError) {
            fmt.Printf("Connection failed: %v\n", connError)
            return
        }
        
        // Generic error handling
        log.Fatal(err)
    }
    defer iterator.Close()
    
    // Handle iteration errors
    for {
        msg, err := iterator.Next(ctx)
        if err == io.EOF {
            break
        }
        if err != nil {
            fmt.Printf("Message processing error: %v\n", err)
            continue // Or break, depending on requirements
        }
        
        // Process message...
    }
}
```

### Resource Cleanup
```go
func ExampleResourceCleanup() {
    ctx := context.Background()
    
    client := claudecode.NewClient()
    
    // Always defer cleanup
    if err := client.Connect(ctx); err != nil {
        log.Fatal(err)
    }
    defer func() {
        if err := client.Disconnect(); err != nil {
            fmt.Printf("Cleanup error: %v\n", err)
        }
    }()
    
    // Use client...
}
```

## Advanced Usage Patterns

### Streaming with Message Processing
```go
func ExampleStreamingWithProcessing() {
    ctx := context.Background()
    client := claudecode.NewClient()
    
    if err := client.Connect(ctx); err != nil {
        log.Fatal(err)
    }
    defer client.Disconnect()
    
    // Create channels for message flow
    queries := make(chan string)
    responses := make(chan claudecode.Message)
    
    // Goroutine to send queries
    go func() {
        defer close(queries)
        queries <- "What is Go?"
        queries <- "Show me an example"
        queries <- "Thank you!"
    }()
    
    // Goroutine to handle responses
    go func() {
        defer close(responses)
        messages := client.ReceiveMessages(ctx)
        for msg := range messages {
            responses <- msg
        }
    }()
    
    // Main loop coordinating queries and responses
    for query := range queries {
        if err := client.Query(ctx, query); err != nil {
            log.Printf("Query error: %v", err)
            continue
        }
        
        // Wait for response (simplified)
        select {
        case response := <-responses:
            fmt.Printf("Q: %s\nA: %v\n\n", query, response)
        case <-ctx.Done():
            return
        }
    }
}
```

## Integration Examples

### Custom Transport (Advanced)
```go
func ExampleCustomTransport() {
    // For testing or custom backends
    mockTransport := &MyMockTransport{}
    
    client := claudecode.NewClient(
        claudecode.WithTransport(mockTransport),
    )
    
    // Use as normal...
}
```

## Best Practices Summary

### Context Usage
- Always pass context as first parameter
- Use context.WithTimeout for operations with time limits
- Respect context cancellation in all operations

### Error Handling  
- Check for specific error types using errors.As()
- Always handle io.EOF for iterators
- Provide helpful error messages and recovery suggestions

### Resource Management
- Always defer cleanup (Close(), Disconnect())
- Use defer immediately after successful resource acquisition
- Handle cleanup errors appropriately

### Concurrency
- Use goroutines for concurrent operations
- Use channels for goroutine communication
- Protect shared state with mutexes if needed