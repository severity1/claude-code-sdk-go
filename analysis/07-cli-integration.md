# CLI Integration Analysis

Analysis of Python SDK's Claude Code CLI discovery, command construction, and process management patterns.

## CLI Discovery Strategy (_internal/transport/subprocess.py)

**Python CLI Finding Logic**:
```python
def _find_cli(self) -> str:
    """Find Claude Code CLI executable with comprehensive search strategy."""
    
    # 1. Check PATH first
    cli_path = shutil.which("claude")
    if cli_path:
        return cli_path
    
    # 2. Check common installation locations
    possible_locations = [
        Path.home() / ".npm-global" / "bin" / "claude",
        Path("/usr/local/bin/claude"),
        Path.home() / ".local" / "bin" / "claude", 
        Path.home() / "node_modules" / ".bin" / "claude",
        Path.home() / ".yarn" / "bin" / "claude",
        # Windows locations
        Path.home() / "AppData" / "Roaming" / "npm" / "claude.cmd",
        Path("C:/Program Files/nodejs/claude.cmd"),
    ]
    
    for location in possible_locations:
        if location.exists() and not location.is_dir():
            return str(location)
    
    # 3. Check Node.js dependency
    node_installed = shutil.which("node") is not None
    
    if not node_installed:
        raise CLINotFoundError(
            "Claude Code requires Node.js, which is not installed.\n\n"
            "Install Node.js from: https://nodejs.org/\n"
            "\nAfter installing Node.js, install Claude Code:\n"
            "  npm install -g @anthropic-ai/claude-code"
        )
    
    # 4. Provide helpful installation guidance
    raise CLINotFoundError(
        "Claude Code not found. Install with:\n"
        "  npm install -g @anthropic-ai/claude-code\n"
        "\nIf already installed locally, try:\n"
        '  export PATH="$HOME/node_modules/.bin:$PATH"\n'
        "\nOr specify the path when creating transport:\n"
        "  SubprocessCLITransport(..., cli_path='/path/to/claude')"
    )
```

**Go CLI Discovery Implementation**:
```go
package claude

import (
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "runtime"
)

func FindCLI() (string, error) {
    // 1. Check PATH first - most common case
    if path, err := exec.LookPath("claude"); err == nil {
        return path, nil
    }
    
    // 2. Check platform-specific common locations
    locations := getCommonCLILocations()
    
    for _, location := range locations {
        if info, err := os.Stat(location); err == nil && !info.IsDir() {
            // Verify it's executable (Unix-like systems)
            if runtime.GOOS != "windows" {
                if info.Mode()&0111 == 0 {
                    continue // Not executable
                }
            }
            return location, nil
        }
    }
    
    // 3. Check Node.js dependency
    if _, err := exec.LookPath("node"); err != nil {
        return "", NewCLINotFoundError(
            "Claude Code requires Node.js, which is not installed.\n\n"+
                "Install Node.js from: https://nodejs.org/\n\n"+
                "After installing Node.js, install Claude Code:\n"+
                "  npm install -g @anthropic-ai/claude-code",
            "",
        )
    }
    
    // 4. Provide installation guidance
    return "", NewCLINotFoundError(
        "Claude Code not found. Install with:\n"+
            "  npm install -g @anthropic-ai/claude-code\n\n"+
            "If already installed locally, try:\n"+
            `  export PATH="$HOME/node_modules/.bin:$PATH"`+"\n\n"+
            "Or specify the path when creating client",
        "",
    )
}

func getCommonCLILocations() []string {
    homeDir, _ := os.UserHomeDir()
    
    var locations []string
    
    switch runtime.GOOS {
    case "windows":
        locations = []string{
            filepath.Join(homeDir, "AppData", "Roaming", "npm", "claude.cmd"),
            filepath.Join("C:", "Program Files", "nodejs", "claude.cmd"),
            filepath.Join(homeDir, ".npm-global", "claude.cmd"),
            filepath.Join(homeDir, "node_modules", ".bin", "claude.cmd"),
        }
    default: // Unix-like systems
        locations = []string{
            filepath.Join(homeDir, ".npm-global", "bin", "claude"),
            "/usr/local/bin/claude",
            filepath.Join(homeDir, ".local", "bin", "claude"),
            filepath.Join(homeDir, "node_modules", ".bin", "claude"),
            filepath.Join(homeDir, ".yarn", "bin", "claude"),
            "/opt/homebrew/bin/claude", // macOS Homebrew ARM
            "/usr/local/homebrew/bin/claude", // macOS Homebrew Intel
        }
    }
    
    return locations
}

// Optional: Version validation
func ValidateCLIVersion(cliPath string) (string, error) {
    cmd := exec.Command(cliPath, "--version")
    output, err := cmd.Output()
    if err != nil {
        return "", fmt.Errorf("failed to get CLI version: %w", err)
    }
    
    version := strings.TrimSpace(string(output))
    
    // Basic version format validation
    if !strings.Contains(version, ".") {
        return "", fmt.Errorf("invalid version format: %s", version)
    }
    
    return version, nil
}
```

