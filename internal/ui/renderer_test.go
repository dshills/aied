package ui

import (
	"testing"

	"github.com/gdamore/tcell/v2"

	"github.com/dshills/aied/internal/buffer"
)

func TestNewRenderer(t *testing.T) {
	// Create a mock screen with known dimensions
	screen := &Screen{width: 80, height: 24}
	renderer := NewRenderer(screen)

	if renderer.viewport.Width != 80 {
		t.Errorf("expected viewport width 80, got %d", renderer.viewport.Width)
	}

	// Height should be screen height - 1 (for status line)
	expectedHeight := 23
	if renderer.viewport.Height != expectedHeight {
		t.Errorf("expected viewport height %d, got %d", expectedHeight, renderer.viewport.Height)
	}

	if renderer.viewport.StartLine != 0 {
		t.Errorf("expected viewport start line 0, got %d", renderer.viewport.StartLine)
	}

	if renderer.viewport.StartCol != 0 {
		t.Errorf("expected viewport start col 0, got %d", renderer.viewport.StartCol)
	}
}

func TestRenderer_AdjustViewport(t *testing.T) {
	screen := &Screen{width: 10, height: 5}
	renderer := NewRenderer(screen)
	// viewport height is 4 (5 - 1 for status line)

	tests := []struct {
		name               string
		cursor             buffer.Position
		lineCount          int
		expectedStartLine  int
		expectedStartCol   int
	}{
		{
			name:              "cursor at origin",
			cursor:            buffer.Position{Line: 0, Col: 0},
			lineCount:         10,
			expectedStartLine: 0,
			expectedStartCol:  0,
		},
		{
			name:              "cursor past viewport height",
			cursor:            buffer.Position{Line: 5, Col: 0},
			lineCount:         10,
			expectedStartLine: 2, // 5 - 4 + 1
			expectedStartCol:  0,
		},
		{
			name:              "cursor past viewport width",
			cursor:            buffer.Position{Line: 0, Col: 15},
			lineCount:         10,
			expectedStartLine: 0,
			expectedStartCol:  6, // 15 - 10 + 1
		},
		{
			name:              "cursor before viewport",
			cursor:            buffer.Position{Line: 0, Col: 0},
			lineCount:         10,
			expectedStartLine: 0,
			expectedStartCol:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset viewport
			renderer.viewport.StartLine = 0
			renderer.viewport.StartCol = 0

			// For test with cursor past viewport, set initial viewport position
			if tt.name == "cursor before viewport" {
				renderer.viewport.StartLine = 5
				renderer.viewport.StartCol = 5
			}

			renderer.adjustViewport(tt.cursor, tt.lineCount)

			if renderer.viewport.StartLine != tt.expectedStartLine {
				t.Errorf("expected start line %d, got %d", tt.expectedStartLine, renderer.viewport.StartLine)
			}

			if renderer.viewport.StartCol != tt.expectedStartCol {
				t.Errorf("expected start col %d, got %d", tt.expectedStartCol, renderer.viewport.StartCol)
			}
		})
	}
}

func TestRenderer_UpdateViewport(t *testing.T) {
	screen := &Screen{width: 80, height: 24}
	renderer := NewRenderer(screen)

	// Update to new size
	newWidth, newHeight := 120, 30
	renderer.UpdateViewport(newWidth, newHeight)

	if renderer.viewport.Width != newWidth {
		t.Errorf("expected viewport width %d, got %d", newWidth, renderer.viewport.Width)
	}

	// Height should be newHeight - 1 (for status line)
	expectedHeight := newHeight - 1
	if renderer.viewport.Height != expectedHeight {
		t.Errorf("expected viewport height %d, got %d", expectedHeight, renderer.viewport.Height)
	}
}

func TestRenderer_UpdateViewport_MinHeight(t *testing.T) {
	screen := &Screen{width: 80, height: 24}
	renderer := NewRenderer(screen)

	// Update to very small height
	renderer.UpdateViewport(80, 1)

	// Height should be at least 1
	if renderer.viewport.Height != 1 {
		t.Errorf("expected minimum viewport height 1, got %d", renderer.viewport.Height)
	}
}

func TestFormatInt(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{0, "0"},
		{1, "1"},
		{10, "10"},
		{123, "123"},
		{9999, "9999"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := formatInt(tt.input)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestNewDefaultStyles(t *testing.T) {
	styles := NewDefaultStyles()

	if styles == nil {
		t.Fatal("expected styles to be created")
	}

	// Test that styles are created (we can't easily test the specific values
	// without exposing internal tcell Style fields, so we just verify they exist)
	if styles.Normal == (tcell.Style{}) {
		t.Error("expected normal style to be set")
	}

	if styles.Cursor == (tcell.Style{}) {
		t.Error("expected cursor style to be set")
	}

	if styles.StatusLine == (tcell.Style{}) {
		t.Error("expected status line style to be set")
	}

	if styles.LineNumber == (tcell.Style{}) {
		t.Error("expected line number style to be set")
	}
}