package modes

import (
	"context"
	
	"github.com/dshills/aied/internal/buffer"
	"github.com/dshills/aied/internal/lsp"
	"github.com/dshills/aied/internal/ui"
	"go.lsp.dev/protocol"
)

// InsertMode implements VIM insert mode behavior
type InsertMode struct{
	lspManager       *lsp.Manager
	showingCompletion bool
	completions      []CompletionItem
	selectedIndex    int
}

// CompletionItem represents a completion option
type CompletionItem struct {
	Label       string
	Detail      string
	InsertText  string
	Kind        string
}

// NewInsertMode creates a new insert mode instance
func NewInsertMode() *InsertMode {
	return &InsertMode{}
}

// SetLSPManager sets the LSP manager for code completion
func (i *InsertMode) SetLSPManager(manager *lsp.Manager) {
	i.lspManager = manager
}

// Type returns the mode type
func (i *InsertMode) Type() ModeType {
	return ModeInsert
}

// HandleInput processes keyboard input in insert mode
func (i *InsertMode) HandleInput(event ui.KeyEvent, buf *buffer.Buffer) ModeResult {
	// Handle completion navigation if showing completions
	if i.showingCompletion {
		switch event.Action {
		case ui.KeyActionUp:
			if i.selectedIndex > 0 {
				i.selectedIndex--
			}
			return ModeResult{Handled: true}
		case ui.KeyActionDown:
			if i.selectedIndex < len(i.completions)-1 {
				i.selectedIndex++
			}
			return ModeResult{Handled: true}
		case ui.KeyActionTab, ui.KeyActionEnter:
			// Accept completion
			if i.selectedIndex < len(i.completions) {
				i.applyCompletion(buf, i.completions[i.selectedIndex])
			}
			i.hideCompletion()
			return ModeResult{Handled: true}
		case ui.KeyActionEscape:
			// Cancel completion
			i.hideCompletion()
			return ModeResult{Handled: true}
		}
	}
	
	switch event.Action {
	case ui.KeyActionEscape:
		// Return to normal mode
		return ModeResult{SwitchToMode: &[]ModeType{ModeNormal}[0], Handled: true}

	case ui.KeyActionChar:
		// Insert the character
		buf.InsertChar(event.Rune)
		
		// Trigger completion on certain characters
		if i.lspManager != nil {
			if event.Rune == '.' || event.Rune == ':' {
				i.triggerCompletion(buf)
			}
		}
		return ModeResult{Handled: true}

	case ui.KeyActionBackspace:
		// Delete character before cursor
		buf.Backspace()
		return ModeResult{Handled: true}

	case ui.KeyActionDelete:
		// Delete character at cursor
		buf.DeleteChar()
		return ModeResult{Handled: true}

	case ui.KeyActionEnter:
		// Insert new line
		buf.InsertLine()
		return ModeResult{Handled: true}

	case ui.KeyActionTab:
		// Insert tab (4 spaces for now)
		buf.InsertChar(' ')
		buf.InsertChar(' ')
		buf.InsertChar(' ')
		buf.InsertChar(' ')
		return ModeResult{Handled: true}

	case ui.KeyActionUp:
		// Move cursor up (allow navigation in insert mode)
		buf.MoveCursor(-1, 0)
		return ModeResult{Handled: true}

	case ui.KeyActionDown:
		// Move cursor down
		buf.MoveCursor(1, 0)
		return ModeResult{Handled: true}

	case ui.KeyActionLeft:
		// Move cursor left
		buf.MoveCursor(0, -1)
		return ModeResult{Handled: true}

	case ui.KeyActionRight:
		// Move cursor right
		buf.MoveCursor(0, 1)
		return ModeResult{Handled: true}

	case ui.KeyActionHome:
		// Move to beginning of line
		cursor := buf.Cursor()
		buf.SetCursor(buffer.Position{Line: cursor.Line, Col: 0})
		return ModeResult{Handled: true}

	case ui.KeyActionEnd:
		// Move to end of line
		cursor := buf.Cursor()
		lineLen := len(buf.CurrentLine())
		buf.SetCursor(buffer.Position{Line: cursor.Line, Col: lineLen})
		return ModeResult{Handled: true}

	case ui.KeyActionCtrlC:
		// Exit insert mode and return to normal mode
		return ModeResult{SwitchToMode: &[]ModeType{ModeNormal}[0], Handled: true}

	case ui.KeyActionCtrlS:
		// Save file (common operation in insert mode)
		if buf.Filename() != "" {
			buf.Save()
		}
		return ModeResult{Handled: true}
		
	case ui.KeyActionCtrlSpace:
		// Manual completion trigger
		if i.lspManager != nil {
			i.triggerCompletion(buf)
		}
		return ModeResult{Handled: true}

	default:
		return ModeResult{Handled: false}
	}
}

