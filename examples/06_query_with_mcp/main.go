// Package main demonstrates Query API with MCP tools (timezone operations).
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
	fmt.Println("Getting current time in multiple timezones using MCP time server")

	ctx := context.Background()
	query := "What time is it in Tokyo and New York?"

	fmt.Printf("\nQuery: %s\n", query)
	fmt.Println("Tools: MCP time server")

	// Configure MCP time server using uvx
	servers := map[string]claudecode.McpServerConfig{
		"time": &claudecode.McpStdioServerConfig{
			Type:    claudecode.McpServerTypeStdio,
			Command: "uvx",
			Args:    []string{"mcp-server-time"},
		},
	}

	// Query with MCP time tools enabled
	iterator, err := claudecode.Query(ctx, query,
		claudecode.WithMcpServers(servers),
		claudecode.WithAllowedTools("mcp__time__get_current_time"),
		claudecode.WithSystemPrompt("You are a helpful assistant. Use the MCP time server to get current time in different timezones."),
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
								fmt.Printf("Time Tool Error: %s\n", content)
							} else if len(content) > 150 {
								fmt.Printf("Time Result: %s...\n", content[:150])
							} else {
								fmt.Printf("Time Result: %s\n", content)
							}
						}
					}
				}
			}
		case *claudecode.ResultMessage:
			if msg.IsError {
				if msg.Result != nil {
					fmt.Printf("Error: %s\n", *msg.Result)
				} else {
					fmt.Printf("Error: unknown error\n")
				}
			}
		}
	}

	fmt.Println("\nTimezone query completed!")
}
