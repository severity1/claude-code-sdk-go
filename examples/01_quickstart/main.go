// Package main demonstrates basic usage of the Claude Code SDK for Go.
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/severity1/claude-code-sdk-go"
)

func main() {
	fmt.Println("Claude Code SDK for Go - Quickstart Example")
	fmt.Println("==========================================")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fmt.Println("🤖 Asking Claude: What is 2+2?")
	
	iterator, err := claudecode.Query(ctx, "What is 2+2?")
	if err != nil {
		log.Fatalf("Failed to create query: %v", err)
	}
	defer iterator.Close()

	fmt.Println("\n📥 Response:")
	
	for {
		message, err := iterator.Next(ctx)
		if err != nil {
			if err.Error() == "no more messages" {
				break
			}
			log.Fatalf("Failed to get next message: %v", err)
		}

		if message == nil {
			break
		}

		switch msg := message.(type) {
		case *claudecode.AssistantMessage:
			for _, block := range msg.Content {
				switch b := block.(type) {
				case *claudecode.TextBlock:
					fmt.Printf("🎯 Answer: %s\n", b.Text)
				case *claudecode.ThinkingBlock:
					fmt.Printf("💭 Claude is thinking: %s\n", b.Thinking)
				}
			}
		case *claudecode.SystemMessage:
			fmt.Println("⚙️  System initialized")
		case *claudecode.ResultMessage:
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
}