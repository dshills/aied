package modes

import (
	"unicode"

	"github.com/dshills/aied/internal/buffer"
	"github.com/dshills/aied/internal/ui"
)

// NormalMode implements VIM normal mode behavior
type NormalMode struct {
	lastCommand rune // For repeat operations (.)
	gPrefix     bool // Whether 'g' was pressed (for two-char commands)
}

// NewNormalMode creates a new normal mode instance
func NewNormalMode() *NormalMode {
	return &NormalMode{}
}

// Type returns the mode type
func (n *NormalMode) Type() ModeType {
	return ModeNormal
}

// HandleInput processes keyboard input in normal mode
func (n *NormalMode) HandleInput(event ui.KeyEvent, buf *buffer.Buffer) ModeResult {
	switch event.Action {
	case ui.KeyActionChar:
		return n.handleCharacter(event.Rune, buf)
	case ui.KeyActionUp, ui.KeyActionDown, ui.KeyActionLeft, ui.KeyActionRight:
		return n.handleArrowKeys(event.Action, buf)
	case ui.KeyActionHome, ui.KeyActionEnd:
		return n.handleHomeEnd(event.Action, buf)
	case ui.KeyActionCtrlC:
		return ModeResult{ExitEditor: true, Handled: true}
	default:
		return ModeResult{Handled: false}
	}
}

// handleCharacter processes character input in normal mode
func (n *NormalMode) handleCharacter(ch rune, buf *buffer.Buffer) ModeResult {
	// Handle g-prefix commands
	if n.gPrefix {
		n.gPrefix = false
		switch ch {
		case 'd':
			// Go to definition
			return n.executeLSPCommand(":definition", buf)
		case 'h':
			// Show hover information
			return n.executeLSPCommand(":hover", buf)
		case 'r':
			// Find references
			return n.executeLSPCommand(":references", buf)
		case 'g':
			// gg - go to first line
			buf.SetCursor(buffer.Position{Line: 0, Col: 0})
			return ModeResult{Handled: true}
		default:
			// Unknown g command
			return ModeResult{Handled: true}
		}
	}
	
	switch ch {
	// Basic movement (hjkl)
	case 'h':
		return n.moveLeft(buf)
	case 'j':
		return n.moveDown(buf)
	case 'k':
		return n.moveUp(buf)
	case 'l':
		return n.moveRight(buf)

	// Line movement
	case '0':
		return n.moveToLineStart(buf)
	case '$':
		return n.moveToLineEnd(buf)
	case '^':
		return n.moveToFirstNonWhitespace(buf)

	// Word movement
	case 'w':
		return n.moveWordForward(buf)
	case 'b':
		return n.moveWordBackward(buf)
	case 'e':
		return n.moveToWordEnd(buf)

	// Mode switching
	case 'i':
		return ModeResult{SwitchToMode: &[]ModeType{ModeInsert}[0], Handled: true}
	case 'a':
		// Move cursor right then enter insert mode
		n.moveRight(buf)
		return ModeResult{SwitchToMode: &[]ModeType{ModeInsert}[0], Handled: true}
	case 'I':
		// Move to start of line then enter insert mode
		n.moveToFirstNonWhitespace(buf)
		return ModeResult{SwitchToMode: &[]ModeType{ModeInsert}[0], Handled: true}
	case 'A':
		// Move to end of line then enter insert mode
		n.moveToLineEnd(buf)
		return ModeResult{SwitchToMode: &[]ModeType{ModeInsert}[0], Handled: true}
	case 'o':
		// Open new line below and enter insert mode
		n.openLineBelow(buf)
		return ModeResult{SwitchToMode: &[]ModeType{ModeInsert}[0], Handled: true}
	case 'O':
		// Open new line above and enter insert mode
		n.openLineAbove(buf)
		return ModeResult{SwitchToMode: &[]ModeType{ModeInsert}[0], Handled: true}

	// Visual mode
	case 'v':
		return ModeResult{SwitchToMode: &[]ModeType{ModeVisual}[0], Handled: true}

	// Command mode
	case ':':
		return ModeResult{SwitchToMode: &[]ModeType{ModeCommand}[0], Handled: true}

	// Deletion
	case 'x':
		return n.deleteChar(buf)
	case 'X':
		return n.deleteCharBefore(buf)

	// Line operations
	case 'd':
		// TODO: Implement dd (delete line) with count support
		return ModeResult{Handled: true}
	case 'y':
		// TODO: Implement yy (yank line) with count support
		return ModeResult{Handled: true}
	case 'p':
		// TODO: Implement paste
		return ModeResult{Handled: true}

	// Undo/Redo
	case 'u':
		// TODO: Implement undo
		return ModeResult{Handled: true}
	
	// Two-character commands
	case 'g':
		n.gPrefix = true
		return ModeResult{Handled: true}

	default:
		return ModeResult{Handled: false}
	}
}

