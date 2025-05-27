package ui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
)

// Screen manages the terminal display and input
type Screen struct {
	tcellScreen tcell.Screen
	width       int
	height      int
	running     bool
}

// NewScreen creates and initializes a new terminal screen
func NewScreen() (*Screen, error) {
	tcellScreen, err := tcell.NewScreen()
	if err != nil {
		return nil, fmt.Errorf("failed to create screen: %w", err)
	}

	if err := tcellScreen.Init(); err != nil {
		return nil, fmt.Errorf("failed to initialize screen: %w", err)
	}

	width, height := tcellScreen.Size()

	screen := &Screen{
		tcellScreen: tcellScreen,
		width:       width,
		height:      height,
		running:     true,
	}

	// Set up initial screen
	tcellScreen.SetStyle(tcell.StyleDefault)
	tcellScreen.Clear()

	return screen, nil
}

// Close shuts down the screen and restores terminal
func (s *Screen) Close() {
	if s.tcellScreen != nil {
		s.running = false
		s.tcellScreen.Fini()
	}
}

// Size returns the current screen dimensions
func (s *Screen) Size() (int, int) {
	return s.width, s.height
}

// IsRunning returns whether the screen is active
func (s *Screen) IsRunning() bool {
	return s.running
}

// Clear clears the entire screen
func (s *Screen) Clear() {
	s.tcellScreen.Clear()
}

// SetCell sets a character at the specified position
func (s *Screen) SetCell(x, y int, ch rune, style tcell.Style) {
	if x >= 0 && x < s.width && y >= 0 && y < s.height {
		s.tcellScreen.SetContent(x, y, ch, nil, style)
	}
}

// SetText sets a string starting at the specified position
func (s *Screen) SetText(x, y int, text string, style tcell.Style) {
	for i, ch := range text {
		if x+i >= s.width {
			break
		}
		s.SetCell(x+i, y, ch, style)
	}
}

// Show refreshes the screen to display all changes
func (s *Screen) Show() {
	s.tcellScreen.Show()
}

// PollEvent returns the next input event
func (s *Screen) PollEvent() tcell.Event {
	return s.tcellScreen.PollEvent()
}

// PostQuit posts a quit event to stop the screen
func (s *Screen) PostQuit() {
	s.running = false
	s.tcellScreen.PostEvent(tcell.NewEventInterrupt(nil))
}

// UpdateSize updates the screen size (typically called on resize)
func (s *Screen) UpdateSize() {
	if s.tcellScreen != nil {
		s.width, s.height = s.tcellScreen.Size()
	}
}

// SetCursor sets the cursor position (if supported by terminal)
func (s *Screen) SetCursor(x, y int) {
	// Note: tcell doesn't directly support cursor positioning
	// This will be handled by drawing a highlighted character
	// at the cursor position during rendering
}