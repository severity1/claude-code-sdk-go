// Package subprocess provides the subprocess transport implementation for Claude Code CLI.
package subprocess

import (
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

	"github.com/severity1/claude-agent-sdk-go/internal/cli"
	"github.com/severity1/claude-agent-sdk-go/internal/control"
	"github.com/severity1/claude-agent-sdk-go/internal/parser"
	"github.com/severity1/claude-agent-sdk-go/internal/shared"
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
	stdin      io.WriteCloser
	stdout     io.ReadCloser
	stderr     *os.File      // Temporary file for stderr isolation
	stderrPipe io.ReadCloser // Pipe for callback-based stderr handling

	// Temporary files (cleaned up on Close)
	mcpConfigFile *os.File // Temporary MCP config file

	// Message parsing
	parser *parser.Parser

	// Stream validation
	validator *shared.StreamValidator

	// Channels for communication
	msgChan chan shared.Message
	errChan chan error

	// Control protocol (for streaming mode only)
	protocol        *control.Protocol
	protocolAdapter *ProtocolAdapter

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
		validator:  shared.NewStreamValidator(),
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
		validator:  shared.NewStreamValidator(),
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

	// Generate MCP config file if McpServers are specified
	opts := t.options
	if t.options != nil && len(t.options.McpServers) > 0 {
		mcpConfigPath, err := t.generateMcpConfigFile()
		if err != nil {
			return fmt.Errorf("failed to generate MCP config file: %w", err)
		}

		// Create modified options with mcp-config in ExtraArgs
		// We don't want to mutate the user's options, so create a shallow copy
		optsCopy := *t.options
		if optsCopy.ExtraArgs == nil {
			optsCopy.ExtraArgs = make(map[string]*string)
		} else {
			// Deep copy the ExtraArgs map
			extraArgsCopy := make(map[string]*string, len(optsCopy.ExtraArgs)+1)
			for k, v := range optsCopy.ExtraArgs {
				extraArgsCopy[k] = v
			}
			optsCopy.ExtraArgs = extraArgsCopy
		}
		optsCopy.ExtraArgs["mcp-config"] = &mcpConfigPath
		opts = &optsCopy
	}

	// Build command with all options
	var args []string
	if t.promptArg != nil {
		// One-shot query with prompt as CLI argument
		args = cli.BuildCommandWithPrompt(t.cliPath, opts, *t.promptArg)
	} else {
		// Streaming mode or regular one-shot
		args = cli.BuildCommand(t.cliPath, opts, t.closeStdin)
	}
	//nolint:gosec // G204: This is the core CLI SDK functionality - subprocess execution is required
	t.cmd = exec.CommandContext(ctx, args[0], args[1:]...)

	// Set up environment and apply to command
	t.cmd.Env = t.buildEnvironment()

	// Set working directory if specified
	if t.options != nil && t.options.Cwd != nil {
		if err := cli.ValidateWorkingDirectory(*t.options.Cwd); err != nil {
			return err
		}
		t.cmd.Dir = *t.options.Cwd
	}

	// Check CLI version and warn if outdated (non-blocking)
	if warning := cli.CheckCLIVersion(ctx, t.cliPath); warning != "" {
		if t.options != nil && t.options.StderrCallback != nil {
			t.options.StderrCallback(warning)
		}
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

	// Handle stderr configuration
	if err := t.setupStderr(); err != nil {
		return err
	}

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

	// Start stderr callback goroutine if callback is configured
	if t.stderrPipe != nil && t.options != nil && t.options.StderrCallback != nil {
		t.wg.Add(1)
		go t.handleStderrCallback()
	}

	// Note: Do NOT close stdin here for one-shot mode
	// The CLI still needs stdin to receive the message, even with --print flag
	// stdin will be closed after sending the message in SendMessage()

	// Set up control protocol for streaming mode only
	// One-shot mode (closeStdin=true) doesn't need control protocol
	if !t.closeStdin {
		t.protocolAdapter = NewProtocolAdapter(t.stdin)
		t.protocol = control.NewProtocol(t.protocolAdapter, t.buildProtocolOptions()...)

		// Start the protocol's background goroutine
		// Note: The protocol's readLoop will block on the closed channel from the adapter,
		// which is intentional - we route messages via handleStdout() -> HandleIncomingMessage()
		if err := t.protocol.Start(t.ctx); err != nil {
			t.cleanup()
			return fmt.Errorf("failed to start control protocol: %w", err)
		}

		// Perform control protocol handshake when hooks, permission callbacks,
		// file checkpointing, or SDK MCP servers are configured
		if t.options != nil && (t.options.Hooks != nil || t.options.CanUseTool != nil || t.options.EnableFileCheckpointing || t.hasSdkMcpServers()) {
			if _, err := t.protocol.Initialize(t.ctx); err != nil {
				t.cleanup()
				return fmt.Errorf("failed to initialize control protocol: %w", err)
			}
		}
	}

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

	// Close control protocol first (before cancelling context)
	if t.protocol != nil {
		_ = t.protocol.Close()
		t.protocol = nil
	}
	if t.protocolAdapter != nil {
		_ = t.protocolAdapter.Close()
		t.protocolAdapter = nil
	}

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

	if t.stderrPipe != nil {
		_ = t.stderrPipe.Close()
		t.stderrPipe = nil
	}

	if t.stderr != nil {
		// Graceful cleanup matching Python SDK pattern
		// Python: except Exception: pass
		_ = t.stderr.Close()
		_ = os.Remove(t.stderr.Name()) // Ignore cleanup errors
		t.stderr = nil
	}

	if t.mcpConfigFile != nil {
		// Clean up temporary MCP config file
		_ = t.mcpConfigFile.Close()
		_ = os.Remove(t.mcpConfigFile.Name()) // Ignore cleanup errors
		t.mcpConfigFile = nil
	}

	// Reset state
	t.cmd = nil
}

// generateMcpConfigFile creates a temporary MCP config file from options.McpServers.
// Returns the file path. The file is stored in t.mcpConfigFile for cleanup.
func (t *Transport) generateMcpConfigFile() (string, error) {
	// Build servers map, stripping Instance field from SDK servers for CLI serialization
	// The CLI doesn't need the Go instance - it routes mcp_message requests to the SDK
	serversForCLI := make(map[string]any)
	for name, config := range t.options.McpServers {
		if sdkConfig, ok := config.(*shared.McpSdkServerConfig); ok {
			// SDK servers: only send type and name to CLI
			serversForCLI[name] = map[string]any{
				"type": string(sdkConfig.Type),
				"name": sdkConfig.Name,
			}
		} else {
			// External servers: pass as-is
			serversForCLI[name] = config
		}
	}

	// Create the MCP config structure matching Claude CLI expected format
	mcpConfig := map[string]interface{}{
		"mcpServers": serversForCLI,
	}

	// Marshal to JSON
	configData, err := json.MarshalIndent(mcpConfig, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal MCP config: %w", err)
	}

	// Create temporary file
	tmpFile, err := os.CreateTemp("", "claude_mcp_config_*.json")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}

	// Write config data
	if _, err := tmpFile.Write(configData); err != nil {
		_ = tmpFile.Close()
		_ = os.Remove(tmpFile.Name())
		return "", fmt.Errorf("failed to write MCP config: %w", err)
	}

	// Sync to ensure data is written
	if err := tmpFile.Sync(); err != nil {
		_ = tmpFile.Close()
		_ = os.Remove(tmpFile.Name())
		return "", fmt.Errorf("failed to sync MCP config file: %w", err)
	}

	// Store for cleanup later
	t.mcpConfigFile = tmpFile

	return tmpFile.Name(), nil
}

