// Package main demonstrates Client API with file tools using WithClient for automatic resource management.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/severity1/claude-code-sdk-go"
)

func main() {
	fmt.Println("Claude Code SDK - Client API with Tools Example")
	fmt.Println("Interactive file operations with context preservation")

	// Setup demo files
	if err := setupFiles(); err != nil {
		log.Fatalf("Setup failed: %v", err)
	}
	defer os.RemoveAll("demo")

	if err := os.Chdir("demo"); err != nil {
		log.Fatalf("Failed to change to demo directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(".."); err != nil {
			log.Printf("Warning: Failed to change back to parent directory: %v", err)
		}
	}()

	ctx := context.Background()

	// Multi-turn conversation with WithClient automatic resource management
	err := claudecode.WithClient(ctx, func(client claudecode.Client) error {
		fmt.Println("\nConnected! Starting interactive session...")

		// Turn 1: Analyze the project
		fmt.Println("\n--- Turn 1: Project Analysis ---")
		query1 := "Read all files in this project and give me an overview of what it does"

		if err := client.Query(ctx, query1); err != nil {
			return fmt.Errorf("turn 1 query failed: %w", err)
		}

		if err := streamResponse(ctx, client); err != nil {
			return fmt.Errorf("turn 1 failed: %w", err)
		}

		// Turn 2: Improve the project
		fmt.Println("\n--- Turn 2: Code Improvements ---")
		query2 := "Based on what you learned, improve the main.go file with better error handling and create a README.md"

		if err := client.Query(ctx, query2); err != nil {
			return fmt.Errorf("turn 2 query failed: %w", err)
		}

		if err := streamResponse(ctx, client); err != nil {
			return fmt.Errorf("turn 2 failed: %w", err)
		}

		fmt.Println("\nInteractive session completed!")
		fmt.Println("Context was preserved across both turns automatically")
		return nil
	}, claudecode.WithAllowedTools("Read", "Write", "Edit"),
		claudecode.WithSystemPrompt("You are a helpful software development assistant."))

	if err != nil {
		log.Fatalf("Session failed: %v", err)
	}
}

func streamResponse(ctx context.Context, client claudecode.Client) error {
	fmt.Println("\nResponse:")

	msgChan := client.ReceiveMessages(ctx)
	for {
		select {
		case message := <-msgChan:
			if message == nil {
				return nil
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
								if len(content) > 100 {
									fmt.Printf("📁 File operation: %s...\n", content[:100])
								} else {
									fmt.Printf("📁 File operation: %s\n", content)
								}
							}
						}
					}
				}
			case *claudecode.ResultMessage:
				if msg.IsError {
					return fmt.Errorf("error: %s", msg.Result)
				}
				return nil
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func setupFiles() error {
	if err := os.MkdirAll("demo", 0755); err != nil {
		return err
	}

	packageJSON := `{
  "name": "demo-web-server",
  "version": "1.0.0",
  "description": "Simple Go web server demo",
  "main": "main.go",
  "keywords": ["go", "web", "demo"],
  "author": "Demo"
}`

	mainGo := `package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, World!")
	})
	
	fmt.Println("Server starting on :8080")
	http.ListenAndServe(":8080", nil)
}`

	files := map[string]string{
		filepath.Join("demo", "package.json"): packageJSON,
		filepath.Join("demo", "main.go"):      mainGo,
	}

	for path, content := range files {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return err
		}
	}

	return nil
}
