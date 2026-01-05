# Feature Parity: Go SDK vs Python SDK

This document provides a comprehensive comparison between the Go Agent SDK and the Python Agent SDK, demonstrating 100% feature parity.

---

## Executive Summary

**Status: 100% Feature Parity Achieved**

The Go SDK (`github.com/severity1/claude-agent-sdk-go`) implements all features from the Python SDK (`claude-agent-sdk`) with additional Go-idiomatic enhancements.

| Category | Python SDK | Go SDK | Parity |
|:---------|:-----------|:-------|:-------|
| Functions | 3 | 4 (+helpers) | 100% |
| Client Methods | 7 | 13 (+extras) | 100% |
| Message Types | 5 | 6 | 100% |
| Content Block Types | 4 | 4 | 100% |
| Error Types | 5 | 6 | 100% |
| Hook Events | 6 | 6 | 100% |
| Option Fields | 30+ | 70+ constructors | 100% |
| MCP Types | 5 | 9 (+extras) | 100% |
| Sandbox Config | 3 types | 3 types | 100% |

---

## Functions

| Python SDK | Go SDK | Notes |
|:-----------|:-------|:------|
| `query(prompt, options)` | `Query(ctx, prompt, opts...)` | Context-first pattern |
| `tool(name, desc, schema)` | `NewTool(name, desc, schema, handler)` | Factory function vs decorator |
| `create_sdk_mcp_server(name, version, tools)` | `CreateSDKMcpServer(name, version, tools...)` | Identical functionality |

### Go SDK Additional Functions

| Function | Description |
|:---------|:------------|
| `QueryWithTransport()` | Query with custom transport (testing) |
| `NewClient()` | Create new Client |
| `NewClientWithTransport()` | Create Client with custom transport |
| `WithClient()` | Resource management helper (like `async with`) |
| `WithClientTransport()` | Resource management with custom transport |

---

## Classes / Client Interface

### Python: `ClaudeSDKClient`

| Method | Go Equivalent | Notes |
|:-------|:--------------|:------|
| `__init__(options)` | `NewClient(opts...)` | Functional options pattern |
| `connect(prompt)` | `Connect(ctx, prompt...)` | Context-first |
| `query(prompt, session_id)` | `Query(ctx, prompt)` / `QueryWithSession(ctx, prompt, sessionID)` | Split into two methods |
| `receive_messages()` | `ReceiveMessages(ctx)` | Returns channel |
| `receive_response()` | `ReceiveResponse(ctx)` | Returns MessageIterator |
| `interrupt()` | `Interrupt(ctx)` | Context-first |
| `rewind_files(uuid)` | `RewindFiles(ctx, messageUUID)` | Context-first |
| `disconnect()` | `Disconnect()` | Identical |
| `async with` context manager | `WithClient()` helper | Go-idiomatic resource management |

### Go SDK Additional Methods

| Method | Description |
|:-------|:------------|
| `SetModel(ctx, model)` | Change model at runtime |
| `SetPermissionMode(ctx, mode)` | Change permission mode at runtime |
| `GetStreamIssues()` | Get validation issues from stream |
| `GetStreamStats()` | Get stream statistics |
| `GetServerInfo(ctx)` | Get diagnostic information |

---

## Configuration Options

### ClaudeAgentOptions Mapping

