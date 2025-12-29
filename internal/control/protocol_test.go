// Package control provides the SDK control protocol for bidirectional communication with Claude CLI.
package control

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"
)

// =============================================================================
// Phase 1: Control Message Type Tests
// =============================================================================

func TestControlMessageTypes(t *testing.T) {
	t.Run("message_type_constants", testMessageTypeConstants)
	t.Run("subtype_constants", testSubtypeConstants)
	t.Run("response_subtype_constants", testResponseSubtypeConstants)
}

func testMessageTypeConstants(t *testing.T) {
	t.Helper()

	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{"control_request", MessageTypeControlRequest, "control_request"},
		{"control_response", MessageTypeControlResponse, "control_response"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.constant != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, tc.constant)
			}
		})
	}
}

func testSubtypeConstants(t *testing.T) {
	t.Helper()

	// These constants must match the Python SDK exactly for parity
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{"interrupt", SubtypeInterrupt, "interrupt"},
		{"can_use_tool", SubtypeCanUseTool, "can_use_tool"},
		{"initialize", SubtypeInitialize, "initialize"},
		{"set_permission_mode", SubtypeSetPermissionMode, "set_permission_mode"},
		{"hook_callback", SubtypeHookCallback, "hook_callback"},
		{"mcp_message", SubtypeMcpMessage, "mcp_message"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.constant != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, tc.constant)
			}
		})
	}
}

func testResponseSubtypeConstants(t *testing.T) {
	t.Helper()

	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{"success", ResponseSubtypeSuccess, "success"},
		{"error", ResponseSubtypeError, "error"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.constant != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, tc.constant)
			}
		})
	}
}

func TestSDKControlRequestSerialization(t *testing.T) {
	t.Run("marshal_interrupt_request", testMarshalInterruptRequest)
	t.Run("marshal_initialize_request", testMarshalInitializeRequest)
	t.Run("unmarshal_control_request", testUnmarshalControlRequest)
}

func testMarshalInterruptRequest(t *testing.T) {
	t.Helper()

	req := SDKControlRequest{
		Type:      MessageTypeControlRequest,
		RequestID: "req_1_abc123",
		Request: InterruptRequest{
			Subtype: SubtypeInterrupt,
		},
	}

	data, err := json.Marshal(req)
	assertControlNoError(t, err)

	var parsed map[string]any
	err = json.Unmarshal(data, &parsed)
	assertControlNoError(t, err)

	assertControlEqual(t, "control_request", parsed["type"])
	assertControlEqual(t, "req_1_abc123", parsed["request_id"])

	request, ok := parsed["request"].(map[string]any)
	if !ok {
		t.Fatal("request field should be an object")
	}
	assertControlEqual(t, "interrupt", request["subtype"])
}

func testMarshalInitializeRequest(t *testing.T) {
	t.Helper()

	req := SDKControlRequest{
		Type:      MessageTypeControlRequest,
		RequestID: "req_2_def456",
		Request: InitializeRequest{
			Subtype: SubtypeInitialize,
		},
	}

	data, err := json.Marshal(req)
	assertControlNoError(t, err)

	var parsed map[string]any
	err = json.Unmarshal(data, &parsed)
	assertControlNoError(t, err)

	assertControlEqual(t, "control_request", parsed["type"])
	assertControlEqual(t, "req_2_def456", parsed["request_id"])

	request, ok := parsed["request"].(map[string]any)
	if !ok {
		t.Fatal("request field should be an object")
	}
	assertControlEqual(t, "initialize", request["subtype"])
}

func testUnmarshalControlRequest(t *testing.T) {
	t.Helper()

	jsonData := `{
		"type": "control_request",
		"request_id": "req_3_ghi789",
		"request": {
			"subtype": "interrupt"
		}
	}`

	var req SDKControlRequest
	err := json.Unmarshal([]byte(jsonData), &req)
	assertControlNoError(t, err)

	assertControlEqual(t, MessageTypeControlRequest, req.Type)
	assertControlEqual(t, "req_3_ghi789", req.RequestID)

	// Request is unmarshaled as map[string]any due to interface{} type
	request, ok := req.Request.(map[string]any)
	if !ok {
		t.Fatal("request field should unmarshal as map[string]any")
	}
	assertControlEqual(t, "interrupt", request["subtype"])
}

