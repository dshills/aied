package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestOpenAIProvider_Configure(t *testing.T) {
	provider := NewOpenAIProvider()
	
	// Test basic configuration
	config := ProviderConfig{
		APIKey:  "test-key",
		Model:   "gpt-3.5-turbo",
		BaseURL: "https://test.openai.com",
		Options: map[string]interface{}{
			"timeout": 60,
		},
	}
	
	err := provider.Configure(config)
	if err != nil {
		t.Fatalf("Configure failed: %v", err)
	}
	
	if provider.apiKey != "test-key" {
		t.Errorf("Expected API key 'test-key', got '%s'", provider.apiKey)
	}
	if provider.model != "gpt-3.5-turbo" {
		t.Errorf("Expected model 'gpt-3.5-turbo', got '%s'", provider.model)
	}
	if provider.baseURL != "https://test.openai.com" {
		t.Errorf("Expected base URL 'https://test.openai.com', got '%s'", provider.baseURL)
	}
}

func TestOpenAIProvider_IsAvailable(t *testing.T) {
	provider := NewOpenAIProvider()
	
	// Should not be available without API key
	if provider.IsAvailable() {
		t.Error("Provider should not be available without API key")
	}
	
	// Configure with API key
	provider.Configure(ProviderConfig{APIKey: "test-key"})
	
	// Should be available with API key
	if !provider.IsAvailable() {
		t.Error("Provider should be available with API key")
	}
}

func TestOpenAIProvider_Complete(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.URL.Path != "/chat/completions" {
			t.Errorf("Expected path /chat/completions, got %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("Expected Authorization header 'Bearer test-key', got '%s'", r.Header.Get("Authorization"))
		}
		
		// Parse request body
		var req openAIChatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}
		
		// Verify request content
		if req.Model != "gpt-4" {
			t.Errorf("Expected model 'gpt-4', got '%s'", req.Model)
		}
		if len(req.Messages) != 2 {
			t.Errorf("Expected 2 messages, got %d", len(req.Messages))
		}
		
		// Send response
		resp := openAIChatResponse{
			Choices: []struct {
				Index   int `json:"index"`
				Message struct {
					Role    string `json:"role"`
					Content string `json:"content"`
				} `json:"message"`
				FinishReason string `json:"finish_reason"`
			}{
				{
					Message: struct {
						Role    string `json:"role"`
						Content string `json:"content"`
					}{
						Role:    "assistant",
						Content: "Completed code",
					},
				},
			},
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()
	
	// Configure provider with test server
	provider := NewOpenAIProvider()
	provider.Configure(ProviderConfig{
		APIKey:  "test-key",
		Model:   "gpt-4",
		BaseURL: server.URL,
	})
	
	// Test completion
	ctx := context.Background()
	req := AIRequest{
		Prompt:   "Complete this",
		Context:  "function test() {",
		Language: "javascript",
		Type:     RequestCompletion,
	}
	
	resp, err := provider.Complete(ctx, req)
	if err != nil {
		t.Fatalf("Complete failed: %v", err)
	}
	
	if resp.Content != "Completed code" {
		t.Errorf("Expected content 'Completed code', got '%s'", resp.Content)
	}
	if resp.Provider != "openai" {
		t.Errorf("Expected provider 'openai', got '%s'", resp.Provider)
	}
}

func TestOpenAIProvider_Chat(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := openAIChatResponse{
			Choices: []struct {
				Index   int `json:"index"`
				Message struct {
					Role    string `json:"role"`
					Content string `json:"content"`
				} `json:"message"`
				FinishReason string `json:"finish_reason"`
			}{
				{
					Message: struct {
						Role    string `json:"role"`
						Content string `json:"content"`
					}{
						Role:    "assistant",
						Content: "Chat response",
					},
				},
			},
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()
	
	provider := NewOpenAIProvider()
	provider.Configure(ProviderConfig{
		APIKey:  "test-key",
		BaseURL: server.URL,
	})
	
	ctx := context.Background()
	req := AIRequest{
		Prompt: "Hello",
		Type:   RequestChat,
	}
	
	resp, err := provider.Chat(ctx, req)
	if err != nil {
		t.Fatalf("Chat failed: %v", err)
	}
	
	if resp.Content != "Chat response" {
		t.Errorf("Expected content 'Chat response', got '%s'", resp.Content)
	}
}

func TestOpenAIProvider_ErrorHandling(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse func(w http.ResponseWriter, r *http.Request)
		wantErr        bool
		errContains    string
	}{
		{
			name: "HTTP Error",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Internal Server Error"))
			},
			wantErr:     true,
			errContains: "status 500",
		},
		{
			name: "No Choices",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				resp := openAIChatResponse{
					Choices: []struct {
						Index   int `json:"index"`
						Message struct {
							Role    string `json:"role"`
							Content string `json:"content"`
						} `json:"message"`
						FinishReason string `json:"finish_reason"`
					}{},
				}
				json.NewEncoder(w).Encode(resp)
			},
			wantErr:     true,
			errContains: "no response choices",
		},
		{
			name: "Invalid JSON",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("invalid json"))
			},
			wantErr:     true,
			errContains: "failed to decode response",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tt.serverResponse))
			defer server.Close()
			
			provider := NewOpenAIProvider()
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
			if err != nil && !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("Error should contain '%s', got '%s'", tt.errContains, err.Error())
			}
		})
	}
}

func TestOpenAIProvider_BuildPrompts(t *testing.T) {
	provider := NewOpenAIProvider()
	
	tests := []struct {
		name     string
		req      AIRequest
		method   string
		contains []string
	}{
		{
			name: "Completion with language",
			req: AIRequest{
				Prompt:   "test",
				Language: "go",
				Context:  "context here",
			},
			method:   "completion",
			contains: []string{"Complete this go code", "Context:", "context here", "Code to complete:", "test"},
		},
		{
			name: "Chat with context",
			req: AIRequest{
				Prompt:  "question",
				Context: "some context",
			},
			method:   "chat",
			contains: []string{"Context:", "some context", "Question: question"},
		},
		{
			name: "Analysis with language",
			req: AIRequest{
				Context:  "code here",
				Language: "python",
				Prompt:   "focus on performance",
			},
			method:   "analysis",
			contains: []string{"Analyze this python code", "Code:", "code here", "Specific focus: focus on performance"},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result string
			switch tt.method {
			case "completion":
				result = provider.buildCompletionPrompt(tt.req)
			case "chat":
				result = provider.buildChatPrompt(tt.req)
			case "analysis":
				result = provider.buildAnalysisPrompt(tt.req)
			}
			
			for _, expected := range tt.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected prompt to contain '%s', got: %s", expected, result)
				}
			}
		})
	}
}