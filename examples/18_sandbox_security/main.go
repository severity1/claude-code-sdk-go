// Package main demonstrates Sandbox Security Configuration.
//
// This example shows how to configure sandboxed bash execution for
// enhanced security when running untrusted or sensitive commands.
// Sandbox security enables:
// - Isolated command execution in a secure environment
// - Auto-approval of sandboxed bash commands
// - Network access control within the sandbox
// - Exclusion of specific commands from sandboxing
//
// Key components:
// - WithSandboxEnabled: Enable/disable sandbox mode
// - WithAutoAllowBashIfSandboxed: Auto-approve bash when sandboxed
// - WithSandboxExcludedCommands: Bypass sandbox for specific commands
// - WithSandboxNetwork: Configure network access in sandbox
// - WithSandbox: Full sandbox settings configuration
// - SandboxSettings: Complete sandbox configuration struct
// - SandboxNetworkConfig: Network-specific settings
//
// NOTE: Sandbox functionality is only available on Linux and macOS.
// Windows users will see sandbox settings but execution behavior may differ.
//
// Run: go run main.go
package main

import (
	"fmt"
	"runtime"

	claudecode "github.com/severity1/claude-agent-sdk-go"
)

func main() {
	fmt.Println("Claude Agent SDK - Sandbox Security Example")
	fmt.Println("============================================")
	fmt.Println()

	// Show platform information
	fmt.Printf("Platform: %s/%s\n", runtime.GOOS, runtime.GOARCH)
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		fmt.Println("Status: Sandbox is supported on this platform")
	} else {
		fmt.Println("Status: Sandbox may have limited support on this platform")
	}
	fmt.Println()

	// Example 1: Basic Sandbox Configuration
	fmt.Println("--- Example 1: Basic Sandbox Configuration ---")
	fmt.Println("Enabling sandbox with auto-allow for bash commands...")
	demonstrateBasicSandbox()

	// Example 2: Command Exclusions
	fmt.Println()
	fmt.Println("--- Example 2: Command Exclusions ---")
	fmt.Println("Excluding specific commands from sandboxing...")
	demonstrateExclusions()

	// Example 3: Network Configuration
	fmt.Println()
	fmt.Println("--- Example 3: Network Configuration ---")
	fmt.Println("Configuring network access in sandbox...")
	demonstrateNetworkConfig()

	// Example 4: Full Sandbox Settings
	fmt.Println()
	fmt.Println("--- Example 4: Full Sandbox Settings ---")
	fmt.Println("Using complete SandboxSettings configuration...")
	demonstrateFullSettings()

	fmt.Println()
	fmt.Println("Sandbox security example completed!")
}

// demonstrateBasicSandbox shows basic sandbox enablement
func demonstrateBasicSandbox() {
	// Create client with basic sandbox configuration
	client := claudecode.NewClient(
		claudecode.WithSandboxEnabled(true),
		claudecode.WithAutoAllowBashIfSandboxed(true),
	)

	fmt.Println("Configuration:")
	fmt.Println("  - Sandbox: enabled")
	fmt.Println("  - Auto-allow bash if sandboxed: true")
	fmt.Println()
	fmt.Println("Effect: Bash commands will run in a sandboxed environment")
	fmt.Println("        and will be auto-approved without user confirmation.")

	_ = client
}

// demonstrateExclusions shows excluding commands from sandbox
func demonstrateExclusions() {
	// Some commands may need to bypass the sandbox for functionality
	excludedCommands := []string{"git", "docker", "npm"}

	client := claudecode.NewClient(
		claudecode.WithSandboxEnabled(true),
		claudecode.WithSandboxExcludedCommands(excludedCommands...),
	)

	fmt.Println("Configuration:")
	fmt.Println("  - Sandbox: enabled")
	fmt.Printf("  - Excluded commands: %v\n", excludedCommands)
	fmt.Println()
	fmt.Println("Effect: Most commands run in sandbox, but git/docker/npm")
	fmt.Println("        run outside for full functionality.")

	_ = client
}

// demonstrateNetworkConfig shows sandbox network configuration
func demonstrateNetworkConfig() {
	// Configure network access for sandbox
	networkConfig := &claudecode.SandboxNetworkConfig{
		AllowUnixSockets:    []string{"/var/run/docker.sock"},
		AllowAllUnixSockets: false,
		AllowLocalBinding:   true,
	}

	client := claudecode.NewClient(
		claudecode.WithSandboxEnabled(true),
		claudecode.WithSandboxNetwork(networkConfig),
	)

	fmt.Println("Network Configuration:")
	fmt.Printf("  - Allowed Unix sockets: %v\n", networkConfig.AllowUnixSockets)
	fmt.Printf("  - Allow all Unix sockets: %v\n", networkConfig.AllowAllUnixSockets)
	fmt.Printf("  - Allow local binding: %v\n", networkConfig.AllowLocalBinding)
	fmt.Println()
	fmt.Println("Effect: Sandbox can access Docker socket and bind to localhost,")
	fmt.Println("        but other Unix socket access is restricted.")

	_ = client
}

// demonstrateFullSettings shows complete SandboxSettings configuration
func demonstrateFullSettings() {
	// HTTP proxy port for demonstration
	httpProxyPort := 8080

	// Create comprehensive sandbox settings
	sandboxSettings := &claudecode.SandboxSettings{
		Enabled:                  true,
		AutoAllowBashIfSandboxed: true,
		ExcludedCommands:         []string{"git", "ssh"},
		AllowUnsandboxedCommands: false,
		Network: &claudecode.SandboxNetworkConfig{
			AllowUnixSockets:    []string{"/var/run/docker.sock"},
			AllowAllUnixSockets: false,
			AllowLocalBinding:   true,
			HTTPProxyPort:       &httpProxyPort,
		},
		IgnoreViolations: &claudecode.SandboxIgnoreViolations{
			File:    []string{"/tmp/*"},
			Network: []string{"localhost:*"},
		},
		EnableWeakerNestedSandbox: false,
	}

	client := claudecode.NewClient(
		claudecode.WithSandbox(sandboxSettings),
	)

	fmt.Println("Full Sandbox Settings:")
	fmt.Printf("  Enabled: %v\n", sandboxSettings.Enabled)
	fmt.Printf("  Auto-allow bash: %v\n", sandboxSettings.AutoAllowBashIfSandboxed)
	fmt.Printf("  Excluded commands: %v\n", sandboxSettings.ExcludedCommands)
	fmt.Printf("  Allow unsandboxed: %v\n", sandboxSettings.AllowUnsandboxedCommands)
	fmt.Printf("  Weaker nested sandbox: %v\n", sandboxSettings.EnableWeakerNestedSandbox)
	fmt.Println()
	fmt.Println("Network Settings:")
	fmt.Printf("  Unix sockets: %v\n", sandboxSettings.Network.AllowUnixSockets)
	fmt.Printf("  Local binding: %v\n", sandboxSettings.Network.AllowLocalBinding)
	fmt.Printf("  HTTP proxy port: %d\n", *sandboxSettings.Network.HTTPProxyPort)
	fmt.Println()
	fmt.Println("Ignored Violations:")
	fmt.Printf("  File patterns: %v\n", sandboxSettings.IgnoreViolations.File)
	fmt.Printf("  Network patterns: %v\n", sandboxSettings.IgnoreViolations.Network)

	_ = client
}
