// Package main demonstrates using Claude Code SDK Query API with MCP tools (AWS API).
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/severity1/claude-code-sdk-go"
)

func main() {
	fmt.Println("Claude Code SDK for Go - Query API with AWS MCP Tools Example")
	fmt.Println("=============================================================")

	if !checkAWSCredentials() {
		fmt.Println("⚠️  AWS credentials not found.")
		fmt.Println("Please configure AWS credentials using one of these methods:")
		fmt.Println("  • Environment variables: AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY")
		fmt.Println("  • AWS credentials file: ~/.aws/credentials")
		fmt.Println("  • AWS CLI: `aws configure`")
		fmt.Println("\n📚 This example will show how the code works with mock explanations.")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()

	fmt.Println("\n🪣 Listing S3 Buckets")
	fmt.Println("====================")

	query := `Please list all my S3 buckets and show me their names, creation dates, and regions.`

	fmt.Printf("🔧 Allowed tools: [mcp__aws-api-mcp__call_aws, mcp__aws-api-mcp__suggest_aws_commands]\n")
	fmt.Printf("❓ Query: %s\n", query)
	fmt.Println("--------------------------------------------------")

	iterator, err := claudecode.Query(ctx, query,
		claudecode.WithAllowedTools(
			"mcp__aws-api-mcp__call_aws", 
			"mcp__aws-api-mcp__suggest_aws_commands"),
		claudecode.WithSystemPrompt("You are an AWS expert. Help list and analyze S3 buckets using AWS MCP tools."),
	)
	if err != nil {
		log.Fatalf("Failed to create query: %v", err)
	}
	defer iterator.Close()

	fmt.Println("🤖 Claude's Response:")

	for {
		message, err := iterator.Next(ctx)
		if err != nil {
			if err.Error() == "no more messages" {
				break
			}
			fmt.Printf("\n⚠️  Error getting next message: %v\n", err)
			fmt.Println("This may be due to MCP tool permission timeout or other issues.")
			break
		}

		if message == nil {
			break
		}

		switch msg := message.(type) {
		case *claudecode.AssistantMessage:
			for _, block := range msg.Content {
				switch b := block.(type) {
				case *claudecode.TextBlock:
					fmt.Print(b.Text)
				case *claudecode.ThinkingBlock:
					fmt.Printf("\n💭 [AWS Analysis: %s]\n", b.Thinking)
				}
			}
		case *claudecode.UserMessage:
			if blocks, ok := msg.Content.([]claudecode.ContentBlock); ok {
				for _, block := range blocks {
					switch b := block.(type) {
					case *claudecode.TextBlock:
						fmt.Printf("📤 AWS Tool: %s\n", b.Text)
					case *claudecode.ToolResultBlock:
						if contentStr, ok := b.Content.(string); ok {
							displayContent := contentStr
							if strings.Contains(contentStr, "<tool_use_error>") {
								displayContent = strings.ReplaceAll(contentStr, "<tool_use_error>", "⚠️ ")
								displayContent = strings.ReplaceAll(displayContent, "</tool_use_error>", "")
								fmt.Printf("🔧 AWS Tool Issue (id: %s): %s\n", b.ToolUseID[:8]+"...", strings.TrimSpace(displayContent))
							} else if len(displayContent) > 150 {
								fmt.Printf("🔧 AWS Tool Result (id: %s): %s...\n", b.ToolUseID[:8]+"...", displayContent[:150])
							} else {
								fmt.Printf("🔧 AWS Tool Result (id: %s): %s\n", b.ToolUseID[:8]+"...", displayContent)
							}
						} else {
							fmt.Printf("🔧 AWS Tool Result (id: %s): <structured data>\n", b.ToolUseID[:8]+"...")
						}
					}
				}
			} else if contentStr, ok := msg.Content.(string); ok {
				fmt.Printf("📤 User: %s\n", contentStr)
			}
		case *claudecode.SystemMessage:
		case *claudecode.ResultMessage:
			if msg.IsError {
				fmt.Printf("\n❌ Claude returned error: %s\n", msg.Result)
				fmt.Println("This may be related to MCP tool permissions or AWS access.")
			}
		default:
			fmt.Printf("\n📦 Unexpected message type: %T\n", message)
		}
	}

	fmt.Println("\n✅ S3 bucket listing completed!")
}

func checkAWSCredentials() bool {
	if os.Getenv("AWS_ACCESS_KEY_ID") != "" && os.Getenv("AWS_SECRET_ACCESS_KEY") != "" {
		return true
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	credentialsPath := homeDir + "/.aws/credentials"
	if _, err := os.Stat(credentialsPath); err == nil {
		return true
	}

	return false
}