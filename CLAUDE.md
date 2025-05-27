# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

AIED (AI Editor) is a terminal-based text editor inspired by VIM with integrated AI capabilities. It combines the efficiency and power of modal editing with modern AI assistance for developers.

## Key Features & Design Goals

- **VIM-style modal editing**: Support for normal, insert, visual, and command modes
- **AI integration**: Built-in AI assistance for code completion, refactoring, and intelligent suggestions
- **Terminal-based**: Runs entirely in the terminal for maximum portability and efficiency
- **Extensible**: Plugin architecture for customization

## Development Setup

### Build Commands
- `go build` - Build the aied binary
- `go build -o aied` - Build with specific output name
- `go install` - Install aied to $GOPATH/bin

### Test Commands
- `go test ./...` - Run all tests
- `go test -v ./...` - Run tests with verbose output
- `go test -cover ./...` - Run tests with coverage

### Development Commands
- `go run main.go` - Run the editor directly
- `go fmt ./...` - Format all Go files
- `go vet ./...` - Run static analysis
- `golangci-lint run` - Run comprehensive linting (if installed)

## Architecture Considerations

### Core Components (to be implemented)
- **Editor Core**: Buffer management, text manipulation, undo/redo
- **Modal System**: VIM-compatible mode handling (normal, insert, visual, command-line)
- **AI Integration Layer**: Interface for AI features (completion, suggestions, refactoring)
- **Terminal UI**: Rendering engine, status line, command palette
- **Command System**: Ex commands, key mappings, configuration
- **Plugin System**: Extension API for custom functionality

### Technical Stack
- **Language**: Go (for performance, concurrency, and simplicity)
- **Terminal UI**: Consider tcell, termbox-go, or bubbletea for terminal handling
- **AI Integration**: 
  - OpenAI API for cloud-based AI features
  - Consider local models via ollama or similar for offline capability
- **Configuration**: YAML/TOML for main config with VIM-compatible command support

### Project Structure
```
aied/
├── main.go                 # Entry point
├── internal/              # Private application code
│   ├── editor/           # Core editor logic
│   ├── ui/              # Terminal UI components
│   ├── modes/           # VIM mode implementations
│   ├── ai/              # AI integration layer
│   ├── buffer/          # Text buffer management
│   └── commands/        # Command system
├── pkg/                  # Public packages
│   ├── config/          # Configuration handling
│   └── plugin/          # Plugin API
└── go.mod               # Go module file
```

## VIM Compatibility Goals

- Support essential VIM motions and commands
- Maintain familiar keybindings where possible
- Compatible configuration syntax for easy migration