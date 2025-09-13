package interfaces

// MessageContent represents any content that can be used in a message.
// This is a sealed interface - external packages cannot implement it.
type MessageContent interface {
	messageContent() // unexported method seals the interface
}

// UserMessageContent represents content that can be used in user messages.
// This is a sealed interface that embeds MessageContent.
type UserMessageContent interface {
	MessageContent
	userMessageContent() // unexported method seals the interface
}

// AssistantMessageContent represents content that can be used in assistant messages.
// This is a sealed interface that embeds MessageContent.
type AssistantMessageContent interface {
	MessageContent
	assistantMessageContent() // unexported method seals the interface
}

// TextContent represents simple text content that can be used in both user and assistant messages.
type TextContent struct {
	Text string `json:"text"`
}

// messageContent implements the sealed MessageContent interface.
func (TextContent) messageContent() {}

// userMessageContent implements the sealed UserMessageContent interface.
func (TextContent) userMessageContent() {}

// BlockListContent represents content that contains multiple content blocks, used in user messages.
type BlockListContent struct {
	Blocks []ContentBlock `json:"blocks"`
}

// messageContent implements the sealed MessageContent interface.
func (BlockListContent) messageContent() {}

// userMessageContent implements the sealed UserMessageContent interface.
func (BlockListContent) userMessageContent() {}

// ThinkingContent represents thinking content with signature, used in assistant messages.
type ThinkingContent struct {
	Thinking  string `json:"thinking"`
	Signature string `json:"signature"`
}

// messageContent implements the sealed MessageContent interface.
func (ThinkingContent) messageContent() {}

// assistantMessageContent implements the sealed AssistantMessageContent interface.
func (ThinkingContent) assistantMessageContent() {}
