package ui

import (
	"fmt"

	"github.com/dshills/aied/internal/buffer"
	"github.com/gdamore/tcell/v2"
)

// UI manages the terminal user interface
type UI struct {
	screen           *Screen
	renderer         *Renderer
	processor        *EventProcessor
	completionPopup  *CompletionPopup
	running          bool
}

// NewUI creates a new terminal UI
func NewUI() (*UI, error) {
	screen, err := NewScreen()
	if err != nil {
		return nil, fmt.Errorf("failed to create screen: %w", err)
	}

	renderer := NewRenderer(screen)
	processor := NewEventProcessor(screen)
	completionPopup := NewCompletionPopup()

	return &UI{
		screen:          screen,
		renderer:        renderer,
		processor:       processor,
		completionPopup: completionPopup,
		running:         true,
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
	
	// Render completion popup if visible
	if ui.completionPopup.IsVisible() {
		ui.completionPopup.Render(ui.renderer.screen, ui.renderer.styles)
		ui.renderer.screen.Show()
	}
}

// RenderWithMode draws the buffer to the screen with mode information
func (ui *UI) RenderWithMode(buf *buffer.Buffer, modeText string) {
	ui.RenderWithModeAndCommand(buf, modeText, "", "")
}

// RenderWithModeAndCommand draws the buffer with mode and command line information
func (ui *UI) RenderWithModeAndCommand(buf *buffer.Buffer, modeText, commandLine, message string) {
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
		
		// Render the line with cursor and diagnostic highlighting
		diagnostics := buf.GetDiagnosticsForLine(bufferLine)
		if len(diagnostics) > 0 {
			ui.renderer.renderLineWithDiagnostics(screenY, line, bufferLine, cursor, diagnostics)
		} else {
			ui.renderer.renderLine(screenY, line, bufferLine, cursor)
		}
	}
	
	// Render status line with mode, command line, and message
	ui.renderer.renderStatusLineWithModeAndCommand(buf, modeText, commandLine, message)
	
	// Render completion popup if visible
	if ui.completionPopup.IsVisible() {
		ui.completionPopup.Render(ui.renderer.screen, ui.renderer.styles)
	}
	
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

// ShowCompletions displays the completion popup with the given items
func (ui *UI) ShowCompletions(items []CompletionItem, x, y int) {
	ui.completionPopup.SetItems(items)
	ui.completionPopup.Show(x, y)
}

// HideCompletions hides the completion popup
func (ui *UI) HideCompletions() {
	ui.completionPopup.Hide()
}

// IsCompletionVisible returns whether the completion popup is visible
func (ui *UI) IsCompletionVisible() bool {
	return ui.completionPopup.IsVisible()
}

// MoveCompletionUp moves the selection up in the completion popup
func (ui *UI) MoveCompletionUp() {
	ui.completionPopup.MoveUp()
}

// MoveCompletionDown moves the selection down in the completion popup
func (ui *UI) MoveCompletionDown() {
	ui.completionPopup.MoveDown()
}

// GetSelectedCompletion returns the currently selected completion item
func (ui *UI) GetSelectedCompletion() *CompletionItem {
	return ui.completionPopup.GetSelectedItem()
}

// GetScreen returns the underlying screen for direct rendering
func (ui *UI) GetScreen() *Screen {
	return ui.screen
}

// GetStyle returns a style for the given name
func GetStyle(name string) tcell.Style {
	switch name {
	case "normal":
		return tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)
	case "selected":
		return tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorBlue)
	case "border":
		return tcell.StyleDefault.Foreground(tcell.ColorGray).Background(tcell.ColorBlack)
	default:
		return tcell.StyleDefault
	}
}