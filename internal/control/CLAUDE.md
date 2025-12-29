# Control Protocol Context

**Context**: SDK control protocol for bidirectional communication with Claude CLI, enabling hooks, permissions, and MCP features.

## Component Focus
- **Control Message Types** - SDKControlRequest, SDKControlResponse, and subtypes matching Python SDK
- **Request/Response Correlation** - Channel-based correlation using unique request IDs
- **Initialize Handshake** - 60-second timeout handshake for streaming mode
- **Message Routing** - Discriminate control vs regular messages

## Package Purpose

This package implements the control protocol infrastructure that enables:
- Tool permission callbacks (Issue #8)
- Hook callbacks (Issue #9)
- MCP message routing (Issue #7)
- Runtime permission mode changes
- Graceful interrupts via protocol

## Key Design Patterns

### Request ID Format
```go
// Format: req_{counter}_{random_hex}
// Example: req_1_a1b2c3d4
func (p *Protocol) generateRequestID() string
```

### Channel-Based Correlation
```go
// Python uses Events, Go uses channels
pendingRequests map[string]chan *ControlResponse
```

### Control Message Types
```go
const (
    MessageTypeControlRequest  = "control_request"
    MessageTypeControlResponse = "control_response"
)

const (
    SubtypeInterrupt         = "interrupt"
    SubtypeCanUseTool        = "can_use_tool"
    SubtypeInitialize        = "initialize"
    SubtypeSetPermissionMode = "set_permission_mode"
    SubtypeHookCallback      = "hook_callback"
    SubtypeMcpMessage        = "mcp_message"
)
```

## Python SDK Parity

This package maintains 100% behavioral parity with the Python SDK's control protocol:

| Feature | Python | Go |
|---------|--------|-----|
| Request ID format | `req_{counter}_{hex}` | Same |
| Initialize timeout | 60 seconds | Same |
| Response correlation | `anyio.Event` | `chan *ControlResponse` |
| Message routing | Type-based switch | Type-based switch |

## Integration Points

### Parser Integration
Control messages are passed through as `RawControlMessage`:
```go
case shared.MessageTypeControlRequest, shared.MessageTypeControlResponse:
    return &shared.RawControlMessage{MessageType: msgType, Data: data}, nil
```

### Transport Interface
Protocol uses a minimal Transport interface:
```go
type Transport interface {
    Write(ctx context.Context, data []byte) error
    Read(ctx context.Context) <-chan []byte
    Close() error
}
```

## Thread Safety

- All Protocol methods are thread-safe via `sync.Mutex`
- Request correlation handles concurrent requests
- Mock transport in tests is thread-safe

## Future Extensions

Issues that build on this foundation:
- **Issue #7**: MCP message routing via `SubtypeMcpMessage`
- **Issue #8**: Permission callbacks via `SubtypeCanUseTool`
- **Issue #9**: Hook callbacks via `SubtypeHookCallback`

Each will add handlers to the Protocol struct without modifying core correlation logic.
