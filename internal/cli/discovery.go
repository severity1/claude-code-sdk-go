// Package cli provides CLI discovery and command building functionality.
package cli

import (
	"os"
	"os/exec"
	"path/filepath"
)

// DiscoveryPaths defines the standard search paths for Claude CLI.
var DiscoveryPaths = []string{
	// Will be populated with dynamic paths in FindCLI()
}

// FindCLI searches for the Claude CLI binary in standard locations.
func FindCLI() (string, error) {
	// 1. Check PATH
	if path, err := exec.LookPath("claude"); err == nil {
		return path, nil
	}

	// 2. Check standard locations
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	locations := []string{
		filepath.Join(home, ".npm-global/bin/claude"),
		"/usr/local/bin/claude",
		filepath.Join(home, ".local/bin/claude"),
		filepath.Join(home, "node_modules/.bin/claude"),
		filepath.Join(home, ".yarn/bin/claude"),
	}

	for _, path := range locations {
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			return path, nil
		}
	}

	// Check Node.js dependency
	if _, err := exec.LookPath("node"); err != nil {
		return "", err // TODO: Return CLINotFoundError with Node.js guidance
	}

	return "", nil // TODO: Return CLINotFoundError with installation guidance
}

// BuildCommand constructs the CLI command with all necessary flags.
func BuildCommand(cliPath string, options interface{}, closeStdin bool) []string {
	cmd := []string{cliPath}

	// Base arguments
	cmd = append(cmd, "--output-format", "stream-json", "--verbose")

	if closeStdin {
		// One-shot mode
		cmd = append(cmd, "--print", "placeholder-prompt")
	} else {
		// Streaming mode
		cmd = append(cmd, "--input-format", "stream-json")
	}

	// TODO: Add all configuration options as CLI flags
	// TODO: Handle ExtraArgs for arbitrary flags

	return cmd
}

// ValidateNodeJS checks if Node.js is available.
func ValidateNodeJS() error {
	if _, err := exec.LookPath("node"); err != nil {
		return err // TODO: Return helpful error with installation instructions
	}
	return nil
}
