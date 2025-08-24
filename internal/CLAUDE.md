# Internal Implementation Patterns

**Context**: Internal Go implementation for Claude Code SDK with transport abstraction and Go-native concurrency

## Component Focus
- **Transport Interface** - Abstract communication with Claude CLI subprocess
- **Go-Native Concurrency** - Goroutines, channels, context-first design
- **Resource Management** - Defer patterns, explicit cleanup, lifecycle management
- **Interface-Driven Design** - Clean abstractions, testable implementations

## Required Patterns

### Transport Abstraction
```go
type Transport interface {
    Connect(ctx context.Context) error
    SendMessage(ctx context.Context, message StreamMessage) error
    ReceiveMessages(ctx context.Context) (<-chan Message, <-chan error)
    Interrupt(ctx context.Context) error
    Close() error
}
```

### Context-First Operations
- All operations accept `context.Context` as first parameter
- Respect context cancellation and timeouts throughout
- Propagate context to all downstream operations
- Use `ctx.Done()` for early termination

### Goroutine Management
- Start goroutines for concurrent I/O operations
- Use channels for communication between goroutines
- Proper cleanup with defer and channel closing
- Avoid goroutine leaks with proper lifecycle management

### Resource Cleanup Patterns
```go
// Example resource management
func (t *Transport) Connect(ctx context.Context) error {
    // Resource acquisition
    defer func() {
        // Ensure cleanup on any exit path
        if err := t.cleanup(); err != nil {
            // Log but don't override primary error
        }
    }()
    
    // Implementation
    return nil
}
```

## Interface Design Principles
- Small, focused interfaces (like Transport)
- Composition over inheritance
- Interface compliance verified at compile time
- Enable testing through interface mocking

## Error Handling Standards
- Return errors explicitly, never panic
- Use `fmt.Errorf` with `%w` for error wrapping
- Support `errors.Is()` and `errors.As()` for error inspection
- Include contextual information in error messages

## Concurrency Safety
- Protect shared state with mutexes when necessary
- Use channels for communication, not shared memory
- Ensure goroutines can be cancelled via context
- Avoid data races with proper synchronization

## Performance Considerations
- Minimize heap allocations where possible
- Use object pooling for frequently allocated objects
- Efficient buffer management for I/O operations
- Lazy initialization for expensive resources

## Integration Requirements
- All components implement appropriate interfaces
- Context propagation throughout call chains
- Proper error handling and reporting
- Resource cleanup in all code paths