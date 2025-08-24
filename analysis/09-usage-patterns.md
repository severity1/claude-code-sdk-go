# Usage Patterns Analysis

Analysis of advanced usage patterns, framework integration, and best practices from the Python SDK.

## Basic Query Pattern

**Python Simple Query**:
```python
import claude_code

# Simple string query
response = await claude_code.query("What's the weather like today?")
for message in response:
    if message.type == "assistant":
        for block in message.content:
            if hasattr(block, 'text'):
                print(block.text)
```

**Go Simple Query**:
```go
package main

import (
    "context"
    "fmt"
    "log"
    
    claude "github.com/your-org/claude-code-sdk-go"
)

func main() {
    ctx := context.Background()
    
    response, err := claude.Query(ctx, "What's the weather like today?", nil)
    if err != nil {
        log.Fatal(err)
    }
    
    for _, message := range response {
        if assistantMsg, ok := message.(*claude.AssistantMessage); ok {
            for _, block := range assistantMsg.Content {
                if textBlock, ok := block.(*claude.TextBlock); ok {
                    fmt.Println(textBlock.Text)
                }
            }
        }
    }
}
```

## Advanced Configuration Patterns

**Python Configuration with Options**:
```python
import claude_code
from pathlib import Path

options = claude_code.ClaudeCodeOptions(
    system_prompt="You are a helpful coding assistant.",
    allowed_tools=["Read", "Edit", "Bash"],
    max_turns=5,
    permission_mode="acceptEdits",
    cwd=Path("./my-project"),
    mcp_servers={
        "filesystem": {
            "command": "npx",
            "args": ["@modelcontextprotocol/server-filesystem", "/tmp"],
        }
    }
)

response = await claude_code.query(
    "Help me refactor this Python function",
    options=options
)
```

**Go Configuration with Builder Pattern**:
```go
func advancedQuery() {
    ctx := context.Background()
    
    options := claude.NewOptions().
        WithSystemPrompt("You are a helpful coding assistant.").
        WithAllowedTools("Read", "Edit", "Bash").
        WithMaxTurns(5).
        WithPermissionMode(claude.PermissionModeAcceptEdits).
        WithWorkingDirectory("./my-project").
        WithMCPServer("filesystem", &claude.McpStdioServerConfig{
            Command: "npx",
            Args:    []string{"@modelcontextprotocol/server-filesystem", "/tmp"},
        })
    
    response, err := claude.Query(ctx, "Help me refactor this Python function", options)
    if err != nil {
        log.Fatal(err)
    }
    
    // Process response...
}
```

## Bidirectional Communication Pattern

**Python Client Usage**:
```python
import claude_code

async def interactive_session():
    client = claude_code.ClaudeSDKClient()
    await client.start()
    
    try:
        # Send initial message
        await client.send_message("Hello, can you help me with Go programming?")
        
        # Process responses
        async for message in client.messages():
            if message.type == "assistant":
                print("Claude:", extract_text(message))
            elif message.type == "system" and message.subtype == "tool_use":
                # Handle tool use
                await handle_tool_use(client, message)
        
        # Send follow-up
        await client.send_message("Can you show me a concrete example?")
        
        # Continue processing...
        async for message in client.messages():
            if message.type == "result":
                print(f"Session complete: {message.session_id}")
                break
                
    finally:
        await client.stop()

def extract_text(message):
    text_parts = []
    for block in message.content:
        if hasattr(block, 'text'):
            text_parts.append(block.text)
    return ' '.join(text_parts)

async def handle_tool_use(client, message):
    """Handle tool use requests with appropriate permissions."""
    # Implementation depends on permission mode
    pass
```

