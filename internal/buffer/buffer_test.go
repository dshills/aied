package buffer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNew(t *testing.T) {
	buf := New()
	
	if buf.LineCount() != 1 {
		t.Errorf("expected 1 line, got %d", buf.LineCount())
	}
	
	if buf.CurrentLine() != "" {
		t.Errorf("expected empty line, got %q", buf.CurrentLine())
	}
	
	cursor := buf.Cursor()
	if cursor.Line != 0 || cursor.Col != 0 {
		t.Errorf("expected cursor at (0,0), got (%d,%d)", cursor.Line, cursor.Col)
	}
	
	if buf.Modified() {
		t.Error("new buffer should not be modified")
	}
	
	if buf.Filename() != "" {
		t.Errorf("new buffer should have empty filename, got %q", buf.Filename())
	}
}

func TestNewFromFile(t *testing.T) {
	// Create a temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := "line 1\nline 2\nline 3"
	
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	
	buf, err := NewFromFile(testFile)
	if err != nil {
		t.Fatalf("failed to create buffer from file: %v", err)
	}
	
	if buf.LineCount() != 3 {
		t.Errorf("expected 3 lines, got %d", buf.LineCount())
	}
	
	line, err := buf.Line(0)
	if err != nil {
		t.Errorf("failed to get line 0: %v", err)
	}
	if line != "line 1" {
		t.Errorf("expected 'line 1', got %q", line)
	}
	
	if buf.Filename() != testFile {
		t.Errorf("expected filename %q, got %q", testFile, buf.Filename())
	}
	
	if buf.Modified() {
		t.Error("buffer loaded from file should not be modified")
	}
}

func TestNewFromNonexistentFile(t *testing.T) {
	_, err := NewFromFile("/nonexistent/file.txt")
	if err == nil {
		t.Error("expected error when opening nonexistent file")
	}
}

func TestInsertChar(t *testing.T) {
	buf := New()
	
	err := buf.InsertChar('h')
	if err != nil {
		t.Errorf("failed to insert char: %v", err)
	}
	
	if buf.CurrentLine() != "h" {
		t.Errorf("expected 'h', got %q", buf.CurrentLine())
	}
	
	cursor := buf.Cursor()
	if cursor.Line != 0 || cursor.Col != 1 {
		t.Errorf("expected cursor at (0,1), got (%d,%d)", cursor.Line, cursor.Col)
	}
	
	if !buf.Modified() {
		t.Error("buffer should be modified after inserting char")
	}
	
	// Insert more chars
	buf.InsertChar('e')
	buf.InsertChar('l')
	buf.InsertChar('l')
	buf.InsertChar('o')
	
	if buf.CurrentLine() != "hello" {
		t.Errorf("expected 'hello', got %q", buf.CurrentLine())
	}
}

func TestDeleteChar(t *testing.T) {
	buf := New()
	buf.InsertChar('h')
	buf.InsertChar('e')
	buf.InsertChar('l')
	buf.InsertChar('l')
	buf.InsertChar('o')
	
	// Move cursor back
	buf.SetCursor(Position{Line: 0, Col: 2})
	
	err := buf.DeleteChar()
	if err != nil {
		t.Errorf("failed to delete char: %v", err)
	}
	
	if buf.CurrentLine() != "helo" {
		t.Errorf("expected 'helo', got %q", buf.CurrentLine())
	}
	
	cursor := buf.Cursor()
	if cursor.Line != 0 || cursor.Col != 2 {
		t.Errorf("expected cursor at (0,2), got (%d,%d)", cursor.Line, cursor.Col)
	}
}

func TestBackspace(t *testing.T) {
	buf := New()
	buf.InsertChar('h')
	buf.InsertChar('e')
	buf.InsertChar('l')
	buf.InsertChar('l')
	buf.InsertChar('o')
	
	err := buf.Backspace()
	if err != nil {
		t.Errorf("failed to backspace: %v", err)
	}
	
	if buf.CurrentLine() != "hell" {
		t.Errorf("expected 'hell', got %q", buf.CurrentLine())
	}
	
	cursor := buf.Cursor()
	if cursor.Line != 0 || cursor.Col != 4 {
		t.Errorf("expected cursor at (0,4), got (%d,%d)", cursor.Line, cursor.Col)
	}
}

