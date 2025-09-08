// Package main demonstrates advanced Client API features with WithClient and error handling.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/severity1/claude-code-sdk-go"
)

func main() {
	fmt.Println("Claude Code SDK - Advanced Client Features Example")
	fmt.Println("WithClient with custom options and error handling")

	ctx := context.Background()

	// Advanced query with custom system prompt and error handling
	err := claudecode.WithClient(ctx, func(client claudecode.Client) error {
		fmt.Println("\nConnected with custom configuration!")

		questions := []string{
			"Explain Go concurrency patterns for web crawlers with goroutine management",
			"Review this Go code for race conditions: func processItems(items []Item) error { var wg sync.WaitGroup; for _, item := range items { go func() { processItem(item) }() } }",
		}

		for i, question := range questions {
			fmt.Printf("\n--- Question %d ---\n", i+1)
			fmt.Printf("Q: %s\n", question)

			if err := client.Query(ctx, question); err != nil {
				return fmt.Errorf("query %d failed: %w", i+1, err)
			}

			if err := streamResponse(ctx, client); err != nil {
				return fmt.Errorf("response %d failed: %w", i+1, err)
			}
		}

		fmt.Println("\nAdvanced session completed!")
		return nil
	},
		// Advanced configuration options
		claudecode.WithSystemPrompt("You are a senior Go developer providing code reviews and architectural guidance."),
		claudecode.WithAllowedTools("Read", "Write"), // Optional tools
	)
	// Advanced error handling
	if err != nil {
		// Check for specific error types
		var cliError *claudecode.CLINotFoundError
		if errors.As(err, &cliError) {
			fmt.Printf("❌ Claude CLI not installed: %v\n", cliError)
			fmt.Println("Install: npm install -g @anthropic-ai/claude-code")
			return
		}

		var connError *claudecode.ConnectionError
		if errors.As(err, &connError) {
			fmt.Printf("⚠️ Connection failed: %v\n", connError)
			fmt.Println("WithClient handled cleanup automatically")
			return
		}

		log.Fatalf("Advanced features failed: %v", err)
	}

	fmt.Println("\nAdvanced features demonstration completed!")
}

func streamResponse(ctx context.Context, client claudecode.Client) error {
	fmt.Println("\nResponse:")

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
}