func TestSDKControlResponseSerialization(t *testing.T) {
	t.Run("marshal_success_response", testMarshalSuccessResponse)
	t.Run("marshal_error_response", testMarshalErrorResponse)
	t.Run("unmarshal_success_response", testUnmarshalSuccessResponse)
	t.Run("unmarshal_error_response", testUnmarshalErrorResponse)
}

func testMarshalSuccessResponse(t *testing.T) {
	t.Helper()

	resp := SDKControlResponse{
		Type: MessageTypeControlResponse,
		Response: Response{
			Subtype:   ResponseSubtypeSuccess,
			RequestID: "req_1_abc123",
			Response: map[string]any{
				"supported_commands": []string{"interrupt", "initialize"},
			},
		},
	}

	data, err := json.Marshal(resp)
	assertControlNoError(t, err)

	var parsed map[string]any
	err = json.Unmarshal(data, &parsed)
	assertControlNoError(t, err)

	assertControlEqual(t, "control_response", parsed["type"])

	response, ok := parsed["response"].(map[string]any)
	if !ok {
		t.Fatal("response field should be an object")
	}
	assertControlEqual(t, "success", response["subtype"])
	assertControlEqual(t, "req_1_abc123", response["request_id"])
}

func testMarshalErrorResponse(t *testing.T) {
	t.Helper()

	resp := SDKControlResponse{
		Type: MessageTypeControlResponse,
		Response: Response{
			Subtype:   ResponseSubtypeError,
			RequestID: "req_2_def456",
			Error:     "initialization failed: timeout",
		},
	}

	data, err := json.Marshal(resp)
	assertControlNoError(t, err)

	var parsed map[string]any
	err = json.Unmarshal(data, &parsed)
	assertControlNoError(t, err)

	response, ok := parsed["response"].(map[string]any)
	if !ok {
		t.Fatal("response field should be an object")
	}
	assertControlEqual(t, "error", response["subtype"])
	assertControlEqual(t, "initialization failed: timeout", response["error"])
}

func testUnmarshalSuccessResponse(t *testing.T) {
	t.Helper()

	jsonData := `{
		"type": "control_response",
		"response": {
			"subtype": "success",
			"request_id": "req_1_abc123",
			"response": {"status": "ok"}
		}
	}`

	var resp SDKControlResponse
	err := json.Unmarshal([]byte(jsonData), &resp)
	assertControlNoError(t, err)

	assertControlEqual(t, MessageTypeControlResponse, resp.Type)
	assertControlEqual(t, ResponseSubtypeSuccess, resp.Response.Subtype)
	assertControlEqual(t, "req_1_abc123", resp.Response.RequestID)
}

func testUnmarshalErrorResponse(t *testing.T) {
	t.Helper()

	jsonData := `{
		"type": "control_response",
		"response": {
			"subtype": "error",
			"request_id": "req_2_def456",
			"error": "unknown command"
		}
	}`

	var resp SDKControlResponse
	err := json.Unmarshal([]byte(jsonData), &resp)
	assertControlNoError(t, err)

	assertControlEqual(t, ResponseSubtypeError, resp.Response.Subtype)
	assertControlEqual(t, "unknown command", resp.Response.Error)
}

func TestInitializeRequestResponse(t *testing.T) {
	t.Run("initialize_request_structure", testInitializeRequestStructure)
	t.Run("initialize_response_structure", testInitializeResponseStructure)
}

func testInitializeRequestStructure(t *testing.T) {
	t.Helper()

	req := InitializeRequest{
		Subtype: SubtypeInitialize,
	}

	data, err := json.Marshal(req)
	assertControlNoError(t, err)

	var parsed map[string]any
	err = json.Unmarshal(data, &parsed)
	assertControlNoError(t, err)

	assertControlEqual(t, "initialize", parsed["subtype"])
}

func testInitializeResponseStructure(t *testing.T) {
	t.Helper()

	jsonData := `{
		"supported_commands": ["interrupt", "initialize", "set_permission_mode"]
	}`

	var resp InitializeResponse
	err := json.Unmarshal([]byte(jsonData), &resp)
	assertControlNoError(t, err)

	if len(resp.SupportedCommands) != 3 {
		t.Errorf("expected 3 supported commands, got %d", len(resp.SupportedCommands))
	}
	assertControlEqual(t, "interrupt", resp.SupportedCommands[0])
}

