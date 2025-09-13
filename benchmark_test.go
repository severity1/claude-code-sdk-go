package claudecode

import (
	"context"
	"testing"
	"time"

	"github.com/severity1/claude-code-sdk-go/pkg/interfaces"
)

// BenchmarkTypedVsInterfaceEmpty compares performance of typed interfaces vs interface{} boxing.
// Tests INTERFACE_SPEC.md Requirement #6: Performance Maintenance.
func BenchmarkTypedVsInterfaceEmpty(b *testing.B) {
	// Test typed interface performance
	b.Run("typed_interface", func(b *testing.B) {
		var msg interfaces.Message
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			// Create typed message
			msg = &UserMessage{
				Content: &TextContent{Text: "benchmark test message"},
			}

			// Type-safe access
			_ = msg.Type()
		}

		// Prevent compiler optimization
		_ = msg
	})

	// Test interface{} boxing performance for comparison
	b.Run("interface_empty_boxing", func(b *testing.B) {
		var msg interface{}
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			// Create with interface{} boxing
			msg = map[string]interface{}{
				"type":    "user",
				"content": map[string]interface{}{"text": "benchmark test message"},
			}

			// Type assertion required
			if m, ok := msg.(map[string]interface{}); ok {
				_ = m["type"]
			}
		}

		// Prevent compiler optimization
		_ = msg
	})
}

// BenchmarkInterfaceComposition tests performance overhead of interface composition.
// Validates segregated interfaces don't introduce performance penalties.
func BenchmarkInterfaceComposition(b *testing.B) {
	ctx, cancel := setupBenchmarkContext(b, 10*time.Second)
	defer cancel()

	b.Run("segregated_interfaces", func(b *testing.B) {
		transport := newBenchmarkMockTransport()
		client := NewClientWithTransport(transport)
		defer disconnectBenchmarkClientSafely(b, client)

		connectBenchmarkClientSafely(b, ctx, client)

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			// Test focused interface usage
			var connectionMgr interfaces.ConnectionManager = client.(interfaces.ConnectionManager)
			var queryExec interfaces.QueryExecutor = client.(interfaces.QueryExecutor)
			var msgReceiver interfaces.MessageReceiver = client.(interfaces.MessageReceiver)

			// Use interfaces
			_ = connectionMgr.IsConnected()
			_ = queryExec.Query(ctx, "benchmark query")
			_ = msgReceiver.ReceiveMessages(ctx)
		}
	})

	b.Run("monolithic_interface", func(b *testing.B) {
		transport := newBenchmarkMockTransport()
		client := NewClientWithTransport(transport)
		defer disconnectBenchmarkClientSafely(b, client)

		connectBenchmarkClientSafely(b, ctx, client)

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			// Test full client interface usage
			_ = client.Status()
			_ = client.Query(ctx, "benchmark query")
			_ = client.ReceiveMessages(ctx)
		}
	})
}

// BenchmarkMessageIterator validates iterator performance with typed messages.
func BenchmarkMessageIterator(b *testing.B) {
	ctx, cancel := setupBenchmarkContext(b, 10*time.Second)
	defer cancel()

	transport := newBenchmarkMockTransport(WithBenchmarkMessages(1000))
	client := NewClientWithTransport(transport)
	defer disconnectBenchmarkClientSafely(b, client)

	connectBenchmarkClientSafely(b, ctx, client)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		iter := client.ReceiveResponse(ctx)

		// Iterate through messages
		messageCount := 0
		for {
			msg, err := iter.Next(ctx)
			if err != nil {
				break
			}
			if msg == nil {
				break
			}
			messageCount++
		}

		_ = iter.Close()
		_ = messageCount
	}
}

