package ai

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestAIManager_RegisterProvider(t *testing.T) {
	manager := NewAIManager()
	provider := NewMockProvider("test-provider")
	
	// Test registering a provider
	err := manager.RegisterProvider(provider)
	if err != nil {
		t.Errorf("Failed to register provider: %v", err)
	}
	
	// Verify provider is registered
	registered, exists := manager.GetProvider("test-provider")
	if !exists {
		t.Error("Provider not found after registration")
	}
	if registered != provider {
		t.Error("Retrieved provider doesn't match registered provider")
	}
	
	// Test registering nil provider
	err = manager.RegisterProvider(nil)
	if err == nil {
		t.Error("Expected error when registering nil provider")
	}
}

func TestAIManager_SetActiveProvider(t *testing.T) {
	manager := NewAIManager()
	provider1 := NewMockProvider("provider1")
	provider2 := NewMockProvider("provider2")
	provider2.SetAvailable(false)
	
	manager.RegisterProvider(provider1)
	manager.RegisterProvider(provider2)
	
	// Test setting available provider
	err := manager.SetActiveProvider("provider1")
	if err != nil {
		t.Errorf("Failed to set active provider: %v", err)
	}
	
	active := manager.GetActiveProvider()
	if active != provider1 {
		t.Error("Active provider doesn't match set provider")
	}
	
	// Test setting unavailable provider
	err = manager.SetActiveProvider("provider2")
	if err == nil {
		t.Error("Expected error when setting unavailable provider")
	}
	
	// Test setting non-existent provider
	err = manager.SetActiveProvider("non-existent")
	if err == nil {
		t.Error("Expected error when setting non-existent provider")
	}
}

func TestAIManager_Request(t *testing.T) {
	ctx := context.Background()
	manager := NewAIManager()
	
	// Set up providers
	provider1 := NewMockProvider(ProviderOpenAI)
	provider1.SetChatFunc(func(ctx context.Context, req AIRequest) (*AIResponse, error) {
		return &AIResponse{
			Content:  "response from provider1",
			Provider: string(ProviderOpenAI),
		}, nil
	})
	
	provider2 := NewMockProvider(ProviderOllama)
	provider2.SetChatFunc(func(ctx context.Context, req AIRequest) (*AIResponse, error) {
		return &AIResponse{
			Content:  "response from provider2",
			Provider: string(ProviderOllama),
		}, nil
	})
	
	manager.RegisterProvider(provider1)
	manager.RegisterProvider(provider2)
	manager.SetActiveProvider(ProviderOpenAI)
	
	// Test successful request
	req := AIRequest{
		Prompt: "test prompt",
		Type:   RequestChat,
	}
	
	resp, err := manager.Request(ctx, req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	if resp.Content != "response from provider1" {
		t.Errorf("Expected response from provider1, got: %s", resp.Content)
	}
	
	// Test fallback when active provider fails
	provider1.SetChatFunc(func(ctx context.Context, req AIRequest) (*AIResponse, error) {
		return nil, errors.New("provider1 failed")
	})
	
	resp, err = manager.Request(ctx, req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	if resp.Content != "response from provider2" {
		t.Errorf("Expected response from provider2 (fallback), got: %s", resp.Content)
	}
}

func TestAIManager_RequestTypes(t *testing.T) {
	ctx := context.Background()
	manager := NewAIManager()
	provider := NewMockProvider("test-provider")
	
	manager.RegisterProvider(provider)
	manager.SetActiveProvider("test-provider")
	
	tests := []struct {
		name        string
		requestType RequestType
		wantMethod  string
	}{
		{"Completion", RequestCompletion, "complete"},
		{"Chat", RequestChat, "chat"},
		{"Explanation", RequestExplanation, "chat"},
		{"Debug", RequestDebug, "chat"},
		{"Documentation", RequestDocumentation, "chat"},
		{"Refactor", RequestRefactor, "analyze"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider.Reset()
			
			req := AIRequest{
				Prompt: "test",
				Type:   tt.requestType,
			}
			
			_, err := manager.Request(ctx, req)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}
			
			// Check which method was called
			switch tt.wantMethod {
			case "complete":
				if len(provider.GetCompleteCalls()) != 1 {
					t.Error("Expected Complete to be called")
				}
			case "chat":
				if len(provider.GetChatCalls()) != 1 {
					t.Error("Expected Chat to be called")
				}
			case "analyze":
				if len(provider.GetAnalyzeCalls()) != 1 {
					t.Error("Expected Analyze to be called")
				}
			}
		})
	}
}

