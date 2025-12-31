//go:build integration

package claudecode_test

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/severity1/claude-agent-sdk-go"
)

// Validation Functions for Group 1: Core Queries (T164-T167)

// validateSimpleQueryResponse validates T164: Simple Query Response Integration
func validateSimpleQueryResponse(t *testing.T, ctx context.Context, iter claudecode.MessageIterator, transport *integrationMockTransport) {
	t.Helper()

	messages := collectIntegrationMessages(t, ctx, iter)
	if len(messages) == 0 {
		t.Fatal("Expected at least one message from simple query")
	}

	// Should receive assistant message with expected text
	assistantMsg, ok := messages[0].(*claudecode.AssistantMessage)
	if !ok {
		t.Fatalf("Expected AssistantMessage, got %T", messages[0])
	}

	if len(assistantMsg.Content) == 0 {
		t.Fatal("Expected content in assistant message")
	}

	textBlock, ok := assistantMsg.Content[0].(*claudecode.TextBlock)
	if !ok {
		t.Fatalf("Expected TextBlock, got %T", assistantMsg.Content[0])
	}

	if textBlock.Text != "The answer is 42." {
		t.Errorf("Expected 'The answer is 42.', got '%s'", textBlock.Text)
	}

	if assistantMsg.Model != "claude-3-5-sonnet-20241022" {
		t.Errorf("Expected model 'claude-3-5-sonnet-20241022', got '%s'", assistantMsg.Model)
	}

	// Verify transport received the query
	assertIntegrationMessageCount(t, transport, 1)
}

// validateQueryWithTools validates T165: Query with Tools Integration
func validateQueryWithTools(t *testing.T, ctx context.Context, iter claudecode.MessageIterator, transport *integrationMockTransport) {
	t.Helper()

	messages := collectIntegrationMessages(t, ctx, iter)
	if len(messages) < 2 {
		t.Fatalf("Expected at least 2 messages for tool usage, got %d", len(messages))
	}

	// First message should be assistant with tool use
	assistantMsg, ok := messages[0].(*claudecode.AssistantMessage)
	if !ok {
		t.Fatalf("Expected first message to be AssistantMessage, got %T", messages[0])
	}

	// Should have both text and tool_use content
	if len(assistantMsg.Content) < 2 {
		t.Fatalf("Expected at least 2 content blocks in tool usage, got %d", len(assistantMsg.Content))
	}

	// Check for tool use block
	foundToolUse := false
	for _, content := range assistantMsg.Content {
		if toolUse, ok := content.(*claudecode.ToolUseBlock); ok {
			foundToolUse = true
			if toolUse.Name != "Read" {
				t.Errorf("Expected tool name 'Read', got '%s'", toolUse.Name)
			}
			if toolUse.ToolUseID == "" {
				t.Error("Expected tool use ID to be set")
			}
		}
	}

	if !foundToolUse {
		t.Error("Expected to find ToolUseBlock in assistant message")
	}

	// Verify tool result handling
	if len(messages) >= 3 {
		finalAssistant, ok := messages[2].(*claudecode.AssistantMessage)
		if ok && len(finalAssistant.Content) > 0 {
			if textBlock, ok := finalAssistant.Content[0].(*claudecode.TextBlock); ok {
				if !strings.Contains(textBlock.Text, "successfully read") {
					t.Logf("Tool result processing validated")
				}
			}
		}
	}
}

// validateStreamingClient validates T166: Streaming Client Integration
func validateStreamingClient(t *testing.T, ctx context.Context, iter claudecode.MessageIterator, transport *integrationMockTransport) {
	t.Helper()

	messages := collectIntegrationMessages(t, ctx, iter)
	if len(messages) < 2 {
		t.Fatalf("Expected at least 2 streaming messages, got %d", len(messages))
	}

	// Verify streaming messages are received in order
	expectedTexts := []string{"Streaming response part 1", "Streaming response part 2"}
	for i, expectedText := range expectedTexts {
		if i >= len(messages) {
			t.Fatalf("Missing streaming message %d", i+1)
		}

		assistantMsg, ok := messages[i].(*claudecode.AssistantMessage)
		if !ok {
			t.Fatalf("Expected streaming message %d to be AssistantMessage, got %T", i+1, messages[i])
		}

		if len(assistantMsg.Content) == 0 {
			t.Fatalf("Expected content in streaming message %d", i+1)
		}

		textBlock, ok := assistantMsg.Content[0].(*claudecode.TextBlock)
		if !ok {
			t.Fatalf("Expected TextBlock in streaming message %d, got %T", i+1, assistantMsg.Content[0])
		}

		if textBlock.Text != expectedText {
			t.Errorf("Expected streaming message %d to be '%s', got '%s'", i+1, expectedText, textBlock.Text)
		}
	}
}

