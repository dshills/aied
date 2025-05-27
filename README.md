# AIED - AI-Powered Terminal Editor

<p align="center">
  <strong>A modern terminal text editor with VIM-style keybindings and integrated AI assistance</strong>
</p>

<p align="center">
  <a href="#features">Features</a> •
  <a href="#installation">Installation</a> •
  <a href="#quick-start">Quick Start</a> •
  <a href="#usage">Usage</a> •
  <a href="#configuration">Configuration</a> •
  <a href="#development">Development</a>
</p>

## Overview

AIED (AI-Enhanced Editor) is a terminal-based text editor that combines the power of VIM-style modal editing with modern AI capabilities. It's designed for developers who want the efficiency of VIM with the intelligence of AI assistants, all within a fast, lightweight terminal application.

## Features

### 🎯 Core Editing
- **VIM-style modal editing**: Normal, Insert, Visual, and Command modes
- **Efficient navigation**: h/j/k/l movement, word/line jumps (w/b/e)
- **Text manipulation**: Delete, yank, paste operations
- **File operations**: Open, save, create new files
- **Buffer management**: Line-based editing with undo/redo support

### 🤖 AI Integration
- **Multiple AI providers**: Seamlessly switch between providers
  - OpenAI (GPT-4, GPT-3.5)
  - Anthropic (Claude 3.5 Sonnet)
  - Google (Gemini 1.5)
  - Ollama (Local models)
- **AI-powered commands**:
  - Code completion at cursor position
  - Code explanation and documentation
  - Refactoring suggestions
  - General programming assistance
- **Automatic fallback**: If one provider fails, automatically tries others
- **Context-aware**: AI understands surrounding code context

### ⚙️ Configuration
- **Flexible configuration**: YAML or JSON format
- **Multiple config locations**: Project, user, and system-wide settings
- **Environment variables**: Override settings without changing files
- **Hot reload**: Update configuration without restarting

### 🚀 Performance
- **Fast startup**: Minimal dependencies, quick load times
- **Efficient rendering**: Optimized terminal drawing
- **Low memory footprint**: Suitable for remote development
- **Cross-platform**: Works on Linux, macOS, and Windows

## Installation

### Using Go Install (Recommended)

```bash
go install github.com/dshills/aied@latest
```

### Building from Source

```bash
# Clone the repository
git clone https://github.com/dshills/aied.git
cd aied

# Build the binary
go build -o aied .

# Optionally, move to PATH
sudo mv aied /usr/local/bin/
```

### Requirements

- Go 1.21 or higher (for building)
- Terminal with UTF-8 support
- (Optional) API keys for AI providers
- (Optional) Ollama for local AI models

## Quick Start

1. **Basic usage**:
   ```bash
   # Open a file
   aied myfile.go
   
   # Create a new file
   aied newfile.py
   
   # Open editor without file
   aied
   ```

2. **Generate configuration**:
   ```bash
   # Start aied and run:
   :configgen
   ```

3. **Set up AI provider** (choose one):
   ```bash
   # OpenAI
   export OPENAI_API_KEY="your-key-here"
   
   # Anthropic
   export ANTHROPIC_API_KEY="your-key-here"
   
   # Google
   export GOOGLE_API_KEY="your-key-here"
   
   # Ollama (install from https://ollama.ai)
   ollama pull llama2
   ```

4. **Try AI features**:
   - Ask a question: `:ai how do I sort a slice in go?`
   - Complete code: Position cursor and type `:aic`
   - Explain code: `:aie`

## Usage

### VIM Commands

#### Normal Mode
| Command | Description |
|---------|-------------|
| `h/j/k/l` | Move cursor left/down/up/right |
| `w/b` | Move forward/backward by word |
| `e` | Move to end of word |
| `0/$` | Move to beginning/end of line |
| `gg/G` | Go to first/last line |
| `i` | Enter Insert mode |
| `v` | Enter Visual mode |
| `x` | Delete character |
| `dd` | Delete line |
| `yy` | Yank (copy) line |
| `p` | Paste |
| `u` | Undo |
| `Ctrl-R` | Redo |
| `:` | Enter Command mode |

#### Insert Mode
| Command | Description |
|---------|-------------|
| `Esc` | Return to Normal mode |
| `Backspace` | Delete previous character |
| `Enter` | Insert new line |
| (Type normally) | Insert text |

#### Command Mode
| Command | Description |
|---------|-------------|
| `:w` | Save file |
| `:q` | Quit |
| `:wq` | Save and quit |
| `:q!` | Quit without saving |
| `:e <file>` | Open file |
| `:new <file>` | Create new file |

### AI Commands

| Command | Description | Example |
|---------|-------------|---------|
| `:ai <question>` | Ask AI anything | `:ai what does this function do?` |
| `:aic` | Complete code at cursor | Place cursor after partial code and run `:aic` |
| `:aie` | Explain current line/selection | `:aie` |
| `:air` | Get refactoring suggestions | `:air` |
| `:aip` | List/switch AI providers | `:aip` or `:aip openai` |

