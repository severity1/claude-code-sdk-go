# Claude Code SDK for Go

Go SDK for programmatic interaction with Claude Code CLI. Provides streaming query capabilities with 100% API parity to the [Python SDK](https://docs.anthropic.com/en/docs/claude-code/sdk).

**🚧 Early Development**: 14/181 tasks complete (8%). API may change during development.

## Installation

```bash
go get github.com/severity1/claude-code-sdk-go
```

**Prerequisites:** Go 1.18+, Node.js, Claude Code (`npm install -g @anthropic-ai/claude-code`)

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "io"
    "log"

    "github.com/severity1/claude-code-sdk-go"
)

func main() {
    ctx := context.Background()
    
    // Streaming query (equivalent to Python's: async for message in query(...))
    iterator, err := claudecode.Query(ctx, "What is 2 + 2?")
    if err != nil {
        log.Fatal(err)
    }
    defer iterator.Close()
    
    for {
        message, err := iterator.Next(ctx)
        if err == io.EOF {
            break
        }
        if err != nil {
            log.Fatal(err)
        }
        
        fmt.Printf("%+v\n", message)
    }
}
```

## Development Progress

Following Test-Driven Development (TDD) methodology with [Python SDK](https://docs.anthropic.com/en/docs/claude-code/sdk) as reference:

- **Phase 1**: Foundation Types & Errors (14/34 tasks - 41%)
- **Overall**: 14/181 tasks complete (8%)

See [TDD_IMPLEMENTATION_TASKS.md](TDD_IMPLEMENTATION_TASKS.md) for detailed progress tracking.

## License

MIT