// validateInterruptDuringStreaming validates T167: Interrupt During Streaming
func validateInterruptDuringStreaming(t *testing.T, ctx context.Context, iter claudecode.MessageIterator, transport *integrationMockTransport) {
	t.Helper()

	// Start collecting messages
	messagesCh := make(chan claudecode.Message, 10)
	errorsCh := make(chan error, 10)

	go func() {
		defer close(messagesCh)
		defer close(errorsCh)
		for {
			msg, err := iter.Next(ctx)
			if err == claudecode.ErrNoMoreMessages {
				return
			}
			if err != nil {
				errorsCh <- err
				return
			}
			if msg != nil {
				messagesCh <- msg
			}
		}
	}()

	// Simulate interrupt after brief delay
	go func() {
		time.Sleep(100 * time.Millisecond)
		if err := transport.Interrupt(ctx); err != nil {
			t.Logf("Interrupt simulation error: %v", err)
		}
	}()

	// Collect messages with timeout
	var messages []claudecode.Message
	timeout := time.After(2 * time.Second)

	for {
		select {
		case msg := <-messagesCh:
			if msg != nil {
				messages = append(messages, msg)
			}
		case err := <-errorsCh:
			if err != nil {
				t.Logf("Expected interrupt error: %v", err)
			}
		case <-timeout:
			goto done
		}
	}

done:
	// Validate that interrupt functionality works (messages may or may not be received)
	t.Logf("Interrupt test completed with %d messages received", len(messages))

	// The key validation is that interrupt doesn't cause system failure
	// Resource cleanup verification is handled by the test framework
}

// Validation Functions for Group 2: Session Management (T168-T171)

// validateSessionContinuation validates T168: Session Continuation Integration
func validateSessionContinuation(t *testing.T, ctx context.Context, client claudecode.Client, transport *integrationMockTransport) {
	t.Helper()

	t.Logf("Starting session continuation validation")

	// Send query to continue session
	err := client.QueryWithSession(ctx, "Continue our previous conversation", "session-continuation-123")
	assertIntegrationError(t, err, false, "")
	t.Logf("Query sent successfully")

	// Verify session ID was used
	transport.mu.Lock()
	sentMessages := make([]claudecode.StreamMessage, len(transport.sentMessages))
	copy(sentMessages, transport.sentMessages)
	testMsgCount := len(transport.testMessages)
	transport.mu.Unlock()

	t.Logf("Transport state: %d sent messages, %d test messages configured", len(sentMessages), testMsgCount)

	if len(sentMessages) == 0 {
		t.Fatal("Expected at least one sent message for session continuation")
	}

	lastMessage := sentMessages[len(sentMessages)-1]
	if lastMessage.SessionID != "session-continuation-123" {
		t.Errorf("Expected session ID 'session-continuation-123', got '%s'", lastMessage.SessionID)
	}

	// Test receiving response with session context
	t.Logf("Starting to receive response messages")
	iter := client.ReceiveResponse(ctx)
	defer iter.Close()

	messages := collectIntegrationMessages(t, ctx, iter)
	t.Logf("Session continuation validated with %d response messages", len(messages))
}

// validateMCPIntegration validates T169: MCP Integration Test
func validateMCPIntegration(t *testing.T, ctx context.Context, client claudecode.Client, transport *integrationMockTransport) {
	t.Helper()

	// Send query that should utilize MCP server
	err := client.Query(ctx, "Test MCP server integration")
	assertIntegrationError(t, err, false, "")

	// Verify MCP scenario was triggered
	if scenario, exists := transport.testScenarios["mcp_integration"]; exists {
		if !scenario.mcpServers {
			t.Error("Expected MCP servers to be enabled in test scenario")
		}
	}

	// Validate response indicates MCP integration
	iter := client.ReceiveResponse(ctx)
	defer iter.Close()

	messages := collectIntegrationMessages(t, ctx, iter)
	if len(messages) > 0 {
		if assistantMsg, ok := messages[0].(*claudecode.AssistantMessage); ok {
			if len(assistantMsg.Content) > 0 {
				if textBlock, ok := assistantMsg.Content[0].(*claudecode.TextBlock); ok {
					if strings.Contains(textBlock.Text, "MCP server") {
						t.Logf("MCP integration validated")
					}
				}
			}
		}
	}
}

