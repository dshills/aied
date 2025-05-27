package ui

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestEventProcessor_ProcessKeyEvent(t *testing.T) {
	// Create a mock screen for testing
	screen := &Screen{} // We don't need a real screen for event processing
	processor := NewEventProcessor(screen)

	tests := []struct {
		name        string
		key         tcell.Key
		rune        rune
		expectedAction KeyAction
	}{
		{"Escape key", tcell.KeyEscape, 0, KeyActionEscape},
		{"Enter key", tcell.KeyEnter, 0, KeyActionEnter},
		{"Backspace key", tcell.KeyBackspace, 0, KeyActionBackspace},
		{"Delete key", tcell.KeyDelete, 0, KeyActionDelete},
		{"Up arrow", tcell.KeyUp, 0, KeyActionUp},
		{"Down arrow", tcell.KeyDown, 0, KeyActionDown},
		{"Left arrow", tcell.KeyLeft, 0, KeyActionLeft},
		{"Right arrow", tcell.KeyRight, 0, KeyActionRight},
		{"Home key", tcell.KeyHome, 0, KeyActionHome},
		{"End key", tcell.KeyEnd, 0, KeyActionEnd},
		{"Ctrl+C", tcell.KeyCtrlC, 0, KeyActionQuit},
		{"Ctrl+S", tcell.KeyCtrlS, 0, KeyActionCtrlS},
		{"Character 'a'", tcell.KeyRune, 'a', KeyActionChar},
		{"Character 'Z'", tcell.KeyRune, 'Z', KeyActionChar},
		{"Space", tcell.KeyRune, ' ', KeyActionChar},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := tcell.NewEventKey(tt.key, tt.rune, tcell.ModNone)
			result := processor.processKeyEvent(event)

			if result.Action != tt.expectedAction {
				t.Errorf("expected action %v, got %v", tt.expectedAction, result.Action)
			}

			if tt.expectedAction == KeyActionChar && result.Rune != tt.rune {
				t.Errorf("expected rune %c, got %c", tt.rune, result.Rune)
			}
		})
	}
}

func TestEventProcessor_ProcessResizeEvent(t *testing.T) {
	screen := &Screen{width: 40, height: 12} // Set initial size
	processor := NewEventProcessor(screen)

	event := tcell.NewEventResize(80, 24)
	result := processor.processResizeEvent(event)

	expected := ResizeEvent{Width: 80, Height: 24}
	if result != expected {
		t.Errorf("expected %+v, got %+v", expected, result)
	}

	// Check that screen size was updated
	if screen.width != 80 || screen.height != 24 {
		t.Errorf("expected screen size to be updated to 80x24, got %dx%d", screen.width, screen.height)
	}
}

func TestIsQuitEvent(t *testing.T) {
	tests := []struct {
		name     string
		event    interface{}
		expected bool
	}{
		{"Quit event", KeyEvent{Action: KeyActionQuit}, true},
		{"Ctrl+C event", KeyEvent{Action: KeyActionCtrlC}, false}, // CtrlC is converted to Quit in processing
		{"Character event", KeyEvent{Action: KeyActionChar}, false},
		{"Resize event", ResizeEvent{}, false},
		{"Nil event", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsQuitEvent(tt.event)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestIsCharEvent(t *testing.T) {
	tests := []struct {
		name     string
		event    interface{}
		expected bool
	}{
		{"Character event", KeyEvent{Action: KeyActionChar, Rune: 'a'}, true},
		{"Quit event", KeyEvent{Action: KeyActionQuit}, false},
		{"Resize event", ResizeEvent{}, false},
		{"Movement event", KeyEvent{Action: KeyActionUp}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsCharEvent(tt.event)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestIsMovementEvent(t *testing.T) {
	tests := []struct {
		name     string
		event    interface{}
		expected bool
	}{
		{"Up arrow", KeyEvent{Action: KeyActionUp}, true},
		{"Down arrow", KeyEvent{Action: KeyActionDown}, true},
		{"Left arrow", KeyEvent{Action: KeyActionLeft}, true},
		{"Right arrow", KeyEvent{Action: KeyActionRight}, true},
		{"Home key", KeyEvent{Action: KeyActionHome}, true},
		{"End key", KeyEvent{Action: KeyActionEnd}, true},
		{"Page up", KeyEvent{Action: KeyActionPageUp}, true},
		{"Page down", KeyEvent{Action: KeyActionPageDown}, true},
		{"Character", KeyEvent{Action: KeyActionChar}, false},
		{"Quit", KeyEvent{Action: KeyActionQuit}, false},
		{"Resize", ResizeEvent{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsMovementEvent(tt.event)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestIsResizeEvent(t *testing.T) {
	tests := []struct {
		name     string
		event    interface{}
		expected bool
	}{
		{"Resize event", ResizeEvent{Width: 80, Height: 24}, true},
		{"Key event", KeyEvent{Action: KeyActionChar}, false},
		{"Nil event", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsResizeEvent(tt.event)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}