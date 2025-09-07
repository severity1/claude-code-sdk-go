// Package main demonstrates basic usage of the Claude Code SDK Query API.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/severity1/claude-code-sdk-go"
)

func main() {
	fmt.Println("Claude Code SDK - Query API Example")
	fmt.Println("Asking: What is 2+2?")

	ctx := context.Background()

	// Create and execute query
	iterator, err := claudecode.Query(ctx, "What is 2+2?")
	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}
	defer iterator.Close()

	fmt.Println("\nResponse:")

	// Iterate through messages
	for {
		message, err := iterator.Next(ctx)
		if err != nil {
			if errors.Is(err, claudecode.ErrNoMoreMessages) {
				break
			}
			log.Fatalf("Failed to get message: %v", err)
		}

		if message == nil {
			break
		}

		// Handle different message types
		switch msg := message.(type) {
		case *claudecode.AssistantMessage:
			for _, block := range msg.Content {
				if textBlock, ok := block.(*claudecode.TextBlock); ok {
					fmt.Print(textBlock.Text)
				}
			}
		case *claudecode.ResultMessage:
			if msg.IsError {
				log.Printf("Error: %s", msg.Result)
			}
		}
	}

	fmt.Println("\nQuery completed!")
}