## Command Construction Architecture

**Python Command Building**:
```python
def _build_command(self) -> list[str]:
    """Build complete CLI command with all options."""
    cmd = [self._cli_path, "--output-format", "stream-json", "--verbose"]
    
    # Core configuration options
    if self._options.system_prompt:
        cmd.extend(["--system-prompt", self._options.system_prompt])
    
    if self._options.append_system_prompt:
        cmd.extend(["--append-system-prompt", self._options.append_system_prompt])
    
    if self._options.allowed_tools:
        cmd.extend(["--allowedTools", ",".join(self._options.allowed_tools)])
    
    if self._options.disallowed_tools:
        cmd.extend(["--disallowedTools", ",".join(self._options.disallowed_tools)])
    
    if self._options.max_turns is not None:
        cmd.extend(["--max-turns", str(self._options.max_turns)])
    
    if self._options.model:
        cmd.extend(["--model", self._options.model])
    
    if self._options.permission_prompt_tool_name:
        cmd.extend(["--permission-prompt-tool", self._options.permission_prompt_tool_name])
    
    if self._options.permission_mode:
        cmd.extend(["--permission-mode", self._options.permission_mode])
    
    if self._options.continue_conversation:
        cmd.append("--continue")
    
    if self._options.resume:
        cmd.extend(["--resume", self._options.resume])
    
    # MCP configuration
    if self._options.mcp_servers:
        if isinstance(self._options.mcp_servers, (str, Path)):
            cmd.extend(["--mcp-config", str(self._options.mcp_servers)])
        else:
            # Write config to temp file
            mcp_config_file = self._write_mcp_config(self._options.mcp_servers)
            cmd.extend(["--mcp-config", mcp_config_file])
    
    # Directory additions (v0.0.19+)
    if self._options.add_dirs:
        for directory in self._options.add_dirs:
            cmd.extend(["--add-dir", str(directory)])
    
    # Settings override (v0.0.18+)
    if self._options.settings:
        cmd.extend(["--settings", self._options.settings])
    
    # Extensibility: Extra arguments for future CLI flags
    for flag, value in self._options.extra_args.items():
        if value is None:
            cmd.append(f"--{flag}")  # Boolean flag
        else:
            cmd.extend([f"--{flag}", str(value)])  # Flag with value
    
    # Input mode configuration
    if self._is_streaming:
        cmd.extend(["--input-format", "stream-json"])
    else:
        # String prompt mode
        cmd.extend(["--print", str(self._prompt)])
    
    return cmd
```

