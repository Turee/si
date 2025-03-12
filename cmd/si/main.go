package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/ture/si/pkg/config"
	"github.com/ture/si/pkg/llm"
)

// CLI represents the command line interface
var CLI struct {
	// Global flags
	ConfigPath string   `name:"config" help:"Path to config file" type:"path"`
	Debug      bool     `name:"debug" help:"Enable debug mode"`
	Version    bool     `name:"version" help:"Show version information"`
	NoStream   bool     `name:"no-stream" help:"Disable streaming responses"`
	Question   []string `arg:"" optional:"" name:"question" help:"Question to ask the LLM"`
}

// For testing purposes, we can override these functions
var (
	osExit         = os.Exit
	loadConfigFunc = config.LoadConfig
)

func main() {
	// Parse command line arguments
	kongCtx := kong.Parse(&CLI,
		kong.Name("si"),
		kong.Description("A command line tool to interact with LLMs"),
		kong.UsageOnError(),
		kong.DefaultEnvars("SI"),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
			Summary: true,
		}),
	)

	// Handle version flag
	if CLI.Version {
		fmt.Println("si version 0.1.0")
		return
	}

	// If no question is provided, show help
	if len(CLI.Question) == 0 {
		kongCtx.PrintUsage(false)
		return
	}

	// Load configuration
	configPath := CLI.ConfigPath
	cfg, err := loadConfigFunc(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Configuration file not found. Please create a configuration file at ~/.config/si.yaml")
			fmt.Println("Example configuration:")
			fmt.Println("```yaml")
			fmt.Println("llm:")
			fmt.Println("  openai:")
			fmt.Println("    # Base URL for the OpenAI API. You can specify:")
			fmt.Println("    # - Full endpoint URL: https://api.openai.com/v1/chat/completions")
			fmt.Println("    # - Base API URL: https://api.openai.com/v1")
			fmt.Println("    # - For Azure, use your Azure OpenAI resource endpoint")
			fmt.Println("    base_url: https://api.openai.com/v1")
			fmt.Println("    # Your OpenAI API key or Azure API key")
			fmt.Println("    api_key: your-api-key")
			fmt.Println("    # Model name to use (default: gpt-4)")
			fmt.Println("    model_name: gpt-4")
			fmt.Println("    # For Azure OpenAI, specify your deployment name")
			fmt.Println("    azure_deployment_name: optional-azure-deployment-name")
			fmt.Println("```")
			osExit(1)
		}
		fmt.Printf("Error loading configuration: %v\n", err)
		osExit(1)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		fmt.Printf("Invalid configuration: %v\n", err)
		osExit(1)
	}

	// Handle the question
	if err := handleQuestion(cfg, CLI.Question); err != nil {
		fmt.Printf("Error: %v\n", err)
		osExit(1)
	}
}

func handleQuestion(cfg *config.Config, question []string) error {
	// Create LLM provider
	provider, err := llm.NewProvider(cfg)
	if err != nil {
		return fmt.Errorf("error creating LLM provider: %w", err)
	}

	// Join all question parts into a single string
	questionStr := strings.Join(question, " ")

	// If streaming is disabled, use the non-streaming API
	if CLI.NoStream {
		// Ask the question
		answer, err := provider.Ask(context.Background(), questionStr)
		if err != nil {
			if err == llm.ErrNotImplemented {
				return fmt.Errorf("this feature is not yet implemented")
			}
			return fmt.Errorf("error asking question: %w", err)
		}

		// Print the answer
		fmt.Println(answer)
		return nil
	}

	// Use streaming API
	err = provider.AskStream(context.Background(), questionStr, func(chunk string) error {
		// Print the chunk without a newline to create a streaming effect
		fmt.Print(chunk)
		// Flush stdout to ensure the chunk is displayed immediately
		return nil
	})

	if err != nil {
		if err == llm.ErrNotImplemented {
			return fmt.Errorf("this feature is not yet implemented")
		}
		return fmt.Errorf("error asking question: %w", err)
	}

	// Print a newline at the end of the response
	fmt.Println()
	return nil
}
