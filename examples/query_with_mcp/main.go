// Package main demonstrates using Claude Code SDK Query API with MCP tools (AWS API).
//
// This example shows how to:
// - Configure Query API to use AWS MCP tools
// - Execute a simple query to list S3 buckets
// - Handle AWS API responses through Claude
//
// Prerequisites:
// - AWS API MCP server installed: `claude mcp add aws-api`
// - AWS credentials configured (AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY)
// - Or AWS credentials in ~/.aws/credentials
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/severity1/claude-code-sdk-go"
)

func main() {
	fmt.Println("Claude Code SDK for Go - Query API with AWS MCP Tools Example")
	fmt.Println("=============================================================")

	// Check if AWS credentials are available
	if !checkAWSCredentials() {
		fmt.Println("⚠️  AWS credentials not found.")
		fmt.Println("Please configure AWS credentials using one of these methods:")
		fmt.Println("  • Environment variables: AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY")
		fmt.Println("  • AWS credentials file: ~/.aws/credentials")
		fmt.Println("  • AWS CLI: `aws configure`")
		fmt.Println("\n📚 This example will show how the code works with mock explanations.")
	}

	// Create context with timeout for AWS operations
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Simple S3 bucket listing with Query API
	fmt.Println("\n🪣 Listing S3 Buckets")
	fmt.Println("====================")

	query := `Please list all my S3 buckets and show me their names, creation dates, and regions.`

	fmt.Printf("🔧 Allowed tools: [mcp__aws-api-mcp__*]\n")
	fmt.Printf("❓ Query: %s\n", query)
	fmt.Println("--------------------------------------------------")

	// Create query with AWS MCP tools
	iterator, err := claudecode.Query(ctx, query,
		claudecode.WithAllowedTools("mcp__aws-api-mcp__*"),
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
					fmt.Print(b.Text)
				case *claudecode.ThinkingBlock:
					// Show Claude's thinking process for AWS operations
					fmt.Printf("\n💭 [AWS Analysis: %s]\n", b.Thinking)
				}
			}
		case *claudecode.SystemMessage:
			// System initialization - can be ignored
		case *claudecode.ResultMessage:
			if msg.IsError {
				log.Fatalf("Claude returned error: %s", msg.Result)
			}
		default:
			fmt.Printf("\n📦 Unexpected message type: %T\n", message)
		}
	}

	fmt.Println("\n✅ S3 bucket listing completed!")
}

// checkAWSCredentials checks if AWS credentials are available
func checkAWSCredentials() bool {
	// Check environment variables
	if os.Getenv("AWS_ACCESS_KEY_ID") != "" && os.Getenv("AWS_SECRET_ACCESS_KEY") != "" {
		return true
	}

	// Check AWS credentials file
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