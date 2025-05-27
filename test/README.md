# AIED LSP Testing

This directory contains test files for LSP functionality in AIED.

## Test Files

### example.go
A Go file with various language constructs for testing:
- Code completion (type `.` after objects)
- Go-to-definition (use `gd` on function calls)
- Hover information (use `gh` on identifiers)
- Error diagnostics (intentional syntax errors)

## How to Test

1. **Code Completion**: In insert mode, type `.` after an object (like `fmt.` or `strings.`) and press Ctrl+Space
2. **Go-to-Definition**: In normal mode, place cursor on a function name and press `gd`
3. **Hover Information**: In normal mode, place cursor on an identifier and press `gh`
4. **Find References**: In normal mode, place cursor on a function name and press `gr`
5. **Diagnostics**: Syntax errors should be highlighted with red underlines

## Commands

- `:hover` - Show hover information at cursor
- `:definition` - Go to definition of symbol at cursor
- `:references` - Find all references to symbol at cursor
- `:rename <new-name>` - Rename symbol at cursor

## Keyboard Shortcuts

- `gh` - Hover information
- `gd` - Go to definition  
- `gr` - Find references
- `Ctrl+Space` - Trigger completion (in insert mode)

## Prerequisites

For Go language support, ensure `gopls` is installed:
```bash
go install golang.org/x/tools/gopls@latest
```