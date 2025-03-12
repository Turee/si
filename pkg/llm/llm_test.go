package llm

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
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
	// Create a test server to simulate the OpenAI API
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request
		assert.Equal(t, "/v1/chat/completions", r.URL.Path)
		assert.Equal(t, "Bearer test-api-key", r.Header.Get("Authorization"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		// Simulate a streaming response
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)

		// Write a sample response
		resp := `data: {"id":"chatcmpl-123","object":"chat.completion.chunk","created":1694268190,"model":"gpt-3.5-turbo-0613","choices":[{"index":0,"delta":{"content":"Hello"},"finish_reason":null}]}

data: {"id":"chatcmpl-123","object":"chat.completion.chunk","created":1694268190,"model":"gpt-3.5-turbo-0613","choices":[{"index":0,"delta":{"content":" world"},"finish_reason":null}]}

data: {"id":"chatcmpl-123","object":"chat.completion.chunk","created":1694268190,"model":"gpt-3.5-turbo-0613","choices":[{"index":0,"delta":{"content":"!"},"finish_reason":null}]}

data: [DONE]
`
		w.Write([]byte(resp))
	}))
	defer server.Close()

	// Create the provider with the test server URL
	cfg := &config.OpenAIConfig{
		BaseURL: server.URL + "/v1",
		APIKey:  "test-api-key",
	}

	provider, err := NewOpenAIProvider(cfg)
	assert.NoError(t, err)

	// Test the Ask method
	answer, err := provider.Ask(context.Background(), "test question")
	assert.NoError(t, err)
	assert.Equal(t, "Hello world!", answer)
}

func TestOpenAIProviderBaseURLHandling(t *testing.T) {
	testCases := []struct {
		name            string
		baseURL         string
		azureDeployment string
		expectedPath    string
	}{
		{
			name:            "Standard OpenAI Base URL",
			baseURL:         "https://api.openai.com/v1",
			azureDeployment: "",
			expectedPath:    "/v1/chat/completions",
		},
		{
			name:            "Full OpenAI Endpoint URL",
			baseURL:         "https://api.openai.com/v1/chat/completions",
			azureDeployment: "",
			expectedPath:    "/v1/chat/completions",
		},
		{
			name:            "Base URL with Trailing Slash",
			baseURL:         "https://api.openai.com/v1/",
			azureDeployment: "",
			expectedPath:    "/v1/chat/completions",
		},
		{
			name:            "Azure OpenAI URL",
			baseURL:         "https://myresource.openai.azure.com",
			azureDeployment: "my-deployment",
			expectedPath:    "/openai/deployments/my-deployment/chat/completions",
		},
		{
			name:            "Azure OpenAI URL with Trailing Slash",
			baseURL:         "https://myresource.openai.azure.com/",
			azureDeployment: "my-deployment",
			expectedPath:    "/openai/deployments/my-deployment/chat/completions",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a test server to capture the request
			var capturedPath string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedPath = r.URL.Path

				// Simulate a streaming response
				w.Header().Set("Content-Type", "text/event-stream")
				w.WriteHeader(http.StatusOK)

				// Write a sample response
				resp := `data: {"id":"chatcmpl-123","object":"chat.completion.chunk","created":1694268190,"model":"gpt-3.5-turbo-0613","choices":[{"index":0,"delta":{"content":"Hello"},"finish_reason":null}]}

data: [DONE]
`
				w.Write([]byte(resp))
			}))
			defer server.Close()

			// Extract the host part from the test server URL
			serverURL := server.URL

			// Create the provider with the test configuration
			cfg := &config.OpenAIConfig{
				BaseURL:             tc.baseURL,
				APIKey:              "test-api-key",
				AzureDeploymentName: tc.azureDeployment,
			}

			// Replace the base URL with our test server URL while preserving the path
			if tc.azureDeployment != "" {
				cfg.BaseURL = serverURL
			} else if strings.Contains(tc.baseURL, "/chat/completions") {
				cfg.BaseURL = serverURL + "/v1/chat/completions"
			} else {
				cfg.BaseURL = serverURL + "/v1"
			}

			provider, err := NewOpenAIProvider(cfg)
			assert.NoError(t, err)

			// Make a request
			var result strings.Builder
			err = provider.AskStream(context.Background(), "test question", func(chunk string) error {
				result.WriteString(chunk)
				return nil
			})

			// Verify the request was made to the expected path
			assert.NoError(t, err)
			assert.Contains(t, capturedPath, tc.expectedPath)
			assert.Equal(t, "Hello", result.String())
		})
	}
}
