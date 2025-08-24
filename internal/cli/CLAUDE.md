# CLI Integration Context

**Context**: Claude CLI discovery, command building, and Node.js dependency validation for Phase 3 implementation

## Component Focus
- **CLI Discovery** - Search 6 locations in specific order for Claude CLI binary
- **Command Building** - Dynamic argument construction from Options struct  
- **Node.js Validation** - Verify Node.js dependency and provide installation guidance
- **Cross-Platform** - Handle differences across Windows, macOS, Linux

## CLI Discovery Sequence (Critical Order)
Must search in exact order for Claude CLI binary:

1. **System PATH** - `exec.LookPath("claude")` first
2. **NPM Global** - `~/.npm-global/bin/claude`  
3. **System Wide** - `/usr/local/bin/claude`
4. **User Local** - `~/.local/bin/claude`
5. **Project Local** - `~/node_modules/.bin/claude`
6. **Yarn Global** - `~/.yarn/bin/claude`

```go
func FindClaudeCLI() (string, error) {
    // 1. System PATH first
    if path, err := exec.LookPath("claude"); err == nil {
        return path, nil
    }
    
    // 2-6. Check specific locations in order
    locations := []string{
        filepath.Join(homeDir, ".npm-global", "bin", "claude"),
        "/usr/local/bin/claude",
        filepath.Join(homeDir, ".local", "bin", "claude"),
        filepath.Join(homeDir, "node_modules", ".bin", "claude"),
        filepath.Join(homeDir, ".yarn", "bin", "claude"),
    }
    
    for _, location := range locations {
        if _, err := os.Stat(location); err == nil {
            return location, nil
        }
    }
    
    return "", NewCLINotFoundError("", "Claude CLI not found in any expected location")
}
```

## Command Building Patterns

### One-Shot vs Streaming Modes
- **One-Shot (Query)**: `--print --output-format stream-json`
- **Streaming (Client)**: `--input-format stream-json --output-format stream-json`

### Dynamic Argument Construction
```go
func BuildCommand(options *Options, closeStdin bool) []string {
    args := []string{}
    
    // Core flags
    if closeStdin {
        args = append(args, "--print")
    } else {
        args = append(args, "--input-format", "stream-json")
    }
    args = append(args, "--output-format", "stream-json", "--verbose")
    
    // Options-based flags
    if options.SystemPrompt != nil {
        args = append(args, "--system-prompt", *options.SystemPrompt)
    }
    
    if options.Model != nil {
        args = append(args, "--model", *options.Model)
    }
    
    // ... continue for all options
    
    return args
}
```

### ExtraArgs Support
```go
// Handle map[string]*string for custom flags
for key, value := range options.ExtraArgs {
    if value == nil {
        args = append(args, "--"+key)  // Boolean flag
    } else {
        args = append(args, "--"+key, *value)  // Flag with value
    }
}
```

## Node.js Dependency Validation
```go
func ValidateNodeJS() error {
    _, err := exec.LookPath("node")
    if err != nil {
        return NewCLINotFoundError("node", 
            "Node.js is required for Claude CLI. Install from https://nodejs.org/")
    }
    return nil
}
```

## CLI Version Detection
```go
func DetectCLIVersion(cliPath string) (string, error) {
    cmd := exec.Command(cliPath, "--version")
    output, err := cmd.Output()
    if err != nil {
        return "", err
    }
    return strings.TrimSpace(string(output)), nil
}
```

## Error Handling with Installation Guidance

### CLINotFoundError Enhancement
```go
func NewCLINotFoundError(path string, message string) *CLINotFoundError {
    helpfulMessage := message + "\n\n" +
        "Installation options:\n" +
        "1. npm install -g @anthropic-ai/claude-code\n" +
        "2. yarn global add @anthropic-ai/claude-code\n" +
        "3. Download from official releases\n\n" +
        "Ensure Node.js is installed: https://nodejs.org/"
    
    return &CLINotFoundError{
        BaseError: BaseError{message: helpfulMessage},
        Path:      path,
    }
}
```

## Working Directory Validation
```go
func ValidateWorkingDirectory(cwd string) error {
    if cwd == "" {
        return nil // Use current directory
    }
    
    info, err := os.Stat(cwd)
    if err != nil {
        return fmt.Errorf("working directory %q does not exist: %w", cwd, err)
    }
    
    if !info.IsDir() {
        return fmt.Errorf("working directory %q is not a directory", cwd)
    }
    
    return nil
}
```

## Configuration Flag Mapping

### All Options Support
Must support all fields from Options struct:
- **Tool Control**: `--allowed-tools`, `--disallowed-tools`
- **System Prompts**: `--system-prompt`, `--append-system-prompt`  
- **Model Config**: `--model`, `--max-thinking-tokens`
- **Permission System**: `--permission-mode`, `--permission-prompt-tool-name`
- **Session Management**: `--continue-conversation`, `--resume`, `--max-turns`
- **File System**: `--cwd`, `--add-dirs`
- **MCP Integration**: `--mcp-servers`

### MCP Server Configuration
```go
func BuildMCPFlags(servers map[string]McpServerConfig) []string {
    if len(servers) == 0 {
        return nil
    }
    
    // Convert to JSON and pass as flag
    data, _ := json.Marshal(servers)
    return []string{"--mcp-servers", string(data)}
}
```

## Cross-Platform Considerations

### Windows Differences
- Use `claude.exe` extension
- Different path separators
- PowerShell vs Command Prompt compatibility

### File Path Handling
```go
func GetExecutableName() string {
    if runtime.GOOS == "windows" {
        return "claude.exe"
    }
    return "claude"
}
```

## Integration Requirements
- Must integrate with subprocess component for process execution
- Provide clear CLI path discovery for transport layer
- Support all configuration options from Options struct
- Handle all error cases with helpful messages

## Testing Considerations
- Mock CLI discovery for unit tests
- Test all argument combinations
- Validate cross-platform compatibility
- Test error scenarios with helpful messages