// Package subprocess provides the subprocess transport implementation for Claude Code CLI.
package subprocess

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/severity1/claude-code-sdk-go/internal/cli"
	"github.com/severity1/claude-code-sdk-go/internal/parser"
	"github.com/severity1/claude-code-sdk-go/internal/shared"
)

const (
	// channelBufferSize is the buffer size for message and error channels.
	channelBufferSize = 10
	// terminationTimeoutSeconds is the timeout for graceful process termination.
	terminationTimeoutSeconds = 5
	// windowsOS is the GOOS value for Windows platform.
	windowsOS = "windows"
)

// Transport implements the Transport interface using subprocess communication.
type Transport struct {
	// Process management
	cmd        *exec.Cmd
	cliPath    string
	options    *shared.Options
	closeStdin bool
	promptArg  *string // For one-shot queries, prompt passed as CLI argument
	entrypoint string  // CLAUDE_CODE_ENTRYPOINT value (sdk-go or sdk-go-client)

	// Connection state
	connected bool
	mu        sync.RWMutex

	// I/O streams
	stdin  io.WriteCloser
	stdout io.ReadCloser
	stderr *os.File

	// Message parsing
	parser *parser.Parser

	// Channels for communication
	msgChan chan shared.Message
	errChan chan error

	// Control and cleanup
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// New creates a new subprocess transport.
func New(cliPath string, options *shared.Options, closeStdin bool, entrypoint string) *Transport {
	return &Transport{
		cliPath:    cliPath,
		options:    options,
		closeStdin: closeStdin,
		entrypoint: entrypoint,
		parser:     parser.New(),
	}
}

// NewWithPrompt creates a new subprocess transport for one-shot queries with prompt as CLI argument.
func NewWithPrompt(cliPath string, options *shared.Options, prompt string) *Transport {
	return &Transport{
		cliPath:    cliPath,
		options:    options,
		closeStdin: true,
		entrypoint: "sdk-go", // Query mode uses sdk-go
		parser:     parser.New(),
		promptArg:  &prompt,
	}
}

// IsConnected returns whether the transport is currently connected.
func (t *Transport) IsConnected() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.connected && t.cmd != nil && t.cmd.Process != nil
}

// Connect starts the Claude CLI subprocess.
func (t *Transport) Connect(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.connected {
		return fmt.Errorf("transport already connected")
	}

	// Build command with all options
	var args []string
	if t.promptArg != nil {
		// One-shot query with prompt as CLI argument
		args = cli.BuildCommandWithPrompt(t.cliPath, t.options, *t.promptArg)
	} else {
		// Streaming mode or regular one-shot
		args = cli.BuildCommand(t.cliPath, t.options, t.closeStdin)
	}
	//nolint:gosec // G204: This is the core CLI SDK functionality - subprocess execution is required
	t.cmd = exec.CommandContext(ctx, args[0], args[1:]...)

	// Set up environment
	t.cmd.Env = append(os.Environ(), "CLAUDE_CODE_ENTRYPOINT="+t.entrypoint)

	// Set working directory if specified
	if t.options != nil && t.options.Cwd != nil {
		if err := cli.ValidateWorkingDirectory(*t.options.Cwd); err != nil {
			return err
		}
		t.cmd.Dir = *t.options.Cwd
	}

	// Set up I/O pipes
	var err error
	if t.promptArg == nil {
		// Only create stdin pipe if we need to send messages via stdin
		t.stdin, err = t.cmd.StdinPipe()
		if err != nil {
			return fmt.Errorf("failed to create stdin pipe: %w", err)
		}
	}

	t.stdout, err = t.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	// Isolate stderr using temporary file to prevent deadlocks
	// This matches Python SDK pattern to avoid subprocess pipe deadlocks
	t.stderr, err = os.CreateTemp("", "claude_stderr_*.log")
	if err != nil {
		return fmt.Errorf("failed to create stderr file: %w", err)
	}
	t.cmd.Stderr = t.stderr

	// Start the process
	if err := t.cmd.Start(); err != nil {
		t.cleanup()
		return shared.NewConnectionError(
			fmt.Sprintf("failed to start Claude CLI: %v", err),
			err,
		)
	}

	// Set up context for goroutine management
	t.ctx, t.cancel = context.WithCancel(ctx)

	// Initialize channels
	t.msgChan = make(chan shared.Message, channelBufferSize)
	t.errChan = make(chan error, channelBufferSize)

	// Start I/O handling goroutines
	t.wg.Add(1)
	go t.handleStdout()

	// Note: Do NOT close stdin here for one-shot mode
	// The CLI still needs stdin to receive the message, even with --print flag
	// stdin will be closed after sending the message in SendMessage()

	t.connected = true
	return nil
}

// SendMessage sends a message to the CLI subprocess.
func (t *Transport) SendMessage(ctx context.Context, message shared.StreamMessage) error {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// For one-shot queries with promptArg, the prompt is already passed as CLI argument
	// so we don't need to send any messages via stdin
	if t.promptArg != nil {
		return nil // No-op for one-shot queries
	}

	if !t.connected || t.stdin == nil {
		return fmt.Errorf("transport not connected or stdin closed")
	}

	// Check context cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Serialize message to JSON
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Send with newline
	_, err = t.stdin.Write(append(data, '\n'))
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	// For one-shot mode, close stdin after sending the message
	if t.closeStdin {
		_ = t.stdin.Close()
		t.stdin = nil
	}

	return nil
}

