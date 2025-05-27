package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGoogleProvider_Configure(t *testing.T) {
	provider := NewGoogleProvider()
	
	// Test without API key
	err := provider.Configure(ProviderConfig{})
	if err == nil {
		t.Error("Expected error when configuring without API key")
	}
	
	// Test with API key
	config := ProviderConfig{
		APIKey:  "test-google-key",
		Model:   "gemini-pro",
		BaseURL: "https://test.google.com",
	}
	
	err = provider.Configure(config)
	if err != nil {
		t.Fatalf("Configure failed: %v", err)
	}
	
	if provider.apiKey != "test-google-key" {
		t.Errorf("Expected API key 'test-google-key', got '%s'", provider.apiKey)
	}
	if provider.model != "gemini-pro" {
		t.Errorf("Expected model 'gemini-pro', got '%s'", provider.model)
	}
}

func TestGoogleProvider_Complete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify URL contains API key
		if !strings.Contains(r.URL.String(), "key=test-key") {
			t.Errorf("Expected URL to contain API key, got %s", r.URL.String())
		}
		
		// Verify path
		expectedPath := "/gemini-1.5-flash:generateContent"
		if !strings.Contains(r.URL.Path, expectedPath) {
			t.Errorf("Expected path to contain '%s', got %s", expectedPath, r.URL.Path)
		}
		
		// Parse request
		var req GoogleRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}
		
		// Verify request has content
		if len(req.Contents) == 0 || len(req.Contents[0].Parts) == 0 {
			t.Error("Expected request to have content")
		}
		
		// Send response
		resp := GoogleResponse{
			Candidates: []GoogleCandidate{
				{
					Content: GoogleContent{
						Parts: []GooglePart{
							{Text: "Completed code from Gemini"},
						},
					},
				},
			},
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()
	
	provider := NewGoogleProvider()
	provider.Configure(ProviderConfig{
		APIKey:  "test-key",
		BaseURL: server.URL,
	})
	
	ctx := context.Background()
	req := AIRequest{
		Prompt:   "Complete this code",
		Language: "java",
		Type:     RequestCompletion,
	}
	
	resp, err := provider.Complete(ctx, req)
	if err != nil {
		t.Fatalf("Complete failed: %v", err)
	}
	
	if resp.Content != "Completed code from Gemini" {
		t.Errorf("Expected content 'Completed code from Gemini', got '%s'", resp.Content)
	}
	if resp.Provider != "google" {
		t.Errorf("Expected provider 'google', got '%s'", resp.Provider)
	}
}

func TestGoogleProvider_Chat(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := GoogleResponse{
			Candidates: []GoogleCandidate{
				{
					Content: GoogleContent{
						Parts: []GooglePart{
							{Text: "Chat response from Gemini"},
						},
					},
				},
			},
		}
		
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()
	
	provider := NewGoogleProvider()
	provider.Configure(ProviderConfig{
		APIKey:  "test-key",
		BaseURL: server.URL,
	})
	
	ctx := context.Background()
	req := AIRequest{
		Prompt:  "Hello Gemini",
		Context: "Some context",
		Type:    RequestChat,
	}
	
	resp, err := provider.Chat(ctx, req)
	if err != nil {
		t.Fatalf("Chat failed: %v", err)
	}
	
	if resp.Content != "Chat response from Gemini" {
		t.Errorf("Expected content 'Chat response from Gemini', got '%s'", resp.Content)
	}
}

func TestGoogleProvider_ErrorHandling(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse func(w http.ResponseWriter, r *http.Request)
		wantErr        bool
		errContains    string
	}{
		{
			name: "API Error",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				resp := GoogleResponse{
					Error: &GoogleError{
						Code:    400,
						Message: "Invalid request",
						Status:  "INVALID_ARGUMENT",
					},
				}
				json.NewEncoder(w).Encode(resp)
			},
			wantErr:     true,
			errContains: "Invalid request",
		},
		{
			name: "No Candidates",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				resp := GoogleResponse{
					Candidates: []GoogleCandidate{},
				}
				json.NewEncoder(w).Encode(resp)
			},
			wantErr:     true,
			errContains: "no response content",
		},
		{
			name: "Empty Parts",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				resp := GoogleResponse{
					Candidates: []GoogleCandidate{
						{
							Content: GoogleContent{
								Parts: []GooglePart{},
							},
						},
					},
				}
				json.NewEncoder(w).Encode(resp)
			},
			wantErr:     true,
			errContains: "no response content",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tt.serverResponse))
			defer server.Close()
			
			provider := NewGoogleProvider()
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

func TestGoogleProvider_NotAvailable(t *testing.T) {
	provider := NewGoogleProvider()
	
	ctx := context.Background()
	req := AIRequest{Prompt: "test"}
	
	// Test Complete without configuration
	_, err := provider.Complete(ctx, req)
	if err == nil || !strings.Contains(err.Error(), "not configured") {
		t.Error("Expected 'not configured' error")
	}
	
	// Test Chat without configuration
	_, err = provider.Chat(ctx, req)
	if err == nil || !strings.Contains(err.Error(), "not configured") {
		t.Error("Expected 'not configured' error")
	}
	
	// Test Analyze without configuration
	_, err = provider.Analyze(ctx, req)
	if err == nil || !strings.Contains(err.Error(), "not configured") {
		t.Error("Expected 'not configured' error")
	}
}