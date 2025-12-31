//go:build integration

package claudecode_test

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/severity1/claude-agent-sdk-go"
)

//go:embed testdata/cli_responses/*
var integrationFixtures embed.FS

// Helper Functions - following client_test.go patterns

// setupIntegrationContext creates a context for integration tests
func setupIntegrationContext(t *testing.T, timeout time.Duration) (context.Context, context.CancelFunc) {
	t.Helper()
	return context.WithTimeout(context.Background(), timeout)
}

// connectIntegrationClientSafely connects a client safely for testing
func connectIntegrationClientSafely(t *testing.T, ctx context.Context, client claudecode.Client) {
	t.Helper()
	if err := client.Connect(ctx); err != nil {
		t.Fatalf("Client connect failed: %v", err)
	}
}

// disconnectIntegrationClientSafely disconnects a client safely for testing
func disconnectIntegrationClientSafely(t *testing.T, client claudecode.Client) {
	t.Helper()
	if err := client.Disconnect(); err != nil {
		t.Errorf("Client disconnect failed: %v", err)
	}
}

// verifyIntegrationResourceCleanup verifies resources are properly cleaned up
func verifyIntegrationResourceCleanup(t *testing.T, transport *integrationMockTransport) {
	t.Helper()

	// Ensure transport is closed before verification - following client_test.go patterns
	if transport != nil {
		transport.Close()
	}

	transport.mu.Lock()
	connected := transport.connected
	closed := transport.closed
	transport.mu.Unlock()

	if connected {
		t.Error("Expected transport to be disconnected after test")
	}

	if !closed {
		t.Error("Expected transport to be closed after test")
	}

	// Verify resource tracker if available
	if transport.resourceTracker != nil {
		transport.resourceTracker.mu.Lock()
		goroutines := transport.resourceTracker.goroutines
		openFiles := transport.resourceTracker.openFiles
		transport.resourceTracker.mu.Unlock()

		if goroutines > 0 {
			t.Errorf("Resource leak detected: %d goroutines not cleaned up", goroutines)
		}
		if openFiles > 0 {
			t.Errorf("Resource leak detected: %d files not closed", openFiles)
		}
	}
}

// collectIntegrationMessages collects messages from an iterator
func collectIntegrationMessages(t *testing.T, ctx context.Context, iter claudecode.MessageIterator) []claudecode.Message {
	t.Helper()

	var messages []claudecode.Message
	start := time.Now()

	for {
		msg, err := iter.Next(ctx)
		if err != nil {
			if err == claudecode.ErrNoMoreMessages {
				t.Logf("Collected %d messages in %v", len(messages), time.Since(start))
				break
			}
			// Fail fast on context timeout/cancellation instead of continuing
			if err == context.DeadlineExceeded || err == context.Canceled {
				t.Fatalf("Iterator failed with timeout/cancellation after %v (collected %d messages): %v",
					time.Since(start), len(messages), err)
			}
			t.Logf("Iterator error (continuing) after %v: %v", time.Since(start), err)
			break
		}
		if msg != nil {
			messages = append(messages, msg)
			t.Logf("Collected message %d: %T", len(messages), msg)
		}
	}

	return messages
}

// assertIntegrationMessageCount verifies message count
func assertIntegrationMessageCount(t *testing.T, transport *integrationMockTransport, expected int) {
	t.Helper()
	transport.mu.Lock()
	actual := len(transport.sentMessages)
	transport.mu.Unlock()

	if actual != expected {
		t.Errorf("Expected %d sent messages, got %d", expected, actual)
	}
}

// assertIntegrationError verifies error conditions
func assertIntegrationError(t *testing.T, err error, wantErr bool, msgContains string) {
	t.Helper()
	if (err != nil) != wantErr {
		t.Errorf("error = %v, wantErr %v", err, wantErr)
		return
	}
	if wantErr && msgContains != "" && !strings.Contains(err.Error(), msgContains) {
		t.Errorf("error = %v, expected message to contain %q", err, msgContains)
	}
}