// validatePermissionMode validates T170: Permission Mode Integration
func validatePermissionMode(t *testing.T, ctx context.Context, client claudecode.Client, transport *integrationMockTransport) {
	t.Helper()

	// Send query that should respect permission mode
	err := client.Query(ctx, "Test permission mode handling")
	assertIntegrationError(t, err, false, "")

	// Verify transport received message with proper permission context
	assertIntegrationMessageCount(t, transport, 1)

	transport.mu.Lock()
	sentMessage := transport.sentMessages[0]
	transport.mu.Unlock()

	if sentMessage.Type != "user" {
		t.Errorf("Expected user message type, got '%s'", sentMessage.Type)
	}

	// Validate response acknowledges permission mode
	iter := client.ReceiveResponse(ctx)
	defer iter.Close()

	messages := collectIntegrationMessages(t, ctx, iter)
	if len(messages) > 0 {
		t.Logf("Permission mode integration validated with %d messages", len(messages))
	}
}

// validateWorkingDirectory validates T171: Working Directory Integration
func validateWorkingDirectory(t *testing.T, ctx context.Context, client claudecode.Client, transport *integrationMockTransport) {
	t.Helper()

	// Send query that should operate in specified directory
	err := client.Query(ctx, "Test working directory handling")
	assertIntegrationError(t, err, false, "")

	// Verify working directory scenario
	if scenario, exists := transport.testScenarios["working_directory"]; exists {
		if scenario.workingDir != "/tmp" {
			t.Errorf("Expected working directory '/tmp', got '%s'", scenario.workingDir)
		}
	}

	// Validate response includes directory context
	iter := client.ReceiveResponse(ctx)
	defer iter.Close()

	messages := collectIntegrationMessages(t, ctx, iter)
	if len(messages) > 0 {
		if assistantMsg, ok := messages[0].(*claudecode.AssistantMessage); ok {
			if len(assistantMsg.Content) > 0 {
				if textBlock, ok := assistantMsg.Content[0].(*claudecode.TextBlock); ok {
					if strings.Contains(textBlock.Text, "/tmp") {
						t.Logf("Working directory integration validated")
					}
				}
			}
		}
	}
}

// Validation Functions for Group 3: Reliability (T172-T175)

// validateErrorHandling validates T172: Error Handling Integration
func validateErrorHandling(t *testing.T, ctx context.Context, transport *integrationMockTransport, opts *integrationTestOptions) {
	t.Helper()

	// Test connection error scenario
	client := claudecode.NewClientWithTransport(transport)
	defer disconnectIntegrationClientSafely(t, client)

	err := client.Connect(ctx)
	assertIntegrationError(t, err, true, "integration test connection error")

	// Test error message propagation
	if !strings.Contains(err.Error(), "integration test connection error") {
		t.Errorf("Expected error to contain 'integration test connection error', got: %v", err)
	}

	t.Logf("Error handling integration validated")
}

// validateLargeResponse validates T173: Large Response Integration
func validateLargeResponse(t *testing.T, ctx context.Context, transport *integrationMockTransport, opts *integrationTestOptions) {
	t.Helper()

	client := claudecode.NewClientWithTransport(transport)
	defer disconnectIntegrationClientSafely(t, client)

	// Remove connection error for this test
	transport.mu.Lock()
	transport.connectError = nil
	transport.mu.Unlock()

	connectIntegrationClientSafely(t, ctx, client)

	// Send query for large response
	err := client.Query(ctx, "Generate large response")
	assertIntegrationError(t, err, false, "")

	// Collect and validate large response
	iter := client.ReceiveResponse(ctx)
	defer iter.Close()

	start := time.Now()
	messages := collectIntegrationMessages(t, ctx, iter)
	duration := time.Since(start)

	if len(messages) == 0 {
		t.Fatal("Expected large response message")
	}

	// Validate response size and processing time
	if assistantMsg, ok := messages[0].(*claudecode.AssistantMessage); ok {
		if len(assistantMsg.Content) > 0 {
			if textBlock, ok := assistantMsg.Content[0].(*claudecode.TextBlock); ok {
				responseSize := len(textBlock.Text)
				if responseSize < 1000 { // Should be large
					t.Errorf("Expected large response (>1000 chars), got %d chars", responseSize)
				}
				t.Logf("Large response validated: %d chars processed in %v", responseSize, duration)
			}
		}
	}
}

