package control

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"
)

// TestQueryControlProtocol tests core Query functionality.
func TestQueryControlProtocol(t *testing.T) {
	tests := []struct {
		name string
		test func(context.Context, *testing.T)
	}{
		{"initialization_handshake", testInitializationHandshake},
		{"request_response_correlation", testRequestResponseCorrelation},
		{"concurrent_requests", testConcurrentRequests},
		{"timeout_handling", testTimeoutHandling},
		{"interrupt", testInterrupt},
		{"set_permission_mode", testSetPermissionMode},
		{"set_model", testSetModel},
		{"set_model_nil", testSetModelNil},
		{"rewind_files", testRewindFiles},
		{"can_use_tool_callback", testCanUseToolCallback},
		{"hook_callback", testHookCallback},
		{"error_response", testErrorResponse},
		{"closed_query", testClosedQuery},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := setupQueryTestContext(t, 10*time.Second)
			defer cancel()
			tt.test(ctx, t)
		})
	}
}

func testInitializationHandshake(ctx context.Context, t *testing.T) {
	t.Helper()

	transport := newQueryMockTransport()
	query := New(transport, WithInitTimeout(5*time.Second))
	transport.setQuery(query) // Enable response delivery

	if err := query.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer func() { _ = query.Close() }()

	// Set up mock response
	transport.setResponse(ResponsePayload{
		Subtype:   SubtypeSuccess,
		RequestID: "", // Will be matched by first request
		Response: map[string]any{
			"commands":     []any{"can_use_tool", "hook_callback", "mcp_message"},
			"output_style": "stream",
		},
	})

	// Perform initialization
	hooks := map[string][]HookMatcher{
		"PreToolUse": {
			{
				Matcher:         nil,
				HookCallbackIDs: []string{"hook_0"},
				Timeout:         intPtr(5000),
			},
		},
	}

	resp, err := query.Initialize(ctx, hooks)
	assertQueryNoError(t, err)

	if resp == nil {
		t.Fatal("Expected initialize response, got nil")
	}

	if len(resp.Commands) != 3 {
		t.Errorf("Expected 3 commands, got %d", len(resp.Commands))
	}

	if resp.OutputStyle != "stream" {
		t.Errorf("Expected output_style 'stream', got %q", resp.OutputStyle)
	}

	if !query.IsInitialized() {
		t.Error("Expected query to be initialized")
	}

	// Verify request format
	if len(transport.written) == 0 {
		t.Fatal("Expected at least one request written")
	}

	var req Request
	if err := json.Unmarshal(transport.written[0], &req); err != nil {
		t.Fatalf("Failed to parse request: %v", err)
	}

	if req.Type != TypeControlRequest {
		t.Errorf("Expected type %q, got %q", TypeControlRequest, req.Type)
	}

	if !strings.HasPrefix(req.RequestID, "req_") {
		t.Errorf("Expected request_id to start with 'req_', got %q", req.RequestID)
	}

	if req.Request["subtype"] != SubtypeInitialize {
		t.Errorf("Expected subtype %q, got %v", SubtypeInitialize, req.Request["subtype"])
	}
}

func testRequestResponseCorrelation(ctx context.Context, t *testing.T) {
	t.Helper()

	transport := newQueryMockTransport()
	query := New(transport)
	transport.setQuery(query) // Enable response delivery

	if err := query.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer func() { _ = query.Close() }()

	// Set up mock to echo request_id in response
	transport.setResponseFunc(func(reqID string) ResponsePayload {
		return ResponsePayload{
			Subtype:   SubtypeSuccess,
			RequestID: reqID,
			Response:  map[string]any{"echoed_id": reqID},
		}
	})

	// Send request
	request := map[string]any{
		"subtype": SubtypeInterrupt,
	}

	resp, err := query.SendRequest(ctx, request)
	assertQueryNoError(t, err)

	// Verify response contains echoed ID
	if resp["echoed_id"] == nil {
		t.Error("Expected echoed_id in response")
	}
}

