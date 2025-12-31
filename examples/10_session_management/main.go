// Package main demonstrates session management with the Claude Code SDK for Go.
// This example shows the clean Query API that replaces the variadic sessionID parameter
// with explicit methods for better clarity and Go idiom compliance.
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/severity1/claude-agent-sdk-go"
)

func main() {
	fmt.Println("Claude Code SDK - Session Management Example")
	fmt.Println("============================================")

	ctx := context.Background()

	if err := runExample(ctx); err != nil {
		log.Fatalf("Example failed: %v", err)
	}
}

func runExample(ctx context.Context) error {
	return claudecode.WithClient(ctx, func(client claudecode.Client) error {
		fmt.Println("\n1. Default session (recommended for most use cases)")
		if err := client.Query(ctx, "Hello! What's 2+2?"); err != nil {
			return fmt.Errorf("default session query: %w", err)
		}
		drainResponse(ctx, client)

		fmt.Println("\n2. Custom session (for isolated conversations)")
		if err := client.QueryWithSession(ctx, "Hello! What's 3+3?", "math-session"); err != nil {
			return fmt.Errorf("custom session query: %w", err)
		}
		drainResponse(ctx, client)

		fmt.Println("\n3. Session isolation demonstration")
		fmt.Println("   Default session asking about previous question:")
		if err := client.Query(ctx, "What was my previous math question?"); err != nil {
			return fmt.Errorf("isolation test: %w", err)
		}
		drainResponse(ctx, client)

		fmt.Println("   Math session asking about previous question:")
		if err := client.QueryWithSession(ctx, "What was my previous math question?", "math-session"); err != nil {
			return fmt.Errorf("isolation test: %w", err)
		}
		drainResponse(ctx, client)

		return nil
	})
}

// drainResponse consumes all messages from the client until completion.
// This is a simple implementation for demonstration purposes.
func drainResponse(ctx context.Context, client claudecode.Client) {
	messages := client.ReceiveMessages(ctx)

	for msg := range messages {
		if msg == nil {
			break
		}

		switch message := msg.(type) {
		case *claudecode.AssistantMessage:
			// Print the first text block content
			for _, block := range message.Content {
				if textBlock, ok := block.(*claudecode.TextBlock); ok {
					fmt.Printf("   → %s\n", textBlock.Text)
					break // Only show first text block for brevity
				}
			}
		case *claudecode.ResultMessage:
			if message.IsError {
				fmt.Printf("   ✗ Error: %s\n", message.Result)
			}
			return // Stream complete
		}
	}
}
