package ai

import (
	"context"
	"fmt"
)

// MockProvider is a mock AI provider for testing
type MockProvider struct {
	name           ProviderType
	available      bool
	completeFunc   func(context.Context, AIRequest) (*AIResponse, error)
	chatFunc       func(context.Context, AIRequest) (*AIResponse, error)
	analyzeFunc    func(context.Context, AIRequest) (*AIResponse, error)
	configureFunc  func(ProviderConfig) error
	configCalls    []ProviderConfig
	completeCalls  []AIRequest
	chatCalls      []AIRequest
	analyzeCalls   []AIRequest
}

// NewMockProvider creates a new mock provider
func NewMockProvider(name ProviderType) *MockProvider {
	return &MockProvider{
		name:      name,
		available: true,
		completeFunc: func(ctx context.Context, req AIRequest) (*AIResponse, error) {
			return &AIResponse{
				Content:  "mock completion",
				Provider: string(name),
				Model:    "mock-model",
			}, nil
		},
		chatFunc: func(ctx context.Context, req AIRequest) (*AIResponse, error) {
			return &AIResponse{
				Content:  "mock chat response",
				Provider: string(name),
				Model:    "mock-model",
			}, nil
		},
		analyzeFunc: func(ctx context.Context, req AIRequest) (*AIResponse, error) {
			return &AIResponse{
				Content:  "mock analysis",
				Provider: string(name),
				Model:    "mock-model",
			}, nil
		},
		configureFunc: func(config ProviderConfig) error {
			return nil
		},
		configCalls:   []ProviderConfig{},
		completeCalls: []AIRequest{},
		chatCalls:     []AIRequest{},
		analyzeCalls:  []AIRequest{},
	}
}

func (m *MockProvider) Name() ProviderType {
	return m.name
}

func (m *MockProvider) IsAvailable() bool {
	return m.available
}

func (m *MockProvider) Complete(ctx context.Context, req AIRequest) (*AIResponse, error) {
	m.completeCalls = append(m.completeCalls, req)
	if m.completeFunc != nil {
		return m.completeFunc(ctx, req)
	}
	return nil, fmt.Errorf("complete not implemented")
}

func (m *MockProvider) Chat(ctx context.Context, req AIRequest) (*AIResponse, error) {
	m.chatCalls = append(m.chatCalls, req)
	if m.chatFunc != nil {
		return m.chatFunc(ctx, req)
	}
	return nil, fmt.Errorf("chat not implemented")
}

func (m *MockProvider) Analyze(ctx context.Context, req AIRequest) (*AIResponse, error) {
	m.analyzeCalls = append(m.analyzeCalls, req)
	if m.analyzeFunc != nil {
		return m.analyzeFunc(ctx, req)
	}
	return nil, fmt.Errorf("analyze not implemented")
}

func (m *MockProvider) Configure(config ProviderConfig) error {
	m.configCalls = append(m.configCalls, config)
	if m.configureFunc != nil {
		return m.configureFunc(config)
	}
	return nil
}

// Test helpers

func (m *MockProvider) SetAvailable(available bool) {
	m.available = available
}

func (m *MockProvider) SetCompleteFunc(f func(context.Context, AIRequest) (*AIResponse, error)) {
	m.completeFunc = f
}

func (m *MockProvider) SetChatFunc(f func(context.Context, AIRequest) (*AIResponse, error)) {
	m.chatFunc = f
}

func (m *MockProvider) SetAnalyzeFunc(f func(context.Context, AIRequest) (*AIResponse, error)) {
	m.analyzeFunc = f
}

func (m *MockProvider) SetConfigureFunc(f func(ProviderConfig) error) {
	m.configureFunc = f
}

func (m *MockProvider) GetCompleteCalls() []AIRequest {
	return m.completeCalls
}

func (m *MockProvider) GetChatCalls() []AIRequest {
	return m.chatCalls
}

func (m *MockProvider) GetAnalyzeCalls() []AIRequest {
	return m.analyzeCalls
}

func (m *MockProvider) GetConfigCalls() []ProviderConfig {
	return m.configCalls
}

func (m *MockProvider) Reset() {
	m.configCalls = []ProviderConfig{}
	m.completeCalls = []AIRequest{}
	m.chatCalls = []AIRequest{}
	m.analyzeCalls = []AIRequest{}
}