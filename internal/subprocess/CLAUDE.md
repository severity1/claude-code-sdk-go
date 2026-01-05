# Module: subprocess

<!-- AUTO-MANAGED: module-description -->
## Purpose

Subprocess management and transport layer. Spawns Claude CLI process, manages stdin/stdout communication, and implements the `Transport` interface for message passing.

<!-- END AUTO-MANAGED -->

<!-- AUTO-MANAGED: architecture -->
## Module Architecture

```
subprocess/
├── transport.go          # Transport struct, Start, Send, Receive
├── transport_test.go     # Transport tests
├── protocol_adapter.go   # ProtocolAdapter for control.Transport interface
└── protocol_adapter_test.go # Adapter tests
```

**Transport Flow**:
1. `Start()`: Spawn CLI subprocess with configured arguments
2. `Send()`: Write JSON messages to stdin
3. `handleStdout()`: Read stdout, parse JSON, route messages
4. Control messages: Route to `control.Protocol.HandleIncomingMessage()`
5. `Close()`: SIGTERM -> wait 5s -> SIGKILL

<!-- END AUTO-MANAGED -->

<!-- AUTO-MANAGED: conventions -->
## Module-Specific Conventions

- Graceful shutdown: SIGTERM with 5s grace period before SIGKILL
- Message routing: Distinguish control vs regular messages by type
- Protocol adapter: Bridges subprocess stdin to `control.Transport` interface
- Resource cleanup: Always close stdin before waiting for process exit

<!-- END AUTO-MANAGED -->

<!-- AUTO-MANAGED: dependencies -->
## Key Dependencies

- `internal/parser`: JSON message parsing
- `internal/control`: Control protocol for hooks/permissions
- `os/exec`: Subprocess management
- `bufio`: Line-by-line stdout reading

<!-- END AUTO-MANAGED -->

<!-- MANUAL -->
## Notes

<!-- END MANUAL -->