**Go Client Usage**:
```go
func interactiveSession() error {
    ctx := context.Background()
    
    client, err := claude.NewClient(nil)
    if err != nil {
        return err
    }
    
    if err := client.Start(ctx); err != nil {
        return err
    }
    defer client.Stop()
    
    // Send initial message
    if err := client.SendMessage(ctx, "Hello, can you help me with Go programming?"); err != nil {
        return err
    }
    
    // Process responses
    for {
        select {
        case message := <-client.Messages():
            if assistantMsg, ok := message.(*claude.AssistantMessage); ok {
                fmt.Printf("Claude: %s\n", extractText(assistantMsg))
            } else if sysMsg, ok := message.(*claude.SystemMessage); ok && sysMsg.Subtype == "tool_use" {
                if err := handleToolUse(ctx, client, sysMsg); err != nil {
                    return err
                }
            }
            
        case err := <-client.Errors():
            if err != nil {
                return err
            }
            
        case <-ctx.Done():
            return ctx.Err()
        }
    }
}

func extractText(message *claude.AssistantMessage) string {
    var textParts []string
    for _, block := range message.Content {
        if textBlock, ok := block.(*claude.TextBlock); ok {
            textParts = append(textParts, textBlock.Text)
        }
    }
    return strings.Join(textParts, " ")
}

func handleToolUse(ctx context.Context, client *claude.Client, message *claude.SystemMessage) error {
    // Handle tool use based on permission mode
    return nil
}
```

## Error Handling Patterns

**Python Error Handling with Recovery**:
```python
import claude_code
import asyncio
import logging

async def robust_query(prompt: str, max_retries: int = 3):
    """Query with automatic retry and error recovery."""
    
    for attempt in range(max_retries):
        try:
            response = await claude_code.query(prompt)
            return list(response)  # Convert to list to fully consume
            
        except claude_code.CLINotFoundError as e:
            logging.error(f"Claude Code not installed: {e}")
            raise  # Don't retry installation issues
            
        except claude_code.ProcessError as e:
            if e.exit_code == 1 and "rate limit" in e.stderr.lower():
                # Exponential backoff for rate limits
                wait_time = 2 ** attempt
                logging.warning(f"Rate limited, waiting {wait_time}s...")
                await asyncio.sleep(wait_time)
                continue
            else:
                logging.error(f"Process error: {e}")
                raise
                
        except claude_code.CLIJSONDecodeError as e:
            if attempt < max_retries - 1:
                logging.warning(f"JSON decode error (attempt {attempt + 1}): {e}")
                continue
            else:
                logging.error(f"Persistent JSON decode error: {e}")
                raise
                
        except Exception as e:
            logging.error(f"Unexpected error: {e}")
            if attempt < max_retries - 1:
                await asyncio.sleep(1)  # Brief pause before retry
                continue
            raise
    
    raise RuntimeError(f"Failed after {max_retries} attempts")
```

**Go Error Handling with Recovery**:
```go
func robustQuery(ctx context.Context, prompt string, maxRetries int) ([]claude.Message, error) {
    var lastErr error
    
    for attempt := 0; attempt < maxRetries; attempt++ {
        response, err := claude.Query(ctx, prompt, nil)
        if err == nil {
            return response, nil
        }
        
        lastErr = err
        
        // Handle specific error types
        var cliNotFoundErr *claude.CLINotFoundError
        if errors.As(err, &cliNotFoundErr) {
            // Don't retry installation issues
            return nil, err
        }
        
        var processErr *claude.ProcessError
        if errors.As(err, &processErr) {
            if processErr.ExitCode == 1 && strings.Contains(strings.ToLower(processErr.Stderr), "rate limit") {
                // Exponential backoff for rate limits
                waitTime := time.Duration(1<<attempt) * time.Second
                log.Printf("Rate limited, waiting %v...", waitTime)
                
                select {
                case <-time.After(waitTime):
                    continue
                case <-ctx.Done():
                    return nil, ctx.Err()
                }
            } else {
                log.Printf("Process error: %v", err)
                return nil, err
            }
        }
        
        var jsonErr *claude.JSONDecodeError
        if errors.As(err, &jsonErr) {
            if attempt < maxRetries-1 {
                log.Printf("JSON decode error (attempt %d): %v", attempt+1, err)
                
                select {
                case <-time.After(time.Second):
                    continue
                case <-ctx.Done():
                    return nil, ctx.Err()
                }
            } else {
                log.Printf("Persistent JSON decode error: %v", err)
                return nil, err
            }
        }
        
        // Other errors - brief pause before retry
        log.Printf("Unexpected error: %v", err)
        if attempt < maxRetries-1 {
            select {
            case <-time.After(time.Second):
                continue
            case <-ctx.Done():
                return nil, ctx.Err()
            }
        }
    }
    
    return nil, fmt.Errorf("failed after %d attempts: %w", maxRetries, lastErr)
}
```

## Streaming Response Processing

