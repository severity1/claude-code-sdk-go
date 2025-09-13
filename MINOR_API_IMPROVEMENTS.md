# Minor API Improvements Analysis

**Status**: Analysis Complete
**Date**: 2025-09-14
**Branch**: `minor-api-improvements`
**Impact**: Low - Non-breaking improvements to API clarity

## Executive Summary

This document analyzes three minor issues in the Claude Code SDK Go public API that, while not critical, could be improved for better clarity, idiomaticity, and adherence to Go best practices. The current API is functional and well-designed, but these refinements would enhance developer experience and prevent potential confusion.

## Issue 1: Variadic sessionID Parameter - Ambiguous API Contract

### Current Implementation
```go
// client.go:261
func (c *ClientImpl) Query(ctx context.Context, prompt string, sessionID ...string) error
```

### Problem Analysis

The variadic `sessionID` parameter creates an ambiguous API contract with several issues:

1. **Unclear semantics**: The variadic syntax (`...string`) suggests the method accepts multiple session IDs, but the implementation only uses the first one:
   ```go
   // client.go:283-286
   sid := defaultSessionID
   if len(sessionID) > 0 && sessionID[0] != "" {
       sid = sessionID[0]
   }
   ```

2. **Silent parameter dropping**: If a user provides multiple session IDs, all except the first are silently ignored without warning or error.

3. **Go idiom violation**: In idiomatic Go, variadic parameters should handle all provided values, not just the first. Examples from stdlib:
   - `fmt.Printf(format, args...)` - uses all args
   - `append(slice, elems...)` - appends all elements
   - `path.Join(elems...)` - joins all path elements

4. **Confusing user experience**:
   ```go
   // What happens here? Only "session1" is used, others ignored
   client.Query(ctx, "Hello", "session1", "session2", "session3")
   ```

### Proposed Solutions

#### Option A: Explicit Optional via Pointer
```go
func (c *ClientImpl) Query(ctx context.Context, prompt string, sessionID *string) error {
    sid := defaultSessionID
    if sessionID != nil && *sessionID != "" {
        sid = *sessionID
    }
    // ...
}

// Usage:
client.Query(ctx, "Hello", nil)           // Use default session
sessionID := "custom"
client.Query(ctx, "Hello", &sessionID)    // Use custom session
```

#### Option B: Separate Methods (Recommended)
```go
func (c *ClientImpl) Query(ctx context.Context, prompt string) error {
    return c.queryWithSession(ctx, prompt, defaultSessionID)
}

func (c *ClientImpl) QueryWithSession(ctx context.Context, prompt string, sessionID string) error {
    return c.queryWithSession(ctx, prompt, sessionID)
}

// Internal helper
func (c *ClientImpl) queryWithSession(ctx context.Context, prompt string, sessionID string) error {
    // Current implementation
}

// Usage:
client.Query(ctx, "Hello")                      // Clear: uses default
client.QueryWithSession(ctx, "Hello", "custom") // Clear: uses custom
```

#### Option C: Options Struct (Most Extensible)
```go
type QueryOptions struct {
    SessionID string
    // Future fields can be added here
}

func (c *ClientImpl) Query(ctx context.Context, prompt string, opts *QueryOptions) error {
    sid := defaultSessionID
    if opts != nil && opts.SessionID != "" {
        sid = opts.SessionID
    }
    // ...
}

// Usage:
client.Query(ctx, "Hello", nil)                           // Default
client.Query(ctx, "Hello", &QueryOptions{SessionID: "custom"}) // Custom
```

### Recommendation

**Option B (Separate Methods)** is recommended because:
- Follows Go stdlib patterns (e.g., `http.Get()` vs `http.NewRequestWithContext()`)
- Clear intent without ambiguity
- No pointer dereferencing needed
- Backwards compatible if we keep the variadic version deprecated

## Issue 2: QueryStream Naming - Misleading Method Purpose

### Current Implementation
```go
// client.go:304
func (c *ClientImpl) QueryStream(ctx context.Context, messages <-chan StreamMessage) error
```

### Problem Analysis

1. **Naming inconsistency**: "Query" implies a request-response pattern, but `QueryStream` is fire-and-forget:
   ```go
   // client.go:315-333
   go func() {
       for {
           select {
           case msg, ok := <-messages:
               // ... sends messages asynchronously
           }
       }
   }()
   return nil // Returns immediately!
   ```

2. **Different operation types with similar names**:
   - `Query()` - Sends one message, synchronous error handling
   - `QueryStream()` - Sends many messages, asynchronous, no error feedback

3. **Goroutine leak risk**: The spawned goroutine (line 316) has no cleanup tracking. If the transport fails, the goroutine may leak.