func testConcurrentRequests(ctx context.Context, t *testing.T) {
	t.Helper()

	transport := newQueryMockTransport()
	query := New(transport)
	transport.setQuery(query) // Enable response delivery

	if err := query.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer func() { _ = query.Close() }()

	// Set up mock to echo request_id with small delay
	transport.setResponseFunc(func(reqID string) ResponsePayload {
		return ResponsePayload{
			Subtype:   SubtypeSuccess,
			RequestID: reqID,
			Response:  map[string]any{"id": reqID},
		}
	})
	transport.setDelay(10 * time.Millisecond)

	const numRequests = 5
	var wg sync.WaitGroup
	errors := make(chan error, numRequests)
	results := make(chan map[string]any, numRequests)

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			request := map[string]any{
				"subtype": SubtypeInterrupt,
				"n":       n,
			}
			resp, err := query.SendRequest(ctx, request)
			if err != nil {
				errors <- err
				return
			}
			results <- resp
		}(i)
	}

	wg.Wait()
	close(errors)
	close(results)

	// Check for errors
	for err := range errors {
		t.Errorf("Concurrent request error: %v", err)
	}

	// Verify all responses received
	count := 0
	for range results {
		count++
	}
	if count != numRequests {
		t.Errorf("Expected %d responses, got %d", numRequests, count)
	}
}

func testTimeoutHandling(ctx context.Context, t *testing.T) {
	t.Helper()

	transport := newQueryMockTransport()
	query := New(transport)

	if err := query.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer func() { _ = query.Close() }()

	// Set up mock to delay longer than timeout
	transport.setDelay(5 * time.Second)
	transport.setResponse(ResponsePayload{
		Subtype:   SubtypeSuccess,
		RequestID: "",
		Response:  map[string]any{},
	})

	// Create short timeout context
	timeoutCtx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	request := map[string]any{
		"subtype": SubtypeInterrupt,
	}

	_, err := query.SendRequest(timeoutCtx, request)
	if err == nil {
		t.Error("Expected timeout error")
	}

	if !strings.Contains(err.Error(), "timeout") && !strings.Contains(err.Error(), "deadline exceeded") {
		t.Errorf("Expected timeout error, got: %v", err)
	}
}

func testInterrupt(ctx context.Context, t *testing.T) {
	t.Helper()

	transport := newQueryMockTransport()
	query := New(transport)
	transport.setQuery(query) // Enable response delivery

	if err := query.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer func() { _ = query.Close() }()

	transport.setResponseFunc(func(reqID string) ResponsePayload {
		return ResponsePayload{
			Subtype:   SubtypeSuccess,
			RequestID: reqID,
			Response:  map[string]any{},
		}
	})

	err := query.Interrupt(ctx)
	assertQueryNoError(t, err)

	// Verify request
	if len(transport.written) == 0 {
		t.Fatal("Expected request written")
	}

	var req Request
	if err := json.Unmarshal(transport.written[0], &req); err != nil {
		t.Fatalf("Failed to parse request: %v", err)
	}

	if req.Request["subtype"] != SubtypeInterrupt {
		t.Errorf("Expected subtype %q, got %v", SubtypeInterrupt, req.Request["subtype"])
	}
}

func testSetPermissionMode(ctx context.Context, t *testing.T) {
	t.Helper()

	transport := newQueryMockTransport()
	query := New(transport)
	transport.setQuery(query) // Enable response delivery

	if err := query.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer func() { _ = query.Close() }()

	transport.setResponseFunc(func(reqID string) ResponsePayload {
		return ResponsePayload{
			Subtype:   SubtypeSuccess,
			RequestID: reqID,
			Response:  map[string]any{},
		}
	})

	err := query.SetPermissionMode(ctx, "bypassPermissions")
	assertQueryNoError(t, err)

	// Verify request
	var req Request
	if err := json.Unmarshal(transport.written[0], &req); err != nil {
		t.Fatalf("Failed to parse request: %v", err)
	}

	if req.Request["subtype"] != SubtypeSetPermissionMode {
		t.Errorf("Expected subtype %q, got %v", SubtypeSetPermissionMode, req.Request["subtype"])
	}

	if req.Request["mode"] != "bypassPermissions" {
		t.Errorf("Expected mode 'bypassPermissions', got %v", req.Request["mode"])
	}
}

