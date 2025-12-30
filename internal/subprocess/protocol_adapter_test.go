package subprocess

import (
	"context"
	"sync"
	"testing"
	"time"
)

// TestProtocolAdapterWrite tests that the adapter writes to stdin correctly.
func TestProtocolAdapterWrite(t *testing.T) {
	t.Helper()

	tests := []struct {
		name      string
		input     []byte
		wantErr   bool
		errString string
	}{
		{
			name:    "writes data successfully",
			input:   []byte(`{"type":"control_request","request_id":"req_1_abc123"}`),
			wantErr: false,
		},
		{
			name:    "writes empty data",
			input:   []byte{},
			wantErr: false,
		},
		{
			name:    "handles large data",
			input:   make([]byte, 1024*1024), // 1MB
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock writer
			var written []byte
			var mu sync.Mutex
			mockWriter := &mockStdinWriter{
				writeFn: func(data []byte) (int, error) {
					mu.Lock()
					defer mu.Unlock()
					written = append(written, data...)
					return len(data), nil
				},
			}

			adapter := NewProtocolAdapter(mockWriter)

			ctx := context.Background()
			err := adapter.Write(ctx, tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("Write() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && len(written) != len(tt.input) {
				t.Errorf("Write() wrote %d bytes, want %d", len(written), len(tt.input))
			}
		})
	}
}

// TestProtocolAdapterWriteContextCancellation tests that Write respects context cancellation.
func TestProtocolAdapterWriteContextCancellation(t *testing.T) {
	mockWriter := &mockStdinWriter{
		writeFn: func(data []byte) (int, error) {
			time.Sleep(100 * time.Millisecond)
			return len(data), nil
		},
	}

	adapter := NewProtocolAdapter(mockWriter)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := adapter.Write(ctx, []byte("test"))

	if err == nil {
		t.Error("Write() should return error when context is cancelled")
	}
}

// TestProtocolAdapterRead tests that Read returns a channel (required by interface but not used).
func TestProtocolAdapterRead(t *testing.T) {
	adapter := NewProtocolAdapter(&mockStdinWriter{})

	ctx := context.Background()
	ch := adapter.Read(ctx)

	if ch == nil {
		t.Error("Read() should return a non-nil channel")
	}

	// Channel should be closed (we don't use the read loop)
	select {
	case _, ok := <-ch:
		if ok {
			t.Error("Read() channel should be closed")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Read() channel should return immediately when closed")
	}
}

// TestProtocolAdapterClose tests cleanup.
func TestProtocolAdapterClose(t *testing.T) {
	adapter := NewProtocolAdapter(&mockStdinWriter{})

	err := adapter.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}

	// Multiple closes should be safe
	err = adapter.Close()
	if err != nil {
		t.Errorf("Close() second call error = %v", err)
	}
}

// TestProtocolAdapterThreadSafety tests concurrent Write operations.
func TestProtocolAdapterThreadSafety(t *testing.T) {
	var writeCount int
	var mu sync.Mutex
	mockWriter := &mockStdinWriter{
		writeFn: func(data []byte) (int, error) {
			mu.Lock()
			writeCount++
			mu.Unlock()
			return len(data), nil
		},
	}

	adapter := NewProtocolAdapter(mockWriter)
	ctx := context.Background()

	const goroutines = 10
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			_ = adapter.Write(ctx, []byte("test"))
		}()
	}

	wg.Wait()

	mu.Lock()
	defer mu.Unlock()
	if writeCount != goroutines {
		t.Errorf("Expected %d writes, got %d", goroutines, writeCount)
	}
}

// mockStdinWriter implements io.Writer for testing
type mockStdinWriter struct {
	writeFn func([]byte) (int, error)
}

func (m *mockStdinWriter) Write(p []byte) (int, error) {
	if m.writeFn != nil {
		return m.writeFn(p)
	}
	return len(p), nil
}
