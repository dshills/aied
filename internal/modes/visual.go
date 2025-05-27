package modes

import (
	"github.com/dshills/aied/internal/buffer"
	"github.com/dshills/aied/internal/ui"
)

// VisualMode implements VIM visual mode behavior
type VisualMode struct {
	startPos buffer.Position // Where selection started
}

// NewVisualMode creates a new visual mode instance
func NewVisualMode() *VisualMode {
	return &VisualMode{}
}

// Type returns the mode type
func (v *VisualMode) Type() ModeType {
	return ModeVisual
}

// HandleInput processes keyboard input in visual mode
func (v *VisualMode) HandleInput(event ui.KeyEvent, buf *buffer.Buffer) ModeResult {
	switch event.Action {
	case ui.KeyActionEscape:
		// Return to normal mode
		return ModeResult{SwitchToMode: &[]ModeType{ModeNormal}[0], Handled: true}

	case ui.KeyActionChar:
		return v.handleCharacter(event.Rune, buf)

	case ui.KeyActionUp, ui.KeyActionDown, ui.KeyActionLeft, ui.KeyActionRight:
		return v.handleArrowKeys(event.Action, buf)

	case ui.KeyActionHome, ui.KeyActionEnd:
		return v.handleHomeEnd(event.Action, buf)

	case ui.KeyActionCtrlC:
		// Return to normal mode
		return ModeResult{SwitchToMode: &[]ModeType{ModeNormal}[0], Handled: true}

	default:
		return ModeResult{Handled: false}
	}
}

// handleCharacter processes character input in visual mode
func (v *VisualMode) handleCharacter(ch rune, buf *buffer.Buffer) ModeResult {
	switch ch {
	// Movement (same as normal mode but extends selection)
	case 'h':
		buf.MoveCursor(0, -1)
		return ModeResult{Handled: true}
	case 'j':
		buf.MoveCursor(1, 0)
		return ModeResult{Handled: true}
	case 'k':
		buf.MoveCursor(-1, 0)
		return ModeResult{Handled: true}
	case 'l':
		buf.MoveCursor(0, 1)
		return ModeResult{Handled: true}

	// Line movement
	case '0':
		cursor := buf.Cursor()
		buf.SetCursor(buffer.Position{Line: cursor.Line, Col: 0})
		return ModeResult{Handled: true}
	case '$':
		cursor := buf.Cursor()
		lineLen := len(buf.CurrentLine())
		buf.SetCursor(buffer.Position{Line: cursor.Line, Col: lineLen})
		return ModeResult{Handled: true}

	// Word movement
	case 'w':
		v.moveWordForward(buf)
		return ModeResult{Handled: true}
	case 'b':
		v.moveWordBackward(buf)
		return ModeResult{Handled: true}
	case 'e':
		v.moveToWordEnd(buf)
		return ModeResult{Handled: true}

	// Operations on selection
	case 'd', 'x':
		// Delete selected text (TODO: implement actual selection deletion)
		return ModeResult{SwitchToMode: &[]ModeType{ModeNormal}[0], Handled: true}
	case 'y':
		// Yank (copy) selected text (TODO: implement actual selection copying)
		return ModeResult{SwitchToMode: &[]ModeType{ModeNormal}[0], Handled: true}

	// Switch to other modes
	case 'i':
		return ModeResult{SwitchToMode: &[]ModeType{ModeInsert}[0], Handled: true}

	default:
		return ModeResult{Handled: false}
	}
}

// Word movement methods (similar to normal mode)
func (v *VisualMode) moveWordForward(buf *buffer.Buffer) {
	cursor := buf.Cursor()
	line := buf.CurrentLine()
	
	if cursor.Col >= len(line) {
		if cursor.Line < buf.LineCount()-1 {
			buf.SetCursor(buffer.Position{Line: cursor.Line + 1, Col: 0})
		}
		return
	}
	
	runes := []rune(line)
	col := cursor.Col
	
	// Skip current word
	for col < len(runes) && runes[col] != ' ' && runes[col] != '\t' {
		col++
	}
	
	// Skip spaces
	for col < len(runes) && (runes[col] == ' ' || runes[col] == '\t') {
		col++
	}
	
	if col >= len(runes) && cursor.Line < buf.LineCount()-1 {
		buf.SetCursor(buffer.Position{Line: cursor.Line + 1, Col: 0})
	} else {
		buf.SetCursor(buffer.Position{Line: cursor.Line, Col: col})
	}
}