**Python Streaming Pattern**:
```python
async def stream_processing():
    """Process responses as they arrive."""
    client = claude_code.ClaudeSDKClient()
    await client.start()
    
    try:
        await client.send_message("Write a long explanation about Go concurrency")
        
        current_response = []
        async for message in client.messages():
            if message.type == "assistant":
                current_response.append(message)
                
                # Process incrementally for better UX
                text = extract_text(message)
                if text:
                    print(text, end="", flush=True)
                    
            elif message.type == "result":
                # End of response
                print(f"\n\nSession {message.session_id} complete")
                print(f"Cost: ${message.total_cost_usd or 0:.4f}")
                break
                
    finally:
        await client.stop()
```

**Go Streaming Pattern**:
```go
func streamProcessing(ctx context.Context) error {
    client, err := claude.NewClient(nil)
    if err != nil {
        return err
    }
    
    if err := client.Start(ctx); err != nil {
        return err
    }
    defer client.Stop()
    
    if err := client.SendMessage(ctx, "Write a long explanation about Go concurrency"); err != nil {
        return err
    }
    
    var currentResponse []claude.Message
    
    for {
        select {
        case message := <-client.Messages():
            if assistantMsg, ok := message.(*claude.AssistantMessage); ok {
                currentResponse = append(currentResponse, message)
                
                // Process incrementally for better UX
                text := extractText(assistantMsg)
                if text != "" {
                    fmt.Print(text)
                }
                
            } else if resultMsg, ok := message.(*claude.ResultMessage); ok {
                // End of response
                cost := 0.0
                if resultMsg.TotalCostUSD != nil {
                    cost = *resultMsg.TotalCostUSD
                }
                fmt.Printf("\n\nSession %s complete\n", resultMsg.SessionID)
                fmt.Printf("Cost: $%.4f\n", cost)
                return nil
            }
            
        case err := <-client.Errors():
            if err != nil {
                return err
            }
            
        case <-ctx.Done():
            return ctx.Err()
        }
    }
}
```

## Framework Integration Patterns

**Python AsyncIO Integration**:
```python
import asyncio
import claude_code

class ClaudeService:
    def __init__(self):
        self.client = None
    
    async def start(self):
        self.client = claude_code.ClaudeSDKClient()
        await self.client.start()
    
    async def stop(self):
        if self.client:
            await self.client.stop()
    
    async def ask(self, question: str) -> str:
        await self.client.send_message(question)
        
        response_parts = []
        async for message in self.client.messages():
            if message.type == "assistant":
                response_parts.append(extract_text(message))
            elif message.type == "result":
                break
        
        return ''.join(response_parts)

# Usage with async context manager
async def main():
    service = ClaudeService()
    try:
        await service.start()
        answer = await service.ask("What is Go's main advantage?")
        print(answer)
    finally:
        await service.stop()

# Run with asyncio
asyncio.run(main())
```

**Go Goroutine and Channel Integration**:
```go
type ClaudeService struct {
    client   *claude.Client
    ctx      context.Context
    cancel   context.CancelFunc
    requests chan QueryRequest
    done     chan struct{}
}

type QueryRequest struct {
    Question string
    Response chan QueryResponse
}

type QueryResponse struct {
    Answer string
    Error  error
}

func NewClaudeService() *ClaudeService {
    ctx, cancel := context.WithCancel(context.Background())
    return &ClaudeService{
        ctx:      ctx,
        cancel:   cancel,
        requests: make(chan QueryRequest, 10),
        done:     make(chan struct{}),
    }
}

func (s *ClaudeService) Start() error {
    client, err := claude.NewClient(nil)
    if err != nil {
        return err
    }
    
    if err := client.Start(s.ctx); err != nil {
        return err
    }
    
    s.client = client
    
    // Start request processing goroutine
    go s.processRequests()
    
    return nil
}

func (s *ClaudeService) Stop() {
    close(s.requests)
    s.cancel()
    <-s.done
    
    if s.client != nil {
        s.client.Stop()
    }
}

func (s *ClaudeService) Ask(ctx context.Context, question string) (string, error) {
    req := QueryRequest{
        Question: question,
        Response: make(chan QueryResponse, 1),
    }
    
    select {
    case s.requests <- req:
    case <-ctx.Done():
        return "", ctx.Err()
    }
    
    select {
    case resp := <-req.Response:
        return resp.Answer, resp.Error
    case <-ctx.Done():
        return "", ctx.Err()
    }
}

func (s *ClaudeService) processRequests() {
    defer close(s.done)
    
    for req := range s.requests {
        answer, err := s.handleQuery(req.Question)
        
        select {
        case req.Response <- QueryResponse{Answer: answer, Error: err}:
        case <-s.ctx.Done():
            return
        }
    }
}

func (s *ClaudeService) handleQuery(question string) (string, error) {
    if err := s.client.SendMessage(s.ctx, question); err != nil {
        return "", err
    }
    
    var responseParts []string
    
    for {
        select {
        case message := <-s.client.Messages():
            if assistantMsg, ok := message.(*claude.AssistantMessage); ok {
                responseParts = append(responseParts, extractText(assistantMsg))
            } else if _, ok := message.(*claude.ResultMessage); ok {
                return strings.Join(responseParts, ""), nil
            }
            
        case err := <-s.client.Errors():
            if err != nil {
                return "", err
            }
            
        case <-s.ctx.Done():
            return "", s.ctx.Err()
        }
    }
}

// Usage
func main() {
    service := NewClaudeService()
    if err := service.Start(); err != nil {
        log.Fatal(err)
    }
    defer service.Stop()
    
    ctx := context.Background()
    answer, err := service.Ask(ctx, "What is Go's main advantage?")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println(answer)
}
```

