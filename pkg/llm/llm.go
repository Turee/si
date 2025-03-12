package llm

import (
	"context"
	"errors"

	"github.com/ture/si/pkg/config"
)

// ErrNotImplemented is returned when a feature is not yet implemented
var ErrNotImplemented = errors.New("not implemented")

// Provider defines the interface for LLM providers
type Provider interface {
	// Ask sends a question to the LLM and returns the response
	Ask(ctx context.Context, question string) (string, error)
}

// NewProvider creates a new LLM provider based on the configuration
func NewProvider(cfg *config.Config) (Provider, error) {
	// For now, we only support OpenAI
	return NewOpenAIProvider(&cfg.LLM.OpenAI)
}

// NewOpenAIProvider creates a new OpenAI provider
func NewOpenAIProvider(cfg *config.OpenAIConfig) (Provider, error) {
	return &openAIProvider{
		cfg: cfg,
	}, nil
}

// openAIProvider implements the Provider interface for OpenAI
type openAIProvider struct {
	cfg *config.OpenAIConfig
}

// Ask implements the Provider interface
func (p *openAIProvider) Ask(ctx context.Context, question string) (string, error) {
	// This is just a placeholder implementation
	// TODO: Implement actual OpenAI API call
	return "", ErrNotImplemented
}
