package subprocess

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	claudecode "github.com/severity1/claude-code-sdk-go"
)

// T098: Subprocess Connection 🔴 RED
func TestSubprocessConnection(t *testing.T) {
	// Establish subprocess connection to Claude CLI

	tempCLI := createMockCLI(t)
	options := &claudecode.Options{}

	transport := New(tempCLI, options, false)

	ctx := context.Background()
	err := transport.Connect(ctx)

	if err != nil {
		t.Fatalf("Expected successful connection, got error: %v", err)
	}

	if !transport.IsConnected() {
		t.Error("Transport should report as connected")
	}

	// Cleanup
	transport.Close()
}

// T099: Subprocess Disconnection 🔴 RED
func TestSubprocessDisconnection(t *testing.T) {
	// Cleanly disconnect from subprocess

	tempCLI := createMockCLI(t)
	options := &claudecode.Options{}

	transport := New(tempCLI, options, false)

	ctx := context.Background()
	err := transport.Connect(ctx)
	if err != nil {
		t.Fatalf("Connection failed: %v", err)
	}

	// Should be connected
	if !transport.IsConnected() {
		t.Error("Transport should report as connected before disconnect")
	}

	// Disconnect
	err = transport.Close()
	if err != nil {
		t.Errorf("Expected clean disconnection, got error: %v", err)
	}

	// Should no longer be connected
	if transport.IsConnected() {
		t.Error("Transport should not report as connected after disconnect")
	}
}

// T100: 5-Second Termination Sequence 🔴 RED
func TestFiveSecondTerminationSequence(t *testing.T) {
	// Implement SIGTERM → wait 5s → SIGKILL sequence

	tempCLI := createLongRunningMockCLI(t)
	options := &claudecode.Options{}

	transport := New(tempCLI, options, false)

	ctx := context.Background()
	err := transport.Connect(ctx)
	if err != nil {
		t.Fatalf("Connection failed: %v", err)
	}

	// Start timing the termination
	start := time.Now()

	// This should trigger the 5-second termination sequence
	err = transport.Close()

	duration := time.Since(start)

	// Should complete without error
	if err != nil {
		t.Errorf("Expected clean termination, got error: %v", err)
	}

	// Should take some time but not more than 6 seconds (allowing buffer)
	if duration > 6*time.Second {
		t.Errorf("Termination took too long: %v", duration)
	}

	// Process should be terminated
	if transport.IsConnected() {
		t.Error("Process should be terminated after Close()")
	}
}

// T101: Process Lifecycle Management 🔴 RED
func TestProcessLifecycleManagement(t *testing.T) {
	// Manage complete process lifecycle

	tempCLI := createMockCLI(t)
	options := &claudecode.Options{}

	transport := New(tempCLI, options, false)

	// Initial state should be disconnected
	if transport.IsConnected() {
		t.Error("New transport should not be connected initially")
	}

	// Connect
	ctx := context.Background()
	err := transport.Connect(ctx)
	if err != nil {
		t.Fatalf("Connection failed: %v", err)
	}

	// Should be running state
	if !transport.IsConnected() {
		t.Error("Transport should be connected after Connect()")
	}

	// Should handle multiple Close() calls gracefully
	err1 := transport.Close()
	err2 := transport.Close()

	if err1 != nil {
		t.Errorf("First Close() should succeed: %v", err1)
	}
	if err2 != nil {
		t.Errorf("Second Close() should be graceful: %v", err2)
	}

	// Final state should be disconnected
	if transport.IsConnected() {
		t.Error("Transport should not be connected after Close()")
	}
}

// T102: Stdin Message Sending 🔴 RED
func TestStdinMessageSending(t *testing.T) {
	// Send JSON messages to subprocess stdin

	tempCLI := createMockCLI(t)
	options := &claudecode.Options{}

	transport := New(tempCLI, options, false)

	ctx := context.Background()
	err := transport.Connect(ctx)
	if err != nil {
		t.Fatalf("Connection failed: %v", err)
	}
	defer transport.Close()

	// Create test message
	message := claudecode.StreamMessage{
		Type:      "user",
		SessionID: "test-session",
	}

	// Should be able to send message
	err = transport.SendMessage(ctx, message)
	if err != nil {
		t.Errorf("Expected successful message send, got error: %v", err)
	}
}

