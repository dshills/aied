package modes

import (
	"testing"

	"github.com/dshills/aied/internal/buffer"
	"github.com/dshills/aied/internal/ui"
)

func TestModeType_String(t *testing.T) {
	tests := []struct {
		mode     ModeType
		expected string
	}{
		{ModeNormal, "NORMAL"},
		{ModeInsert, "INSERT"},
		{ModeVisual, "VISUAL"},
		{ModeCommand, "COMMAND"},
		{ModeType(999), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.mode.String()
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestNewModeManager(t *testing.T) {
	mm := NewModeManager()

	if mm == nil {
		t.Fatal("expected mode manager to be created")
	}

	// Should start in Normal mode
	if mm.CurrentModeType() != ModeNormal {
		t.Errorf("expected to start in Normal mode, got %v", mm.CurrentModeType())
	}

	// Should have all modes registered
	expectedModes := []ModeType{ModeNormal, ModeInsert, ModeVisual}
	for _, modeType := range expectedModes {
		if _, exists := mm.modes[modeType]; !exists {
			t.Errorf("expected mode %v to be registered", modeType)
		}
	}
}

func TestModeManager_SwitchToMode(t *testing.T) {
	mm := NewModeManager()
	buf := buffer.New()

	// Switch to Insert mode
	mm.SwitchToMode(ModeInsert, buf)
	if mm.CurrentModeType() != ModeInsert {
		t.Errorf("expected Insert mode, got %v", mm.CurrentModeType())
	}

	// Switch to Visual mode
	mm.SwitchToMode(ModeVisual, buf)
	if mm.CurrentModeType() != ModeVisual {
		t.Errorf("expected Visual mode, got %v", mm.CurrentModeType())
	}

	// Switch back to Normal mode
	mm.SwitchToMode(ModeNormal, buf)
	if mm.CurrentModeType() != ModeNormal {
		t.Errorf("expected Normal mode, got %v", mm.CurrentModeType())
	}
}

func TestModeManager_HandleInput(t *testing.T) {
	mm := NewModeManager()
	buf := buffer.New()

	// Test mode switching from Normal to Insert
	event := ui.KeyEvent{Action: ui.KeyActionChar, Rune: 'i'}
	result := mm.HandleInput(event, buf)

	if !result.Handled {
		t.Error("expected input to be handled")
	}

	if mm.CurrentModeType() != ModeInsert {
		t.Errorf("expected to switch to Insert mode, got %v", mm.CurrentModeType())
	}

	// Test character insertion in Insert mode
	charEvent := ui.KeyEvent{Action: ui.KeyActionChar, Rune: 'a'}
	result = mm.HandleInput(charEvent, buf)

	if !result.Handled {
		t.Error("expected character input to be handled in Insert mode")
	}

	if buf.CurrentLine() != "a" {
		t.Errorf("expected buffer to contain 'a', got %q", buf.CurrentLine())
	}

	// Test switching back to Normal mode
	escapeEvent := ui.KeyEvent{Action: ui.KeyActionEscape}
	result = mm.HandleInput(escapeEvent, buf)

	if !result.Handled {
		t.Error("expected escape to be handled")
	}

	if mm.CurrentModeType() != ModeNormal {
		t.Errorf("expected to switch back to Normal mode, got %v", mm.CurrentModeType())
	}
}

func TestModeManager_GetStatusText(t *testing.T) {
	mm := NewModeManager()
	buf := buffer.New()

	// Normal mode should have empty status text
	statusText := mm.GetStatusText()
	if statusText != "" {
		t.Errorf("expected empty status text for Normal mode, got %q", statusText)
	}

	// Switch to Insert mode and check status
	mm.SwitchToMode(ModeInsert, buf)
	statusText = mm.GetStatusText()
	if statusText != "-- INSERT --" {
		t.Errorf("expected '-- INSERT --', got %q", statusText)
	}

	// Switch to Visual mode and check status
	mm.SwitchToMode(ModeVisual, buf)
	statusText = mm.GetStatusText()
	if statusText != "-- VISUAL --" {
		t.Errorf("expected '-- VISUAL --', got %q", statusText)
	}
}