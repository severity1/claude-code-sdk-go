package claudecode

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

// =============================================================================
// Test Functions (primary purpose)
// =============================================================================

// TestNewToolCreation tests NewTool constructor and McpTool accessors.
func TestNewToolCreation(t *testing.T) {
	tests := []struct {
		name        string
		toolName    string
		description string
		schema      map[string]any
	}{
		{
			name:        "basic_tool",
			toolName:    "add",
			description: "Add two numbers",
			schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"a": map[string]any{"type": "number"},
					"b": map[string]any{"type": "number"},
				},
			},
		},
		{
			name:        "empty_schema",
			toolName:    "noop",
			description: "Does nothing",
			schema:      map[string]any{},
		},
		{
			name:        "nil_schema",
			toolName:    "simple",
			description: "Simple tool",
			schema:      nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := func(_ context.Context, _ map[string]any) (*McpToolResult, error) {
				return &McpToolResult{Content: []McpContent{{Type: "text", Text: "ok"}}}, nil
			}

			tool := NewTool(tt.toolName, tt.description, tt.schema, handler)

			if tool.Name() != tt.toolName {
				t.Errorf("Name() = %q, want %q", tool.Name(), tt.toolName)
			}
			if tool.Description() != tt.description {
				t.Errorf("Description() = %q, want %q", tool.Description(), tt.description)
			}
			// Schema comparison is by reference equality for nil
			if tt.schema == nil && tool.InputSchema() != nil {
				t.Errorf("InputSchema() = %v, want nil", tool.InputSchema())
			}
			if tt.schema != nil && tool.InputSchema() == nil {
				t.Errorf("InputSchema() = nil, want %v", tt.schema)
			}
		})
	}
}

// TestToolHandlerExecution tests that tool handlers are called correctly.
func TestToolHandlerExecution(t *testing.T) {
	ctx, cancel := setupMcpTestContext(t, 5*time.Second)
	defer cancel()

	tests := []struct {
		name      string
		args      map[string]any
		wantText  string
		wantError bool
	}{
		{
			name:     "basic_execution",
			args:     map[string]any{"a": 1.0, "b": 2.0},
			wantText: "3.00",
		},
		{
			name:     "empty_args",
			args:     map[string]any{},
			wantText: "0.00",
		},
		{
			name:     "nil_args",
			args:     nil,
			wantText: "0.00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := func(_ context.Context, args map[string]any) (*McpToolResult, error) {
				a, _ := args["a"].(float64)
				b, _ := args["b"].(float64)
				return &McpToolResult{
					Content: []McpContent{{Type: "text", Text: formatFloat(a + b)}},
				}, nil
			}

			tool := NewTool("add", "Add numbers", nil, handler)
			result, err := tool.Call(ctx, tt.args)

			if tt.wantError && err == nil {
				t.Error("Expected error, got nil")
				return
			}
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			if tt.wantError {
				return
			}

			if len(result.Content) != 1 {
				t.Errorf("Expected 1 content item, got %d", len(result.Content))
				return
			}
			if result.Content[0].Text != tt.wantText {
				t.Errorf("Content text = %q, want %q", result.Content[0].Text, tt.wantText)
			}
		})
	}
}

// TestToolHandlerErrorPropagation tests that handler errors are properly returned.
func TestToolHandlerErrorPropagation(t *testing.T) {
	ctx, cancel := setupMcpTestContext(t, 5*time.Second)
	defer cancel()

	expectedErr := errors.New("handler error")
	handler := func(_ context.Context, _ map[string]any) (*McpToolResult, error) {
		return nil, expectedErr
	}

	tool := NewTool("failing", "Always fails", nil, handler)
	_, err := tool.Call(ctx, nil)

	if err == nil {
		t.Error("Expected error, got nil")
		return
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("Error = %v, want %v", err, expectedErr)
	}
}

// TestToolNilHandler tests that nil handlers return an error.
func TestToolNilHandler(t *testing.T) {
	ctx, cancel := setupMcpTestContext(t, 5*time.Second)
	defer cancel()

	tool := NewTool("nohandler", "No handler", nil, nil)
	_, err := tool.Call(ctx, nil)

	if err == nil {
		t.Error("Expected error for nil handler, got nil")
	}
}

// TestCreateSDKMcpServerWithTools tests server creation with tools.
func TestCreateSDKMcpServerWithTools(t *testing.T) {
	addTool := NewTool("add", "Add", nil, dummyHandler)
	subTool := NewTool("sub", "Subtract", nil, dummyHandler)

	server := CreateSDKMcpServer("calculator", "1.0.0", addTool, subTool)

	if server.Type != McpServerTypeSdk {
		t.Errorf("Type = %q, want %q", server.Type, McpServerTypeSdk)
	}
	if server.Name != "calculator" {
		t.Errorf("Name = %q, want %q", server.Name, "calculator")
	}
	if server.Instance == nil {
		t.Error("Instance is nil")
	}
}