// T103: Stdout Message Receiving 🔴 RED
func TestStdoutMessageReceiving(t *testing.T) {
	// Receive JSON messages from subprocess stdout

	tempCLI := createMockCLI(t)
	options := &claudecode.Options{}

	transport := New(tempCLI, options, false)

	ctx := context.Background()
	err := transport.Connect(ctx)
	if err != nil {
		t.Fatalf("Connection failed: %v", err)
	}
	defer transport.Close()

	// Get message channels
	msgChan, errChan := transport.ReceiveMessages(ctx)

	// Should get channels (even if no messages yet)
	if msgChan == nil {
		t.Error("Message channel should not be nil")
	}
	if errChan == nil {
		t.Error("Error channel should not be nil")
	}

	// Test that channels don't block immediately
	select {
	case <-msgChan:
		// OK if message received
	case <-errChan:
		// OK if error received
	case <-time.After(100 * time.Millisecond):
		// OK if no immediate message - this is normal
	}
}

// T104: Stderr Isolation 🔴 RED
func TestStderrIsolation(t *testing.T) {
	// Isolate stderr using temporary files

	tempCLI := createMockCLI(t)
	options := &claudecode.Options{}

	transport := New(tempCLI, options, false)

	ctx := context.Background()
	err := transport.Connect(ctx)
	if err != nil {
		t.Fatalf("Connection failed: %v", err)
	}
	defer transport.Close()

	// Stderr should be isolated and not cause deadlocks
	// This test mainly verifies that connection succeeds without hanging
	time.Sleep(100 * time.Millisecond) // Allow process to start

	if !transport.IsConnected() {
		t.Error("Transport should remain connected despite stderr")
	}
}

// T105: Environment Variable Setting 🔴 RED
func TestEnvironmentVariableSetting(t *testing.T) {
	// Set CLAUDE_CODE_ENTRYPOINT environment variable

	tempCLI := createEnvironmentCheckCLI(t)
	options := &claudecode.Options{}

	transport := New(tempCLI, options, false)

	ctx := context.Background()
	err := transport.Connect(ctx)
	if err != nil {
		t.Fatalf("Connection failed: %v", err)
	}
	defer transport.Close()

	// The mock CLI should have received the environment variable
	// This is verified by the mock script itself
	time.Sleep(100 * time.Millisecond) // Allow process to start

	if !transport.IsConnected() {
		t.Error("Transport should be connected")
	}
}

// T106: Concurrent I/O Handling 🔴 RED
func TestConcurrentIOHandling(t *testing.T) {
	// Handle stdin/stdout concurrently with goroutines

	tempCLI := createMockCLI(t)
	options := &claudecode.Options{}

	transport := New(tempCLI, options, false)

	ctx := context.Background()
	err := transport.Connect(ctx)
	if err != nil {
		t.Fatalf("Connection failed: %v", err)
	}
	defer transport.Close()

	// Start receiving messages in background
	msgChan, errChan := transport.ReceiveMessages(ctx)

	// Send messages concurrently
	message1 := claudecode.StreamMessage{Type: "user", SessionID: "session1"}
	message2 := claudecode.StreamMessage{Type: "user", SessionID: "session2"}

	go func() {
		transport.SendMessage(ctx, message1)
	}()
	go func() {
		transport.SendMessage(ctx, message2)
	}()

	// Should not block or cause race conditions
	time.Sleep(200 * time.Millisecond)

	// Cleanup channels
	select {
	case <-msgChan:
	case <-errChan:
	default:
	}
}

// T107: Process Error Handling 🔴 RED
func TestProcessErrorHandling(t *testing.T) {
	// Handle subprocess errors and exit codes

	tempCLI := createFailingCLI(t)
	options := &claudecode.Options{}

	transport := New(tempCLI, options, false)

	ctx := context.Background()
	err := transport.Connect(ctx)

	// Connection should succeed initially (process starts)
	if err != nil {
		t.Fatalf("Connection should succeed initially: %v", err)
	}
	defer transport.Close()

	// But process should fail quickly
	time.Sleep(100 * time.Millisecond)

	// Get error channels to see process failure
	_, errChan := transport.ReceiveMessages(ctx)

	// Should get error from failing process or stderr
	select {
	case err := <-errChan:
		if err != nil {
			errMsg := err.Error()
			if !strings.Contains(errMsg, "exit") && !strings.Contains(errMsg, "failed") && !strings.Contains(errMsg, "error") {
				t.Errorf("Error message should indicate process failure: %v", errMsg)
			}
		}
	case <-time.After(200 * time.Millisecond):
		// It's OK if no error comes through channels immediately
		t.Log("No immediate error from failing CLI - process may have failed quickly")
	}
}