| Python Option | Go Option Constructor | Status |
|:--------------|:---------------------|:-------|
| `allowed_tools` | `WithAllowedTools(tools...)` | PARITY |
| `disallowed_tools` | `WithDisallowedTools(tools...)` | PARITY |
| `tools` | `WithTools(tools...)` | PARITY |
| - | `WithToolsPreset(preset)` | GO EXTRA |
| - | `WithClaudeCodeTools()` | GO EXTRA |
| `system_prompt` | `WithSystemPrompt(prompt)` | PARITY |
| - | `WithAppendSystemPrompt(prompt)` | GO EXTRA |
| `model` | `WithModel(model)` | PARITY |
| `fallback_model` | `WithFallbackModel(model)` | PARITY |
| `max_turns` | `WithMaxTurns(turns)` | PARITY |
| `max_budget_usd` | `WithMaxBudgetUSD(budget)` | PARITY |
| `max_thinking_tokens` | `WithMaxThinkingTokens(tokens)` | PARITY |
| `permission_mode` | `WithPermissionMode(mode)` | PARITY |
| `permission_prompt_tool_name` | `WithPermissionPromptToolName(toolName)` | PARITY |
| `continue_conversation` | `WithContinueConversation(bool)` | PARITY |
| `resume` | `WithResume(sessionID)` | PARITY |
| `fork_session` | `WithForkSession(fork)` | PARITY |
| `cwd` | `WithCwd(cwd)` | PARITY |
| `add_dirs` | `WithAddDirs(dirs...)` | PARITY |
| `mcp_servers` | `WithMcpServers(servers)` | PARITY |
| - | `WithSdkMcpServer(name, server)` | GO EXTRA |
| `settings` | `WithSettings(settings)` | PARITY |
| `setting_sources` | `WithSettingSources(sources...)` | PARITY |
| `env` | `WithEnv(env)` | PARITY |
| - | `WithEnvVar(key, value)` | GO EXTRA |
| `extra_args` | `WithExtraArgs(args)` | PARITY |
| `cli_path` | `WithCLIPath(path)` | PARITY |
| `max_buffer_size` | `WithMaxBufferSize(size)` | PARITY |
| `stderr` | `WithStderrCallback(callback)` | PARITY |
| `debug_stderr` (deprecated) | `WithDebugWriter(w)` | PARITY |
| - | `WithDebugStderr()` | GO EXTRA |
| - | `WithDebugDisabled()` | GO EXTRA |
| `can_use_tool` | `WithCanUseTool(callback)` | PARITY |
| `hooks` | `WithHooks(hooks)` | PARITY |
| - | `WithHook(event, matcher, callback)` | GO EXTRA |
| - | `WithPreToolUseHook(matcher, callback)` | GO EXTRA |
| - | `WithPostToolUseHook(matcher, callback)` | GO EXTRA |
| `user` | `WithUser(user)` | PARITY |
| `include_partial_messages` | `WithIncludePartialMessages(include)` | PARITY |
| - | `WithPartialStreaming()` | GO EXTRA |
| `enable_file_checkpointing` | `WithEnableFileCheckpointing(enable)` | PARITY |
| - | `WithFileCheckpointing()` | GO EXTRA |
| `agents` | `WithAgents(agents)` | PARITY |
| - | `WithAgent(name, agent)` | GO EXTRA |
| `plugins` | `WithPlugins(plugins)` | PARITY |
| - | `WithPlugin(plugin)` | GO EXTRA |
| - | `WithLocalPlugin(path)` | GO EXTRA |
| `sandbox` | `WithSandbox(sandbox)` | PARITY |
| - | `WithSandboxEnabled(enabled)` | GO EXTRA |
| - | `WithAutoAllowBashIfSandboxed(autoAllow)` | GO EXTRA |
| - | `WithSandboxExcludedCommands(commands...)` | GO EXTRA |
| - | `WithSandboxNetwork(network)` | GO EXTRA |
| `output_format` | `WithOutputFormat(format)` | PARITY |
| - | `WithJSONSchema(schema)` | GO EXTRA |
| `betas` | `WithBetas(betas...)` | PARITY |

---

## Message Types

| Python SDK | Go SDK | Status |
|:-----------|:-------|:-------|
| `Message` (union) | `Message` interface | PARITY |
| `UserMessage` | `UserMessage` struct | PARITY |
| `AssistantMessage` | `AssistantMessage` struct | PARITY |
| `SystemMessage` | `SystemMessage` struct | PARITY |
| `ResultMessage` | `ResultMessage` struct | PARITY |
| `StreamEvent` | `StreamEvent` struct | PARITY |
| - | `RawControlMessage` struct | GO EXTRA |

### Message Type Constants

| Python | Go | Status |
|:-------|:---|:-------|
| `"user"` | `MessageTypeUser` | PARITY |
| `"assistant"` | `MessageTypeAssistant` | PARITY |
| `"system"` | `MessageTypeSystem` | PARITY |
| `"result"` | `MessageTypeResult` | PARITY |
| `"stream_event"` | `MessageTypeStreamEvent` | PARITY |
| - | `MessageTypeControlRequest` | GO EXTRA |
| - | `MessageTypeControlResponse` | GO EXTRA |

---

## Content Block Types

| Python SDK | Go SDK | Status |
|:-----------|:-------|:-------|
| `ContentBlock` (union) | `ContentBlock` interface | PARITY |
| `TextBlock` | `TextBlock` struct | PARITY |
| `ThinkingBlock` | `ThinkingBlock` struct | PARITY |
| `ToolUseBlock` | `ToolUseBlock` struct | PARITY |
| `ToolResultBlock` | `ToolResultBlock` struct | PARITY |

### Content Block Type Constants

