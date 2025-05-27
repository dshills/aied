package commands

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/dshills/aied/internal/ai"
	"github.com/dshills/aied/internal/buffer"
	"github.com/dshills/aied/internal/config"
)

// ConfigGenerateCommand generates an example configuration file
type ConfigGenerateCommand struct{}

func NewConfigGenerateCommand() *ConfigGenerateCommand {
	return &ConfigGenerateCommand{}
}

func (c *ConfigGenerateCommand) Name() string {
	return "configgen"
}

func (c *ConfigGenerateCommand) Aliases() []string {
	return []string{"cg"}
}

func (c *ConfigGenerateCommand) Execute(args []string, buf *buffer.Buffer) CommandResult {
	path := ".aied.yaml"
	if len(args) > 0 {
		path = args[0]
	}
	
	// Make sure path has correct extension
	ext := filepath.Ext(path)
	if ext != ".yaml" && ext != ".yml" && ext != ".json" {
		path += ".yaml"
	}
	
	if err := config.GenerateExample(path); err != nil {
		return CommandResult{
			Success:    false,
			Message:    fmt.Sprintf("Failed to generate config: %s", err.Error()),
			SwitchMode: true,
		}
	}
	
	return CommandResult{
		Success:    true,
		Message:    fmt.Sprintf("Generated example config at: %s", path),
		SwitchMode: true,
	}
}

func (c *ConfigGenerateCommand) Help() string {
	return "Generate example configuration file"
}

// ConfigShowCommand shows current configuration
type ConfigShowCommand struct{}

func NewConfigShowCommand() *ConfigShowCommand {
	return &ConfigShowCommand{}
}

func (c *ConfigShowCommand) Name() string {
	return "config"
}

func (c *ConfigShowCommand) Aliases() []string {
	return []string{"cfg"}
}

func (c *ConfigShowCommand) Execute(args []string, buf *buffer.Buffer) CommandResult {
	cfg, err := config.Load()
	if err != nil {
		return CommandResult{
			Success:    false,
			Message:    fmt.Sprintf("Failed to load config: %s", err.Error()),
			SwitchMode: true,
		}
	}
	
	var info strings.Builder
	info.WriteString("Configuration:\n")
	info.WriteString(fmt.Sprintf("  Editor:\n"))
	info.WriteString(fmt.Sprintf("    Tab size: %d\n", cfg.Editor.TabSize))
	info.WriteString(fmt.Sprintf("    Indent style: %s\n", cfg.Editor.IndentStyle))
	info.WriteString(fmt.Sprintf("    Line numbers: %v\n", cfg.Editor.LineNumbers))
	info.WriteString(fmt.Sprintf("  AI:\n"))
	info.WriteString(fmt.Sprintf("    Default provider: %s\n", cfg.AI.DefaultProvider))
	info.WriteString(fmt.Sprintf("    Context lines: %d\n", cfg.AI.ContextLines))
	info.WriteString(fmt.Sprintf("  Providers:\n"))
	
	for _, p := range cfg.Providers {
		status := "disabled"
		if p.Enabled {
			status = "enabled"
		}
		info.WriteString(fmt.Sprintf("    %s: %s (model: %s)\n", p.Type, status, p.Model))
	}
	
	return CommandResult{
		Success:    true,
		Message:    info.String(),
		SwitchMode: true,
	}
}

func (c *ConfigShowCommand) Help() string {
	return "Show current configuration"
}

// ConfigReloadCommand reloads configuration from disk
type ConfigReloadCommand struct{}

func NewConfigReloadCommand() *ConfigReloadCommand {
	return &ConfigReloadCommand{}
}

func (c *ConfigReloadCommand) Name() string {
	return "configreload"
}

func (c *ConfigReloadCommand) Aliases() []string {
	return []string{"cr"}
}

func (c *ConfigReloadCommand) Execute(args []string, buf *buffer.Buffer) CommandResult {
	// Load new config
	cfg, err := config.Load()
	if err != nil {
		return CommandResult{
			Success:    false,
			Message:    fmt.Sprintf("Failed to reload config: %s", err.Error()),
			SwitchMode: true,
		}
	}
	
	// Re-initialize AI providers
	if aiManager != nil {
		aiManager.ConfigureProviders(cfg.Providers)
		
		// Set default provider if specified
		if cfg.AI.DefaultProvider != "" {
			aiManager.SetActiveProvider(ai.ProviderType(cfg.AI.DefaultProvider))
		}
	}
	
	return CommandResult{
		Success:    true,
		Message:    "Configuration reloaded",
		SwitchMode: true,
	}
}

func (c *ConfigReloadCommand) Help() string {
	return "Reload configuration from disk"
}