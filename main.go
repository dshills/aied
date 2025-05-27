package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/dshills/aied/internal/ai"
	"github.com/dshills/aied/internal/buffer"
	"github.com/dshills/aied/internal/commands"
	"github.com/dshills/aied/internal/config"
	"github.com/dshills/aied/internal/lsp"
	"github.com/dshills/aied/internal/modes"
	"github.com/dshills/aied/internal/ui"
	"go.lsp.dev/protocol"
)

// Global diagnostics cache - TODO: move to buffer manager when we support multiple buffers
var diagnosticsCache = make(map[string][]buffer.Diagnostic)

func main() {
	// Initialize AI system
	aiManager := initializeAI()
	commands.SetAIManager(aiManager)
	
	// Initialize LSP system
	lspManager := initializeLSP()
	if lspManager != nil {
		defer lspManager.StopAll()
		commands.SetLSPManager(lspManager)
	}

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
		
		// Open file in LSP if available
		if lspManager != nil && buf.Filename() != "" {
			content := lsp.GetBufferContent(buf)
			ctx := context.Background()
			if err := lspManager.OpenFile(ctx, buf.Filename(), content); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Failed to open file in LSP: %v\n", err)
			}
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

	// Create mode manager (starts in Normal mode)
	modeManager := modes.NewModeManager()
	
	// Set LSP manager if available
	if lspManager != nil {
		modeManager.SetLSPManager(lspManager)
	}

	// Initial render with mode
	modeText := modeManager.GetStatusText()
	terminalUI.RenderWithMode(buf, modeText)

	// Main event loop
	for terminalUI.IsRunning() {
		event := terminalUI.WaitForEvent()

		switch ev := event.(type) {
		case ui.KeyEvent:
			// Handle input through mode system
			result := modeManager.HandleInput(ev, buf)
			
			if result.ExitEditor {
				break // quit requested
			}
			
			// Handle unhandled events with fallback logic
			if !result.Handled {
				if handleFallbackKeyEvent(ev, buf, terminalUI) {
					break // quit requested
				}
			}
		case ui.ResizeEvent:
			terminalUI.HandleResize(ev)
		}

		// Update buffer diagnostics if available
		if buf.Filename() != "" {
			if diags, ok := diagnosticsCache[buf.Filename()]; ok {
				buf.SetDiagnostics(diags)
			}
		}
		
		// Re-render after any changes with current mode
		modeText := modeManager.GetStatusText()
		
		// Check if we're in command mode and need to show command line
		if commandLine, message, isCommandMode := modeManager.GetCommandInfo(); isCommandMode {
			terminalUI.RenderWithModeAndCommand(buf, modeText, commandLine, message)
		} else {
			terminalUI.RenderWithMode(buf, modeText)
		}
		
		// Show completion popup if in insert mode and completions are available
		if insertMode, ok := modeManager.CurrentMode().(*modes.InsertMode); ok {
			if completions, selectedIndex, showing := insertMode.GetCompletions(); showing {
				renderCompletionPopup(terminalUI, buf, completions, selectedIndex)
			}
		}
		
	}
}

// handleFallbackKeyEvent processes unhandled keyboard input and returns true if quit was requested
func handleFallbackKeyEvent(event ui.KeyEvent, buf *buffer.Buffer, terminalUI *ui.UI) bool {
	switch event.Action {
	case ui.KeyActionQuit, ui.KeyActionCtrlC:
		return true

	case ui.KeyActionCtrlS:
		// Global save command
		if buf.Filename() != "" {
			buf.Save()
		}

	default:
		// Ignore other keys - they should be handled by modes
	}

	return false
}

// initializeAI sets up the AI system with available providers
func initializeAI() *ai.AIManager {
	aiManager := ai.NewAIManager()
	
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to load config: %v\n", err)
		// Continue with defaults
		cfg = config.DefaultConfig()
	}
	
	// Configure providers from config file
	err = aiManager.ConfigureProviders(cfg.Providers)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to configure providers: %v\n", err)
	}
	
	// Set default provider if specified
	if cfg.AI.DefaultProvider != "" {
		if err := aiManager.SetActiveProvider(ai.ProviderType(cfg.AI.DefaultProvider)); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to set default provider: %v\n", err)
		}
	}
	
	return aiManager
}