// Movement methods
func (n *NormalMode) moveLeft(buf *buffer.Buffer) ModeResult {
	buf.MoveCursor(0, -1)
	return ModeResult{Handled: true}
}

func (n *NormalMode) moveDown(buf *buffer.Buffer) ModeResult {
	buf.MoveCursor(1, 0)
	return ModeResult{Handled: true}
}

func (n *NormalMode) moveUp(buf *buffer.Buffer) ModeResult {
	buf.MoveCursor(-1, 0)
	return ModeResult{Handled: true}
}

func (n *NormalMode) moveRight(buf *buffer.Buffer) ModeResult {
	cursor := buf.Cursor()
	lineLen := len(buf.CurrentLine())
	// In normal mode, don't move past the last character
	if cursor.Col < lineLen-1 || lineLen == 0 {
		buf.MoveCursor(0, 1)
	}
	return ModeResult{Handled: true}
}

func (n *NormalMode) moveToLineStart(buf *buffer.Buffer) ModeResult {
	cursor := buf.Cursor()
	buf.SetCursor(buffer.Position{Line: cursor.Line, Col: 0})
	return ModeResult{Handled: true}
}

func (n *NormalMode) moveToLineEnd(buf *buffer.Buffer) ModeResult {
	cursor := buf.Cursor()
	lineLen := len(buf.CurrentLine())
	// In normal mode, cursor stops at last character, not after it
	endCol := lineLen - 1
	if endCol < 0 {
		endCol = 0
	}
	buf.SetCursor(buffer.Position{Line: cursor.Line, Col: endCol})
	return ModeResult{Handled: true}
}

func (n *NormalMode) moveToFirstNonWhitespace(buf *buffer.Buffer) ModeResult {
	cursor := buf.Cursor()
	line := buf.CurrentLine()
	
	col := 0
	for i, ch := range line {
		if !unicode.IsSpace(ch) {
			col = i
			break
		}
	}
	
	buf.SetCursor(buffer.Position{Line: cursor.Line, Col: col})
	return ModeResult{Handled: true}
}

func (n *NormalMode) moveWordForward(buf *buffer.Buffer) ModeResult {
	cursor := buf.Cursor()
	line := buf.CurrentLine()
	
	if cursor.Col >= len(line) {
		// Move to next line
		if cursor.Line < buf.LineCount()-1 {
			buf.SetCursor(buffer.Position{Line: cursor.Line + 1, Col: 0})
		}
		return ModeResult{Handled: true}
	}
	
	runes := []rune(line)
	col := cursor.Col
	
	// Skip current word
	for col < len(runes) && !unicode.IsSpace(runes[col]) {
		col++
	}
	
	// Skip spaces
	for col < len(runes) && unicode.IsSpace(runes[col]) {
		col++
	}
	
	// If we reached the end, move to next line
	if col >= len(runes) && cursor.Line < buf.LineCount()-1 {
		buf.SetCursor(buffer.Position{Line: cursor.Line + 1, Col: 0})
	} else {
		buf.SetCursor(buffer.Position{Line: cursor.Line, Col: col})
	}
	
	return ModeResult{Handled: true}
}

func (n *NormalMode) moveWordBackward(buf *buffer.Buffer) ModeResult {
	cursor := buf.Cursor()
	
	if cursor.Col <= 0 {
		// Move to previous line end
		if cursor.Line > 0 {
			prevLine, _ := buf.Line(cursor.Line - 1)
			endCol := len(prevLine) - 1
			if endCol < 0 {
				endCol = 0
			}
			buf.SetCursor(buffer.Position{Line: cursor.Line - 1, Col: endCol})
		}
		return ModeResult{Handled: true}
	}
	
	line := buf.CurrentLine()
	runes := []rune(line)
	col := cursor.Col - 1
	
	// Skip spaces
	for col >= 0 && unicode.IsSpace(runes[col]) {
		col--
	}
	
	// Skip current word
	for col >= 0 && !unicode.IsSpace(runes[col]) {
		col--
	}
	
	col++ // Move to start of word
	if col < 0 {
		col = 0
	}
	
	buf.SetCursor(buffer.Position{Line: cursor.Line, Col: col})
	return ModeResult{Handled: true}
}

