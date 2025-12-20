package claudecode

import (
	"github.com/severity1/claude-code-sdk-go/internal/shared"
)

// Options contains configuration for Claude Code CLI interactions.
type Options = shared.Options

// PermissionMode defines the permission handling mode.
type PermissionMode = shared.PermissionMode

// McpServerType defines the type of MCP server.
type McpServerType = shared.McpServerType

// McpServerConfig represents an MCP server configuration.
type McpServerConfig = shared.McpServerConfig

// McpStdioServerConfig represents a stdio MCP server configuration.
type McpStdioServerConfig = shared.McpStdioServerConfig

// McpSSEServerConfig represents an SSE MCP server configuration.
type McpSSEServerConfig = shared.McpSSEServerConfig

// McpHTTPServerConfig represents an HTTP MCP server configuration.
type McpHTTPServerConfig = shared.McpHTTPServerConfig

// SdkBeta represents a beta feature identifier.
type SdkBeta = shared.SdkBeta

// Re-export constants
const (
	PermissionModeDefault           = shared.PermissionModeDefault
	PermissionModeAcceptEdits       = shared.PermissionModeAcceptEdits
	PermissionModePlan              = shared.PermissionModePlan
	PermissionModeBypassPermissions = shared.PermissionModeBypassPermissions
	McpServerTypeStdio              = shared.McpServerTypeStdio
	McpServerTypeSSE                = shared.McpServerTypeSSE
	McpServerTypeHTTP               = shared.McpServerTypeHTTP
	SdkBetaContext1M                = shared.SdkBetaContext1M
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

// WithFallbackModel sets the fallback model when primary model is unavailable.
func WithFallbackModel(model string) Option {
	return func(o *Options) {
		o.FallbackModel = &model
	}
}

// WithMaxBudgetUSD sets the maximum budget in USD for API usage.
func WithMaxBudgetUSD(budget float64) Option {
	return func(o *Options) {
		o.MaxBudgetUSD = &budget
	}
}

// WithUser sets the user identifier for tracking and billing.
func WithUser(user string) Option {
	return func(o *Options) {
		o.User = &user
	}
}

// WithMaxBufferSize sets the maximum buffer size for CLI output.
func WithMaxBufferSize(size int) Option {
	return func(o *Options) {
		o.MaxBufferSize = &size
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

// WithEnv sets environment variables for the subprocess.
// Multiple calls to WithEnv or WithEnvVar merge the values.
// Later calls override earlier ones for the same key.
func WithEnv(env map[string]string) Option {
	return func(o *Options) {
		if o.ExtraEnv == nil {
			o.ExtraEnv = make(map[string]string)
		}
		// Merge pattern - idiomatic Go
		for k, v := range env {
			o.ExtraEnv[k] = v
		}
	}
}

// WithEnvVar sets a single environment variable for the subprocess.
// This is a convenience method for setting individual variables.
func WithEnvVar(key, value string) Option {
	return func(o *Options) {
		if o.ExtraEnv == nil {
			o.ExtraEnv = make(map[string]string)
		}
		o.ExtraEnv[key] = value
	}
}

// WithBetas sets the SDK beta features to enable.
// See https://docs.anthropic.com/en/api/beta-headers
func WithBetas(betas ...SdkBeta) Option {
	return func(o *Options) {
		o.Betas = betas
	}
}

const customTransportMarker = "custom_transport"

// WithTransport sets a custom transport for testing.
// Since Transport is not part of Options struct, this is handled in client creation.
func WithTransport(_ Transport) Option {
	return func(o *Options) {
		// This will be handled in client implementation
		// For now, we'll use a special marker in ExtraArgs
		if o.ExtraArgs == nil {
			o.ExtraArgs = make(map[string]*string)
		}
		marker := customTransportMarker
		o.ExtraArgs["__transport_marker__"] = &marker
	}
}

// NewOptions creates Options with default values using functional options pattern.
func NewOptions(opts ...Option) *Options {
	// Create options with defaults from shared package
	options := shared.NewOptions()

	// Apply functional options
	for _, opt := range opts {
		opt(options)
	}

	return options
}
