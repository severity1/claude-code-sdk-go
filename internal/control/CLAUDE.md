# Module: control

<!-- AUTO-MANAGED: module-description -->
## Purpose

Bidirectional control protocol for Claude CLI communication. Manages request/response correlation, permission callbacks, lifecycle hooks, and SDK MCP server integration.

<!-- END AUTO-MANAGED -->

<!-- AUTO-MANAGED: architecture -->
## Module Architecture

```
control/
├── protocol.go           # Protocol struct, Initialize, SendControlRequest
├── protocol_test.go      # Protocol unit tests
├── protocol_hook_test.go # Hook system tests
├── protocol_mcp_test.go  # MCP server tests
├── types.go              # Request/Response types, Initialize handshake
├── types_hook.go         # Hook event types, HookMatcher, HookCallback
└── types_hook_test.go    # Hook type tests
```

**Protocol Flow**:
1. `Initialize()`: Handshake with CLI, negotiate capabilities
2. `SendControlRequest()`: Send JSON-RPC style requests with correlation IDs
3. `HandleIncomingMessage()`: Route responses to pending requests
4. Hook/Permission callbacks: Invoked on tool use events

<!-- END AUTO-MANAGED -->

<!-- AUTO-MANAGED: conventions -->
## Module-Specific Conventions

- Request correlation: Use unique request IDs for response matching
- Thread safety: All state access protected by mutex
- Timeout handling: Default 60s init timeout, configurable via `WithInitTimeout`
- Hook registration: `RegisterHook()` returns callback ID for later removal

<!-- END AUTO-MANAGED -->

<!-- AUTO-MANAGED: dependencies -->
## Key Dependencies

- `control.Transport`: Interface for stdin/stdout communication (implemented by subprocess)
- `crypto/rand`: Generate unique request IDs
- `sync`: Mutex for thread-safe state management

<!-- END AUTO-MANAGED -->

<!-- MANUAL -->
## Notes

<!-- END MANUAL -->
