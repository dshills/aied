package commands

import (
	"fmt"
	"strings"

	"github.com/dshills/aied/internal/buffer"
)

// CommandResult represents the result of executing a command
type CommandResult struct {
	Success    bool   // Whether the command executed successfully
	Message    string // Success or error message
	ExitEditor bool   // Whether to exit the editor
	SwitchMode bool   // Whether to switch back to Normal mode
}

// Command represents a VIM ex command
type Command interface {
	// Name returns the command name (e.g., "write", "quit")
	Name() string
	
	// Aliases returns alternative names for the command (e.g., ["w"] for write)
	Aliases() []string
	
	// Execute runs the command with the given arguments
	Execute(args []string, buf *buffer.Buffer) CommandResult
	
	// Help returns help text for the command
	Help() string
}

// CommandRegistry manages available commands
type CommandRegistry struct {
	commands map[string]Command
}

// NewCommandRegistry creates a new command registry
func NewCommandRegistry() *CommandRegistry {
	registry := &CommandRegistry{
		commands: make(map[string]Command),
	}
	
	// Register built-in commands
	registry.RegisterCommand(NewWriteCommand())
	registry.RegisterCommand(NewQuitCommand())
	registry.RegisterCommand(NewForceQuitCommand())
	registry.RegisterCommand(NewWriteQuitCommand())
	registry.RegisterCommand(NewEditCommand())
	registry.RegisterCommand(NewNewCommand())
	
	// Register AI commands
	registry.RegisterCommand(NewAICompleteCommand())
	registry.RegisterCommand(NewAIExplainCommand())
	registry.RegisterCommand(NewAIRefactorCommand())
	registry.RegisterCommand(NewAIChatCommand())
	registry.RegisterCommand(NewAIProviderCommand())
	
	return registry
}

// RegisterCommand registers a command and its aliases
func (cr *CommandRegistry) RegisterCommand(cmd Command) {
	// Register main name
	cr.commands[cmd.Name()] = cmd
	
	// Register aliases
	for _, alias := range cmd.Aliases() {
		cr.commands[alias] = cmd
	}
}

// GetCommand returns a command by name or alias
func (cr *CommandRegistry) GetCommand(name string) (Command, bool) {
	cmd, exists := cr.commands[name]
	return cmd, exists
}

// ListCommands returns all registered command names
func (cr *CommandRegistry) ListCommands() []string {
	var names []string
	seen := make(map[string]bool)
	
	for _, cmd := range cr.commands {
		mainName := cmd.Name()
		if !seen[mainName] {
			names = append(names, mainName)
			seen[mainName] = true
		}
	}
	
	return names
}

// CommandParser handles parsing command line input
type CommandParser struct{}

// NewCommandParser creates a new command parser
func NewCommandParser() *CommandParser {
	return &CommandParser{}
}

// ParseCommand parses a command line into command name and arguments
func (cp *CommandParser) ParseCommand(cmdLine string) (string, []string, error) {
	// Remove leading/trailing whitespace
	cmdLine = strings.TrimSpace(cmdLine)
	
	if cmdLine == "" {
		return "", nil, fmt.Errorf("empty command")
	}
	
	// Split into parts
	parts := strings.Fields(cmdLine)
	if len(parts) == 0 {
		return "", nil, fmt.Errorf("empty command")
	}
	
	cmdName := parts[0]
	args := parts[1:]
	
	return cmdName, args, nil
}

// CommandExecutor handles command execution
type CommandExecutor struct {
	registry *CommandRegistry
	parser   *CommandParser
}

// NewCommandExecutor creates a new command executor
func NewCommandExecutor() *CommandExecutor {
	return &CommandExecutor{
		registry: NewCommandRegistry(),
		parser:   NewCommandParser(),
	}
}

// Execute parses and executes a command line
func (ce *CommandExecutor) Execute(cmdLine string, buf *buffer.Buffer) CommandResult {
	// Parse the command line
	cmdName, args, err := ce.parser.ParseCommand(cmdLine)
	if err != nil {
		return CommandResult{
			Success:    false,
			Message:    fmt.Sprintf("Error: %s", err.Error()),
			SwitchMode: true,
		}
	}
	
	// Get the command
	cmd, exists := ce.registry.GetCommand(cmdName)
	if !exists {
		return CommandResult{
			Success:    false,
			Message:    fmt.Sprintf("Unknown command: %s", cmdName),
			SwitchMode: true,
		}
	}
	
	// Execute the command
	result := cmd.Execute(args, buf)
	
	// Always switch back to normal mode unless explicitly requested not to
	if !result.ExitEditor {
		result.SwitchMode = true
	}
	
	return result
}

// GetCommands returns the command registry for introspection
func (ce *CommandExecutor) GetCommands() *CommandRegistry {
	return ce.registry
}