func TestBackspaceJoinLines(t *testing.T) {
	buf := New()
	buf.InsertChar('h')
	buf.InsertChar('i')
	buf.InsertLine()
	buf.InsertChar('b')
	buf.InsertChar('y')
	buf.InsertChar('e')
	
	// Buffer should be: "hi\nbye"
	if buf.LineCount() != 2 {
		t.Errorf("expected 2 lines, got %d", buf.LineCount())
	}
	
	// Move cursor to beginning of second line
	buf.SetCursor(Position{Line: 1, Col: 0})
	
	err := buf.Backspace()
	if err != nil {
		t.Errorf("failed to backspace at line start: %v", err)
	}
	
	if buf.LineCount() != 1 {
		t.Errorf("expected 1 line after join, got %d", buf.LineCount())
	}
	
	if buf.CurrentLine() != "hibye" {
		t.Errorf("expected 'hibye', got %q", buf.CurrentLine())
	}
	
	cursor := buf.Cursor()
	if cursor.Line != 0 || cursor.Col != 2 {
		t.Errorf("expected cursor at (0,2), got (%d,%d)", cursor.Line, cursor.Col)
	}
}

func TestInsertLine(t *testing.T) {
	buf := New()
	buf.InsertChar('h')
	buf.InsertChar('e')
	buf.InsertChar('l')
	buf.InsertChar('l')
	buf.InsertChar('o')
	
	// Move cursor to middle
	buf.SetCursor(Position{Line: 0, Col: 2})
	
	err := buf.InsertLine()
	if err != nil {
		t.Errorf("failed to insert line: %v", err)
	}
	
	if buf.LineCount() != 2 {
		t.Errorf("expected 2 lines, got %d", buf.LineCount())
	}
	
	line0, _ := buf.Line(0)
	if line0 != "he" {
		t.Errorf("expected 'he' on line 0, got %q", line0)
	}
	
	line1, _ := buf.Line(1)
	if line1 != "llo" {
		t.Errorf("expected 'llo' on line 1, got %q", line1)
	}
	
	cursor := buf.Cursor()
	if cursor.Line != 1 || cursor.Col != 0 {
		t.Errorf("expected cursor at (1,0), got (%d,%d)", cursor.Line, cursor.Col)
	}
}

func TestDeleteLine(t *testing.T) {
	buf := New()
	buf.InsertChar('l')
	buf.InsertChar('i')
	buf.InsertChar('n')
	buf.InsertChar('e')
	buf.InsertChar('1')
	buf.InsertLine()
	buf.InsertChar('l')
	buf.InsertChar('i')
	buf.InsertChar('n')
	buf.InsertChar('e')
	buf.InsertChar('2')
	buf.InsertLine()
	buf.InsertChar('l')
	buf.InsertChar('i')
	buf.InsertChar('n')
	buf.InsertChar('e')
	buf.InsertChar('3')
	
	// Buffer: "line1\nline2\nline3"
	if buf.LineCount() != 3 {
		t.Errorf("expected 3 lines, got %d", buf.LineCount())
	}
	
	// Move to middle line
	buf.SetCursor(Position{Line: 1, Col: 0})
	
	err := buf.DeleteLine()
	if err != nil {
		t.Errorf("failed to delete line: %v", err)
	}
	
	if buf.LineCount() != 2 {
		t.Errorf("expected 2 lines after delete, got %d", buf.LineCount())
	}
	
	line0, _ := buf.Line(0)
	if line0 != "line1" {
		t.Errorf("expected 'line1' on line 0, got %q", line0)
	}
	
	line1, _ := buf.Line(1)
	if line1 != "line3" {
		t.Errorf("expected 'line3' on line 1, got %q", line1)
	}
}

func TestDeleteLastLine(t *testing.T) {
	buf := New()
	buf.InsertChar('t')
	buf.InsertChar('e')
	buf.InsertChar('s')
	buf.InsertChar('t')
	
	err := buf.DeleteLine()
	if err != nil {
		t.Errorf("failed to delete last line: %v", err)
	}
	
	if buf.LineCount() != 1 {
		t.Errorf("expected 1 line after deleting last line, got %d", buf.LineCount())
	}
	
	if buf.CurrentLine() != "" {
		t.Errorf("expected empty line after deleting last line, got %q", buf.CurrentLine())
	}
	
	cursor := buf.Cursor()
	if cursor.Line != 0 || cursor.Col != 0 {
		t.Errorf("expected cursor at (0,0), got (%d,%d)", cursor.Line, cursor.Col)
	}
}