func (n *NormalMode) moveToWordEnd(buf *buffer.Buffer) ModeResult {
	cursor := buf.Cursor()
	line := buf.CurrentLine()
	runes := []rune(line)
	
	if cursor.Col >= len(runes)-1 {
		// Move to next line
		if cursor.Line < buf.LineCount()-1 {
			nextLine, _ := buf.Line(cursor.Line + 1)
			nextRunes := []rune(nextLine)
			// Find first non-space character, then find end of that word
			col := 0
			for col < len(nextRunes) && unicode.IsSpace(nextRunes[col]) {
				col++
			}
			for col < len(nextRunes) && !unicode.IsSpace(nextRunes[col]) {
				col++
			}
			if col > 0 {
				col--
			}
			buf.SetCursor(buffer.Position{Line: cursor.Line + 1, Col: col})
		}
		return ModeResult{Handled: true}
	}
	
	col := cursor.Col + 1
	
	// Skip spaces
	for col < len(runes) && unicode.IsSpace(runes[col]) {
		col++
	}
	
	// Move to end of word
	for col < len(runes) && !unicode.IsSpace(runes[col]) {
		col++
	}
	
	if col > 0 {
		col--
	}
	
	buf.SetCursor(buffer.Position{Line: cursor.Line, Col: col})
	return ModeResult{Handled: true}
}

// Line operations
func (n *NormalMode) openLineBelow(buf *buffer.Buffer) ModeResult {
	cursor := buf.Cursor()
	// Move to end of current line
	lineLen := len(buf.CurrentLine())
	buf.SetCursor(buffer.Position{Line: cursor.Line, Col: lineLen})
	// Insert new line
	buf.InsertLine()
	return ModeResult{Handled: true}
}

func (n *NormalMode) openLineAbove(buf *buffer.Buffer) ModeResult {
	cursor := buf.Cursor()
	// Move to beginning of current line
	buf.SetCursor(buffer.Position{Line: cursor.Line, Col: 0})
	// Insert empty line above
	buf.InsertEmptyLine()
	return ModeResult{Handled: true}
}

// Deletion operations
func (n *NormalMode) deleteChar(buf *buffer.Buffer) ModeResult {
	buf.DeleteChar()
	return ModeResult{Handled: true}
}

func (n *NormalMode) deleteCharBefore(buf *buffer.Buffer) ModeResult {
	buf.Backspace()
	return ModeResult{Handled: true}
}

// Arrow key handling in normal mode
func (n *NormalMode) handleArrowKeys(action ui.KeyAction, buf *buffer.Buffer) ModeResult {
	switch action {
	case ui.KeyActionUp:
		return n.moveUp(buf)
	case ui.KeyActionDown:
		return n.moveDown(buf)
	case ui.KeyActionLeft:
		return n.moveLeft(buf)
	case ui.KeyActionRight:
		return n.moveRight(buf)
	}
	return ModeResult{Handled: false}
}

func (n *NormalMode) handleHomeEnd(action ui.KeyAction, buf *buffer.Buffer) ModeResult {
	switch action {
	case ui.KeyActionHome:
		return n.moveToLineStart(buf)
	case ui.KeyActionEnd:
		return n.moveToLineEnd(buf)
	}
	return ModeResult{Handled: false}
}

// Mode lifecycle methods
func (n *NormalMode) OnEnter(buf *buffer.Buffer) {
	if buf == nil {
		return
	}
	
	// In normal mode, ensure cursor doesn't go past line end
	cursor := buf.Cursor()
	lineLen := len(buf.CurrentLine())
	if cursor.Col >= lineLen && lineLen > 0 {
		buf.SetCursor(buffer.Position{Line: cursor.Line, Col: lineLen - 1})
	}
}

func (n *NormalMode) OnExit(buf *buffer.Buffer) {
	// Nothing special needed when exiting normal mode
}

func (n *NormalMode) GetStatusText() string {
	if n.gPrefix {
		return "g"
	}
	return ""
}

// executeLSPCommand executes an LSP command via command mode
func (n *NormalMode) executeLSPCommand(command string, buf *buffer.Buffer) ModeResult {
	// We need to execute the command through the command system
	// For now, just return a message
	// TODO: Integrate with command execution system
	return ModeResult{Handled: true}
}