# Core Types Analysis

Complete analysis of the Python SDK's type system including messages, content blocks, and configuration options.

## Type System Overview (types.py)

The Python SDK uses a sophisticated type system with dataclasses, TypedDict, and union types:

```python
from dataclasses import dataclass, field
from typing import Any, Literal, TypedDict
from typing_extensions import NotRequired  # For Python < 3.11 compatibility
```

## Permission System

**Permission Modes**:
```python
PermissionMode = Literal["default", "acceptEdits", "plan", "bypassPermissions"]
```

**Go Implementation**:
```go
type PermissionMode string

const (
    PermissionModeDefault     PermissionMode = "default"
    PermissionModeAcceptEdits PermissionMode = "acceptEdits" 
    PermissionModePlan        PermissionMode = "plan"
    PermissionModeBypass      PermissionMode = "bypassPermissions"
)
```

## MCP Server Configuration

**Python Union Type Approach**:
```python
class McpStdioServerConfig(TypedDict):
    type: NotRequired[Literal["stdio"]]  # Optional for backwards compatibility
    command: str
    args: NotRequired[list[str]]
    env: NotRequired[dict[str, str]]

class McpSSEServerConfig(TypedDict):
    type: Literal["sse"]
    url: str
    headers: NotRequired[dict[str, str]]

class McpHttpServerConfig(TypedDict):
    type: Literal["http"]
    url: str
    headers: NotRequired[dict[str, str]]

McpServerConfig = McpStdioServerConfig | McpSSEServerConfig | McpHttpServerConfig
```

**Go Interface Approach**:
```go
type McpServerConfig interface {
    GetType() string
}

type McpStdioServerConfig struct {
    Type    string            `json:"type,omitempty"`
    Command string            `json:"command"`
    Args    []string          `json:"args,omitempty"`
    Env     map[string]string `json:"env,omitempty"`
}

func (m *McpStdioServerConfig) GetType() string {
    if m.Type == "" {
        return "stdio" // Default for backwards compatibility
    }
    return m.Type
}

type McpSSEServerConfig struct {
    Type    string            `json:"type"`
    URL     string            `json:"url"`
    Headers map[string]string `json:"headers,omitempty"`
}

func (m *McpSSEServerConfig) GetType() string { return "sse" }

type McpHttpServerConfig struct {
    Type    string            `json:"type"`
    URL     string            `json:"url"`
    Headers map[string]string `json:"headers,omitempty"`
}

func (m *McpHttpServerConfig) GetType() string { return "http" }
```

## Content Block Types

**Python Dataclass Approach**:
```python
@dataclass
class TextBlock:
    """Text content block."""
    text: str

@dataclass
class ThinkingBlock:
    """Thinking content block."""
    thinking: str
    signature: str

@dataclass
class ToolUseBlock:
    """Tool use content block."""
    id: str
    name: str
    input: dict[str, Any]

@dataclass
class ToolResultBlock:
    """Tool result content block."""
    tool_use_id: str
    content: str | list[dict[str, Any]] | None = None
    is_error: bool | None = None

ContentBlock = TextBlock | ThinkingBlock | ToolUseBlock | ToolResultBlock
```

**Go Struct Approach**:
```go
type ContentBlock interface {
    GetType() string
}

type TextBlock struct {
    Text string `json:"text"`
}

func (t *TextBlock) GetType() string { return "text" }

type ThinkingBlock struct {
    Thinking  string `json:"thinking"`
    Signature string `json:"signature"`
}

func (t *ThinkingBlock) GetType() string { return "thinking" }

type ToolUseBlock struct {
    ID    string                 `json:"id"`
    Name  string                 `json:"name"`
    Input map[string]interface{} `json:"input"`
}

func (t *ToolUseBlock) GetType() string { return "tool_use" }

type ToolResultBlock struct {
    ToolUseID string      `json:"tool_use_id"`
    Content   interface{} `json:"content,omitempty"`
    IsError   *bool       `json:"is_error,omitempty"`
}

func (t *ToolResultBlock) GetType() string { return "tool_result" }
```

## Message Types

**Python Message Hierarchy**:
```python
@dataclass
class UserMessage:
    """User message."""
    content: str | list[ContentBlock]

@dataclass
class AssistantMessage:
    """Assistant message with content blocks."""
    content: list[ContentBlock]
    model: str

@dataclass
class SystemMessage:
    """System message with metadata."""
    subtype: str
    data: dict[str, Any]

@dataclass
class ResultMessage:
    """Result message with cost and usage information."""
    subtype: str
    duration_ms: int
    duration_api_ms: int
    is_error: bool
    num_turns: int
    session_id: str
    total_cost_usd: float | None = None
    usage: dict[str, Any] | None = None
    result: str | None = None

Message = UserMessage | AssistantMessage | SystemMessage | ResultMessage
```

