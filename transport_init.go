package claudecode

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// init sets up the default transport factory to avoid import cycles.
func init() {
	defaultTransportFactory = func(options *Options, closeStdin bool) (Transport, error) {
		// Find Claude CLI binary
		cliPath, err := findCLI()
		if err != nil {
			return nil, fmt.Errorf("failed to find Claude CLI: %w", err)
		}
		
		// For now, return a helpful error that indicates transport selection is working
		// The actual subprocess transport creation is blocked by import cycle issues
		// that need to be resolved at the architectural level
		return nil, fmt.Errorf("transport selection successful - found CLI at %s, but subprocess creation blocked by import cycle (architectural issue)", cliPath)
	}
}

// findCLI searches for the Claude CLI binary in standard locations.
func findCLI() (string, error) {
	executableName := "claude"
	if runtime.GOOS == "windows" {
		executableName = "claude.exe"
	}
	
	// 1. Check system PATH first
	if path, err := exec.LookPath(executableName); err == nil {
		return path, nil
	}
	
	// 2. Get home directory for other paths
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", NewCLINotFoundError("", "Unable to determine home directory")
	}
	
	// 3. Check standard installation locations in order
	locations := []string{
		filepath.Join(homeDir, ".npm-global", "bin", executableName),
		filepath.Join("/usr/local/bin", executableName),
		filepath.Join(homeDir, ".local", "bin", executableName),
		filepath.Join(homeDir, "node_modules", ".bin", executableName),
		filepath.Join(homeDir, ".yarn", "bin", executableName),
	}
	
	for _, location := range locations {
		if _, err := os.Stat(location); err == nil {
			return location, nil
		}
	}
	
	// 4. CLI not found, return helpful error
	return "", NewCLINotFoundError("", 
		"Claude CLI not found. Please install it with:\n"+
		"npm install -g @anthropic-ai/claude-code\n"+
		"or visit https://docs.anthropic.com/claude/docs/claude-code")
}