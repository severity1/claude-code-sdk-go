// Package main demonstrates Query API with file tools (Read/Write).
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/severity1/claude-code-sdk-go"
)

func main() {
	fmt.Println("Claude Code SDK - Query API with Tools Example")
	fmt.Println("File operations: Read analysis and Write generation")

	ctx := context.Background()

	// Setup simple demo files
	if err := setupFiles(); err != nil {
		log.Fatalf("Setup failed: %v", err)
	}
	defer os.RemoveAll("demo")

	// Change to demo directory
	if err := os.Chdir("demo"); err != nil {
		log.Fatalf("Failed to change to demo directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(".."); err != nil {
			log.Printf("Warning: Failed to change back to parent directory: %v", err)
		}
	}()

	// Example 1: Read and analyze files
	fmt.Println("\n--- Example 1: File Analysis ---")
	analyzeQuery := `Read README.md and config.json, then analyze:
1. What this project does
2. Current configuration settings
3. Any potential improvements`

	if err := queryWithTools(ctx, analyzeQuery, []string{"Read"}); err != nil {
		log.Printf("Analysis failed: %v", err)
	}

	// Example 2: Generate documentation
	fmt.Println("\n--- Example 2: Generate Documentation ---")
	docQuery := `Read all files in the current directory and create a PROJECT_SUMMARY.md with:
- Project overview
- Configuration details
- Usage instructions

Use proper markdown formatting.`

	if err := queryWithTools(ctx, docQuery, []string{"Read", "Write"}); err != nil {
		log.Printf("Documentation failed: %v", err)
	}

	fmt.Println("\nTool examples completed!")
}

func queryWithTools(ctx context.Context, question string, allowedTools []string) error {
	fmt.Printf("Tools: %v\n", allowedTools)
	fmt.Printf("Query: %s\n", question)

	iterator, err := claudecode.Query(ctx, question,
		claudecode.WithAllowedTools(allowedTools...),
		claudecode.WithSystemPrompt("You are a helpful assistant. Explain your file operations clearly."),
	)
	if err != nil {
		return fmt.Errorf("query failed: %w", err)
	}
	defer iterator.Close()

	fmt.Println("\nResponse:")

	for {
		message, err := iterator.Next(ctx)
		if err != nil {
			if errors.Is(err, claudecode.ErrNoMoreMessages) {
				break
			}
			return fmt.Errorf("failed to get next message: %w", err)
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
								fmt.Printf("⚠️ Tool error: %s\n", content)
							} else if len(content) > 100 {
								fmt.Printf("📁 Tool: %s...\n", content[:100])
							} else {
								fmt.Printf("📁 Tool: %s\n", content)
							}
						}
					}
				}
			}
		case *claudecode.ResultMessage:
			if msg.IsError {
				return fmt.Errorf("error: %s", msg.Result)
			}
		}
	}

	fmt.Println("\n✅ Completed")
	return nil
}

func setupFiles() error {
	if err := os.MkdirAll("demo", 0755); err != nil {
		return err
	}

	readme := `# Demo Web Server

Simple HTTP server with configuration management.

## Features
- HTTP routing
- JSON configuration
- Basic logging

## Usage
1. Configure settings in config.json
2. Run: go run main.go
3. Visit: http://localhost:8080
`

	config := `{
  "server": {
    "port": 8080,
    "host": "localhost"
  },
  "database": {
    "url": "sqlite://app.db",
    "timeout": 5000
  },
  "debug": true
}`

	files := map[string]string{
		filepath.Join("demo", "README.md"):   readme,
		filepath.Join("demo", "config.json"): config,
	}

	for path, content := range files {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return err
		}
	}

	return nil
}