// initializeLSP sets up the LSP system
func initializeLSP() *lsp.Manager {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to load config for LSP: %v\n", err)
		cfg = config.DefaultConfig()
	}
	
	// Check if LSP is enabled
	if !cfg.LSP.Enabled {
		return nil
	}
	
	// Get working directory
	workDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to get working directory: %v\n", err)
		workDir = "."
	}
	workDir, _ = filepath.Abs(workDir)
	
	// Create LSP manager
	lspManager := lsp.NewManager(workDir)
	
	// Convert config to LSP server configs
	var serverConfigs []lsp.ServerConfig
	for _, srv := range cfg.LSP.Servers {
		if srv.Enabled {
			serverConfigs = append(serverConfigs, lsp.ServerConfig{
				Name:       srv.Name,
				Command:    srv.Command,
				Args:       srv.Args,
				Languages:  srv.Languages,
				Extensions: srv.Extensions,
			})
		}
	}
	
	// Configure servers
	lspManager.Configure(serverConfigs)
	
	// Auto-start servers if configured
	if cfg.LSP.AutoStart {
		ctx := context.Background()
		for _, srv := range serverConfigs {
			if err := lspManager.Start(ctx, srv.Name); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Failed to start LSP server %s: %v\n", srv.Name, err)
			} else {
				fmt.Printf("Started LSP server: %s\n", srv.Name)
			}
		}
	}
	
	// Set up diagnostics handler
	if cfg.LSP.ShowDiagnostics {
		lspManager.SetDiagnosticsHandler(func(filename string, diagnostics []protocol.Diagnostic) {
			// Convert LSP diagnostics to buffer diagnostics
			var bufDiags []buffer.Diagnostic
			for _, diag := range diagnostics {
				bufDiags = append(bufDiags, buffer.Diagnostic{
					Line:     int(diag.Range.Start.Line),
					Column:   int(diag.Range.Start.Character),
					Severity: int(diag.Severity),
					Message:  diag.Message,
					Source:   diag.Source,
				})
			}
			
			// Store diagnostics globally for now
			diagnosticsCache[filename] = bufDiags
		})
	}
	
	return lspManager
}

// renderCompletionPopup renders a simple completion popup
func renderCompletionPopup(terminalUI *ui.UI, buf *buffer.Buffer, completions []modes.CompletionItem, selectedIndex int) {
	if len(completions) == 0 {
		return
	}
	
	screen := terminalUI.GetScreen()
	cursor := buf.Cursor()
	viewport := terminalUI.GetViewport()
	
	// Calculate popup position (below cursor)
	popupX := cursor.Col - viewport.StartCol + 1
	popupY := cursor.Line - viewport.StartLine + 1
	
	// Popup dimensions
	maxWidth := 40
	maxHeight := 10
	popupHeight := len(completions)
	if popupHeight > maxHeight {
		popupHeight = maxHeight
	}
	
	// Calculate actual width needed
	popupWidth := 0
	for i, item := range completions {
		if i >= maxHeight {
			break
		}
		itemWidth := len(item.Label) + len(item.Kind) + 3 // 3 for " - "
		if itemWidth > popupWidth {
			popupWidth = itemWidth
		}
	}
	if popupWidth > maxWidth {
		popupWidth = maxWidth
	}
	if popupWidth < 20 {
		popupWidth = 20
	}
	
	// Adjust position if popup would go off screen
	screenWidth, screenHeight := screen.Size()
	if popupX + popupWidth >= screenWidth {
		popupX = screenWidth - popupWidth - 1
	}
	if popupY + popupHeight >= screenHeight - 1 { // -1 for status line
		popupY = cursor.Line - viewport.StartLine - popupHeight
		if popupY < 0 {
			popupY = 0
		}
	}
	
	// Render popup background and border
	normalStyle := ui.GetStyle("normal")
	selectedStyle := ui.GetStyle("selected")
	borderStyle := ui.GetStyle("border")
	
	// Draw border
	for y := 0; y < popupHeight + 2; y++ {
		for x := 0; x < popupWidth + 2; x++ {
			ch := ' '
			style := borderStyle
			
			if y == 0 || y == popupHeight + 1 {
				// Top and bottom border
				if x == 0 && y == 0 {
					ch = '┌'
				} else if x == popupWidth + 1 && y == 0 {
					ch = '┐'
				} else if x == 0 && y == popupHeight + 1 {
					ch = '└'
				} else if x == popupWidth + 1 && y == popupHeight + 1 {
					ch = '┘'
				} else {
					ch = '─'
				}
			} else if x == 0 || x == popupWidth + 1 {
				// Side borders
				ch = '│'
			} else {
				// Interior
				style = normalStyle
			}
			
			screen.SetCell(popupX + x, popupY + y, ch, style)
		}
	}
	
	// Draw completion items
	for i, item := range completions {
		if i >= popupHeight {
			break
		}
		
		style := normalStyle
		if i == selectedIndex {
			style = selectedStyle
		}
		
		// Format item text
		text := item.Label
		if item.Kind != "" {
			text = item.Kind + ": " + text
		}
		
		// Truncate if too long
		if len(text) > popupWidth {
			text = text[:popupWidth-3] + "..."
		}
		
		// Clear the line
		for x := 1; x < popupWidth + 1; x++ {
			screen.SetCell(popupX + x, popupY + i + 1, ' ', style)
		}
		
		// Draw the text
		screen.SetText(popupX + 2, popupY + i + 1, text, style)
	}
	
	screen.Show()
}

// updateLSPBuffer sends buffer changes to LSP server
func updateLSPBuffer(lspManager *lsp.Manager, buf *buffer.Buffer) {
	if buf.Filename() == "" {
		return
	}
	
	content := buf.String()
	ctx := context.Background()
	
	// Update the file content in LSP
	err := lspManager.UpdateFile(ctx, buf.Filename(), content, 1)
	if err != nil {
		// Silently ignore LSP update errors for now
		return
	}
}