// loadIntegrationFixture loads a test fixture from embedded files
func loadIntegrationFixture(t *testing.T, name string) []claudecode.Message {
	t.Helper()

	data, err := integrationFixtures.ReadFile("testdata/cli_responses/" + name + ".json")
	if err != nil {
		// Return a simple mock message if fixture loading fails
		t.Logf("Failed to load fixture %s, using mock: %v", name, err)
		return []claudecode.Message{
			&claudecode.AssistantMessage{
				Content: []claudecode.ContentBlock{&claudecode.TextBlock{Text: "Mock response from fixture"}},
				Model:   "claude-3-5-sonnet-20241022",
			},
		}
	}

	var rawMessages []json.RawMessage
	if err := json.Unmarshal(data, &rawMessages); err != nil {
		t.Logf("Failed to unmarshal fixture %s, using mock: %v", name, err)
		return []claudecode.Message{
			&claudecode.AssistantMessage{
				Content: []claudecode.ContentBlock{&claudecode.TextBlock{Text: "Mock response from fixture"}},
				Model:   "claude-3-5-sonnet-20241022",
			},
		}
	}

	messages := make([]claudecode.Message, 0, len(rawMessages))
	for _, raw := range rawMessages {
		// Parse each message using the same parser logic as the SDK
		var msgType struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal(raw, &msgType); err != nil {
			t.Logf("Skipping malformed message in fixture %s: %v", name, err)
			continue
		}

		var msg claudecode.Message
		switch msgType.Type {
		case "user":
			msg = &claudecode.UserMessage{}
		case "assistant":
			msg = &claudecode.AssistantMessage{}
		case "system":
			msg = &claudecode.SystemMessage{}
		case "result":
			msg = &claudecode.ResultMessage{}
		default:
			t.Logf("Skipping unknown message type in fixture %s: %s", name, msgType.Type)
			continue
		}

		if err := json.Unmarshal(raw, msg); err != nil {
			t.Logf("Failed to parse message in fixture %s: %v", name, err)
			continue
		}

		messages = append(messages, msg)
	}

	if len(messages) == 0 {
		// Fallback to simple mock message
		messages = append(messages, &claudecode.AssistantMessage{
			Content: []claudecode.ContentBlock{&claudecode.TextBlock{Text: "Mock response"}},
			Model:   "claude-3-5-sonnet-20241022",
		})
	}

	return messages
}

// Mock Transport Factory Functions - following client_test.go patterns

type IntegrationMockTransportOption func(*integrationMockTransport)

func newIntegrationMockTransport(options ...IntegrationMockTransportOption) *integrationMockTransport {
	transport := &integrationMockTransport{
		testScenarios:   make(map[string]*integrationScenario),
		resourceTracker: &integrationResourceTracker{},
		// Initialize empty test messages slice if not set by options
		testMessages: []claudecode.Message{},
	}

	for _, option := range options {
		option(transport)
	}

	// If no messages were set by options, provide a default message for basic functionality
	if len(transport.testMessages) == 0 {
		transport.testMessages = []claudecode.Message{
			&claudecode.AssistantMessage{
				Content: []claudecode.ContentBlock{&claudecode.TextBlock{Text: "Default mock response"}},
				Model:   "claude-3-5-sonnet-20241022",
			},
		}
	}

	return transport
}

// Functional Options for Mock Transport

func WithIntegrationSimpleResponse(text string) IntegrationMockTransportOption {
	return func(t *integrationMockTransport) {
		t.testMessages = []claudecode.Message{
			&claudecode.AssistantMessage{
				Content: []claudecode.ContentBlock{&claudecode.TextBlock{Text: text}},
				Model:   "claude-3-5-sonnet-20241022",
			},
		}
	}
}

