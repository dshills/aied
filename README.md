# AIED - AI-Powered Terminal Editor

AIED is a terminal-based text editor with VIM-style key bindings and integrated AI assistance. It supports multiple AI providers including OpenAI, Anthropic, Google Gemini, and local models via Ollama.

## Features

- **VIM-style editing**: Normal, Insert, Visual, and Command modes
- **AI Integration**: Code completion, explanation, refactoring, and chat
- **Multi-provider support**: OpenAI, Anthropic, Google, and Ollama
- **Configurable**: YAML/JSON configuration files
- **Terminal UI**: Built with tcell for cross-platform support

## Installation

```bash
go install github.com/dshills/aied@latest
```

Or build from source:

```bash
git clone https://github.com/dshills/aied.git
cd aied
go build .
```

## Configuration

AIED can be configured using YAML or JSON files. Configuration is loaded from:

1. `.aied.yaml` or `.aied.json` in the current directory
2. `~/.aied.yaml` or `~/.aied.json` in your home directory
3. `~/.config/aied/config.yaml` or `~/.config/aied/config.json`
4. Environment variables (override file settings)

### Generate Example Configuration

```bash
# In the editor
:configgen

# Or specify a path
:configgen myconfig.yaml
```

### Example Configuration

```yaml
editor:
  tab_size: 4
  indent_style: spaces
  line_numbers: true
  theme: default

ai:
  default_provider: ollama
  context_lines: 10
  max_tokens: 1000
  temperature: 0.3

providers:
  - type: openai
    api_key: ${OPENAI_API_KEY}
    model: gpt-4
    enabled: true
    
  - type: anthropic
    api_key: ${ANTHROPIC_API_KEY}
    model: claude-3-5-sonnet-20241022
    enabled: true
    
  - type: ollama
    base_url: http://localhost:11434
    model: llama2
    enabled: true
```

### Environment Variables

AI providers can be configured via environment variables:

- `OPENAI_API_KEY`: OpenAI API key
- `ANTHROPIC_API_KEY`: Anthropic API key
- `GOOGLE_API_KEY`: Google AI API key
- `OLLAMA_BASE_URL`: Ollama server URL (default: http://localhost:11434)
- `OLLAMA_MODEL`: Ollama model name (default: llama2)

## Usage

### Basic Usage

```bash
# Open a file
aied filename.go

# Create new file
aied
```

### VIM Modes

- **Normal Mode**: Navigate and edit with VIM commands
- **Insert Mode**: Type text (press `i` from Normal mode)
- **Visual Mode**: Select text (press `v` from Normal mode)
- **Command Mode**: Execute commands (press `:` from Normal mode)

### AI Commands

- `:ai <question>` - Ask AI a question about your code
- `:aic` - Complete code at cursor position
- `:aie` - Explain code at cursor or selection
- `:air` - Get refactoring suggestions
- `:aip` - List available AI providers
- `:aip <provider>` - Switch to a specific provider

### File Commands

- `:w` - Save current file
- `:q` - Quit editor
- `:wq` - Save and quit
- `:q!` - Force quit without saving
- `:e <filename>` - Open file
- `:new <filename>` - Create new file

### Configuration Commands

- `:config` - Show current configuration
- `:configgen [path]` - Generate example config file
- `:configreload` - Reload configuration from disk

## Building from Source

```bash
# Clone repository
git clone https://github.com/dshills/aied.git
cd aied

# Build
go build .

# Run tests
go test ./...
```

## Requirements

- Go 1.21 or higher
- Terminal with UTF-8 support
- (Optional) Ollama for local AI models
- (Optional) API keys for cloud AI providers

## License

MIT License - see LICENSE file for details