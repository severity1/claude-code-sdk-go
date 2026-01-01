// Package main demonstrates Plugin Configuration.
//
// This example shows how to configure local plugins for the Claude Code
// CLI using the SDK. Plugins extend Claude's capabilities with custom
// commands, tools, and integrations. Plugin configuration enables:
// - Loading custom plugins from local paths
// - Configuring multiple plugins simultaneously
// - Extending Claude Code with domain-specific functionality
// - Integration with existing plugin ecosystems
//
// Key components:
// - WithLocalPlugin: Convenience function for local plugin paths
// - WithPlugin: Add a single plugin with explicit configuration
// - WithPlugins: Add multiple plugins at once
// - SdkPluginConfig: Plugin configuration structure
// - SdkPluginTypeLocal: Plugin type constant for local plugins
//
// NOTE: This example demonstrates the API configuration without requiring
// actual plugins to exist. In production, plugins must be valid Claude Code
// plugin directories.
//
// Run: go run main.go
package main

import (
	"fmt"

	claudecode "github.com/severity1/claude-agent-sdk-go"
)

func main() {
	fmt.Println("Claude Agent SDK - Plugins Configuration Example")
	fmt.Println("=================================================")
	fmt.Println()

	// Example 1: Single Plugin with WithLocalPlugin
	fmt.Println("--- Example 1: Single Plugin Configuration ---")
	fmt.Println("Using WithLocalPlugin convenience function...")
	demonstrateSinglePlugin()

	// Example 2: Single Plugin with Explicit Config
	fmt.Println()
	fmt.Println("--- Example 2: Explicit Plugin Configuration ---")
	fmt.Println("Using WithPlugin with SdkPluginConfig...")
	demonstrateExplicitConfig()

	// Example 3: Multiple Plugins
	fmt.Println()
	fmt.Println("--- Example 3: Multiple Plugins ---")
	fmt.Println("Configuring multiple plugins with WithPlugins...")
	demonstrateMultiplePlugins()

	// Example 4: Plugin Configuration Patterns
	fmt.Println()
	fmt.Println("--- Example 4: Configuration Patterns ---")
	fmt.Println("Common plugin configuration patterns...")
	showConfigurationPatterns()

	fmt.Println()
	fmt.Println("Plugins configuration example completed!")
}

// demonstrateSinglePlugin shows the WithLocalPlugin convenience function
func demonstrateSinglePlugin() {
	// WithLocalPlugin is the simplest way to add a local plugin
	pluginPath := "/path/to/my-plugin"

	// Create a client with the plugin (demonstration only)
	client := claudecode.NewClient(
		claudecode.WithLocalPlugin(pluginPath),
	)

	fmt.Printf("Configured plugin path: %s\n", pluginPath)
	fmt.Printf("Client created with plugin configuration\n")

	// Note: We don't connect since this is just a configuration demo
	_ = client
}

// demonstrateExplicitConfig shows using SdkPluginConfig directly
func demonstrateExplicitConfig() {
	// Create explicit plugin configuration
	pluginConfig := claudecode.SdkPluginConfig{
		Type: claudecode.SdkPluginTypeLocal,
		Path: "/path/to/custom-plugin",
	}

	fmt.Printf("Plugin Type: %s\n", pluginConfig.Type)
	fmt.Printf("Plugin Path: %s\n", pluginConfig.Path)
	fmt.Println()

	// Create a client with explicit config
	client := claudecode.NewClient(
		claudecode.WithPlugin(pluginConfig),
	)

	fmt.Println("Client created with explicit plugin configuration")
	_ = client
}

// demonstrateMultiplePlugins shows configuring multiple plugins
func demonstrateMultiplePlugins() {
	// Define multiple plugins
	plugins := []claudecode.SdkPluginConfig{
		{
			Type: claudecode.SdkPluginTypeLocal,
			Path: "/plugins/code-formatter",
		},
		{
			Type: claudecode.SdkPluginTypeLocal,
			Path: "/plugins/test-runner",
		},
		{
			Type: claudecode.SdkPluginTypeLocal,
			Path: "/plugins/deployment-helper",
		},
	}

	fmt.Printf("Configuring %d plugins:\n", len(plugins))
	for i, p := range plugins {
		fmt.Printf("  %d. [%s] %s\n", i+1, p.Type, p.Path)
	}
	fmt.Println()

	// Create client with all plugins
	client := claudecode.NewClient(
		claudecode.WithPlugins(plugins),
	)

	fmt.Println("Client created with multiple plugins")
	_ = client
}

// showConfigurationPatterns demonstrates common plugin patterns
func showConfigurationPatterns() {
	fmt.Println("Pattern 1: Chaining WithLocalPlugin calls")
	fmt.Println("  claudecode.NewClient(")
	fmt.Println("      claudecode.WithLocalPlugin(\"/plugins/formatter\"),")
	fmt.Println("      claudecode.WithLocalPlugin(\"/plugins/linter\"),")
	fmt.Println("  )")
	fmt.Println()

	fmt.Println("Pattern 2: Using WithPlugins for bulk configuration")
	fmt.Println("  plugins := []claudecode.SdkPluginConfig{...}")
	fmt.Println("  claudecode.NewClient(claudecode.WithPlugins(plugins))")
	fmt.Println()

	fmt.Println("Pattern 3: Combining with other options")
	fmt.Println("  claudecode.WithClient(ctx, handler,")
	fmt.Println("      claudecode.WithLocalPlugin(\"/path/to/plugin\"),")
	fmt.Println("      claudecode.WithAllowedTools(\"Read\", \"Write\"),")
	fmt.Println("      claudecode.WithMaxTurns(5),")
	fmt.Println("  )")
	fmt.Println()

	fmt.Println("Available plugin types:")
	fmt.Printf("  - SdkPluginTypeLocal: %q\n", claudecode.SdkPluginTypeLocal)
}