func WithIntegrationToolUsage() IntegrationMockTransportOption {
	return func(t *integrationMockTransport) {
		// Load actual tool usage fixture data (4 messages: assistant+tool → user+result → assistant+response → result)
		// For now, create proper messages manually since fixture loading needs test context
		t.testMessages = []claudecode.Message{
			&claudecode.AssistantMessage{
				Content: []claudecode.ContentBlock{
					&claudecode.TextBlock{Text: "I'll read the file for you."},
					&claudecode.ToolUseBlock{
						ToolUseID: "toolu_123456789",
						Name:      "Read",
						Input:     map[string]any{"file_path": "/example/file.txt"},
					},
				},
				Model: "claude-3-5-sonnet-20241022",
			},
			&claudecode.UserMessage{
				Content: []claudecode.ContentBlock{
					&claudecode.ToolResultBlock{
						ToolUseID: "toolu_123456789",
						Content:   "File contents: Hello, World!\nThis is a test file.",
						IsError:   boolPtr(false),
					},
				},
			},
			&claudecode.AssistantMessage{
				Content: []claudecode.ContentBlock{
					&claudecode.TextBlock{Text: "I've successfully read the file. The contents are:\n\nHello, World!\nThis is a test file.\n\nThe file contains a simple greeting and indicates it's a test file."},
				},
				Model: "claude-3-5-sonnet-20241022",
			},
			&claudecode.ResultMessage{
				SessionID:     "test-session-tool-456",
				IsError:       false,
				DurationMs:    2340,
				DurationAPIMs: 1560,
				NumTurns:      2,
				TotalCostUSD:  floatPtr(0.0034),
			},
		}
	}
}

func WithIntegrationStreamingResponse(parts ...string) IntegrationMockTransportOption {
	return func(t *integrationMockTransport) {
		messages := make([]claudecode.Message, len(parts))
		for i, part := range parts {
			messages[i] = &claudecode.AssistantMessage{
				Content: []claudecode.ContentBlock{&claudecode.TextBlock{Text: part}},
				Model:   "claude-3-5-sonnet-20241022",
			}
		}
		t.testMessages = messages
	}
}

func WithIntegrationInterruptScenario() IntegrationMockTransportOption {
	return func(t *integrationMockTransport) {
		t.testMessages = []claudecode.Message{
			&claudecode.AssistantMessage{
				Content: []claudecode.ContentBlock{&claudecode.TextBlock{Text: "This response can be interrupted"}},
				Model:   "claude-3-5-sonnet-20241022",
			},
		}
	}
}

func WithIntegrationSessionContinuation(sessionID string) IntegrationMockTransportOption {
	return func(t *integrationMockTransport) {
		scenario := &integrationScenario{
			name:      "session_continuation",
			sessionID: sessionID,
			messages: []claudecode.Message{
				&claudecode.UserMessage{Content: "Continue our previous conversation"},
				&claudecode.AssistantMessage{
					Content: []claudecode.ContentBlock{&claudecode.TextBlock{Text: "I remember our discussion"}},
					Model:   "claude-3-5-sonnet-20241022",
				},
			},
		}
		t.testScenarios["session_continuation"] = scenario
		t.testMessages = scenario.messages
	}
}

func WithIntegrationMCPServers() IntegrationMockTransportOption {
	return func(t *integrationMockTransport) {
		scenario := &integrationScenario{
			name:       "mcp_integration",
			mcpServers: true,
			messages: []claudecode.Message{
				&claudecode.AssistantMessage{
					Content: []claudecode.ContentBlock{&claudecode.TextBlock{Text: "MCP server integration working"}},
					Model:   "claude-3-5-sonnet-20241022",
				},
			},
		}
		t.testScenarios["mcp_integration"] = scenario
		t.testMessages = scenario.messages
	}
}

