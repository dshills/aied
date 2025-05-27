package ui

import (
	"github.com/gdamore/tcell/v2"
)

// KeyAction represents different types of key actions
type KeyAction int

const (
	KeyActionNone KeyAction = iota
	KeyActionQuit
	KeyActionChar
	KeyActionBackspace
	KeyActionDelete
	KeyActionEnter
	KeyActionTab
	KeyActionEscape
	KeyActionUp
	KeyActionDown
	KeyActionLeft
	KeyActionRight
	KeyActionHome
	KeyActionEnd
	KeyActionPageUp
	KeyActionPageDown
	KeyActionCtrlC
	KeyActionCtrlD
	KeyActionCtrlS
	KeyActionCtrlQ
	KeyActionCtrlZ
	KeyActionResize
)

// KeyEvent represents a processed keyboard event
type KeyEvent struct {
	Action KeyAction
	Rune   rune
	Key    tcell.Key
	Mods   tcell.ModMask
}

// ResizeEvent represents a terminal resize event
type ResizeEvent struct {
	Width  int
	Height int
}

// EventProcessor handles terminal events and converts them to editor events
type EventProcessor struct {
	screen *Screen
}

// NewEventProcessor creates a new event processor
func NewEventProcessor(screen *Screen) *EventProcessor {
	return &EventProcessor{
		screen: screen,
	}
}

// ProcessEvent converts tcell events to editor-specific events
func (ep *EventProcessor) ProcessEvent(event tcell.Event) interface{} {
	switch ev := event.(type) {
	case *tcell.EventKey:
		return ep.processKeyEvent(ev)
	case *tcell.EventResize:
		return ep.processResizeEvent(ev)
	case *tcell.EventInterrupt:
		return KeyEvent{Action: KeyActionQuit}
	default:
		return nil
	}
}

// processKeyEvent converts tcell key events to KeyEvent
func (ep *EventProcessor) processKeyEvent(ev *tcell.EventKey) KeyEvent {
	keyEvent := KeyEvent{
		Key:  ev.Key(),
		Mods: ev.Modifiers(),
		Rune: ev.Rune(),
	}

	// Handle special keys first
	switch ev.Key() {
	case tcell.KeyEscape:
		keyEvent.Action = KeyActionEscape
	case tcell.KeyEnter:
		keyEvent.Action = KeyActionEnter
	case tcell.KeyTab:
		keyEvent.Action = KeyActionTab
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		keyEvent.Action = KeyActionBackspace
	case tcell.KeyDelete:
		keyEvent.Action = KeyActionDelete
	case tcell.KeyUp:
		keyEvent.Action = KeyActionUp
	case tcell.KeyDown:
		keyEvent.Action = KeyActionDown
	case tcell.KeyLeft:
		keyEvent.Action = KeyActionLeft
	case tcell.KeyRight:
		keyEvent.Action = KeyActionRight
	case tcell.KeyHome:
		keyEvent.Action = KeyActionHome
	case tcell.KeyEnd:
		keyEvent.Action = KeyActionEnd
	case tcell.KeyPgUp:
		keyEvent.Action = KeyActionPageUp
	case tcell.KeyPgDn:
		keyEvent.Action = KeyActionPageDown
	case tcell.KeyCtrlC:
		keyEvent.Action = KeyActionCtrlC
	case tcell.KeyCtrlD:
		keyEvent.Action = KeyActionCtrlD
	case tcell.KeyCtrlS:
		keyEvent.Action = KeyActionCtrlS
	case tcell.KeyCtrlQ:
		keyEvent.Action = KeyActionCtrlQ
	case tcell.KeyCtrlZ:
		keyEvent.Action = KeyActionCtrlZ
	case tcell.KeyRune:
		// Handle printable characters
		if ev.Rune() != 0 {
			keyEvent.Action = KeyActionChar
		}
	default:
		keyEvent.Action = KeyActionNone
	}

	// Handle Ctrl+C as quit
	if ev.Key() == tcell.KeyCtrlC {
		keyEvent.Action = KeyActionQuit
	}

	return keyEvent
}

// processResizeEvent converts tcell resize events to ResizeEvent
func (ep *EventProcessor) processResizeEvent(ev *tcell.EventResize) ResizeEvent {
	width, height := ev.Size()
	// Update screen size if we have a real screen
	if ep.screen.tcellScreen != nil {
		ep.screen.UpdateSize()
	} else {
		// For testing, manually update the size
		ep.screen.width = width
		ep.screen.height = height
	}
	return ResizeEvent{
		Width:  width,
		Height: height,
	}
}

// WaitForEvent blocks until an event is available and returns it
func (ep *EventProcessor) WaitForEvent() interface{} {
	event := ep.screen.PollEvent()
	return ep.ProcessEvent(event)
}

// IsQuitEvent checks if an event represents a quit action
func IsQuitEvent(event interface{}) bool {
	if keyEvent, ok := event.(KeyEvent); ok {
		return keyEvent.Action == KeyActionQuit
	}
	return false
}

// IsCharEvent checks if an event represents a character input
func IsCharEvent(event interface{}) bool {
	if keyEvent, ok := event.(KeyEvent); ok {
		return keyEvent.Action == KeyActionChar
	}
	return false
}

// IsMovementEvent checks if an event represents cursor movement
func IsMovementEvent(event interface{}) bool {
	if keyEvent, ok := event.(KeyEvent); ok {
		switch keyEvent.Action {
		case KeyActionUp, KeyActionDown, KeyActionLeft, KeyActionRight,
			KeyActionHome, KeyActionEnd, KeyActionPageUp, KeyActionPageDown:
			return true
		}
	}
	return false
}

// IsResizeEvent checks if an event represents a terminal resize
func IsResizeEvent(event interface{}) bool {
	_, ok := event.(ResizeEvent)
	return ok
}