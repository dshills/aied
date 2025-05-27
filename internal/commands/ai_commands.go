package commands

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/dshills/aied/internal/ai"
	"github.com/dshills/aied/internal/buffer"
)

// Global AI manager - will be initialized from main
var aiManager *ai.AIManager

// SetAIManager sets the global AI manager for commands
func SetAIManager(manager *ai.AIManager) {
	aiManager = manager
}

// AICompleteCommand implements AI-powered code completion
type AICompleteCommand struct{}

func NewAICompleteCommand() *AICompleteCommand {
	return &AICompleteCommand{}
}

func (c *AICompleteCommand) Name() string {
	return "aicomplete"
}

func (c *AICompleteCommand) Aliases() []string {
	return []string{"aic"}
}

func (c *AICompleteCommand) Execute(args []string, buf *buffer.Buffer) CommandResult {
	if aiManager == nil {
		return CommandResult{
			Success:    false,
			Message:    "AI manager not initialized",
			SwitchMode: true,
		}
	}

	// Get current line or selection
	cursor := buf.Cursor()
	currentLine := buf.CurrentLine()
	
	// Get context (surrounding lines)
	contextLines := 10
	startLine := cursor.Line - contextLines
	if startLine < 0 {
		startLine = 0
	}
	endLine := cursor.Line + contextLines
	if endLine >= buf.LineCount() {
		endLine = buf.LineCount() - 1
	}
	
	var contextBuilder strings.Builder
	for i := startLine; i <= endLine; i++ {
		if i == cursor.Line {
			contextBuilder.WriteString(">>> ")
		}
		lineContent, _ := buf.Line(i)
		contextBuilder.WriteString(lineContent)
		contextBuilder.WriteString("\n")
	}

	// Create AI request
	req := ai.AIRequest{
		Prompt:   currentLine[:cursor.Col], // Everything before cursor
		Context:  contextBuilder.String(),
		Language: detectLanguage(buf.Filename()),
		Type:     ai.RequestCompletion,
	}

	// Make AI request with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := aiManager.Request(ctx, req)
	if err != nil {
		return CommandResult{
			Success:    false,
			Message:    fmt.Sprintf("AI error: %s", err.Error()),
			SwitchMode: true,
		}
	}

	// Insert completion at cursor
	buf.InsertTextAt(cursor.Line, cursor.Col, resp.Content)

	return CommandResult{
		Success:    true,
		Message:    fmt.Sprintf("Completed with %s", resp.Provider),
		SwitchMode: true,
	}
}

func (c *AICompleteCommand) Help() string {
	return "Complete code at cursor position using AI"
}

// AIExplainCommand explains selected code or current line
type AIExplainCommand struct{}

func NewAIExplainCommand() *AIExplainCommand {
	return &AIExplainCommand{}
}

func (c *AIExplainCommand) Name() string {
	return "aiexplain"
}

func (c *AIExplainCommand) Aliases() []string {
	return []string{"aie"}
}

func (c *AIExplainCommand) Execute(args []string, buf *buffer.Buffer) CommandResult {
	if aiManager == nil {
		return CommandResult{
			Success:    false,
			Message:    "AI manager not initialized",
			SwitchMode: true,
		}
	}

	// Get current line or selection (for now just current line)
	codeToExplain := buf.CurrentLine()

	if len(args) > 0 {
		// If args provided, use them as the code to explain
		codeToExplain = strings.Join(args, " ")
	}

	// Create AI request
	req := ai.AIRequest{
		Prompt:   fmt.Sprintf("Explain this code: %s", codeToExplain),
		Language: detectLanguage(buf.Filename()),
		Type:     ai.RequestExplanation,
	}

	// Make AI request
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := aiManager.Request(ctx, req)
	if err != nil {
		return CommandResult{
			Success:    false,
			Message:    fmt.Sprintf("AI error: %s", err.Error()),
			SwitchMode: true,
		}
	}

	// For now, just show the message - later we'll add a popup or split view
	return CommandResult{
		Success:    true,
		Message:    resp.Content,
		SwitchMode: true,
	}
}

func (c *AIExplainCommand) Help() string {
	return "Explain code at cursor or provided as argument"
}

// AIRefactorCommand suggests refactoring for selected code
type AIRefactorCommand struct{}

func NewAIRefactorCommand() *AIRefactorCommand {
	return &AIRefactorCommand{}
}

func (c *AIRefactorCommand) Name() string {
	return "airefactor"
}

func (c *AIRefactorCommand) Aliases() []string {
	return []string{"air"}
}

func (c *AIRefactorCommand) Execute(args []string, buf *buffer.Buffer) CommandResult {
	if aiManager == nil {
		return CommandResult{
			Success:    false,
			Message:    "AI manager not initialized",
			SwitchMode: true,
		}
	}

	// Get current function or block (for now just current line)
	codeToRefactor := buf.CurrentLine()

	// Create AI request
	req := ai.AIRequest{
		Prompt:   codeToRefactor,
		Language: detectLanguage(buf.Filename()),
		Type:     ai.RequestRefactor,
	}

	// Make AI request
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := aiManager.Request(ctx, req)
	if err != nil {
		return CommandResult{
			Success:    false,
			Message:    fmt.Sprintf("AI error: %s", err.Error()),
			SwitchMode: true,
		}
	}

	// For now, show suggestion - later we'll add preview and apply
	return CommandResult{
		Success:    true,
		Message:    fmt.Sprintf("Refactor suggestion: %s", resp.Content),
		SwitchMode: true,
	}
}

