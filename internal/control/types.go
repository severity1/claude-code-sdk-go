// Package control provides the control protocol for bidirectional communication
// between the SDK and CLI. This mirrors the Python SDK's Query class functionality.
package control

import (
	"context"
	"time"
)

// Message type constants matching Python SDK
const (
	TypeControlRequest  = "control_request"
	TypeControlResponse = "control_response"
)

// Subtype constants matching Python SDK SDKControl*Request types
const (
	SubtypeInitialize        = "initialize"
	SubtypeInterrupt         = "interrupt"
	SubtypeCanUseTool        = "can_use_tool"
	SubtypeSetPermissionMode = "set_permission_mode"
	SubtypeSetModel          = "set_model"
	SubtypeRewindFiles       = "rewind_files"
	SubtypeHookCallback      = "hook_callback"
	SubtypeMcpMessage        = "mcp_message"
)

// Response subtype constants
const (
	SubtypeSuccess = "success"
	SubtypeError   = "error"
)

// DefaultInitTimeout is the default timeout for initialization handshake.
const DefaultInitTimeout = 60 * time.Second

// Request represents a control request sent to CLI (matches Python SDKControlRequest).
type Request struct {
	Type      string         `json:"type"`       // Always "control_request"
	RequestID string         `json:"request_id"` // Format: "req_{n}_{hex}"
	Request   map[string]any `json:"request"`    // Contains subtype and payload
}

// Response represents a control response from CLI (matches Python SDKControlResponse).
type Response struct {
	Type     string          `json:"type"`     // Always "control_response"
	Response ResponsePayload `json:"response"` // Contains subtype, request_id, response/error
}

// ResponsePayload is the inner response structure.
type ResponsePayload struct {
	Subtype   string         `json:"subtype"`            // "success" or "error"
	RequestID string         `json:"request_id"`         // Matches the request
	Response  map[string]any `json:"response,omitempty"` // Response data for success
	Error     string         `json:"error,omitempty"`    // Error message for error
}

// InitializeRequest is the payload for initialize control requests.
type InitializeRequest struct {
	Subtype string         `json:"subtype"` // Always "initialize"
	Hooks   map[string]any `json:"hooks,omitempty"`
}

// InitializeResponse is the response from CLI initialization.
type InitializeResponse struct {
	Commands    []string `json:"commands,omitempty"`
	OutputStyle string   `json:"output_style,omitempty"`
}

// InterruptRequest is the payload for interrupt control requests.
type InterruptRequest struct {
	Subtype string `json:"subtype"` // Always "interrupt"
}

// SetPermissionModeRequest is the payload for set_permission_mode control requests.
type SetPermissionModeRequest struct {
	Subtype string `json:"subtype"` // Always "set_permission_mode"
	Mode    string `json:"mode"`
}

// SetModelRequest is the payload for set_model control requests.
type SetModelRequest struct {
	Subtype string  `json:"subtype"`         // Always "set_model"
	Model   *string `json:"model,omitempty"` // nil to reset to default
}

// RewindFilesRequest is the payload for rewind_files control requests.
type RewindFilesRequest struct {
	Subtype       string `json:"subtype"`         // Always "rewind_files"
	UserMessageID string `json:"user_message_id"` // UUID of the checkpoint
}

// CanUseToolRequest is received from CLI for permission callbacks (incoming request).
type CanUseToolRequest struct {
	Subtype               string         `json:"subtype"` // Always "can_use_tool"
	ToolName              string         `json:"tool_name"`
	Input                 map[string]any `json:"input"`
	PermissionSuggestions []any          `json:"permission_suggestions,omitempty"`
	BlockedPath           *string        `json:"blocked_path,omitempty"`
}

// CanUseToolResponse is sent back to CLI after permission callback.
type CanUseToolResponse struct {
	Behavior           string         `json:"behavior"` // "allow" or "deny"
	UpdatedInput       map[string]any `json:"updatedInput,omitempty"`
	UpdatedPermissions []any          `json:"updatedPermissions,omitempty"`
	Message            string         `json:"message,omitempty"`
	Interrupt          bool           `json:"interrupt,omitempty"`
}

// HookCallbackRequest is received from CLI for hook callbacks (incoming request).
type HookCallbackRequest struct {
	Subtype    string  `json:"subtype"` // Always "hook_callback"
	CallbackID string  `json:"callback_id"`
	Input      any     `json:"input"`
	ToolUseID  *string `json:"tool_use_id,omitempty"`
}

// McpMessageRequest is the payload for mcp_message control requests.
type McpMessageRequest struct {
	Subtype    string         `json:"subtype"` // Always "mcp_message"
	ServerName string         `json:"server_name"`
	Message    map[string]any `json:"message"`
}

// HookMatcher configures hook matching for a specific event.
type HookMatcher struct {
	Matcher         any      `json:"matcher,omitempty"`
	HookCallbackIDs []string `json:"hookCallbackIds"`
	Timeout         *int     `json:"timeout,omitempty"`
}

// Transport is the minimal interface required by Query for control protocol communication.
// This avoids import cycles by not depending on the full Transport interface.
type Transport interface {
	// Write sends raw bytes to the CLI subprocess stdin.
	Write(ctx context.Context, data []byte) error
}

// CanUseToolHandler processes incoming can_use_tool requests from CLI.
type CanUseToolHandler func(ctx context.Context, req CanUseToolRequest) (CanUseToolResponse, error)

// HookHandler processes incoming hook_callback requests from CLI.
type HookHandler func(ctx context.Context, input any, toolUseID *string) (map[string]any, error)
