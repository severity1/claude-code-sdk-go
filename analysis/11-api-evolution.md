# API Evolution Analysis

Analysis of Python SDK evolution patterns and recent feature additions critical for Go implementation.

## Version History and Feature Evolution

### Version 0.0.19 (Latest) - Recent Enhancements
**Key Additions**:
- **`add_dirs` Support**: `ClaudeCodeOptions.add_dirs` for `--add-dir` CLI flag
  - Allows multiple directory additions: `for directory in self._options.add_dirs: cmd.extend(["--add-dir", str(directory)])`
  - Supports both string and Path objects
- **MCP Stderr Fix**: Resolved hanging when MCP servers log to Claude Code stderr
  - Critical for subprocess stability with MCP integration

### Version 0.0.18 - Settings Configuration
**Key Additions**:
- **Settings Configuration**: `ClaudeCodeOptions.settings` for `--settings` CLI flag
  - Accepts JSON string or file path: `cmd.extend(["--settings", self._options.settings])`
  - Provides runtime configuration override capability

### Version 0.0.17 - Framework Compatibility Breakthrough
**Major Changes**:
- **Trio Support**: Removed dependency on asyncio for full trio compatibility
- **Framework Agnostic**: anyio-based implementation supports multiple async frameworks
- **Architecture Impact**: Core SDK doesn't import asyncio or trio directly

### Version 0.0.16 - Architectural Revolution
**Major Additions**:
- **ClaudeSDKClient Introduction**: Bidirectional streaming conversation capability
  - Fundamental architecture change adding persistent, stateful client
- **Message Input Support**: query() function now accepts Message objects, not just strings
- **Working Directory Validation**: Explicit error if cwd does not exist
  - `if self._cwd and not Path(self._cwd).exists(): raise CLIConnectionError(...)`

### Version 0.0.14 - Stability & Buffering Improvements
**Critical Fixes**:
- **Stderr Safety Limits**: Added safety limits to Claude Code CLI stderr reading
  - Prevents memory exhaustion from excessive stderr output
- **Multi-line Buffering**: Improved handling of JSON messages split across multiple stream reads
  - Foundation for speculative parsing strategy

### Version 0.0.13 - Protocol Alignment & Field Evolution
**Important Changes**:
- **MCP Type Updates**: Updated MCP (Model Context Protocol) types to align with Claude Code
- **Multi-line Buffer Fix**: Fixed multi-line buffering issue
- **Cost Field Rename**: `cost_usd` → `total_cost_usd` in API responses
  - **CRITICAL**: Go implementation must use `total_cost_usd`, not `cost_usd`
- **Optional Cost Handling**: Fixed optional cost fields handling

## Evolution Patterns Analysis

### 1. Incremental Feature Addition Pattern
Each version adds specific CLI flag support:
- v0.0.19: `--add-dir` support
- v0.0.18: `--settings` support  
- Earlier versions: Various tool and configuration flags

**Go Implementation Implication**: Must support all current CLI flags and be extensible for future additions.

### 2. Stability Focus Pattern
Regular fixes for buffering and process management:
- v0.0.14: Stderr safety limits
- v0.0.13: Multi-line buffering fixes
- Ongoing: Process lifecycle improvements

**Go Implementation Implication**: Must implement robust buffering and process management from the start.

### 3. Framework Compatibility Evolution
Progressive improvement of async framework support:
- Early versions: asyncio-specific
- v0.0.17: anyio adoption for framework agnosticism
- Current: Works with asyncio, trio, and others

**Go Implementation Implication**: Go's unified goroutine model provides natural framework agnosticism.

### 4. Protocol Alignment Pattern
Continuous alignment with Claude Code CLI changes:
- MCP type updates
- Cost field naming changes
- New CLI flag support

**Go Implementation Implication**: Must track Python SDK changes and CLI evolution.

### 5. Backward Compatibility Maintenance
Changes maintain existing API surface:
- New fields are optional
- New methods don't break existing usage
- Configuration changes use sensible defaults

**Go Implementation Implication**: Design for forward compatibility with optional fields and extensible configuration.

## Recent Feature Implementation Details

### add_dirs Implementation
**Python Implementation**:
```python
if self._options.add_dirs:
    # Convert all paths to strings and add each directory
    for directory in self._options.add_dirs:
        cmd.extend(["--add-dir", str(directory)])
```