// =============================================================================
// Phase 2: Request/Response Correlation Tests
// =============================================================================

func TestRequestIDGeneration(t *testing.T) {
	t.Run("format_matches_python_sdk", testRequestIDFormat)
	t.Run("unique_ids", testRequestIDUniqueness)
	t.Run("counter_increments", testRequestIDCounterIncrements)
}

func testRequestIDFormat(t *testing.T) {
	t.Helper()

	transport := newControlMockTransport()
	protocol := NewProtocol(transport)

	id := protocol.generateRequestID()

	// Format: req_{counter}_{random_hex}
	// Example: req_1_a1b2c3d4
	if len(id) < 10 {
		t.Errorf("request ID too short: %s", id)
	}

	// Should start with "req_"
	if id[:4] != "req_" {
		t.Errorf("request ID should start with 'req_', got: %s", id)
	}
}

func testRequestIDUniqueness(t *testing.T) {
	t.Helper()

	transport := newControlMockTransport()
	protocol := NewProtocol(transport)

	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id := protocol.generateRequestID()
		if ids[id] {
			t.Errorf("duplicate request ID generated: %s", id)
		}
		ids[id] = true
	}
}

func testRequestIDCounterIncrements(t *testing.T) {
	t.Helper()

	transport := newControlMockTransport()
	protocol := NewProtocol(transport)

	id1 := protocol.generateRequestID()
	id2 := protocol.generateRequestID()
	id3 := protocol.generateRequestID()

	// IDs should be different
	if id1 == id2 || id2 == id3 || id1 == id3 {
		t.Errorf("request IDs should be unique: %s, %s, %s", id1, id2, id3)
	}
}

func TestRequestResponseCorrelation(t *testing.T) {
	t.Run("response_matched_to_request", testResponseMatchedToRequest)
	t.Run("unknown_request_id_ignored", testUnknownRequestIDIgnored)
}

func testResponseMatchedToRequest(t *testing.T) {
	t.Helper()

	ctx, cancel := setupControlTestContext(t, 5*time.Second)
	defer cancel()

	transport := newControlMockTransport()
	protocol := NewProtocol(transport)

	// Start protocol to process messages
	err := protocol.Start(ctx)
	assertControlNoError(t, err)
	defer func() { _ = protocol.Close() }()

	// Send a request and inject the response
	go func() {
		// Wait for the request to be sent
		time.Sleep(50 * time.Millisecond)

		// Get the request ID from the sent message
		transport.mu.Lock()
		if len(transport.writtenData) == 0 {
			transport.mu.Unlock()
			return
		}
		var req SDKControlRequest
		_ = json.Unmarshal(transport.writtenData[0], &req)
		transport.mu.Unlock()

		// Inject matching response
		transport.injectResponse(req.RequestID, map[string]any{"status": "ok"})
	}()

	// Send control request
	result, err := protocol.SendControlRequest(ctx, InterruptRequest{Subtype: SubtypeInterrupt}, 5*time.Second)
	assertControlNoError(t, err)

	// Verify response was received
	if result == nil {
		t.Fatal("expected response, got nil")
	}
}

func testUnknownRequestIDIgnored(t *testing.T) {
	t.Helper()

	ctx, cancel := setupControlTestContext(t, 2*time.Second)
	defer cancel()

	transport := newControlMockTransport()
	protocol := NewProtocol(transport)

	err := protocol.Start(ctx)
	assertControlNoError(t, err)
	defer func() { _ = protocol.Close() }()

	// Inject response with unknown request ID
	transport.injectResponse("req_unknown_12345678", map[string]any{"status": "ok"})

	// Give time for message to be processed
	time.Sleep(100 * time.Millisecond)

	// Protocol should still be running (no panic or error)
	if protocol.IsClosed() {
		t.Error("protocol should not be closed after unknown response")
	}
}

func TestRequestTimeout(t *testing.T) {
	t.Run("timeout_after_duration", testTimeoutAfterDuration)
}

func testTimeoutAfterDuration(t *testing.T) {
	t.Helper()

	ctx, cancel := setupControlTestContext(t, 5*time.Second)
	defer cancel()

	transport := newControlMockTransport()
	protocol := NewProtocol(transport)

	err := protocol.Start(ctx)
	assertControlNoError(t, err)
	defer func() { _ = protocol.Close() }()

	// Send request with short timeout (no response will be sent)
	start := time.Now()
	_, err = protocol.SendControlRequest(ctx, InterruptRequest{Subtype: SubtypeInterrupt}, 100*time.Millisecond)
	duration := time.Since(start)

	// Should get timeout error
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}

	// Duration should be approximately 100ms
	if duration < 90*time.Millisecond || duration > 200*time.Millisecond {
		t.Errorf("timeout duration should be ~100ms, got %v", duration)
	}
}

