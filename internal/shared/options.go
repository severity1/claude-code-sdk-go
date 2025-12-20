package shared

import (
	"fmt"
	"io"
)

const (
	// DefaultMaxThinkingTokens is the default maximum number of thinking tokens.
	DefaultMaxThinkingTokens = 8000
)

// PermissionMode represents the different permission handling modes.
type PermissionMode string

const (
	// PermissionModeDefault is the standard permission handling mode.
	PermissionModeDefault PermissionMode = "default"
	// PermissionModeAcceptEdits automatically accepts all edit permissions.
	PermissionModeAcceptEdits PermissionMode = "acceptEdits"
	// PermissionModePlan enables plan mode for task execution.
	PermissionModePlan PermissionMode = "plan"
	// PermissionModeBypassPermissions bypasses all permission checks.
	PermissionModeBypassPermissions PermissionMode = "bypassPermissions"
)

// SdkBeta represents a beta feature identifier.
// See https://docs.anthropic.com/en/api/beta-headers
type SdkBeta string

const (
	// SdkBetaContext1M enables the 1M context window beta feature.
	SdkBetaContext1M SdkBeta = "context-1m-2025-08-07"
)

// ToolsPreset represents a preset tools configuration.
type ToolsPreset struct {
	Type   string `json:"type"`   // Always "preset"
	Preset string `json:"preset"` // e.g., "claude_code"
}

// SettingSource represents a settings source location.
type SettingSource string

const (
	// SettingSourceUser loads user-level settings.
	SettingSourceUser SettingSource = "user"
	// SettingSourceProject loads project-level settings.
	SettingSourceProject SettingSource = "project"
	// SettingSourceLocal loads local/workspace-level settings.
	SettingSourceLocal SettingSource = "local"
)

// Options configures the Claude Code SDK behavior.
type Options struct {
	// Tool Control
	AllowedTools    []string `json:"allowed_tools,omitempty"`
	DisallowedTools []string `json:"disallowed_tools,omitempty"`

	// Tools configures available tools.
	// Can be []string (list of tool names) or ToolsPreset (preset configuration).
	Tools any `json:"tools,omitempty"`

	// Beta Features
	Betas []SdkBeta `json:"betas,omitempty"`

	// System Prompts & Model
	SystemPrompt       *string `json:"system_prompt,omitempty"`
	AppendSystemPrompt *string `json:"append_system_prompt,omitempty"`
	Model              *string `json:"model,omitempty"`
	FallbackModel      *string `json:"fallback_model,omitempty"`
	MaxThinkingTokens  int     `json:"max_thinking_tokens,omitempty"`

	// Budget & Billing
	MaxBudgetUSD *float64 `json:"max_budget_usd,omitempty"`
	User         *string  `json:"user,omitempty"`

	// Buffer Configuration (internal)
	MaxBufferSize *int `json:"max_buffer_size,omitempty"`

	// Permission & Safety System
	PermissionMode           *PermissionMode `json:"permission_mode,omitempty"`
	PermissionPromptToolName *string         `json:"permission_prompt_tool_name,omitempty"`

	// Session & State Management
	ContinueConversation bool            `json:"continue_conversation,omitempty"`
	Resume               *string         `json:"resume,omitempty"`
	MaxTurns             int             `json:"max_turns,omitempty"`
	Settings             *string         `json:"settings,omitempty"`
	ForkSession          bool            `json:"fork_session,omitempty"`
	SettingSources       []SettingSource `json:"setting_sources,omitempty"`

	// File System & Context
	Cwd     *string  `json:"cwd,omitempty"`
	AddDirs []string `json:"add_dirs,omitempty"`

	// MCP Integration
	McpServers map[string]McpServerConfig `json:"mcp_servers,omitempty"`

	// Extensibility
	ExtraArgs map[string]*string `json:"extra_args,omitempty"`

	// ExtraEnv specifies additional environment variables for the subprocess.
	// These are merged with the system environment variables.
	ExtraEnv map[string]string `json:"extra_env,omitempty"`

	// CLI Path (for testing and custom installations)
	CLIPath *string `json:"cli_path,omitempty"`

	// DebugWriter specifies where to write debug output from the CLI subprocess.
	// If nil (default), stderr is isolated to a temporary file to prevent deadlocks.
	// Common values: os.Stderr, io.Discard, or a custom io.Writer.
	DebugWriter io.Writer `json:"-"` // Not serialized
}

// McpServerType represents the type of MCP server.
type McpServerType string

const (
	// McpServerTypeStdio represents a stdio-based MCP server.
	McpServerTypeStdio McpServerType = "stdio"
	// McpServerTypeSSE represents a Server-Sent Events MCP server.
	McpServerTypeSSE McpServerType = "sse"
	// McpServerTypeHTTP represents an HTTP-based MCP server.
	McpServerTypeHTTP McpServerType = "http"
)

// McpServerConfig represents MCP server configuration.
type McpServerConfig interface {
	GetType() McpServerType
}

// McpStdioServerConfig configures an MCP stdio server.
type McpStdioServerConfig struct {
	Type    McpServerType     `json:"type"`
	Command string            `json:"command"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
}

// GetType returns the server type for McpStdioServerConfig.
func (c *McpStdioServerConfig) GetType() McpServerType {
	return McpServerTypeStdio
}

// McpSSEServerConfig configures an MCP Server-Sent Events server.
type McpSSEServerConfig struct {
	Type    McpServerType     `json:"type"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers,omitempty"`
}

// GetType returns the server type for McpSSEServerConfig.
func (c *McpSSEServerConfig) GetType() McpServerType {
	return McpServerTypeSSE
}

// McpHTTPServerConfig configures an MCP HTTP server.
type McpHTTPServerConfig struct {
	Type    McpServerType     `json:"type"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers,omitempty"`
}

// GetType returns the server type for McpHTTPServerConfig.
func (c *McpHTTPServerConfig) GetType() McpServerType {
	return McpServerTypeHTTP
}

// Validate checks the options for valid values and constraints.
func (o *Options) Validate() error {
	// Validate MaxThinkingTokens
	if o.MaxThinkingTokens < 0 {
		return fmt.Errorf("MaxThinkingTokens must be non-negative, got %d", o.MaxThinkingTokens)
	}

	// Validate MaxTurns
	if o.MaxTurns < 0 {
		return fmt.Errorf("MaxTurns must be non-negative, got %d", o.MaxTurns)
	}

	// Validate tool conflicts (same tool in both allowed and disallowed)
	allowedSet := make(map[string]bool)
	for _, tool := range o.AllowedTools {
		allowedSet[tool] = true
	}

	for _, tool := range o.DisallowedTools {
		if allowedSet[tool] {
			return fmt.Errorf("tool '%s' cannot be in both AllowedTools and DisallowedTools", tool)
		}
	}

	return nil
}

// NewOptions creates Options with default values.
func NewOptions() *Options {
	return &Options{
		AllowedTools:      []string{},
		DisallowedTools:   []string{},
		Betas:             []SdkBeta{},
		MaxThinkingTokens: DefaultMaxThinkingTokens,
		AddDirs:           []string{},
		McpServers:        make(map[string]McpServerConfig),
		ExtraArgs:         make(map[string]*string),
		ExtraEnv:          make(map[string]string),
		SettingSources:    []SettingSource{},
	}
}