// GetValidator returns the stream validator for diagnostic purposes.
// This allows clients to check for validation issues like missing tool results.
func (t *Transport) GetValidator() *shared.StreamValidator {
	return t.validator
}

// SetModel changes the AI model during a streaming session.
// This method requires control protocol integration which is only available
// in streaming mode (when closeStdin is false).
func (t *Transport) SetModel(ctx context.Context, model *string) error {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if !t.connected {
		return fmt.Errorf("transport not connected")
	}

	// Control protocol integration is only available in streaming mode
	if t.closeStdin {
		return fmt.Errorf("SetModel not available in one-shot mode")
	}

	// Delegate to control protocol
	if t.protocol == nil {
		return fmt.Errorf("control protocol not initialized")
	}

	return t.protocol.SetModel(ctx, model)
}

// SetPermissionMode changes the permission mode during a streaming session.
// This method requires control protocol integration which is only available
// in streaming mode (when closeStdin is false).
func (t *Transport) SetPermissionMode(ctx context.Context, mode string) error {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if !t.connected {
		return fmt.Errorf("transport not connected")
	}

	// Control protocol integration is only available in streaming mode
	if t.closeStdin {
		return fmt.Errorf("SetPermissionMode not available in one-shot mode")
	}

	// Delegate to control protocol
	if t.protocol == nil {
		return fmt.Errorf("control protocol not initialized")
	}

	return t.protocol.SetPermissionMode(ctx, mode)
}

