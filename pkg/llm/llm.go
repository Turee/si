package llm

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/Turee/si/pkg/config"
)

// ErrNotImplemented is returned when a feature is not yet implemented
var ErrNotImplemented = errors.New("not implemented")

// Provider defines the interface for LLM providers
type Provider interface {
	// Ask sends a question to the LLM and returns the response
	Ask(ctx context.Context, question string) (string, error)

	// AskStream sends a question to the LLM and streams the response
	AskStream(ctx context.Context, question string, callback func(chunk string) error) error
}

// NewProvider creates a new LLM provider based on the configuration
func NewProvider(cfg *config.Config) (Provider, error) {
	// For now, we only support OpenAI
	return NewOpenAIProvider(&cfg.LLM.OpenAI)
}

// NewOpenAIProvider creates a new OpenAI provider
func NewOpenAIProvider(cfg *config.OpenAIConfig) (Provider, error) {
	return &openAIProvider{
		cfg:    cfg,
		client: &http.Client{},
	}, nil
}

// openAIProvider implements the Provider interface for OpenAI
type openAIProvider struct {
	cfg    *config.OpenAIConfig
	client *http.Client
}

// OpenAI API request and response structures
type openAIRequest struct {
	Model       string    `json:"model"`
	Messages    []message `json:"messages"`
	Stream      bool      `json:"stream"`
	Temperature float64   `json:"temperature,omitempty"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []choice `json:"choices"`
}

type choice struct {
	Index        int     `json:"index"`
	Message      message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

// Streaming response structures
type streamResponse struct {
	ID      string         `json:"id"`
	Object  string         `json:"object"`
	Created int64          `json:"created"`
	Model   string         `json:"model"`
	Choices []streamChoice `json:"choices"`
}

type streamChoice struct {
	Index        int         `json:"index"`
	Delta        streamDelta `json:"delta"`
	FinishReason string      `json:"finish_reason"`
}

type streamDelta struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
}

// Ask implements the Provider interface
func (p *openAIProvider) Ask(ctx context.Context, question string) (string, error) {
	var result strings.Builder

	err := p.AskStream(ctx, question, func(chunk string) error {
		result.WriteString(chunk)
		return nil
	})

	if err != nil {
		return "", err
	}

	return result.String(), nil
}

// AskStream implements the Provider interface for streaming responses
func (p *openAIProvider) AskStream(ctx context.Context, question string, callback func(chunk string) error) error {
	// Determine the API endpoint
	baseURL := p.cfg.BaseURL
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}

	// Determine the model to use
	model := p.cfg.ModelName
	if model == "" {
		model = "gpt-4"
	}

	// Create the request
	reqBody := openAIRequest{
		Model: model,
		Messages: []message{
			{
				Role:    "system",
				Content: "You are an AI assistant being used from a terminal. Provide concise, direct responses optimized for command-line viewing. Prioritize brevity and clarity. Use markdown formatting when helpful for readability. Avoid unnecessary pleasantries or verbose explanations unless specifically requested.",
			},
			{
				Role:    "user",
				Content: question,
			},
		},
		Stream: true,
	}

	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Determine the API endpoint based on whether we're using Azure or standard OpenAI
	var endpoint string
	if p.cfg.AzureDeploymentName != "" {
		// Azure OpenAI endpoint format
		if !strings.HasSuffix(baseURL, "/") {
			baseURL += "/"
		}
		endpoint = fmt.Sprintf("%sopenai/deployments/%s/chat/completions?api-version=2024-12-01-preview",
			baseURL, p.cfg.AzureDeploymentName)
	} else {
		// Standard OpenAI endpoint
		// If the user provided a complete URL including the endpoint, use it directly
		if strings.Contains(baseURL, "/chat/completions") {
			endpoint = baseURL
		} else {
			// Otherwise, ensure the URL doesn't have a trailing slash and add the endpoint
			baseURL = strings.TrimSuffix(baseURL, "/")
			endpoint = fmt.Sprintf("%s/chat/completions", baseURL)
		}
	}

	// Create the HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, strings.NewReader(string(reqJSON)))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// Set the API key header based on whether we're using Azure or not
	if p.cfg.AzureDeploymentName != "" {
		req.Header.Set("api-key", p.cfg.APIKey)
	} else {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.cfg.APIKey))
	}

	// Send the request
	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check for errors
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Process the streaming response
	reader := bufio.NewReader(resp.Body)

	for {
		// Read a line from the response
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("error reading response: %w", err)
		}

		// Skip empty lines and "data: [DONE]"
		line = strings.TrimSpace(line)
		if line == "" || line == "data: [DONE]" {
			continue
		}

		// Remove "data: " prefix
		if strings.HasPrefix(line, "data: ") {
			line = strings.TrimPrefix(line, "data: ")
		} else {
			continue // Skip non-data lines
		}

		// Parse the JSON
		var streamResp streamResponse
		if err := json.Unmarshal([]byte(line), &streamResp); err != nil {
			return fmt.Errorf("error parsing response: %w", err)
		}

		// Process the choices
		for _, choice := range streamResp.Choices {
			if choice.Delta.Content != "" {
				if err := callback(choice.Delta.Content); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
