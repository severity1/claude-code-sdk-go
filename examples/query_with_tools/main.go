// Package main demonstrates using Claude Code SDK Query API with Read/Write tools.
//
// This example shows how to:
// - Configure Query API to use specific tools (Read, Write, Edit)
// - Execute queries that automatically invoke file operations
// - Handle responses that include both text analysis and tool results
// - Demonstrate practical file manipulation workflows
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

	// Create context with timeout for file operations
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Setup working directory and sample files
	workDir := "./query_tools_demo"
	if err := setupDemoFiles(workDir); err != nil {
		log.Fatalf("Failed to setup demo files: %v", err)
	}
	defer cleanupDemoFiles(workDir)

	// Change to the working directory so Claude can access the files
	originalDir, _ := os.Getwd()
	os.Chdir(workDir)
	defer os.Chdir(originalDir)

	fmt.Printf("📁 Working in directory: %s\n", workDir)
	fmt.Println("📄 Demo files created: config.json, README.md, data.txt")

	// Example 1: Query with Read tool to analyze existing files
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

	// Example 2: Query with Read + Write tools to create documentation
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

	// Example 3: Query with Read + Write + Edit tools for refactoring
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

	// Example 4: Query with restricted tools (security demonstration)
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
	fmt.Println("\n✨ What was demonstrated:")
	fmt.Println("   • Read tool: Analyze existing files and configuration")
	fmt.Println("   • Write tool: Create new documentation and summaries")
	fmt.Println("   • Edit tool: Modify existing files with improvements")
	fmt.Println("   • Tool restriction: Security by limiting available tools")
	fmt.Println("   • Automatic tool selection: Claude chooses appropriate tools")
	fmt.Println("   • Real file operations: Actual changes made to filesystem")
	
	fmt.Println("\n📂 Files created/modified during this demo:")
	listFiles(workDir)
}

// runQueryWithTools executes a query with specific tool restrictions
func runQueryWithTools(ctx context.Context, title, question string, allowedTools []string) error {
	fmt.Printf("\n🎯 %s\n", title)
	fmt.Printf("🔧 Allowed tools: %v\n", allowedTools)
	fmt.Printf("❓ Query: %s\n", strings.TrimSpace(question))
	fmt.Println(strings.Repeat("-", 50))

	// Create query with tool restrictions
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
		case *claudecode.SystemMessage:
			// System initialization (can be ignored)
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

// setupDemoFiles creates sample files for the demonstration
func setupDemoFiles(workDir string) error {
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return err
	}

	// Create README.md
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

	// Create config.json
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

	// Create data.txt
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

// cleanupDemoFiles removes the demo directory
func cleanupDemoFiles(workDir string) {
	os.RemoveAll(workDir)
}

// listFiles shows what files exist in the directory
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