// RewindFiles reverts tracked files to their state at a specific user message.
// This method requires control protocol integration which is only available
// in streaming mode (when closeStdin is false).
// Returns error if not connected, not in streaming mode, or protocol not initialized.
func (t *Transport) RewindFiles(ctx context.Context, userMessageID string) error {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if !t.connected {
		return fmt.Errorf("transport not connected")
	}

	// Control protocol integration is only available in streaming mode
	if t.closeStdin {
		return fmt.Errorf("RewindFiles not available in one-shot mode")
	}

	// Delegate to control protocol
	if t.protocol == nil {
		return fmt.Errorf("control protocol not initialized")
	}

	return t.protocol.RewindFiles(ctx, userMessageID)
}

// buildProtocolOptions constructs control protocol options from transport configuration.
// This extracts callback wiring logic from Connect to reduce cyclomatic complexity.
func (t *Transport) buildProtocolOptions() []control.ProtocolOption {
	var opts []control.ProtocolOption

	// Wire permission callback if configured
	if t.options != nil && t.options.CanUseTool != nil {
		// Create adapter that converts between shared.Options (any types)
		// and control package (strongly-typed) to avoid import cycles
		optionsCallback := t.options.CanUseTool
		opts = append(opts,
			control.WithCanUseToolCallback(func(
				ctx context.Context,
				toolName string,
				input map[string]any,
				permCtx control.ToolPermissionContext,
			) (control.PermissionResult, error) {
				// Call the Options callback with any-typed permCtx
				result, err := optionsCallback(ctx, toolName, input, permCtx)
				if err != nil {
					return nil, err
				}

				// Convert result back to strongly-typed PermissionResult
				if pr, ok := result.(control.PermissionResult); ok {
					return pr, nil
				}

				// Fallback: deny if result type is unexpected
				return control.NewPermissionResultDeny("invalid permission result type"), nil
			}))
	}

	// Wire hooks if configured
	if t.options != nil && t.options.Hooks != nil {
		// Convert from any to strongly-typed hooks map
		if hooks, ok := t.options.Hooks.(map[control.HookEvent][]control.HookMatcher); ok {
			opts = append(opts, control.WithHooks(hooks))
		}
	}

	// Wire SDK MCP servers to protocol (Issue #7)
	if t.options != nil && len(t.options.McpServers) > 0 {
		sdkServers := make(map[string]control.McpServer)
		for name, config := range t.options.McpServers {
			if sdkConfig, ok := config.(*shared.McpSdkServerConfig); ok && sdkConfig.Instance != nil {
				sdkServers[name] = sdkConfig.Instance
			}
		}
		if len(sdkServers) > 0 {
			opts = append(opts, control.WithSdkMcpServers(sdkServers))
		}
	}

	return opts
}

// hasSdkMcpServers checks if any SDK MCP servers are configured.
// Returns true if at least one SDK server with a valid Instance exists.
func (t *Transport) hasSdkMcpServers() bool {
	if t.options == nil || len(t.options.McpServers) == 0 {
		return false
	}
	for _, config := range t.options.McpServers {
		if sdkConfig, ok := config.(*shared.McpSdkServerConfig); ok && sdkConfig.Instance != nil {
			return true
		}
	}
	return false
}

// buildEnvironment constructs the environment variables for the subprocess.
// This extracts environment setup logic from Connect to reduce cyclomatic complexity.
func (t *Transport) buildEnvironment() []string {
	env := os.Environ()

	// Set entrypoint to identify SDK to CLI
	env = append(env, "CLAUDE_CODE_ENTRYPOINT="+t.entrypoint)

	// Enable file checkpointing if requested (matches Python SDK)
	if t.options != nil && t.options.EnableFileCheckpointing {
		env = append(env, "CLAUDE_CODE_ENABLE_SDK_FILE_CHECKPOINTING=true")
	}

	// Add user-specified environment variables
	if t.options != nil && t.options.ExtraEnv != nil {
		for key, value := range t.options.ExtraEnv {
			env = append(env, fmt.Sprintf("%s=%s", key, value))
		}
	}

	return env
}
