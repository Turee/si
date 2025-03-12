package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	LLM LLMConfig `yaml:"llm"`
}

// LLMConfig represents the configuration for LLM providers
type LLMConfig struct {
	OpenAI OpenAIConfig `yaml:"openai"`
}

// OpenAIConfig represents the configuration for OpenAI
type OpenAIConfig struct {
	BaseURL             string `yaml:"base_url"`
	APIKey              string `yaml:"api_key"`
	ModelName           string `yaml:"model_name,omitempty"`
	AzureDeploymentName string `yaml:"azure_deployment_name,omitempty"`
}

// DefaultConfigPath returns the default path for the configuration file
func DefaultConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(homeDir, ".config", "si.yaml")
}

// LoadConfig loads the configuration from the specified path
func LoadConfig(path string) (*Config, error) {
	if path == "" {
		path = DefaultConfigPath()
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Check if OpenAI API key is provided
	if c.LLM.OpenAI.APIKey == "" {
		return fmt.Errorf("OpenAI API key is required")
	}

	return nil
}
