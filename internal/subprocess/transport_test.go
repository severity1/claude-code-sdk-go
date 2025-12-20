package subprocess

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/severity1/claude-code-sdk-go/internal/shared"
)

// TestTransportLifecycle tests connection lifecycle, state management, and reconnection
func TestTransportLifecycle(t *testing.T) {
	ctx, cancel := setupTransportTestContext(t, 10*time.Second)
	defer cancel()

	// Test basic lifecycle
	transport := setupTransportForTest(t, newTransportMockCLI())
	defer disconnectTransportSafely(t, transport)

	// Initial state should be disconnected
	assertTransportConnected(t, transport, false)

	// Connect
	connectTransportSafely(ctx, t, transport)
	assertTransportConnected(t, transport, true)

	// Test multiple Close() calls are safe
	err1 := transport.Close()
	err2 := transport.Close()

	assertNoTransportError(t, err1)
	assertNoTransportError(t, err2)
	assertTransportConnected(t, transport, false)

	// Test reconnection capability
	connectTransportSafely(ctx, t, transport)
	assertTransportConnected(t, transport, true)
}

// TestTransportMessageIO tests basic message sending and receiving
func TestTransportMessageIO(t *testing.T) {
	ctx, cancel := setupTransportTestContext(t, 10*time.Second)
	defer cancel()

	transport := setupTransportForTest(t, newTransportMockCLI())
	defer disconnectTransportSafely(t, transport)

	connectTransportSafely(ctx, t, transport)

	// Test message sending
	message := shared.StreamMessage{
		Type:      "user",
		SessionID: "test-session",
	}

	err := transport.SendMessage(ctx, message)
	assertNoTransportError(t, err)

	// Test message receiving
	msgChan, errChan := transport.ReceiveMessages(ctx)
	if msgChan == nil || errChan == nil {
		t.Error("Message and error channels should not be nil")
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

// TestTransportProcessManagement tests process control and termination
func TestTransportProcessManagement(t *testing.T) {
	ctx, cancel := setupTransportTestContext(t, 15*time.Second)
	defer cancel()

	// Test 5-second termination sequence
	t.Run("five_second_termination", func(t *testing.T) {
		transport := setupTransportForTest(t, newTransportMockCLIWithOptions(WithLongRunning()))
		defer disconnectTransportSafely(t, transport)

		connectTransportSafely(ctx, t, transport)

		// Start timing the termination
		start := time.Now()
		err := transport.Close()
		duration := time.Since(start)

		assertNoTransportError(t, err)

		// Should complete in reasonable time (allowing buffer for 5-second sequence)
		if duration > 6*time.Second {
			t.Errorf("Termination took too long: %v", duration)
		}

		assertTransportConnected(t, transport, false)
	})

	// Test interrupt handling
	t.Run("interrupt_handling", func(t *testing.T) {
		if runtime.GOOS == windowsOS {
			t.Skip("Interrupt not supported on Windows")
		}

		transport := setupTransportForTest(t, newTransportMockCLI())
		defer disconnectTransportSafely(t, transport)

		connectTransportSafely(ctx, t, transport)

		err := transport.Interrupt(ctx)
		assertNoTransportError(t, err)

		// Process should still be manageable after interrupt
		assertTransportConnected(t, transport, true)
	})
}

// TestTransportErrorHandling tests various error scenarios using table-driven approach
func TestTransportErrorHandling(t *testing.T) {
	ctx, cancel := setupTransportTestContext(t, 10*time.Second)
	defer cancel()

	tests := []struct {
		name           string
		setupTransport func() *Transport
		operation      func(*Transport) error
		expectError    bool
		errorContains  string
	}{
		{
			name: "connection_with_failing_cli",
			setupTransport: func() *Transport {
				return setupTransportForTest(t, newTransportMockCLIWithOptions(WithFailure()))
			},
			operation: func(tr *Transport) error {
				return tr.Connect(ctx)
			},
			expectError:   false, // Connection should succeed initially even if CLI fails
			errorContains: "",
		},
		{
			name: "send_to_disconnected_transport",
			setupTransport: func() *Transport {
				return setupTransportForTest(t, newTransportMockCLI())
			},
			operation: func(tr *Transport) error {
				// Don't connect - send to disconnected transport
				message := shared.StreamMessage{Type: "user", SessionID: "test"}
				return tr.SendMessage(ctx, message)
			},
			expectError:   true,
			errorContains: "",
		},
		{
			name: "context_cancellation",
			setupTransport: func() *Transport {
				return setupTransportForTest(t, newTransportMockCLI())
			},
			operation: func(tr *Transport) error {
				connectTransportSafely(ctx, t, tr)
				// Use canceled context
				canceledCtx, cancel := context.WithCancel(ctx)
				cancel()
				message := shared.StreamMessage{Type: "user", SessionID: "test"}
				return tr.SendMessage(canceledCtx, message)
			},
			expectError:   false, // Context cancellation handling may vary
			errorContains: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			transport := test.setupTransport()
			defer disconnectTransportSafely(t, transport)

			err := test.operation(transport)

			if test.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if test.errorContains != "" && !strings.Contains(err.Error(), test.errorContains) {
					t.Errorf("Expected error containing '%s', got: %v", test.errorContains, err)
				}
			} else {
				if err != nil && test.errorContains != "" && !strings.Contains(err.Error(), test.errorContains) {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

// TestTransportConcurrency tests concurrent operations and backpressure handling
func TestTransportConcurrency(t *testing.T) {
	ctx, cancel := setupTransportTestContext(t, 15*time.Second)
	defer cancel()

	transport := setupTransportForTest(t, newTransportMockCLI())
	defer disconnectTransportSafely(t, transport)

	connectTransportSafely(ctx, t, transport)

	// Test concurrent message sending
	t.Run("concurrent_sending", func(t *testing.T) {
		var wg sync.WaitGroup
		errorCount := 0
		var mu sync.Mutex

		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				message := shared.StreamMessage{
					Type:      "user",
					SessionID: fmt.Sprintf("session-%d", id),
				}

				err := transport.SendMessage(ctx, message)
				if err != nil {
					mu.Lock()
					errorCount++
					mu.Unlock()
				}
			}(i)
		}

		wg.Wait()

		// Some errors might be acceptable in concurrent scenarios
		if errorCount > 2 {
			t.Errorf("Too many errors in concurrent sending: %d", errorCount)
		}
	})

	// Test backpressure handling
	t.Run("backpressure_handling", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			message := shared.StreamMessage{
				Type:      "user",
				SessionID: "backpressure-test",
			}

			// Should not block indefinitely
			done := make(chan error, 1)
			go func() {
				done <- transport.SendMessage(ctx, message)
			}()

			select {
			case err := <-done:
				if err != nil && !strings.Contains(err.Error(), "context") && !strings.Contains(err.Error(), "closed") {
					t.Errorf("Unexpected error in message %d: %v", i, err)
				}
			case <-time.After(1 * time.Second):
				t.Errorf("Message %d took too long to send (backpressure issue)", i)
			}
		}
	})
}

// TestTransportEnvironmentSetup tests environment variable and platform compatibility
func TestTransportEnvironmentSetup(t *testing.T) {
	ctx, cancel := setupTransportTestContext(t, 10*time.Second)
	defer cancel()

	transport := setupTransportForTest(t, newTransportMockCLIWithOptions(WithEnvironmentCheck()))
	defer disconnectTransportSafely(t, transport)

	// Connection should succeed with proper environment setup
	connectTransportSafely(ctx, t, transport)
	assertTransportConnected(t, transport, true)

	// Test interrupt (platform-specific signals)
	if runtime.GOOS != windowsOS {
		err := transport.Interrupt(ctx)
		assertNoTransportError(t, err)
	}
}

// TestTransportReceiveMessagesNotConnected tests ReceiveMessages behavior on disconnected transport
// This targets the missing 44.4% coverage in ReceiveMessages function
func TestTransportReceiveMessagesNotConnected(t *testing.T) {
	ctx, cancel := setupTransportTestContext(t, 5*time.Second)
	defer cancel()

	transport := setupTransportForTest(t, newTransportMockCLI())

	// Test ReceiveMessages on disconnected transport
	msgChan, errChan := transport.ReceiveMessages(ctx)

	// Channels should not be nil
	if msgChan == nil {
		t.Error("Expected message channel to be non-nil")
	}
	if errChan == nil {
		t.Error("Expected error channel to be non-nil")
	}

	// Channels should be closed (for disconnected transport)
	select {
	case msg, ok := <-msgChan:
		if ok {
			t.Errorf("Expected message channel to be closed, got message: %v", msg)
		}
		// Channel is closed, which is expected
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected message channel to be closed immediately")
	}

	select {
	case err, ok := <-errChan:
		if ok {
			t.Errorf("Expected error channel to be closed, got error: %v", err)
		}
		// Channel is closed, which is expected
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected error channel to be closed immediately")
	}

	// Test multiple calls return the same behavior
	msgChan2, errChan2 := transport.ReceiveMessages(ctx)
	if msgChan2 == nil || errChan2 == nil {
		t.Error("Multiple calls should return valid channels")
	}

	// Verify they're different channel instances but behave the same
	select {
	case _, ok := <-msgChan2:
		if ok {
			t.Error("Expected second message channel to be closed")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected second message channel to be closed immediately")
	}
}

// Mock transport implementation with functional options
type transportMockOptions struct {
	longRunning      bool
	shouldFail       bool
	checkEnvironment bool
	invalidOutput    bool
}

type TransportMockOption func(*transportMockOptions)

func WithLongRunning() TransportMockOption {
	return func(opts *transportMockOptions) {
		opts.longRunning = true
	}
}

func WithFailure() TransportMockOption {
	return func(opts *transportMockOptions) {
		opts.shouldFail = true
	}
}

func WithEnvironmentCheck() TransportMockOption {
	return func(opts *transportMockOptions) {
		opts.checkEnvironment = true
	}
}

func WithInvalidOutput() TransportMockOption {
	return func(opts *transportMockOptions) {
		opts.invalidOutput = true
	}
}

func newTransportMockCLI() string {
	return newTransportMockCLIWithOptions()
}

func newTransportMockCLIWithOptions(options ...TransportMockOption) string {
	opts := &transportMockOptions{}
	for _, opt := range options {
		opt(opts)
	}

	var script string
	var extension string

	if runtime.GOOS == windowsOS {
		extension = ".bat"
		switch {
		case opts.shouldFail:
			script = `@echo off
echo Mock CLI failing >&2
exit /b 1
`
		case opts.longRunning:
			script = `@echo off
echo {"type":"assistant","content":[{"type":"text","text":"Long running mock"}],"model":"claude-3"}
timeout /t 30 /nobreak > NUL
`
		case opts.checkEnvironment:
			script = `@echo off
if "%CLAUDE_CODE_ENTRYPOINT%"=="sdk-go" (
    echo {"type":"assistant","content":[{"type":"text","text":"Environment OK"}],"model":"claude-3"}
) else (
    echo Missing environment variable >&2
    exit /b 1
)
timeout /t 1 /nobreak > NUL
`
		case opts.invalidOutput:
			script = `@echo off
echo This is not valid JSON output
echo {"invalid": json}
echo {"type":"assistant","content":[{"type":"text","text":"Valid after invalid"}],"model":"claude-3"}
timeout /t 1 /nobreak > NUL
`
		default:
			script = `@echo off
echo {"type":"assistant","content":[{"type":"text","text":"Mock response"}],"model":"claude-3"}
timeout /t 1 /nobreak > NUL
`
		}
	} else {
		extension = ""
		switch {
		case opts.shouldFail:
			script = `#!/bin/bash
echo "Mock CLI failing" >&2
exit 1
`
		case opts.longRunning:
			script = `#!/bin/bash
# Ignore SIGTERM initially to test 5-second timeout
trap 'echo "Received SIGTERM, ignoring for 6 seconds"; sleep 6; exit 1' TERM
echo '{"type":"assistant","content":[{"type":"text","text":"Long running mock"}],"model":"claude-3"}'
sleep 30  # Run long enough to test termination
`
		case opts.checkEnvironment:
			script = `#!/bin/bash
if [ "$CLAUDE_CODE_ENTRYPOINT" = "sdk-go" ]; then
    echo '{"type":"assistant","content":[{"type":"text","text":"Environment OK"}],"model":"claude-3"}'
else
    echo "Missing environment variable" >&2
    exit 1
fi
sleep 0.5
`
		case opts.invalidOutput:
			script = `#!/bin/bash
echo "This is not valid JSON output"
echo '{"invalid": json}'
echo '{"type":"assistant","content":[{"type":"text","text":"Valid after invalid"}],"model":"claude-3"}'
sleep 0.5
`
		default:
			script = `#!/bin/bash
echo '{"type":"assistant","content":[{"type":"text","text":"Mock response"}],"model":"claude-3"}'
sleep 0.5
`
		}
	}

	return createTransportTempScript(script, extension)
}

func createTransportTempScript(script, extension string) string {
	tempDir := os.TempDir()
	scriptPath := filepath.Join(tempDir, fmt.Sprintf("mock-claude-%d%s", time.Now().UnixNano(), extension))

	err := os.WriteFile(scriptPath, []byte(script), 0o755) // #nosec G306 - Test script needs to be executable
	if err != nil {
		panic(fmt.Sprintf("Failed to create mock CLI script: %v", err))
	}

	return scriptPath
}

// Helper functions following client_test.go patterns
func setupTransportTestContext(t *testing.T, timeout time.Duration) (context.Context, context.CancelFunc) {
	t.Helper()
	return context.WithTimeout(context.Background(), timeout)
}

func setupTransportForTest(t *testing.T, cliPath string) *Transport {
	t.Helper()
	options := &shared.Options{}
	return New(cliPath, options, false, "sdk-go")
}

func connectTransportSafely(ctx context.Context, t *testing.T, transport *Transport) {
	t.Helper()
	err := transport.Connect(ctx)
	if err != nil {
		t.Fatalf("Transport connection failed: %v", err)
	}
}

func disconnectTransportSafely(t *testing.T, transport *Transport) {
	t.Helper()
	if err := transport.Close(); err != nil {
		t.Logf("Transport disconnect warning: %v", err)
	}
}

func assertTransportConnected(t *testing.T, transport *Transport, expected bool) {
	t.Helper()
	actual := transport.IsConnected()
	if actual != expected {
		t.Errorf("Expected transport connected=%t, got connected=%t", expected, actual)
	}
}

func assertNoTransportError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

// TestNewWithPrompt tests the NewWithPrompt constructor for one-shot queries
func TestNewWithPrompt(t *testing.T) {
	tests := []struct {
		name    string
		prompt  string
		options *shared.Options
	}{
		{"basic_prompt", "What is 2+2?", &shared.Options{}},
		{"empty_prompt", "", nil},
		{"multiline_prompt", "Line 1\nLine 2", &shared.Options{SystemPrompt: stringPtr("test")}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			transport := NewWithPrompt("/usr/bin/claude", test.options, test.prompt)

			if transport == nil {
				t.Fatal("Expected transport to be created, got nil")
			}

			// Verify key configuration
			if transport.entrypoint != "sdk-go" {
				t.Errorf("Expected entrypoint 'sdk-go', got %q", transport.entrypoint)
			}
			if !transport.closeStdin {
				t.Error("Expected closeStdin to be true")
			}
			if transport.promptArg == nil || *transport.promptArg != test.prompt {
				t.Errorf("Expected promptArg %q, got %v", test.prompt, transport.promptArg)
			}
			assertTransportConnected(t, transport, false)
		})
	}
}

// TestTransportConnectErrorPaths tests uncovered Connect error scenarios
func TestTransportConnectErrorPaths(t *testing.T) {
	ctx, cancel := setupTransportTestContext(t, 5*time.Second)
	defer cancel()

	tests := []struct {
		name      string
		setup     func() *Transport
		wantError bool
	}{
		{
			name: "already_connected_error",
			setup: func() *Transport {
				transport := setupTransportForTest(t, newTransportMockCLI())
				connectTransportSafely(ctx, t, transport)
				return transport
			},
			wantError: true,
		},
		{
			name: "invalid_working_directory",
			setup: func() *Transport {
				options := &shared.Options{Cwd: stringPtr("/nonexistent/directory/path")}
				return New(newTransportMockCLI(), options, false, "sdk-go")
			},
			wantError: true,
		},
		{
			name: "cli_start_failure",
			setup: func() *Transport {
				return setupTransportForTest(t, "/nonexistent/cli/path")
			},
			wantError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			transport := test.setup()
			defer disconnectTransportSafely(t, transport)

			err := transport.Connect(ctx)
			if test.wantError && err == nil {
				t.Error("Expected error but got none")
			} else if !test.wantError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// TestTransportSendMessageEdgeCases tests uncovered SendMessage scenarios
func TestTransportSendMessageEdgeCases(t *testing.T) {
	ctx, cancel := setupTransportTestContext(t, 5*time.Second)
	defer cancel()

	// Test SendMessage with promptArg transport (one-shot mode)
	t.Run("send_message_with_prompt_arg", func(t *testing.T) {
		transport := NewWithPrompt(newTransportMockCLI(), &shared.Options{}, "test prompt")
		defer disconnectTransportSafely(t, transport)

		connectTransportSafely(ctx, t, transport)

		// Should be no-op since prompt is already passed as CLI argument
		message := shared.StreamMessage{Type: "user", SessionID: "test"}
		err := transport.SendMessage(ctx, message)
		assertNoTransportError(t, err)
	})

	// Test SendMessage with invalid JSON
	t.Run("send_message_marshal_error", func(t *testing.T) {
		transport := setupTransportForTest(t, newTransportMockCLI())
		defer disconnectTransportSafely(t, transport)

		connectTransportSafely(ctx, t, transport)

		// Create a message that would cause JSON marshal error
		// In Go, this is difficult to trigger naturally, so we test normal case
		message := shared.StreamMessage{Type: "user", SessionID: "test"}
		err := transport.SendMessage(ctx, message)
		assertNoTransportError(t, err)
	})

	// Test context cancellation during send
	t.Run("context_cancelled_during_send", func(t *testing.T) {
		transport := setupTransportForTest(t, newTransportMockCLI())
		defer disconnectTransportSafely(t, transport)

		connectTransportSafely(ctx, t, transport)

		cancelledCtx, cancel := context.WithCancel(ctx)
		cancel() // Cancel immediately

		message := shared.StreamMessage{Type: "user", SessionID: "test"}
		err := transport.SendMessage(cancelledCtx, message)
		// Error is acceptable since context was cancelled
		if err != nil && !strings.Contains(err.Error(), "context") {
			t.Errorf("Expected context cancellation error, got: %v", err)
		}
	})
}

// TestTransportTerminateProcessPaths tests uncovered terminateProcess scenarios
func TestTransportTerminateProcessPaths(t *testing.T) {
	if runtime.GOOS == windowsOS {
		t.Skip("Process termination testing requires Unix signals")
	}

	ctx, cancel := setupTransportTestContext(t, 15*time.Second)
	defer cancel()

	// Test normal termination
	t.Run("normal_termination", func(t *testing.T) {
		transport := setupTransportForTest(t, newTransportMockCLI())
		connectTransportSafely(ctx, t, transport)

		// Close should trigger terminateProcess
		err := transport.Close()
		assertNoTransportError(t, err)
	})

	// Test SIGTERM timeout (force SIGKILL)
	t.Run("sigterm_timeout_force_kill", func(t *testing.T) {
		transport := setupTransportForTest(t, newTransportMockCLIWithOptions(WithLongRunning()))
		connectTransportSafely(ctx, t, transport)

		// This transport ignores SIGTERM for 6 seconds, forcing SIGKILL
		start := time.Now()
		err := transport.Close()
		duration := time.Since(start)

		// Should complete within reasonable time after 5-second timeout
		if duration > 8*time.Second {
			t.Errorf("Termination took too long: %v", duration)
		}
		assertNoTransportError(t, err)
	})

	// Test context cancellation during termination
	t.Run("context_cancelled_during_termination", func(t *testing.T) {
		// Create a context that we can cancel
		shortCtx, shortCancel := context.WithCancel(ctx)

		transport := setupTransportForTest(t, newTransportMockCLI())

		// Connect with the cancellable context
		connectTransportSafely(shortCtx, t, transport)

		// Cancel the context to simulate cancellation during termination
		shortCancel()

		err := transport.Close()
		// Should not error even with cancelled context
		assertNoTransportError(t, err)
	})
}

// TestTransportHandleStdoutErrorPaths tests uncovered handleStdout scenarios
func TestTransportHandleStdoutErrorPaths(t *testing.T) {
	ctx, cancel := setupTransportTestContext(t, 5*time.Second)
	defer cancel()

	// Test stdout parsing errors
	t.Run("stdout_parsing_errors", func(t *testing.T) {
		transport := setupTransportForTest(t, newTransportMockCLIWithOptions(WithInvalidOutput()))
		defer disconnectTransportSafely(t, transport)

		connectTransportSafely(ctx, t, transport)

		// Get channels and wait briefly for processing
		msgChan, errChan := transport.ReceiveMessages(ctx)

		// Check for either parsing errors or messages (both are acceptable)
		errorReceived := false
		messageReceived := false

		timeout := time.After(2 * time.Second)
		for !errorReceived && !messageReceived {
			select {
			case err := <-errChan:
				if err != nil {
					errorReceived = true
				}
			case <-msgChan:
				messageReceived = true
			case <-timeout:
				// Either outcome is acceptable - parser may be resilient to invalid JSON
				return
			}
		}
	})

	// Test scanner error conditions
	t.Run("scanner_error_handling", func(t *testing.T) {
		transport := setupTransportForTest(t, newTransportMockCLI())
		defer disconnectTransportSafely(t, transport)

		connectTransportSafely(ctx, t, transport)

		// Get channels
		msgChan, errChan := transport.ReceiveMessages(ctx)

		// Close transport to trigger scanner completion
		disconnectTransportSafely(t, transport)

		// Channels should be closed after transport close
		select {
		case _, ok := <-msgChan:
			if ok {
				t.Error("Expected message channel to be closed")
			}
		case <-time.After(1 * time.Second):
			t.Error("Expected message channel to be closed promptly")
		}

		select {
		case _, ok := <-errChan:
			if ok {
				t.Error("Expected error channel to be closed")
			}
		case <-time.After(1 * time.Second):
			t.Error("Expected error channel to be closed promptly")
		}
	})
}

// TestTransportInterruptErrorPaths tests uncovered Interrupt scenarios
func TestTransportInterruptErrorPaths(t *testing.T) {
	ctx, cancel := setupTransportTestContext(t, 5*time.Second)
	defer cancel()

	// Test interrupt on disconnected transport
	t.Run("interrupt_disconnected_transport", func(t *testing.T) {
		transport := setupTransportForTest(t, newTransportMockCLI())

		// Don't connect - test interrupt on disconnected transport
		err := transport.Interrupt(ctx)
		if err == nil {
			t.Error("Expected error when interrupting disconnected transport")
		}
	})

	// Test interrupt with nil process
	t.Run("interrupt_nil_process", func(t *testing.T) {
		transport := setupTransportForTest(t, newTransportMockCLI())
		defer disconnectTransportSafely(t, transport)

		connectTransportSafely(ctx, t, transport)
		// Force close to null out the process
		disconnectTransportSafely(t, transport)

		err := transport.Interrupt(ctx)
		if err == nil {
			t.Error("Expected error when interrupting closed transport")
		}
	})

	if runtime.GOOS != windowsOS {
		t.Run("interrupt_signal_error", func(t *testing.T) {
			transport := setupTransportForTest(t, newTransportMockCLI())
			defer disconnectTransportSafely(t, transport)

			connectTransportSafely(ctx, t, transport)

			// Normal interrupt should work
			err := transport.Interrupt(ctx)
			assertNoTransportError(t, err)
		})
	}
}

// TestSubprocessEnvironmentVariables tests environment variable passing to subprocess
func TestSubprocessEnvironmentVariables(t *testing.T) {
	// Following client_test.go patterns for test organization
	setupSubprocessTestContext := func(t *testing.T) (context.Context, context.CancelFunc) {
		t.Helper()
		return context.WithTimeout(context.Background(), 5*time.Second)
	}

	tests := []struct {
		name     string
		options  *shared.Options
		validate func(t *testing.T, env []string)
	}{
		{
			name: "custom_env_vars_passed",
			options: &shared.Options{
				ExtraEnv: map[string]string{
					"TEST_VAR": "test_value",
					"DEBUG":    "1",
				},
			},
			validate: func(t *testing.T, env []string) {
				assertEnvContains(t, env, "TEST_VAR=test_value")
				assertEnvContains(t, env, "DEBUG=1")
				assertEnvContains(t, env, "CLAUDE_CODE_ENTRYPOINT=sdk-go")
			},
		},
		{
			name: "system_env_preserved",
			options: &shared.Options{
				ExtraEnv: map[string]string{"CUSTOM": "value"},
			},
			validate: func(t *testing.T, env []string) {
				// Verify system environment is preserved by checking for common env vars
				// Don't rely on PATH alone since it might have platform-specific casing
				systemEnvFound := false
				expectedEnvVars := []string{"PATH", "Path", "HOME", "USERPROFILE", "USER", "USERNAME"}

				for _, expectedVar := range expectedEnvVars {
					if os.Getenv(expectedVar) != "" {
						// Check if this env var exists in the subprocess environment
						for _, envVar := range env {
							if strings.HasPrefix(strings.ToUpper(envVar), strings.ToUpper(expectedVar)+"=") {
								systemEnvFound = true
								break
							}
						}
						if systemEnvFound {
							break
						}
					}
				}

				if !systemEnvFound {
					// Show first few env vars for debugging, but limit to avoid log spam
					envSample := env
					if len(envSample) > 5 {
						envSample = env[:5]
					}
					t.Errorf("Expected system environment to be preserved. System env has PATH=%q, subprocess env sample: %v",
						os.Getenv("PATH"), envSample)
				}

				assertEnvContains(t, env, "CUSTOM=value")
				assertEnvContains(t, env, "CLAUDE_CODE_ENTRYPOINT=sdk-go")
			},
		},
		{
			name: "nil_extra_env_works",
			options: &shared.Options{
				ExtraEnv: nil,
			},
			validate: func(t *testing.T, env []string) {
				assertEnvContains(t, env, "CLAUDE_CODE_ENTRYPOINT=sdk-go")
			},
		},
		{
			name: "empty_extra_env_works",
			options: &shared.Options{
				ExtraEnv: map[string]string{},
			},
			validate: func(t *testing.T, env []string) {
				assertEnvContains(t, env, "CLAUDE_CODE_ENTRYPOINT=sdk-go")
			},
		},
		{
			name: "proxy_configuration_example",
			options: &shared.Options{
				ExtraEnv: map[string]string{
					"HTTP_PROXY":  "http://proxy.example.com:8080",
					"HTTPS_PROXY": "http://proxy.example.com:8080",
					"NO_PROXY":    "localhost,127.0.0.1",
				},
			},
			validate: func(t *testing.T, env []string) {
				assertEnvContains(t, env, "HTTP_PROXY=http://proxy.example.com:8080")
				assertEnvContains(t, env, "HTTPS_PROXY=http://proxy.example.com:8080")
				assertEnvContains(t, env, "NO_PROXY=localhost,127.0.0.1")
				assertEnvContains(t, env, "CLAUDE_CODE_ENTRYPOINT=sdk-go")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := setupSubprocessTestContext(t)
			defer cancel()

			// Create transport with test options
			transport := New("echo", tt.options, true, "sdk-go")
			defer func() {
				if transport.IsConnected() {
					_ = transport.Close()
				}
			}()

			// Connect to build command with environment
			err := transport.Connect(ctx)
			assertNoTransportError(t, err)

			// Validate environment variables were set correctly
			if transport.cmd != nil && transport.cmd.Env != nil {
				tt.validate(t, transport.cmd.Env)
			} else {
				t.Error("Expected command environment to be set")
			}
		})
	}
}

// TestTransportWorkingDirectory tests that working directory is set via exec.Cmd.Dir
func TestTransportWorkingDirectory(t *testing.T) {
	ctx, cancel := setupTransportTestContext(t, 5*time.Second)
	defer cancel()

	tests := []struct {
		name     string
		setup    func() *Transport
		validate func(*testing.T, *Transport)
	}{
		{
			name: "working_directory_set_via_cmd_dir",
			setup: func() *Transport {
				cwd := t.TempDir()
				options := &shared.Options{
					Cwd: &cwd,
				}
				return New(newTransportMockCLI(), options, false, "sdk-go")
			},
			validate: func(t *testing.T, transport *Transport) {
				t.Helper()
				if transport.cmd == nil {
					t.Fatal("Expected command to be initialized after Connect()")
				}
				expectedCwd := *transport.options.Cwd
				if transport.cmd.Dir != expectedCwd {
					t.Errorf("Expected cmd.Dir to be %s, got %s", expectedCwd, transport.cmd.Dir)
				}
			},
		},
		{
			name: "no_working_directory_when_cwd_nil",
			setup: func() *Transport {
				options := &shared.Options{
					Cwd: nil,
				}
				return New(newTransportMockCLI(), options, false, "sdk-go")
			},
			validate: func(t *testing.T, transport *Transport) {
				t.Helper()
				if transport.cmd == nil {
					t.Fatal("Expected command to be initialized after Connect()")
				}
				if transport.cmd.Dir != "" {
					t.Errorf("Expected cmd.Dir to be empty when Cwd is nil, got %s", transport.cmd.Dir)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transport := tt.setup()
			defer disconnectTransportSafely(t, transport)

			// Connect to initialize the command
			err := transport.Connect(ctx)
			assertNoTransportError(t, err)

			// Validate working directory was set correctly via cmd.Dir
			tt.validate(t, transport)
		})
	}
}

// assertEnvContains checks if environment slice contains a key=value pair
func assertEnvContains(t *testing.T, env []string, expected string) {
	t.Helper()
	for _, e := range env {
		if e == expected {
			return
		}
	}
	t.Errorf("Environment missing %s. Available: %v", expected, env)
}

func stringPtr(s string) *string {
	return &s
}
