// Package main demonstrates streaming with Client API using automatic resource management.
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/severity1/claude-agent-sdk-go"
)

func main() {
	fmt.Println("Claude Agent SDK - Client Streaming Example")
	fmt.Println("Asking: Explain Go goroutines with a simple example")

	ctx := context.Background()
	question := "Explain what Go goroutines are and show a simple example"

	// WithClient handles connection lifecycle automatically
	err := claudecode.WithClient(ctx, func(client claudecode.Client) error {
		fmt.Println("\nConnected! Streaming response:")

		if err := client.Query(ctx, question); err != nil {
			return fmt.Errorf("query failed: %w", err)
		}

		// Stream messages in real-time
		msgChan := client.ReceiveMessages(ctx)
		for {
			select {
			case message := <-msgChan:
				if message == nil {
					return nil // Stream ended
				}

				switch msg := message.(type) {
				case *claudecode.AssistantMessage:
					// Print streaming text as it arrives
					for _, block := range msg.Content {
						if textBlock, ok := block.(*claudecode.TextBlock); ok {
							fmt.Print(textBlock.Text)
						}
					}
				case *claudecode.ResultMessage:
					if msg.IsError {
						if msg.Result != nil {
							return fmt.Errorf("error: %s", *msg.Result)
						}
						return fmt.Errorf("error: unknown error")
					}
					return nil // Success, stream complete
				}
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	})
	if err != nil {
		if cliErr := claudecode.AsCLINotFoundError(err); cliErr != nil {
			fmt.Printf("Claude CLI not found: %v\n", cliErr)
			fmt.Println("Install with: npm install -g @anthropic-ai/claude-code")
			return
		}
		if connErr := claudecode.AsConnectionError(err); connErr != nil {
			fmt.Printf("Connection failed: %v\n", connErr)
			return
		}
		log.Fatalf("Streaming failed: %v", err)
	}

	fmt.Println("\n\nStreaming completed!")
}
