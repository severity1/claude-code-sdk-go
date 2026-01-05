# Contributing

Thank you for your interest in contributing to the Claude Agent SDK for Go!

## Prerequisites

- **Go 1.18+** - [Download Go](https://go.dev/dl/)
- **Node.js** - [Download Node.js](https://nodejs.org/)
- **Claude Code CLI** - `npm install -g @anthropic-ai/claude-code`

## Development Setup

```bash
# Clone the repository
git clone https://github.com/severity1/claude-agent-sdk-go.git
cd claude-agent-sdk-go

# Verify installation
go build ./...
go test ./...
```

## Code Style

### Formatting and Linting

Run these before every commit:

```bash
go fmt ./...           # Format code
go vet ./...           # Static analysis
golangci-lint run      # Comprehensive linting (if available)
```

### Go Conventions

- **Interface-driven design**: Define behavior through interfaces
- **Context-first**: All blocking functions accept `context.Context` as first parameter
- **Error wrapping**: Use `fmt.Errorf("context: %w", err)` for error chains
- **Functional options**: Use `WithXxx()` pattern for configuration
- **No unnecessary exports**: Keep identifiers unexported unless needed

### Naming

- Interfaces: Describe behavior (e.g., `Transport`, `Message`)
- Implementations: Concrete names (e.g., `ClientImpl`, `mockTransport`)
- Options: `WithXxx()` pattern (e.g., `WithSystemPrompt()`)
- Errors: `XxxError` suffix (e.g., `CLINotFoundError`)

## Testing

### Running Tests

```bash
go test ./...              # Run all tests
go test -v ./...           # Verbose output
go test -race ./...        # Race condition detection
go test -cover ./...       # Coverage report
```

### Test Patterns

Reference `client_test.go` as the gold standard for testing patterns.

**Test File Organization:**
```go
// 1. Test functions (primary purpose)
func TestFeature(t *testing.T) {...}

// 2. Mock implementations
type mockTransport struct {...}

// 3. Helper functions
func setupTest(t *testing.T) {...}
```

**Table-Driven Tests:**
```go
tests := []struct {
    name    string
    input   string
    want    string
    wantErr bool
}{
    {"valid input", "hello", "HELLO", false},
    {"empty input", "", "", true},
}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        got, err := Process(tt.input)
        if (err != nil) != tt.wantErr {
            t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
        }
        if got != tt.want {
            t.Errorf("got %v, want %v", got, tt.want)
        }
    })
}
```

**Helper Functions:**
```go
func setupTestContext(t *testing.T, timeout time.Duration) (context.Context, context.CancelFunc) {
    t.Helper()  // Critical for correct line reporting
    return context.WithTimeout(context.Background(), timeout)
}
```

**Thread-Safe Mocks:**
```go
type mockTransport struct {
    mu        sync.Mutex
    connected bool
    messages  []Message
}

func (m *mockTransport) SendMessage(msg Message) {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.messages = append(m.messages, msg)
}
```

## Commit Conventions

Use conventional commit messages:

```
<type>: <description>

[optional body]

[optional footer]
```

### Types

| Type | Description |
|------|-------------|
| `feat` | New feature |
| `fix` | Bug fix |
| `docs` | Documentation only |
| `test` | Adding or updating tests |
| `refactor` | Code change that neither fixes a bug nor adds a feature |
| `chore` | Maintenance tasks |

### Examples

```
feat: Add permission callback support (Issue #8)

Implement WithCanUseTool option for programmatic tool access control.

- Add CanUseToolCallback type
- Add PermissionResult types (Allow, Deny, AskUser)
- Update control protocol to handle permission requests
```

```
fix: Handle connection timeout correctly

The transport was not respecting context cancellation during
the initial handshake, causing hangs on slow connections.
```

### Issue References

Reference issues in commits:
- `(Issue #N)` - Related to issue
- `Closes #N` - Closes issue when merged (in PR body)

## Pull Request Process

### 1. Create a Feature Branch

```bash
git checkout main
git pull origin main
git checkout -b feature/issue-N-short-description
```

### 2. Make Changes

- Write tests first (TDD approach)
- Implement the feature
- Ensure all tests pass
- Run linting

### 3. Commit Changes

```bash
git add .
git commit -m "feat: Description (Issue #N)"
```

### 4. Push and Create PR

```bash
git push -u origin feature/issue-N-short-description
```

Then create a PR on GitHub with:
- Clear title matching commit convention
- Description of changes
- Link to related issue
- Test plan

### 5. PR Requirements

- [ ] All tests pass
- [ ] No linting errors
- [ ] Code follows style guidelines
- [ ] Documentation updated (if applicable)
- [ ] Commit messages follow conventions

## Issue Guidelines

### Before Creating an Issue

1. Search existing issues to avoid duplicates
2. Check if it's already fixed in `main`

### Creating an Issue

Include:
- **Clear title**: Descriptive summary
- **Description**: What you expected vs. what happened
- **Reproduction steps**: Minimal example to reproduce
- **Environment**: Go version, OS, CLI version
- **Proposed solution**: (Optional) How you think it should be fixed

### Issue Labels

| Label | Description |
|-------|-------------|
| `bug` | Something isn't working |
| `enhancement` | New feature request |
| `docs` | Documentation improvement |
| `good first issue` | Good for newcomers |

## Architecture

See [ARCHITECTURE.md](ARCHITECTURE.md) for system design documentation.

## Questions?

- Open a [GitHub Issue](https://github.com/severity1/claude-agent-sdk-go/issues)
- Check existing [documentation](docs/)
