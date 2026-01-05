# Component Structure

This document describes the package organization and component responsibilities of the Claude Agent SDK for Go.

## Package Layout

```
github.com/severity1/claude-agent-sdk-go
├── (root)                  # Public API layer
│   ├── client.go           # Client interface and WithClient
│   ├── query.go            # Query API and iterators
│   ├── types.go            # Type re-exports
│   ├── options.go          # Functional options
│   ├── errors.go           # Error types and helpers
│   ├── mcp.go              # MCP server types
│   └── transport.go        # Transport interface
│
├── internal/
│   ├── cli/                # CLI discovery
│   ├── control/            # Control protocol
│   ├── parser/             # JSON parsing
│   ├── shared/             # Shared types
│   └── subprocess/         # Process management
│
└── examples/               # Usage examples (01-20)
```

## Public API Layer

The root package (`claudecode`) exposes the SDK's public interface.

### client.go

Defines the `Client` interface for bidirectional streaming communication:

```go
type Client interface {
    Connect(ctx context.Context, prompt ...StreamMessage) error
    Disconnect() error
    Query(ctx context.Context, prompt string) error
    QueryWithSession(ctx context.Context, prompt string, sessionID string) error
    ReceiveMessages(ctx context.Context) <-chan Message
    // ... additional methods
}
```

Also provides `WithClient()` - a Go-idiomatic context manager for automatic resource management.

### query.go

Implements the Query API for one-shot operations:

```go
func Query(ctx context.Context, prompt string, opts ...Option) (MessageIterator, error)
```

Returns a `MessageIterator` for processing response messages.

### types.go

Re-exports types from internal packages for public use:

- Message types: `UserMessage`, `AssistantMessage`, `SystemMessage`, `ResultMessage`
- Content blocks: `TextBlock`, `ThinkingBlock`, `ToolUseBlock`, `ToolResultBlock`
- Interfaces: `Message`, `ContentBlock`

### options.go

Functional options for configuration:

```go
claudecode.Query(ctx, "prompt",
    claudecode.WithSystemPrompt("..."),
    claudecode.WithAllowedTools("Read", "Write"),
    claudecode.WithMaxTurns(10),
)
```

### errors.go

Structured error types with helper functions:

- `CLINotFoundError` - CLI binary not found
- `ConnectionError` - Connection failed
- `MessageParseError` - JSON parsing failed

Helper functions: `AsCLINotFoundError()`, `AsConnectionError()`, `AsMessageParseError()`

## Internal Packages

### internal/cli

**Purpose**: CLI discovery and command building

**Key Files**:
- `discovery.go` - `FindCLI()`, version checking, path resolution

**Responsibilities**:
- Locate Claude CLI binary in PATH and common locations
- Validate CLI version compatibility
- Build command-line arguments from Options
- Provide installation guidance when CLI not found

### internal/control

**Purpose**: Bidirectional control protocol for advanced features

**Key Files**:
- `protocol.go` - Request/response correlation, message routing
- `types.go` - Control request/response types
- `types_hook.go` - Hook event types and callbacks

**Responsibilities**:
- Initialize handshake with CLI
- Correlate requests with responses using unique IDs
- Route permission callback requests
- Execute lifecycle hooks (PreToolUse, PostToolUse)
- Handle SDK MCP server messages

### internal/parser

**Purpose**: JSON message parsing with speculative parsing

**Key Files**:
- `json.go` - `Parser` struct, `ProcessLine()`

**Responsibilities**:
- Parse streaming JSON from CLI stdout
- Handle incomplete JSON with speculative parsing
- Discriminate message types based on `"type"` field
- Protect against buffer overflow (1MB limit)

### internal/shared

**Purpose**: Shared types used across packages

**Key Files**:
- `message.go` - Message and ContentBlock interfaces
- `errors.go` - Error type definitions
- `options.go` - Options struct
- `stream.go` - StreamIssue, StreamStats
- `validator.go` - Stream validation

**Responsibilities**:
- Define core interfaces (Message, ContentBlock)
- Define concrete message types
- Define error hierarchy
- Provide configuration structures

### internal/subprocess

**Purpose**: Subprocess management and transport layer

**Key Files**:
- `transport.go` - `Transport` struct, process lifecycle
- `protocol_adapter.go` - Adapter for control protocol

**Responsibilities**:
- Spawn and manage CLI subprocess
- Handle stdin/stdout/stderr communication
- Route messages between CLI and SDK
- Implement graceful shutdown (SIGTERM -> 5s -> SIGKILL)
- Validate message streams

## Component Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                        Application                               │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                     Public API (claudecode)                      │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────────────┐ │
│  │ Query()  │  │ Client   │  │ Options  │  │ Error Helpers    │ │
│  │          │  │ WithClient│  │ WithXxx()│  │ AsXxxError()     │ │
│  └──────────┘  └──────────┘  └──────────┘  └──────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                     Transport Interface                          │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                   internal/subprocess                            │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │ Transport: Process lifecycle, I/O streams, message routing│   │
│  └──────────────────────────────────────────────────────────┘   │
│                          │                                       │
│           ┌──────────────┴──────────────┐                       │
│           ▼                             ▼                       │
│  ┌─────────────────┐          ┌─────────────────┐               │
│  │ internal/parser │          │ internal/control│               │
│  │ JSON parsing    │          │ Control protocol│               │
│  └─────────────────┘          └─────────────────┘               │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Claude Code CLI                             │
│                    (subprocess: stdin/stdout)                    │
└─────────────────────────────────────────────────────────────────┘
```

## Examples Directory

The `examples/` directory contains 20 progressive examples:

| Range | Category | Examples |
|-------|----------|----------|
| 01-03 | Getting Started | Quickstart, streaming, multi-turn |
| 04-07 | Tool Integration | File tools, MCP servers |
| 08-10 | Production Patterns | Error handling, sessions |
| 11-14 | Security & Lifecycle | Permissions, hooks, checkpointing |
| 15-20 | Advanced | Subagents, structured output, debugging |

Each example is self-contained with its own `main.go` demonstrating specific features.
