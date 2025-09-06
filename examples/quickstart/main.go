// Package main demonstrates basic usage of the Claude Code SDK for Go.
//
// This example shows how to:
// - Execute a simple query with Claude Code
// - Handle different message types from the response
// - Properly iterate through streaming messages
// - Clean up resources
package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/severity1/claude-code-sdk-go"
)

func main() {
	fmt.Println("Claude Code SDK for Go - Quickstart Example")
	fmt.Println("==========================================")

	// Create context with reasonable timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Basic query example
	fmt.Println("🤖 Asking Claude: What is 2+2?")
	
	iterator, err := claudecode.Query(ctx, "What is 2+2?")
	if err != nil {
		log.Fatalf("Failed to create query: %v", err)
	}
	defer iterator.Close()

	fmt.Println("\n📥 Response:")
	
	// Iterate through all messages
	for {
		message, err := iterator.Next(ctx)
		if err != nil {
			if err.Error() == "no more messages" {
				break // Normal completion
			}
			log.Fatalf("Failed to get next message: %v", err)
		}

		if message == nil {
			break // Stream ended
		}

		// Handle different message types
		switch msg := message.(type) {
		case *claudecode.AssistantMessage:
			// This is Claude's response - extract text content
			for _, block := range msg.Content {
				switch b := block.(type) {
				case *claudecode.TextBlock:
					fmt.Printf("🎯 Answer: %s\n", b.Text)
				case *claudecode.ThinkingBlock:
					fmt.Printf("💭 Claude is thinking: %s\n", b.Thinking)
				}
			}
		case *claudecode.SystemMessage:
			// System initialization message (can be ignored for basic usage)
			fmt.Println("⚙️  System initialized")
		case *claudecode.ResultMessage:
			// Final result with metadata
			if msg.IsError {
				fmt.Printf("❌ Error: %s\n", msg.Result)
			} else {
				fmt.Printf("✅ Completed successfully\n")
			}
		default:
			fmt.Printf("📦 Other message type: %T\n", message)
		}
	}

	fmt.Println("\n🎉 Query completed!")
	
	// Example with options
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("📋 Example with system prompt:")
	
	iterator2, err := claudecode.Query(ctx, "Hello there!", 
		claudecode.WithSystemPrompt("You are a friendly assistant. Keep responses brief."))
	if err != nil {
		log.Fatalf("Failed to create query with options: %v", err)
	}
	defer iterator2.Close()

	fmt.Println("\n📥 Response:")
	
	for {
		message, err := iterator2.Next(ctx)
		if err != nil {
			if err.Error() == "no more messages" {
				break
			}
			log.Fatalf("Failed to get next message: %v", err)
		}

		if message == nil {
			break
		}

		// Extract just the text response for cleaner output
		if assistantMsg, ok := message.(*claudecode.AssistantMessage); ok {
			for _, block := range assistantMsg.Content {
				if textBlock, ok := block.(*claudecode.TextBlock); ok {
					fmt.Printf("🎯 Response: %s\n", textBlock.Text)
				}
			}
		}
	}

	fmt.Println("\n✨ All examples completed successfully!")
}