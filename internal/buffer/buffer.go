package buffer

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Position represents a cursor position in the buffer
type Position struct {
	Line int // 0-based line number
	Col  int // 0-based column number
}

// Buffer represents a text buffer with cursor tracking
type Buffer struct {
	lines    []string // Text content stored as lines
	cursor   Position // Current cursor position
	filename string   // Associated filename (empty for new buffer)
	modified bool     // Whether buffer has unsaved changes
}

// New creates a new empty buffer
func New() *Buffer {
	return &Buffer{
		lines:    []string{""},
		cursor:   Position{Line: 0, Col: 0},
		filename: "",
		modified: false,
	}
}

// NewFromFile creates a buffer from an existing file
func NewFromFile(filename string) (*Buffer, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %q: %w", filename, err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read file %q: %w", filename, err)
	}

	// Ensure at least one empty line
	if len(lines) == 0 {
		lines = []string{""}
	}

	return &Buffer{
		lines:    lines,
		cursor:   Position{Line: 0, Col: 0},
		filename: filename,
		modified: false,
	}, nil
}

// LineCount returns the number of lines in the buffer
func (b *Buffer) LineCount() int {
	return len(b.lines)
}

// Line returns the content of the specified line (0-based)
func (b *Buffer) Line(lineNum int) (string, error) {
	if lineNum < 0 || lineNum >= len(b.lines) {
		return "", fmt.Errorf("line %d out of range [0-%d]", lineNum, len(b.lines)-1)
	}
	return b.lines[lineNum], nil
}

// CurrentLine returns the content of the current cursor line
func (b *Buffer) CurrentLine() string {
	if b.cursor.Line >= 0 && b.cursor.Line < len(b.lines) {
		return b.lines[b.cursor.Line]
	}
	return ""
}

// Cursor returns the current cursor position
func (b *Buffer) Cursor() Position {
	return b.cursor
}

// SetCursor moves the cursor to the specified position with bounds checking
func (b *Buffer) SetCursor(pos Position) {
	// Clamp line to valid range
	if pos.Line < 0 {
		pos.Line = 0
	} else if pos.Line >= len(b.lines) {
		pos.Line = len(b.lines) - 1
	}

	// Clamp column to valid range for the line
	lineLen := len(b.lines[pos.Line])
	if pos.Col < 0 {
		pos.Col = 0
	} else if pos.Col > lineLen {
		pos.Col = lineLen
	}

	b.cursor = pos
}

// MoveCursor moves the cursor by the specified delta
func (b *Buffer) MoveCursor(deltaLine, deltaCol int) {
	newPos := Position{
		Line: b.cursor.Line + deltaLine,
		Col:  b.cursor.Col + deltaCol,
	}
	b.SetCursor(newPos)
}

// Filename returns the associated filename
func (b *Buffer) Filename() string {
	return b.filename
}

// SetFilename sets the associated filename
func (b *Buffer) SetFilename(filename string) {
	b.filename = filename
}

// Modified returns whether the buffer has unsaved changes
func (b *Buffer) Modified() bool {
	return b.modified
}

// setModified marks the buffer as modified or unmodified
func (b *Buffer) setModified(modified bool) {
	b.modified = modified
}

// String returns the entire buffer content as a string
func (b *Buffer) String() string {
	return strings.Join(b.lines, "\n")
}

// Lines returns a copy of all lines in the buffer
func (b *Buffer) Lines() []string {
	result := make([]string, len(b.lines))
	copy(result, b.lines)
	return result
}

// InsertChar inserts a character at the current cursor position
func (b *Buffer) InsertChar(ch rune) error {
	if b.cursor.Line < 0 || b.cursor.Line >= len(b.lines) {
		return fmt.Errorf("cursor line %d out of range", b.cursor.Line)
	}

	line := b.lines[b.cursor.Line]
	if b.cursor.Col < 0 || b.cursor.Col > len(line) {
		return fmt.Errorf("cursor column %d out of range for line length %d", b.cursor.Col, len(line))
	}

	// Insert character at cursor position
	newLine := line[:b.cursor.Col] + string(ch) + line[b.cursor.Col:]
	b.lines[b.cursor.Line] = newLine

	// Move cursor forward
	b.cursor.Col++
	b.setModified(true)

	return nil
}

// DeleteChar deletes the character at the current cursor position
func (b *Buffer) DeleteChar() error {
	if b.cursor.Line < 0 || b.cursor.Line >= len(b.lines) {
		return fmt.Errorf("cursor line %d out of range", b.cursor.Line)
	}

	line := b.lines[b.cursor.Line]
	if b.cursor.Col < 0 || b.cursor.Col >= len(line) {
		return fmt.Errorf("cursor column %d out of range for line length %d", b.cursor.Col, len(line))
	}

	// Can't delete if cursor is at end of line
	if b.cursor.Col >= len(line) {
		return fmt.Errorf("cannot delete at end of line")
	}

	// Delete character at cursor position
	newLine := line[:b.cursor.Col] + line[b.cursor.Col+1:]
	b.lines[b.cursor.Line] = newLine
	b.setModified(true)

	return nil
}

