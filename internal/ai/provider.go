package ai

import (
	"context"
	"fmt"
)

// ProviderType represents different AI providers
type ProviderType string

const (
	ProviderOpenAI    ProviderType = "openai"
	ProviderAnthropic ProviderType = "anthropic"
	ProviderGoogle    ProviderType = "google"
	ProviderOllama    ProviderType = "ollama"
)

// AIRequest represents a request to an AI provider
type AIRequest struct {
	Prompt      string            // The main prompt/question
	Context     string            // Additional context (code, file content, etc.)
	Language    string            // Programming language for context
	Type        RequestType       // Type of AI assistance requested
	Options     map[string]interface{} // Provider-specific options
}

// RequestType represents different types of AI assistance
type RequestType string

const (
	RequestCompletion   RequestType = "completion"   // Code completion
	RequestExplanation  RequestType = "explanation"  // Explain code/concept
	RequestRefactor     RequestType = "refactor"     // Suggest refactoring
	RequestDebug        RequestType = "debug"        // Debug help
	RequestDocumentation RequestType = "documentation" // Generate docs
	RequestChat         RequestType = "chat"         // General chat/help
)

// AIResponse represents the response from an AI provider
type AIResponse struct {
	Content     string   // The AI's response content
	Suggestions []string // List of suggestions (for completion)
	Confidence  float64  // Confidence score (0.0 to 1.0)
	Error       error    // Any error that occurred
	Provider    string   // Which provider generated this response
	Model       string   // Which model was used
}

// Provider defines the interface that all AI providers must implement
type Provider interface {
	// Name returns the provider name
	Name() ProviderType
	
	// IsAvailable checks if the provider is configured and available
	IsAvailable() bool
	
	// Complete generates code completion suggestions
	Complete(ctx context.Context, req AIRequest) (*AIResponse, error)
	
	// Chat handles general conversational requests
	Chat(ctx context.Context, req AIRequest) (*AIResponse, error)
	
	// Analyze provides code analysis and suggestions
	Analyze(ctx context.Context, req AIRequest) (*AIResponse, error)
	
	// Configure sets up the provider with configuration
	Configure(config ProviderConfig) error
}

// ProviderConfig holds configuration for AI providers
type ProviderConfig struct {
	Type       ProviderType          `yaml:"type"`
	APIKey     string                `yaml:"api_key"`
	BaseURL    string                `yaml:"base_url"`
	Model      string                `yaml:"model"`
	Options    map[string]interface{} `yaml:"options"`
	Enabled    bool                  `yaml:"enabled"`
}

// AIManager manages multiple AI providers and routing
type AIManager struct {
	providers     map[ProviderType]Provider
	activeProvider ProviderType
	fallbackOrder []ProviderType
}

// NewAIManager creates a new AI manager
func NewAIManager() *AIManager {
	return &AIManager{
		providers: make(map[ProviderType]Provider),
		fallbackOrder: []ProviderType{
			ProviderOllama,    // Local first (no API costs)
			ProviderOpenAI,    // Popular and reliable
			ProviderAnthropic, // High quality
			ProviderGoogle,    // Good alternative
		},
	}
}

// RegisterProvider registers an AI provider
func (am *AIManager) RegisterProvider(provider Provider) error {
	if provider == nil {
		return fmt.Errorf("provider cannot be nil")
	}
	
	am.providers[provider.Name()] = provider
	
	// Set as active if it's the first available provider
	if am.activeProvider == "" && provider.IsAvailable() {
		am.activeProvider = provider.Name()
	}
	
	return nil
}

// SetActiveProvider sets the active AI provider
func (am *AIManager) SetActiveProvider(providerType ProviderType) error {
	provider, exists := am.providers[providerType]
	if !exists {
		return fmt.Errorf("provider %s not registered", providerType)
	}
	
	if !provider.IsAvailable() {
		return fmt.Errorf("provider %s not available", providerType)
	}
	
	am.activeProvider = providerType
	return nil
}

// GetActiveProvider returns the currently active provider
func (am *AIManager) GetActiveProvider() Provider {
	if am.activeProvider == "" {
		return nil
	}
	
	return am.providers[am.activeProvider]
}

// GetProvider returns a specific provider by type
func (am *AIManager) GetProvider(providerType ProviderType) (Provider, bool) {
	provider, exists := am.providers[providerType]
	return provider, exists
}

// ListProviders returns all registered providers
func (am *AIManager) ListProviders() map[ProviderType]Provider {
	return am.providers
}

// ListAvailableProviders returns only available providers
func (am *AIManager) ListAvailableProviders() map[ProviderType]Provider {
	available := make(map[ProviderType]Provider)
	for pType, provider := range am.providers {
		if provider.IsAvailable() {
			available[pType] = provider
		}
	}
	return available
}

// Request makes an AI request using the active provider with fallback
func (am *AIManager) Request(ctx context.Context, req AIRequest) (*AIResponse, error) {
	// Try active provider first
	if am.activeProvider != "" {
		if provider := am.providers[am.activeProvider]; provider != nil && provider.IsAvailable() {
			response, err := am.makeRequest(ctx, provider, req)
			if err == nil {
				return response, nil
			}
			// Log error but continue to fallback
		}
	}
	
	// Try fallback providers
	for _, providerType := range am.fallbackOrder {
		provider, exists := am.providers[providerType]
		if !exists || !provider.IsAvailable() || providerType == am.activeProvider {
			continue
		}
		
		response, err := am.makeRequest(ctx, provider, req)
		if err == nil {
			return response, nil
		}
	}
	
	return nil, fmt.Errorf("no available AI providers")
}

// makeRequest makes a request to a specific provider based on request type
func (am *AIManager) makeRequest(ctx context.Context, provider Provider, req AIRequest) (*AIResponse, error) {
	switch req.Type {
	case RequestCompletion:
		return provider.Complete(ctx, req)
	case RequestChat, RequestExplanation, RequestDebug, RequestDocumentation:
		return provider.Chat(ctx, req)
	case RequestRefactor:
		return provider.Analyze(ctx, req)
	default:
		return provider.Chat(ctx, req)
	}
}

// ConfigureProviders configures multiple providers from config
func (am *AIManager) ConfigureProviders(configs []ProviderConfig) error {
	for _, config := range configs {
		if !config.Enabled {
			continue
		}
		
		provider, err := CreateProvider(config.Type)
		if err != nil {
			continue // Skip unavailable providers
		}
		
		if err := provider.Configure(config); err != nil {
			continue // Skip failed configurations
		}
		
		am.RegisterProvider(provider)
	}
	
	return nil
}

// CreateProvider creates a provider instance by type
func CreateProvider(providerType ProviderType) (Provider, error) {
	switch providerType {
	case ProviderOpenAI:
		return NewOpenAIProvider(), nil
	case ProviderAnthropic:
		return NewAnthropicProvider(), nil
	case ProviderGoogle:
		return NewGoogleProvider(), nil
	case ProviderOllama:
		return NewOllamaProvider(), nil
	default:
		return nil, fmt.Errorf("unknown provider type: %s", providerType)
	}
}