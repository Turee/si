package llm

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ture/si/pkg/config"
)

// TestNewProvider tests the NewProvider function
func TestNewProvider(t *testing.T) {
	// Create a test config
	cfg := &config.Config{
		LLM: config.LLMConfig{
			OpenAI: config.OpenAIConfig{
				BaseURL: "https://api.openai.com/v1",
				APIKey:  "test-api-key",
			},
		},
	}

	// Create a provider
	provider, err := NewProvider(cfg)

	// Check that no error occurred
	assert.NoError(t, err)

	// Check that the provider is not nil
	assert.NotNil(t, provider)

	// Check that the provider is of the correct type
	_, ok := provider.(*openAIProvider)
	assert.True(t, ok)
}

// TestOpenAIProviderAsk tests the Ask method of the openAIProvider
func TestOpenAIProviderAsk(t *testing.T) {
	// Create a test config
	cfg := &config.OpenAIConfig{
		BaseURL: "https://api.openai.com/v1",
		APIKey:  "test-api-key",
	}

	// Create a provider
	provider, err := NewOpenAIProvider(cfg)
	assert.NoError(t, err)

	// Ask a question
	_, err = provider.Ask(context.Background(), "What is the capital of France?")

	// Since we haven't implemented the actual LLM functionality,
	// we expect an ErrNotImplemented error
	assert.Error(t, err)
	assert.Equal(t, ErrNotImplemented, err)
}
