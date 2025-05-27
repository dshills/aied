package main

import (
	"fmt"
	"os"

	"github.com/dshills/aied/internal/ai"
	"github.com/dshills/aied/internal/buffer"
	"github.com/dshills/aied/internal/commands"
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

	// Initialize providers based on environment variables
	// OpenAI
	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
		provider := ai.NewOpenAIProvider()
		provider.Configure(ai.ProviderConfig{
			APIKey: apiKey,
			Model:  "gpt-4",
		})
		aiManager.RegisterProvider(provider)
	}

	// Anthropic
	if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
		provider := ai.NewAnthropicProvider()
		provider.Configure(ai.ProviderConfig{
			APIKey: apiKey,
			Model:  "claude-3-5-sonnet-20241022",
		})
		aiManager.RegisterProvider(provider)
	}

	// Google
	if apiKey := os.Getenv("GOOGLE_API_KEY"); apiKey != "" {
		provider := ai.NewGoogleProvider()
		provider.Configure(ai.ProviderConfig{
			APIKey: apiKey,
			Model:  "gemini-1.5-flash",
		})
		aiManager.RegisterProvider(provider)
	}

	// Ollama (always try to register - it will check if available)
	ollamaProvider := ai.NewOllamaProvider()
	ollamaProvider.Configure(ai.ProviderConfig{
		Model: "llama2",
	})
	aiManager.RegisterProvider(ollamaProvider)

	return aiManager
}
