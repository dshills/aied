package modes

import (
	"github.com/dshills/aied/internal/buffer"
	"github.com/dshills/aied/internal/ui"
)

// ModeType represents the different editor modes
type ModeType int

const (
	ModeNormal ModeType = iota
	ModeInsert
	ModeVisual
	ModeCommand
)

// String returns the string representation of the mode
func (m ModeType) String() string {
	switch m {
	case ModeNormal:
		return "NORMAL"
	case ModeInsert:
		return "INSERT"
	case ModeVisual:
		return "VISUAL"
	case ModeCommand:
		return "COMMAND"
	default:
		return "UNKNOWN"
	}
}

// ModeResult represents the result of processing an input in a mode
type ModeResult struct {
	SwitchToMode *ModeType // If non-nil, switch to this mode
	Handled      bool      // Whether the input was handled
	ExitEditor   bool      // Whether to exit the editor
}

// Mode interface defines the behavior that all editor modes must implement
type Mode interface {
	// Type returns the mode type
	Type() ModeType

	// HandleInput processes a keyboard event and returns the result
	HandleInput(event ui.KeyEvent, buf *buffer.Buffer) ModeResult

	// OnEnter is called when entering this mode
	OnEnter(buf *buffer.Buffer)

	// OnExit is called when leaving this mode
	OnExit(buf *buffer.Buffer)

	// GetStatusText returns mode-specific status information
	GetStatusText() string
}

// ModeManager manages the current mode and mode transitions
type ModeManager struct {
	currentMode Mode
	modes       map[ModeType]Mode
}

// NewModeManager creates a new mode manager
func NewModeManager() *ModeManager {
	mm := &ModeManager{
		modes: make(map[ModeType]Mode),
	}

	// Register all available modes
	mm.RegisterMode(NewNormalMode())
	mm.RegisterMode(NewInsertMode())
	mm.RegisterMode(NewVisualMode())

	// Start in Normal mode
	mm.SwitchToMode(ModeNormal, nil)

	return mm
}

// RegisterMode registers a mode with the manager
func (mm *ModeManager) RegisterMode(mode Mode) {
	mm.modes[mode.Type()] = mode
}

// SwitchToMode switches to the specified mode
func (mm *ModeManager) SwitchToMode(modeType ModeType, buf *buffer.Buffer) {
	if newMode, exists := mm.modes[modeType]; exists {
		// Exit current mode if we have one
		if mm.currentMode != nil && buf != nil {
			mm.currentMode.OnExit(buf)
		}

		// Switch to new mode
		mm.currentMode = newMode
		
		// Only call OnEnter if we have a buffer
		if buf != nil {
			mm.currentMode.OnEnter(buf)
		}
	}
}

// HandleInput processes input through the current mode
func (mm *ModeManager) HandleInput(event ui.KeyEvent, buf *buffer.Buffer) ModeResult {
	if mm.currentMode == nil {
		return ModeResult{Handled: false}
	}

	result := mm.currentMode.HandleInput(event, buf)

	// Handle mode switching
	if result.SwitchToMode != nil {
		mm.SwitchToMode(*result.SwitchToMode, buf)
	}

	return result
}

// CurrentMode returns the current mode
func (mm *ModeManager) CurrentMode() Mode {
	return mm.currentMode
}

// CurrentModeType returns the current mode type
func (mm *ModeManager) CurrentModeType() ModeType {
	if mm.currentMode == nil {
		return ModeNormal
	}
	return mm.currentMode.Type()
}

// GetStatusText returns the current mode's status text
func (mm *ModeManager) GetStatusText() string {
	if mm.currentMode == nil {
		return ""
	}
	return mm.currentMode.GetStatusText()
}