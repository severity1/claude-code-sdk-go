# Advanced Features

This document describes the SDK's advanced features for production use cases.

## Permission Callbacks

Permission callbacks allow programmatic control over tool access, enabling custom security policies.

### How It Works

When Claude requests to use a tool, the SDK intercepts the request and invokes your callback before allowing execution.

```
Claude: "I want to use the Bash tool to run 'rm -rf /'"
    │
    ▼
SDK intercepts tool request
    │
    ▼
Your callback receives:
    - tool: "Bash"
    - input: {"command": "rm -rf /"}
    - context: {file paths, etc.}
    │
    ▼
Your callback returns:
    - Allow: Tool executes
    - Deny: Tool blocked with message
    - AskUser: Prompt user for decision
```

### Configuration

```go
client := claudecode.NewClient(
    claudecode.WithCanUseTool(func(
        ctx context.Context,
        tool string,
        input map[string]any,
        permCtx claudecode.ToolPermissionContext,
    ) (claudecode.PermissionResult, error) {
        // Example: Block dangerous commands
        if tool == "Bash" {
            cmd := input["command"].(string)
            if strings.Contains(cmd, "rm -rf") {
                return claudecode.NewPermissionResultDeny(
                    "Destructive commands not allowed",
                ), nil
            }
        }

        // Example: Restrict file access to project directory
        if tool == "Read" || tool == "Write" {
            path := input["path"].(string)
            if !strings.HasPrefix(path, "/project/") {
                return claudecode.NewPermissionResultDeny(
                    "Access restricted to /project/",
                ), nil
            }
        }

        return claudecode.NewPermissionResultAllow(), nil
    }),
)
```

### Permission Results

```go
// Allow tool execution
claudecode.NewPermissionResultAllow()

// Deny with reason
claudecode.NewPermissionResultDeny("Reason for denial")

// Prompt user for decision
claudecode.NewPermissionResultAskUser("Should I allow this?")
```

## Lifecycle Hooks

Hooks intercept tool execution at key points, enabling logging, modification, or blocking.

### Hook Events

| Event | When Fired | Use Cases |
|-------|------------|-----------|
| `PreToolUse` | Before tool executes | Logging, validation, blocking |
| `PostToolUse` | After tool completes | Logging, result modification |

### Configuration

```go
client := claudecode.NewClient(
    claudecode.WithHooks(map[claudecode.HookEvent][]claudecode.HookMatcher{
        claudecode.HookEventPreToolUse: {
            {
                // Match specific tools
                ToolName: "Bash",
                Callback: func(ctx context.Context, event claudecode.HookEvent, data map[string]any) (claudecode.HookResult, error) {
                    cmd := data["input"].(map[string]any)["command"].(string)
                    log.Printf("Bash command: %s", cmd)

                    // Block dangerous patterns
                    if strings.Contains(cmd, "sudo") {
                        return claudecode.HookResultBlock("sudo not allowed"), nil
                    }

                    return claudecode.HookResultContinue(), nil
                },
            },
            {
                // Match all tools
                Callback: func(ctx context.Context, event claudecode.HookEvent, data map[string]any) (claudecode.HookResult, error) {
                    log.Printf("Tool: %s", data["tool"])
                    return claudecode.HookResultContinue(), nil
                },
            },
        },
        claudecode.HookEventPostToolUse: {
            {
                Callback: func(ctx context.Context, event claudecode.HookEvent, data map[string]any) (claudecode.HookResult, error) {
                    log.Printf("Tool %s completed", data["tool"])
                    return claudecode.HookResultContinue(), nil
                },
            },
        },
    }),
)
```

### Hook Results

```go
// Allow execution to continue
claudecode.HookResultContinue()

// Block execution with message
claudecode.HookResultBlock("Reason for blocking")

// Modify context (PreToolUse only)
claudecode.HookResultModify(modifiedData)
```

## MCP Server Integration

MCP (Model Context Protocol) servers extend Claude's capabilities with external tools.

### External MCP Servers

Configure external MCP servers (stdio, HTTP, SSE):

```go
client := claudecode.NewClient(
    claudecode.WithMcpServers(map[string]claudecode.McpServerConfig{
        "time-server": {
            Command: "uvx",
            Args:    []string{"mcp-server-time"},
        },
        "database": {
            Command: "node",
            Args:    []string{"./mcp-database-server.js"},
            Env: map[string]string{
                "DATABASE_URL": "postgres://...",
            },
        },
    }),
)

// Use MCP tools
iterator, err := claudecode.Query(ctx, "What time is it in Tokyo?",
    claudecode.WithAllowedTools("mcp__time-server__get_time"),
)
```

### SDK MCP Servers (In-Process)

Create custom tools that run in your Go process:

```go
// Define a custom tool
weatherTool := claudecode.NewTool(
    "get_weather",
    "Get current weather for a location",
    map[string]any{
        "type": "object",
        "properties": map[string]any{
            "location": map[string]any{
                "type":        "string",
                "description": "City name",
            },
        },
        "required": []string{"location"},
    },
    func(ctx context.Context, input map[string]any) (any, error) {
        location := input["location"].(string)
        // Call weather API...
        return map[string]any{
            "temperature": 72,
            "conditions":  "sunny",
        }, nil
    },
)

// Create SDK MCP server
server := claudecode.CreateSDKMcpServer("weather", weatherTool)

// Use with client
client := claudecode.NewClient(
    claudecode.WithSdkMcpServer(server),
)
```

## File Checkpointing

Track file changes and rewind to previous states.

### How It Works

