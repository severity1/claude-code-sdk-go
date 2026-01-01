# Shared Types Context

**Context**: Shared types and interfaces used across internal packages to resolve import cycles

## Component Focus
- **Core Types** - Message types, Options struct, error types shared across packages  
- **Import Cycle Resolution** - Neutral location for types that both main and internal packages need
- **Interface Contracts** - Shared contracts without circular dependencies

## Package Purpose
This package exists solely to break the circular dependency between the main package and internal packages like subprocess, parser, and cli. It contains only the minimal set of types that need to be shared.

## Design Principles
- **Minimal Surface Area** - Only types that cause import cycles
- **No Business Logic** - Pure data types and interfaces only
- **No External Dependencies** - Keep imports minimal
- **Stable API** - Changes here affect multiple packages

## Types Included
- **Options** - Configuration struct used by CLI, subprocess, and main package
- **Message Types** - User/Assistant/System/Result messages used by parser and main
- **Error Types** - SDK errors used by subprocess, CLI, and main package  
- **StreamMessage** - Stream communication type used by transport layer

## Usage Pattern
Internal packages import from here instead of main package:
```go
import "github.com/severity1/claude-agent-sdk-go/internal/shared"
```

Main package re-exports types for public API compatibility:
```go  
type Options = shared.Options
```

This maintains public API while breaking import cycles.