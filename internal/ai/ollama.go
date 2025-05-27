package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type OllamaProvider struct {
	baseURL string
	model   string
	client  *http.Client
}

type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type OllamaChatRequest struct {
	Model    string           `json:"model"`
	Messages []OllamaMessage  `json:"messages"`
	Stream   bool             `json:"stream"`
}

type OllamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OllamaResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
	Error    string `json:"error,omitempty"`
}

type OllamaChatResponse struct {
	Message OllamaMessage `json:"message"`
	Done    bool          `json:"done"`
	Error   string        `json:"error,omitempty"`
}

type OllamaTagsResponse struct {
	Models []OllamaModel `json:"models"`
}

type OllamaModel struct {
	Name string `json:"name"`
}

func NewOllamaProvider() *OllamaProvider {
	return &OllamaProvider{
		baseURL: "http://localhost:11434",
		model:   "llama2",
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (p *OllamaProvider) Name() ProviderType {
	return ProviderOllama
}

func (p *OllamaProvider) IsAvailable() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	req, err := http.NewRequestWithContext(ctx, "GET", p.baseURL+"/api/tags", nil)
	if err != nil {
		return false
	}
	
	resp, err := p.client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	
	return resp.StatusCode == http.StatusOK
}

func (p *OllamaProvider) Configure(config ProviderConfig) error {
	if config.BaseURL != "" {
		p.baseURL = config.BaseURL
	}
	
	if config.Model != "" {
		p.model = config.Model
	}
	
	return nil
}

func (p *OllamaProvider) Complete(ctx context.Context, req AIRequest) (*AIResponse, error) {
	prompt := p.buildPrompt(req, "completion")
	
	ollamaReq := OllamaRequest{
		Model:  p.model,
		Prompt: prompt,
		Stream: false,
	}

	jsonData, err := json.Marshal(ollamaReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/api/generate", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	var ollamaResp OllamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	if ollamaResp.Error != "" {
		return nil, fmt.Errorf("ollama API error: %s", ollamaResp.Error)
	}

	response := &AIResponse{
		Content:  ollamaResp.Response,
		Provider: string(ProviderOllama),
		Model:    p.model,
	}

	return response, nil
}

func (p *OllamaProvider) Chat(ctx context.Context, req AIRequest) (*AIResponse, error) {
	messages := []OllamaMessage{
		{Role: "user", Content: req.Prompt},
	}
	
	if req.Context != "" {
		messages = []OllamaMessage{
			{Role: "system", Content: req.Context},
			{Role: "user", Content: req.Prompt},
		}
	}
	
	ollamaReq := OllamaChatRequest{
		Model:    p.model,
		Messages: messages,
		Stream:   false,
	}

	jsonData, err := json.Marshal(ollamaReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/api/chat", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	var ollamaResp OllamaChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	if ollamaResp.Error != "" {
		return nil, fmt.Errorf("ollama API error: %s", ollamaResp.Error)
	}

	response := &AIResponse{
		Content:  ollamaResp.Message.Content,
		Provider: string(ProviderOllama),
		Model:    p.model,
	}

	return response, nil
}

func (p *OllamaProvider) Analyze(ctx context.Context, req AIRequest) (*AIResponse, error) {
	prompt := p.buildPrompt(req, "analysis")
	
	ollamaReq := OllamaRequest{
		Model:  p.model,
		Prompt: prompt,
		Stream: false,
	}

	jsonData, err := json.Marshal(ollamaReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/api/generate", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	var ollamaResp OllamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	if ollamaResp.Error != "" {
		return nil, fmt.Errorf("ollama API error: %s", ollamaResp.Error)
	}

	response := &AIResponse{
		Content:  ollamaResp.Response,
		Provider: string(ProviderOllama),
		Model:    p.model,
	}

	return response, nil
}

func (p *OllamaProvider) buildPrompt(req AIRequest, requestType string) string {
	switch requestType {
	case "completion":
		return fmt.Sprintf("Complete the following code:\n\n%s", req.Prompt)
	case "analysis":
		return fmt.Sprintf("Analyze the following code and provide suggestions for improvement:\n\n%s", req.Prompt)
	default:
		return req.Prompt
	}
}