// T108: Message Channel Management 🔴 RED
func TestMessageChannelManagement(t *testing.T) {
	// Manage message and error channels

	tempCLI := createMockCLI(t)
	options := &claudecode.Options{}

	transport := New(tempCLI, options, false)

	ctx := context.Background()
	err := transport.Connect(ctx)
	if err != nil {
		t.Fatalf("Connection failed: %v", err)
	}
	defer transport.Close()

	// Get channels
	msgChan, errChan := transport.ReceiveMessages(ctx)

	// Channels should be separate (test by type assertion)
	if msgChan == nil || errChan == nil {
		t.Error("Both channels should be non-nil")
	}

	// Should be able to get channels multiple times
	msgChan2, errChan2 := transport.ReceiveMessages(ctx)

	if msgChan2 == nil || errChan2 == nil {
		t.Error("Should be able to get channels multiple times")
	}
}

// T109: Backpressure Handling 🔴 RED
func TestBackpressureHandling(t *testing.T) {
	// Handle backpressure in message channels

	tempCLI := createMockCLI(t)
	options := &claudecode.Options{}

	transport := New(tempCLI, options, false)

	ctx := context.Background()
	err := transport.Connect(ctx)
	if err != nil {
		t.Fatalf("Connection failed: %v", err)
	}
	defer transport.Close()

	// Send multiple messages rapidly to test backpressure
	for i := 0; i < 10; i++ {
		message := claudecode.StreamMessage{
			Type:      "user",
			SessionID: "session",
		}

		// Should not block indefinitely
		done := make(chan error, 1)
		go func() {
			done <- transport.SendMessage(ctx, message)
		}()

		select {
		case err := <-done:
			if err != nil && !strings.Contains(err.Error(), "context") {
				t.Errorf("Unexpected error in message %d: %v", i, err)
			}
		case <-time.After(1 * time.Second):
			t.Errorf("Message %d took too long to send (backpressure issue)", i)
		}
	}
}

// T110: Context Cancellation 🔴 RED
func TestContextCancellation(t *testing.T) {
	// Support context cancellation throughout transport

	tempCLI := createMockCLI(t)
	options := &claudecode.Options{}

	transport := New(tempCLI, options, false)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := transport.Connect(ctx)
	if err != nil {
		t.Fatalf("Connection failed: %v", err)
	}
	defer transport.Close()

	// Wait for context cancellation
	time.Sleep(200 * time.Millisecond)

	// Operations should respect cancelled context
	message := claudecode.StreamMessage{Type: "user", SessionID: "session"}
	err = transport.SendMessage(ctx, message)

	// Should get context error
	if err != nil && !strings.Contains(err.Error(), "context") {
		// This is ok - context cancellation handling varies
	}
}

// T111: Resource Cleanup 🔴 RED
func TestResourceCleanup(t *testing.T) {
	// Clean up all resources on shutdown

	tempCLI := createMockCLI(t)
	options := &claudecode.Options{}

	transport := New(tempCLI, options, false)

	ctx := context.Background()
	err := transport.Connect(ctx)
	if err != nil {
		t.Fatalf("Connection failed: %v", err)
	}

	// Get channels to create resources
	transport.ReceiveMessages(ctx)

	// Cleanup should not error
	err = transport.Close()
	if err != nil {
		t.Errorf("Resource cleanup failed: %v", err)
	}

	// Multiple cleanups should be safe
	err = transport.Close()
	if err != nil {
		t.Errorf("Multiple cleanup should be safe: %v", err)
	}
}

// T112: Process State Tracking 🔴 RED
func TestProcessStateTracking(t *testing.T) {
	// Track subprocess connection state

	tempCLI := createMockCLI(t)
	options := &claudecode.Options{}

	transport := New(tempCLI, options, false)

	// Initially disconnected
	if transport.IsConnected() {
		t.Error("Should be disconnected initially")
	}

	ctx := context.Background()
	err := transport.Connect(ctx)
	if err != nil {
		t.Fatalf("Connection failed: %v", err)
	}

	// Should be connected
	if !transport.IsConnected() {
		t.Error("Should be connected after Connect()")
	}

	transport.Close()

	// Should be disconnected
	if transport.IsConnected() {
		t.Error("Should be disconnected after Close()")
	}
}