func testSetModel(ctx context.Context, t *testing.T) {
	t.Helper()

	transport := newQueryMockTransport()
	query := New(transport)
	transport.setQuery(query) // Enable response delivery

	if err := query.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer func() { _ = query.Close() }()

	transport.setResponseFunc(func(reqID string) ResponsePayload {
		return ResponsePayload{
			Subtype:   SubtypeSuccess,
			RequestID: reqID,
			Response:  map[string]any{},
		}
	})

	model := "claude-sonnet-4-5"
	err := query.SetModel(ctx, &model)
	assertQueryNoError(t, err)

	// Verify request
	var req Request
	if err := json.Unmarshal(transport.written[0], &req); err != nil {
		t.Fatalf("Failed to parse request: %v", err)
	}

	if req.Request["subtype"] != SubtypeSetModel {
		t.Errorf("Expected subtype %q, got %v", SubtypeSetModel, req.Request["subtype"])
	}

	if req.Request["model"] != "claude-sonnet-4-5" {
		t.Errorf("Expected model 'claude-sonnet-4-5', got %v", req.Request["model"])
	}
}

func testSetModelNil(ctx context.Context, t *testing.T) {
	t.Helper()

	transport := newQueryMockTransport()
	query := New(transport)
	transport.setQuery(query) // Enable response delivery

	if err := query.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer func() { _ = query.Close() }()

	transport.setResponseFunc(func(reqID string) ResponsePayload {
		return ResponsePayload{
			Subtype:   SubtypeSuccess,
			RequestID: reqID,
			Response:  map[string]any{},
		}
	})

	err := query.SetModel(ctx, nil)
	assertQueryNoError(t, err)

	// Verify request - model should not be present
	var req Request
	if err := json.Unmarshal(transport.written[0], &req); err != nil {
		t.Fatalf("Failed to parse request: %v", err)
	}

	if _, exists := req.Request["model"]; exists {
		t.Error("Expected model to be absent for nil")
	}
}

func testRewindFiles(ctx context.Context, t *testing.T) {
	t.Helper()

	transport := newQueryMockTransport()
	query := New(transport)
	transport.setQuery(query) // Enable response delivery

	if err := query.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer func() { _ = query.Close() }()

	transport.setResponseFunc(func(reqID string) ResponsePayload {
		return ResponsePayload{
			Subtype:   SubtypeSuccess,
			RequestID: reqID,
			Response:  map[string]any{},
		}
	})

	err := query.RewindFiles(ctx, "uuid-12345-abcde")
	assertQueryNoError(t, err)

	// Verify request
	var req Request
	if err := json.Unmarshal(transport.written[0], &req); err != nil {
		t.Fatalf("Failed to parse request: %v", err)
	}

	if req.Request["subtype"] != SubtypeRewindFiles {
		t.Errorf("Expected subtype %q, got %v", SubtypeRewindFiles, req.Request["subtype"])
	}

	if req.Request["user_message_id"] != "uuid-12345-abcde" {
		t.Errorf("Expected user_message_id 'uuid-12345-abcde', got %v", req.Request["user_message_id"])
	}
}