// TestCreateSDKMcpServerEmpty tests server creation with no tools.
func TestCreateSDKMcpServerEmpty(t *testing.T) {
	server := CreateSDKMcpServer("empty", "1.0.0")

	if server.Type != McpServerTypeSdk {
		t.Errorf("Type = %q, want %q", server.Type, McpServerTypeSdk)
	}
	if server.Instance == nil {
		t.Error("Instance is nil")
	}
}

// TestCreateSDKMcpServerNilTools tests that nil tools are ignored.
func TestCreateSDKMcpServerNilTools(t *testing.T) {
	addTool := NewTool("add", "Add", nil, dummyHandler)

	server := CreateSDKMcpServer("test", "1.0.0", nil, addTool, nil)

	if server.Instance == nil {
		t.Fatal("Instance is nil")
	}

	ctx, cancel := setupMcpTestContext(t, 5*time.Second)
	defer cancel()

	tools, err := server.Instance.ListTools(ctx)
	if err != nil {
		t.Fatalf("ListTools error: %v", err)
	}
	if len(tools) != 1 {
		t.Errorf("Expected 1 tool, got %d", len(tools))
	}
}

// TestSdkMcpServerListTools tests the ListTools method.
func TestSdkMcpServerListTools(t *testing.T) {
	ctx, cancel := setupMcpTestContext(t, 5*time.Second)
	defer cancel()

	addTool := NewTool("add", "Add numbers", map[string]any{"type": "object"}, dummyHandler)
	mulTool := NewTool("mul", "Multiply numbers", map[string]any{"type": "object"}, dummyHandler)

	server := CreateSDKMcpServer("math", "1.0.0", addTool, mulTool)
	tools, err := server.Instance.ListTools(ctx)

	if err != nil {
		t.Fatalf("ListTools error: %v", err)
	}
	if len(tools) != 2 {
		t.Errorf("Expected 2 tools, got %d", len(tools))
	}

	// Check tools are in the list (order may vary due to map iteration)
	foundAdd, foundMul := false, false
	for _, tool := range tools {
		if tool.Name == "add" {
			foundAdd = true
		}
		if tool.Name == "mul" {
			foundMul = true
		}
	}
	if !foundAdd {
		t.Error("Tool 'add' not found in list")
	}
	if !foundMul {
		t.Error("Tool 'mul' not found in list")
	}
}

// TestSdkMcpServerCallTool tests tool execution through the server.
func TestSdkMcpServerCallTool(t *testing.T) {
	ctx, cancel := setupMcpTestContext(t, 5*time.Second)
	defer cancel()

	handler := func(_ context.Context, args map[string]any) (*McpToolResult, error) {
		a, _ := args["a"].(float64)
		b, _ := args["b"].(float64)
		return &McpToolResult{
			Content: []McpContent{{Type: "text", Text: formatFloat(a + b)}},
		}, nil
	}

	addTool := NewTool("add", "Add", nil, handler)
	server := CreateSDKMcpServer("calc", "1.0.0", addTool)

	result, err := server.Instance.CallTool(ctx, "add", map[string]any{"a": 10.0, "b": 5.0})
	if err != nil {
		t.Fatalf("CallTool error: %v", err)
	}
	if len(result.Content) != 1 {
		t.Fatalf("Expected 1 content item, got %d", len(result.Content))
	}
	if result.Content[0].Text != "15.00" {
		t.Errorf("Result = %q, want %q", result.Content[0].Text, "15.00")
	}
}

// TestSdkMcpServerCallToolNotFound tests error when tool doesn't exist.
func TestSdkMcpServerCallToolNotFound(t *testing.T) {
	ctx, cancel := setupMcpTestContext(t, 5*time.Second)
	defer cancel()

	server := CreateSDKMcpServer("empty", "1.0.0")
	_, err := server.Instance.CallTool(ctx, "nonexistent", nil)

	if err == nil {
		t.Error("Expected error for nonexistent tool, got nil")
	}
}

// TestSdkMcpServerIsErrorResult tests the IsError field in results.
func TestSdkMcpServerIsErrorResult(t *testing.T) {
	ctx, cancel := setupMcpTestContext(t, 5*time.Second)
	defer cancel()

	handler := func(_ context.Context, _ map[string]any) (*McpToolResult, error) {
		return &McpToolResult{
			Content: []McpContent{{Type: "text", Text: "Something went wrong"}},
			IsError: true,
		}, nil
	}

	errorTool := NewTool("error", "Returns error", nil, handler)
	server := CreateSDKMcpServer("test", "1.0.0", errorTool)

	result, err := server.Instance.CallTool(ctx, "error", nil)
	if err != nil {
		t.Fatalf("CallTool error: %v", err)
	}
	if !result.IsError {
		t.Error("Expected IsError = true")
	}
}