**Go Message Implementation**:
```go
type Message interface {
    GetType() string
}

type UserMessage struct {
    Content []ContentBlock `json:"content"`
}

func (u *UserMessage) GetType() string { return "user" }

type AssistantMessage struct {
    Content []ContentBlock `json:"content"`
    Model   string         `json:"model"`
}

func (a *AssistantMessage) GetType() string { return "assistant" }

type SystemMessage struct {
    Subtype string                 `json:"subtype"`
    Data    map[string]interface{} `json:"data"`
}

func (s *SystemMessage) GetType() string { return "system" }

type ResultMessage struct {
    Subtype       string                 `json:"subtype"`
    DurationMS    int                    `json:"duration_ms"`
    DurationAPIMS int                    `json:"duration_api_ms"`
    IsError       bool                   `json:"is_error"`
    NumTurns      int                    `json:"num_turns"`
    SessionID     string                 `json:"session_id"`
    TotalCostUSD  *float64               `json:"total_cost_usd,omitempty"`
    Usage         map[string]interface{} `json:"usage,omitempty"`
    Result        *string                `json:"result,omitempty"`
}

func (r *ResultMessage) GetType() string { return "result" }
```

## Configuration Options

**Python Options Class**:
```python
@dataclass
class ClaudeCodeOptions:
    """Query options for Claude SDK."""
    allowed_tools: list[str] = field(default_factory=list)
    max_thinking_tokens: int = 8000
    system_prompt: str | None = None
    append_system_prompt: str | None = None
    mcp_servers: dict[str, McpServerConfig] | str | Path = field(default_factory=dict)
    permission_mode: PermissionMode | None = None
    continue_conversation: bool = False
    resume: str | None = None
    max_turns: int | None = None
    disallowed_tools: list[str] = field(default_factory=list)
    model: str | None = None
    permission_prompt_tool_name: str | None = None
    cwd: str | Path | None = None
    settings: str | None = None
    add_dirs: list[str | Path] = field(default_factory=list)
    extra_args: dict[str, str | None] = field(default_factory=dict)
```

**Go Options with Builder Pattern**:
```go
type Options struct {
    AllowedTools             []string
    MaxThinkingTokens        int
    SystemPrompt             string
    AppendSystemPrompt       string
    McpServers               map[string]McpServerConfig
    PermissionMode           PermissionMode
    ContinueConversation     bool
    Resume                   string
    MaxTurns                 int
    DisallowedTools          []string
    Model                    string
    PermissionPromptToolName string
    Cwd                      string
    Settings                 string
    AddDirs                  []string
    ExtraArgs                map[string]*string // nil for boolean flags
}

func NewOptions() *Options {
    return &Options{
        MaxThinkingTokens: 8000,
        McpServers:        make(map[string]McpServerConfig),
        ExtraArgs:         make(map[string]*string),
    }
}

// Builder methods
func (o *Options) WithSystemPrompt(prompt string) *Options {
    o.SystemPrompt = prompt
    return o
}

func (o *Options) WithAllowedTools(tools ...string) *Options {
    o.AllowedTools = tools
    return o
}

func (o *Options) WithPermissionMode(mode PermissionMode) *Options {
    o.PermissionMode = mode
    return o
}
```

## Stream Message Protocol

**Internal Protocol Types**:
```python
# For streaming communication between SDK and CLI
class StreamMessage(TypedDict):
    type: str
    message: dict[str, Any] | None
    parent_tool_use_id: str | None
    session_id: str
    request_id: str | None  # For control messages
    request: dict[str, Any] | None
    response: dict[str, Any] | None
```

**Go Stream Message**:
```go
type StreamMessage struct {
    Type             string                 `json:"type"`
    Message          interface{}            `json:"message,omitempty"`
    ParentToolUseID  *string                `json:"parent_tool_use_id,omitempty"`
    SessionID        string                 `json:"session_id,omitempty"`
    RequestID        string                 `json:"request_id,omitempty"`
    Request          map[string]interface{} `json:"request,omitempty"`
    Response         map[string]interface{} `json:"response,omitempty"`
}
```

## Key Implementation Insights

1. **Mixed Content Support**: Both UserMessage and AssistantMessage can contain mixed content blocks
2. **Optional Fields**: Use pointer types in Go for optional fields (matching Python's None)
3. **Raw Data Preservation**: SystemMessage stores entire raw data dict for extensibility
4. **Union Type Handling**: Use interfaces with type discrimination methods
5. **Backward Compatibility**: MCP stdio type is optional for compatibility
6. **Type Safety**: Strong typing throughout with interface-based polymorphism

This type system provides the foundation for 100% API compatibility while being Go-native.