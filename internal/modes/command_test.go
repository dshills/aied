package modes

import (
	"testing"

	"github.com/dshills/aied/internal/buffer"
	"github.com/dshills/aied/internal/ui"
)

func TestCommandMode_Type(t *testing.T) {
	mode := NewCommandMode()
	if mode.Type() != ModeCommand {
		t.Errorf("expected mode type %v, got %v", ModeCommand, mode.Type())
	}
}

func TestCommandMode_HandleInput(t *testing.T) {
	mode := NewCommandMode()
	buf := buffer.New()

	// Test character input
	event := ui.KeyEvent{Action: ui.KeyActionChar, Rune: 'w'}
	result := mode.HandleInput(event, buf)
	if !result.Handled {
		t.Error("expected character input to be handled")
	}
	if mode.GetCommandLine() != ":w" {
		t.Errorf("expected command line ':w', got %q", mode.GetCommandLine())
	}

	// Test backspace
	event = ui.KeyEvent{Action: ui.KeyActionBackspace}
	result = mode.HandleInput(event, buf)
	if !result.Handled {
		t.Error("expected backspace to be handled")
	}
	if mode.GetCommandLine() != ":" {
		t.Errorf("expected command line ':', got %q", mode.GetCommandLine())
	}

	// Test escape (cancel)
	event = ui.KeyEvent{Action: ui.KeyActionEscape}
	result = mode.HandleInput(event, buf)
	if !result.Handled {
		t.Error("expected escape to be handled")
	}
	if result.SwitchToMode == nil || *result.SwitchToMode != ModeNormal {
		t.Error("expected escape to switch to Normal mode")
	}
	if mode.GetCommandLine() != ":" {
		t.Errorf("expected command line to be reset, got %q", mode.GetCommandLine())
	}
}

func TestCommandMode_ExecuteCommand(t *testing.T) {
	mode := NewCommandMode()
	buf := buffer.New()

	// Type a quit command
	chars := "q"
	for _, ch := range chars {
		event := ui.KeyEvent{Action: ui.KeyActionChar, Rune: ch}
		mode.HandleInput(event, buf)
	}

	// Execute with Enter
	event := ui.KeyEvent{Action: ui.KeyActionEnter}
	result := mode.HandleInput(event, buf)

	if !result.Handled {
		t.Error("expected enter to be handled")
	}
	if !result.ExitEditor {
		t.Error("expected quit command to exit editor")
	}
}

func TestCommandMode_InvalidCommand(t *testing.T) {
	mode := NewCommandMode()
	buf := buffer.New()

	// Type an invalid command
	chars := "invalid"
	for _, ch := range chars {
		event := ui.KeyEvent{Action: ui.KeyActionChar, Rune: ch}
		mode.HandleInput(event, buf)
	}

	// Execute with Enter
	event := ui.KeyEvent{Action: ui.KeyActionEnter}
	result := mode.HandleInput(event, buf)

	if !result.Handled {
		t.Error("expected enter to be handled")
	}
	if result.ExitEditor {
		t.Error("expected invalid command not to exit editor")
	}
	if result.SwitchToMode == nil || *result.SwitchToMode != ModeNormal {
		t.Error("expected invalid command to switch back to Normal mode")
	}

	// Should have an error message
	message := mode.GetMessage()
	if message == "" {
		t.Error("expected error message for invalid command")
	}
}

func TestCommandMode_EmptyCommand(t *testing.T) {
	mode := NewCommandMode()
	buf := buffer.New()

	// Execute empty command with Enter
	event := ui.KeyEvent{Action: ui.KeyActionEnter}
	result := mode.HandleInput(event, buf)

	if !result.Handled {
		t.Error("expected enter to be handled")
	}
	if result.ExitEditor {
		t.Error("expected empty command not to exit editor")
	}
	if result.SwitchToMode == nil || *result.SwitchToMode != ModeNormal {
		t.Error("expected empty command to switch back to Normal mode")
	}
}

func TestCommandMode_OnEnterExit(t *testing.T) {
	mode := NewCommandMode()
	buf := buffer.New()

	// Add some command text
	event := ui.KeyEvent{Action: ui.KeyActionChar, Rune: 'w'}
	mode.HandleInput(event, buf)

	// OnEnter should clear command line
	mode.OnEnter(buf)
	if mode.GetCommandLine() != ":" {
		t.Errorf("expected command line to be cleared on enter, got %q", mode.GetCommandLine())
	}

	// Add command text again
	event = ui.KeyEvent{Action: ui.KeyActionChar, Rune: 'q'}
	mode.HandleInput(event, buf)

	// OnExit should clear command line
	mode.OnExit(buf)
	if mode.GetCommandLine() != ":" {
		t.Errorf("expected command line to be cleared on exit, got %q", mode.GetCommandLine())
	}
}

func TestCommandMode_GetStatusText(t *testing.T) {
	mode := NewCommandMode()
	expected := "-- COMMAND --"
	if mode.GetStatusText() != expected {
		t.Errorf("expected status text %q, got %q", expected, mode.GetStatusText())
	}
}