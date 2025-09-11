package interfaces

// McpServerType represents the type of MCP server.
type McpServerType string

// McpServerConfig represents MCP server configuration.
// Uses Type() method instead of GetType() for consistency.
type McpServerConfig interface {
	Type() McpServerType
}
