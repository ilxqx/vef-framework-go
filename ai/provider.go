package ai

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ModelProvider defines the interface for model providers.
type ModelProvider interface {
	// Name returns the provider's unique identifier.
	Name() string
	// CreateModel creates a new chat model instance.
	CreateModel(ctx context.Context, cfg *ModelConfig) (ToolableChatModel, error)
}

// AgentFactory defines the interface for agent factories.
type AgentFactory interface {
	// Name returns the agent type name.
	Name() string
	// CreateBuilder creates a new agent builder.
	CreateBuilder() AgentBuilder
}

// ModelConfig contains configuration for creating a model.
type ModelConfig struct {
	// Provider is the name of the model provider.
	Provider string
	// Model is the name of the model to use.
	Model string
	// ApiKey is the API key for authentication.
	ApiKey string
	// BaseUrl is the base URL for the API endpoint.
	BaseUrl string
	// Temperature controls randomness (0.0 to 1.0).
	Temperature float64
	// MaxTokens limits the maximum tokens in the response.
	MaxTokens int
	// Timeout is the request timeout duration.
	Timeout time.Duration
}

var (
	modelProviders = make(map[string]ModelProvider)
	agentFactories = make(map[string]AgentFactory)
	providerMu     sync.RWMutex
)

// RegisterModelProvider registers a model provider.
// It panics if a provider with the same name is already registered.
func RegisterModelProvider(p ModelProvider) {
	providerMu.Lock()
	defer providerMu.Unlock()

	name := p.Name()
	if _, exists := modelProviders[name]; exists {
		panic(fmt.Sprintf("ai: model provider %q already registered", name))
	}
	modelProviders[name] = p
}

// RegisterAgentFactory registers an agent factory.
// It panics if a factory with the same name is already registered.
func RegisterAgentFactory(f AgentFactory) {
	providerMu.Lock()
	defer providerMu.Unlock()

	name := f.Name()
	if _, exists := agentFactories[name]; exists {
		panic(fmt.Sprintf("ai: agent factory %q already registered", name))
	}
	agentFactories[name] = f
}

// NewChatModel creates a new chat model using the registered provider.
func NewChatModel(ctx context.Context, cfg *ModelConfig) (ToolableChatModel, error) {
	providerMu.RLock()
	p, ok := modelProviders[cfg.Provider]
	providerMu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("ai: unknown model provider %q", cfg.Provider)
	}
	return p.CreateModel(ctx, cfg)
}

// NewAgentBuilder creates a new agent builder using the registered factory.
func NewAgentBuilder(agentType string) (AgentBuilder, error) {
	providerMu.RLock()
	f, ok := agentFactories[agentType]
	providerMu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("ai: unknown agent type %q", agentType)
	}
	return f.CreateBuilder(), nil
}

// ListModelProviders returns the names of all registered model providers.
func ListModelProviders() []string {
	providerMu.RLock()
	defer providerMu.RUnlock()

	names := make([]string, 0, len(modelProviders))
	for name := range modelProviders {
		names = append(names, name)
	}
	return names
}

// ListAgentFactories returns the names of all registered agent factories.
func ListAgentFactories() []string {
	providerMu.RLock()
	defer providerMu.RUnlock()

	names := make([]string, 0, len(agentFactories))
	for name := range agentFactories {
		names = append(names, name)
	}
	return names
}