// ReceiveMessages returns channels for receiving messages and errors.
func (t *Transport) ReceiveMessages(_ context.Context) (<-chan shared.Message, <-chan error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if !t.connected {
		// Return closed channels if not connected
		msgChan := make(chan shared.Message)
		errChan := make(chan error)
		close(msgChan)
		close(errChan)
		return msgChan, errChan
	}

	return t.msgChan, t.errChan
}

// Interrupt sends an interrupt signal to the subprocess.
func (t *Transport) Interrupt(_ context.Context) error {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if !t.connected || t.cmd == nil || t.cmd.Process == nil {
		return fmt.Errorf("process not running")
	}

	// Windows doesn't support os.Interrupt signal
	if runtime.GOOS == windowsOS {
		return fmt.Errorf("interrupt not supported by windows")
	}

	// Send interrupt signal (Unix/Linux/macOS)
	return t.cmd.Process.Signal(os.Interrupt)
}

// Close terminates the subprocess connection.
func (t *Transport) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.connected {
		return nil // Already closed
	}

	t.connected = false

	// Cancel context to stop goroutines
	if t.cancel != nil {
		t.cancel()
	}

	// Close stdin if open
	if t.stdin != nil {
		_ = t.stdin.Close()
		t.stdin = nil
	}

	// Wait for goroutines to finish with timeout
	done := make(chan struct{})
	go func() {
		t.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Goroutines finished gracefully
	case <-time.After(terminationTimeoutSeconds * time.Second):
		// Timeout: proceed with cleanup anyway
		// Goroutines should terminate when process is killed
	}

	// Terminate process with 5-second timeout
	var err error
	if t.cmd != nil && t.cmd.Process != nil {
		err = t.terminateProcess()
	}

	// Cleanup resources
	t.cleanup()

	return err
}

// handleStdout processes stdout in a separate goroutine
func (t *Transport) handleStdout() {
	defer t.wg.Done()
	defer close(t.msgChan)
	defer close(t.errChan)

	scanner := bufio.NewScanner(t.stdout)

	for scanner.Scan() {
		select {
		case <-t.ctx.Done():
			return
		default:
		}

		line := scanner.Text()
		if line == "" {
			continue
		}

		// Parse line with the parser
		messages, err := t.parser.ProcessLine(line)
		if err != nil {
			select {
			case t.errChan <- err:
			case <-t.ctx.Done():
				return
			}
			continue
		}

		// Send parsed messages
		for _, msg := range messages {
			if msg != nil {
				select {
				case t.msgChan <- msg:
				case <-t.ctx.Done():
					return
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		select {
		case t.errChan <- fmt.Errorf("stdout scanner error: %w", err):
		case <-t.ctx.Done():
		}
	}
}

// isProcessAlreadyFinishedError checks if an error indicates the process has already terminated.
// This follows the Python SDK pattern of suppressing "process not found" type errors.
func isProcessAlreadyFinishedError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "process already finished") ||
		strings.Contains(errStr, "process already released") ||
		strings.Contains(errStr, "no child processes") ||
		strings.Contains(errStr, "signal: killed")
}

// terminateProcess implements the 5-second SIGTERM → SIGKILL sequence
func (t *Transport) terminateProcess() error {
	if t.cmd == nil || t.cmd.Process == nil {
		return nil
	}

	// Send SIGTERM
	if err := t.cmd.Process.Signal(syscall.SIGTERM); err != nil {
		// If process is already finished, that's success
		if isProcessAlreadyFinishedError(err) {
			return nil
		}
		// If SIGTERM fails for other reasons, try SIGKILL immediately
		killErr := t.cmd.Process.Kill()
		if killErr != nil && !isProcessAlreadyFinishedError(killErr) {
			return killErr
		}
		return nil // Don't return error for expected termination
	}

	// Wait exactly 5 seconds
	done := make(chan error, 1)
	// Capture cmd while we know it's valid to avoid data race
	cmd := t.cmd
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case err := <-done:
		// Normal termination or expected signals are not errors
		if err != nil {
			// Check if it's an expected exit signal
			if strings.Contains(err.Error(), "signal:") {
				return nil // Expected signal termination
			}
		}
		return err
	case <-time.After(terminationTimeoutSeconds * time.Second):
		// Force kill after 5 seconds
		if killErr := t.cmd.Process.Kill(); killErr != nil && !isProcessAlreadyFinishedError(killErr) {
			return killErr
		}
		// Wait for process to exit after kill
		<-done
		return nil
	case <-t.ctx.Done():
		// Context canceled - force kill immediately
		if killErr := t.cmd.Process.Kill(); killErr != nil && !isProcessAlreadyFinishedError(killErr) {
			return killErr
		}
		// Wait for process to exit after kill, but don't return context error
		// since this is normal cleanup behavior
		<-done
		return nil
	}
}

// cleanup cleans up all resources
func (t *Transport) cleanup() {
	if t.stdout != nil {
		_ = t.stdout.Close()
		t.stdout = nil
	}

	if t.stderr != nil {
		// Graceful cleanup matching Python SDK pattern
		// Python: except Exception: pass
		_ = t.stderr.Close()
		_ = os.Remove(t.stderr.Name()) // Ignore cleanup errors
		t.stderr = nil
	}

	// Reset state
	t.cmd = nil
}