**Go Command Builder**:
```go
type CommandBuilder struct {
    cliPath   string
    options   *Options
    streaming bool
    prompt    interface{}
}

func NewCommandBuilder(cliPath string, options *Options) *CommandBuilder {
    return &CommandBuilder{
        cliPath: cliPath,
        options: options,
    }
}

func (cb *CommandBuilder) SetPrompt(prompt interface{}) *CommandBuilder {
    cb.prompt = prompt
    _, cb.streaming = prompt.([]Message) // Check if it's a message stream
    return cb
}

func (cb *CommandBuilder) Build() []string {
    args := []string{
        "--output-format", "stream-json",
        "--verbose",
    }
    
    // Core configuration
    if cb.options.SystemPrompt != "" {
        args = append(args, "--system-prompt", cb.options.SystemPrompt)
    }
    
    if cb.options.AppendSystemPrompt != "" {
        args = append(args, "--append-system-prompt", cb.options.AppendSystemPrompt)
    }
    
    if len(cb.options.AllowedTools) > 0 {
        args = append(args, "--allowedTools", strings.Join(cb.options.AllowedTools, ","))
    }
    
    if len(cb.options.DisallowedTools) > 0 {
        args = append(args, "--disallowedTools", strings.Join(cb.options.DisallowedTools, ","))
    }
    
    if cb.options.MaxTurns > 0 {
        args = append(args, "--max-turns", fmt.Sprintf("%d", cb.options.MaxTurns))
    }
    
    if cb.options.Model != "" {
        args = append(args, "--model", cb.options.Model)
    }
    
    if cb.options.PermissionPromptToolName != "" {
        args = append(args, "--permission-prompt-tool", cb.options.PermissionPromptToolName)
    }
    
    if cb.options.PermissionMode != "" {
        args = append(args, "--permission-mode", string(cb.options.PermissionMode))
    }
    
    if cb.options.ContinueConversation {
        args = append(args, "--continue")
    }
    
    if cb.options.Resume != "" {
        args = append(args, "--resume", cb.options.Resume)
    }
    
    // MCP configuration
    args = cb.addMCPConfig(args)
    
    // Directory additions
    for _, dir := range cb.options.AddDirs {
        args = append(args, "--add-dir", dir)
    }
    
    // Settings override
    if cb.options.Settings != "" {
        args = append(args, "--settings", cb.options.Settings)
    }
    
    // Extra arguments for extensibility
    for flag, value := range cb.options.ExtraArgs {
        if value == nil {
            args = append(args, "--"+flag) // Boolean flag
        } else {
            args = append(args, "--"+flag, *value) // Flag with value
        }
    }
    
    // Input mode
    if cb.streaming {
        args = append(args, "--input-format", "stream-json")
    } else if promptStr, ok := cb.prompt.(string); ok {
        args = append(args, "--print", promptStr)
    }
    
    return args
}

func (cb *CommandBuilder) addMCPConfig(args []string) []string {
    if len(cb.options.McpServers) == 0 {
        return args
    }
    
    // Write MCP config to temporary file
    configFile, err := cb.writeMCPConfig(cb.options.McpServers)
    if err != nil {
        // Log error but continue - MCP is optional
        log.Printf("Failed to write MCP config: %v", err)
        return args
    }
    
    return append(args, "--mcp-config", configFile)
}
```

## MCP Configuration File Generation

**Python MCP Config Writing**:
```python
def _write_mcp_config(self, servers: dict[str, McpServerConfig]) -> str:
    """Write MCP server configuration to temporary file."""
    import json
    import tempfile
    
    with tempfile.NamedTemporaryFile(
        mode='w', 
        suffix='.json',
        prefix='claude_mcp_',
        delete=False
    ) as f:
        # Convert TypedDict to regular dict for JSON serialization
        config = {
            "mcpServers": {
                name: dict(server_config)
                for name, server_config in servers.items()
            }
        }
        json.dump(config, f, indent=2)
        return f.name
```