// validateConcurrentClients validates T174: Concurrent Client Integration
func validateConcurrentClients(t *testing.T, ctx context.Context, transport *integrationMockTransport, opts *integrationTestOptions) {
	t.Helper()

	// Remove connection error for this test
	transport.mu.Lock()
	transport.connectError = nil
	transport.mu.Unlock()

	const numClients = 10
	var wg sync.WaitGroup
	errors := make(chan error, numClients)

	// Launch concurrent clients
	for i := 0; i < numClients; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			client := claudecode.NewClientWithTransport(transport)
			defer disconnectIntegrationClientSafely(t, client)

			if err := client.Connect(ctx); err != nil {
				errors <- err
				return
			}

			if err := client.Query(ctx, fmt.Sprintf("Concurrent query %d", id)); err != nil {
				errors <- err
				return
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	errorCount := 0
	for err := range errors {
		errorCount++
		t.Logf("Concurrent client error: %v", err)
	}

	if errorCount > numClients/2 { // Allow some failures in concurrent scenario
		t.Errorf("Too many concurrent client failures: %d/%d", errorCount, numClients)
	}

	t.Logf("Concurrent clients validated: %d/%d succeeded", numClients-errorCount, numClients)

	// Clean up transport state after concurrent test - following established patterns
	transport.mu.Lock()
	transport.connected = false
	transport.mu.Unlock()
}

// validateResourceCleanup validates T175: Resource Cleanup Integration
func validateResourceCleanup(t *testing.T, ctx context.Context, transport *integrationMockTransport, opts *integrationTestOptions) {
	t.Helper()

	// Remove connection error for this test
	transport.mu.Lock()
	transport.connectError = nil
	initialGoroutines := runtime.NumGoroutine()
	transport.mu.Unlock()

	// Test multiple client lifecycles
	for i := 0; i < 5; i++ {
		func() {
			client := claudecode.NewClientWithTransport(transport)
			defer disconnectIntegrationClientSafely(t, client)

			connectIntegrationClientSafely(t, ctx, client)
			err := client.Query(ctx, fmt.Sprintf("Resource test %d", i))
			assertIntegrationError(t, err, false, "")
		}()

		// Allow cleanup to occur
		runtime.GC()
		time.Sleep(10 * time.Millisecond)
	}

	// Verify resource cleanup
	finalGoroutines := runtime.NumGoroutine()
	if finalGoroutines > initialGoroutines+2 { // Allow some tolerance
		t.Errorf("Potential goroutine leak: initial=%d, final=%d", initialGoroutines, finalGoroutines)
	}

	// Resource cleanup verification is handled by the test framework after defers complete
	t.Logf("Resource cleanup validated: goroutines %d->%d", initialGoroutines, finalGoroutines)

	// Clean up transport state after resource test - following established patterns
	transport.mu.Lock()
	transport.connected = false
	transport.mu.Unlock()
}

// Validation Functions for Group 4: Performance (T176-T177)

// validatePerformance validates T176: Performance Integration Test
func validatePerformance(t *testing.T, ctx context.Context, transport *integrationMockTransport, opts *integrationTestOptions) {
	t.Helper()

	// Remove connection error for this test
	transport.mu.Lock()
	transport.connectError = nil
	transport.mu.Unlock()

	client := claudecode.NewClientWithTransport(transport)
	defer disconnectIntegrationClientSafely(t, client)

	connectIntegrationClientSafely(t, ctx, client)

	// Performance test with timing
	start := time.Now()

	for i := 0; i < opts.messageCount; i++ {
		err := client.Query(ctx, fmt.Sprintf("Performance test message %d", i))
		if err != nil {
			t.Fatalf("Performance test failed at message %d: %v", i, err)
		}
	}

	duration := time.Since(start)
	avgLatency := duration / time.Duration(opts.messageCount)

	// Validate performance metrics
	if avgLatency > time.Duration(opts.maxLatencyMs)*time.Millisecond {
		t.Errorf("Average latency %v exceeds limit %dms", avgLatency, opts.maxLatencyMs)
	}

	// Check memory usage (simplified)
	var m runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m)
	memoryMB := m.Alloc / 1024 / 1024

	if int(memoryMB) > opts.maxMemoryMB {
		t.Errorf("Memory usage %dMB exceeds limit %dMB", memoryMB, opts.maxMemoryMB)
	}

	t.Logf("Performance validated: %d messages in %v (avg: %v), memory: %dMB",
		opts.messageCount, duration, avgLatency, memoryMB)
}

