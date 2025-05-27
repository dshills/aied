package lsp

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"go.lsp.dev/protocol"
)

// ServerConfig represents configuration for a language server
type ServerConfig struct {
	Name       string   // Server name (e.g., "gopls", "rust-analyzer")
	Command    string   // Command to start the server
	Args       []string // Command arguments
	Languages  []string // Language IDs this server handles
	Extensions []string // File extensions this server handles
}

// Manager manages multiple LSP clients for different languages
type Manager struct {
	mu            sync.RWMutex
	clients       map[string]*Client // keyed by server name
	langToServer  map[string]string  // language ID to server name
	extToServer   map[string]string  // file extension to server name
	configs       []ServerConfig
	rootPath      string
	
	// Callbacks
	onDiagnostics func(filename string, diagnostics []protocol.Diagnostic)
}

// NewManager creates a new LSP manager
func NewManager(rootPath string) *Manager {
	return &Manager{
		clients:      make(map[string]*Client),
		langToServer: make(map[string]string),
		extToServer:  make(map[string]string),
		rootPath:     rootPath,
	}
}

// Configure adds server configurations
func (m *Manager) Configure(configs []ServerConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.configs = configs
	
	// Build language and extension mappings
	for _, config := range configs {
		for _, lang := range config.Languages {
			m.langToServer[lang] = config.Name
		}
		for _, ext := range config.Extensions {
			m.extToServer[ext] = config.Name
		}
	}
}

// Start starts a specific language server
func (m *Manager) Start(ctx context.Context, serverName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// Check if already started
	if _, exists := m.clients[serverName]; exists {
		return nil
	}
	
	// Find config
	var config *ServerConfig
	for _, cfg := range m.configs {
		if cfg.Name == serverName {
			config = &cfg
			break
		}
	}
	
	if config == nil {
		return fmt.Errorf("no configuration found for server %s", serverName)
	}
	
	// Create and start client
	client := NewClient(serverName, m.rootPath)
	
	// Set diagnostics handler
	client.SetDiagnosticsHandler(func(filename string, diagnostics []protocol.Diagnostic) {
		if m.onDiagnostics != nil {
			m.onDiagnostics(filename, diagnostics)
		}
	})
	
	if err := client.Start(ctx, config.Command, config.Args...); err != nil {
		return fmt.Errorf("failed to start %s: %w", serverName, err)
	}
	
	m.clients[serverName] = client
	return nil
}

// Stop stops a specific language server
func (m *Manager) Stop(serverName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	client, exists := m.clients[serverName]
	if !exists {
		return nil
	}
	
	delete(m.clients, serverName)
	return client.Stop()
}

// StopAll stops all language servers
func (m *Manager) StopAll() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	for name, client := range m.clients {
		client.Stop()
		delete(m.clients, name)
	}
}

// GetClient returns the appropriate client for a file
func (m *Manager) GetClient(filename string) (*Client, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	// Determine server based on file extension
	ext := filepath.Ext(filename)
	if ext != "" {
		ext = ext[1:] // Remove leading dot
	}
	
	serverName, exists := m.extToServer[ext]
	if !exists {
		return nil, fmt.Errorf("no language server configured for extension .%s", ext)
	}
	
	client, exists := m.clients[serverName]
	if !exists {
		return nil, fmt.Errorf("language server %s not started", serverName)
	}
	
	return client, nil
}

// GetClientByLanguage returns the appropriate client for a language ID
func (m *Manager) GetClientByLanguage(languageID string) (*Client, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	serverName, exists := m.langToServer[languageID]
	if !exists {
		return nil, fmt.Errorf("no language server configured for language %s", languageID)
	}
	
	client, exists := m.clients[serverName]
	if !exists {
		return nil, fmt.Errorf("language server %s not started", serverName)
	}
	
	return client, nil
}

// OpenFile opens a file in the appropriate language server
func (m *Manager) OpenFile(ctx context.Context, filename string, content string) error {
	client, err := m.GetClient(filename)
	if err != nil {
		return err
	}
	
	languageID := m.getLanguageID(filename)
	return client.OpenFile(ctx, filename, content, languageID)
}

// CloseFile closes a file in the appropriate language server
func (m *Manager) CloseFile(ctx context.Context, filename string) error {
	client, err := m.GetClient(filename)
	if err != nil {
		return err
	}
	
	return client.CloseFile(ctx, filename)
}

// UpdateFile updates a file in the appropriate language server
func (m *Manager) UpdateFile(ctx context.Context, filename string, content string, version int32) error {
	client, err := m.GetClient(filename)
	if err != nil {
		return err
	}
	
	return client.UpdateFile(ctx, filename, content, version)
}

// Completion requests completions for a file position
func (m *Manager) Completion(ctx context.Context, filename string, line, col int) ([]protocol.CompletionItem, error) {
	client, err := m.GetClient(filename)
	if err != nil {
		return nil, err
	}
	
	return client.GetCompletion(ctx, filename, uint32(line), uint32(col))
}

// Hover requests hover information for a file position
func (m *Manager) Hover(ctx context.Context, filename string, line, col int) (*protocol.Hover, error) {
	client, err := m.GetClient(filename)
	if err != nil {
		return nil, err
	}
	
	return client.GetHover(ctx, filename, uint32(line), uint32(col))
}

