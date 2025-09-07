// Package main demonstrates using Claude Code SDK Query API with Read/Write tools.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/severity1/claude-code-sdk-go"
)

func main() {
	fmt.Println("Claude Code SDK for Go - Query API with Read/Write Tools Example")
	fmt.Println("================================================================")

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	workDir := "./query_tools_demo"
	if err := setupDemoFiles(workDir); err != nil {
		log.Fatalf("Failed to setup demo files: %v", err)
	}
	defer cleanupDemoFiles(workDir)

	originalDir, _ := os.Getwd()
	os.Chdir(workDir)
	defer os.Chdir(originalDir)

	fmt.Printf("📁 Working in directory: %s\n", workDir)
	fmt.Println("📄 Demo files created: config.json, README.md, data.txt")

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("📖 EXAMPLE 1: File Analysis with Read Tool")
	fmt.Println(strings.Repeat("=", 60))

	query1 := `Please read the README.md file and config.json file in the current directory. 
Then analyze their contents and tell me:
1. What this project appears to be about
2. What configuration settings are present
3. Any potential issues or improvements you notice`

	if err := runQueryWithTools(ctx, "File Analysis", query1, []string{"Read"}); err != nil {
		log.Printf("Query 1 failed: %v", err)
	}

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("📝 EXAMPLE 2: Documentation Generation with Read/Write Tools")
	fmt.Println(strings.Repeat("=", 60))

	query2 := `Please read all the files in the current directory (README.md, config.json, data.txt) and create a comprehensive project summary.
Write this summary to a new file called PROJECT_SUMMARY.md with the following sections:
- Project Overview
- Configuration Details  
- Data Analysis
- Recommendations

Make it well-formatted with proper markdown.`

	if err := runQueryWithTools(ctx, "Documentation Generation", query2, []string{"Read", "Write"}); err != nil {
		log.Printf("Query 2 failed: %v", err)
	}

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🔧 EXAMPLE 3: Configuration Update with Read/Write/Edit Tools")
	fmt.Println(strings.Repeat("=", 60))

	query3 := `Please read the config.json file and:
1. Update the database URL to use a production-ready connection string
2. Add a new "logging" section with appropriate log levels
3. Increase the timeout values for better reliability
4. Write the updated configuration back to config.json

Explain the changes you made and why.`

	if err := runQueryWithTools(ctx, "Configuration Update", query3, []string{"Read", "Write", "Edit"}); err != nil {
		log.Printf("Query 3 failed: %v", err)
	}

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🔒 EXAMPLE 4: Read-Only Analysis (Security Restricted)")
	fmt.Println(strings.Repeat("=", 60))

	query4 := `Please analyze all files in the current directory and suggest improvements.
Try to create a backup of the important files.`

	fmt.Println("Note: This query is restricted to Read-only tools for security")
	if err := runQueryWithTools(ctx, "Read-Only Security Demo", query4, []string{"Read"}); err != nil {
		log.Printf("Query 4 failed: %v", err)
	}

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🎉 Query with Tools Examples Completed!")

	fmt.Println("\n📂 Files created/modified during this demo:")
	listFiles(workDir)
}

