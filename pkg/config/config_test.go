package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")
	
	configContent := `llm:
  openai:
    base_url: https://api.openai.com/v1
    api_key: test-api-key
    azure_deployment_name: test-deployment
`
	
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}
	
	// Load the config
	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	
	// Verify the config values
	if config.LLM.OpenAI.BaseURL != "https://api.openai.com/v1" {
		t.Errorf("Expected BaseURL to be 'https://api.openai.com/v1', got '%s'", config.LLM.OpenAI.BaseURL)
	}
	
	if config.LLM.OpenAI.APIKey != "test-api-key" {
		t.Errorf("Expected APIKey to be 'test-api-key', got '%s'", config.LLM.OpenAI.APIKey)
	}
	
	if config.LLM.OpenAI.AzureDeploymentName != "test-deployment" {
		t.Errorf("Expected AzureDeploymentName to be 'test-deployment', got '%s'", config.LLM.OpenAI.AzureDeploymentName)
	}
}

func TestValidate(t *testing.T) {
	// Test valid config
	validConfig := &Config{
		LLM: LLMConfig{
			OpenAI: OpenAIConfig{
				BaseURL: "https://api.openai.com/v1",
				APIKey:  "test-api-key",
			},
		},
	}
	
	if err := validConfig.Validate(); err != nil {
		t.Errorf("Expected valid config to pass validation, got error: %v", err)
	}
	
	// Test invalid config (missing API key)
	invalidConfig := &Config{
		LLM: LLMConfig{
			OpenAI: OpenAIConfig{
				BaseURL: "https://api.openai.com/v1",
				APIKey:  "",
			},
		},
	}
	
	if err := invalidConfig.Validate(); err == nil {
		t.Error("Expected invalid config to fail validation, but it passed")
	}
} 