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
├── main.go                 # Entry point with complete VIM editor
├── internal/              # Private application code
│   ├── buffer/           # ✅ Text buffer management (complete)
│   ├── ui/              # ✅ Terminal UI components (complete)
│   ├── modes/           # ✅ VIM mode implementations (complete)
│   ├── commands/        # ✅ VIM command system (complete)
│   ├── editor/           # Core editor logic
│   └── ai/              # AI integration layer
├── pkg/                  # Public packages
│   ├── config/          # Configuration handling
│   └── plugin/          # Plugin API
└── go.mod               # Go module file
```

### Completed Components

#### Buffer System (`internal/buffer/`)
- Complete text buffer with line-based storage
- Cursor position management with bounds checking
- Character and line operations (insert, delete, backspace)
- File I/O with proper error handling
- 75.2% test coverage

#### Terminal UI (`internal/ui/`)
- tcell-based terminal interface
- Screen management and rendering
- Keyboard input processing and event handling
- Viewport scrolling to keep cursor visible
- Status line with file info and cursor position
- Mode-aware rendering and command line display
- Comprehensive test coverage

#### VIM Modes System (`internal/modes/`)
- Complete modal editing system with Normal, Insert, Visual, and Command modes
- Mode manager with seamless mode transitions
- VIM-compatible navigation (hjkl, word movement, line operations)
- Mode-specific key bindings and behaviors
- Status line integration showing current mode
- Comprehensive test coverage for all modes and transitions

#### VIM Command System (`internal/commands/`)
- Complete ex-command implementation with Command mode
- Essential file commands: :w (write), :q (quit), :wq (write-quit), :q! (force quit)
- Editor commands: :e (edit), :new (new file) with proper error handling
- Command parsing and execution with comprehensive error messages
- Command registry for extensibility and alias support
- Real-time command line display and feedback
- Comprehensive test coverage for all commands and error cases

## VIM Compatibility Goals

- Support essential VIM motions and commands
- Maintain familiar keybindings where possible
- Compatible configuration syntax for easy migration