func runQueryWithTools(ctx context.Context, title, question string, allowedTools []string) error {
	fmt.Printf("\n🎯 %s\n", title)
	fmt.Printf("🔧 Allowed tools: %v\n", allowedTools)
	fmt.Printf("❓ Query: %s\n", strings.TrimSpace(question))
	fmt.Println(strings.Repeat("-", 50))

	iterator, err := claudecode.Query(ctx, question,
		claudecode.WithAllowedTools(allowedTools...),
		claudecode.WithSystemPrompt("You are a helpful assistant that can read and write files. Be thorough and explain your actions clearly."),
	)
	if err != nil {
		return fmt.Errorf("failed to create query: %w", err)
	}
	defer iterator.Close()

	fmt.Println("🤖 Claude's Response:")
	responseReceived := false

	for {
		message, err := iterator.Next(ctx)
		if err != nil {
			if err.Error() == "no more messages" {
				break
			}
			return fmt.Errorf("failed to get next message: %w", err)
		}

		if message == nil {
			break
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
					switch b := block.(type) {
					case *claudecode.TextBlock:
						fmt.Printf("📤 Tool: %s\n", b.Text)
					case *claudecode.ToolResultBlock:
						if contentStr, ok := b.Content.(string); ok {
							displayContent := contentStr
							if strings.Contains(contentStr, "<tool_use_error>") {
								displayContent = strings.ReplaceAll(contentStr, "<tool_use_error>", "⚠️ ")
								displayContent = strings.ReplaceAll(displayContent, "</tool_use_error>", "")
								fmt.Printf("🔧 Tool Issue (id: %s): %s\n", b.ToolUseID[:8]+"...", strings.TrimSpace(displayContent))
							} else if len(displayContent) > 150 {
								fmt.Printf("🔧 Tool Result (id: %s): %s...\n", b.ToolUseID[:8]+"...", displayContent[:150])
							} else {
								fmt.Printf("🔧 Tool Result (id: %s): %s\n", b.ToolUseID[:8]+"...", displayContent)
							}
						} else {
							fmt.Printf("🔧 Tool Result (id: %s): <structured data>\n", b.ToolUseID[:8]+"...")
						}
					}
				}
			} else if contentStr, ok := msg.Content.(string); ok {
				fmt.Printf("📤 User: %s\n", contentStr)
			}
		case *claudecode.SystemMessage:
		case *claudecode.ResultMessage:
			if msg.IsError {
				return fmt.Errorf("claude returned error: %s", msg.Result)
			}
		default:
			fmt.Printf("\n📦 Unexpected message type: %T\n", message)
		}
	}

	if !responseReceived {
		return fmt.Errorf("no response received")
	}

	fmt.Println("\n✅ Query completed successfully")
	return nil
}

func setupDemoFiles(workDir string) error {
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return err
	}

	readme := `# Demo Project

This is a sample project for demonstrating Claude Code SDK with file tools.

## Features
- Configuration management
- Data processing
- Documentation generation

## Usage
Run the main application with your preferred settings.

## Requirements
- Go 1.18+
- Valid configuration file
`

	if err := os.WriteFile(filepath.Join(workDir, "README.md"), []byte(readme), 0644); err != nil {
		return err
	}

	config := `{
  "database": {
    "url": "sqlite://./dev.db",
    "timeout": 5000,
    "max_connections": 10
  },
  "server": {
    "port": 8080,
    "host": "localhost"
  },
  "features": {
    "debug": true,
    "metrics": false
  }
}`

	if err := os.WriteFile(filepath.Join(workDir, "config.json"), []byte(config), 0644); err != nil {
		return err
	}

	data := `Sample Data File
================

User records:
- John Doe, Admin, Active
- Jane Smith, User, Active  
- Bob Johnson, User, Inactive

System metrics:
- CPU Usage: 45%
- Memory: 2.1GB/8GB
- Disk: 120GB/500GB
- Uptime: 72 hours

Last updated: 2024-01-15
`

	if err := os.WriteFile(filepath.Join(workDir, "data.txt"), []byte(data), 0644); err != nil {
		return err
	}

	return nil
}

func cleanupDemoFiles(workDir string) {
	os.RemoveAll(workDir)
}

func listFiles(dir string) {
	files, err := os.ReadDir(dir)
	if err != nil {
		fmt.Printf("   Error reading directory: %v\n", err)
		return
	}

	for _, file := range files {
		if file.IsDir() {
			fmt.Printf("   📁 %s/\n", file.Name())
		} else {
			info, _ := file.Info()
			fmt.Printf("   📄 %s (%d bytes)\n", file.Name(), info.Size())
		}
	}
}