func TestJoinLines(t *testing.T) {
	buf := New()
	buf.InsertChar('h')
	buf.InsertChar('e')
	buf.InsertChar('l')
	buf.InsertChar('l')
	buf.InsertChar('o')
	buf.InsertLine()
	buf.InsertChar('w')
	buf.InsertChar('o')
	buf.InsertChar('r')
	buf.InsertChar('l')
	buf.InsertChar('d')
	
	// Move to first line
	buf.SetCursor(Position{Line: 0, Col: 0})
	
	err := buf.JoinLines()
	if err != nil {
		t.Errorf("failed to join lines: %v", err)
	}
	
	if buf.LineCount() != 1 {
		t.Errorf("expected 1 line after join, got %d", buf.LineCount())
	}
	
	if buf.CurrentLine() != "hello world" {
		t.Errorf("expected 'hello world', got %q", buf.CurrentLine())
	}
}

func TestSetCursor(t *testing.T) {
	buf := New()
	buf.InsertChar('h')
	buf.InsertChar('e')
	buf.InsertChar('l')
	buf.InsertChar('l')
	buf.InsertChar('o')
	buf.InsertLine()
	buf.InsertChar('w')
	buf.InsertChar('o')
	buf.InsertChar('r')
	buf.InsertChar('l')
	buf.InsertChar('d')
	
	// Test valid position
	buf.SetCursor(Position{Line: 0, Col: 2})
	cursor := buf.Cursor()
	if cursor.Line != 0 || cursor.Col != 2 {
		t.Errorf("expected cursor at (0,2), got (%d,%d)", cursor.Line, cursor.Col)
	}
	
	// Test position beyond line end (should clamp)
	buf.SetCursor(Position{Line: 0, Col: 10})
	cursor = buf.Cursor()
	if cursor.Line != 0 || cursor.Col != 5 {
		t.Errorf("expected cursor clamped to (0,5), got (%d,%d)", cursor.Line, cursor.Col)
	}
	
	// Test negative position (should clamp)
	buf.SetCursor(Position{Line: -1, Col: -1})
	cursor = buf.Cursor()
	if cursor.Line != 0 || cursor.Col != 0 {
		t.Errorf("expected cursor clamped to (0,0), got (%d,%d)", cursor.Line, cursor.Col)
	}
	
	// Test line beyond buffer (should clamp)
	buf.SetCursor(Position{Line: 10, Col: 0})
	cursor = buf.Cursor()
	if cursor.Line != 1 || cursor.Col != 0 {
		t.Errorf("expected cursor clamped to (1,0), got (%d,%d)", cursor.Line, cursor.Col)
	}
}

func TestSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test_save.txt")
	
	// Create buffer with content
	buf := New()
	buf.InsertChar('h')
	buf.InsertChar('e')
	buf.InsertChar('l')
	buf.InsertChar('l')
	buf.InsertChar('o')
	buf.InsertLine()
	buf.InsertChar('w')
	buf.InsertChar('o')
	buf.InsertChar('r')
	buf.InsertChar('l')
	buf.InsertChar('d')
	
	// Save buffer
	err := buf.SaveAs(testFile)
	if err != nil {
		t.Errorf("failed to save buffer: %v", err)
	}
	
	if buf.Modified() {
		t.Error("buffer should not be modified after save")
	}
	
	if buf.Filename() != testFile {
		t.Errorf("expected filename %q, got %q", testFile, buf.Filename())
	}
	
	// Load buffer and verify content
	buf2, err := NewFromFile(testFile)
	if err != nil {
		t.Errorf("failed to load buffer: %v", err)
	}
	
	if buf2.String() != buf.String() {
		t.Errorf("loaded buffer content mismatch.\nExpected: %q\nGot: %q", buf.String(), buf2.String())
	}
}

func TestString(t *testing.T) {
	buf := New()
	buf.InsertChar('l')
	buf.InsertChar('i')
	buf.InsertChar('n')
	buf.InsertChar('e')
	buf.InsertChar('1')
	buf.InsertLine()
	buf.InsertChar('l')
	buf.InsertChar('i')
	buf.InsertChar('n')
	buf.InsertChar('e')
	buf.InsertChar('2')
	
	expected := "line1\nline2"
	if buf.String() != expected {
		t.Errorf("expected %q, got %q", expected, buf.String())
	}
}