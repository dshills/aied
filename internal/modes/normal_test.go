package modes

import (
	"testing"

	"github.com/dshills/aied/internal/buffer"
	"github.com/dshills/aied/internal/ui"
)

func TestNormalMode_BasicMovement(t *testing.T) {
	mode := NewNormalMode()
	buf := buffer.New()
	
	// Add some content to test movement
	buf.InsertChar('h')
	buf.InsertChar('e')
	buf.InsertChar('l')
	buf.InsertChar('l')
	buf.InsertChar('o')
	buf.InsertLine()
	buf.InsertChar('w')
	buf.InsertChar('o')
	buf.InsertChar('r')
	buf.InsertChar('l')
	buf.InsertChar('d')

	// Move to start for testing
	buf.SetCursor(buffer.Position{Line: 0, Col: 0})

	tests := []struct {
		name     string
		key      rune
		expected buffer.Position
	}{
		{"move right", 'l', buffer.Position{Line: 0, Col: 1}},
		{"move down", 'j', buffer.Position{Line: 1, Col: 1}},
		{"move left", 'h', buffer.Position{Line: 1, Col: 0}},
		{"move up", 'k', buffer.Position{Line: 0, Col: 0}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := ui.KeyEvent{Action: ui.KeyActionChar, Rune: tt.key}
			result := mode.HandleInput(event, buf)

			if !result.Handled {
				t.Error("expected movement to be handled")
			}

			cursor := buf.Cursor()
			if cursor != tt.expected {
				t.Errorf("expected cursor at %+v, got %+v", tt.expected, cursor)
			}
		})
	}
}

func TestNormalMode_LineMovement(t *testing.T) {
	mode := NewNormalMode()
	buf := buffer.New()
	
	// Add content: "  hello world"
	buf.InsertChar(' ')
	buf.InsertChar(' ')
	buf.InsertChar('h')
	buf.InsertChar('e')
	buf.InsertChar('l')
	buf.InsertChar('l')
	buf.InsertChar('o')
	buf.InsertChar(' ')
	buf.InsertChar('w')
	buf.InsertChar('o')
	buf.InsertChar('r')
	buf.InsertChar('l')
	buf.InsertChar('d')

	// Start in middle of line
	buf.SetCursor(buffer.Position{Line: 0, Col: 5})

	tests := []struct {
		name     string
		key      rune
		expected buffer.Position
	}{
		{"move to line start", '0', buffer.Position{Line: 0, Col: 0}},
		{"move to line end", '$', buffer.Position{Line: 0, Col: 12}}, // last char
		{"move to first non-whitespace", '^', buffer.Position{Line: 0, Col: 2}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset position
			buf.SetCursor(buffer.Position{Line: 0, Col: 5})
			
			event := ui.KeyEvent{Action: ui.KeyActionChar, Rune: tt.key}
			result := mode.HandleInput(event, buf)

			if !result.Handled {
				t.Error("expected line movement to be handled")
			}

			cursor := buf.Cursor()
			if cursor != tt.expected {
				t.Errorf("expected cursor at %+v, got %+v", tt.expected, cursor)
			}
		})
	}
}

func TestNormalMode_ModeSwitch(t *testing.T) {
	mode := NewNormalMode()
	buf := buffer.New()
	buf.InsertChar('t')
	buf.InsertChar('e')
	buf.InsertChar('s')
	buf.InsertChar('t')
	buf.SetCursor(buffer.Position{Line: 0, Col: 2}) // Middle of "test"

	tests := []struct {
		name         string
		key          rune
		expectedMode ModeType
		description  string
	}{
		{"insert at cursor", 'i', ModeInsert, "should switch to insert mode"},
		{"append after cursor", 'a', ModeInsert, "should move right and switch to insert"},
		{"visual mode", 'v', ModeVisual, "should switch to visual mode"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset position
			buf.SetCursor(buffer.Position{Line: 0, Col: 2})
			
			event := ui.KeyEvent{Action: ui.KeyActionChar, Rune: tt.key}
			result := mode.HandleInput(event, buf)

			if !result.Handled {
				t.Error("expected mode switch to be handled")
			}

			if result.SwitchToMode == nil {
				t.Error("expected mode switch to be requested")
			} else if *result.SwitchToMode != tt.expectedMode {
				t.Errorf("expected switch to %v, got %v", tt.expectedMode, *result.SwitchToMode)
			}
		})
	}
}

