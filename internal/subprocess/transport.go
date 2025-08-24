// Package subprocess provides the subprocess transport implementation for Claude Code CLI.
package subprocess

import (
	"context"
	"os/exec"
)

// Transport implements the Transport interface using subprocess communication.
type Transport struct {
	cmd        *exec.Cmd
	cliPath    string
	options    interface{} // TODO: Import proper Options type
	closeStdin bool
	connected  bool
}

// New creates a new subprocess transport.
func New(cliPath string, options interface{}, closeStdin bool) *Transport {
	return &Transport{
		cliPath:    cliPath,
		options:    options,
		closeStdin: closeStdin,
	}
}

// Connect starts the Claude CLI subprocess.
func (t *Transport) Connect(ctx context.Context) error {
	// TODO: Implement subprocess connection logic
	t.connected = true
	return nil
}

// SendMessage sends a message to the CLI subprocess.
func (t *Transport) SendMessage(ctx context.Context, message interface{}) error {
	// TODO: Implement message sending logic
	return nil
}

// ReceiveMessages returns channels for receiving messages and errors.
func (t *Transport) ReceiveMessages(ctx context.Context) (<-chan interface{}, <-chan error) {
	// TODO: Implement message receiving logic
	msgChan := make(chan interface{})
	errChan := make(chan error)
	close(msgChan)
	close(errChan)
	return msgChan, errChan
}

// Interrupt sends an interrupt signal to the subprocess.
func (t *Transport) Interrupt(ctx context.Context) error {
	// TODO: Implement interrupt logic
	return nil
}

// Close terminates the subprocess connection.
func (t *Transport) Close() error {
	// TODO: Implement cleanup logic
	t.connected = false
	return nil
}
