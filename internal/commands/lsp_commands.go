package commands

import (
	"context"
	"fmt"

	"github.com/dshills/aied/internal/buffer"
	"github.com/dshills/aied/internal/lsp"
)

// Global LSP manager reference
var lspManager *lsp.Manager

// SetLSPManager sets the global LSP manager
func SetLSPManager(manager *lsp.Manager) {
	lspManager = manager
}

// HoverCommand shows hover information at cursor position
type HoverCommand struct{}

func NewHoverCommand() Command {
	return &HoverCommand{}
}

func (c *HoverCommand) Name() string {
	return "hover"
}

func (c *HoverCommand) Aliases() []string {
	return []string{"lsp-hover"}
}

func (c *HoverCommand) Execute(args []string, buf *buffer.Buffer) CommandResult {
	if lspManager == nil {
		return CommandResult{
			Success: false,
			Message: "LSP not available",
		}
	}
	
	if buf.Filename() == "" {
		return CommandResult{
			Success: false,
			Message: "No file associated with buffer",
		}
	}
	
	cursor := buf.Cursor()
	ctx := context.Background()
	
	hover, err := lspManager.Hover(ctx, buf.Filename(), cursor.Line, cursor.Col)
	if err != nil {
		return CommandResult{
			Success: false,
			Message: fmt.Sprintf("Hover failed: %v", err),
		}
	}
	
	if hover == nil || hover.Contents.Value == "" {
		return CommandResult{
			Success: true,
			Message: "No hover information available",
		}
	}
	
	return CommandResult{
		Success: true,
		Message: hover.Contents.Value,
	}
}

func (c *HoverCommand) Help() string {
	return "Show hover information at cursor position"
}

// DefinitionCommand jumps to definition
type DefinitionCommand struct{}

func NewDefinitionCommand() Command {
	return &DefinitionCommand{}
}

func (c *DefinitionCommand) Name() string {
	return "definition"
}

func (c *DefinitionCommand) Aliases() []string {
	return []string{"def", "lsp-definition"}
}

func (c *DefinitionCommand) Execute(args []string, buf *buffer.Buffer) CommandResult {
	if lspManager == nil {
		return CommandResult{
			Success: false,
			Message: "LSP not available",
		}
	}
	
	if buf.Filename() == "" {
		return CommandResult{
			Success: false,
			Message: "No file associated with buffer",
		}
	}
	
	cursor := buf.Cursor()
	ctx := context.Background()
	
	locations, err := lspManager.Definition(ctx, buf.Filename(), cursor.Line, cursor.Col)
	if err != nil {
		return CommandResult{
			Success: false,
			Message: fmt.Sprintf("Definition failed: %v", err),
		}
	}
	
	if len(locations) == 0 {
		return CommandResult{
			Success: true,
			Message: "No definition found",
		}
	}
	
	// For now, just show the first location
	loc := locations[0]
	return CommandResult{
		Success: true,
		Message: fmt.Sprintf("Definition at %s:%d:%d", loc.URI, loc.Range.Start.Line+1, loc.Range.Start.Character+1),
	}
}

func (c *DefinitionCommand) Help() string {
	return "Go to definition of symbol at cursor"
}

// ReferencesCommand finds references
type ReferencesCommand struct{}

func NewReferencesCommand() Command {
	return &ReferencesCommand{}
}

func (c *ReferencesCommand) Name() string {
	return "references"
}

func (c *ReferencesCommand) Aliases() []string {
	return []string{"refs", "lsp-references"}
}

func (c *ReferencesCommand) Execute(args []string, buf *buffer.Buffer) CommandResult {
	if lspManager == nil {
		return CommandResult{
			Success: false,
			Message: "LSP not available",
		}
	}
	
	if buf.Filename() == "" {
		return CommandResult{
			Success: false,
			Message: "No file associated with buffer",
		}
	}
	
	cursor := buf.Cursor()
	ctx := context.Background()
	
	locations, err := lspManager.References(ctx, buf.Filename(), cursor.Line, cursor.Col)
	if err != nil {
		return CommandResult{
			Success: false,
			Message: fmt.Sprintf("References failed: %v", err),
		}
	}
	
	if len(locations) == 0 {
		return CommandResult{
			Success: true,
			Message: "No references found",
		}
	}
	
	// Format references
	msg := fmt.Sprintf("Found %d references:", len(locations))
	for i, loc := range locations {
		if i < 5 { // Show first 5 references
			msg += fmt.Sprintf("\n  %s:%d:%d", loc.URI, loc.Range.Start.Line+1, loc.Range.Start.Character+1)
		}
	}
	if len(locations) > 5 {
		msg += fmt.Sprintf("\n  ... and %d more", len(locations)-5)
	}
	
	return CommandResult{
		Success: true,
		Message: msg,
	}
}

func (c *ReferencesCommand) Help() string {
	return "Find all references to symbol at cursor"
}

// RenameCommand renames a symbol
type RenameCommand struct{}

func NewRenameCommand() Command {
	return &RenameCommand{}
}

func (c *RenameCommand) Name() string {
	return "rename"
}

func (c *RenameCommand) Aliases() []string {
	return []string{"lsp-rename"}
}

func (c *RenameCommand) Execute(args []string, buf *buffer.Buffer) CommandResult {
	if lspManager == nil {
		return CommandResult{
			Success: false,
			Message: "LSP not available",
		}
	}
	
	if len(args) < 1 {
		return CommandResult{
			Success: false,
			Message: "Usage: :rename <new-name>",
		}
	}
	
	if buf.Filename() == "" {
		return CommandResult{
			Success: false,
			Message: "No file associated with buffer",
		}
	}
	
	newName := args[0]
	cursor := buf.Cursor()
	ctx := context.Background()
	
	workspaceEdit, err := lspManager.Rename(ctx, buf.Filename(), cursor.Line, cursor.Col, newName)
	if err != nil {
		return CommandResult{
			Success: false,
			Message: fmt.Sprintf("Rename failed: %v", err),
		}
	}
	
	if workspaceEdit == nil || len(workspaceEdit.Changes) == 0 {
		return CommandResult{
			Success: true,
			Message: "No changes to apply",
		}
	}
	
	// Count total edits
	totalEdits := 0
	for _, edits := range workspaceEdit.Changes {
		totalEdits += len(edits)
	}
	
	return CommandResult{
		Success: true,
		Message: fmt.Sprintf("Rename would affect %d locations in %d files", totalEdits, len(workspaceEdit.Changes)),
	}
}

func (c *RenameCommand) Help() string {
	return "Rename symbol at cursor"
}