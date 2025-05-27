package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/dshills/aied/internal/ai"
	"gopkg.in/yaml.v3"
)

// Config represents the main configuration structure
type Config struct {
	Editor    EditorConfig              `yaml:"editor" json:"editor"`
	Providers []ai.ProviderConfig       `yaml:"providers" json:"providers"`
	AI        AIConfig                  `yaml:"ai" json:"ai"`
	LSP       LSPConfig                 `yaml:"lsp" json:"lsp"`
}

// EditorConfig holds editor-specific settings
type EditorConfig struct {
	TabSize      int    `yaml:"tab_size" json:"tab_size"`
	IndentStyle  string `yaml:"indent_style" json:"indent_style"` // "tabs" or "spaces"
	LineNumbers  bool   `yaml:"line_numbers" json:"line_numbers"`
	Theme        string `yaml:"theme" json:"theme"`
	AutoSave     bool   `yaml:"auto_save" json:"auto_save"`
	AutoSaveDelay int   `yaml:"auto_save_delay" json:"auto_save_delay"` // seconds
}

// AIConfig holds AI-specific settings
type AIConfig struct {
	DefaultProvider     string   `yaml:"default_provider" json:"default_provider"`
	EnableCompletion    bool     `yaml:"enable_completion" json:"enable_completion"`
	CompletionDelay     int      `yaml:"completion_delay" json:"completion_delay"` // milliseconds
	ContextLines        int      `yaml:"context_lines" json:"context_lines"`
	MaxTokens           int      `yaml:"max_tokens" json:"max_tokens"`
	Temperature         float64  `yaml:"temperature" json:"temperature"`
	EnabledCommands     []string `yaml:"enabled_commands" json:"enabled_commands"`
}

// LSPConfig holds LSP-specific settings
type LSPConfig struct {
	Enabled          bool              `yaml:"enabled" json:"enabled"`
	AutoStart        bool              `yaml:"auto_start" json:"auto_start"`
	ShowDiagnostics  bool              `yaml:"show_diagnostics" json:"show_diagnostics"`
	CompletionTrigger string           `yaml:"completion_trigger" json:"completion_trigger"` // "auto" or "manual"
	Servers          []LSPServerConfig `yaml:"servers" json:"servers"`
}