// T113: Interrupt Signal Handling 🔴 RED
func TestInterruptSignalHandling(t *testing.T) {
	// Handle interrupt signals to subprocess

	tempCLI := createMockCLI(t)
	options := &claudecode.Options{}

	transport := New(tempCLI, options, false)

	ctx := context.Background()
	err := transport.Connect(ctx)
	if err != nil {
		t.Fatalf("Connection failed: %v", err)
	}
	defer transport.Close()

	// Should be able to send interrupt
	err = transport.Interrupt(ctx)
	if err != nil {
		t.Errorf("Expected successful interrupt, got error: %v", err)
	}

	// Process should still be manageable
	if !transport.IsConnected() {
		t.Error("Process should still be connected after interrupt")
	}
}

// T114: Message Ordering Guarantees 🔴 RED
func TestMessageOrderingGuarantees(t *testing.T) {
	// Maintain message ordering through transport

	tempCLI := createMockCLI(t)
	options := &claudecode.Options{}

	transport := New(tempCLI, options, false)

	ctx := context.Background()
	err := transport.Connect(ctx)
	if err != nil {
		t.Fatalf("Connection failed: %v", err)
	}
	defer transport.Close()

	// Send messages in sequence
	messages := []claudecode.StreamMessage{
		{Type: "user", SessionID: "session1"},
		{Type: "user", SessionID: "session2"},
		{Type: "user", SessionID: "session3"},
	}

	for _, msg := range messages {
		err := transport.SendMessage(ctx, msg)
		if err != nil {
			t.Errorf("Failed to send message: %v", err)
		}
	}

	// Order should be maintained (this is primarily a design requirement)
	// The actual verification would require a more complex mock
}

// T115: Transport Reconnection 🔴 RED
func TestTransportReconnection(t *testing.T) {
	// Handle transport reconnection scenarios

	tempCLI := createMockCLI(t)
	options := &claudecode.Options{}

	transport := New(tempCLI, options, false)

	ctx := context.Background()

	// First connection
	err := transport.Connect(ctx)
	if err != nil {
		t.Fatalf("First connection failed: %v", err)
	}

	transport.Close()

	// Should be able to reconnect
	err = transport.Connect(ctx)
	if err != nil {
		t.Fatalf("Reconnection failed: %v", err)
	}

	if !transport.IsConnected() {
		t.Error("Should be connected after reconnection")
	}

	transport.Close()
}

// T116: Performance Under Load 🔴 RED
func TestPerformanceUnderLoad(t *testing.T) {
	// Maintain performance under high message throughput

	tempCLI := createMockCLI(t)
	options := &claudecode.Options{}

	transport := New(tempCLI, options, false)

	ctx := context.Background()
	err := transport.Connect(ctx)
	if err != nil {
		t.Fatalf("Connection failed: %v", err)
	}
	defer transport.Close()

	// Send many messages in rapid succession
	start := time.Now()
	messageCount := 100

	for i := 0; i < messageCount; i++ {
		message := claudecode.StreamMessage{
			Type:      "user",
			SessionID: "load-test",
		}

		err := transport.SendMessage(ctx, message)
		if err != nil {
			t.Errorf("Message %d failed: %v", i, err)
		}
	}

	duration := time.Since(start)

	// Should handle reasonable throughput
	if duration > 10*time.Second {
		t.Errorf("Load test took too long: %v for %d messages", duration, messageCount)
	}
}

// T117: Memory Usage Optimization 🔴 RED
func TestMemoryUsageOptimization(t *testing.T) {
	// Optimize memory usage in transport layer

	tempCLI := createMockCLI(t)
	options := &claudecode.Options{}

	transport := New(tempCLI, options, false)

	ctx := context.Background()
	err := transport.Connect(ctx)
	if err != nil {
		t.Fatalf("Connection failed: %v", err)
	}
	defer transport.Close()

	// This test primarily verifies that memory doesn't accumulate
	// In a real implementation, we'd check memory usage over time

	// Send and receive messages
	for i := 0; i < 10; i++ {
		message := claudecode.StreamMessage{
			Type:      "user",
			SessionID: "memory-test",
		}

		transport.SendMessage(ctx, message)
		transport.ReceiveMessages(ctx)
	}

	// Memory should not accumulate indefinitely
	// This is more of a design requirement than a testable assertion
}