func TestConcurrentRequests(t *testing.T) {
	t.Run("thread_safe_concurrent_requests", testThreadSafeConcurrentRequests)
}

func testThreadSafeConcurrentRequests(t *testing.T) {
	t.Helper()

	ctx, cancel := setupControlTestContext(t, 10*time.Second)
	defer cancel()

	transport := newControlMockTransport()
	protocol := NewProtocol(transport)

	err := protocol.Start(ctx)
	assertControlNoError(t, err)
	defer func() { _ = protocol.Close() }()

	const numGoroutines = 10
	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines)

	// Auto-respond to all requests
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				transport.mu.Lock()
				for i, data := range transport.writtenData {
					if transport.responded[i] {
						continue
					}
					var req SDKControlRequest
					if err := json.Unmarshal(data, &req); err == nil {
						transport.responded[i] = true
						transport.mu.Unlock()
						transport.injectResponse(req.RequestID, map[string]any{"status": "ok"})
						transport.mu.Lock()
					}
				}
				transport.mu.Unlock()
				time.Sleep(10 * time.Millisecond)
			}
		}
	}()

	// Launch concurrent requests
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := protocol.SendControlRequest(ctx, InterruptRequest{Subtype: SubtypeInterrupt}, 5*time.Second)
			if err != nil {
				errors <- err
			}
		}()
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Errorf("concurrent request error: %v", err)
	}
}

// =============================================================================
// Phase 3: Initialize Handshake Tests
// =============================================================================

func TestInitializeHandshake(t *testing.T) {
	t.Run("success", testInitializeSuccess)
	t.Run("timeout", testInitializeTimeout)
	t.Run("error_response", testInitializeErrorResponse)
	t.Run("cached_result", testInitializeCachedResult)
}

func testInitializeSuccess(t *testing.T) {
	t.Helper()

	ctx, cancel := setupControlTestContext(t, 5*time.Second)
	defer cancel()

	transport := newControlMockTransport()
	protocol := NewProtocol(transport)

	err := protocol.Start(ctx)
	assertControlNoError(t, err)
	defer func() { _ = protocol.Close() }()

	// Auto-respond to initialize request
	go func() {
		time.Sleep(50 * time.Millisecond)
		transport.mu.Lock()
		if len(transport.writtenData) > 0 {
			var req SDKControlRequest
			if err := json.Unmarshal(transport.writtenData[0], &req); err == nil {
				transport.mu.Unlock()
				transport.injectResponse(req.RequestID, map[string]any{
					"supported_commands": []string{"interrupt", "initialize"},
				})
				return
			}
		}
		transport.mu.Unlock()
	}()

	resp, err := protocol.Initialize(ctx)
	assertControlNoError(t, err)

	if resp == nil {
		t.Fatal("expected initialize response, got nil")
	}
}

func testInitializeTimeout(t *testing.T) {
	t.Helper()

	ctx, cancel := setupControlTestContext(t, 5*time.Second)
	defer cancel()

	transport := newControlMockTransport()
	// Use short init timeout for test
	protocol := NewProtocol(transport, WithInitTimeout(100*time.Millisecond))

	err := protocol.Start(ctx)
	assertControlNoError(t, err)
	defer func() { _ = protocol.Close() }()

	// Don't send any response - should timeout
	_, err = protocol.Initialize(ctx)

	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
}