**Go MCP Config Writing**:
```go
func (cb *CommandBuilder) writeMCPConfig(servers map[string]McpServerConfig) (string, error) {
    // Create config structure
    config := map[string]interface{}{
        "mcpServers": make(map[string]interface{}),
    }
    
    // Convert server configs to JSON-serializable format
    mcpServers := config["mcpServers"].(map[string]interface{})
    for name, serverConfig := range servers {
        configData, err := cb.serverConfigToMap(serverConfig)
        if err != nil {
            return "", fmt.Errorf("failed to convert server config %s: %w", name, err)
        }
        mcpServers[name] = configData
    }
    
    // Create temporary file
    tmpFile, err := os.CreateTemp("", "claude_mcp_*.json")
    if err != nil {
        return "", fmt.Errorf("failed to create MCP config temp file: %w", err)
    }
    
    // Write JSON config
    encoder := json.NewEncoder(tmpFile)
    encoder.SetIndent("", "  ")
    if err := encoder.Encode(config); err != nil {
        tmpFile.Close()
        os.Remove(tmpFile.Name())
        return "", fmt.Errorf("failed to write MCP config: %w", err)
    }
    
    if err := tmpFile.Close(); err != nil {
        os.Remove(tmpFile.Name())
        return "", fmt.Errorf("failed to close MCP config file: %w", err)
    }
    
    return tmpFile.Name(), nil
}

func (cb *CommandBuilder) serverConfigToMap(config McpServerConfig) (map[string]interface{}, error) {
    // Use reflection or type switches to convert to map
    switch c := config.(type) {
    case *McpStdioServerConfig:
        result := map[string]interface{}{
            "command": c.Command,
        }
        if c.Type != "" {
            result["type"] = c.Type
        }
        if len(c.Args) > 0 {
            result["args"] = c.Args
        }
        if len(c.Env) > 0 {
            result["env"] = c.Env
        }
        return result, nil
        
    case *McpSSEServerConfig:
        result := map[string]interface{}{
            "type": "sse",
            "url":  c.URL,
        }
        if len(c.Headers) > 0 {
            result["headers"] = c.Headers
        }
        return result, nil
        
    case *McpHttpServerConfig:
        result := map[string]interface{}{
            "type": "http", 
            "url":  c.URL,
        }
        if len(c.Headers) > 0 {
            result["headers"] = c.Headers
        }
        return result, nil
        
    default:
        return nil, fmt.Errorf("unknown MCP server config type: %T", config)
    }
}
```

## Environment Variable Configuration

**Python Environment Setup**:
```python
async def connect(self) -> None:
    """Start subprocess with proper environment."""
    # ... other setup ...
    
    # Set SDK identifier for telemetry/debugging
    env = {**os.environ, "CLAUDE_CODE_ENTRYPOINT": "sdk-py"}
    
    self._process = await anyio.open_process(
        cmd,
        stdin=PIPE,
        stdout=PIPE,
        stderr=self._stderr_file,
        cwd=self._cwd,
        env=env,  # Pass modified environment
    )
```

**Go Environment Setup**:
```go
func (t *SubprocessTransport) Connect(ctx context.Context) error {
    cmd, err := t.buildCommand()
    if err != nil {
        return err
    }
    
    // Set environment variables
    cmd.Env = append(os.Environ(), "CLAUDE_CODE_ENTRYPOINT=sdk-go")
    
    // Add working directory if specified
    if t.cwd != "" {
        if _, err := os.Stat(t.cwd); os.IsNotExist(err) {
            return NewConnectionError(
                fmt.Sprintf("working directory does not exist: %s", t.cwd),
                err,
            )
        }
        cmd.Dir = t.cwd
    }
    
    // Set up pipes and start process
    // ... rest of connection logic
}
```

## Working Directory Validation

**Python Working Directory Handling**:
```python
async def connect(self) -> None:
    # Validate working directory exists before starting process
    if self._cwd and not Path(self._cwd).exists():
        raise CLIConnectionError(f"Working directory does not exist: {self._cwd}")
    
    # ... continue with process startup ...
```

**Go Working Directory Validation**:
```go
func (t *SubprocessTransport) validateWorkingDirectory() error {
    if t.cwd == "" {
        return nil // No validation needed if no cwd specified
    }
    
    info, err := os.Stat(t.cwd)
    if os.IsNotExist(err) {
        return NewConnectionError(
            fmt.Sprintf("working directory does not exist: %s", t.cwd),
            err,
        )
    }
    if err != nil {
        return fmt.Errorf("failed to check working directory: %w", err)
    }
    
    if !info.IsDir() {
        return NewConnectionError(
            fmt.Sprintf("working directory path is not a directory: %s", t.cwd),
            nil,
        )
    }
    
    return nil
}
```