// BenchmarkClientOperations tests core client operations performance.
func BenchmarkClientOperations(b *testing.B) {
	ctx, cancel := setupBenchmarkContext(b, 30*time.Second)
	defer cancel()

	b.Run("connect_disconnect", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			transport := newBenchmarkMockTransport()
			client := NewClientWithTransport(transport)

			err := client.Connect(ctx)
			if err != nil {
				b.Fatalf("Connect failed: %v", err)
			}

			err = client.Close()
			if err != nil {
				b.Fatalf("Disconnect failed: %v", err)
			}
		}
	})

	b.Run("query_operations", func(b *testing.B) {
		transport := newBenchmarkMockTransport()
		client := NewClientWithTransport(transport)
		defer disconnectBenchmarkClientSafely(b, client)

		connectBenchmarkClientSafely(b, ctx, client)

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			err := client.Query(ctx, "benchmark query")
			if err != nil {
				b.Fatalf("Query failed: %v", err)
			}
		}
	})
}

// Helper Functions - following client_test.go patterns

// setupBenchmarkContext creates a context for benchmark tests
func setupBenchmarkContext(b *testing.B, timeout time.Duration) (context.Context, context.CancelFunc) {
	b.Helper()
	return context.WithTimeout(context.Background(), timeout)
}

// benchmarkMockTransport implements Transport interface for benchmarking
type benchmarkMockTransport struct {
	connected bool
	messages  []Message
}

// BenchmarkMockTransportOption allows configuration of benchmark mock transport
type BenchmarkMockTransportOption func(*benchmarkMockTransport)

// WithBenchmarkMessages configures the mock transport with test messages
func WithBenchmarkMessages(count int) BenchmarkMockTransportOption {
	return func(t *benchmarkMockTransport) {
		messages := make([]Message, count)
		for i := 0; i < count; i++ {
			messages[i] = &AssistantMessage{
				Content: []ContentBlock{&TextBlock{Text: "benchmark message"}},
				Model:   "claude-3-5-sonnet-20241022",
			}
		}
		t.messages = messages
	}
}

// newBenchmarkMockTransport creates a new benchmark mock transport
func newBenchmarkMockTransport(opts ...BenchmarkMockTransportOption) *benchmarkMockTransport {
	transport := &benchmarkMockTransport{
		messages: []Message{
			&AssistantMessage{
				Content: []ContentBlock{&TextBlock{Text: "default benchmark message"}},
				Model:   "claude-3-5-sonnet-20241022",
			},
		},
	}

	for _, opt := range opts {
		opt(transport)
	}

	return transport
}

func (t *benchmarkMockTransport) Connect(ctx context.Context) error {
	t.connected = true
	return nil
}

func (t *benchmarkMockTransport) SendMessage(ctx context.Context, message StreamMessage) error {
	return nil
}

func (t *benchmarkMockTransport) ReceiveMessages(ctx context.Context) (<-chan Message, <-chan error) {
	msgChan := make(chan Message, len(t.messages))
	errChan := make(chan error, 1)

	go func() {
		defer close(msgChan)
		defer close(errChan)

		for _, msg := range t.messages {
			select {
			case msgChan <- msg:
			case <-ctx.Done():
				errChan <- ctx.Err()
				return
			}
		}
	}()

	return msgChan, errChan
}

func (t *benchmarkMockTransport) Interrupt(ctx context.Context) error {
	return nil
}

func (t *benchmarkMockTransport) Close() error {
	t.connected = false
	return nil
}

// connectBenchmarkClientSafely connects a client for benchmark tests
func connectBenchmarkClientSafely(b *testing.B, ctx context.Context, client Client) {
	b.Helper()
	err := client.Connect(ctx)
	if err != nil {
		b.Fatalf("Failed to connect benchmark client: %v", err)
	}
}

// disconnectBenchmarkClientSafely disconnects a client for benchmark tests
func disconnectBenchmarkClientSafely(b *testing.B, client Client) {
	b.Helper()
	if err := client.Close(); err != nil {
		b.Logf("Warning: Failed to disconnect benchmark client: %v", err)
	}
}
