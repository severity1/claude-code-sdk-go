// Package main demonstrates Partial Streaming for real-time updates.
//
// This example shows how to receive streaming updates as Claude generates
// responses, enabling real-time UI updates and progress indicators.
// Partial streaming enables:
// - Character-by-character or chunk-by-chunk text display
// - Real-time typing indicators and progress bars
// - Early processing of response content
// - Building responsive chat interfaces
//
// Key components:
// - WithPartialStreaming: Enable partial message streaming
// - WithIncludePartialMessages: Explicit control over partial messages
// - StreamEvent: Message type for streaming events
// - StreamEventType* constants: Event type identifiers
//
// Stream event types:
// - content_block_start: A new content block is starting
// - content_block_delta: Incremental content update
// - content_block_stop: Content block completed
// - message_start: Message generation beginning
// - message_delta: Message-level update
// - message_stop: Message generation complete
//
// Run: go run main.go
package main

import (
	"context"
	"fmt"
	"time"

	claudecode "github.com/severity1/claude-agent-sdk-go"
)

func main() {
	fmt.Println("Claude Agent SDK - Partial Streaming Example")
	fmt.Println("=============================================")
	fmt.Println()

	// Example 1: Stream Event Types
	fmt.Println("--- Example 1: Stream Event Types ---")
	fmt.Println("Available stream event type constants:")
	showStreamEventTypes()

	// Example 2: Basic Partial Streaming
	fmt.Println()
	fmt.Println("--- Example 2: Basic Partial Streaming ---")
	fmt.Println("Enabling partial streaming to receive real-time updates...")
	runPartialStreamingExample()

	// Example 3: Typing Indicator Pattern
	fmt.Println()
	fmt.Println("--- Example 3: Typing Indicator Pattern ---")
	fmt.Println("Building a typing indicator with stream events...")
	demonstrateTypingIndicator()

	fmt.Println()
	fmt.Println("Partial streaming example completed!")
}

// showStreamEventTypes displays all available stream event type constants
func showStreamEventTypes() {
	fmt.Printf("  ContentBlockStart: %q\n", claudecode.StreamEventTypeContentBlockStart)
	fmt.Printf("  ContentBlockDelta: %q\n", claudecode.StreamEventTypeContentBlockDelta)
	fmt.Printf("  ContentBlockStop:  %q\n", claudecode.StreamEventTypeContentBlockStop)
	fmt.Printf("  MessageStart:      %q\n", claudecode.StreamEventTypeMessageStart)
	fmt.Printf("  MessageDelta:      %q\n", claudecode.StreamEventTypeMessageDelta)
	fmt.Printf("  MessageStop:       %q\n", claudecode.StreamEventTypeMessageStop)
}

// runPartialStreamingExample demonstrates receiving StreamEvent messages
func runPartialStreamingExample() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	var eventCount int
	var deltaCount int
	var textBuffer string

	err := claudecode.WithClient(ctx, func(client claudecode.Client) error {
		// Ask for a short response to see streaming in action
		if err := client.Query(ctx, "Count from 1 to 5, putting each number on its own line."); err != nil {
			return err
		}

		msgChan := client.ReceiveMessages(ctx)

		for {
			select {
			case message := <-msgChan:
				if message == nil {
					return nil
				}

				switch msg := message.(type) {
				case *claudecode.StreamEvent:
					eventCount++
					eventType, _ := msg.Event["type"].(string)

					switch eventType {
					case claudecode.StreamEventTypeContentBlockStart:
						fmt.Printf("[%s] Content block starting...\n", eventType)

					case claudecode.StreamEventTypeContentBlockDelta:
						deltaCount++
						// Extract delta text from the event
						if delta, ok := msg.Event["delta"].(map[string]any); ok {
							if text, ok := delta["text"].(string); ok {
								textBuffer += text
								// Show each delta as it arrives
								fmt.Printf("[delta %d] %q\n", deltaCount, text)
							}
						}

					case claudecode.StreamEventTypeContentBlockStop:
						fmt.Printf("[%s] Content block complete\n", eventType)

					case claudecode.StreamEventTypeMessageStart:
						fmt.Printf("[%s] Message generation starting\n", eventType)

					case claudecode.StreamEventTypeMessageStop:
						fmt.Printf("[%s] Message generation complete\n", eventType)

					default:
						fmt.Printf("[%s] Event received\n", eventType)
					}

				case *claudecode.AssistantMessage:
					// With partial streaming, we also get the full message
					for _, block := range msg.Content {
						if textBlock, ok := block.(*claudecode.TextBlock); ok {
							fmt.Printf("\nFull message: %s\n", textBlock.Text)
						}
					}

				case *claudecode.ResultMessage:
					if msg.IsError {
						if msg.Result != nil {
							return fmt.Errorf("error: %s", *msg.Result)
						}
						return fmt.Errorf("error: unknown error")
					}
					return nil
				}

			case <-ctx.Done():
				return ctx.Err()
			}
		}
	},
		claudecode.WithPartialStreaming(),
		claudecode.WithMaxTurns(1),
	)

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("\nSummary:\n")
	fmt.Printf("  Total stream events: %d\n", eventCount)
	fmt.Printf("  Content deltas: %d\n", deltaCount)
	fmt.Printf("  Accumulated text length: %d chars\n", len(textBuffer))
}

// demonstrateTypingIndicator shows how to build a typing indicator
func demonstrateTypingIndicator() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	fmt.Println("Simulating a chat interface with typing indicator...")
	fmt.Println()

	isTyping := false
	var charCount int

	err := claudecode.WithClient(ctx, func(client claudecode.Client) error {
		if err := client.Query(ctx, "Say hello in exactly 10 words."); err != nil {
			return err
		}

		msgChan := client.ReceiveMessages(ctx)

		for {
			select {
			case message := <-msgChan:
				if message == nil {
					return nil
				}

				switch msg := message.(type) {
				case *claudecode.StreamEvent:
					eventType, _ := msg.Event["type"].(string)

					switch eventType {
					case claudecode.StreamEventTypeContentBlockStart:
						if !isTyping {
							isTyping = true
							fmt.Print("Claude is typing")
						}

					case claudecode.StreamEventTypeContentBlockDelta:
						// Show typing indicator dots
						charCount++
						if charCount%5 == 0 {
							fmt.Print(".")
						}

					case claudecode.StreamEventTypeMessageStop:
						if isTyping {
							fmt.Println(" done!")
							isTyping = false
						}
					}

				case *claudecode.AssistantMessage:
					fmt.Print("\nClaude: ")
					for _, block := range msg.Content {
						if textBlock, ok := block.(*claudecode.TextBlock); ok {
							fmt.Println(textBlock.Text)
						}
					}

				case *claudecode.ResultMessage:
					if msg.IsError {
						if msg.Result != nil {
							return fmt.Errorf("error: %s", *msg.Result)
						}
						return fmt.Errorf("error: unknown error")
					}
					return nil
				}

			case <-ctx.Done():
				return ctx.Err()
			}
		}
	},
		claudecode.WithPartialStreaming(),
		claudecode.WithMaxTurns(1),
	)

	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}