func (v *VisualMode) moveWordBackward(buf *buffer.Buffer) {
	cursor := buf.Cursor()
	
	if cursor.Col <= 0 {
		if cursor.Line > 0 {
			prevLine, _ := buf.Line(cursor.Line - 1)
			buf.SetCursor(buffer.Position{Line: cursor.Line - 1, Col: len(prevLine)})
		}
		return
	}
	
	line := buf.CurrentLine()
	runes := []rune(line)
	col := cursor.Col - 1
	
	// Skip spaces
	for col >= 0 && (runes[col] == ' ' || runes[col] == '\t') {
		col--
	}
	
	// Skip current word
	for col >= 0 && runes[col] != ' ' && runes[col] != '\t' {
		col--
	}
	
	col++
	if col < 0 {
		col = 0
	}
	
	buf.SetCursor(buffer.Position{Line: cursor.Line, Col: col})
}

func (v *VisualMode) moveToWordEnd(buf *buffer.Buffer) {
	cursor := buf.Cursor()
	line := buf.CurrentLine()
	runes := []rune(line)
	
	if cursor.Col >= len(runes)-1 {
		if cursor.Line < buf.LineCount()-1 {
			nextLine, _ := buf.Line(cursor.Line + 1)
			nextRunes := []rune(nextLine)
			col := 0
			for col < len(nextRunes) && (nextRunes[col] == ' ' || nextRunes[col] == '\t') {
				col++
			}
			for col < len(nextRunes) && nextRunes[col] != ' ' && nextRunes[col] != '\t' {
				col++
			}
			if col > 0 {
				col--
			}
			buf.SetCursor(buffer.Position{Line: cursor.Line + 1, Col: col})
		}
		return
	}
	
	col := cursor.Col + 1
	
	// Skip spaces
	for col < len(runes) && (runes[col] == ' ' || runes[col] == '\t') {
		col++
	}
	
	// Move to end of word
	for col < len(runes) && runes[col] != ' ' && runes[col] != '\t' {
		col++
	}
	
	if col > 0 {
		col--
	}
	
	buf.SetCursor(buffer.Position{Line: cursor.Line, Col: col})
}

// Arrow key handling
func (v *VisualMode) handleArrowKeys(action ui.KeyAction, buf *buffer.Buffer) ModeResult {
	switch action {
	case ui.KeyActionUp:
		buf.MoveCursor(-1, 0)
	case ui.KeyActionDown:
		buf.MoveCursor(1, 0)
	case ui.KeyActionLeft:
		buf.MoveCursor(0, -1)
	case ui.KeyActionRight:
		buf.MoveCursor(0, 1)
	}
	return ModeResult{Handled: true}
}

func (v *VisualMode) handleHomeEnd(action ui.KeyAction, buf *buffer.Buffer) ModeResult {
	cursor := buf.Cursor()
	switch action {
	case ui.KeyActionHome:
		buf.SetCursor(buffer.Position{Line: cursor.Line, Col: 0})
	case ui.KeyActionEnd:
		lineLen := len(buf.CurrentLine())
		buf.SetCursor(buffer.Position{Line: cursor.Line, Col: lineLen})
	}
	return ModeResult{Handled: true}
}

// OnEnter is called when entering visual mode
func (v *VisualMode) OnEnter(buf *buffer.Buffer) {
	if buf == nil {
		return
	}
	// Remember where selection started
	v.startPos = buf.Cursor()
}

// OnExit is called when leaving visual mode
func (v *VisualMode) OnExit(buf *buffer.Buffer) {
	// Clear selection (for now, just reset)
	// TODO: Implement proper selection clearing in renderer
}

// GetStatusText returns mode-specific status information
func (v *VisualMode) GetStatusText() string {
	return "-- VISUAL --"
}

// GetSelection returns the current selection range
func (v *VisualMode) GetSelection(buf *buffer.Buffer) (buffer.Position, buffer.Position) {
	currentPos := buf.Cursor()
	
	// Return selection in order (start before end)
	if v.startPos.Line < currentPos.Line || 
		(v.startPos.Line == currentPos.Line && v.startPos.Col < currentPos.Col) {
		return v.startPos, currentPos
	}
	
	return currentPos, v.startPos
}