// TestSdkMcpServerImageContent tests image content in results.
func TestSdkMcpServerImageContent(t *testing.T) {
	ctx, cancel := setupMcpTestContext(t, 5*time.Second)
	defer cancel()

	handler := func(_ context.Context, _ map[string]any) (*McpToolResult, error) {
		return &McpToolResult{
			Content: []McpContent{
				{Type: "image", Data: "base64data", MimeType: "image/png"},
			},
		}, nil
	}

	imageTool := NewTool("image", "Returns image", nil, handler)
	server := CreateSDKMcpServer("test", "1.0.0", imageTool)

	result, err := server.Instance.CallTool(ctx, "image", nil)
	if err != nil {
		t.Fatalf("CallTool error: %v", err)
	}
	if len(result.Content) != 1 {
		t.Fatalf("Expected 1 content item, got %d", len(result.Content))
	}
	content := result.Content[0]
	if content.Type != "image" {
		t.Errorf("Type = %q, want %q", content.Type, "image")
	}
	if content.Data != "base64data" {
		t.Errorf("Data = %q, want %q", content.Data, "base64data")
	}
	if content.MimeType != "image/png" {
		t.Errorf("MimeType = %q, want %q", content.MimeType, "image/png")
	}
}

// TestSdkMcpServerConcurrentCalls tests thread safety with concurrent calls.
func TestSdkMcpServerConcurrentCalls(t *testing.T) {
	ctx, cancel := setupMcpTestContext(t, 10*time.Second)
	defer cancel()

	callCount := 0
	var mu sync.Mutex

	handler := func(_ context.Context, _ map[string]any) (*McpToolResult, error) {
		mu.Lock()
		callCount++
		mu.Unlock()
		return &McpToolResult{
			Content: []McpContent{{Type: "text", Text: "ok"}},
		}, nil
	}

	tool := NewTool("concurrent", "Concurrent test", nil, handler)
	server := CreateSDKMcpServer("test", "1.0.0", tool)

	const numGoroutines = 100
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			_, err := server.Instance.CallTool(ctx, "concurrent", nil)
			if err != nil {
				t.Errorf("Concurrent call error: %v", err)
			}
		}()
	}

	wg.Wait()

	mu.Lock()
	finalCount := callCount
	mu.Unlock()

	if finalCount != numGoroutines {
		t.Errorf("Call count = %d, want %d", finalCount, numGoroutines)
	}
}

// TestSdkMcpServerName tests the Name and Version methods.
func TestSdkMcpServerName(t *testing.T) {
	server := CreateSDKMcpServer("myserver", "2.5.0")

	if server.Instance.Name() != "myserver" {
		t.Errorf("Name() = %q, want %q", server.Instance.Name(), "myserver")
	}
	if server.Instance.Version() != "2.5.0" {
		t.Errorf("Version() = %q, want %q", server.Instance.Version(), "2.5.0")
	}
}

// =============================================================================
// Helper Functions (utilities)
// =============================================================================

// setupMcpTestContext creates a context with timeout for MCP tests.
func setupMcpTestContext(t *testing.T, timeout time.Duration) (context.Context, context.CancelFunc) {
	t.Helper()
	return context.WithTimeout(context.Background(), timeout)
}

// dummyHandler is a no-op handler for tests that don't need actual execution.
func dummyHandler(_ context.Context, _ map[string]any) (*McpToolResult, error) {
	return &McpToolResult{
		Content: []McpContent{{Type: "text", Text: "dummy"}},
	}, nil
}

// formatFloat formats a float with 2 decimal places.
func formatFloat(f float64) string {
	return formatFloatPrecision(f, 2)
}

// formatFloatPrecision formats a float with specified precision.
func formatFloatPrecision(f float64, precision int) string {
	format := "%." + string(rune('0'+precision)) + "f"
	return formatWithPattern(format, f)
}

// formatWithPattern formats using the specified pattern.
func formatWithPattern(pattern string, v any) string {
	switch val := v.(type) {
	case float64:
		return floatToString(val, pattern)
	default:
		return ""
	}
}

// floatToString converts float to string with pattern.
func floatToString(f float64, pattern string) string {
	// Simple implementation for tests
	if pattern == "%.2f" {
		intPart := int(f)
		fracPart := int((f - float64(intPart)) * 100)
		if fracPart < 0 {
			fracPart = -fracPart
		}
		return intToString(intPart) + "." + padLeft(intToString(fracPart), 2, '0')
	}
	return ""
}

// intToString converts int to string.
func intToString(i int) string {
	if i == 0 {
		return "0"
	}
	neg := i < 0
	if neg {
		i = -i
	}
	var digits []byte
	for i > 0 {
		digits = append([]byte{byte('0' + i%10)}, digits...)
		i /= 10
	}
	if neg {
		digits = append([]byte{'-'}, digits...)
	}
	return string(digits)
}

// padLeft pads string on the left with char to reach length.
func padLeft(s string, length int, char byte) string {
	for len(s) < length {
		s = string(char) + s
	}
	return s
}
