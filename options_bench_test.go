package claudecode

import (
	"testing"
)

// sink prevents dead code elimination by the compiler.
var sink any

// BenchmarkNewOptions measures functional options pattern performance.
func BenchmarkNewOptions(b *testing.B) {
	model := "claude-sonnet-4-5"
	prompt := "You are a helpful assistant"

	tests := []struct {
		name string
		opts []Option
	}{
		{
			name: "minimal",
			opts: nil,
		},
		{
			name: "with_model",
			opts: []Option{
				WithModel(model),
			},
		},
		{
			name: "common",
			opts: []Option{
				WithModel(model),
				WithSystemPrompt(prompt),
				WithMaxThinkingTokens(16000),
			},
		},
		{
			name: "with_tools",
			opts: []Option{
				WithModel(model),
				WithAllowedTools("Read", "Write", "Bash", "Glob", "Grep"),
				WithDisallowedTools("Edit"),
			},
		},
		{
			name: "with_permission",
			opts: []Option{
				WithModel(model),
				WithPermissionMode(PermissionModeAcceptEdits),
			},
		},
		{
			name: "full",
			opts: []Option{
				WithModel(model),
				WithSystemPrompt(prompt),
				WithMaxThinkingTokens(16000),
				WithAllowedTools("Read", "Write", "Bash"),
				WithPermissionMode(PermissionModeAcceptEdits),
				WithMaxBudgetUSD(1.0),
				WithMaxTurns(10),
				WithCwd("/home/user/project"),
			},
		},
	}

	for _, tc := range tests {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				result := NewOptions(tc.opts...)
				sink = result
			}
		})
	}
}

// BenchmarkNewOptions_MCPServers measures MCP server configuration performance.
func BenchmarkNewOptions_MCPServers(b *testing.B) {
	servers := map[string]McpServerConfig{
		"aws": &McpStdioServerConfig{
			Type:    McpServerTypeStdio,
			Command: "aws-mcp",
			Args:    []string{"--profile", "default"},
		},
		"db": &McpStdioServerConfig{
			Type:    McpServerTypeStdio,
			Command: "db-mcp",
			Args:    []string{"--host", "localhost"},
		},
	}

	opts := []Option{
		WithMcpServers(servers),
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		result := NewOptions(opts...)
		sink = result
	}
}

// BenchmarkNewOptions_Agents measures agent configuration performance.
func BenchmarkNewOptions_Agents(b *testing.B) {
	agents := map[string]AgentDefinition{
		"code-reviewer": {
			Description: "Reviews code for quality",
			Prompt:      "You are a code reviewer...",
			Tools:       []string{"Read", "Glob", "Grep"},
			Model:       AgentModelSonnet,
		},
		"test-writer": {
			Description: "Writes tests for code",
			Prompt:      "You are a test writer...",
			Tools:       []string{"Read", "Write"},
			Model:       AgentModelHaiku,
		},
	}

	opts := []Option{
		WithAgents(agents),
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		result := NewOptions(opts...)
		sink = result
	}
}

// BenchmarkNewOptions_Sandbox measures sandbox configuration performance.
func BenchmarkNewOptions_Sandbox(b *testing.B) {
	sandbox := SandboxSettings{
		Enabled:                  true,
		AutoAllowBashIfSandboxed: true,
		ExcludedCommands:         []string{"git", "npm"},
		Network: &SandboxNetworkConfig{
			AllowUnixSockets:  []string{"/tmp/socket"},
			AllowLocalBinding: true,
		},
	}

	opts := []Option{
		WithSandbox(&sandbox),
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		result := NewOptions(opts...)
		sink = result
	}
}

// BenchmarkWithOption_Individual measures individual option application.
func BenchmarkWithOption_Individual(b *testing.B) {
	tests := []struct {
		name string
		opt  Option
	}{
		{"model", WithModel("claude-sonnet-4-5")},
		{"system_prompt", WithSystemPrompt("You are helpful")},
		{"max_thinking", WithMaxThinkingTokens(16000)},
		{"permission_mode", WithPermissionMode(PermissionModeAcceptEdits)},
		{"allowed_tools", WithAllowedTools("Read", "Write", "Bash")},
	}

	for _, tc := range tests {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				result := NewOptions(tc.opt)
				sink = result
			}
		})
	}
}