// OnEnter is called when entering insert mode
func (i *InsertMode) OnEnter(buf *buffer.Buffer) {
	if buf == nil {
		return
	}
	// In insert mode, cursor can be positioned after the last character
	// No special adjustment needed
}

// OnExit is called when leaving insert mode
func (i *InsertMode) OnExit(buf *buffer.Buffer) {
	if buf == nil {
		return
	}
	
	// When leaving insert mode, adjust cursor to be on a character (not after)
	// This follows VIM behavior
	cursor := buf.Cursor()
	lineLen := len(buf.CurrentLine())
	
	if cursor.Col > 0 && cursor.Col >= lineLen && lineLen > 0 {
		buf.SetCursor(buffer.Position{Line: cursor.Line, Col: lineLen - 1})
	}
}

// GetStatusText returns mode-specific status information
func (i *InsertMode) GetStatusText() string {
	if i.showingCompletion {
		return "-- INSERT (completing) --"
	}
	return "-- INSERT --"
}

// triggerCompletion requests completions from LSP
func (i *InsertMode) triggerCompletion(buf *buffer.Buffer) {
	if buf.Filename() == "" {
		return
	}
	
	cursor := buf.Cursor()
	ctx := context.Background()
	
	completions, err := i.lspManager.Completion(ctx, buf.Filename(), cursor.Line, cursor.Col)
	if err != nil || len(completions) == 0 {
		i.hideCompletion()
		return
	}
	
	// Convert LSP completions to our format
	var items []CompletionItem
	for _, comp := range completions {
		insertText := comp.InsertText
		if insertText == "" {
			insertText = comp.Label
		}
		
		items = append(items, CompletionItem{
			Label:      comp.Label,
			Detail:     comp.Detail,
			InsertText: insertText,
			Kind:       getCompletionKindString(comp.Kind),
		})
	}
	
	if len(items) > 0 {
		i.completions = items
		i.selectedIndex = 0
		i.showingCompletion = true
	}
}

// applyCompletion applies the selected completion
func (i *InsertMode) applyCompletion(buf *buffer.Buffer, item CompletionItem) {
	if item.InsertText == "" {
		return
	}
	
	// Get current word being typed
	cursor := buf.Cursor()
	line := buf.CurrentLine()
	
	// Find word start
	wordStart := cursor.Col
	for wordStart > 0 && isIdentifierChar(rune(line[wordStart-1])) {
		wordStart--
	}
	
	// Delete current partial word
	for j := cursor.Col; j > wordStart; j-- {
		buf.Backspace()
	}
	
	// Insert completion text
	for _, ch := range item.InsertText {
		buf.InsertChar(ch)
	}
}

// hideCompletion hides the completion popup
func (i *InsertMode) hideCompletion() {
	i.showingCompletion = false
	i.completions = nil
	i.selectedIndex = 0
}

// isIdentifierChar checks if a character is part of an identifier
func isIdentifierChar(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || 
	       (r >= '0' && r <= '9') || r == '_'
}

// getCompletionKindString converts completion kind to string
func getCompletionKindString(kind protocol.CompletionItemKind) string {
	switch kind {
	case protocol.CompletionItemKindText:
		return "Text"
	case protocol.CompletionItemKindMethod:
		return "Method"
	case protocol.CompletionItemKindFunction:
		return "Function"
	case protocol.CompletionItemKindConstructor:
		return "Constructor"
	case protocol.CompletionItemKindField:
		return "Field"
	case protocol.CompletionItemKindVariable:
		return "Variable"
	case protocol.CompletionItemKindClass:
		return "Class"
	case protocol.CompletionItemKindInterface:
		return "Interface"
	case protocol.CompletionItemKindModule:
		return "Module"
	case protocol.CompletionItemKindProperty:
		return "Property"
	case protocol.CompletionItemKindUnit:
		return "Unit"
	case protocol.CompletionItemKindValue:
		return "Value"
	case protocol.CompletionItemKindEnum:
		return "Enum"
	case protocol.CompletionItemKindKeyword:
		return "Keyword"
	case protocol.CompletionItemKindSnippet:
		return "Snippet"
	case protocol.CompletionItemKindColor:
		return "Color"
	case protocol.CompletionItemKindFile:
		return "File"
	case protocol.CompletionItemKindReference:
		return "Reference"
	case protocol.CompletionItemKindFolder:
		return "Folder"
	case protocol.CompletionItemKindEnumMember:
		return "EnumMember"
	case protocol.CompletionItemKindConstant:
		return "Constant"
	case protocol.CompletionItemKindStruct:
		return "Struct"
	case protocol.CompletionItemKindEvent:
		return "Event"
	case protocol.CompletionItemKindOperator:
		return "Operator"
	case protocol.CompletionItemKindTypeParameter:
		return "TypeParameter"
	default:
		return "Unknown"
	}
}

// GetCompletions returns current completion items for display
func (i *InsertMode) GetCompletions() ([]CompletionItem, int, bool) {
	return i.completions, i.selectedIndex, i.showingCompletion
}