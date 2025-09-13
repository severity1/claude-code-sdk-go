package claudecode_test

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/severity1/claude-code-sdk-go"
	"github.com/severity1/claude-code-sdk-go/pkg/interfaces"
)

// TestInterfaceCompatibilitySimple tests basic interface compatibility without dependencies
func TestInterfaceCompatibilitySimple(t *testing.T) {
	// Create a simple mock transport that works
	transport := &simpleMockTransport{}
	client := claudecode.NewClientWithTransport(transport)

	// Test that ClientImpl implements all segregated interfaces using reflection
	tests := []struct {
		name          string
		interfaceType reflect.Type
	}{
		{
			name:          "ConnectionManager",
			interfaceType: reflect.TypeOf((*interfaces.ConnectionManager)(nil)).Elem(),
		},
		{
			name:          "QueryExecutor",
			interfaceType: reflect.TypeOf((*interfaces.QueryExecutor)(nil)).Elem(),
		},
		{
			name:          "MessageReceiver",
			interfaceType: reflect.TypeOf((*interfaces.MessageReceiver)(nil)).Elem(),
		},
		{
			name:          "ProcessController",
			interfaceType: reflect.TypeOf((*interfaces.ProcessController)(nil)).Elem(),
		},
		{
			name:          "Client (composition)",
			interfaceType: reflect.TypeOf((*interfaces.Client)(nil)).Elem(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clientType := reflect.TypeOf(client)
			if !clientType.Implements(tt.interfaceType) {
				t.Errorf("ClientImpl does not implement %s interface", tt.interfaceType.Name())
				return
			}
			t.Logf("✅ ClientImpl successfully implements %s interface", tt.interfaceType.Name())
		})
	}

	// Test casting to segregated interfaces
	t.Run("Interface casting", func(t *testing.T) {
		// Test that we can cast client to each segregated interface
		if _, ok := client.(interfaces.ConnectionManager); !ok {
			t.Error("Client cannot be cast to ConnectionManager interface")
		}

		if _, ok := client.(interfaces.QueryExecutor); !ok {
			t.Error("Client cannot be cast to QueryExecutor interface")
		}

		if _, ok := client.(interfaces.MessageReceiver); !ok {
			t.Error("Client cannot be cast to MessageReceiver interface")
		}

		if _, ok := client.(interfaces.ProcessController); !ok {
			t.Error("Client cannot be cast to ProcessController interface")
		}

		if _, ok := client.(interfaces.Client); !ok {
			t.Error("Client cannot be cast to interfaces.Client")
		}

		t.Log("✅ All interface casting successful")
	})

	// Test specialized interface patterns
	t.Run("Specialized interfaces", func(t *testing.T) {
		// Test SimpleQuerier pattern
		if _, ok := client.(interfaces.SimpleQuerier); !ok {
			t.Error("Client cannot be cast to SimpleQuerier interface")
		}

		// Test StreamClient pattern
		if _, ok := client.(interfaces.StreamClient); !ok {
			t.Error("Client cannot be cast to StreamClient interface")
		}

		t.Log("✅ All specialized interface casting successful")
	})
}

// TestMethodSignatureCompatibility tests that the method signatures match exactly
func TestMethodSignatureCompatibility(t *testing.T) {
	transport := &simpleMockTransport{}
	client := claudecode.NewClientWithTransport(transport)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	t.Run("ConnectionManager methods", func(t *testing.T) {
		connMgr := client.(interfaces.ConnectionManager)

		// Test IsConnected before connection
		if connMgr.IsConnected() {
			t.Error("Expected IsConnected() to return false before connection")
		}

		// Test Connect with correct signature
		if err := connMgr.Connect(ctx); err != nil {
			t.Errorf("Connect failed: %v", err)
		}

		// Test IsConnected after connection
		if !connMgr.IsConnected() {
			t.Error("Expected IsConnected() to return true after connection")
		}

		// Test Close
		if err := connMgr.Close(); err != nil {
			t.Errorf("Close failed: %v", err)
		}

		t.Log("✅ ConnectionManager method signatures work correctly")
	})

	t.Run("QueryExecutor methods", func(t *testing.T) {
		queryExec := client.(interfaces.QueryExecutor)

		// Reconnect for query test
		if err := client.Connect(ctx); err != nil {
			t.Fatalf("Failed to reconnect: %v", err)
		}

		// Test Query with simplified signature (no sessionID)
		if err := queryExec.Query(ctx, "Test query"); err != nil {
			t.Errorf("Query failed: %v", err)
		}

		_ = client.Close()
		t.Log("✅ QueryExecutor method signatures work correctly")
	})
}

// Simple mock transport for testing interface compatibility
type simpleMockTransport struct {
	connected bool
}

func (s *simpleMockTransport) Connect(ctx context.Context) error {
	s.connected = true
	return nil
}

func (s *simpleMockTransport) SendMessage(ctx context.Context, message claudecode.StreamMessage) error {
	return nil
}

func (s *simpleMockTransport) ReceiveMessages(ctx context.Context) (<-chan claudecode.Message, <-chan error) {
	msgChan := make(chan claudecode.Message)
	errChan := make(chan error)
	close(msgChan)
	close(errChan)
	return msgChan, errChan
}

func (s *simpleMockTransport) Interrupt(ctx context.Context) error {
	return nil
}

func (s *simpleMockTransport) Close() error {
	s.connected = false
	return nil
}
