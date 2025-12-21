package claudecode

import (
	"io"
	"os"

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

// ToolsPreset represents a preset tools configuration.
type ToolsPreset = shared.ToolsPreset

// SettingSource represents a settings source location.
type SettingSource = shared.SettingSource

// SandboxSettings configures sandbox behavior for bash command execution.
type SandboxSettings = shared.SandboxSettings

// SandboxNetworkConfig configures network access within sandbox.
type SandboxNetworkConfig = shared.SandboxNetworkConfig

// SandboxIgnoreViolations specifies patterns to ignore during sandbox violations.
type SandboxIgnoreViolations = shared.SandboxIgnoreViolations

// SdkPluginType represents the type of SDK plugin.
type SdkPluginType = shared.SdkPluginType

// SdkPluginConfig represents a plugin configuration.
type SdkPluginConfig = shared.SdkPluginConfig

// OutputFormat specifies the format for structured output.
type OutputFormat = shared.OutputFormat

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
	SettingSourceUser               = shared.SettingSourceUser
	SettingSourceProject            = shared.SettingSourceProject
	SettingSourceLocal              = shared.SettingSourceLocal
	SdkPluginTypeLocal              = shared.SdkPluginTypeLocal
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

// WithTools sets available tools as a list of tool names.
func WithTools(tools ...string) Option {
	return func(o *Options) {
		o.Tools = tools
	}
}

// WithToolsPreset sets tools to a preset configuration.
func WithToolsPreset(preset string) Option {
	return func(o *Options) {
		o.Tools = ToolsPreset{
			Type:   "preset",
			Preset: preset,
		}
	}
}

// WithClaudeCodeTools sets tools to the claude_code preset.
func WithClaudeCodeTools() Option {
	return WithToolsPreset("claude_code")
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

// WithForkSession enables forking to a new session ID when resuming.
// When true, resumed sessions fork to a new session ID rather than
// continuing the previous session.
func WithForkSession(fork bool) Option {
	return func(o *Options) {
		o.ForkSession = fork
	}
}

// WithSettingSources sets which settings sources to load.
// Valid sources are SettingSourceUser, SettingSourceProject, and SettingSourceLocal.
func WithSettingSources(sources ...SettingSource) Option {
	return func(o *Options) {
		o.SettingSources = sources
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

// WithSandbox sets the sandbox settings for bash command isolation.
func WithSandbox(sandbox *SandboxSettings) Option {
	return func(o *Options) {
		o.Sandbox = sandbox
	}
}

// WithSandboxEnabled enables or disables sandbox.
// If sandbox settings don't exist, they are initialized.
func WithSandboxEnabled(enabled bool) Option {
	return func(o *Options) {
		if o.Sandbox == nil {
			o.Sandbox = &SandboxSettings{}
		}
		o.Sandbox.Enabled = enabled
	}
}

// WithAutoAllowBashIfSandboxed sets whether to auto-approve bash when sandboxed.
// If sandbox settings don't exist, they are initialized.
func WithAutoAllowBashIfSandboxed(autoAllow bool) Option {
	return func(o *Options) {
		if o.Sandbox == nil {
			o.Sandbox = &SandboxSettings{}
		}
		o.Sandbox.AutoAllowBashIfSandboxed = autoAllow
	}
}

// WithSandboxExcludedCommands sets commands that always bypass sandbox.
// If sandbox settings don't exist, they are initialized.
func WithSandboxExcludedCommands(commands ...string) Option {
	return func(o *Options) {
		if o.Sandbox == nil {
			o.Sandbox = &SandboxSettings{}
		}
		o.Sandbox.ExcludedCommands = commands
	}
}

// WithSandboxNetwork sets the network configuration for sandbox.
// If sandbox settings don't exist, they are initialized.
func WithSandboxNetwork(network *SandboxNetworkConfig) Option {
	return func(o *Options) {
		if o.Sandbox == nil {
			o.Sandbox = &SandboxSettings{}
		}
		o.Sandbox.Network = network
	}
}

// WithPlugins sets the plugin configurations.
// This replaces any previously configured plugins.
func WithPlugins(plugins []SdkPluginConfig) Option {
	return func(o *Options) {
		o.Plugins = plugins
	}
}

// WithPlugin appends a single plugin configuration.
// Multiple calls accumulate plugins.
func WithPlugin(plugin SdkPluginConfig) Option {
	return func(o *Options) {
		o.Plugins = append(o.Plugins, plugin)
	}
}

// WithLocalPlugin appends a local plugin by path.
// This is a convenience method for the common case of local plugins.
func WithLocalPlugin(path string) Option {
	return func(o *Options) {
		o.Plugins = append(o.Plugins, SdkPluginConfig{
			Type: SdkPluginTypeLocal,
			Path: path,
		})
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

// WithDebugWriter sets the writer for CLI debug output.
// If not set, stderr is isolated to a temporary file (default behavior).
// Common values: os.Stderr, io.Discard, or a custom io.Writer like bytes.Buffer.
func WithDebugWriter(w io.Writer) Option {
	return func(o *Options) {
		o.DebugWriter = w
	}
}

// WithDebugStderr redirects CLI debug output to os.Stderr.
// This is useful for seeing debug output in real-time during development.
func WithDebugStderr() Option {
	return WithDebugWriter(os.Stderr)
}

// WithDebugDisabled discards all CLI debug output.
// This is more explicit than the default nil behavior but has the same effect.
func WithDebugDisabled() Option {
	return WithDebugWriter(io.Discard)
}

// OutputFormatJSONSchema creates an OutputFormat for JSON schema constraints.
func OutputFormatJSONSchema(schema map[string]any) *OutputFormat {
	return &OutputFormat{
		Type:   "json_schema",
		Schema: schema,
	}
}

// WithOutputFormat sets the output format for structured responses.
func WithOutputFormat(format *OutputFormat) Option {
	return func(o *Options) {
		o.OutputFormat = format
	}
}

// WithJSONSchema is a convenience function that sets a JSON schema output format.
// This is equivalent to WithOutputFormat(OutputFormatJSONSchema(schema)).
func WithJSONSchema(schema map[string]any) Option {
	return func(o *Options) {
		if schema == nil {
			o.OutputFormat = nil
			return
		}
		o.OutputFormat = OutputFormatJSONSchema(schema)
	}
}
