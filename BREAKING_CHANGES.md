# Breaking Changes and Migration Guide

**Version**: v0.x → v1.0
**Status**: Planned Breaking Changes
**Impact**: API Contract Changes Requiring Code Updates

## Overview

This document outlines breaking changes planned for v1.0 release of the Claude Code SDK for Go. These changes address critical API design issues that could cause data loss, resource leaks, and confusion in production environments.

## Critical Breaking Changes (P1 - Immediate)

### 1. Query Method Signature Change

**Issue**: Variadic sessionID parameter causes silent data loss

#### Before (v0.x)
```go
func (c *ClientImpl) Query(ctx context.Context, prompt string, sessionID ...string) error
```

#### After (v1.0)
```go
func (c *ClientImpl) Query(ctx context.Context, prompt string) error
func (c *ClientImpl) QueryWithSession(ctx context.Context, prompt string, sessionID string) error
```

#### Migration
```go
// Before
client.Query(ctx, "Hello")                     // Still works
client.Query(ctx, "Hello", "session1")         // Needs update
client.Query(ctx, "Hello", "s1", "s2", "s3")  // BROKEN - only s1 was used!

// After
client.Query(ctx, "Hello")                     // Default session
client.QueryWithSession(ctx, "Hello", "session1") // Custom session
```

#### Automated Migration Script
```bash
# Find all Query calls with session IDs
rg 'Query\(ctx,[^,]+,[^)]+\)' --type go

# Replace with QueryWithSession (manual review recommended)
# client.Query(ctx, prompt, sessionID) → client.QueryWithSession(ctx, prompt, sessionID)
```

### 2. QueryStream Renamed to StartMessageStream

**Issue**: Method name doesn't reflect fire-and-forget behavior

#### Before (v0.x)
```go
func (c *ClientImpl) QueryStream(ctx context.Context, messages <-chan StreamMessage) error
```

#### After (v1.0)
```go
func (c *ClientImpl) StartMessageStream(ctx context.Context, messages <-chan StreamMessage) error
```

#### Migration
```go
// Before
err := client.QueryStream(ctx, msgChan)

// After
err := client.StartMessageStream(ctx, msgChan)
```

#### Behavioral Changes
- Now properly respects context cancellation
- Tracks goroutine lifecycle
- Returns error if stream already active

## High Priority Breaking Changes (P2 - Before v1.0)

### 3. Enhanced Context Cancellation

**Issue**: Goroutines don't respect context cancellation

#### Impact
All streaming operations now properly exit when context is cancelled:
```go
ctx, cancel := context.WithCancel(context.Background())
client.StartMessageStream(ctx, messages)
cancel() // Stream goroutine now exits immediately
```

### 4. Required Lifecycle Management

**Issue**: No way to clean up resources properly

#### New Requirements
```go
// MUST call Close() to prevent goroutine leaks
client := claudecode.NewClient()
defer client.Close() // Required in v1.0

// Close() now:
// - Cancels all active streams
// - Waits for goroutines to exit
// - Cleans up transport resources
```

### 5. Connection State API

**Issue**: No way to check if client is ready

#### New Methods
```go
// Check connection before operations
if !client.IsConnected() {
    if err := client.WaitForReady(ctx); err != nil {
        return err
    }
}
```

## Medium Priority Breaking Changes (P3)

### 6. ReceiveResponse Deprecated

**Issue**: Redundant API with ReceiveMessages

#### Before (v0.x)
```go
// Two ways to do the same thing
messages := client.ReceiveMessages(ctx)
iterator := client.ReceiveResponse(ctx)
```

#### After (v1.0)
```go
// Only one idiomatic way
messages := client.ReceiveMessages(ctx)
for msg := range messages {
    // Process message
}
```

#### Migration for Iterator Users
```go
// If you need iterator pattern, create your own:
type MessageIterator struct {
    messages <-chan Message
}

func (it *MessageIterator) Next(ctx context.Context) (Message, error) {
    select {
    case msg, ok := <-it.messages:
        if !ok {
            return nil, io.EOF
        }
        return msg, nil
    case <-ctx.Done():
        return nil, ctx.Err()
    }
}

// Usage
iterator := &MessageIterator{messages: client.ReceiveMessages(ctx)}
```

## Migration Timeline

### Phase 1: v0.9.0 (Deprecation Warnings)
- Add deprecation notices to affected methods
- Introduce new methods alongside old ones
- Update documentation with migration guide

### Phase 2: v0.10.0 (Parallel APIs)
- New methods become primary
- Old methods marked as deprecated
- Emit runtime warnings for deprecated usage

### Phase 3: v1.0.0 (Breaking Changes)
- Remove all deprecated methods
- Enforce new API contract
- Release with major version bump

## Compatibility Package

For gradual migration, a compatibility package will be provided:

```go
import "github.com/severity1/claude-code-sdk-go/compat/v0"

// Provides v0 API wrapping v1 implementation
client := v0.NewClient()
client.Query(ctx, "Hello", "session") // Still works but deprecated
```

## Testing Your Migration

### Validation Script
```go
// migration_test.go
package main

import (
    "context"
    "testing"
    claudecode "github.com/severity1/claude-code-sdk-go"
)

func TestMigration(t *testing.T) {
    client := claudecode.NewClient()
    defer client.Close() // Required in v1.0

    ctx := context.Background()

    // Test new API
    if err := client.Query(ctx, "test"); err != nil {
        t.Errorf("Query failed: %v", err)
    }

    if err := client.QueryWithSession(ctx, "test", "session1"); err != nil {
        t.Errorf("QueryWithSession failed: %v", err)
    }

    // Verify connection state
    if !client.IsConnected() {
        t.Error("Client should be connected")
    }
}
```

## Common Migration Issues

### Issue: Goroutine Leaks After Migration
**Solution**: Ensure `client.Close()` is called with `defer`

### Issue: Context Cancellation Not Working
**Solution**: Update to use new context-aware stream methods

### Issue: Multiple Session IDs Being Passed
**Solution**: Review code for multi-session patterns - these were never supported

### Issue: Iterator Pattern Dependencies
**Solution**: Use provided iterator wrapper or switch to channel-based approach

## Support and Resources

- **Migration Guide**: This document
- **Examples**: See `/examples/migration/` directory
- **Support**: File issues at github.com/severity1/claude-code-sdk-go/issues
- **Compatibility Package**: Available for 6 months after v1.0 release

## Summary

These breaking changes are necessary to:
1. **Prevent data loss** from silent parameter dropping
2. **Fix resource leaks** from untracked goroutines
3. **Improve API clarity** with better naming
4. **Ensure reliability** with proper lifecycle management
5. **Reduce confusion** by eliminating redundant APIs

While breaking changes are disruptive, these fixes address critical production issues and establish a stable API for long-term use.