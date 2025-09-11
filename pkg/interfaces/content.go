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
