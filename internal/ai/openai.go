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

// OpenAIProvider implements the Provider interface for OpenAI
type OpenAIProvider struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

// NewOpenAIProvider creates a new OpenAI provider
func NewOpenAIProvider() *OpenAIProvider {
	return &OpenAIProvider{
		baseURL: "https://api.openai.com/v1",
		model:   "gpt-4", // Default model
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Name returns the provider name
func (o *OpenAIProvider) Name() ProviderType {
	return ProviderOpenAI
}

// IsAvailable checks if the provider is configured and available
func (o *OpenAIProvider) IsAvailable() bool {
	return o.apiKey != ""
}

// Configure sets up the provider with configuration
func (o *OpenAIProvider) Configure(config ProviderConfig) error {
	o.apiKey = config.APIKey
	
	if config.BaseURL != "" {
		o.baseURL = config.BaseURL
	}
	
	if config.Model != "" {
		o.model = config.Model
	}
	
	// Configure client timeout from options
	if config.Options != nil {
		if timeout, ok := config.Options["timeout"].(int); ok {
			o.client.Timeout = time.Duration(timeout) * time.Second
		}
	}
	
	return nil
}

// OpenAI API structures
type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIChatRequest struct {
	Model       string          `json:"model"`
	Messages    []openAIMessage `json:"messages"`
	MaxTokens   int             `json:"max_tokens,omitempty"`
	Temperature float64         `json:"temperature,omitempty"`
	Stream      bool            `json:"stream"`
}

type openAIChatResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// Complete generates code completion suggestions
func (o *OpenAIProvider) Complete(ctx context.Context, req AIRequest) (*AIResponse, error) {
	prompt := o.buildCompletionPrompt(req)
	
	chatReq := openAIChatRequest{
		Model: o.model,
		Messages: []openAIMessage{
			{Role: "system", Content: "You are a helpful code completion assistant. Provide concise, accurate code completions."},
			{Role: "user", Content: prompt},
		},
		MaxTokens:   150,
		Temperature: 0.1, // Low temperature for more deterministic completions
		Stream:      false,
	}
	
	response, err := o.makeRequest(ctx, chatReq)
	if err != nil {
		return nil, err
	}
	
	return &AIResponse{
		Content:    response.Choices[0].Message.Content,
		Confidence: 0.8, // Default confidence
		Provider:   string(ProviderOpenAI),
		Model:      o.model,
	}, nil
}

// Chat handles general conversational requests
func (o *OpenAIProvider) Chat(ctx context.Context, req AIRequest) (*AIResponse, error) {
	prompt := o.buildChatPrompt(req)
	
	chatReq := openAIChatRequest{
		Model: o.model,
		Messages: []openAIMessage{
			{Role: "system", Content: o.getSystemPrompt(req.Type)},
			{Role: "user", Content: prompt},
		},
		MaxTokens:   1000,
		Temperature: 0.3,
		Stream:      false,
	}
	
	response, err := o.makeRequest(ctx, chatReq)
	if err != nil {
		return nil, err
	}
	
	return &AIResponse{
		Content:    response.Choices[0].Message.Content,
		Confidence: 0.9,
		Provider:   string(ProviderOpenAI),
		Model:      o.model,
	}, nil
}

// Analyze provides code analysis and suggestions
func (o *OpenAIProvider) Analyze(ctx context.Context, req AIRequest) (*AIResponse, error) {
	prompt := o.buildAnalysisPrompt(req)
	
	chatReq := openAIChatRequest{
		Model: o.model,
		Messages: []openAIMessage{
			{Role: "system", Content: "You are an expert code reviewer and refactoring assistant. Provide specific, actionable suggestions."},
			{Role: "user", Content: prompt},
		},
		MaxTokens:   800,
		Temperature: 0.2,
		Stream:      false,
	}
	
	response, err := o.makeRequest(ctx, chatReq)
	if err != nil {
		return nil, err
	}
	
	return &AIResponse{
		Content:    response.Choices[0].Message.Content,
		Confidence: 0.85,
		Provider:   string(ProviderOpenAI),
		Model:      o.model,
	}, nil
}

// makeRequest makes an HTTP request to OpenAI API
func (o *OpenAIProvider) makeRequest(ctx context.Context, req openAIChatRequest) (*openAIChatResponse, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	
	httpReq, err := http.NewRequestWithContext(ctx, "POST", o.baseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+o.apiKey)
	
	resp, err := o.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	
	var response openAIChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("no response choices returned")
	}
	
	return &response, nil
}

// buildCompletionPrompt builds a prompt for code completion
func (o *OpenAIProvider) buildCompletionPrompt(req AIRequest) string {
	var prompt strings.Builder
	
	if req.Language != "" {
		prompt.WriteString(fmt.Sprintf("Complete this %s code:\n\n", req.Language))
	} else {
		prompt.WriteString("Complete this code:\n\n")
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
func (o *OpenAIProvider) buildChatPrompt(req AIRequest) string {
	var prompt strings.Builder
	
	if req.Context != "" {
		prompt.WriteString("Context:\n")
		prompt.WriteString(req.Context)
		prompt.WriteString("\n\n")
	}
	
	prompt.WriteString("Question: ")
	prompt.WriteString(req.Prompt)
	
	return prompt.String()
}

// buildAnalysisPrompt builds a prompt for code analysis
func (o *OpenAIProvider) buildAnalysisPrompt(req AIRequest) string {
	var prompt strings.Builder
	
	if req.Language != "" {
		prompt.WriteString(fmt.Sprintf("Analyze this %s code and suggest improvements:\n\n", req.Language))
	} else {
		prompt.WriteString("Analyze this code and suggest improvements:\n\n")
	}
	
	prompt.WriteString("Code:\n")
	prompt.WriteString(req.Context)
	
	if req.Prompt != "" {
		prompt.WriteString("\n\nSpecific focus: ")
		prompt.WriteString(req.Prompt)
	}
	
	return prompt.String()
}

// getSystemPrompt returns appropriate system prompt for request type
func (o *OpenAIProvider) getSystemPrompt(reqType RequestType) string {
	switch reqType {
	case RequestExplanation:
		return "You are a helpful programming tutor. Explain code concepts clearly and concisely."
	case RequestDebug:
		return "You are an expert debugger. Help identify and fix issues in code."
	case RequestDocumentation:
		return "You are a technical writer. Generate clear, comprehensive documentation."
	default:
		return "You are a helpful AI programming assistant."
	}
}