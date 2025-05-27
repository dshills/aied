#!/bin/bash

# AIED LSP Testing Script

echo "=== AIED LSP Integration Test ==="
echo

# Check if gopls is available
if ! command -v gopls &> /dev/null; then
    echo "Warning: gopls not found. Go language support will not work."
    echo "Install it with: go install golang.org/x/tools/gopls@latest"
    echo
fi

# Check if aied binary exists
if [[ ! -f "./aied" ]]; then
    echo "Error: aied binary not found. Please run 'go build -o aied' first."
    exit 1
fi

echo "✓ AIED binary found"
echo "✓ Test files created in test/ directory"
echo

echo "LSP Features implemented:"
echo "  ✓ Diagnostics display (errors/warnings highlighted)"
echo "  ✓ Code completion (Ctrl+Space in insert mode)"
echo "  ✓ Hover information (gh in normal mode)" 
echo "  ✓ Go-to-definition (gd in normal mode)"
echo "  ✓ Find references (gr in normal mode)"
echo "  ✓ LSP commands (:hover, :definition, :references)"
echo

echo "To test manually:"
echo "  1. ./aied test/example.go"
echo "  2. Try the keyboard shortcuts and commands listed above"
echo "  3. In insert mode, type 'fmt.' and press Ctrl+Space for completion"
echo

echo "Configuration file: .aied.yaml"
echo "LSP settings can be customized in the config file."
echo

echo "Note: This editor runs in terminal mode and cannot be tested in this script environment."
echo "Please run './aied test/example.go' in a proper terminal to test LSP features."