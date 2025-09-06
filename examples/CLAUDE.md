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
**Pattern**: Least privilege with wildcard tool restrictions, read-only database operations (SELECT only), multi-service integration with explicit boundaries

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

## Example Directory Structure

### Available Examples
- `quickstart/` - Basic Query and Client API usage
- `client_streaming/` - Real-time streaming patterns  
- `client_multi_turn/` - Context preservation across interactions
- `client_advanced/` - Error handling, retries, production patterns
- `query_with_tools/` - File operations with Query API
- `client_with_tools/` - Interactive file workflows with Client API
- `query_with_mcp/` - MCP tool integration (AWS, databases) with Query API
- `client_with_mcp/` - Interactive MCP workflows with Client API
- `client_vs_query/` - API comparison and selection guidance
- `tools_comparison/` - Tool usage patterns and performance analysis

### Integration Requirements
- All examples must demonstrate proper resource cleanup
- Context-first design throughout all patterns
- Tool security restrictions as primary examples
- Go-native concurrency patterns where applicable