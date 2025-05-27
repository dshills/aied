package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// AnthropicProvider implements the Provider interface for Anthropic Claude
type AnthropicProvider struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

// NewAnthropicProvider creates a new Anthropic provider
func NewAnthropicProvider() *AnthropicProvider {
	return &AnthropicProvider{
		baseURL: "https://api.anthropic.com/v1",
		model:   "claude-3-5-sonnet-20241022", // Default model
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Name returns the provider name
func (a *AnthropicProvider) Name() ProviderType {
	return ProviderAnthropic
}

// IsAvailable checks if the provider is configured and available
func (a *AnthropicProvider) IsAvailable() bool {
	return a.apiKey != ""
}

// Configure sets up the provider with configuration
func (a *AnthropicProvider) Configure(config ProviderConfig) error {
	a.apiKey = config.APIKey
	
	if config.BaseURL != "" {
		a.baseURL = config.BaseURL
	}
	
	if config.Model != "" {
		a.model = config.Model
	}
	
	// Configure client timeout from options
	if config.Options != nil {
		if timeout, ok := config.Options["timeout"].(int); ok {
			a.client.Timeout = time.Duration(timeout) * time.Second
		}
	}
	
	return nil
}

// Anthropic API structures
type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	Messages  []anthropicMessage `json:"messages"`
	System    string             `json:"system,omitempty"`
}

type anthropicResponse struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Role    string `json:"role"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Model        string `json:"model"`
	StopReason   string `json:"stop_reason"`
	StopSequence string `json:"stop_sequence"`
	Usage        struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

// Complete generates code completion suggestions
func (a *AnthropicProvider) Complete(ctx context.Context, req AIRequest) (*AIResponse, error) {
	prompt := a.buildCompletionPrompt(req)
	
	anthropicReq := anthropicRequest{
		Model:     a.model,
		MaxTokens: 150,
		System:    "You are a helpful code completion assistant. Provide concise, accurate code completions without explanations.",
		Messages: []anthropicMessage{
			{Role: "user", Content: prompt},
		},
	}
	
	response, err := a.makeRequest(ctx, anthropicReq)
	if err != nil {
		return nil, err
	}
	
	content := ""
	if len(response.Content) > 0 {
		content = response.Content[0].Text
	}
	
	return &AIResponse{
		Content:    content,
		Confidence: 0.85, // Claude is generally very reliable
		Provider:   string(ProviderAnthropic),
		Model:      a.model,
	}, nil
}

// Chat handles general conversational requests
func (a *AnthropicProvider) Chat(ctx context.Context, req AIRequest) (*AIResponse, error) {
	prompt := a.buildChatPrompt(req)
	
	anthropicReq := anthropicRequest{
		Model:     a.model,
		MaxTokens: 1000,
		System:    a.getSystemPrompt(req.Type),
		Messages: []anthropicMessage{
			{Role: "user", Content: prompt},
		},
	}
	
	response, err := a.makeRequest(ctx, anthropicReq)
	if err != nil {
		return nil, err
	}
	
	content := ""
	if len(response.Content) > 0 {
		content = response.Content[0].Text
	}
	
	return &AIResponse{
		Content:    content,
		Confidence: 0.9, // Claude excels at conversational tasks
		Provider:   string(ProviderAnthropic),
		Model:      a.model,
	}, nil
}

// Analyze provides code analysis and suggestions
func (a *AnthropicProvider) Analyze(ctx context.Context, req AIRequest) (*AIResponse, error) {
	prompt := a.buildAnalysisPrompt(req)
	
	anthropicReq := anthropicRequest{
		Model:     a.model,
		MaxTokens: 800,
		System:    "You are an expert code reviewer and refactoring assistant. Provide specific, actionable suggestions with clear explanations.",
		Messages: []anthropicMessage{
			{Role: "user", Content: prompt},
		},
	}
	
	response, err := a.makeRequest(ctx, anthropicReq)
	if err != nil {
		return nil, err
	}
	
	content := ""
	if len(response.Content) > 0 {
		content = response.Content[0].Text
	}
	
	return &AIResponse{
		Content:    content,
		Confidence: 0.9, // Claude is excellent at code analysis
		Provider:   string(ProviderAnthropic),
		Model:      a.model,
	}, nil
}

// makeRequest makes an HTTP request to Anthropic API
func (a *AnthropicProvider) makeRequest(ctx context.Context, req anthropicRequest) (*anthropicResponse, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	
	httpReq, err := http.NewRequestWithContext(ctx, "POST", a.baseURL+"/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", a.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")
	
	resp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	
	var response anthropicResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	return &response, nil
}

// buildCompletionPrompt builds a prompt for code completion
func (a *AnthropicProvider) buildCompletionPrompt(req AIRequest) string {
	var prompt strings.Builder
	
	if req.Language != "" {
		prompt.WriteString(fmt.Sprintf("Complete this %s code. Only provide the completion, no explanations:\n\n", req.Language))
	} else {
		prompt.WriteString("Complete this code. Only provide the completion, no explanations:\n\n")
	}
	
	if req.Context != "" {
		prompt.WriteString("Context:\n")
		prompt.WriteString(req.Context)
		prompt.WriteString("\n\n")
	}
	
	prompt.WriteString("Code to complete:\n")
	prompt.WriteString(req.Prompt)
	
	return prompt.String()
}

// buildChatPrompt builds a prompt for chat/explanation requests
func (a *AnthropicProvider) buildChatPrompt(req AIRequest) string {
	var prompt strings.Builder
	
	if req.Context != "" {
		prompt.WriteString("Context:\n")
		prompt.WriteString(req.Context)
		prompt.WriteString("\n\n")
	}
	
	prompt.WriteString(req.Prompt)
	
	return prompt.String()
}

// buildAnalysisPrompt builds a prompt for code analysis
func (a *AnthropicProvider) buildAnalysisPrompt(req AIRequest) string {
	var prompt strings.Builder
	
	if req.Language != "" {
		prompt.WriteString(fmt.Sprintf("Analyze this %s code and suggest specific improvements:\n\n", req.Language))
	} else {
		prompt.WriteString("Analyze this code and suggest specific improvements:\n\n")
	}
	
	prompt.WriteString("Code:\n")
	prompt.WriteString(req.Context)
	
	if req.Prompt != "" {
		prompt.WriteString("\n\nSpecific focus: ")
		prompt.WriteString(req.Prompt)
	}
	
	prompt.WriteString("\n\nProvide specific, actionable suggestions with explanations.")
	
	return prompt.String()
}

// getSystemPrompt returns appropriate system prompt for request type
func (a *AnthropicProvider) getSystemPrompt(reqType RequestType) string {
	switch reqType {
	case RequestExplanation:
		return "You are an expert programming educator. Explain code concepts clearly with examples when helpful."
	case RequestDebug:
		return "You are an expert debugger. Analyze code carefully to identify issues and provide clear solutions."
	case RequestDocumentation:
		return "You are a technical documentation expert. Generate clear, comprehensive documentation that follows best practices."
	case RequestRefactor:
		return "You are a senior software engineer specializing in code refactoring. Suggest improvements that enhance readability, maintainability, and performance."
	default:
		return "You are a helpful AI programming assistant with expertise across multiple programming languages and best practices."
	}
}