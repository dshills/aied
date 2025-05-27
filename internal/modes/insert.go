package modes

import (
	"github.com/dshills/aied/internal/buffer"
	"github.com/dshills/aied/internal/ui"
)

// InsertMode implements VIM insert mode behavior
type InsertMode struct{}

// NewInsertMode creates a new insert mode instance
func NewInsertMode() *InsertMode {
	return &InsertMode{}
}

// Type returns the mode type
func (i *InsertMode) Type() ModeType {
	return ModeInsert
}

// HandleInput processes keyboard input in insert mode
func (i *InsertMode) HandleInput(event ui.KeyEvent, buf *buffer.Buffer) ModeResult {
	switch event.Action {
	case ui.KeyActionEscape:
		// Return to normal mode
		return ModeResult{SwitchToMode: &[]ModeType{ModeNormal}[0], Handled: true}

	case ui.KeyActionChar:
		// Insert the character
		buf.InsertChar(event.Rune)
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
	return "-- INSERT --"
}