## Cross-Platform Process Handling

**Python Cross-Platform Process Management**:
```python
# Python's anyio handles cross-platform differences automatically
self._process = await anyio.open_process(
    cmd,
    stdin=PIPE,
    stdout=PIPE,
    stderr=self._stderr_file,
    cwd=self._cwd,
    env=env,
)
```

**Go Cross-Platform Process Management**:
```go
func (t *SubprocessTransport) startProcess(ctx context.Context) error {
    cmd, err := t.buildCommand()
    if err != nil {
        return err
    }
    
    // Platform-specific process setup
    if err := t.setupPlatformSpecific(cmd); err != nil {
        return err
    }
    
    // Create pipes
    stdin, err := cmd.StdinPipe()
    if err != nil {
        return fmt.Errorf("failed to create stdin pipe: %w", err)
    }
    
    stdout, err := cmd.StdoutPipe()
    if err != nil {
        return fmt.Errorf("failed to create stdout pipe: %w", err)
    }
    
    // Platform-specific stderr handling
    stderrFile, err := t.setupStderr()
    if err != nil {
        return fmt.Errorf("failed to setup stderr: %w", err)
    }
    cmd.Stderr = stderrFile
    
    // Start the process
    if err := cmd.Start(); err != nil {
        return t.handleStartError(err)
    }
    
    t.cmd = cmd
    t.stdin = stdin
    t.stdout = stdout
    t.stderrFile = stderrFile
    
    return nil
}

func (t *SubprocessTransport) setupPlatformSpecific(cmd *exec.Cmd) error {
    switch runtime.GOOS {
    case "windows":
        // Windows-specific setup
        cmd.SysProcAttr = &syscall.SysProcAttr{
            CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
        }
    default:
        // Unix-like systems
        cmd.SysProcAttr = &syscall.SysProcAttr{
            Setpgid: true, // Create new process group for clean termination
        }
    }
    return nil
}

func (t *SubprocessTransport) setupStderr() (*os.File, error) {
    return os.CreateTemp("", "claude_stderr_*.log")
}
```

## CLI Version Compatibility

**Python Version Detection** (implicit in error handling):
```python
# Python SDK handles version compatibility through error messages
# and feature detection rather than explicit version checking
def _handle_unsupported_flag(self, flag: str, stderr: str) -> bool:
    """Check if error is due to unsupported CLI flag."""
    unsupported_patterns = [
        f"unknown option: {flag}",
        f"invalid argument: {flag}",
        "unrecognized arguments:",
    ]
    
    return any(pattern in stderr.lower() for pattern in unsupported_patterns)
```

**Go Version Compatibility**:
```go
type CLIVersion struct {
    Major int
    Minor int
    Patch int
    Raw   string
}

func (t *SubprocessTransport) detectCLIVersion() (*CLIVersion, error) {
    cmd := exec.Command(t.cliPath, "--version")
    output, err := cmd.Output()
    if err != nil {
        return nil, fmt.Errorf("failed to get CLI version: %w", err)
    }
    
    version, err := parseCLIVersion(strings.TrimSpace(string(output)))
    if err != nil {
        return nil, fmt.Errorf("failed to parse CLI version: %w", err)
    }
    
    return version, nil
}

func (t *SubprocessTransport) supportsFlag(flag string, version *CLIVersion) bool {
    switch flag {
    case "--add-dir":
        return version.IsAtLeast(0, 0, 19)
    case "--settings":
        return version.IsAtLeast(0, 0, 18)
    default:
        return true // Assume supported for unknown flags
    }
}

func (v *CLIVersion) IsAtLeast(major, minor, patch int) bool {
    if v.Major > major {
        return true
    }
    if v.Major == major && v.Minor > minor {
        return true
    }
    if v.Major == major && v.Minor == minor && v.Patch >= patch {
        return true
    }
    return false
}
```

This comprehensive CLI integration system provides robust executable discovery, command construction, and process management while handling cross-platform differences and maintaining compatibility with evolving CLI features.