### Configuration Commands

| Command | Description |
|---------|-------------|
| `:config` | Show current configuration |
| `:configgen [path]` | Generate example config file |
| `:configreload` | Reload configuration from disk |

## Configuration

### Configuration File Locations

AIED searches for configuration in the following order (first found wins):

1. `.aied.yaml` or `.aied.json` (current directory)
2. `~/.aied.yaml` or `~/.aied.json` (home directory)
3. `~/.config/aied/config.yaml` or `~/.config/aied/config.json`
4. `/etc/aied/config.yaml` or `/etc/aied/config.json` (system-wide)

### Configuration Structure

```yaml
# Editor settings
editor:
  tab_size: 4                    # Number of spaces for tab
  indent_style: spaces           # "spaces" or "tabs"
  line_numbers: true             # Show line numbers
  theme: default                 # Color theme
  auto_save: false               # Auto-save on focus loss
  auto_save_delay: 60            # Seconds before auto-save

# AI settings
ai:
  default_provider: ollama       # Default AI provider
  enable_completion: true        # Enable AI completions
  completion_delay: 500          # Milliseconds before completion
  context_lines: 10              # Lines of context for AI
  max_tokens: 1000               # Max tokens in AI response
  temperature: 0.3               # AI creativity (0.0-1.0)
  enabled_commands:              # Which AI commands to enable
    - ai
    - aic
    - aie
    - air
    - aip

# AI Provider configurations
providers:
  # OpenAI
  - type: openai
    api_key: ${OPENAI_API_KEY}   # Environment variable reference
    model: gpt-4
    base_url: https://api.openai.com/v1
    enabled: true
    options:
      timeout: 30

  # Anthropic Claude
  - type: anthropic
    api_key: ${ANTHROPIC_API_KEY}
    model: claude-3-5-sonnet-20241022
    base_url: https://api.anthropic.com/v1
    enabled: true

  # Google Gemini
  - type: google
    api_key: ${GOOGLE_API_KEY}
    model: gemini-1.5-flash
    base_url: https://generativelanguage.googleapis.com/v1beta/models
    enabled: false

  # Ollama (Local)
  - type: ollama
    base_url: http://localhost:11434
    model: llama2
    enabled: true
```

### Environment Variables

Environment variables override configuration file settings:

```bash
# AI Provider Keys
export OPENAI_API_KEY="sk-..."
export ANTHROPIC_API_KEY="sk-ant-..."
export GOOGLE_API_KEY="AIza..."

# Ollama Settings
export OLLAMA_BASE_URL="http://localhost:11434"
export OLLAMA_MODEL="codellama"
```

## Development

### Project Structure

```
aied/
├── cmd/                    # Command-line interface
├── internal/              # Internal packages
│   ├── ai/               # AI provider implementations
│   ├── buffer/           # Text buffer management
│   ├── commands/         # Ex commands (:w, :q, etc.)
│   ├── config/           # Configuration management
│   ├── modes/            # VIM modes (normal, insert, etc.)
│   └── ui/               # Terminal UI rendering
├── .aied.yaml.example    # Example configuration
├── go.mod               # Go modules
└── main.go              # Entry point
```

### Building and Testing

```bash
# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Build for current platform
go build -o aied .

# Build for multiple platforms
GOOS=linux GOARCH=amd64 go build -o aied-linux-amd64
GOOS=darwin GOARCH=amd64 go build -o aied-darwin-amd64
GOOS=windows GOARCH=amd64 go build -o aied-windows-amd64.exe
```

### Adding New AI Providers

1. Implement the `Provider` interface in `internal/ai/`:
```go
type Provider interface {
    Name() ProviderType
    IsAvailable() bool
    Complete(ctx context.Context, req AIRequest) (*AIResponse, error)
    Chat(ctx context.Context, req AIRequest) (*AIResponse, error)
    Analyze(ctx context.Context, req AIRequest) (*AIResponse, error)
    Configure(config ProviderConfig) error
}
```

2. Add provider type to `internal/ai/provider.go`
3. Update factory in `CreateProvider()` function
4. Add configuration support in `internal/config/config.go`

### Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Roadmap

- [x] Core VIM-style editing
- [x] AI provider integration
- [x] Configuration system
- [ ] Syntax highlighting
- [ ] Multiple buffers/windows
- [ ] Search and replace
- [ ] Macros and registers
- [ ] Plugin system
- [ ] LSP support
- [ ] Git integration

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Built with [tcell](https://github.com/gdamore/tcell) for terminal handling
- Inspired by VIM and modern AI-powered development tools
- Thanks to all contributors and users

## Support

- **Issues**: [GitHub Issues](https://github.com/dshills/aied/issues)
- **Discussions**: [GitHub Discussions](https://github.com/dshills/aied/discussions)
- **Documentation**: [Wiki](https://github.com/dshills/aied/wiki)

---

<p align="center">
  Made with ❤️ by developers, for developers
</p>