When enabled, the SDK tracks all file modifications and associates them with user messages. You can rewind files to their state at any previous message.

```
Message 1: "Create file.txt"
    └─► file.txt created (checkpoint 1)

Message 2: "Add content to file.txt"
    └─► file.txt modified (checkpoint 2)

Message 3: "Delete file.txt"
    └─► file.txt deleted (checkpoint 3)

RewindFiles(message1UUID):
    └─► file.txt restored to checkpoint 1 state
```

### Configuration

```go
err := claudecode.WithClient(ctx, func(client claudecode.Client) error {
    // Send queries...
    if err := client.Query(ctx, "Create a config file"); err != nil {
        return err
    }

    // Capture message UUIDs from UserMessage responses
    var messageUUID string
    for msg := range client.ReceiveMessages(ctx) {
        if userMsg, ok := msg.(*claudecode.UserMessage); ok {
            messageUUID = userMsg.GetUUID()
        }
    }

    // Later, rewind to that state
    if err := client.RewindFiles(ctx, messageUUID); err != nil {
        return err
    }

    return nil
}, claudecode.WithFileCheckpointing())
```

## Session Management

Maintain isolated conversation contexts within a single connection.

### Default Session

```go
// All queries share the default session
client.Query(ctx, "Remember: x = 5")
client.Query(ctx, "What is x?")  // Knows x = 5
```

### Named Sessions

```go
// Create isolated sessions
client.QueryWithSession(ctx, "Remember: a = 1", "session-a")
client.QueryWithSession(ctx, "Remember: b = 2", "session-b")

// Each session is isolated
client.QueryWithSession(ctx, "What is a?", "session-a")  // Knows a = 1
client.QueryWithSession(ctx, "What is b?", "session-b")  // Knows b = 2
client.QueryWithSession(ctx, "What is a?", "session-b")  // Doesn't know a
```

### Use Cases

- **Multi-tenant applications**: Separate context per user
- **Parallel workflows**: Multiple independent tasks
- **A/B testing**: Compare responses with different contexts

## Dynamic Model Switching

Change the AI model during a streaming session.

```go
err := claudecode.WithClient(ctx, func(client claudecode.Client) error {
    // Start with default model
    if err := client.Query(ctx, "Quick question"); err != nil {
        return err
    }

    // Switch to more capable model for complex task
    model := "claude-opus-4-5"
    if err := client.SetModel(ctx, &model); err != nil {
        return err
    }

    if err := client.Query(ctx, "Complex analysis task"); err != nil {
        return err
    }

    // Reset to default
    if err := client.SetModel(ctx, nil); err != nil {
        return err
    }

    return nil
})
```

## Stream Diagnostics

Monitor stream health and detect issues.

### Stream Issues

```go
issues := client.GetStreamIssues()
for _, issue := range issues {
    log.Printf("Issue: %s at %v", issue.Description, issue.Timestamp)
}
```

### Stream Statistics

```go
stats := client.GetStreamStats()
log.Printf("Messages: %d", stats.MessageCount)
log.Printf("Tool requests: %d", stats.ToolRequestCount)
log.Printf("Tool results: %d", stats.ToolResultCount)
log.Printf("Duration: %v", stats.Duration)
```

### Validation

The stream validator tracks tool requests and results to detect:
- Missing tool results (tool requested but never completed)
- Orphan tool results (result without matching request)
- Incomplete streams (connection lost mid-stream)

## Programmatic Agents

Define custom agents for specialized tasks.

```go
iterator, err := claudecode.Query(ctx, "Review this code",
    claudecode.WithAgent("security-reviewer", claudecode.AgentDefinition{
        Description: "Reviews code for security vulnerabilities",
        Prompt:      "You are a security expert. Focus on OWASP top 10...",
        Tools:       []string{"Read", "Grep", "Glob"},
        Model:       claudecode.AgentModelSonnet,
    }),
)

// Multiple agents for complex workflows
iterator, err := claudecode.Query(ctx, "Analyze and improve",
    claudecode.WithAgents(map[string]claudecode.AgentDefinition{
        "analyzer": {
            Description: "Analyzes code quality",
            Prompt:      "You are a code quality expert...",
            Tools:       []string{"Read", "Grep"},
            Model:       claudecode.AgentModelSonnet,
        },
        "improver": {
            Description: "Suggests improvements",
            Prompt:      "You suggest code improvements...",
            Tools:       []string{"Read", "Write"},
            Model:       claudecode.AgentModelOpus,
        },
    }),
)
```

### Agent Models

```go
claudecode.AgentModelSonnet  // Fast, efficient
claudecode.AgentModelOpus    // Most capable
claudecode.AgentModelHaiku   // Fastest, lightweight
claudecode.AgentModelInherit // Use parent's model
```

## Structured Output

Constrain responses to specific JSON schemas.

```go
schema := map[string]any{
    "type": "object",
    "properties": map[string]any{
        "summary": map[string]any{"type": "string"},
        "score":   map[string]any{"type": "number"},
        "issues":  map[string]any{
            "type":  "array",
            "items": map[string]any{"type": "string"},
        },
    },
    "required": []string{"summary", "score"},
}

iterator, err := claudecode.Query(ctx, "Review this code",
    claudecode.WithJSONSchema(schema),
)

// Result contains structured output
for {
    msg, _ := iterator.Next(ctx)
    if result, ok := msg.(*claudecode.ResultMessage); ok {
        // result.StructuredOutput contains parsed JSON
        output := result.StructuredOutput.(map[string]any)
        summary := output["summary"].(string)
        score := output["score"].(float64)
    }
}
```