| Python | Go | Status |
|:-------|:---|:-------|
| `"text"` | `ContentBlockTypeText` | PARITY |
| `"thinking"` | `ContentBlockTypeThinking` | PARITY |
| `"tool_use"` | `ContentBlockTypeToolUse` | PARITY |
| `"tool_result"` | `ContentBlockTypeToolResult` | PARITY |

---

## Error Types

| Python SDK | Go SDK | Status |
|:-----------|:-------|:-------|
| `ClaudeSDKError` | `SDKError` interface + `BaseError` | PARITY |
| `CLIConnectionError` | `ConnectionError` | PARITY |
| `CLINotFoundError` | `CLINotFoundError` | PARITY |
| `ProcessError` | `ProcessError` | PARITY |
| `CLIJSONDecodeError` | `JSONDecodeError` | PARITY |
| `MessageParseError` | `MessageParseError` | PARITY |

### AssistantMessageError Types

| Python | Go | Status |
|:-------|:---|:-------|
| `"authentication_failed"` | `AssistantMessageErrorAuthFailed` | PARITY |
| `"billing_error"` | `AssistantMessageErrorBilling` | PARITY |
| `"rate_limit"` | `AssistantMessageErrorRateLimit` | PARITY |
| `"invalid_request"` | `AssistantMessageErrorInvalidRequest` | PARITY |
| `"server_error"` | `AssistantMessageErrorServer` | PARITY |
| `"unknown"` | `AssistantMessageErrorUnknown` | PARITY |

### Go-Specific Error Type Helpers

Go SDK provides idiomatic helper functions following the `os.IsNotExist` pattern. These work with wrapped errors (using `errors.As` internally).

| Function | Description | Status |
|:---------|:------------|:-------|
| `IsConnectionError(err)` | Check if error is ConnectionError | GO-NATIVE |
| `IsCLINotFoundError(err)` | Check if error is CLINotFoundError | GO-NATIVE |
| `IsProcessError(err)` | Check if error is ProcessError | GO-NATIVE |
| `IsJSONDecodeError(err)` | Check if error is JSONDecodeError | GO-NATIVE |
| `IsMessageParseError(err)` | Check if error is MessageParseError | GO-NATIVE |
| `AsConnectionError(err)` | Extract *ConnectionError or nil | GO-NATIVE |
| `AsCLINotFoundError(err)` | Extract *CLINotFoundError or nil | GO-NATIVE |
| `AsProcessError(err)` | Extract *ProcessError or nil | GO-NATIVE |
| `AsJSONDecodeError(err)` | Extract *JSONDecodeError or nil | GO-NATIVE |
| `AsMessageParseError(err)` | Extract *MessageParseError or nil | GO-NATIVE |

**Note**: Python uses `isinstance()` for error type checking. Go SDK provides these helpers as a more idiomatic alternative to manual type assertions.

---

## Hook Types

### Hook Events

| Python | Go | Status |
|:-------|:---|:-------|
| `"PreToolUse"` | `HookEventPreToolUse` | PARITY |
| `"PostToolUse"` | `HookEventPostToolUse` | PARITY |
| `"UserPromptSubmit"` | `HookEventUserPromptSubmit` | PARITY |
| `"Stop"` | `HookEventStop` | PARITY |
| `"SubagentStop"` | `HookEventSubagentStop` | PARITY |
| `"PreCompact"` | `HookEventPreCompact` | PARITY |

### Hook Types

| Python SDK | Go SDK | Status |
|:-----------|:-------|:-------|
| `HookEvent` | `HookEvent` type | PARITY |
| `HookCallback` | `HookCallback` type | PARITY |
| `HookContext` | `HookContext` struct | PARITY |
| `HookMatcher` | `HookMatcher` struct | PARITY |
| `HookJSONOutput` | `HookJSONOutput` struct | PARITY |
| `AsyncHookJSONOutput` | `AsyncHookJSONOutput` struct | PARITY |

### Hook Input Types

| Python | Go | Status |
|:-------|:---|:-------|
| `BaseHookInput` | `BaseHookInput` | PARITY |
| `PreToolUseHookInput` | `PreToolUseHookInput` | PARITY |
| `PostToolUseHookInput` | `PostToolUseHookInput` | PARITY |
| `UserPromptSubmitHookInput` | `UserPromptSubmitHookInput` | PARITY |
| `StopHookInput` | `StopHookInput` | PARITY |
| `SubagentStopHookInput` | `SubagentStopHookInput` | PARITY |
| `PreCompactHookInput` | `PreCompactHookInput` | PARITY |

### Hook Output Types