// LSPServerConfig holds configuration for a specific LSP server
type LSPServerConfig struct {
	Name       string            `yaml:"name" json:"name"`
	Command    string            `yaml:"command" json:"command"`
	Args       []string          `yaml:"args" json:"args"`
	Languages  []string          `yaml:"languages" json:"languages"`
	Extensions []string          `yaml:"extensions" json:"extensions"`
	Enabled    bool              `yaml:"enabled" json:"enabled"`
	Settings   map[string]interface{} `yaml:"settings" json:"settings"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Editor: EditorConfig{
			TabSize:       4,
			IndentStyle:   "spaces",
			LineNumbers:   true,
			Theme:         "default",
			AutoSave:      false,
			AutoSaveDelay: 60,
		},
		AI: AIConfig{
			DefaultProvider:  "ollama",
			EnableCompletion: true,
			CompletionDelay:  500,
			ContextLines:     10,
			MaxTokens:        1000,
			Temperature:      0.3,
			EnabledCommands:  []string{"ai", "aic", "aie", "air", "aip"},
		},
		Providers: []ai.ProviderConfig{
			{
				Type:    ai.ProviderOllama,
				BaseURL: "http://localhost:11434",
				Model:   "llama2",
				Enabled: true,
			},
		},
		LSP: LSPConfig{
			Enabled:          true,
			AutoStart:        true,
			ShowDiagnostics:  true,
			CompletionTrigger: "manual",
			Servers: []LSPServerConfig{
				{
					Name:       "gopls",
					Command:    "gopls",
					Languages:  []string{"go"},
					Extensions: []string{"go"},
					Enabled:    true,
				},
			},
		},
	}
}

// ConfigPaths returns the search paths for config files
func ConfigPaths() []string {
	var paths []string
	
	// Current directory
	if cwd, err := os.Getwd(); err == nil {
		paths = append(paths, filepath.Join(cwd, ".aied.yaml"))
		paths = append(paths, filepath.Join(cwd, ".aied.json"))
	}
	
	// Home directory
	if home, err := os.UserHomeDir(); err == nil {
		paths = append(paths, filepath.Join(home, ".aied.yaml"))
		paths = append(paths, filepath.Join(home, ".aied.json"))
		paths = append(paths, filepath.Join(home, ".config", "aied", "config.yaml"))
		paths = append(paths, filepath.Join(home, ".config", "aied", "config.json"))
	}
	
	// System-wide (Unix-like systems)
	paths = append(paths, "/etc/aied/config.yaml")
	paths = append(paths, "/etc/aied/config.json")
	
	return paths
}

// Load loads configuration from file system
func Load() (*Config, error) {
	config := DefaultConfig()
	
	// Try each config path in order
	for _, path := range ConfigPaths() {
		if err := loadFromFile(path, config); err == nil {
			// Successfully loaded config
			break
		}
	}
	
	// Override with environment variables
	loadFromEnv(config)
	
	return config, nil
}

// LoadFromFile loads configuration from a specific file
func LoadFromFile(path string) (*Config, error) {
	config := DefaultConfig()
	if err := loadFromFile(path, config); err != nil {
		return nil, err
	}
	loadFromEnv(config)
	return config, nil
}

// LoadFromFileOnly loads configuration from a specific file without env overrides
func LoadFromFileOnly(path string) (*Config, error) {
	config := DefaultConfig()
	if err := loadFromFile(path, config); err != nil {
		return nil, err
	}
	return config, nil
}

// loadFromFile loads configuration from a file
func loadFromFile(path string, config *Config) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	
	ext := filepath.Ext(path)
	switch ext {
	case ".yaml", ".yml":
		return yaml.Unmarshal(data, config)
	case ".json":
		return json.Unmarshal(data, config)
	default:
		return fmt.Errorf("unsupported config file format: %s", ext)
	}
}

// loadFromEnv loads configuration from environment variables
func loadFromEnv(config *Config) {
	// Check for API keys in environment
	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
		addOrUpdateProvider(config, ai.ProviderConfig{
			Type:    ai.ProviderOpenAI,
			APIKey:  apiKey,
			Model:   "gpt-4",
			Enabled: true,
		})
	}
	
	if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
		addOrUpdateProvider(config, ai.ProviderConfig{
			Type:    ai.ProviderAnthropic,
			APIKey:  apiKey,
			Model:   "claude-3-5-sonnet-20241022",
			Enabled: true,
		})
	}
	
	if apiKey := os.Getenv("GOOGLE_API_KEY"); apiKey != "" {
		addOrUpdateProvider(config, ai.ProviderConfig{
			Type:    ai.ProviderGoogle,
			APIKey:  apiKey,
			Model:   "gemini-1.5-flash",
			Enabled: true,
		})
	}
	
	// Check for Ollama configuration
	if baseURL := os.Getenv("OLLAMA_BASE_URL"); baseURL != "" {
		addOrUpdateProvider(config, ai.ProviderConfig{
			Type:    ai.ProviderOllama,
			BaseURL: baseURL,
			Model:   getEnvOrDefault("OLLAMA_MODEL", "llama2"),
			Enabled: true,
		})
	}
}

// addOrUpdateProvider adds or updates a provider configuration
func addOrUpdateProvider(config *Config, provider ai.ProviderConfig) {
	for i, p := range config.Providers {
		if p.Type == provider.Type {
			// Update existing provider
			if provider.APIKey != "" {
				config.Providers[i].APIKey = provider.APIKey
			}
			if provider.BaseURL != "" {
				config.Providers[i].BaseURL = provider.BaseURL
			}
			if provider.Model != "" {
				config.Providers[i].Model = provider.Model
			}
			config.Providers[i].Enabled = provider.Enabled
			return
		}
	}
	// Add new provider
	config.Providers = append(config.Providers, provider)
}

// getEnvOrDefault gets an environment variable or returns a default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Save saves the configuration to a file
func (c *Config) Save(path string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	
	var data []byte
	var err error
	
	ext := filepath.Ext(path)
	switch ext {
	case ".yaml", ".yml":
		data, err = yaml.Marshal(c)
	case ".json":
		data, err = json.MarshalIndent(c, "", "  ")
	default:
		return fmt.Errorf("unsupported config file format: %s", ext)
	}
	
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	
	return os.WriteFile(path, data, 0644)
}

// GenerateExample generates an example configuration file
func GenerateExample(path string) error {
	config := &Config{
		Editor: EditorConfig{
			TabSize:       4,
			IndentStyle:   "spaces",
			LineNumbers:   true,
			Theme:         "monokai",
			AutoSave:      true,
			AutoSaveDelay: 30,
		},
		AI: AIConfig{
			DefaultProvider:  "openai",
			EnableCompletion: true,
			CompletionDelay:  300,
			ContextLines:     15,
			MaxTokens:        1500,
			Temperature:      0.2,
			EnabledCommands:  []string{"ai", "aic", "aie", "air", "aip"},
		},
		Providers: []ai.ProviderConfig{
			{
				Type:    ai.ProviderOpenAI,
				APIKey:  "your-openai-api-key-here",
				Model:   "gpt-4",
				BaseURL: "https://api.openai.com/v1",
				Enabled: true,
				Options: map[string]interface{}{
					"timeout": 30,
				},
			},
			{
				Type:    ai.ProviderAnthropic,
				APIKey:  "your-anthropic-api-key-here", 
				Model:   "claude-3-5-sonnet-20241022",
				BaseURL: "https://api.anthropic.com/v1",
				Enabled: true,
			},
			{
				Type:    ai.ProviderGoogle,
				APIKey:  "your-google-api-key-here",
				Model:   "gemini-1.5-flash",
				BaseURL: "https://generativelanguage.googleapis.com/v1beta/models",
				Enabled: false,
			},
			{
				Type:    ai.ProviderOllama,
				BaseURL: "http://localhost:11434",
				Model:   "llama2",
				Enabled: true,
			},
		},
		LSP: LSPConfig{
			Enabled:          true,
			AutoStart:        true,
			ShowDiagnostics:  true,
			CompletionTrigger: "manual",
			Servers: []LSPServerConfig{
				{
					Name:       "gopls",
					Command:    "gopls",
					Languages:  []string{"go"},
					Extensions: []string{"go"},
					Enabled:    true,
				},
				{
					Name:       "rust-analyzer",
					Command:    "rust-analyzer",
					Languages:  []string{"rust"},
					Extensions: []string{"rs"},
					Enabled:    false,
				},
				{
					Name:       "pyright",
					Command:    "pyright-langserver",
					Args:       []string{"--stdio"},
					Languages:  []string{"python"},
					Extensions: []string{"py"},
					Enabled:    false,
				},
				{
					Name:       "typescript-language-server",
					Command:    "typescript-language-server",
					Args:       []string{"--stdio"},
					Languages:  []string{"javascript", "typescript"},
					Extensions: []string{"js", "ts", "jsx", "tsx"},
					Enabled:    false,
				},
			},
		},
	}
	
	return config.Save(path)
}