4. **No error propagation**: Send errors in the goroutine are silently swallowed:
   ```go
   // client.go:323-326
   if err := transport.SendMessage(ctx, msg); err != nil {
       // Log error but continue processing
       return // Just returns, error is lost
   }
   ```

### Proposed Solutions

#### Option A: Rename to Clarify Intent
```go
// Makes it clear this is for continuous sending
func (c *ClientImpl) SendStream(ctx context.Context, messages <-chan StreamMessage) error

// Or more explicit
func (c *ClientImpl) StreamMessages(ctx context.Context, messages <-chan StreamMessage) error
```

#### Option B: Make It Blocking with Error Channel (Safer)
```go
func (c *ClientImpl) SendMessages(ctx context.Context, messages <-chan StreamMessage) <-chan error {
    errChan := make(chan error, 1)

    go func() {
        defer close(errChan)
        for {
            select {
            case msg, ok := <-messages:
                if !ok {
                    return
                }
                if err := c.transport.SendMessage(ctx, msg); err != nil {
                    errChan <- err
                    return
                }
            case <-ctx.Done():
                errChan <- ctx.Err()
                return
            }
        }
    }()

    return errChan
}
```

### Recommendation

**Rename to `SendStream()`** and document the fire-and-forget behavior clearly. The current implementation is valid for streaming use cases, but the name should reflect what it actually does.

## Issue 3: ReceiveResponse vs ReceiveMessages - Redundant APIs

### Current Implementation
```go
// client.go:336
func (c *ClientImpl) ReceiveMessages(ctx context.Context) <-chan Message

// client.go:355
func (c *ClientImpl) ReceiveResponse(ctx context.Context) MessageIterator
```

### Problem Analysis

1. **Same underlying data source**: Both methods read from the same channel:
   ```go
   // ReceiveMessages - line 341
   return msgChan

   // ReceiveResponse - returns iterator wrapping same msgChan
   ```

2. **Iterator adds no value**: The `MessageIterator` just wraps the channel with a `Next()` method, which Go developers can easily do with `range`:
   ```go
   // Using channel (idiomatic Go)
   for msg := range client.ReceiveMessages(ctx) {
       // process
   }

   // Using iterator (unnecessary abstraction)
   iter := client.ReceiveResponse(ctx)
   for {
       msg, err := iter.Next(ctx)
       if err != nil { break }
       // process
   }
   ```

3. **Confuses API users**: Two methods for the same thing creates decision paralysis.

4. **Maintenance burden**: Two APIs to document, test, and maintain.

5. **Against Go philosophy**: Go prefers one obvious way to do things. Channels ARE Go's iterators.

### Proposed Solution

Remove `ReceiveResponse()` and keep only `ReceiveMessages()`:

```go
// Keep only the channel-based API
func (c *ClientImpl) ReceiveMessages(ctx context.Context) <-chan Message

// Remove ReceiveResponse - users can create their own iterator if needed
// Delete: func (c *ClientImpl) ReceiveResponse(ctx context.Context) MessageIterator
```

### Rationale

- **Go stdlib precedent**:
  - `time.Tick()` returns `<-chan Time`, not an iterator
  - `signal.Notify()` uses channels
  - Database drivers use `rows.Next()` but that's for SQL result sets, not message streams

- **Channels are sufficient**: Range loops over channels are idiomatic and well-understood

- **Simpler API**: One way to receive messages reduces cognitive load

### Migration Path

If we must maintain backwards compatibility:
```go
// Deprecate but keep for compatibility
// Deprecated: Use ReceiveMessages() instead
func (c *ClientImpl) ReceiveResponse(ctx context.Context) MessageIterator {
    return &channelIterator{
        ctx:     ctx,
        msgChan: c.ReceiveMessages(ctx),
    }
}
```

## Implementation Priority

1. **High Priority**: Fix variadic sessionID (Option B - separate methods)
   - Prevents user confusion
   - Clear API contract
   - Easy to implement

2. **Medium Priority**: Rename QueryStream to SendStream
   - Simple rename
   - Better describes actual behavior
   - Update documentation

3. **Low Priority**: Remove ReceiveResponse
   - Can deprecate first
   - Less critical since both work
   - Give users time to migrate

## Impact Assessment

### Breaking Changes
- These changes would break existing code
- However, the SDK is young (v0.x) and can accept breaking changes
- Clear migration path can be provided

### Benefits
- Clearer API contracts
- Better adherence to Go idioms
- Reduced API surface area
- Prevention of subtle bugs
- Improved developer experience

## Conclusion

While the current API is functional and well-designed overall, these minor improvements would:
1. Eliminate ambiguity in the API contract
2. Better follow Go stdlib patterns
3. Reduce redundancy
4. Prevent potential bugs (parameter dropping, goroutine leaks)

The changes are minor but would significantly improve the clarity and idiomaticity of the SDK's public API.