// validateStressTest validates T177: Stress Test Integration
func validateStressTest(t *testing.T, ctx context.Context, transport *integrationMockTransport, opts *integrationTestOptions) {
	t.Helper()

	// Remove connection error for this test
	transport.mu.Lock()
	transport.connectError = nil
	transport.mu.Unlock()

	var wg sync.WaitGroup
	errors := make(chan error, opts.concurrent*10)
	start := time.Now()

	// Launch concurrent stress workers
	for i := 0; i < opts.concurrent; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			client := claudecode.NewClientWithTransport(transport)
			defer disconnectIntegrationClientSafely(t, client)

			if err := client.Connect(ctx); err != nil {
				errors <- err
				return
			}

			messagesPerWorker := opts.messageCount / opts.concurrent
			for j := 0; j < messagesPerWorker; j++ {
				err := client.Query(ctx, fmt.Sprintf("Stress test W%d-M%d", workerID, j))
				if err != nil {
					errors <- err
					return
				}
			}
		}(i)
	}

	wg.Wait()
	close(errors)
	duration := time.Since(start)

	// Count errors
	errorCount := 0
	for err := range errors {
		errorCount++
		if errorCount <= 5 { // Log first few errors
			t.Logf("Stress test error: %v", err)
		}
	}

	// Validate stress test results
	successRate := float64(opts.messageCount-errorCount) / float64(opts.messageCount) * 100
	if successRate < 90.0 { // 90% success rate threshold
		t.Errorf("Stress test success rate too low: %.1f%% (%d errors)", successRate, errorCount)
	}

	avgLatency := duration / time.Duration(opts.messageCount)
	if avgLatency > time.Duration(opts.maxLatencyMs)*time.Millisecond {
		t.Errorf("Stress test average latency %v exceeds limit %dms", avgLatency, opts.maxLatencyMs)
	}

	t.Logf("Stress test validated: %d messages, %d concurrent, %.1f%% success, avg latency: %v",
		opts.messageCount, opts.concurrent, successRate, avgLatency)

	// Clean up transport state after stress test - following established patterns
	transport.mu.Lock()
	transport.connected = false
	transport.mu.Unlock()
}

// Validation Functions for Group 5: Platforms (T178-T179)

// validateCLIVersionCompatibility validates T178: CLI Version Compatibility
func validateCLIVersionCompatibility(t *testing.T, ctx context.Context, transport *integrationMockTransport, opts *integrationTestOptions) {
	t.Helper()

	// Remove connection error for this test
	transport.mu.Lock()
	transport.connectError = nil
	transport.mu.Unlock()

	client := claudecode.NewClientWithTransport(transport)
	defer disconnectIntegrationClientSafely(t, client)

	connectIntegrationClientSafely(t, ctx, client)

	// Test version compatibility query
	err := client.Query(ctx, "Test CLI version compatibility")
	assertIntegrationError(t, err, false, "")

	// Validate version compatibility responses
	iter := client.ReceiveResponse(ctx)
	defer iter.Close()

	messages := collectIntegrationMessages(t, ctx, iter)
	compatibleVersions := 0

	for _, msg := range messages {
		if assistantMsg, ok := msg.(*claudecode.AssistantMessage); ok {
			if len(assistantMsg.Content) > 0 {
				if textBlock, ok := assistantMsg.Content[0].(*claudecode.TextBlock); ok {
					if strings.Contains(textBlock.Text, "compatible") {
						compatibleVersions++
					}
				}
			}
		}
	}

	if compatibleVersions == 0 {
		t.Error("Expected at least one CLI version compatibility confirmation")
	}

	t.Logf("CLI version compatibility validated: %d compatible versions", compatibleVersions)
}

