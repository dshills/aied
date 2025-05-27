package ui

import (
	"github.com/gdamore/tcell/v2"

	"github.com/dshills/aied/internal/buffer"
)

// Viewport represents the visible area of the buffer
type Viewport struct {
	StartLine int // First visible line (0-based)
	StartCol  int // First visible column (0-based)
	Width     int // Viewport width in characters
	Height    int // Viewport height in lines (excluding status line)
}

// Renderer handles drawing the buffer content to the screen
type Renderer struct {
	screen   *Screen
	viewport Viewport
	styles   *StyleConfig
}

// StyleConfig defines the visual styling for different elements
type StyleConfig struct {
	Normal     tcell.Style
	Cursor     tcell.Style
	StatusLine tcell.Style
	LineNumber tcell.Style
}

// NewRenderer creates a new renderer for the given screen
func NewRenderer(screen *Screen) *Renderer {
	width, height := screen.Size()
	
	// Reserve last line for status
	viewportHeight := height - 1
	if viewportHeight < 1 {
		viewportHeight = 1
	}

	return &Renderer{
		screen: screen,
		viewport: Viewport{
			StartLine: 0,
			StartCol:  0,
			Width:     width,
			Height:    viewportHeight,
		},
		styles: NewDefaultStyles(),
	}
}

// NewDefaultStyles creates the default color scheme
func NewDefaultStyles() *StyleConfig {
	return &StyleConfig{
		Normal:     tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack),
		Cursor:     tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorWhite),
		StatusLine: tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorSilver),
		LineNumber: tcell.StyleDefault.Foreground(tcell.ColorGray).Background(tcell.ColorBlack),
	}
}

// RenderBuffer draws the buffer content to the screen
func (r *Renderer) RenderBuffer(buf *buffer.Buffer) {
	r.screen.Clear()
	
	// Get buffer information
	cursor := buf.Cursor()
	lineCount := buf.LineCount()
	
	// Adjust viewport to keep cursor visible
	r.adjustViewport(cursor, lineCount)
	
	// Render visible lines
	for screenY := 0; screenY < r.viewport.Height; screenY++ {
		bufferLine := r.viewport.StartLine + screenY
		
		if bufferLine >= lineCount {
			// Past end of buffer, draw empty line
			r.renderEmptyLine(screenY)
			continue
		}
		
		line, err := buf.Line(bufferLine)
		if err != nil {
			// Error getting line, draw empty
			r.renderEmptyLine(screenY)
			continue
		}
		
		// Render the line with cursor highlighting
		r.renderLine(screenY, line, bufferLine, cursor)
	}
	
	// Render status line
	r.renderStatusLine(buf)
	
	r.screen.Show()
}

// adjustViewport ensures the cursor is visible by adjusting the viewport
func (r *Renderer) adjustViewport(cursor buffer.Position, lineCount int) {
	// Vertical scrolling
	if cursor.Line < r.viewport.StartLine {
		r.viewport.StartLine = cursor.Line
	} else if cursor.Line >= r.viewport.StartLine+r.viewport.Height {
		r.viewport.StartLine = cursor.Line - r.viewport.Height + 1
	}
	
	// Horizontal scrolling
	if cursor.Col < r.viewport.StartCol {
		r.viewport.StartCol = cursor.Col
	} else if cursor.Col >= r.viewport.StartCol+r.viewport.Width {
		r.viewport.StartCol = cursor.Col - r.viewport.Width + 1
	}
	
	// Ensure viewport doesn't go negative
	if r.viewport.StartLine < 0 {
		r.viewport.StartLine = 0
	}
	if r.viewport.StartCol < 0 {
		r.viewport.StartCol = 0
	}
}

// renderLine draws a single line with cursor highlighting
func (r *Renderer) renderLine(screenY int, line string, bufferLine int, cursor buffer.Position) {
	// Convert line to runes for proper unicode handling
	runes := []rune(line)
	
	for screenX := 0; screenX < r.viewport.Width; screenX++ {
		bufferCol := r.viewport.StartCol + screenX
		style := r.styles.Normal
		
		var ch rune = ' '
		
		// Get character if within line bounds
		if bufferCol < len(runes) {
			ch = runes[bufferCol]
		}
		
		// Highlight cursor position
		if bufferLine == cursor.Line && bufferCol == cursor.Col {
			style = r.styles.Cursor
			// Show a space if at end of line
			if bufferCol >= len(runes) {
				ch = ' '
			}
		}
		
		r.screen.SetCell(screenX, screenY, ch, style)
	}
}

// renderEmptyLine draws an empty line (past end of buffer)
func (r *Renderer) renderEmptyLine(screenY int) {
	for screenX := 0; screenX < r.viewport.Width; screenX++ {
		r.screen.SetCell(screenX, screenY, ' ', r.styles.Normal)
	}
}

// renderStatusLine draws the status line at the bottom
func (r *Renderer) renderStatusLine(buf *buffer.Buffer) {
	r.renderStatusLineWithMode(buf, "")
}

// renderStatusLineWithMode draws the status line with mode information
func (r *Renderer) renderStatusLineWithMode(buf *buffer.Buffer, modeText string) {
	_, height := r.screen.Size()
	statusY := height - 1
	
	cursor := buf.Cursor()
	filename := buf.Filename()
	if filename == "" {
		filename = "[No Name]"
	}
	
	modified := ""
	if buf.Modified() {
		modified = "[+]"
	}
	
	// Format: "filename [+] - Line: 1, Col: 1 - MODE"
	status := ""
	if len(filename) + len(modified) > 0 {
		status = filename + " " + modified + " - "
	}
	status += "Line: " + formatInt(cursor.Line+1) + ", Col: " + formatInt(cursor.Col+1)
	
	if modeText != "" {
		status += " - " + modeText
	}
	
	// Clear the status line
	for x := 0; x < r.viewport.Width; x++ {
		r.screen.SetCell(x, statusY, ' ', r.styles.StatusLine)
	}
	
	// Draw status text
	r.screen.SetText(0, statusY, status, r.styles.StatusLine)
}

// UpdateViewport updates the viewport size (called on resize)
func (r *Renderer) UpdateViewport(width, height int) {
	r.viewport.Width = width
	// Reserve last line for status
	r.viewport.Height = height - 1
	if r.viewport.Height < 1 {
		r.viewport.Height = 1
	}
}

// GetViewport returns the current viewport
func (r *Renderer) GetViewport() Viewport {
	return r.viewport
}

// formatInt converts an integer to string (simple implementation)
func formatInt(n int) string {
	if n == 0 {
		return "0"
	}
	
	digits := []rune{}
	for n > 0 {
		digits = append([]rune{rune('0'+n%10)}, digits...)
		n /= 10
	}
	
	return string(digits)
}