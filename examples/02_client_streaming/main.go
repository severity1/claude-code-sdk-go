// Package main demonstrates streaming usage of the Claude Code SDK Client API.
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
	fmt.Println("Claude Code SDK for Go - Client Streaming Example")
	fmt.Println("================================================")

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	fmt.Println("🔗 Creating streaming client...")
	client := claudecode.NewClient()

	fmt.Println("📡 Connecting to Claude Code...")
	if err := client.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect client: %v", err)
	}

	defer func() {
		fmt.Println("\n🧹 Cleaning up connection...")
		if err := client.Disconnect(); err != nil {
			log.Printf("Warning: Failed to disconnect cleanly: %v", err)
		}
		fmt.Println("✅ Connection closed")
	}()

	fmt.Println("✅ Connected successfully!")

	question := "Explain what Go goroutines are and show a simple example"
	fmt.Printf("\n🤖 Asking Claude: %s\n", question)

	if err := client.Query(ctx, question); err != nil {
		log.Fatalf("Failed to send query: %v", err)
	}

	fmt.Println("\n📥 Streaming response:")
	fmt.Println(strings.Repeat("-", 50))

	responseReceived := false
	msgChan := client.ReceiveMessages(ctx)
	for {
		select {
		case message := <-msgChan:
			if message == nil {
				fmt.Println("\n📡 Stream ended")
				goto streamDone
			}

			switch msg := message.(type) {
			case *claudecode.AssistantMessage:
				responseReceived = true
				for _, block := range msg.Content {
					switch b := block.(type) {
					case *claudecode.TextBlock:
						fmt.Print(b.Text)
					case *claudecode.ThinkingBlock:
						fmt.Printf("\n💭 [Thinking: %s]\n", b.Thinking)
					}
				}
			case *claudecode.UserMessage:
				if blocks, ok := msg.Content.([]claudecode.ContentBlock); ok {
					for _, block := range blocks {
						if textBlock, ok := block.(*claudecode.TextBlock); ok {
							fmt.Printf("📤 User: %s\n", textBlock.Text)
						}
					}
				} else if contentStr, ok := msg.Content.(string); ok {
					fmt.Printf("📤 User: %s\n", contentStr)
				}
			case *claudecode.SystemMessage:
				fmt.Println("⚙️  System initialized")
			case *claudecode.ResultMessage:
				if msg.IsError {
					fmt.Printf("\n❌ Issue: %s\n", msg.Result)
				} else {
					fmt.Printf("\n✅ Stream completed successfully\n")
				}
			default:
				fmt.Printf("\n📦 Received message type: %T\n", message)
			}
		case <-ctx.Done():
			fmt.Println("\n⏰ Context timeout")
			goto streamDone
		}
	}

streamDone:

	if !responseReceived {
		fmt.Println("⚠️  No response received - check your Claude Code installation")
	} else {
		fmt.Println(strings.Repeat("-", 50))
		fmt.Println("\n🎉 Streaming example completed!")
	}
}
