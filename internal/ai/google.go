package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type GoogleProvider struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

type GoogleRequest struct {
	Contents []GoogleContent `json:"contents"`
}

type GoogleContent struct {
	Parts []GooglePart `json:"parts"`
	Role  string       `json:"role,omitempty"`
}

type GooglePart struct {
	Text string `json:"text"`
}

type GoogleResponse struct {
	Candidates []GoogleCandidate `json:"candidates"`
	Error      *GoogleError      `json:"error,omitempty"`
}

type GoogleCandidate struct {
	Content GoogleContent `json:"content"`
}

type GoogleError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

func NewGoogleProvider() *GoogleProvider {
	return &GoogleProvider{
		baseURL: "https://generativelanguage.googleapis.com/v1beta/models",
		model:   "gemini-1.5-flash",
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (p *GoogleProvider) Name() ProviderType {
	return ProviderGoogle
}

func (p *GoogleProvider) IsAvailable() bool {
	return p.apiKey != ""
}

func (p *GoogleProvider) Configure(config ProviderConfig) error {
	if config.APIKey == "" {
		return fmt.Errorf("google provider requires API key")
	}
	
	p.apiKey = config.APIKey
	
	if config.BaseURL != "" {
		p.baseURL = config.BaseURL
	}
	
	if config.Model != "" {
		p.model = config.Model
	}
	
	return nil
}

func (p *GoogleProvider) Complete(ctx context.Context, req AIRequest) (*AIResponse, error) {
	return p.generateContent(ctx, req, "completion")
}

func (p *GoogleProvider) Chat(ctx context.Context, req AIRequest) (*AIResponse, error) {
	return p.generateContent(ctx, req, "chat")
}

func (p *GoogleProvider) Analyze(ctx context.Context, req AIRequest) (*AIResponse, error) {
	return p.generateContent(ctx, req, "analysis")
}

func (p *GoogleProvider) generateContent(ctx context.Context, req AIRequest, requestType string) (*AIResponse, error) {
	if !p.IsAvailable() {
		return nil, fmt.Errorf("google provider not configured")
	}

	prompt := p.buildPrompt(req, requestType)
	
	googleReq := GoogleRequest{
		Contents: []GoogleContent{
			{
				Parts: []GooglePart{
					{Text: prompt},
				},
			},
		},
	}

	jsonData, err := json.Marshal(googleReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	url := fmt.Sprintf("%s/%s:generateContent?key=%s", p.baseURL, p.model, p.apiKey)
	
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	var googleResp GoogleResponse
	if err := json.NewDecoder(resp.Body).Decode(&googleResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	if googleResp.Error != nil {
		return nil, fmt.Errorf("google API error: %s", googleResp.Error.Message)
	}

	if len(googleResp.Candidates) == 0 || len(googleResp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no response content from google")
	}

	response := &AIResponse{
		Content:  googleResp.Candidates[0].Content.Parts[0].Text,
		Provider: string(ProviderGoogle),
		Model:    p.model,
	}

	return response, nil
}

func (p *GoogleProvider) buildPrompt(req AIRequest, requestType string) string {
	switch requestType {
	case "completion":
		return fmt.Sprintf("Complete the following code:\n\n%s", req.Prompt)
	case "chat":
		if req.Context != "" {
			return fmt.Sprintf("Context: %s\n\nQuestion: %s", req.Context, req.Prompt)
		}
		return req.Prompt
	case "analysis":
		return fmt.Sprintf("Analyze the following code and provide suggestions:\n\n%s", req.Prompt)
	default:
		return req.Prompt
	}
}