func TestNormalMode_WordMovement(t *testing.T) {
	mode := NewNormalMode()
	buf := buffer.New()
	
	// Add content: "hello world test"
	content := "hello world test"
	for _, ch := range content {
		buf.InsertChar(ch)
	}
	buf.SetCursor(buffer.Position{Line: 0, Col: 0})

	tests := []struct {
		name        string
		key         rune
		startPos    buffer.Position
		expectedCol int
	}{
		{"word forward from start", 'w', buffer.Position{Line: 0, Col: 0}, 6},  // "world"
		{"word forward from middle", 'w', buffer.Position{Line: 0, Col: 3}, 6}, // "world"
		{"word backward", 'b', buffer.Position{Line: 0, Col: 8}, 6},             // back to "world"
		{"word end", 'e', buffer.Position{Line: 0, Col: 0}, 4},                 // end of "hello"
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.SetCursor(tt.startPos)
			
			event := ui.KeyEvent{Action: ui.KeyActionChar, Rune: tt.key}
			result := mode.HandleInput(event, buf)

			if !result.Handled {
				t.Error("expected word movement to be handled")
			}

			cursor := buf.Cursor()
			if cursor.Col != tt.expectedCol {
				t.Errorf("expected column %d, got %d", tt.expectedCol, cursor.Col)
			}
		})
	}
}

func TestNormalMode_Deletion(t *testing.T) {
	mode := NewNormalMode()
	buf := buffer.New()
	
	// Add content: "hello"
	buf.InsertChar('h')
	buf.InsertChar('e')
	buf.InsertChar('l')
	buf.InsertChar('l')
	buf.InsertChar('o')
	buf.SetCursor(buffer.Position{Line: 0, Col: 2}) // on 'l'

	// Test 'x' (delete character at cursor)
	event := ui.KeyEvent{Action: ui.KeyActionChar, Rune: 'x'}
	result := mode.HandleInput(event, buf)

	if !result.Handled {
		t.Error("expected deletion to be handled")
	}

	expected := "helo"
	if buf.CurrentLine() != expected {
		t.Errorf("expected %q after deletion, got %q", expected, buf.CurrentLine())
	}
}

func TestNormalMode_OpenLines(t *testing.T) {
	mode := NewNormalMode()
	buf := buffer.New()
	
	// Add content: "hello"
	buf.InsertChar('h')
	buf.InsertChar('e')
	buf.InsertChar('l')
	buf.InsertChar('l')
	buf.InsertChar('o')

	// Test 'o' (open line below)
	event := ui.KeyEvent{Action: ui.KeyActionChar, Rune: 'o'}
	result := mode.HandleInput(event, buf)

	if !result.Handled {
		t.Error("expected line opening to be handled")
	}

	if result.SwitchToMode == nil || *result.SwitchToMode != ModeInsert {
		t.Error("expected switch to insert mode after opening line")
	}

	if buf.LineCount() != 2 {
		t.Errorf("expected 2 lines after opening line below, got %d", buf.LineCount())
	}

	cursor := buf.Cursor()
	if cursor.Line != 1 || cursor.Col != 0 {
		t.Errorf("expected cursor at (1,0) after opening line, got (%d,%d)", cursor.Line, cursor.Col)
	}
}

func TestNormalMode_OnEnter(t *testing.T) {
	mode := NewNormalMode()
	buf := buffer.New()
	
	// Add content and position cursor past line end (simulating coming from insert mode)
	buf.InsertChar('h')
	buf.InsertChar('i')
	buf.SetCursor(buffer.Position{Line: 0, Col: 5}) // Past end of line

	mode.OnEnter(buf)

	cursor := buf.Cursor()
	if cursor.Col != 1 { // Should be on last character
		t.Errorf("expected cursor to be adjusted to column 1, got %d", cursor.Col)
	}
}