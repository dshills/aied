package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAnthropicProvider_Configure(t *testing.T) {
	provider := NewAnthropicProvider()
	
	config := ProviderConfig{
		APIKey:  "test-anthropic-key",
		Model:   "claude-3-opus",
		BaseURL: "https://test.anthropic.com",
		Options: map[string]interface{}{
			"timeout": 45,
		},
	}
	
	err := provider.Configure(config)
	if err != nil {
		t.Fatalf("Configure failed: %v", err)
	}
	
	if provider.apiKey != "test-anthropic-key" {
		t.Errorf("Expected API key 'test-anthropic-key', got '%s'", provider.apiKey)
	}
	if provider.model != "claude-3-opus" {
		t.Errorf("Expected model 'claude-3-opus', got '%s'", provider.model)
	}
	if provider.baseURL != "https://test.anthropic.com" {
		t.Errorf("Expected base URL 'https://test.anthropic.com', got '%s'", provider.baseURL)
	}
}

func TestAnthropicProvider_Complete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify headers
		if r.URL.Path != "/messages" {
			t.Errorf("Expected path /messages, got %s", r.URL.Path)
		}
		if r.Header.Get("x-api-key") != "test-key" {
			t.Errorf("Expected x-api-key header 'test-key', got '%s'", r.Header.Get("x-api-key"))
		}
		if r.Header.Get("anthropic-version") != "2023-06-01" {
			t.Errorf("Expected anthropic-version header '2023-06-01', got '%s'", r.Header.Get("anthropic-version"))
		}
		
		// Parse request
		var req anthropicRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}
		
		// Verify system prompt for completion
		if !strings.Contains(req.System, "code completion assistant") {
			t.Errorf("Expected system prompt to contain 'code completion assistant', got '%s'", req.System)
		}
		
		// Send response
		resp := anthropicResponse{
			Content: []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			}{
				{
					Type: "text",
					Text: "Completed code here",
				},
			},
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()
	
	provider := NewAnthropicProvider()
	provider.Configure(ProviderConfig{
		APIKey:  "test-key",
		BaseURL: server.URL,
	})
	
	ctx := context.Background()
	req := AIRequest{
		Prompt:   "Complete this",
		Language: "python",
		Type:     RequestCompletion,
	}
	
	resp, err := provider.Complete(ctx, req)
	if err != nil {
		t.Fatalf("Complete failed: %v", err)
	}
	
	if resp.Content != "Completed code here" {
		t.Errorf("Expected content 'Completed code here', got '%s'", resp.Content)
	}
	if resp.Provider != "anthropic" {
		t.Errorf("Expected provider 'anthropic', got '%s'", resp.Provider)
	}
}

func TestAnthropicProvider_Chat(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req anthropicRequest
		json.NewDecoder(r.Body).Decode(&req)
		
		// Check that system prompt is set based on request type
		expectedSystem := "You are a helpful AI programming assistant"
		if !strings.Contains(req.System, expectedSystem) {
			t.Errorf("Expected system to contain '%s', got '%s'", expectedSystem, req.System)
		}
		
		resp := anthropicResponse{
			Content: []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			}{
				{
					Type: "text",
					Text: "Chat response from Claude",
				},
			},
		}
		
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()
	
	provider := NewAnthropicProvider()
	provider.Configure(ProviderConfig{
		APIKey:  "test-key",
		BaseURL: server.URL,
	})
	
	ctx := context.Background()
	req := AIRequest{
		Prompt: "Hello Claude",
		Type:   RequestChat,
	}
	
	resp, err := provider.Chat(ctx, req)
	if err != nil {
		t.Fatalf("Chat failed: %v", err)
	}
	
	if resp.Content != "Chat response from Claude" {
		t.Errorf("Expected content 'Chat response from Claude', got '%s'", resp.Content)
	}
}

func TestAnthropicProvider_Analyze(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req anthropicRequest
		json.NewDecoder(r.Body).Decode(&req)
		
		// Check system prompt for analysis
		if !strings.Contains(req.System, "code reviewer") {
			t.Errorf("Expected system to contain 'code reviewer', got '%s'", req.System)
		}
		
		resp := anthropicResponse{
			Content: []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			}{
				{
					Type: "text",
					Text: "Code analysis results",
				},
			},
		}
		
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()
	
	provider := NewAnthropicProvider()
	provider.Configure(ProviderConfig{
		APIKey:  "test-key",
		BaseURL: server.URL,
	})
	
	ctx := context.Background()
	req := AIRequest{
		Context:  "def bad_function():\n    pass",
		Language: "python",
		Type:     RequestRefactor,
	}
	
	resp, err := provider.Analyze(ctx, req)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}
	
	if resp.Content != "Code analysis results" {
		t.Errorf("Expected content 'Code analysis results', got '%s'", resp.Content)
	}
}

func TestAnthropicProvider_ErrorHandling(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse func(w http.ResponseWriter, r *http.Request)
		wantErr        bool
		errContains    string
	}{
		{
			name: "HTTP Error",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("Bad Request"))
			},
			wantErr:     true,
			errContains: "status 400",
		},
		{
			name: "Empty Content",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				resp := anthropicResponse{
					Content: []struct {
						Type string `json:"type"`
						Text string `json:"text"`
					}{},
				}
				json.NewEncoder(w).Encode(resp)
			},
			wantErr: false, // Should handle empty content gracefully
		},
		{
			name: "Invalid JSON",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("{invalid json"))
			},
			wantErr:     true,
			errContains: "failed to decode response",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tt.serverResponse))
			defer server.Close()
			
			provider := NewAnthropicProvider()
			provider.Configure(ProviderConfig{
				APIKey:  "test-key",
				BaseURL: server.URL,
			})
			
			ctx := context.Background()
			req := AIRequest{Prompt: "test"}
			
			_, err := provider.Chat(ctx, req)
			if (err != nil) != tt.wantErr {
				t.Errorf("Chat() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("Error should contain '%s', got '%s'", tt.errContains, err.Error())
			}
		})
	}
}

func TestAnthropicProvider_GetSystemPrompt(t *testing.T) {
	provider := NewAnthropicProvider()
	
	tests := []struct {
		reqType RequestType
		expect  string
	}{
		{RequestExplanation, "programming educator"},
		{RequestDebug, "expert debugger"},
		{RequestDocumentation, "technical documentation expert"},
		{RequestRefactor, "senior software engineer"},
		{RequestChat, "helpful AI programming assistant"},
	}
	
	for _, tt := range tests {
		t.Run(string(tt.reqType), func(t *testing.T) {
			prompt := provider.getSystemPrompt(tt.reqType)
			if !strings.Contains(prompt, tt.expect) {
				t.Errorf("System prompt for %s should contain '%s', got '%s'", tt.reqType, tt.expect, prompt)
			}
		})
	}
}