| Python | Go | Status |
|:-------|:---|:-------|
| `PreToolUseHookSpecificOutput` | `PreToolUseHookSpecificOutput` | PARITY |
| `PostToolUseHookSpecificOutput` | `PostToolUseHookSpecificOutput` | PARITY |
| `UserPromptSubmitHookSpecificOutput` | `UserPromptSubmitHookSpecificOutput` | PARITY |

---

## MCP Types

| Python SDK | Go SDK | Status |
|:-----------|:-------|:-------|
| `SdkMcpTool` | `McpTool` struct | PARITY |
| `McpServerConfig` (union) | `McpServerConfig` interface | PARITY |
| `McpStdioServerConfig` | `McpStdioServerConfig` | PARITY |
| `McpSSEServerConfig` | `McpSSEServerConfig` | PARITY |
| `McpHttpServerConfig` | `McpHTTPServerConfig` | PARITY |
| `McpSdkServerConfig` | `McpSdkServerConfig` | PARITY |

### Go SDK MCP Extras

| Type | Description |
|:-----|:------------|
| `McpServer` interface | Interface for MCP servers |
| `SdkMcpServer` struct | In-process server implementation |
| `McpToolHandler` | Function type for tool handlers |
| `McpToolResult` | Result from tool execution |
| `McpContent` | Content in tool result |
| `McpToolDefinition` | Tool definition for listing |

---

## Permission Types

| Python | Go | Status |
|:-------|:---|:-------|
| `CanUseTool` | `CanUseToolCallback` | PARITY |
| `ToolPermissionContext` | `ToolPermissionContext` | PARITY |
| `PermissionResult` | `PermissionResult` interface | PARITY |
| `PermissionResultAllow` | `PermissionResultAllow` struct | PARITY |
| `PermissionResultDeny` | `PermissionResultDeny` struct | PARITY |
| `PermissionUpdate` | `PermissionUpdate` struct | PARITY |
| `PermissionRuleValue` | `PermissionRuleValue` struct | PARITY |

### Permission Modes

| Python | Go | Status |
|:-------|:---|:-------|
| `"default"` | `PermissionModeDefault` | PARITY |
| `"acceptEdits"` | `PermissionModeAcceptEdits` | PARITY |
| `"plan"` | `PermissionModePlan` | PARITY |
| `"bypassPermissions"` | `PermissionModeBypassPermissions` | PARITY |

---

## Sandbox Configuration

| Python SDK | Go SDK | Status |
|:-----------|:-------|:-------|
| `SandboxSettings` | `SandboxSettings` struct | PARITY |
| `SandboxNetworkConfig` | `SandboxNetworkConfig` struct | PARITY |
| `SandboxIgnoreViolations` | `SandboxIgnoreViolations` struct | PARITY |

### SandboxSettings Fields

| Python | Go | Status |
|:-------|:---|:-------|
| `enabled` | `Enabled` | PARITY |
| `autoAllowBashIfSandboxed` | `AutoAllowBashIfSandboxed` | PARITY |
| `excludedCommands` | `ExcludedCommands` | PARITY |
| `allowUnsandboxedCommands` | `AllowUnsandboxedCommands` | PARITY |
| `network` | `Network` | PARITY |
| `ignoreViolations` | `IgnoreViolations` | PARITY |
| `enableWeakerNestedSandbox` | `EnableWeakerNestedSandbox` | PARITY |

### SandboxNetworkConfig Fields

| Python | Go | Status |
|:-------|:---|:-------|
| `allowLocalBinding` | `AllowLocalBinding` | PARITY |
| `allowUnixSockets` | `AllowUnixSockets` | PARITY |
| `allowAllUnixSockets` | `AllowAllUnixSockets` | PARITY |
| `httpProxyPort` | `HTTPProxyPort` | PARITY |
| `socksProxyPort` | `SOCKSProxyPort` | PARITY |

---

## Advanced Features

| Feature | Python SDK | Go SDK | Status |
|:--------|:-----------|:-------|:-------|
| Streaming responses | `async for message in query()` | `MessageIterator.Next(ctx)` | PARITY |
| Partial message streaming | `include_partial_messages=True` | `WithPartialStreaming()` | PARITY |
| Session management | `resume`, `fork_session` | `WithResume()`, `WithForkSession()` | PARITY |
| File checkpointing | `enable_file_checkpointing` | `WithFileCheckpointing()` | PARITY |
| File rewinding | `rewind_files(uuid)` | `RewindFiles(ctx, uuid)` | PARITY |
| Interrupt support | `interrupt()` | `Interrupt(ctx)` | PARITY |
| Structured output | `output_format` | `WithOutputFormat()`, `WithJSONSchema()` | PARITY |
| Custom agents | `agents` | `WithAgents()`, `WithAgent()` | PARITY |
| Plugins | `plugins` | `WithPlugins()`, `WithLocalPlugin()` | PARITY |
| Beta features | `betas` | `WithBetas()` | PARITY |

