// Package control provides the SDK control protocol for bidirectional communication with Claude CLI.
// This package enables features like tool permission callbacks, hook callbacks, and MCP message routing.
package control

// Message type constants for control protocol discrimination.
const (
	// MessageTypeControlRequest is sent TO the CLI to request an action.
	MessageTypeControlRequest = "control_request"
	// MessageTypeControlResponse is received FROM the CLI as a response.
	MessageTypeControlResponse = "control_response"
)

// Request subtype constants matching Python SDK for 100% parity.
const (
	// SubtypeInterrupt requests interruption of current operation.
	SubtypeInterrupt = "interrupt"
	// SubtypeCanUseTool requests permission to use a tool.
	SubtypeCanUseTool = "can_use_tool"
	// SubtypeInitialize performs the control protocol handshake.
	SubtypeInitialize = "initialize"
	// SubtypeSetPermissionMode changes the permission mode at runtime.
	SubtypeSetPermissionMode = "set_permission_mode"
	// SubtypeSetModel changes the AI model at runtime.
	SubtypeSetModel = "set_model"
	// SubtypeHookCallback invokes a registered hook callback.
	SubtypeHookCallback = "hook_callback"
	// SubtypeMcpMessage routes an MCP message to an SDK MCP server.
	SubtypeMcpMessage = "mcp_message"
)

// Response subtype constants for control responses.
const (
	// ResponseSubtypeSuccess indicates the request succeeded.
	ResponseSubtypeSuccess = "success"
	// ResponseSubtypeError indicates the request failed.
	ResponseSubtypeError = "error"
)

// SDKControlRequest represents a control request sent TO the CLI.
// This is the envelope that wraps all control request types.
type SDKControlRequest struct {
	// Type is always MessageTypeControlRequest.
	Type string `json:"type"`
	// RequestID is a unique identifier for request/response correlation.
	// Format: req_{counter}_{random_hex}
	RequestID string `json:"request_id"`
	// Request contains the actual request payload (InterruptRequest, InitializeRequest, etc.).
	Request any `json:"request"`
}

// SDKControlResponse represents a control response received FROM the CLI.
// This is the envelope that wraps all control response types.
type SDKControlResponse struct {
	// Type is always MessageTypeControlResponse.
	Type string `json:"type"`
	// Response contains the actual response data.
	Response Response `json:"response"`
}

// Response is the inner response structure within SDKControlResponse.
type Response struct {
	// Subtype is either ResponseSubtypeSuccess or ResponseSubtypeError.
	Subtype string `json:"subtype"`
	// RequestID matches the request that this response is for.
	RequestID string `json:"request_id"`
	// Response contains the response data (only for success).
	Response any `json:"response,omitempty"`
	// Error contains the error message (only for error).
	Error string `json:"error,omitempty"`
}

// InterruptRequest requests interruption of the current operation.
type InterruptRequest struct {
	// Subtype is always SubtypeInterrupt.
	Subtype string `json:"subtype"`
}

// InitializeRequest performs the control protocol handshake.
// This must be sent before any other control requests in streaming mode.
type InitializeRequest struct {
	// Subtype is always SubtypeInitialize.
	Subtype string `json:"subtype"`
	// Hooks will be added in Issue #9 for hook registration.
}

// InitializeResponse contains the CLI's response to initialization.
type InitializeResponse struct {
	// SupportedCommands lists the control commands supported by this CLI version.
	SupportedCommands []string `json:"supported_commands,omitempty"`
}

// SetPermissionModeRequest changes the permission mode at runtime.
type SetPermissionModeRequest struct {
	// Subtype is always SubtypeSetPermissionMode.
	Subtype string `json:"subtype"`
	// Mode is the new permission mode to set.
	Mode string `json:"mode"`
}

// SetModelRequest changes the AI model at runtime.
// This matches Python SDK's set_model() behavior exactly.
type SetModelRequest struct {
	// Subtype is always SubtypeSetModel.
	Subtype string `json:"subtype"`
	// Model is the new model to use. Use nil to reset to default.
	// Examples: "claude-sonnet-4-5", "claude-opus-4-1-20250805"
	Model *string `json:"model"`
}
