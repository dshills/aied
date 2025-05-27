package commands

import (
	"fmt"
	"os"

	"github.com/dshills/aied/internal/buffer"
)

// WriteCommand implements the :w (write) command
type WriteCommand struct{}

// NewWriteCommand creates a new write command
func NewWriteCommand() *WriteCommand {
	return &WriteCommand{}
}

func (w *WriteCommand) Name() string {
	return "write"
}

func (w *WriteCommand) Aliases() []string {
	return []string{"w"}
}

func (w *WriteCommand) Execute(args []string, buf *buffer.Buffer) CommandResult {
	var filename string
	
	if len(args) > 0 {
		// Use provided filename
		filename = args[0]
	} else {
		// Use buffer's current filename
		filename = buf.Filename()
		if filename == "" {
			return CommandResult{
				Success: false,
				Message: "No file name specified",
			}
		}
	}
	
	// Save the file
	if len(args) > 0 {
		// Save as new filename
		err := buf.SaveAs(filename)
		if err != nil {
			return CommandResult{
				Success: false,
				Message: fmt.Sprintf("Error writing file: %s", err.Error()),
			}
		}
	} else {
		// Save to current filename
		err := buf.Save()
		if err != nil {
			return CommandResult{
				Success: false,
				Message: fmt.Sprintf("Error writing file: %s", err.Error()),
			}
		}
	}
	
	return CommandResult{
		Success: true,
		Message: fmt.Sprintf("File written: %s", filename),
	}
}

func (w *WriteCommand) Help() string {
	return ":w [filename] - Write buffer to file"
}

// QuitCommand implements the :q (quit) command
type QuitCommand struct{}

// NewQuitCommand creates a new quit command
func NewQuitCommand() *QuitCommand {
	return &QuitCommand{}
}

func (q *QuitCommand) Name() string {
	return "quit"
}

func (q *QuitCommand) Aliases() []string {
	return []string{"q"}
}

func (q *QuitCommand) Execute(args []string, buf *buffer.Buffer) CommandResult {
	// Check if buffer has unsaved changes
	if buf.Modified() {
		return CommandResult{
			Success: false,
			Message: "Buffer has unsaved changes. Use :q! to force quit or :wq to save and quit",
		}
	}
	
	return CommandResult{
		Success:    true,
		Message:    "Goodbye!",
		ExitEditor: true,
	}
}

func (q *QuitCommand) Help() string {
	return ":q - Quit editor (fails if unsaved changes)"
}

// ForceQuitCommand implements the :q! (force quit) command
type ForceQuitCommand struct{}

// NewForceQuitCommand creates a new force quit command
func NewForceQuitCommand() *ForceQuitCommand {
	return &ForceQuitCommand{}
}

func (fq *ForceQuitCommand) Name() string {
	return "quit!"
}

func (fq *ForceQuitCommand) Aliases() []string {
	return []string{"q!"}
}

func (fq *ForceQuitCommand) Execute(args []string, buf *buffer.Buffer) CommandResult {
	return CommandResult{
		Success:    true,
		Message:    "Goodbye!",
		ExitEditor: true,
	}
}

func (fq *ForceQuitCommand) Help() string {
	return ":q! - Force quit editor (discards unsaved changes)"
}

// WriteQuitCommand implements the :wq (write and quit) command
type WriteQuitCommand struct{}

// NewWriteQuitCommand creates a new write and quit command
func NewWriteQuitCommand() *WriteQuitCommand {
	return &WriteQuitCommand{}
}

func (wq *WriteQuitCommand) Name() string {
	return "wq"
}

func (wq *WriteQuitCommand) Aliases() []string {
	return []string{"x", "exit"}
}

func (wq *WriteQuitCommand) Execute(args []string, buf *buffer.Buffer) CommandResult {
	// First try to write
	writeCmd := NewWriteCommand()
	writeResult := writeCmd.Execute(args, buf)
	
	if !writeResult.Success {
		return writeResult
	}
	
	// If write succeeded, quit
	return CommandResult{
		Success:    true,
		Message:    "File written and editor closed",
		ExitEditor: true,
	}
}

func (wq *WriteQuitCommand) Help() string {
	return ":wq [filename] - Write buffer and quit editor"
}

// EditCommand implements the :e (edit) command
type EditCommand struct{}

// NewEditCommand creates a new edit command
func NewEditCommand() *EditCommand {
	return &EditCommand{}
}

func (e *EditCommand) Name() string {
	return "edit"
}

func (e *EditCommand) Aliases() []string {
	return []string{"e"}
}

func (e *EditCommand) Execute(args []string, buf *buffer.Buffer) CommandResult {
	if len(args) == 0 {
		// Reload current file
		filename := buf.Filename()
		if filename == "" {
			return CommandResult{
				Success: false,
				Message: "No file to reload",
			}
		}
		
		// Check if buffer has unsaved changes
		if buf.Modified() {
			return CommandResult{
				Success: false,
				Message: "Buffer has unsaved changes. Use :e! to force reload",
			}
		}
		
		// TODO: Reload file into current buffer
		return CommandResult{
			Success: false,
			Message: "File reloading not yet implemented",
		}
	}
	
	filename := args[0]
	
	// Check if buffer has unsaved changes
	if buf.Modified() {
		return CommandResult{
			Success: false,
			Message: "Buffer has unsaved changes. Save first or use :e! to discard",
		}
	}
	
	// Check if file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return CommandResult{
			Success: false,
			Message: fmt.Sprintf("File not found: %s", filename),
		}
	}
	
	// TODO: Load new file into buffer
	// For now, just return a message
	return CommandResult{
		Success: false,
		Message: "File loading not yet implemented",
	}
}

func (e *EditCommand) Help() string {
	return ":e [filename] - Edit file (loads new file or reloads current)"
}

// NewCommand implements the :new command
type NewCommand struct{}

// NewNewCommand creates a new "new" command
func NewNewCommand() *NewCommand {
	return &NewCommand{}
}

func (n *NewCommand) Name() string {
	return "new"
}

func (n *NewCommand) Aliases() []string {
	return []string{"enew"}
}

func (n *NewCommand) Execute(args []string, buf *buffer.Buffer) CommandResult {
	// Check if buffer has unsaved changes
	if buf.Modified() {
		return CommandResult{
			Success: false,
			Message: "Buffer has unsaved changes. Save first",
		}
	}
	
	// TODO: Clear buffer and reset to new file
	// For now, just return a message
	return CommandResult{
		Success: false,
		Message: "New file creation not yet implemented",
	}
}

func (n *NewCommand) Help() string {
	return ":new - Create new empty buffer"
}