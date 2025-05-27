package main

import (
	"fmt"
	"os"

	"github.com/dshills/aied/internal/ai"
	"github.com/dshills/aied/internal/buffer"
	"github.com/dshills/aied/internal/commands"
	"github.com/dshills/aied/internal/config"
	"github.com/dshills/aied/internal/modes"
	"github.com/dshills/aied/internal/ui"
)

func main() {
	// Initialize AI system
	aiManager := initializeAI()
	commands.SetAIManager(aiManager)

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

	// Create mode manager (starts in Normal mode)
	modeManager := modes.NewModeManager()

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

		// Re-render after any changes with current mode
		modeText := modeManager.GetStatusText()
		
		// Check if we're in command mode and need to show command line
		if commandLine, message, isCommandMode := modeManager.GetCommandInfo(); isCommandMode {
			terminalUI.RenderWithModeAndCommand(buf, modeText, commandLine, message)
		} else {
			terminalUI.RenderWithMode(buf, modeText)
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