func testInitializeErrorResponse(t *testing.T) {
	t.Helper()

	ctx, cancel := setupControlTestContext(t, 5*time.Second)
	defer cancel()

	transport := newControlMockTransport()
	protocol := NewProtocol(transport)

	err := protocol.Start(ctx)
	assertControlNoError(t, err)
	defer func() { _ = protocol.Close() }()

	// Respond with error
	go func() {
		time.Sleep(50 * time.Millisecond)
		transport.mu.Lock()
		if len(transport.writtenData) > 0 {
			var req SDKControlRequest
			if err := json.Unmarshal(transport.writtenData[0], &req); err == nil {
				transport.mu.Unlock()
				transport.injectErrorResponse(req.RequestID, "initialization failed")
				return
			}
		}
		transport.mu.Unlock()
	}()

	_, err = protocol.Initialize(ctx)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func testInitializeCachedResult(t *testing.T) {
	t.Helper()

	ctx, cancel := setupControlTestContext(t, 5*time.Second)
	defer cancel()

	transport := newControlMockTransport()
	protocol := NewProtocol(transport)

	err := protocol.Start(ctx)
	assertControlNoError(t, err)
	defer func() { _ = protocol.Close() }()

	// Respond to first initialize
	go func() {
		time.Sleep(50 * time.Millisecond)
		transport.mu.Lock()
		if len(transport.writtenData) > 0 {
			var req SDKControlRequest
			if err := json.Unmarshal(transport.writtenData[0], &req); err == nil {
				transport.mu.Unlock()
				transport.injectResponse(req.RequestID, map[string]any{
					"supported_commands": []string{"interrupt"},
				})
				return
			}
		}
		transport.mu.Unlock()
	}()

	resp1, err := protocol.Initialize(ctx)
	assertControlNoError(t, err)

	// Second call should return cached result without sending request
	initialWriteCount := transport.getWriteCount()
	resp2, err := protocol.Initialize(ctx)
	assertControlNoError(t, err)

	// Should be same instance (cached)
	if resp1 != resp2 {
		t.Error("expected cached result, got different instance")
	}

	// Should not have sent another request
	if transport.getWriteCount() != initialWriteCount {
		t.Error("should not have sent another initialize request")
	}
}

// =============================================================================
// Phase 4: Message Routing Tests
// =============================================================================

func TestMessageRouting(t *testing.T) {
	t.Run("route_control_response", testRouteControlResponse)
	t.Run("route_regular_message", testRouteRegularMessage)
	t.Run("route_unknown_type", testRouteUnknownType)
}

func testRouteControlResponse(t *testing.T) {
	t.Helper()

	ctx, cancel := setupControlTestContext(t, 5*time.Second)
	defer cancel()

	transport := newControlMockTransport()
	protocol := NewProtocol(transport)

	err := protocol.Start(ctx)
	assertControlNoError(t, err)
	defer func() { _ = protocol.Close() }()

	// Set up pending request
	requestID := "req_test_12345678"
	responseChan := make(chan *Response, 1)
	protocol.setPendingRequest(requestID, responseChan)

	// Route a control response
	msg := map[string]any{
		"type": MessageTypeControlResponse,
		"response": map[string]any{
			"subtype":    ResponseSubtypeSuccess,
			"request_id": requestID,
			"response":   map[string]any{"status": "ok"},
		},
	}

	err = protocol.HandleIncomingMessage(ctx, msg)
	assertControlNoError(t, err)

	// Verify response was routed to channel
	select {
	case resp := <-responseChan:
		if resp == nil {
			t.Fatal("expected response, got nil")
		}
		assertControlEqual(t, ResponseSubtypeSuccess, resp.Subtype)
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for response")
	}
}

func testRouteRegularMessage(t *testing.T) {
	t.Helper()

	ctx, cancel := setupControlTestContext(t, 5*time.Second)
	defer cancel()

	transport := newControlMockTransport()
	protocol := NewProtocol(transport)

	err := protocol.Start(ctx)
	assertControlNoError(t, err)
	defer func() { _ = protocol.Close() }()

	// Route a regular message (user message)
	msg := map[string]any{
		"type": "user",
		"message": map[string]any{
			"content": "hello",
		},
	}

	err = protocol.HandleIncomingMessage(ctx, msg)
	assertControlNoError(t, err)

	// Verify message was forwarded to message stream
	select {
	case received := <-protocol.ReceiveMessages():
		if received["type"] != "user" {
			t.Errorf("expected user message, got %v", received["type"])
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for message")
	}
}

func testRouteUnknownType(t *testing.T) {
	t.Helper()

	ctx, cancel := setupControlTestContext(t, 5*time.Second)
	defer cancel()

	transport := newControlMockTransport()
	protocol := NewProtocol(transport)

	err := protocol.Start(ctx)
	assertControlNoError(t, err)
	defer func() { _ = protocol.Close() }()

	// Route an unknown message type
	msg := map[string]any{
		"type": "unknown_future_type",
		"data": "some data",
	}

	err = protocol.HandleIncomingMessage(ctx, msg)
	assertControlNoError(t, err)

	// Unknown messages should be forwarded to stream (forward compatibility)
	select {
	case received := <-protocol.ReceiveMessages():
		if received["type"] != "unknown_future_type" {
			t.Errorf("expected unknown_future_type, got %v", received["type"])
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for message")
	}
}

// =============================================================================
// Phase 6: Interrupt via Protocol Tests
// =============================================================================

func TestInterruptViaProtocol(t *testing.T) {
	t.Run("sends_interrupt_request", testInterruptSendsRequest)
}

func testInterruptSendsRequest(t *testing.T) {
	t.Helper()

	ctx, cancel := setupControlTestContext(t, 5*time.Second)
	defer cancel()

	transport := newControlMockTransport()
	protocol := NewProtocol(transport)

	err := protocol.Start(ctx)
	assertControlNoError(t, err)
	defer func() { _ = protocol.Close() }()

	// Auto-respond to interrupt request
	go func() {
		time.Sleep(50 * time.Millisecond)
		transport.mu.Lock()
		if len(transport.writtenData) > 0 {
			var req SDKControlRequest
			if err := json.Unmarshal(transport.writtenData[0], &req); err == nil {
				transport.mu.Unlock()
				transport.injectResponse(req.RequestID, nil)
				return
			}
		}
		transport.mu.Unlock()
	}()

	err = protocol.Interrupt(ctx)
	assertControlNoError(t, err)

	// Verify interrupt request was sent
	transport.mu.Lock()
	defer transport.mu.Unlock()

	if len(transport.writtenData) == 0 {
		t.Fatal("expected interrupt request to be sent")
	}

	var req SDKControlRequest
	err = json.Unmarshal(transport.writtenData[0], &req)
	assertControlNoError(t, err)

	// Verify it's an interrupt request
	request, ok := req.Request.(map[string]any)
	if !ok {
		t.Fatal("request should be a map")
	}
	assertControlEqual(t, SubtypeInterrupt, request["subtype"])
}

// =============================================================================
// Mock Transport for Control Protocol Tests
// =============================================================================

type controlMockTransport struct {
	mu          sync.Mutex
	writtenData [][]byte
	responded   []bool
	readChan    chan []byte
	writeErr    error
	closed      bool
}

func newControlMockTransport() *controlMockTransport {
	return &controlMockTransport{
		readChan:  make(chan []byte, 100),
		responded: make([]bool, 0),
	}
}

func (m *controlMockTransport) Write(_ context.Context, data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.writeErr != nil {
		return m.writeErr
	}

	// Make a copy of the data
	dataCopy := make([]byte, len(data))
	copy(dataCopy, data)
	m.writtenData = append(m.writtenData, dataCopy)
	m.responded = append(m.responded, false)
	return nil
}

func (m *controlMockTransport) Read(_ context.Context) <-chan []byte {
	return m.readChan
}

func (m *controlMockTransport) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.closed {
		m.closed = true
		close(m.readChan)
	}
	return nil
}

func (m *controlMockTransport) injectResponse(requestID string, response any) {
	resp := SDKControlResponse{
		Type: MessageTypeControlResponse,
		Response: Response{
			Subtype:   ResponseSubtypeSuccess,
			RequestID: requestID,
			Response:  response,
		},
	}
	data, _ := json.Marshal(resp)
	m.readChan <- data
}

func (m *controlMockTransport) injectErrorResponse(requestID string, errorMsg string) {
	resp := SDKControlResponse{
		Type: MessageTypeControlResponse,
		Response: Response{
			Subtype:   ResponseSubtypeError,
			RequestID: requestID,
			Error:     errorMsg,
		},
	}
	data, _ := json.Marshal(resp)
	m.readChan <- data
}

func (m *controlMockTransport) getWriteCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.writtenData)
}

// =============================================================================
// Test Helpers
// =============================================================================

func setupControlTestContext(t *testing.T, timeout time.Duration) (context.Context, context.CancelFunc) {
	t.Helper()
	return context.WithTimeout(context.Background(), timeout)
}

func assertControlNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func assertControlEqual(t *testing.T, expected, actual any) {
	t.Helper()
	if expected != actual {
		t.Errorf("expected %v, got %v", expected, actual)
	}
}
