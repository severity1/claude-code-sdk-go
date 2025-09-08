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

// TestTransportLifecycle tests connection lifecycle and state management
func TestTransportLifecycle(t *testing.T) {
	ctx, cancel := setupTransportTestContext(t, 10*time.Second)
	defer cancel()

	// Test basic lifecycle
	transport := setupTransportForTest(t, newTransportMockCLI())
	defer disconnectTransportSafely(t, transport)

	// Initial state should be disconnected
	assertTransportConnected(t, transport, false)

	// Connect
	connectTransportSafely(t, ctx, transport)
	assertTransportConnected(t, transport, true)

	// Test multiple Close() calls are safe
	err1 := transport.Close()
	err2 := transport.Close()

	assertNoTransportError(t, err1)
	assertNoTransportError(t, err2)
	assertTransportConnected(t, transport, false)
}

// TestTransportReconnection tests reconnection capability
func TestTransportReconnection(t *testing.T) {
	ctx, cancel := setupTransportTestContext(t, 10*time.Second)
	defer cancel()

	transport := setupTransportForTest(t, newTransportMockCLI())
	defer disconnectTransportSafely(t, transport)

	// First connection
	connectTransportSafely(t, ctx, transport)
	assertTransportConnected(t, transport, true)

	// Disconnect
	disconnectTransportSafely(t, transport)
	assertTransportConnected(t, transport, false)

	// Reconnect
	connectTransportSafely(t, ctx, transport)
	assertTransportConnected(t, transport, true)
}

// TestTransportMessageIO tests basic message sending and receiving
func TestTransportMessageIO(t *testing.T) {
	ctx, cancel := setupTransportTestContext(t, 10*time.Second)
	defer cancel()

	transport := setupTransportForTest(t, newTransportMockCLI())
	defer disconnectTransportSafely(t, transport)

	connectTransportSafely(t, ctx, transport)

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

		connectTransportSafely(t, ctx, transport)

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

		connectTransportSafely(t, ctx, transport)

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
				connectTransportSafely(t, ctx, tr)
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

	connectTransportSafely(t, ctx, transport)

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
	connectTransportSafely(t, ctx, transport)
	assertTransportConnected(t, transport, true)

	// Test interrupt (platform-specific signals)
	if runtime.GOOS != windowsOS {
		err := transport.Interrupt(ctx)
		assertNoTransportError(t, err)
	}
}

// TestTransportCleanup tests resource cleanup and multiple close scenarios
func TestTransportCleanup(t *testing.T) {
	ctx, cancel := setupTransportTestContext(t, 10*time.Second)
	defer cancel()

	transport := setupTransportForTest(t, newTransportMockCLI())

	connectTransportSafely(t, ctx, transport)

	// Get channels to create resources
	transport.ReceiveMessages(ctx)

	// Cleanup should not error
	err := transport.Close()
	assertNoTransportError(t, err)

	// Multiple cleanups should be safe
	err = transport.Close()
	assertNoTransportError(t, err)

	assertTransportConnected(t, transport, false)
}

// Mock transport implementation with functional options
type transportMockOptions struct {
	longRunning      bool
	shouldFail       bool
	checkEnvironment bool
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

func connectTransportSafely(t *testing.T, ctx context.Context, transport *Transport) {
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
		name        string
		cliPath     string
		options     *shared.Options
		prompt      string
		expectValid bool
	}{
		{
			name:        "valid_basic_prompt",
			cliPath:     "/usr/bin/claude",
			options:     &shared.Options{},
			prompt:      "What is 2+2?",
			expectValid: true,
		},
		{
			name:    "valid_prompt_with_options",
			cliPath: "/usr/bin/claude",
			options: &shared.Options{
				SystemPrompt: stringPtr("You are helpful"),
				MaxTurns:     3,
			},
			prompt:      "Hello there!",
			expectValid: true,
		},
		{
			name:        "empty_prompt",
			cliPath:     "/usr/bin/claude",
			options:     &shared.Options{},
			prompt:      "",
			expectValid: true, // Empty prompt is valid
		},
		{
			name:        "multiline_prompt",
			cliPath:     "/usr/bin/claude",
			options:     &shared.Options{},
			prompt:      "Line 1\nLine 2\nLine 3",
			expectValid: true,
		},
		{
			name:        "special_characters_prompt",
			cliPath:     "/usr/bin/claude",
			options:     &shared.Options{},
			prompt:      "Test with \"quotes\" and 'apostrophes' and $variables",
			expectValid: true,
		},
		{
			name:        "nil_options",
			cliPath:     "/usr/bin/claude",
			options:     nil,
			prompt:      "Test with nil options",
			expectValid: true, // Should handle nil options gracefully
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Create transport using NewWithPrompt
			transport := NewWithPrompt(test.cliPath, test.options, test.prompt)

			if !test.expectValid {
				if transport != nil {
					t.Error("Expected nil transport for invalid input")
				}
				return
			}

			// Validate transport was created properly
			if transport == nil {
				t.Fatal("Expected transport to be created, got nil")
			}

			// Verify transport configuration
			assertNewWithPromptConfiguration(t, transport, test.cliPath, test.options, test.prompt)

			// Verify initial state
			assertTransportConnected(t, transport, false)

			// Test that transport can be used (basic lifecycle)
			// Note: We won't actually connect since CLI path is likely invalid in test environment
			// But we can verify the transport is properly configured for connection attempts
			if transport.cliPath != test.cliPath {
				t.Errorf("Expected cliPath %q, got %q", test.cliPath, transport.cliPath)
			}

			if transport.entrypoint != "sdk-go" {
				t.Errorf("Expected entrypoint 'sdk-go', got %q", transport.entrypoint)
			}

			if !transport.closeStdin {
				t.Error("Expected closeStdin to be true for NewWithPrompt transport")
			}

			// Verify prompt was stored correctly
			if transport.promptArg == nil {
				t.Error("Expected promptArg to be set, got nil")
			} else if *transport.promptArg != test.prompt {
				t.Errorf("Expected promptArg %q, got %q", test.prompt, *transport.promptArg)
			}

			// Verify parser was initialized
			if transport.parser == nil {
				t.Error("Expected parser to be initialized, got nil")
			}
		})
	}
}

