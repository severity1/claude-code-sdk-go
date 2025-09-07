// Package main demonstrates the differences between Query API and Client API patterns.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/severity1/claude-code-sdk-go"
)

func main() {
	fmt.Println("Claude Code SDK - Query vs Client API Comparison")
	fmt.Println("Demonstrating when to use each API pattern")

	ctx := context.Background()
	question := "What are the key differences between channels and mutexes in Go?"

	fmt.Printf("\nComparing both APIs with question: %s\n", question)

	// Demonstrate Query API
	fmt.Println("\n--- Query API Example ---")
	fmt.Println("Best for: One-shot operations, automation, CI/CD")

	if err := queryExample(ctx, question); err != nil {
		log.Printf("Query API failed: %v", err)
	}

	// Demonstrate Client API
	fmt.Println("\n--- Client API Example ---")
	fmt.Println("Best for: Multi-turn conversations, interactive sessions")

	if err := clientExample(ctx, question); err != nil {
		log.Printf("Client API failed: %v", err)
	}

	// Show Client API advantage
	fmt.Println("\n--- Client API Advantage: Context Preservation ---")
	if err := conversationExample(ctx); err != nil {
		log.Printf("Conversation failed: %v", err)
	}

	fmt.Println("\n🎯 API Selection Guide:")
	fmt.Println("Query API: Stateless, simple, automated scripts")
	fmt.Println("Client API: Stateful, conversations, interactive apps")
}

func queryExample(ctx context.Context, question string) error {
	fmt.Println("Creating one-shot query...")

	iterator, err := claudecode.Query(ctx, question)
	if err != nil {
		return fmt.Errorf("query failed: %w", err)
	}
	defer iterator.Close()

	fmt.Println("Response:")
	for {
		message, err := iterator.Next(ctx)
		if err != nil {
			if errors.Is(err, claudecode.ErrNoMoreMessages) {
				break
			}
			return fmt.Errorf("failed to get next message: %w", err)
		}

		if message == nil {
			break
		}

		if assistantMsg, ok := message.(*claudecode.AssistantMessage); ok {
			for _, block := range assistantMsg.Content {
				if textBlock, ok := block.(*claudecode.TextBlock); ok {
					fmt.Print(textBlock.Text)
				}
			}
		}
	}

	fmt.Println("✅ Query completed - automatic cleanup")
	return nil
}

func clientExample(ctx context.Context, question string) error {
	fmt.Println("Using WithClient for automatic resource management...")

	return claudecode.WithClient(ctx, func(client claudecode.Client) error {
		fmt.Println("Connected! Sending query...")

		if err := client.Query(ctx, question); err != nil {
			return fmt.Errorf("query failed: %w", err)
		}

		fmt.Println("Response:")
		msgChan := client.ReceiveMessages(ctx)
		for {
			select {
			case message := <-msgChan:
				if message == nil {
					return nil
				}

				switch msg := message.(type) {
				case *claudecode.AssistantMessage:
					for _, block := range msg.Content {
						if textBlock, ok := block.(*claudecode.TextBlock); ok {
							fmt.Print(textBlock.Text)
						}
					}
				case *claudecode.ResultMessage:
					if msg.IsError {
						return fmt.Errorf("error: %s", msg.Result)
					}
					return nil
				}
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	})
}

func conversationExample(ctx context.Context) error {
	fmt.Println("Multi-turn conversation showing context preservation...")

	questions := []string{
		"What is dependency injection?",
		"Show me a Go example of the pattern you just described",
		"What are the testing benefits of that approach?",
	}

	return claudecode.WithClient(ctx, func(client claudecode.Client) error {
		for i, question := range questions {
			fmt.Printf("\nTurn %d: %s\n", i+1, question)

			if err := client.Query(ctx, question); err != nil {
				return fmt.Errorf("turn %d failed: %w", i+1, err)
			}

			// Show first few lines of response
			if err := showFirstLines(ctx, client, 3); err != nil {
				return fmt.Errorf("turn %d display failed: %w", i+1, err)
			}

			// Drain remaining messages
			drainMessages(client.ReceiveMessages(ctx))
		}

		fmt.Println("\n✅ Context preserved across all turns!")
		fmt.Println("Each question built on the previous response automatically.")
		return nil
	})
}

// showFirstLines displays the first few lines of response from the client
func showFirstLines(ctx context.Context, client claudecode.Client, maxLines int) error {
	msgChan := client.ReceiveMessages(ctx)
	lines := 0

	for lines < maxLines {
		select {
		case message := <-msgChan:
			if message == nil {
				return nil
			}

			if assistantMsg, ok := message.(*claudecode.AssistantMessage); ok {
				for _, block := range assistantMsg.Content {
					if textBlock, ok := block.(*claudecode.TextBlock); ok {
						text := textBlock.Text
						if len(text) > 100 {
							fmt.Printf("  %s...\n", text[:100])
						} else {
							fmt.Printf("  %s\n", text)
						}
						lines++
						if lines >= maxLines {
							return nil
						}
					}
				}
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

// drainMessages consumes remaining messages from a channel
func drainMessages(msgChan <-chan claudecode.Message) {
	for {
		select {
		case message := <-msgChan:
			if message == nil {
				return
			}
		default:
			return
		}
	}
}
