package modes

import (
	"github.com/dshills/aied/internal/buffer"
	"github.com/dshills/aied/internal/commands"
	"github.com/dshills/aied/internal/ui"
)

// CommandMode implements VIM command mode behavior (:commands)
type CommandMode struct {
	commandLine string                    // Current command being typed
	executor    *commands.CommandExecutor // Command executor
	message     string                    // Last command result message
}

// NewCommandMode creates a new command mode instance
func NewCommandMode() *CommandMode {
	return &CommandMode{
		commandLine: "",
		executor:    commands.NewCommandExecutor(),
		message:     "",
	}
}

// Type returns the mode type
func (c *CommandMode) Type() ModeType {
	return ModeCommand
}

// HandleInput processes keyboard input in command mode
func (c *CommandMode) HandleInput(event ui.KeyEvent, buf *buffer.Buffer) ModeResult {
	switch event.Action {
	case ui.KeyActionEscape:
		// Cancel command mode
		c.commandLine = ""
		c.message = ""
		return ModeResult{SwitchToMode: &[]ModeType{ModeNormal}[0], Handled: true}

	case ui.KeyActionEnter:
		// Execute the command
		return c.executeCommand(buf)

	case ui.KeyActionBackspace:
		// Remove last character from command line
		if len(c.commandLine) > 0 {
			c.commandLine = c.commandLine[:len(c.commandLine)-1]
		}
		return ModeResult{Handled: true}

	case ui.KeyActionChar:
		// Add character to command line
		c.commandLine += string(event.Rune)
		return ModeResult{Handled: true}

	case ui.KeyActionCtrlC:
		// Cancel command mode
		c.commandLine = ""
		c.message = ""
		return ModeResult{SwitchToMode: &[]ModeType{ModeNormal}[0], Handled: true}

	default:
		return ModeResult{Handled: false}
	}
}

// executeCommand executes the current command line
func (c *CommandMode) executeCommand(buf *buffer.Buffer) ModeResult {
	if c.commandLine == "" {
		// Empty command, just return to normal mode
		return ModeResult{SwitchToMode: &[]ModeType{ModeNormal}[0], Handled: true}
	}

	// Execute the command
	result := c.executor.Execute(c.commandLine, buf)
	
	// Store the result message
	c.message = result.Message
	
	// Clear the command line
	c.commandLine = ""

	// Handle the result
	if result.ExitEditor {
		return ModeResult{
			ExitEditor: true,
			Handled:    true,
		}
	}

	if result.SwitchMode {
		return ModeResult{
			SwitchToMode: &[]ModeType{ModeNormal}[0],
			Handled:      true,
		}
	}

	// Stay in command mode if not explicitly switching
	return ModeResult{Handled: true}
}

// OnEnter is called when entering command mode
func (c *CommandMode) OnEnter(buf *buffer.Buffer) {
	if buf == nil {
		return
	}
	// Clear any previous command and message
	c.commandLine = ""
	c.message = ""
}

// OnExit is called when leaving command mode
func (c *CommandMode) OnExit(buf *buffer.Buffer) {
	// Clear command line when leaving command mode
	c.commandLine = ""
}

// GetStatusText returns mode-specific status information
func (c *CommandMode) GetStatusText() string {
	return "-- COMMAND --"
}

// GetCommandLine returns the current command line being typed
func (c *CommandMode) GetCommandLine() string {
	return ":" + c.commandLine
}

// GetMessage returns the last command result message
func (c *CommandMode) GetMessage() string {
	return c.message
}

// ClearMessage clears the stored message
func (c *CommandMode) ClearMessage() {
	c.message = ""
}