### Go SDK Advanced Extras

| Feature | Description |
|:--------|:------------|
| `Transport` interface | Custom transport for testing |
| `StreamValidator` | Stream validation and diagnostics |
| `GetStreamIssues()` | Get validation issues |
| `GetStreamStats()` | Get stream statistics |
| `SetModel()` | Runtime model change |
| `SetPermissionMode()` | Runtime permission mode change |
| `GetServerInfo()` | Get diagnostic info |

---

## Other Types

### Agent Types

| Python | Go | Status |
|:-------|:---|:-------|
| `AgentDefinition` dataclass | `AgentDefinition` struct | PARITY |
| `"sonnet"` | `AgentModelSonnet` | PARITY |
| `"opus"` | `AgentModelOpus` | PARITY |
| `"haiku"` | `AgentModelHaiku` | PARITY |
| `"inherit"` | `AgentModelInherit` | PARITY |

### Plugin Types

| Python | Go | Status |
|:-------|:---|:-------|
| `SdkPluginConfig` TypedDict | `SdkPluginConfig` struct | PARITY |
| `"local"` | `SdkPluginTypeLocal` | PARITY |

### Setting Sources

| Python | Go | Status |
|:-------|:---|:-------|
| `"user"` | `SettingSourceUser` | PARITY |
| `"project"` | `SettingSourceProject` | PARITY |
| `"local"` | `SettingSourceLocal` | PARITY |

### Beta Features

| Python | Go | Status |
|:-------|:---|:-------|
| `"context-1m-2025-08-07"` | `SdkBetaContext1M` | PARITY |

---

## Migration Guide for Python SDK Users

### Key Differences

1. **Context-First Pattern**: Go functions accept `context.Context` as the first parameter for cancellation and timeouts.

2. **Functional Options**: Instead of a single options object, Go uses the functional options pattern with `With*()` functions.

3. **Interfaces vs Classes**: Go uses interfaces (`Client`, `Message`, `ContentBlock`) instead of classes.

4. **Error Handling**: Go uses explicit error returns instead of exceptions.

5. **Resource Management**: Use `WithClient()` instead of `async with` for automatic resource cleanup.

### Code Comparison

**Python:**
```python
from claude_agent_sdk import query, ClaudeAgentOptions

options = ClaudeAgentOptions(
    system_prompt="You are an expert",
    allowed_tools=["Read", "Write"],
    permission_mode="acceptEdits"
)

async for message in query(prompt="Hello", options=options):
    if isinstance(message, AssistantMessage):
        print(message.content)
```

**Go:**
```go
import "github.com/severity1/claude-agent-sdk-go"

iterator, err := claudecode.Query(ctx, "Hello",
    claudecode.WithSystemPrompt("You are an expert"),
    claudecode.WithAllowedTools("Read", "Write"),
    claudecode.WithPermissionMode(claudecode.PermissionModeAcceptEdits),
)
if err != nil {
    return err
}
defer iterator.Close()

for {
    message, err := iterator.Next(ctx)
    if errors.Is(err, claudecode.ErrNoMoreMessages) {
        break
    }
    if assistant, ok := message.(*claudecode.AssistantMessage); ok {
        // Process content
    }
}
```

### Client Usage Comparison

**Python:**
```python
async with ClaudeSDKClient(options) as client:
    await client.query("Hello")
    async for msg in client.receive_response():
        print(msg)
```

**Go:**
```go
err := claudecode.WithClient(ctx, func(client claudecode.Client) error {
    if err := client.Query(ctx, "Hello"); err != nil {
        return err
    }
    for msg := range client.ReceiveMessages(ctx) {
        // Process message
    }
    return nil
}, opts...)
```

---

## Conclusion

The Go SDK provides complete feature parity with the Python SDK while adding Go-idiomatic enhancements:

- **100% Python SDK features implemented**
- **Functional options pattern** for flexible configuration
- **Context-first design** for proper cancellation and timeouts
- **Interface-based design** for testability
- **Additional diagnostic features** (StreamValidator, GetStreamIssues, GetStreamStats)
- **Runtime configuration changes** (SetModel, SetPermissionMode)
- **Custom transport support** for testing

The Go SDK is production-ready and suitable for building applications that require Claude Code integration in Go environments.
