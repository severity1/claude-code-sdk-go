package interfaces

// Message represents any message type in the Claude Code protocol.
type Message interface {
	Type() string
}

// ContentBlock represents any content block within a message.
// Uses consistent Type() method naming instead of BlockType().
type ContentBlock interface {
	Type() string
}
