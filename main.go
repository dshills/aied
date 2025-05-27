package main

import (
	"fmt"
	"os"

	"github.com/dshills/aied/internal/buffer"
	"github.com/dshills/aied/internal/ui"
)

func main() {
	// Create a new buffer
	var buf *buffer.Buffer
	var err error

	// Check if a filename was provided
	if len(os.Args) > 1 {
		filename := os.Args[1]
		buf, err = buffer.NewFromFile(filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening file %q: %v\n", filename, err)
			os.Exit(1)
		}
	} else {
		buf = buffer.New()
	}

	// Create the terminal UI
	terminalUI, err := ui.NewUI()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize terminal UI: %v\n", err)
		os.Exit(1)
	}
	defer terminalUI.Close()

	// Initial render
	terminalUI.Render(buf)

	// Main event loop
	for terminalUI.IsRunning() {
		event := terminalUI.WaitForEvent()

		switch ev := event.(type) {
		case ui.KeyEvent:
			if handleKeyEvent(ev, buf, terminalUI) {
				break // quit requested
			}
		case ui.ResizeEvent:
			terminalUI.HandleResize(ev)
		}

		// Re-render after any changes
		terminalUI.Render(buf)
	}
}

// handleKeyEvent processes keyboard input and returns true if quit was requested
func handleKeyEvent(event ui.KeyEvent, buf *buffer.Buffer, terminalUI *ui.UI) bool {
	switch event.Action {
	case ui.KeyActionQuit, ui.KeyActionCtrlC:
		return true

	case ui.KeyActionChar:
		// Insert character
		buf.InsertChar(event.Rune)

	case ui.KeyActionBackspace:
		buf.Backspace()

	case ui.KeyActionDelete:
		buf.DeleteChar()

	case ui.KeyActionEnter:
		buf.InsertLine()

	case ui.KeyActionUp:
		buf.MoveCursor(-1, 0)

	case ui.KeyActionDown:
		buf.MoveCursor(1, 0)

	case ui.KeyActionLeft:
		buf.MoveCursor(0, -1)

	case ui.KeyActionRight:
		buf.MoveCursor(0, 1)

	case ui.KeyActionHome:
		cursor := buf.Cursor()
		buf.SetCursor(buffer.Position{Line: cursor.Line, Col: 0})

	case ui.KeyActionEnd:
		cursor := buf.Cursor()
		lineLen := len(buf.CurrentLine())
		buf.SetCursor(buffer.Position{Line: cursor.Line, Col: lineLen})

	case ui.KeyActionCtrlS:
		// Save file
		if buf.Filename() != "" {
			buf.Save()
		}

	default:
		// Ignore other keys for now
	}

	return false
}