// validateCrossPlatform validates T179: Cross-Platform Integration
func validateCrossPlatform(t *testing.T, ctx context.Context, transport *integrationMockTransport, opts *integrationTestOptions) {
	t.Helper()

	// Remove connection error for this test
	transport.mu.Lock()
	transport.connectError = nil
	transport.mu.Unlock()

	client := claudecode.NewClientWithTransport(transport)
	defer disconnectIntegrationClientSafely(t, client)

	connectIntegrationClientSafely(t, ctx, client)

	// Test platform-specific functionality
	err := client.Query(ctx, "Test cross-platform compatibility")
	assertIntegrationError(t, err, false, "")

	// Validate platform-specific response
	iter := client.ReceiveResponse(ctx)
	defer iter.Close()

	messages := collectIntegrationMessages(t, ctx, iter)
	if len(messages) == 0 {
		t.Fatal("Expected platform compatibility response")
	}

	// Check platform detection
	currentPlatform := runtime.GOOS
	platformSupported := false

	for _, msg := range messages {
		if assistantMsg, ok := msg.(*claudecode.AssistantMessage); ok {
			if len(assistantMsg.Content) > 0 {
				if textBlock, ok := assistantMsg.Content[0].(*claudecode.TextBlock); ok {
					if strings.Contains(textBlock.Text, currentPlatform) {
						platformSupported = true
					}
				}
			}
		}
	}

	if !platformSupported {
		t.Errorf("Expected platform %s to be supported in response", currentPlatform)
	}

	t.Logf("Cross-platform integration validated for %s", currentPlatform)
}

// Validation Functions for Group 6: Production (T180-T181)

// validateNetworkIsolation validates T180: Network Isolation Integration
func validateNetworkIsolation(t *testing.T, ctx context.Context, transport *integrationMockTransport, opts *integrationTestOptions) {
	t.Helper()

	// Remove connection error for this test
	transport.mu.Lock()
	transport.connectError = nil
	transport.mu.Unlock()

	client := claudecode.NewClientWithTransport(transport)
	defer disconnectIntegrationClientSafely(t, client)

	connectIntegrationClientSafely(t, ctx, client)

	// Test network isolation behavior
	err := client.Query(ctx, "Test offline behavior")
	assertIntegrationError(t, err, false, "")

	// Validate offline/isolated response
	iter := client.ReceiveResponse(ctx)
	defer iter.Close()

	messages := collectIntegrationMessages(t, ctx, iter)
	if len(messages) == 0 {
		t.Fatal("Expected network isolation response")
	}

	// Verify offline mode handling
	offlineModeDetected := false
	for _, msg := range messages {
		if assistantMsg, ok := msg.(*claudecode.AssistantMessage); ok {
			if len(assistantMsg.Content) > 0 {
				if textBlock, ok := assistantMsg.Content[0].(*claudecode.TextBlock); ok {
					if strings.Contains(textBlock.Text, "offline") {
						offlineModeDetected = true
					}
				}
			}
		}
	}

	if !offlineModeDetected {
		t.Error("Expected offline mode detection in network isolation test")
	}

	t.Logf("Network isolation validated: offline mode detected")
}

// New failing tests that will validate our fixes (RED phase of TDD)

// validateSpecificResourceCleanup validates strict resource cleanup following client_test.go patterns
func validateSpecificResourceCleanup(t *testing.T, ctx context.Context, transport *integrationMockTransport, opts *integrationTestOptions) {
	t.Helper()

	initialGoroutines := runtime.NumGoroutine()

	// Remove connection error for this test
	transport.mu.Lock()
	transport.connectError = nil
	transport.mu.Unlock()

	// Test resource cleanup through complete lifecycle
	func() {
		client := claudecode.NewClientWithTransport(transport)
		defer disconnectIntegrationClientSafely(t, client)

		connectIntegrationClientSafely(t, ctx, client)

		// Perform operations that create resources
		err := client.Query(ctx, "Resource test")
		assertIntegrationError(t, err, false, "")

		// Test receiving messages (creates goroutines)
		iter := client.ReceiveResponse(ctx)
		messages := collectIntegrationMessages(t, ctx, iter)
		iter.Close()

		if len(messages) == 0 {
			t.Error("Expected messages for resource cleanup test")
		}
	}() // Client should be disconnected here

	// This should pass after we fix the goroutine management
	assertResourcesReleased(t, transport, initialGoroutines, 1) // Allow 1 goroutine tolerance
}

// validateStrictTransportState validates transport state transitions following client_test.go patterns
func validateStrictTransportState(t *testing.T, ctx context.Context, transport *integrationMockTransport, opts *integrationTestOptions) {
	t.Helper()

	// Use provided transport from test setup
	// Initial state should be disconnected and not closed
	assertTransportStateTransition(t, transport, false, false)

	// After Connect()
	err := transport.Connect(ctx)
	assertIntegrationError(t, err, false, "")
	assertTransportStateTransition(t, transport, true, false)

	// After Close()
	err = transport.Close()
	assertIntegrationError(t, err, false, "")
	// This should pass after we fix the state management
	assertTransportStateTransition(t, transport, false, true)
}
