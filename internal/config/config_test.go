package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dshills/aied/internal/ai"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	
	// Test editor defaults
	if cfg.Editor.TabSize != 4 {
		t.Errorf("Expected tab size 4, got %d", cfg.Editor.TabSize)
	}
	if cfg.Editor.IndentStyle != "spaces" {
		t.Errorf("Expected indent style 'spaces', got %s", cfg.Editor.IndentStyle)
	}
	if !cfg.Editor.LineNumbers {
		t.Error("Expected line numbers to be enabled by default")
	}
	
	// Test AI defaults
	if cfg.AI.DefaultProvider != "ollama" {
		t.Errorf("Expected default provider 'ollama', got %s", cfg.AI.DefaultProvider)
	}
	if cfg.AI.ContextLines != 10 {
		t.Errorf("Expected context lines 10, got %d", cfg.AI.ContextLines)
	}
	
	// Test default provider
	if len(cfg.Providers) != 1 {
		t.Errorf("Expected 1 default provider, got %d", len(cfg.Providers))
	}
	if cfg.Providers[0].Type != ai.ProviderOllama {
		t.Errorf("Expected default provider to be Ollama, got %s", cfg.Providers[0].Type)
	}
}

func TestLoadFromFile(t *testing.T) {
	// Create temporary test files
	tempDir := t.TempDir()
	
	// Test YAML config
	yamlPath := filepath.Join(tempDir, "test.yaml")
	yamlContent := `
editor:
  tab_size: 2
  indent_style: tabs
ai:
  default_provider: openai
  context_lines: 20
providers:
  - type: openai
    api_key: test-key
    model: gpt-3.5-turbo
    enabled: true
`
	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
		t.Fatal(err)
	}
	
	cfg, err := LoadFromFileOnly(yamlPath)
	if err != nil {
		t.Fatal(err)
	}
	
	if cfg.Editor.TabSize != 2 {
		t.Errorf("Expected tab size 2, got %d", cfg.Editor.TabSize)
	}
	if cfg.Editor.IndentStyle != "tabs" {
		t.Errorf("Expected indent style 'tabs', got %s", cfg.Editor.IndentStyle)
	}
	if cfg.AI.DefaultProvider != "openai" {
		t.Errorf("Expected default provider 'openai', got %s", cfg.AI.DefaultProvider)
	}
	if cfg.AI.ContextLines != 20 {
		t.Errorf("Expected context lines 20, got %d", cfg.AI.ContextLines)
	}
	
	// Check provider was loaded
	found := false
	for _, p := range cfg.Providers {
		if p.Type == ai.ProviderOpenAI && p.APIKey == "test-key" {
			found = true
			break
		}
	}
	if !found {
		t.Error("OpenAI provider not found or API key not set")
	}
}

func TestLoadFromEnv(t *testing.T) {
	// Set test environment variables
	os.Setenv("OPENAI_API_KEY", "test-openai-key")
	os.Setenv("ANTHROPIC_API_KEY", "test-anthropic-key")
	defer os.Unsetenv("OPENAI_API_KEY")
	defer os.Unsetenv("ANTHROPIC_API_KEY")
	
	cfg := DefaultConfig()
	loadFromEnv(cfg)
	
	// Check OpenAI provider was added
	openaiFound := false
	for _, p := range cfg.Providers {
		if p.Type == ai.ProviderOpenAI && p.APIKey == "test-openai-key" {
			openaiFound = true
		}
	}
	if !openaiFound {
		t.Error("OpenAI provider not added from environment")
	}
	
	// Check Anthropic provider was added
	anthropicFound := false
	for _, p := range cfg.Providers {
		if p.Type == ai.ProviderAnthropic && p.APIKey == "test-anthropic-key" {
			anthropicFound = true
		}
	}
	if !anthropicFound {
		t.Error("Anthropic provider not added from environment")
	}
}

func TestSaveConfig(t *testing.T) {
	tempDir := t.TempDir()
	
	// Test saving YAML
	yamlPath := filepath.Join(tempDir, "save-test.yaml")
	cfg := DefaultConfig()
	cfg.Editor.TabSize = 8
	
	if err := cfg.Save(yamlPath); err != nil {
		t.Fatal(err)
	}
	
	// Load it back
	loaded, err := LoadFromFile(yamlPath)
	if err != nil {
		t.Fatal(err)
	}
	
	if loaded.Editor.TabSize != 8 {
		t.Errorf("Expected tab size 8 after save/load, got %d", loaded.Editor.TabSize)
	}
	
	// Test saving JSON
	jsonPath := filepath.Join(tempDir, "save-test.json")
	if err := cfg.Save(jsonPath); err != nil {
		t.Fatal(err)
	}
	
	// Load JSON back
	loaded, err = LoadFromFile(jsonPath)
	if err != nil {
		t.Fatal(err)
	}
	
	if loaded.Editor.TabSize != 8 {
		t.Errorf("Expected tab size 8 after JSON save/load, got %d", loaded.Editor.TabSize)
	}
}

func TestGenerateExample(t *testing.T) {
	tempDir := t.TempDir()
	examplePath := filepath.Join(tempDir, "example.yaml")
	
	if err := GenerateExample(examplePath); err != nil {
		t.Fatal(err)
	}
	
	// Check file exists
	if _, err := os.Stat(examplePath); os.IsNotExist(err) {
		t.Error("Example file was not created")
	}
	
	// Try to load it
	cfg, err := LoadFromFile(examplePath)
	if err != nil {
		t.Fatal(err)
	}
	
	// Verify it has multiple providers
	if len(cfg.Providers) < 4 {
		t.Errorf("Expected at least 4 providers in example, got %d", len(cfg.Providers))
	}
}