func (c *AIRefactorCommand) Help() string {
	return "Get AI suggestions for refactoring code"
}

// AIChatCommand opens AI chat for general help
type AIChatCommand struct{}

func NewAIChatCommand() *AIChatCommand {
	return &AIChatCommand{}
}

func (c *AIChatCommand) Name() string {
	return "ai"
}

func (c *AIChatCommand) Aliases() []string {
	return []string{}
}

func (c *AIChatCommand) Execute(args []string, buf *buffer.Buffer) CommandResult {
	if aiManager == nil {
		return CommandResult{
			Success:    false,
			Message:    "AI manager not initialized",
			SwitchMode: true,
		}
	}

	if len(args) == 0 {
		return CommandResult{
			Success:    false,
			Message:    "Usage: :ai <question>",
			SwitchMode: true,
		}
	}

	question := strings.Join(args, " ")

	// Get current file context
	var contextBuilder strings.Builder
	contextBuilder.WriteString(fmt.Sprintf("File: %s\n", buf.Filename()))
	contextBuilder.WriteString(fmt.Sprintf("Language: %s\n", detectLanguage(buf.Filename())))
	
	// Add current line info
	cursor := buf.Cursor()
	contextBuilder.WriteString(fmt.Sprintf("Current line %d: %s\n", cursor.Line+1, buf.CurrentLine()))
	contextBuilder.WriteString(fmt.Sprintf("Cursor at column %d\n", cursor.Col))

	// Create AI request
	req := ai.AIRequest{
		Prompt:   question,
		Context:  contextBuilder.String(),
		Type:     ai.RequestChat,
	}

	// Make AI request
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := aiManager.Request(ctx, req)
	if err != nil {
		return CommandResult{
			Success:    false,
			Message:    fmt.Sprintf("AI error: %s", err.Error()),
			SwitchMode: true,
		}
	}

	return CommandResult{
		Success:    true,
		Message:    resp.Content,
		SwitchMode: true,
	}
}

func (c *AIChatCommand) Help() string {
	return "Ask AI a general question about your code"
}

// AIProviderCommand manages AI providers
type AIProviderCommand struct{}

func NewAIProviderCommand() *AIProviderCommand {
	return &AIProviderCommand{}
}

func (c *AIProviderCommand) Name() string {
	return "aiprovider"
}

func (c *AIProviderCommand) Aliases() []string {
	return []string{"aip"}
}

func (c *AIProviderCommand) Execute(args []string, buf *buffer.Buffer) CommandResult {
	if aiManager == nil {
		return CommandResult{
			Success:    false,
			Message:    "AI manager not initialized",
			SwitchMode: true,
		}
	}

	if len(args) == 0 {
		// List available providers
		providers := aiManager.ListAvailableProviders()
		active := aiManager.GetActiveProvider()
		
		var providerList []string
		for pType, provider := range providers {
			marker := ""
			if active != nil && provider.Name() == active.Name() {
				marker = " (active)"
			}
			providerList = append(providerList, fmt.Sprintf("%s%s", pType, marker))
		}
		
		if len(providerList) == 0 {
			return CommandResult{
				Success:    true,
				Message:    "No AI providers available",
				SwitchMode: true,
			}
		}
		
		return CommandResult{
			Success:    true,
			Message:    fmt.Sprintf("Available providers: %s", strings.Join(providerList, ", ")),
			SwitchMode: true,
		}
	}

	// Set active provider
	providerType := ai.ProviderType(args[0])
	err := aiManager.SetActiveProvider(providerType)
	if err != nil {
		return CommandResult{
			Success:    false,
			Message:    fmt.Sprintf("Failed to set provider: %s", err.Error()),
			SwitchMode: true,
		}
	}

	return CommandResult{
		Success:    true,
		Message:    fmt.Sprintf("Active provider set to: %s", providerType),
		SwitchMode: true,
	}
}

func (c *AIProviderCommand) Help() string {
	return "List or set active AI provider"
}

// Helper function to detect programming language from filename
func detectLanguage(filename string) string {
	parts := strings.Split(filename, ".")
	if len(parts) < 2 {
		return "text"
	}
	
	ext := parts[len(parts)-1]
	switch ext {
	case "go":
		return "go"
	case "py":
		return "python"
	case "js", "jsx":
		return "javascript"
	case "ts", "tsx":
		return "typescript"
	case "java":
		return "java"
	case "c":
		return "c"
	case "cpp", "cc", "cxx":
		return "cpp"
	case "rs":
		return "rust"
	case "rb":
		return "ruby"
	case "php":
		return "php"
	case "sh", "bash":
		return "bash"
	case "md":
		return "markdown"
	default:
		return "text"
	}
}