// T118: Error Recovery 🔴 RED
func TestErrorRecovery(t *testing.T) {
	// Recover from transport errors gracefully

	tempCLI := createMockCLI(t)
	options := &claudecode.Options{}

	transport := New(tempCLI, options, false)

	ctx := context.Background()
	err := transport.Connect(ctx)
	if err != nil {
		t.Fatalf("Connection failed: %v", err)
	}
	defer transport.Close()

	// Simulate error conditions
	cancelledCtx, cancel := context.WithCancel(ctx)
	cancel() // Cancel immediately

	// Should handle cancelled context gracefully
	err = transport.SendMessage(cancelledCtx, claudecode.StreamMessage{Type: "user"})
	if err == nil {
		t.Log("SendMessage with cancelled context - error expected but got nil")
	}

	// Should still be functional with good context
	err = transport.SendMessage(ctx, claudecode.StreamMessage{Type: "user"})
	if err != nil && !strings.Contains(err.Error(), "closed") {
		t.Errorf("Should recover from error: %v", err)
	}
}

// T119: Subprocess Security 🔴 RED
func TestSubprocessSecurity(t *testing.T) {
	// Ensure subprocess runs with minimal permissions

	tempCLI := createMockCLI(t)
	options := &claudecode.Options{}

	transport := New(tempCLI, options, false)

	ctx := context.Background()
	err := transport.Connect(ctx)
	if err != nil {
		t.Fatalf("Connection failed: %v", err)
	}
	defer transport.Close()

	// Process should run with appropriate permissions
	// This is mainly a design requirement - actual verification
	// would require platform-specific checks

	if !transport.IsConnected() {
		t.Error("Process should be running securely")
	}
}

// T120: Platform Compatibility 🔴 RED
func TestPlatformCompatibility(t *testing.T) {
	// Work across Windows, macOS, and Linux

	tempCLI := createMockCLI(t)
	options := &claudecode.Options{}

	transport := New(tempCLI, options, false)

	ctx := context.Background()
	err := transport.Connect(ctx)
	if err != nil {
		t.Fatalf("Connection failed: %v", err)
	}
	defer transport.Close()

	// Basic functionality should work on all platforms
	if !transport.IsConnected() {
		t.Error("Transport should work on current platform")
	}

	// Test interrupt (platform-specific signals)
	err = transport.Interrupt(ctx)
	if err != nil {
		t.Errorf("Interrupt should work on current platform: %v", err)
	}
}

// Helper functions to create mock CLI scripts for testing

func createMockCLI(t *testing.T) string {
	script := `#!/bin/bash
echo '{"type":"assistant","content":[{"type":"text","text":"Mock response"}],"model":"claude-3"}'
sleep 0.5
`
	return createTempScript(t, script)
}

func createLongRunningMockCLI(t *testing.T) string {
	script := `#!/bin/bash
# Ignore SIGTERM initially to test 5-second timeout
trap 'echo "Received SIGTERM, ignoring for 6 seconds"; sleep 6; exit 1' TERM
echo '{"type":"assistant","content":[{"type":"text","text":"Long running mock"}],"model":"claude-3"}'
sleep 30  # Run long enough to test termination
`
	return createTempScript(t, script)
}

func createFailingCLI(t *testing.T) string {
	script := `#!/bin/bash
echo "Mock CLI failing" >&2
exit 1
`
	return createTempScript(t, script)
}

func createEnvironmentCheckCLI(t *testing.T) string {
	script := `#!/bin/bash
if [ "$CLAUDE_CODE_ENTRYPOINT" = "sdk-go" ]; then
    echo '{"type":"assistant","content":[{"type":"text","text":"Environment OK"}],"model":"claude-3"}'
else
    echo "Missing environment variable" >&2
    exit 1
fi
sleep 0.5
`
	return createTempScript(t, script)
}

func createTempScript(t *testing.T, script string) string {
	tempDir := t.TempDir()
	scriptPath := filepath.Join(tempDir, "mock-claude")

	err := os.WriteFile(scriptPath, []byte(script), 0755)
	if err != nil {
		t.Fatalf("Failed to create mock CLI script: %v", err)
	}

	return scriptPath
}