func testCanUseToolCallback(ctx context.Context, t *testing.T) {
	t.Helper()

	transport := newQueryMockTransport()

	callbackCalled := false
	canUseTool := func(_ context.Context, req CanUseToolRequest) (CanUseToolResponse, error) {
		callbackCalled = true
		if req.ToolName != "Bash" {
			t.Errorf("Expected tool_name 'Bash', got %q", req.ToolName)
		}
		return CanUseToolResponse{
			Behavior: "allow",
		}, nil
	}

	query := New(transport, WithCanUseTool(canUseTool))

	if err := query.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer func() { _ = query.Close() }()

	// Simulate incoming can_use_tool request from CLI
	incomingRequest := map[string]any{
		"type":       TypeControlRequest,
		"request_id": "cli_req_1",
		"request": map[string]any{
			"subtype":   SubtypeCanUseTool,
			"tool_name": "Bash",
			"input":     map[string]any{"command": "ls"},
		},
	}

	err := query.HandleIncoming(ctx, incomingRequest)
	assertQueryNoError(t, err)

	if !callbackCalled {
		t.Error("Expected canUseTool callback to be called")
	}

	// Verify response was sent
	if len(transport.written) == 0 {
		t.Fatal("Expected response written")
	}

	var resp Response
	if err := json.Unmarshal(transport.written[0], &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp.Type != TypeControlResponse {
		t.Errorf("Expected type %q, got %q", TypeControlResponse, resp.Type)
	}

	if resp.Response.RequestID != "cli_req_1" {
		t.Errorf("Expected request_id 'cli_req_1', got %q", resp.Response.RequestID)
	}

	if resp.Response.Subtype != SubtypeSuccess {
		t.Errorf("Expected subtype %q, got %q", SubtypeSuccess, resp.Response.Subtype)
	}
}

func testHookCallback(ctx context.Context, t *testing.T) {
	t.Helper()

	transport := newQueryMockTransport()
	query := New(transport)

	if err := query.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer func() { _ = query.Close() }()

	// Register hook callback
	hookCalled := false
	query.RegisterHookCallback("hook_0", func(_ context.Context, _ any, _ *string) (map[string]any, error) {
		hookCalled = true
		return map[string]any{"continue": true}, nil
	})

	// Simulate incoming hook_callback request from CLI
	incomingRequest := map[string]any{
		"type":       TypeControlRequest,
		"request_id": "cli_hook_1",
		"request": map[string]any{
			"subtype":     SubtypeHookCallback,
			"callback_id": "hook_0",
			"input":       map[string]any{"tool_name": "Bash"},
		},
	}

	err := query.HandleIncoming(ctx, incomingRequest)
	assertQueryNoError(t, err)

	if !hookCalled {
		t.Error("Expected hook callback to be called")
	}

	// Verify response was sent
	if len(transport.written) == 0 {
		t.Fatal("Expected response written")
	}

	var resp Response
	if err := json.Unmarshal(transport.written[0], &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp.Response.Subtype != SubtypeSuccess {
		t.Errorf("Expected subtype %q, got %q", SubtypeSuccess, resp.Response.Subtype)
	}
}

func testErrorResponse(ctx context.Context, t *testing.T) {
	t.Helper()

	transport := newQueryMockTransport()
	query := New(transport)
	transport.setQuery(query) // Enable response delivery

	if err := query.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer func() { _ = query.Close() }()

	// Set up error response
	transport.setResponse(ResponsePayload{
		Subtype:   SubtypeError,
		RequestID: "",
		Error:     "permission denied",
	})

	request := map[string]any{
		"subtype": SubtypeInterrupt,
	}

	_, err := query.SendRequest(ctx, request)
	if err == nil {
		t.Error("Expected error from error response")
	}

	if !strings.Contains(err.Error(), "permission denied") {
		t.Errorf("Expected 'permission denied' in error, got: %v", err)
	}
}

func testClosedQuery(ctx context.Context, t *testing.T) {
	t.Helper()

	transport := newQueryMockTransport()
	query := New(transport)

	if err := query.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Close the query
	_ = query.Close()

	// Attempt to send request
	request := map[string]any{
		"subtype": SubtypeInterrupt,
	}

	_, err := query.SendRequest(ctx, request)
	if err == nil {
		t.Error("Expected error from closed query")
	}

	if !strings.Contains(err.Error(), "closed") {
		t.Errorf("Expected 'closed' in error, got: %v", err)
	}
}

// TestQueryConcurrency tests thread safety of Query operations.
func TestQueryConcurrency(t *testing.T) {
	ctx, cancel := setupQueryTestContext(t, 30*time.Second)
	defer cancel()

	transport := newQueryMockTransport()
	query := New(transport)
	transport.setQuery(query) // Enable response delivery

	if err := query.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer func() { _ = query.Close() }()

	transport.setResponseFunc(func(reqID string) ResponsePayload {
		return ResponsePayload{
			Subtype:   SubtypeSuccess,
			RequestID: reqID,
			Response:  map[string]any{},
		}
	})

	const numGoroutines = 10
	const operationsPerGoroutine = 5

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines*operationsPerGoroutine)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine; j++ {
				err := query.Interrupt(ctx)
				if err != nil {
					errors <- fmt.Errorf("goroutine %d operation %d: %w", id, j, err)
				}
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Error(err)
	}
}

// TestQueryGenerateCallbackID tests callback ID generation.
func TestQueryGenerateCallbackID(t *testing.T) {
	transport := newQueryMockTransport()
	query := New(transport)

	ids := make(map[string]bool)
	for i := 0; i < 10; i++ {
		id := query.GenerateCallbackID()
		if ids[id] {
			t.Errorf("Duplicate callback ID: %s", id)
		}
		ids[id] = true

		if !strings.HasPrefix(id, "hook_") {
			t.Errorf("Expected callback ID to start with 'hook_', got %q", id)
		}
	}
}

// Mock transport implementation

type queryMockTransport struct {
	mu           sync.Mutex
	written      [][]byte
	response     ResponsePayload
	responseFunc func(reqID string) ResponsePayload
	delay        time.Duration
	writeError   error
	query        *Query // Reference to query for delivering responses
}

func newQueryMockTransport() *queryMockTransport {
	return &queryMockTransport{
		written: make([][]byte, 0),
	}
}

func (m *queryMockTransport) setQuery(q *Query) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.query = q
}

