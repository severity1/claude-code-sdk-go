# Project Structure Analysis

Analysis of Python SDK project configuration, build system, and development workflow.

## pyproject.toml - Project Configuration

**Key Findings:**
- **Version**: 0.0.20 (current as of analysis)
- **Python Requirements**: >=3.10, supports up to 3.13
- **Build System**: Uses Hatchling (modern Python packaging)
- **License**: MIT
- **Official Anthropic Project**: Published by Anthropic with official support

**Dependencies:**
- **Core**: `anyio>=4.0.0` (async I/O framework, enables trio/asyncio compatibility)
- **Compatibility**: `typing_extensions>=4.0.0` (for Python <3.11 support)

**Development Dependencies:**
- **Testing**: pytest, pytest-asyncio, pytest-cov
- **Type Checking**: mypy with very strict configuration
- **Linting**: ruff with comprehensive rule set
- **Async Frameworks**: anyio[trio] for trio support

**Configuration Quality:**
- **MyPy**: Extremely strict type checking (strict=true, no untyped defs, etc.)
- **Ruff**: Comprehensive linting (pycodestyle, pyflakes, isort, bugbear, etc.)
- **Testing**: Proper test path configuration with importlib mode

## Development Workflow (CLAUDE.md)

**Development Commands:**
- **Linting**: `python -m ruff check src/ tests/ --fix` (auto-fix)
- **Formatting**: `python -m ruff format src/ tests/`
- **Type Checking**: `python -m mypy src/` (strict typing, source only)
- **Testing**: `python -m pytest tests/` (full suite) or specific files

**Architecture Overview:**
```
src/claude_code_sdk/
├── client.py           # ClaudeSDKClient (interactive sessions)
├── query.py            # query() function (one-shot)
├── types.py            # Type definitions
└── _internal/          # Internal implementation
    ├── transport/subprocess_cli.py  # CLI subprocess management
    └── message_parser.py           # Message parsing logic
```

**Development Quality Standards:**
- **Separate tooling for different purposes**: Linting, formatting, type checking, testing
- **Source-only type checking**: Internal types don't get public type validation
- **Auto-fix capabilities**: Development workflow supports automatic fixes
- **Granular testing**: Can run specific test files during development

## Version Management (scripts/update_version.py)

**Version Management Strategy:**
- **Dual Location Updates**: Updates both `pyproject.toml` and `src/claude_code_sdk/__init__.py`
- **Regex-based Updates**: Uses precise regex to avoid unintended changes
- **Safety Features**: `count=1` ensures only one replacement per file
- **CLI Interface**: Simple command-line tool for version bumps

**Implementation Details:**
- **Regex Pattern**: `^version = "[^"]*"` and `^__version__ = "[^"]*"`
- **File Operations**: Direct file read/write with pathlib
- **Error Handling**: Basic argument validation

## Go SDK Implications

### Build System
- **Go modules**: Use standard Go module system (go.mod, go.sum)
- **Version management**: Use git tags for semantic versioning
- **Build tools**: Standard Go toolchain (go build, go test, go fmt)

### Development Workflow
- **Linting**: Use golangci-lint with comprehensive rule set
- **Formatting**: gofmt and goimports for code formatting
- **Testing**: go test with race detection and coverage
- **Type checking**: Go's built-in type system provides compile-time checking

### Quality Standards
- **Zero external dependencies**: Match Python SDK's minimal dependency approach
- **Comprehensive testing**: Unit tests, integration tests, and benchmarks
- **Documentation**: Complete godoc coverage for public APIs
- **CI/CD**: Automated testing across Go versions and platforms

### Project Structure
```
claude-code-sdk-go/
├── go.mod              # Module definition
├── types.go            # Core types and interfaces
├── client.go           # Client implementation
├── query.go            # Query function
├── errors.go           # Error types
├── transport.go        # Transport interface
├── internal/           # Internal implementation
│   ├── subprocess/     # Subprocess transport
│   ├── parser/         # Message parsing
│   └── cli/            # CLI integration
├── examples/           # Usage examples
└── testdata/           # Test fixtures
```

This structure mirrors the Python SDK's logical organization while following Go conventions.