**Go Implementation Required**:
```go
if len(options.AddDirs) > 0 {
    for _, dir := range options.AddDirs {
        args = append(args, "--add-dir", dir)
    }
}
```

### settings Implementation
**Python Implementation**:
```python
if self._options.settings:
    cmd.extend(["--settings", self._options.settings])
```

**Go Implementation Required**:
```go
if options.Settings != "" {
    args = append(args, "--settings", options.Settings)
}
```

### MCP Stderr Isolation
**Problem**: MCP servers logging to stderr caused SDK to hang
**Solution**: Enhanced stderr management with proper isolation
**Go Implementation Required**: Robust stderr handling with temporary files and proper cleanup

## Field Evolution Tracking

### Critical Field Name Changes
- **ResultMessage.cost_usd** → **ResultMessage.total_cost_usd** (v0.0.13)
- This affects JSON parsing and struct field names

**Go Implementation**:
```go
type ResultMessage struct {
    // ... other fields ...
    TotalCostUSD *float64 `json:"total_cost_usd,omitempty"` // NOT cost_usd
    // ... other fields ...
}
```

### Optional Field Evolution
Fields have become optional over time:
- `total_cost_usd`: Always optional (`float | None`)
- `usage`: Always optional (`dict[str, Any] | None`)
- `result`: Always optional (`str | None`)

**Go Implementation**:
```go
// Use pointer types for optional fields
type ResultMessage struct {
    TotalCostUSD *float64               `json:"total_cost_usd,omitempty"`
    Usage        map[string]interface{} `json:"usage,omitempty"`
    Result       *string                `json:"result,omitempty"`
}
```

## CLI Flag Evolution

### Supported CLI Flags (as of v0.0.19)
```python
# Base flags
cmd = [self._cli_path, "--output-format", "stream-json", "--verbose"]

# Configuration flags
--system-prompt
--append-system-prompt
--allowedTools (comma-separated)
--disallowedTools (comma-separated)
--max-turns
--model
--permission-prompt-tool
--permission-mode
--continue
--resume
--settings           # Added v0.0.18
--add-dir           # Added v0.0.19 (multiple)
--mcp-config

# Input/output format flags
--input-format stream-json    # For streaming mode
--print <prompt>             # For string mode
```

### Extra Args Support
```python
# Support for arbitrary future CLI flags
for flag, value in self._options.extra_args.items():
    if value is None:
        cmd.append(f"--{flag}")      # Boolean flag
    else:
        cmd.extend([f"--{flag}", str(value)])  # Flag with value
```

**Go Implementation Pattern**:
```go
type Options struct {
    // ... standard fields ...
    ExtraArgs map[string]*string `json:"extra_args,omitempty"`
}

// In command building
for flag, value := range options.ExtraArgs {
    if value == nil {
        args = append(args, "--"+flag)
    } else {
        args = append(args, "--"+flag, *value)
    }
}
```

## Future-Proofing Strategies

### 1. Extensible Configuration
Design configuration to easily add new CLI flags:
```go
type Options struct {
    // Current fields...
    ExtraArgs map[string]*string // Future CLI flags
}
```

### 2. Version Detection
Implement CLI version detection for feature compatibility:
```go
func (t *Transport) detectCLIVersion() (string, error) {
    cmd := exec.Command(t.cliPath, "--version")
    output, err := cmd.Output()
    // Parse version and determine feature support
}
```

### 3. Optional Field Handling
Use pointer types consistently for optional fields:
```go
// Always use pointers for optional fields
type ResultMessage struct {
    TotalCostUSD *float64 `json:"total_cost_usd,omitempty"`
    Usage        *map[string]interface{} `json:"usage,omitempty"`
    Result       *string  `json:"result,omitempty"`
}
```

### 4. Protocol Buffer Strategy
Consider protobuf or similar for future protocol stability:
- Type-safe message evolution
- Backward/forward compatibility
- Efficient serialization

## Go SDK Evolution Recommendations

1. **Track Python SDK releases** and incorporate changes promptly
2. **Implement all current CLI flags** including latest additions
3. **Use correct field names** (total_cost_usd, not cost_usd)
4. **Design for extensibility** with ExtraArgs pattern
5. **Implement robust stderr handling** to prevent MCP hanging issues
6. **Version compatibility checking** for CLI feature support
7. **Comprehensive testing** for all supported CLI configurations

This evolution analysis ensures the Go SDK will maintain compatibility as the Python SDK and Claude Code CLI continue to evolve.