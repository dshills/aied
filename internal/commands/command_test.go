package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dshills/aied/internal/buffer"
)

func TestCommandParser_ParseCommand(t *testing.T) {
	parser := NewCommandParser()

	tests := []struct {
		name        string
		input       string
		expectedCmd string
		expectedArgs []string
		expectError bool
	}{
		{"simple command", "w", "w", []string{}, false},
		{"command with args", "w filename.txt", "w", []string{"filename.txt"}, false},
		{"command with multiple args", "s/old/new/g", "s/old/new/g", []string{}, false},
		{"whitespace", "  q  ", "q", []string{}, false},
		{"empty command", "", "", nil, true},
		{"only whitespace", "   ", "", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, args, err := parser.ParseCommand(tt.input)

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if cmd != tt.expectedCmd {
				t.Errorf("expected command %q, got %q", tt.expectedCmd, cmd)
			}

			if len(args) != len(tt.expectedArgs) {
				t.Errorf("expected %d args, got %d", len(tt.expectedArgs), len(args))
				return
			}

			for i, arg := range args {
				if arg != tt.expectedArgs[i] {
					t.Errorf("expected arg %d to be %q, got %q", i, tt.expectedArgs[i], arg)
				}
			}
		})
	}
}

func TestCommandRegistry(t *testing.T) {
	registry := NewCommandRegistry()

	// Test that basic commands are registered
	expectedCommands := []string{"w", "q", "q!", "wq", "x", "e"}
	
	for _, cmdName := range expectedCommands {
		cmd, exists := registry.GetCommand(cmdName)
		if !exists {
			t.Errorf("expected command %q to be registered", cmdName)
		}
		if cmd == nil {
			t.Errorf("expected command %q to not be nil", cmdName)
		}
	}

	// Test that unknown command returns false
	_, exists := registry.GetCommand("unknown")
	if exists {
		t.Error("expected unknown command to not exist")
	}
}

func TestWriteCommand(t *testing.T) {
	cmd := NewWriteCommand()

	if cmd.Name() != "write" {
		t.Errorf("expected name 'write', got %q", cmd.Name())
	}

	aliases := cmd.Aliases()
	if len(aliases) != 1 || aliases[0] != "w" {
		t.Errorf("expected alias 'w', got %v", aliases)
	}

	// Test with temporary file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	buf := buffer.New()
	buf.InsertChar('h')
	buf.InsertChar('e')
	buf.InsertChar('l')
	buf.InsertChar('l')
	buf.InsertChar('o')

	// Test write with filename argument
	result := cmd.Execute([]string{testFile}, buf)
	if !result.Success {
		t.Errorf("expected write to succeed, got error: %s", result.Message)
	}

	// Verify file was created
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Error("expected file to be created")
	}

	// Test write without filename (should use buffer's filename)
	result = cmd.Execute([]string{}, buf)
	if !result.Success {
		t.Errorf("expected write to succeed, got error: %s", result.Message)
	}
}

func TestQuitCommand(t *testing.T) {
	cmd := NewQuitCommand()

	// Test quit with unmodified buffer
	buf := buffer.New()
	result := cmd.Execute([]string{}, buf)
	if !result.Success {
		t.Errorf("expected quit to succeed with clean buffer, got: %s", result.Message)
	}
	if !result.ExitEditor {
		t.Error("expected quit to exit editor")
	}

	// Test quit with modified buffer
	buf.InsertChar('a')
	result = cmd.Execute([]string{}, buf)
	if result.Success {
		t.Error("expected quit to fail with modified buffer")
	}
	if result.ExitEditor {
		t.Error("expected quit not to exit editor with modified buffer")
	}
}

func TestForceQuitCommand(t *testing.T) {
	cmd := NewForceQuitCommand()

	// Test force quit with modified buffer
	buf := buffer.New()
	buf.InsertChar('a')
	result := cmd.Execute([]string{}, buf)
	if !result.Success {
		t.Errorf("expected force quit to succeed, got: %s", result.Message)
	}
	if !result.ExitEditor {
		t.Error("expected force quit to exit editor")
	}
}

func TestWriteQuitCommand(t *testing.T) {
	cmd := NewWriteQuitCommand()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	buf := buffer.New()
	buf.InsertChar('t')
	buf.InsertChar('e')
	buf.InsertChar('s')
	buf.InsertChar('t')

	// Test write and quit
	result := cmd.Execute([]string{testFile}, buf)
	if !result.Success {
		t.Errorf("expected write-quit to succeed, got: %s", result.Message)
	}
	if !result.ExitEditor {
		t.Error("expected write-quit to exit editor")
	}

	// Verify file was created
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Error("expected file to be created")
	}
}

func TestCommandExecutor(t *testing.T) {
	executor := NewCommandExecutor()
	buf := buffer.New()

	// Test successful command
	result := executor.Execute("q", buf)
	if !result.Success {
		t.Errorf("expected quit command to succeed, got: %s", result.Message)
	}

	// Test unknown command
	result = executor.Execute("unknown", buf)
	if result.Success {
		t.Error("expected unknown command to fail")
	}

	// Test empty command
	result = executor.Execute("", buf)
	if result.Success {
		t.Error("expected empty command to fail")
	}
}