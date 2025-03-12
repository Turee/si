package main

import (
	"fmt"
	"os"

	"github.com/ture/si/pkg/config"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig("")
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Configuration file not found. Please create a configuration file at ~/.config/si.yaml")
			fmt.Println("Example configuration:")
			fmt.Println("```yaml")
			fmt.Println("llm:")
			fmt.Println("  openai:")
			fmt.Println("    base_url: https://api.openai.com/v1")
			fmt.Println("    api_key: your-api-key")
			fmt.Println("    azure_deployment_name: optional-azure-deployment-name")
			fmt.Println("```")
			os.Exit(1)
		}
		fmt.Printf("Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		fmt.Printf("Invalid configuration: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Configuration loaded successfully!")
	// TODO: Implement the rest of the application
} 