## Testing Patterns

**Python Testing with Mock Transport**:
```python
import pytest
import claude_code
from unittest.mock import AsyncMock, MagicMock

@pytest.fixture
def mock_transport():
    transport = MagicMock()
    transport.connect = AsyncMock()
    transport.disconnect = AsyncMock()
    transport.send_request = AsyncMock()
    transport.is_connected.return_value = True
    
    # Mock message stream
    async def mock_messages():
        yield {"type": "assistant", "content": [{"type": "text", "text": "Hello!"}]}
        yield {"type": "result", "session_id": "test-123", "is_error": False}
    
    transport.receive_messages.return_value = mock_messages()
    return transport

@pytest.mark.asyncio
async def test_query_with_mock(mock_transport):
    # Inject mock transport
    with patch('claude_code._internal.transport.SubprocessCLITransport', return_value=mock_transport):
        response = await claude_code.query("Hello")
        messages = list(response)
        
        assert len(messages) == 2
        assert messages[0].type == "assistant"
        assert messages[1].type == "result"
```

**Go Testing with Mock Transport**:
```go
type MockTransport struct {
    messages []map[string]interface{}
    errors   []error
    connected bool
}

func (m *MockTransport) Connect(ctx context.Context) error {
    m.connected = true
    return nil
}

func (m *MockTransport) Disconnect() error {
    m.connected = false
    return nil
}

func (m *MockTransport) SendMessage(ctx context.Context, message claude.StreamMessage) error {
    return nil
}

func (m *MockTransport) ReceiveMessages(ctx context.Context) (<-chan map[string]interface{}, <-chan error) {
    msgChan := make(chan map[string]interface{}, len(m.messages))
    errChan := make(chan error, len(m.errors))
    
    go func() {
        defer close(msgChan)
        defer close(errChan)
        
        for _, msg := range m.messages {
            msgChan <- msg
        }
        for _, err := range m.errors {
            errChan <- err
        }
    }()
    
    return msgChan, errChan
}

func (m *MockTransport) Interrupt(ctx context.Context) error { return nil }
func (m *MockTransport) IsConnected() bool { return m.connected }

func TestQueryWithMock(t *testing.T) {
    transport := &MockTransport{
        messages: []map[string]interface{}{
            {
                "type": "assistant",
                "content": []interface{}{
                    map[string]interface{}{"type": "text", "text": "Hello!"},
                },
            },
            {
                "type": "result",
                "session_id": "test-123",
                "is_error": false,
            },
        },
    }
    
    // Create client with mock transport
    client := &claude.Client{transport: transport}
    
    ctx := context.Background()
    messages, err := client.Query(ctx, "Hello")
    
    assert.NoError(t, err)
    assert.Len(t, messages, 2)
    assert.IsType(t, &claude.AssistantMessage{}, messages[0])
    assert.IsType(t, &claude.ResultMessage{}, messages[1])
}
```

These usage patterns demonstrate the flexibility and power of both SDKs while showing how Go's native concurrency and type safety can provide excellent developer experience.