func (m *queryMockTransport) Write(_ context.Context, data []byte) error {
	m.mu.Lock()
	writeError := m.writeError
	delay := m.delay
	response := m.response
	responseFunc := m.responseFunc
	query := m.query
	m.mu.Unlock()

	if writeError != nil {
		return writeError
	}

	// Strip trailing newline for easier parsing
	data = []byte(strings.TrimSuffix(string(data), "\n"))

	m.mu.Lock()
	m.written = append(m.written, data)
	m.mu.Unlock()

	// Parse request to get request_id
	var req Request
	if err := json.Unmarshal(data, &req); err != nil {
		return nil // Not a control request
	}

	// Simulate async response delivery
	go func() {
		if delay > 0 {
			time.Sleep(delay)
		}

		var payload ResponsePayload
		if responseFunc != nil {
			payload = responseFunc(req.RequestID)
		} else {
			payload = response
			payload.RequestID = req.RequestID
		}

		// Deliver response through Query's HandleIncoming
		if query != nil {
			respMsg := map[string]any{
				"type": TypeControlResponse,
				"response": map[string]any{
					"subtype":    payload.Subtype,
					"request_id": payload.RequestID,
					"response":   payload.Response,
					"error":      payload.Error,
				},
			}
			_ = query.HandleIncoming(context.Background(), respMsg)
		}
	}()

	return nil
}

func (m *queryMockTransport) setResponse(resp ResponsePayload) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.response = resp
	m.responseFunc = nil
}

func (m *queryMockTransport) setResponseFunc(fn func(reqID string) ResponsePayload) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.responseFunc = fn
}

func (m *queryMockTransport) setDelay(d time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.delay = d
}

// Helper functions

func setupQueryTestContext(t *testing.T, timeout time.Duration) (context.Context, context.CancelFunc) {
	t.Helper()
	return context.WithTimeout(context.Background(), timeout)
}

func assertQueryNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func intPtr(i int) *int {
	return &i
}
