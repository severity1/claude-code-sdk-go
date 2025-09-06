# Usage Patterns Context

**Context**: Example usage patterns for Claude Code SDK Go with Query vs Client API patterns, tool integration, and Go-native concurrency

## Component Focus
- **Query API Patterns** - One-shot automation, scripting, CI/CD integration with automatic cleanup
- **Client API Patterns** - Interactive conversations, multi-turn workflows, streaming responses with persistent connections
- **Tool Integration Patterns** - Core tools (Read/Write/Edit) and MCP tools (AWS, databases) with security restrictions
- **Go-Native Concurrency** - Goroutines, channels, context patterns for SDK users
- **Error Handling Examples** - Proper error handling and resource cleanup patterns

## Required Usage Patterns

### Query API (One-Shot Operations)
**Use Cases**: Automation scripts, batch processing, CI/CD pipelines, documentation generation
**Pattern**: Resource acquisition → defer cleanup → iterator processing with proper error handling

### Query with Tool Security Restrictions
**Pattern**: Tool allowlists for file operations, MCP integration with safety controls using `WithAllowedTools()` and `WithDisallowedTools()`

### Client API (Streaming/Multi-Turn)
**Use Cases**: Interactive conversations, progressive workflows, context preservation
**Pattern**: Connect → defer cleanup → channel-based streaming with context preservation across interactions

## Tool Integration Patterns

### Core File Tools Security Model
**Pattern**: Principle of least privilege with read-only restrictions for security audits, progressive file workflows with Client API context preservation

### MCP Tool Security Patterns
**Pattern**: Least privilege with explicit tool names (no wildcards), read-only database operations (SELECT only), multi-service integration with explicit boundaries

## Required Go-Native Patterns

### Context-First Design
**Pattern**: Context as first parameter throughout, timeout/cancellation handling, context propagation to all blocking operations

### Concurrent Query Pattern
**Pattern**: Goroutines with proper closure variable capture, channel-based result collection, essential resource cleanup in concurrent operations

## Error Handling Requirements

### Structured Error Handling
**Pattern**: Type-specific error checking with `errors.As()`, error wrapping with `%w`, normal completion vs error differentiation in iterators

### Resource Cleanup Requirements
**Pattern**: Immediate defer after resource acquisition, goroutine panic recovery with resource cleanup, connection lifecycle management

## Example Directory Structure (Learning Path)

### Numbered Examples (Easiest → Hardest)
- `01_quickstart/` - Basic Query API usage, message handling
- `02_client_streaming/` - Real-time streaming patterns with Client API
- `03_client_multi_turn/` - Context preservation across interactions
- `04_query_with_tools/` - File operations with Query API
- `05_client_with_tools/` - Interactive file workflows with Client API
- `06_query_with_mcp/` - MCP tool integration (AWS) with Query API
- `07_client_with_mcp/` - Interactive MCP workflows with Client API
- `08_client_advanced/` - Error handling, retries, production patterns
- `09_client_vs_query/` - API comparison and selection guidance

### Integration Requirements
- All examples must demonstrate proper resource cleanup
- Context-first design throughout all patterns
- Tool security restrictions as primary examples
- Go-native concurrency patterns where applicable
- Explicit MCP tool names (no wildcards) for security compliance
- Progressive complexity from basic queries to advanced cloud workflows