// Definition requests definition location for a file position
func (m *Manager) Definition(ctx context.Context, filename string, line, col int) ([]protocol.Location, error) {
	client, err := m.GetClient(filename)
	if err != nil {
		return nil, err
	}
	
	return client.GetDefinition(ctx, filename, uint32(line), uint32(col))
}

// References finds all references to a symbol
func (m *Manager) References(ctx context.Context, filename string, line, col int) ([]protocol.Location, error) {
	// TODO: Implement references when client supports it
	return nil, fmt.Errorf("references not implemented")
}

// Rename renames a symbol
func (m *Manager) Rename(ctx context.Context, filename string, line, col int, newName string) (*protocol.WorkspaceEdit, error) {
	// TODO: Implement rename when client supports it
	return nil, fmt.Errorf("rename not implemented")
}

// GetDiagnostics returns diagnostics for a file
func (m *Manager) GetDiagnostics(filename string) []protocol.Diagnostic {
	client, err := m.GetClient(filename)
	if err != nil {
		return nil
	}
	
	return client.GetDiagnostics(filename)
}

// SetDiagnosticsHandler sets the callback for diagnostics
func (m *Manager) SetDiagnosticsHandler(handler func(filename string, diagnostics []protocol.Diagnostic)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onDiagnostics = handler
}

// getLanguageID determines the language ID for a file
func (m *Manager) getLanguageID(filename string) string {
	ext := filepath.Ext(filename)
	if ext != "" {
		ext = ext[1:] // Remove leading dot
	}
	
	// Common language ID mappings
	switch strings.ToLower(ext) {
	case "go":
		return "go"
	case "rs":
		return "rust"
	case "py":
		return "python"
	case "js", "mjs":
		return "javascript"
	case "jsx":
		return "javascriptreact"
	case "ts", "mts":
		return "typescript"
	case "tsx":
		return "typescriptreact"
	case "c":
		return "c"
	case "cc", "cpp", "cxx", "c++":
		return "cpp"
	case "h", "hpp", "hxx", "h++":
		return "cpp"
	case "java":
		return "java"
	case "rb":
		return "ruby"
	case "php":
		return "php"
	case "cs":
		return "csharp"
	case "swift":
		return "swift"
	case "kt":
		return "kotlin"
	case "scala":
		return "scala"
	case "clj":
		return "clojure"
	case "ex", "exs":
		return "elixir"
	case "erl", "hrl":
		return "erlang"
	case "hs":
		return "haskell"
	case "ml", "mli":
		return "ocaml"
	case "vim":
		return "vim"
	case "lua":
		return "lua"
	case "sh", "bash":
		return "shellscript"
	case "yaml", "yml":
		return "yaml"
	case "json":
		return "json"
	case "xml":
		return "xml"
	case "html", "htm":
		return "html"
	case "css":
		return "css"
	case "scss":
		return "scss"
	case "less":
		return "less"
	case "sql":
		return "sql"
	case "md", "markdown":
		return "markdown"
	case "tex":
		return "latex"
	case "r":
		return "r"
	case "m":
		return "matlab"
	case "jl":
		return "julia"
	case "nim":
		return "nim"
	case "zig":
		return "zig"
	case "dart":
		return "dart"
	case "toml":
		return "toml"
	case "ini":
		return "ini"
	case "dockerfile":
		return "dockerfile"
	case "makefile", "mk":
		return "makefile"
	default:
		return "plaintext"
	}
}

// DefaultConfigs returns default LSP server configurations
func DefaultConfigs() []ServerConfig {
	return []ServerConfig{
		{
			Name:       "gopls",
			Command:    "gopls",
			Languages:  []string{"go"},
			Extensions: []string{"go"},
		},
		{
			Name:       "rust-analyzer",
			Command:    "rust-analyzer",
			Languages:  []string{"rust"},
			Extensions: []string{"rs"},
		},
		{
			Name:       "pyright",
			Command:    "pyright-langserver",
			Args:       []string{"--stdio"},
			Languages:  []string{"python"},
			Extensions: []string{"py", "pyi"},
		},
		{
			Name:       "typescript-language-server",
			Command:    "typescript-language-server",
			Args:       []string{"--stdio"},
			Languages:  []string{"javascript", "javascriptreact", "typescript", "typescriptreact"},
			Extensions: []string{"js", "jsx", "ts", "tsx", "mjs", "mts"},
		},
		{
			Name:       "clangd",
			Command:    "clangd",
			Languages:  []string{"c", "cpp"},
			Extensions: []string{"c", "cc", "cpp", "cxx", "h", "hpp"},
		},
		{
			Name:       "lua-language-server",
			Command:    "lua-language-server",
			Languages:  []string{"lua"},
			Extensions: []string{"lua"},
		},
		{
			Name:       "bash-language-server",
			Command:    "bash-language-server",
			Args:       []string{"start"},
			Languages:  []string{"shellscript"},
			Extensions: []string{"sh", "bash"},
		},
		{
			Name:       "yaml-language-server",
			Command:    "yaml-language-server",
			Args:       []string{"--stdio"},
			Languages:  []string{"yaml"},
			Extensions: []string{"yaml", "yml"},
		},
		{
			Name:       "json-language-server",
			Command:    "vscode-json-language-server",
			Args:       []string{"--stdio"},
			Languages:  []string{"json", "jsonc"},
			Extensions: []string{"json", "jsonc"},
		},
	}
}