package ui

import (
	"fmt"

	"github.com/dshills/aied/internal/buffer"
)

// UI manages the terminal user interface
type UI struct {
	screen    *Screen
	renderer  *Renderer
	processor *EventProcessor
	running   bool
}

// NewUI creates a new terminal UI
func NewUI() (*UI, error) {
	screen, err := NewScreen()
	if err != nil {
		return nil, fmt.Errorf("failed to create screen: %w", err)
	}

	renderer := NewRenderer(screen)
	processor := NewEventProcessor(screen)

	return &UI{
		screen:    screen,
		renderer:  renderer,
		processor: processor,
		running:   true,
	}, nil
}

// Close shuts down the UI and restores the terminal
func (ui *UI) Close() {
	ui.running = false
	if ui.screen != nil {
		ui.screen.Close()
	}
}

// IsRunning returns whether the UI is active
func (ui *UI) IsRunning() bool {
	return ui.running && ui.screen.IsRunning()
}

// Render draws the buffer to the screen
func (ui *UI) Render(buf *buffer.Buffer) {
	ui.renderer.RenderBuffer(buf)
}

// RenderWithMode draws the buffer to the screen with mode information
func (ui *UI) RenderWithMode(buf *buffer.Buffer, modeText string) {
	ui.renderer.screen.Clear()
	
	// Get buffer information
	cursor := buf.Cursor()
	lineCount := buf.LineCount()
	
	// Adjust viewport to keep cursor visible
	ui.renderer.adjustViewport(cursor, lineCount)
	
	// Render visible lines
	for screenY := 0; screenY < ui.renderer.viewport.Height; screenY++ {
		bufferLine := ui.renderer.viewport.StartLine + screenY
		
		if bufferLine >= lineCount {
			// Past end of buffer, draw empty line
			ui.renderer.renderEmptyLine(screenY)
			continue
		}
		
		line, err := buf.Line(bufferLine)
		if err != nil {
			// Error getting line, draw empty
			ui.renderer.renderEmptyLine(screenY)
			continue
		}
		
		// Render the line with cursor highlighting
		ui.renderer.renderLine(screenY, line, bufferLine, cursor)
	}
	
	// Render status line with mode
	ui.renderer.renderStatusLineWithMode(buf, modeText)
	
	ui.renderer.screen.Show()
}

// WaitForEvent blocks until an input event is available
func (ui *UI) WaitForEvent() interface{} {
	return ui.processor.WaitForEvent()
}

// HandleResize processes a terminal resize event
func (ui *UI) HandleResize(event ResizeEvent) {
	ui.renderer.UpdateViewport(event.Width, event.Height)
}

// GetSize returns the current terminal size
func (ui *UI) GetSize() (int, int) {
	return ui.screen.Size()
}

// ShowMessage displays a temporary message (for future use)
func (ui *UI) ShowMessage(message string) {
	// TODO: Implement message display in command line area
	// For now, this is a placeholder for future message functionality
	_ = message
}

// Clear clears the screen
func (ui *UI) Clear() {
	ui.screen.Clear()
}

// SetCursor sets the visual cursor position
func (ui *UI) SetCursor(x, y int) {
	ui.screen.SetCursor(x, y)
}

// PostQuit signals the UI to quit
func (ui *UI) PostQuit() {
	ui.screen.PostQuit()
}

// GetViewport returns the current viewport information
func (ui *UI) GetViewport() Viewport {
	return ui.renderer.GetViewport()
}