func WithIntegrationPermissionModes() IntegrationMockTransportOption {
	return func(t *integrationMockTransport) {
		t.testMessages = []claudecode.Message{
			&claudecode.AssistantMessage{
				Content: []claudecode.ContentBlock{&claudecode.TextBlock{Text: "Permission mode respected"}},
				Model:   "claude-3-5-sonnet-20241022",
			},
		}
	}
}

func WithIntegrationWorkingDirectory(workDir string) IntegrationMockTransportOption {
	return func(t *integrationMockTransport) {
		scenario := &integrationScenario{
			name:       "working_directory",
			workingDir: workDir,
			messages: []claudecode.Message{
				&claudecode.AssistantMessage{
					Content: []claudecode.ContentBlock{&claudecode.TextBlock{Text: fmt.Sprintf("Working in directory: %s", workDir)}},
					Model:   "claude-3-5-sonnet-20241022",
				},
			},
		}
		t.testScenarios["working_directory"] = scenario
		t.testMessages = scenario.messages
	}
}

func WithIntegrationErrorScenarios() IntegrationMockTransportOption {
	return func(t *integrationMockTransport) {
		t.testMessages = loadIntegrationFixture(&testing.T{}, "error_responses")
		t.connectError = fmt.Errorf("integration test connection error")
	}
}

func WithIntegrationLargeResponse(sizeBytes int) IntegrationMockTransportOption {
	return func(t *integrationMockTransport) {
		// Create a large response message
		largeText := strings.Repeat("This is a large response test. ", sizeBytes/32)
		t.testMessages = []claudecode.Message{
			&claudecode.AssistantMessage{
				Content: []claudecode.ContentBlock{&claudecode.TextBlock{Text: largeText}},
				Model:   "claude-3-5-sonnet-20241022",
			},
		}
	}
}

func WithIntegrationConcurrentScenario(clientCount int) IntegrationMockTransportOption {
	return func(t *integrationMockTransport) {
		messages := make([]claudecode.Message, clientCount)
		for i := 0; i < clientCount; i++ {
			messages[i] = &claudecode.AssistantMessage{
				Content: []claudecode.ContentBlock{&claudecode.TextBlock{Text: fmt.Sprintf("Concurrent response %d", i+1)}},
				Model:   "claude-3-5-sonnet-20241022",
			}
		}
		t.testMessages = messages
	}
}

func WithIntegrationResourceTracking() IntegrationMockTransportOption {
	return func(t *integrationMockTransport) {
		t.resourceTracker = &integrationResourceTracker{
			goroutines:        0,  // Start clean for proper tracking
			allocatedMemoryMB: 10, // Mock starting memory
			openFiles:         5,  // Mock starting files
			connections:       1,  // Mock starting connections
		}
	}
}

func WithIntegrationPerformanceScenario(messageCount int) IntegrationMockTransportOption {
	return func(t *integrationMockTransport) {
		messages := make([]claudecode.Message, messageCount)
		for i := 0; i < messageCount; i++ {
			messages[i] = &claudecode.AssistantMessage{
				Content: []claudecode.ContentBlock{&claudecode.TextBlock{Text: fmt.Sprintf("Performance test message %d", i+1)}},
				Model:   "claude-3-5-sonnet-20241022",
			}
		}
		t.testMessages = messages
	}
}

func WithIntegrationStressScenario(messageCount, concurrency int) IntegrationMockTransportOption {
	return func(t *integrationMockTransport) {
		messages := make([]claudecode.Message, messageCount)
		for i := 0; i < messageCount; i++ {
			messages[i] = &claudecode.AssistantMessage{
				Content: []claudecode.ContentBlock{&claudecode.TextBlock{Text: fmt.Sprintf("Stress test message %d", i+1)}},
				Model:   "claude-3-5-sonnet-20241022",
			}
		}
		t.testMessages = messages

		// Initialize resource tracking for stress test
		t.resourceTracker = &integrationResourceTracker{
			goroutines: concurrency,
		}
	}
}