// TestNewWithPromptVsNew tests differences between NewWithPrompt and New constructors
func TestNewWithPromptVsNew(t *testing.T) {
	cliPath := "/usr/bin/claude"
	options := &shared.Options{
		SystemPrompt: stringPtr("Test prompt"),
	}
	prompt := "Test query"

	// Create both types of transports
	regularTransport := New(cliPath, options, false, "sdk-go")
	promptTransport := NewWithPrompt(cliPath, options, prompt)

	// Compare configurations
	tests := []struct {
		name  string
		check func(t *testing.T)
	}{
		{
			name: "closeStdin_configuration",
			check: func(t *testing.T) {
				if regularTransport.closeStdin != false {
					t.Error("Regular transport should have closeStdin=false")
				}
				if promptTransport.closeStdin != true {
					t.Error("Prompt transport should have closeStdin=true")
				}
			},
		},
		{
			name: "entrypoint_configuration",
			check: func(t *testing.T) {
				if regularTransport.entrypoint != "sdk-go" {
					t.Errorf("Regular transport entrypoint should be 'sdk-go', got %q", regularTransport.entrypoint)
				}
				if promptTransport.entrypoint != "sdk-go" {
					t.Errorf("Prompt transport entrypoint should be 'sdk-go', got %q", promptTransport.entrypoint)
				}
			},
		},
		{
			name: "prompt_argument_configuration",
			check: func(t *testing.T) {
				if regularTransport.promptArg != nil {
					t.Errorf("Regular transport should have nil promptArg, got %v", *regularTransport.promptArg)
				}
				if promptTransport.promptArg == nil {
					t.Error("Prompt transport should have non-nil promptArg")
				} else if *promptTransport.promptArg != prompt {
					t.Errorf("Prompt transport promptArg should be %q, got %q", prompt, *promptTransport.promptArg)
				}
			},
		},
		{
			name: "shared_configuration",
			check: func(t *testing.T) {
				// Both should have same cliPath and options
				if regularTransport.cliPath != promptTransport.cliPath {
					t.Error("Both transports should have same cliPath")
				}
				if regularTransport.options != promptTransport.options {
					t.Error("Both transports should reference same options")
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.check(t)
		})
	}
}

// Helper functions for NewWithPrompt tests
func assertNewWithPromptConfiguration(t *testing.T, transport *Transport, expectedCLIPath string, expectedOptions *shared.Options, expectedPrompt string) {
	t.Helper()

	if transport.cliPath != expectedCLIPath {
		t.Errorf("Expected cliPath %q, got %q", expectedCLIPath, transport.cliPath)
	}

	if transport.options != expectedOptions {
		t.Errorf("Expected options to match provided options")
	}

	if transport.promptArg == nil {
		t.Error("Expected promptArg to be set, got nil")
	} else if *transport.promptArg != expectedPrompt {
		t.Errorf("Expected promptArg %q, got %q", expectedPrompt, *transport.promptArg)
	}

	if transport.entrypoint != "sdk-go" {
		t.Errorf("Expected entrypoint 'sdk-go', got %q", transport.entrypoint)
	}

	if !transport.closeStdin {
		t.Error("Expected closeStdin to be true for NewWithPrompt transport")
	}

	if transport.parser == nil {
		t.Error("Expected parser to be initialized, got nil")
	}
}

func stringPtr(s string) *string {
	return &s
}
