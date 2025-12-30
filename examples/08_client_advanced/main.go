// Package main demonstrates advanced Client API features with WithClient,
// dynamic model switching, and error handling.
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
	fmt.Println("WithClient with dynamic model switching and error handling")

	ctx := context.Background()

	// Advanced query with custom system prompt and error handling
	err := claudecode.WithClient(ctx, func(client claudecode.Client) error {
		fmt.Println("\nConnected with custom configuration!")

		// First question with default model
		fmt.Println("\n--- Question 1 (default model) ---")
		question1 := "Explain Go concurrency patterns for web crawlers with goroutine management"
		fmt.Printf("Q: %s\n", question1)

		if err := client.Query(ctx, question1); err != nil {
			return fmt.Errorf("query 1 failed: %w", err)
		}
		if err := streamResponse(ctx, client); err != nil {
			return fmt.Errorf("response 1 failed: %w", err)
		}

		// Dynamic model switch mid-conversation
		fmt.Println("\n--- Switching model to claude-sonnet-4-5 ---")
		sonnetModel := "claude-sonnet-4-5"
		if err := client.SetModel(ctx, &sonnetModel); err != nil {
			// Model switch is best-effort - log but continue
			fmt.Printf("Note: Model switch failed (may not be supported): %v\n", err)
		} else {
			fmt.Println("Model switched successfully!")
		}

		// Second question with new model
		fmt.Println("\n--- Question 2 (after model switch) ---")
		question2 := "Review this Go code for race conditions: func processItems(items []Item) error { var wg sync.WaitGroup; for _, item := range items { go func() { processItem(item) }() } }"
		fmt.Printf("Q: %s\n", question2)

		if err := client.Query(ctx, question2); err != nil {
			return fmt.Errorf("query 2 failed: %w", err)
		}
		if err := streamResponse(ctx, client); err != nil {
			return fmt.Errorf("response 2 failed: %w", err)
		}

		// Reset to default model
		fmt.Println("\n--- Resetting to default model ---")
		if err := client.SetModel(ctx, nil); err != nil {
			fmt.Printf("Note: Model reset failed: %v\n", err)
		} else {
			fmt.Println("Model reset to default!")
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
