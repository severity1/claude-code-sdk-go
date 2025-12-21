package control

import (
	"encoding/json"
	"testing"
)

// TestRequestSerialization tests JSON serialization of control requests.
func TestRequestSerialization(t *testing.T) {
	tests := []struct {
		name     string
		request  Request
		expected string
	}{
		{
			name: "initialize_request",
			request: Request{
				Type:      TypeControlRequest,
				RequestID: "req_1_abcd1234",
				Request: map[string]any{
					"subtype": SubtypeInitialize,
					"hooks":   nil,
				},
			},
			expected: `{"type":"control_request","request_id":"req_1_abcd1234","request":{"subtype":"initialize","hooks":null}}`,
		},
		{
			name: "interrupt_request",
			request: Request{
				Type:      TypeControlRequest,
				RequestID: "req_2_efgh5678",
				Request: map[string]any{
					"subtype": SubtypeInterrupt,
				},
			},
			expected: `{"type":"control_request","request_id":"req_2_efgh5678","request":{"subtype":"interrupt"}}`,
		},
		{
			name: "set_permission_mode_request",
			request: Request{
				Type:      TypeControlRequest,
				RequestID: "req_3_ijkl9012",
				Request: map[string]any{
					"subtype": SubtypeSetPermissionMode,
					"mode":    "bypassPermissions",
				},
			},
			expected: `{"type":"control_request","request_id":"req_3_ijkl9012","request":{"subtype":"set_permission_mode","mode":"bypassPermissions"}}`,
		},
		{
			name: "set_model_request",
			request: Request{
				Type:      TypeControlRequest,
				RequestID: "req_4_mnop3456",
				Request: map[string]any{
					"subtype": SubtypeSetModel,
					"model":   "claude-sonnet-4-5",
				},
			},
			expected: `{"type":"control_request","request_id":"req_4_mnop3456","request":{"subtype":"set_model","model":"claude-sonnet-4-5"}}`,
		},
		{
			name: "rewind_files_request",
			request: Request{
				Type:      TypeControlRequest,
				RequestID: "req_5_qrst7890",
				Request: map[string]any{
					"subtype":         SubtypeRewindFiles,
					"user_message_id": "uuid-12345-abcde",
				},
			},
			expected: `{"type":"control_request","request_id":"req_5_qrst7890","request":{"subtype":"rewind_files","user_message_id":"uuid-12345-abcde"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.request)
			assertTypesNoError(t, err)

			// Compare by unmarshaling both to normalize key order
			var got, want map[string]any
			if err := json.Unmarshal(data, &got); err != nil {
				t.Fatalf("Failed to unmarshal result: %v", err)
			}
			if err := json.Unmarshal([]byte(tt.expected), &want); err != nil {
				t.Fatalf("Failed to unmarshal expected: %v", err)
			}

			assertTypesJSONEqual(t, got, want)
		})
	}
}

// TestResponseDeserialization tests JSON deserialization of control responses.
func TestResponseDeserialization(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		validate func(*testing.T, Response)
	}{
		{
			name: "success_response",
			json: `{"type":"control_response","response":{"subtype":"success","request_id":"req_1_abcd1234","response":{"commands":["can_use_tool","mcp_message"]}}}`,
			validate: func(t *testing.T, resp Response) {
				t.Helper()
				if resp.Type != TypeControlResponse {
					t.Errorf("Expected type %q, got %q", TypeControlResponse, resp.Type)
				}
				if resp.Response.Subtype != SubtypeSuccess {
					t.Errorf("Expected subtype %q, got %q", SubtypeSuccess, resp.Response.Subtype)
				}
				if resp.Response.RequestID != "req_1_abcd1234" {
					t.Errorf("Expected request_id %q, got %q", "req_1_abcd1234", resp.Response.RequestID)
				}
				if resp.Response.Error != "" {
					t.Errorf("Expected empty error, got %q", resp.Response.Error)
				}
			},
		},
		{
			name: "error_response",
			json: `{"type":"control_response","response":{"subtype":"error","request_id":"req_2_efgh5678","error":"permission denied"}}`,
			validate: func(t *testing.T, resp Response) {
				t.Helper()
				if resp.Type != TypeControlResponse {
					t.Errorf("Expected type %q, got %q", TypeControlResponse, resp.Type)
				}
				if resp.Response.Subtype != SubtypeError {
					t.Errorf("Expected subtype %q, got %q", SubtypeError, resp.Response.Subtype)
				}
				if resp.Response.RequestID != "req_2_efgh5678" {
					t.Errorf("Expected request_id %q, got %q", "req_2_efgh5678", resp.Response.RequestID)
				}
				if resp.Response.Error != "permission denied" {
					t.Errorf("Expected error %q, got %q", "permission denied", resp.Response.Error)
				}
			},
		},
		{
			name: "initialize_response",
			json: `{"type":"control_response","response":{"subtype":"success","request_id":"req_3_init","response":{"commands":["can_use_tool","hook_callback","mcp_message"],"output_style":"stream"}}}`,
			validate: func(t *testing.T, resp Response) {
				t.Helper()
				if resp.Response.Response == nil {
					t.Fatal("Expected response data, got nil")
				}
				commands, ok := resp.Response.Response["commands"].([]any)
				if !ok {
					t.Fatalf("Expected commands array, got %T", resp.Response.Response["commands"])
				}
				if len(commands) != 3 {
					t.Errorf("Expected 3 commands, got %d", len(commands))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resp Response
			err := json.Unmarshal([]byte(tt.json), &resp)
			assertTypesNoError(t, err)
			tt.validate(t, resp)
		})
	}
}

// TestCanUseToolRequestDeserialization tests deserialization of incoming can_use_tool requests.
func TestCanUseToolRequestDeserialization(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		validate func(*testing.T, CanUseToolRequest)
	}{
		{
			name: "basic_request",
			json: `{"subtype":"can_use_tool","tool_name":"Bash","input":{"command":"ls -la"}}`,
			validate: func(t *testing.T, req CanUseToolRequest) {
				t.Helper()
				if req.Subtype != SubtypeCanUseTool {
					t.Errorf("Expected subtype %q, got %q", SubtypeCanUseTool, req.Subtype)
				}
				if req.ToolName != "Bash" {
					t.Errorf("Expected tool_name %q, got %q", "Bash", req.ToolName)
				}
				if req.Input["command"] != "ls -la" {
					t.Errorf("Expected command %q, got %v", "ls -la", req.Input["command"])
				}
			},
		},
		{
			name: "with_suggestions",
			json: `{"subtype":"can_use_tool","tool_name":"Write","input":{"file_path":"/etc/passwd"},"permission_suggestions":[{"type":"addRules","behavior":"deny"}],"blocked_path":"/etc/passwd"}`,
			validate: func(t *testing.T, req CanUseToolRequest) {
				t.Helper()
				if len(req.PermissionSuggestions) != 1 {
					t.Errorf("Expected 1 suggestion, got %d", len(req.PermissionSuggestions))
				}
				if req.BlockedPath == nil || *req.BlockedPath != "/etc/passwd" {
					t.Errorf("Expected blocked_path %q, got %v", "/etc/passwd", req.BlockedPath)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req CanUseToolRequest
			err := json.Unmarshal([]byte(tt.json), &req)
			assertTypesNoError(t, err)
			tt.validate(t, req)
		})
	}
}

// TestCanUseToolResponseSerialization tests serialization of can_use_tool responses.
func TestCanUseToolResponseSerialization(t *testing.T) {
	tests := []struct {
		name     string
		response CanUseToolResponse
		validate func(*testing.T, map[string]any)
	}{
		{
			name: "allow_response",
			response: CanUseToolResponse{
				Behavior: "allow",
			},
			validate: func(t *testing.T, data map[string]any) {
				t.Helper()
				if data["behavior"] != "allow" {
					t.Errorf("Expected behavior %q, got %v", "allow", data["behavior"])
				}
			},
		},
		{
			name: "allow_with_updated_input",
			response: CanUseToolResponse{
				Behavior:     "allow",
				UpdatedInput: map[string]any{"command": "ls -l"},
			},
			validate: func(t *testing.T, data map[string]any) {
				t.Helper()
				if data["behavior"] != "allow" {
					t.Errorf("Expected behavior %q, got %v", "allow", data["behavior"])
				}
				updatedInput, ok := data["updatedInput"].(map[string]any)
				if !ok {
					t.Fatalf("Expected updatedInput map, got %T", data["updatedInput"])
				}
				if updatedInput["command"] != "ls -l" {
					t.Errorf("Expected command %q, got %v", "ls -l", updatedInput["command"])
				}
			},
		},
		{
			name: "deny_response",
			response: CanUseToolResponse{
				Behavior: "deny",
				Message:  "Operation not permitted",
			},
			validate: func(t *testing.T, data map[string]any) {
				t.Helper()
				if data["behavior"] != "deny" {
					t.Errorf("Expected behavior %q, got %v", "deny", data["behavior"])
				}
				if data["message"] != "Operation not permitted" {
					t.Errorf("Expected message %q, got %v", "Operation not permitted", data["message"])
				}
			},
		},
		{
			name: "deny_with_interrupt",
			response: CanUseToolResponse{
				Behavior:  "deny",
				Message:   "Dangerous operation",
				Interrupt: true,
			},
			validate: func(t *testing.T, data map[string]any) {
				t.Helper()
				if data["behavior"] != "deny" {
					t.Errorf("Expected behavior %q, got %v", "deny", data["behavior"])
				}
				if data["interrupt"] != true {
					t.Errorf("Expected interrupt true, got %v", data["interrupt"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.response)
			assertTypesNoError(t, err)

			var result map[string]any
			if err := json.Unmarshal(data, &result); err != nil {
				t.Fatalf("Failed to unmarshal result: %v", err)
			}
			tt.validate(t, result)
		})
	}
}

// TestHookCallbackRequestDeserialization tests deserialization of hook callback requests.
func TestHookCallbackRequestDeserialization(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		validate func(*testing.T, HookCallbackRequest)
	}{
		{
			name: "basic_hook_callback",
			json: `{"subtype":"hook_callback","callback_id":"hook_0","input":{"tool_name":"Bash","tool_input":{"command":"rm -rf /"}}}`,
			validate: func(t *testing.T, req HookCallbackRequest) {
				t.Helper()
				if req.Subtype != SubtypeHookCallback {
					t.Errorf("Expected subtype %q, got %q", SubtypeHookCallback, req.Subtype)
				}
				if req.CallbackID != "hook_0" {
					t.Errorf("Expected callback_id %q, got %q", "hook_0", req.CallbackID)
				}
				if req.ToolUseID != nil {
					t.Errorf("Expected nil tool_use_id, got %v", req.ToolUseID)
				}
			},
		},
		{
			name: "hook_callback_with_tool_use_id",
			json: `{"subtype":"hook_callback","callback_id":"hook_1","input":{},"tool_use_id":"toolu_12345"}`,
			validate: func(t *testing.T, req HookCallbackRequest) {
				t.Helper()
				if req.ToolUseID == nil || *req.ToolUseID != "toolu_12345" {
					t.Errorf("Expected tool_use_id %q, got %v", "toolu_12345", req.ToolUseID)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req HookCallbackRequest
			err := json.Unmarshal([]byte(tt.json), &req)
			assertTypesNoError(t, err)
			tt.validate(t, req)
		})
	}
}

// TestHookMatcherSerialization tests serialization of hook matchers for initialization.
func TestHookMatcherSerialization(t *testing.T) {
	timeout := 5000
	matcher := HookMatcher{
		Matcher:         nil,
		HookCallbackIDs: []string{"hook_0", "hook_1"},
		Timeout:         &timeout,
	}

	data, err := json.Marshal(matcher)
	assertTypesNoError(t, err)

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	ids, ok := result["hookCallbackIds"].([]any)
	if !ok {
		t.Fatalf("Expected hookCallbackIds array, got %T", result["hookCallbackIds"])
	}
	if len(ids) != 2 {
		t.Errorf("Expected 2 callback IDs, got %d", len(ids))
	}
	if result["timeout"].(float64) != 5000 {
		t.Errorf("Expected timeout 5000, got %v", result["timeout"])
	}
}

// TestRequestIDFormat verifies the expected request ID format.
func TestRequestIDFormat(t *testing.T) {
	// Request IDs should match format: req_{counter}_{hex}
	validIDs := []string{
		"req_1_abcd1234",
		"req_42_deadbeef",
		"req_100_12345678",
	}

	for _, id := range validIDs {
		// Create a request with this ID and verify it serializes correctly
		req := Request{
			Type:      TypeControlRequest,
			RequestID: id,
			Request: map[string]any{
				"subtype": SubtypeInterrupt,
			},
		}

		data, err := json.Marshal(req)
		assertTypesNoError(t, err)

		var result map[string]any
		if err := json.Unmarshal(data, &result); err != nil {
			t.Fatalf("Failed to unmarshal: %v", err)
		}

		if result["request_id"] != id {
			t.Errorf("Expected request_id %q, got %v", id, result["request_id"])
		}
	}
}

// Helper functions

func assertTypesNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func assertTypesJSONEqual(t *testing.T, got, want map[string]any) {
	t.Helper()
	gotBytes, _ := json.Marshal(got)
	wantBytes, _ := json.Marshal(want)
	if string(gotBytes) != string(wantBytes) {
		t.Errorf("JSON mismatch:\ngot:  %s\nwant: %s", gotBytes, wantBytes)
	}
}