// Backspace deletes the character before the cursor position
func (b *Buffer) Backspace() error {
	if b.cursor.Line < 0 || b.cursor.Line >= len(b.lines) {
		return fmt.Errorf("cursor line %d out of range", b.cursor.Line)
	}

	// If at beginning of line, try to join with previous line
	if b.cursor.Col == 0 {
		if b.cursor.Line == 0 {
			return fmt.Errorf("cannot backspace at beginning of buffer")
		}

		// Get content of current line
		currentLine := b.lines[b.cursor.Line]
		
		// Move cursor to end of previous line
		prevLineLen := len(b.lines[b.cursor.Line-1])
		b.cursor.Line--
		b.cursor.Col = prevLineLen

		// Append current line content to previous line
		b.lines[b.cursor.Line] += currentLine

		// Remove the now-empty line
		b.lines = append(b.lines[:b.cursor.Line+1], b.lines[b.cursor.Line+2:]...)
		b.setModified(true)

		return nil
	}

	// Delete character before cursor
	line := b.lines[b.cursor.Line]
	newLine := line[:b.cursor.Col-1] + line[b.cursor.Col:]
	b.lines[b.cursor.Line] = newLine

	// Move cursor back
	b.cursor.Col--
	b.setModified(true)

	return nil
}

// InsertLine inserts a new line at the current cursor position
// The text after the cursor on the current line moves to the new line
func (b *Buffer) InsertLine() error {
	if b.cursor.Line < 0 || b.cursor.Line >= len(b.lines) {
		return fmt.Errorf("cursor line %d out of range", b.cursor.Line)
	}

	currentLine := b.lines[b.cursor.Line]
	if b.cursor.Col < 0 || b.cursor.Col > len(currentLine) {
		return fmt.Errorf("cursor column %d out of range for line length %d", b.cursor.Col, len(currentLine))
	}

	// Split the current line at cursor position
	leftPart := currentLine[:b.cursor.Col]
	rightPart := currentLine[b.cursor.Col:]

	// Update current line with left part
	b.lines[b.cursor.Line] = leftPart

	// Insert new line with right part
	newLines := make([]string, len(b.lines)+1)
	copy(newLines[:b.cursor.Line+1], b.lines[:b.cursor.Line+1])
	newLines[b.cursor.Line+1] = rightPart
	copy(newLines[b.cursor.Line+2:], b.lines[b.cursor.Line+1:])
	b.lines = newLines

	// Move cursor to beginning of new line
	b.cursor.Line++
	b.cursor.Col = 0
	b.setModified(true)

	return nil
}

// InsertEmptyLine inserts an empty line above the current line
func (b *Buffer) InsertEmptyLine() error {
	if b.cursor.Line < 0 || b.cursor.Line >= len(b.lines) {
		return fmt.Errorf("cursor line %d out of range", b.cursor.Line)
	}

	// Insert empty line at current position
	newLines := make([]string, len(b.lines)+1)
	copy(newLines[:b.cursor.Line], b.lines[:b.cursor.Line])
	newLines[b.cursor.Line] = ""
	copy(newLines[b.cursor.Line+1:], b.lines[b.cursor.Line:])
	b.lines = newLines

	// Cursor stays at the new empty line
	b.cursor.Col = 0
	b.setModified(true)

	return nil
}

// DeleteLine deletes the current line
func (b *Buffer) DeleteLine() error {
	if len(b.lines) <= 1 {
		// Don't delete the last line, just clear it
		b.lines[0] = ""
		b.cursor = Position{Line: 0, Col: 0}
		b.setModified(true)
		return nil
	}

	if b.cursor.Line < 0 || b.cursor.Line >= len(b.lines) {
		return fmt.Errorf("cursor line %d out of range", b.cursor.Line)
	}

	// Remove the current line
	newLines := make([]string, len(b.lines)-1)
	copy(newLines[:b.cursor.Line], b.lines[:b.cursor.Line])
	copy(newLines[b.cursor.Line:], b.lines[b.cursor.Line+1:])
	b.lines = newLines

	// Adjust cursor position
	if b.cursor.Line >= len(b.lines) {
		b.cursor.Line = len(b.lines) - 1
	}
	
	// Clamp cursor column to line length
	if b.cursor.Line >= 0 && b.cursor.Line < len(b.lines) {
		lineLen := len(b.lines[b.cursor.Line])
		if b.cursor.Col > lineLen {
			b.cursor.Col = lineLen
		}
	}

	b.setModified(true)
	return nil
}

// JoinLines joins the current line with the next line
func (b *Buffer) JoinLines() error {
	if b.cursor.Line < 0 || b.cursor.Line >= len(b.lines)-1 {
		return fmt.Errorf("cannot join line %d (no next line)", b.cursor.Line)
	}

	currentLine := b.lines[b.cursor.Line]
	nextLine := b.lines[b.cursor.Line+1]

	// Join lines with a space if both have content
	separator := ""
	if len(currentLine) > 0 && len(nextLine) > 0 {
		separator = " "
	}

	// Combine the lines
	b.lines[b.cursor.Line] = currentLine + separator + nextLine

	// Remove the next line
	newLines := make([]string, len(b.lines)-1)
	copy(newLines[:b.cursor.Line+1], b.lines[:b.cursor.Line+1])
	copy(newLines[b.cursor.Line+1:], b.lines[b.cursor.Line+2:])
	b.lines = newLines

	b.setModified(true)
	return nil
}

// Save writes the buffer content to its associated file
func (b *Buffer) Save() error {
	if b.filename == "" {
		return fmt.Errorf("no filename set for buffer")
	}
	return b.SaveAs(b.filename)
}

// SaveAs writes the buffer content to the specified file
func (b *Buffer) SaveAs(filename string) error {
	if filename == "" {
		return fmt.Errorf("filename cannot be empty")
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file %q: %w", filename, err)
	}
	defer file.Close()

	content := b.String()
	if _, err := file.WriteString(content); err != nil {
		return fmt.Errorf("failed to write to file %q: %w", filename, err)
	}

	// Update buffer state
	b.filename = filename
	b.setModified(false)

	return nil
}