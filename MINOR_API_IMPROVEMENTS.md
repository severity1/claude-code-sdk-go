# Minor API Improvements Analysis

**Status**: Analysis Complete
**Date**: 2025-09-14
**Branch**: `minor-api-improvements`
**Impact**: Low - Non-breaking improvements to API clarity

## Executive Summary

This document analyzes one minor issue in the Claude Code SDK Go public API that could be improved for better clarity and adherence to Go idioms. The current API is functional and well-designed, but this refinement would enhance developer experience and prevent potential confusion.

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

## Implementation Priority

**Priority**: Fix variadic sessionID (Option B - separate methods)
- Prevents user confusion
- Clear API contract
- Easy to implement

## Impact Assessment

### Breaking Changes
- This change would break existing code
- However, the SDK is young (v0.x) and can accept breaking changes
- Clear migration path can be provided

### Benefits
- Clearer API contract
- Better adherence to Go idioms
- Prevention of subtle bugs
- Improved developer experience

## Conclusion

The variadic `sessionID` parameter represents a legitimate API design concern that could cause confusion for users. The "silent parameter dropping" behavior where only the first of multiple provided session IDs is used violates the principle of least surprise.

While not a critical bug, this issue should be addressed before v1.0 to avoid breaking changes later and to align with Go idioms where variadic parameters typically handle all provided values.