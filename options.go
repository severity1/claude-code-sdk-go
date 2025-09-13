package claudecode

import (
	"github.com/severity1/claude-code-sdk-go/internal/shared"
	"github.com/severity1/claude-code-sdk-go/pkg/interfaces"
)

// Options contains configuration for Claude Code CLI interactions.
type Options = interfaces.Options

// PermissionMode defines the permission handling mode.
type PermissionMode = interfaces.PermissionMode

// McpServerType defines the type of MCP server.
type McpServerType = interfaces.McpServerType

// McpServerConfig represents an MCP server configuration.
type McpServerConfig = interfaces.McpServerConfig

// McpStdioServerConfig represents a stdio MCP server configuration.
type McpStdioServerConfig = interfaces.McpStdioServerConfig

// McpSSEServerConfig represents an SSE MCP server configuration.
type McpSSEServerConfig = interfaces.McpSSEServerConfig

// McpHTTPServerConfig represents an HTTP MCP server configuration.
type McpHTTPServerConfig = interfaces.McpHTTPServerConfig

// Re-export constants
const (
	PermissionModeDefault           = interfaces.PermissionModeDefault
	PermissionModeAcceptEdits       = interfaces.PermissionModeAcceptEdits
	PermissionModePlan              = interfaces.PermissionModePlan
	PermissionModeBypassPermissions = interfaces.PermissionModeBypassPermissions
	McpServerTypeStdio              = interfaces.McpServerTypeStdio
	McpServerTypeSSE                = interfaces.McpServerTypeSSE
	McpServerTypeHTTP               = interfaces.McpServerTypeHTTP
	DefaultMaxThinkingTokens        = interfaces.DefaultMaxThinkingTokens
)

// Option configures Options using the functional options pattern.
type Option func(*Options)

// WithAllowedTools sets the allowed tools list.
func WithAllowedTools(tools ...string) Option {
	return func(o *Options) {
		o.AllowedTools = tools
	}
}

// WithDisallowedTools sets the disallowed tools list.
func WithDisallowedTools(tools ...string) Option {
	return func(o *Options) {
		o.DisallowedTools = tools
	}
}

// WithSystemPrompt sets the system prompt.
func WithSystemPrompt(prompt string) Option {
	return func(o *Options) {
		o.SystemPrompt = &prompt
	}
}

// WithAppendSystemPrompt sets the append system prompt.
func WithAppendSystemPrompt(prompt string) Option {
	return func(o *Options) {
		o.AppendSystemPrompt = &prompt
	}
}

// WithModel sets the model to use.
func WithModel(model string) Option {
	return func(o *Options) {
		o.Model = &model
	}
}

// WithMaxThinkingTokens sets the maximum thinking tokens.
func WithMaxThinkingTokens(tokens int) Option {
	return func(o *Options) {
		o.MaxThinkingTokens = tokens
	}
}

// WithPermissionMode sets the permission mode.
func WithPermissionMode(mode PermissionMode) Option {
	return func(o *Options) {
		o.PermissionMode = &mode
	}
}

// WithPermissionPromptToolName sets the permission prompt tool name.
func WithPermissionPromptToolName(toolName string) Option {
	return func(o *Options) {
		o.PermissionPromptToolName = &toolName
	}
}

// WithContinueConversation enables conversation continuation.
func WithContinueConversation(continueConversation bool) Option {
	return func(o *Options) {
		o.ContinueConversation = continueConversation
	}
}

// WithResume sets the session ID to resume.
func WithResume(sessionID string) Option {
	return func(o *Options) {
		o.Resume = &sessionID
	}
}

// WithCwd sets the working directory.
func WithCwd(cwd string) Option {
	return func(o *Options) {
		o.Cwd = &cwd
	}
}

// WithAddDirs adds directories to the context.
func WithAddDirs(dirs ...string) Option {
	return func(o *Options) {
		o.AddDirs = dirs
	}
}

// WithMcpServers sets the MCP server configurations.
func WithMcpServers(servers map[string]McpServerConfig) Option {
	return func(o *Options) {
		o.McpServers = servers
	}
}

// WithMaxTurns sets the maximum number of conversation turns.
func WithMaxTurns(turns int) Option {
	return func(o *Options) {
		o.MaxTurns = turns
	}
}

// WithSettings sets the settings file path or JSON string.
func WithSettings(settings string) Option {
	return func(o *Options) {
		o.Settings = &settings
	}
}

// WithExtraArgs sets arbitrary CLI flags via ExtraArgs.
func WithExtraArgs(args map[string]*string) Option {
	return func(o *Options) {
		o.ExtraArgs = args
	}
}

// WithCLIPath sets a custom CLI path.
func WithCLIPath(path string) Option {
	return func(o *Options) {
		o.CLIPath = &path
	}
}

// WithTransport sets a custom transport for testing.
// Since Transport is not part of Options struct, this is handled in client creation.
func WithTransport(transport Transport) Option {
	return func(o *Options) {
		// This will be handled in client implementation
		// For now, we'll use a special marker in ExtraArgs
		if o.ExtraArgs == nil {
			o.ExtraArgs = make(map[string]*string)
		}
		const customTransportMarker = "custom_transport"
		marker := customTransportMarker
		o.ExtraArgs["__transport_marker__"] = &marker
	}
}

// NewOptions creates Options with default values using functional options pattern.
func NewOptions(opts ...Option) *Options {
	// Create options with defaults from interfaces package
	options := interfaces.NewOptions()

	// Apply functional options
	for _, opt := range opts {
		opt(options)
	}

	return options
}

// toSharedOptions converts interfaces.Options to shared.Options for subprocess compatibility.
// This is a temporary bridge function during the migration to new interfaces.
func toSharedOptions(o *Options) *shared.Options {
	// Convert interfaces.Options back to shared.Options for subprocess
	sharedOpts := shared.NewOptions()

	// Copy all fields from interfaces.Options to shared.Options
	sharedOpts.AllowedTools = o.AllowedTools
	sharedOpts.DisallowedTools = o.DisallowedTools
	sharedOpts.SystemPrompt = o.SystemPrompt
	sharedOpts.AppendSystemPrompt = o.AppendSystemPrompt
	sharedOpts.Model = o.Model
	sharedOpts.MaxThinkingTokens = o.MaxThinkingTokens

	// Convert PermissionMode
	if o.PermissionMode != nil {
		sharedMode := shared.PermissionMode(*o.PermissionMode)
		sharedOpts.PermissionMode = &sharedMode
	}

	sharedOpts.PermissionPromptToolName = o.PermissionPromptToolName
	sharedOpts.ContinueConversation = o.ContinueConversation
	sharedOpts.Resume = o.Resume
	sharedOpts.MaxTurns = o.MaxTurns
	sharedOpts.Settings = o.Settings
	sharedOpts.Cwd = o.Cwd
	sharedOpts.AddDirs = o.AddDirs
	sharedOpts.ExtraArgs = o.ExtraArgs
	sharedOpts.CLIPath = o.CLIPath

	// Convert MCP servers - create empty map since subprocess doesn't need them
	sharedOpts.McpServers = make(map[string]shared.McpServerConfig)

	return sharedOpts
}