func WithIntegrationCLIVersions(versions ...string) IntegrationMockTransportOption {
	return func(t *integrationMockTransport) {
		messages := make([]claudecode.Message, len(versions))
		for i, version := range versions {
			messages[i] = &claudecode.AssistantMessage{
				Content: []claudecode.ContentBlock{&claudecode.TextBlock{Text: fmt.Sprintf("CLI version %s compatible", version)}},
				Model:   "claude-3-5-sonnet-20241022",
			}
		}
		t.testMessages = messages
	}
}

func WithIntegrationPlatformScenarios() IntegrationMockTransportOption {
	return func(t *integrationMockTransport) {
		platform := runtime.GOOS
		t.testMessages = []claudecode.Message{
			&claudecode.AssistantMessage{
				Content: []claudecode.ContentBlock{&claudecode.TextBlock{Text: fmt.Sprintf("Platform %s supported", platform)}},
				Model:   "claude-3-5-sonnet-20241022",
			},
		}
	}
}

func WithIntegrationNetworkIsolation() IntegrationMockTransportOption {
	return func(t *integrationMockTransport) {
		t.testMessages = []claudecode.Message{
			&claudecode.AssistantMessage{
				Content: []claudecode.ContentBlock{&claudecode.TextBlock{Text: "Operating in offline mode"}},
				Model:   "claude-3-5-sonnet-20241022",
			},
		}
	}
}

// Resource tracking helper methods
func (r *integrationResourceTracker) trackGoroutine() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.goroutines++
}

func (r *integrationResourceTracker) releaseGoroutine() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.goroutines > 0 {
		r.goroutines--
	}
}

func (r *integrationResourceTracker) trackConnection() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.connections++
}

func (r *integrationResourceTracker) releaseConnection() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.connections > 0 {
		r.connections--
	}
}

// Resource cleanup verification helpers following client_test.go patterns
func setupResourceCleanupTest(t *testing.T) (*integrationMockTransport, int) {
	t.Helper()

	initialGoroutines := runtime.NumGoroutine()
	transport := newIntegrationMockTransport(WithIntegrationResourceTracking())
	return transport, initialGoroutines
}

func assertResourcesReleased(t *testing.T, transport *integrationMockTransport, initialGoroutines, expectedGoroutines int) {
	t.Helper()

	// Give minimal time for cleanup - reduce artificial delays
	runtime.GC()
	time.Sleep(10 * time.Millisecond)

	finalGoroutines := runtime.NumGoroutine()

	// Check actual goroutine count
	if finalGoroutines > initialGoroutines+expectedGoroutines {
		t.Errorf("Goroutine leak detected: initial=%d, final=%d, expected_max=%d",
			initialGoroutines, finalGoroutines, initialGoroutines+expectedGoroutines)
	}

	// Verify mock transport state
	if transport.resourceTracker != nil {
		transport.resourceTracker.mu.Lock()
		goroutines := transport.resourceTracker.goroutines
		transport.resourceTracker.mu.Unlock()

		if goroutines > expectedGoroutines {
			t.Errorf("Mock transport goroutine leak: expected max %d, got %d", expectedGoroutines, goroutines)
		}
	}
}

func setupTransportStateTest(t *testing.T) *integrationMockTransport {
	t.Helper()
	return newIntegrationMockTransport()
}

func assertTransportStateTransition(t *testing.T, transport *integrationMockTransport, expectedConnected, expectedClosed bool) {
	t.Helper()

	transport.mu.Lock()
	connected := transport.connected
	closed := transport.closed
	transport.mu.Unlock()

	if connected != expectedConnected {
		t.Errorf("Expected transport connected=%v, got connected=%v", expectedConnected, connected)
	}

	if closed != expectedClosed {
		t.Errorf("Expected transport closed=%v, got closed=%v", expectedClosed, closed)
	}
}

// Helper function to create bool pointers
func boolPtr(b bool) *bool {
	return &b
}

func floatPtr(f float64) *float64 {
	return &f
}
