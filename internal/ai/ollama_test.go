package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestOllamaProvider_Configure(t *testing.T) {
	provider := NewOllamaProvider()
	
	config := ProviderConfig{
		BaseURL: "http://localhost:8080",
		Model:   "codellama",
	}
	
	err := provider.Configure(config)
	if err != nil {
		t.Fatalf("Configure failed: %v", err)
	}
	
	if provider.baseURL != "http://localhost:8080" {
		t.Errorf("Expected base URL 'http://localhost:8080', got '%s'", provider.baseURL)
	}
	if provider.model != "codellama" {
		t.Errorf("Expected model 'codellama', got '%s'", provider.model)
	}
}

func TestOllamaProvider_IsAvailable(t *testing.T) {
	// Test with unavailable server
	provider := NewOllamaProvider()
	provider.baseURL = "http://localhost:99999" // Non-existent port
	
	if provider.IsAvailable() {
		t.Error("Provider should not be available with invalid server")
	}
	
	// Test with mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/tags" {
			resp := OllamaTagsResponse{
				Models: []OllamaModel{
					{Name: "llama2"},
					{Name: "codellama"},
				},
			}
			json.NewEncoder(w).Encode(resp)
		}
	}))
	defer server.Close()
	
	provider.baseURL = server.URL
	if !provider.IsAvailable() {
		t.Error("Provider should be available with valid server")
	}
}

func TestOllamaProvider_Complete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/generate" {
			t.Errorf("Expected path /api/generate, got %s", r.URL.Path)
		}
		
		// Parse request
		var req OllamaRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}
		
		// Verify request
		if req.Model != "llama2" {
			t.Errorf("Expected model 'llama2', got '%s'", req.Model)
		}
		if req.Stream != false {
			t.Error("Expected stream to be false")
		}
		if !strings.Contains(req.Prompt, "Complete the following code") {
			t.Error("Expected prompt to contain completion instruction")
		}
		
		// Send response
		resp := OllamaResponse{
			Response: "Completed code from Ollama",
			Done:     true,
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()
	
	provider := NewOllamaProvider()
	provider.Configure(ProviderConfig{
		BaseURL: server.URL,
		Model:   "llama2",
	})
	
	ctx := context.Background()
	req := AIRequest{
		Prompt:   "function test() {",
		Language: "javascript",
		Type:     RequestCompletion,
	}
	
	resp, err := provider.Complete(ctx, req)
	if err != nil {
		t.Fatalf("Complete failed: %v", err)
	}
	
	if resp.Content != "Completed code from Ollama" {
		t.Errorf("Expected content 'Completed code from Ollama', got '%s'", resp.Content)
	}
	if resp.Provider != "ollama" {
		t.Errorf("Expected provider 'ollama', got '%s'", resp.Provider)
	}
}

func TestOllamaProvider_Chat(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/chat" {
			t.Errorf("Expected path /api/chat, got %s", r.URL.Path)
		}
		
		// Parse request
		var req OllamaChatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}
		
		// Verify messages
		if len(req.Messages) == 0 {
			t.Error("Expected at least one message")
		}
		
		// Send response
		resp := OllamaChatResponse{
			Message: OllamaMessage{
				Role:    "assistant",
				Content: "Chat response from Ollama",
			},
			Done: true,
		}
		
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()
	
	provider := NewOllamaProvider()
	provider.Configure(ProviderConfig{
		BaseURL: server.URL,
	})
	
	ctx := context.Background()
	req := AIRequest{
		Prompt:  "Hello Ollama",
		Context: "System context",
		Type:    RequestChat,
	}
	
	resp, err := provider.Chat(ctx, req)
	if err != nil {
		t.Fatalf("Chat failed: %v", err)
	}
	
	if resp.Content != "Chat response from Ollama" {
		t.Errorf("Expected content 'Chat response from Ollama', got '%s'", resp.Content)
	}
}

func TestOllamaProvider_Analyze(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req OllamaRequest
		json.NewDecoder(r.Body).Decode(&req)
		
		// Check that analyze prompt is properly formatted
		if !strings.Contains(req.Prompt, "Analyze the following code") {
			t.Errorf("Expected prompt to contain 'Analyze the following code', got: %s", req.Prompt)
		}
		
		resp := OllamaResponse{
			Response: "Analysis from Ollama",
			Done:     true,
		}
		
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()
	
	provider := NewOllamaProvider()
	provider.Configure(ProviderConfig{
		BaseURL: server.URL,
	})
	
	ctx := context.Background()
	req := AIRequest{
		Context:  "def analyze_me():\n    pass",
		Language: "python",
		Type:     RequestRefactor,
	}
	
	resp, err := provider.Analyze(ctx, req)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}
	
	if resp.Content != "Analysis from Ollama" {
		t.Errorf("Expected content 'Analysis from Ollama', got '%s'", resp.Content)
	}
}

func TestOllamaProvider_ErrorHandling(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse func(w http.ResponseWriter, r *http.Request)
		wantErr        bool
		errContains    string
	}{
		{
			name: "API Error in Response",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				resp := OllamaResponse{
					Error: "Model not found",
				}
				json.NewEncoder(w).Encode(resp)
			},
			wantErr:     true,
			errContains: "Model not found",
		},
		{
			name: "Chat API Error",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				resp := OllamaChatResponse{
					Error: "Invalid request",
				}
				json.NewEncoder(w).Encode(resp)
			},
			wantErr:     true,
			errContains: "Invalid request",
		},
		{
			name: "HTTP Error",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte("Not Found"))
			},
			wantErr:     true,
			errContains: "failed to decode response",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tt.serverResponse))
			defer server.Close()
			
			provider := NewOllamaProvider()
			provider.Configure(ProviderConfig{
				BaseURL: server.URL,
			})
			
			ctx := context.Background()
			req := AIRequest{Prompt: "test"}
			
			// Try both Complete and Chat depending on error type
			_, err := provider.Complete(ctx, req)
			if err == nil {
				_, err = provider.Chat(ctx, req)
			}
			
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("Error should contain '%s', got '%s'", tt.errContains, err.Error())
			}
		})
	}
}

func TestOllamaProvider_BuildPrompt(t *testing.T) {
	provider := NewOllamaProvider()
	
	tests := []struct {
		name        string
		req         AIRequest
		requestType string
		contains    string
	}{
		{
			name: "Completion prompt",
			req: AIRequest{
				Prompt: "test code",
			},
			requestType: "completion",
			contains:    "Complete the following code:\n\ntest code",
		},
		{
			name: "Analysis prompt with focus",
			req: AIRequest{
				Context: "code here",
				Prompt:  "focus on security",
			},
			requestType: "analysis",
			contains:    "Specific focus: focus on security",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := provider.buildPrompt(tt.req, tt.requestType)
			if !strings.Contains(result, tt.contains) {
				t.Errorf("Expected prompt to contain '%s', got: %s", tt.contains, result)
			}
		})
	}
}