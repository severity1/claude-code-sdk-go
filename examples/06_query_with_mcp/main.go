// Package main demonstrates Query API with MCP tools (AWS operations).
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/severity1/claude-agent-sdk-go"
)

func main() {
	fmt.Println("Claude Code SDK - Query API with MCP Tools Example")
	fmt.Println("AWS S3 bucket listing using MCP tools")

	ctx := context.Background()
	query := "List my S3 buckets with their names, creation dates, and regions"

	fmt.Printf("\nQuery: %s\n", query)
	fmt.Println("Tools: AWS MCP tools")

	// Query with MCP tools enabled
	iterator, err := claudecode.Query(ctx, query,
		claudecode.WithAllowedTools(
			"mcp__aws-api-mcp__call_aws",
			"mcp__aws-api-mcp__suggest_aws_commands"),
		claudecode.WithSystemPrompt("You are an AWS expert. Use AWS MCP tools to help with S3 operations."),
	)
	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}
	defer iterator.Close()

	fmt.Println("\nResponse:")

	for {
		message, err := iterator.Next(ctx)
		if err != nil {
			if errors.Is(err, claudecode.ErrNoMoreMessages) {
				break
			}
			log.Printf("Message error: %v", err)
			break
		}

		if message == nil {
			break
		}

		switch msg := message.(type) {
		case *claudecode.AssistantMessage:
			for _, block := range msg.Content {
				if textBlock, ok := block.(*claudecode.TextBlock); ok {
					fmt.Print(textBlock.Text)
				}
			}
		case *claudecode.UserMessage:
			if blocks, ok := msg.Content.([]claudecode.ContentBlock); ok {
				for _, block := range blocks {
					if toolResult, ok := block.(*claudecode.ToolResultBlock); ok {
						if content, ok := toolResult.Content.(string); ok {
							if strings.Contains(content, "tool_use_error") {
								fmt.Printf("⚠️ AWS Tool Error: %s\n", content)
							} else if len(content) > 150 {
								fmt.Printf("🔧 AWS Result: %s...\n", content[:150])
							} else {
								fmt.Printf("🔧 AWS Result: %s\n", content)
							}
						}
					}
				}
			}
		case *claudecode.ResultMessage:
			if msg.IsError {
				fmt.Printf("Error: %s\n", msg.Result)
			}
		}
	}

	fmt.Println("\nS3 bucket listing completed!")
}