func TestAIManager_ConfigureProviders(t *testing.T) {
	manager := NewAIManager()
	
	configs := []ProviderConfig{
		{
			Type:    ProviderOpenAI,
			APIKey:  "key1",
			Enabled: true,
		},
		{
			Type:    ProviderAnthropic,
			APIKey:  "key2",
			Enabled: false, // Should not be registered
		},
		{
			Type:    ProviderGoogle,
			APIKey:  "key3",
			Enabled: true,
		},
	}
	
	err := manager.ConfigureProviders(configs)
	if err != nil {
		t.Fatalf("ConfigureProviders failed: %v", err)
	}
	
	// Check that enabled providers were registered
	provider1, exists := manager.GetProvider(ProviderOpenAI)
	if !exists {
		t.Error("OpenAI provider should be registered")
	} else if !provider1.IsAvailable() {
		t.Error("OpenAI provider should be available (has API key)")
	}
	
	_, exists = manager.GetProvider(ProviderAnthropic)
	if exists {
		t.Error("Anthropic provider should not be registered (disabled)")
	}
	
	provider3, exists := manager.GetProvider(ProviderGoogle)
	if !exists {
		t.Error("Google provider should be registered")
	} else if !provider3.IsAvailable() {
		t.Error("Google provider should be available (has API key)")
	}
}

func TestAIManager_ListProviders(t *testing.T) {
	manager := NewAIManager()
	provider1 := NewMockProvider("provider1")
	provider2 := NewMockProvider("provider2")
	provider2.SetAvailable(false)
	
	manager.RegisterProvider(provider1)
	manager.RegisterProvider(provider2)
	
	// Test ListProviders
	allProviders := manager.ListProviders()
	if len(allProviders) != 2 {
		t.Errorf("Expected 2 providers, got %d", len(allProviders))
	}
	
	// Test ListAvailableProviders
	availableProviders := manager.ListAvailableProviders()
	if len(availableProviders) != 1 {
		t.Errorf("Expected 1 available provider, got %d", len(availableProviders))
	}
	if _, exists := availableProviders["provider1"]; !exists {
		t.Error("provider1 should be in available providers")
	}
}

func TestAIManager_NoProviders(t *testing.T) {
	ctx := context.Background()
	manager := NewAIManager()
	
	req := AIRequest{
		Prompt: "test",
		Type:   RequestChat,
	}
	
	// Test request with no providers
	_, err := manager.Request(ctx, req)
	if err == nil {
		t.Error("Expected error when no providers are available")
	}
}

func TestAIManager_ContextCancellation(t *testing.T) {
	manager := NewAIManager()
	provider := NewMockProvider("test-provider")
	
	// Make provider respect context cancellation
	provider.SetChatFunc(func(ctx context.Context, req AIRequest) (*AIResponse, error) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(100 * time.Millisecond):
			return &AIResponse{Content: "success"}, nil
		}
	})
	
	manager.RegisterProvider(provider)
	manager.SetActiveProvider("test-provider")
	
	// Create a context that will be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	
	// Cancel immediately
	cancel()
	
	req := AIRequest{
		Prompt: "test",
		Type:   RequestChat,
	}
	
	_, err := manager.Request(ctx, req)
	if err == nil {
		t.Error("Expected error when context is cancelled")
	}
}

func TestCreateProvider(t *testing.T) {
	tests := []struct {
		name         string
		providerType ProviderType
		wantErr      bool
	}{
		{"OpenAI", ProviderOpenAI, false},
		{"Anthropic", ProviderAnthropic, false},
		{"Google", ProviderGoogle, false},
		{"Ollama", ProviderOllama, false},
		{"Unknown", "unknown-provider", true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := CreateProvider(tt.providerType)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateProvider() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && provider == nil {
				t.Error("Expected provider to be created")
			}
			if !tt.wantErr && provider.Name() != tt.providerType {
				t.Errorf("Provider name mismatch: got %v, want %v", provider.Name(), tt.providerType)
			}
		})
	}
}