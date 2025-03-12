package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ture/si/pkg/config"
)

// TestVersionFlag tests the --version flag
func TestVersionFlag(t *testing.T) {
	// Save original os.Args and restore after test
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Set up test args
	os.Args = []string{"si", "--version"}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Reset after test
	defer func() { os.Stdout = oldStdout }()

	// Call main with exit handling
	exitCalled := false
	oldOsExit := osExit
	defer func() { osExit = oldOsExit }()
	osExit = func(code int) {
		exitCalled = true
	}

	main()

	// Read captured output
	w.Close()
	var buf bytes.Buffer
	_, err := buf.ReadFrom(r)
	require.NoError(t, err)

	// Check output
	assert.Contains(t, buf.String(), "si version")
	assert.False(t, exitCalled, "os.Exit should not have been called")
}

// TestHelpFlag tests the --help flag
func TestHelpFlag(t *testing.T) {
	if os.Getenv("TEST_MAIN_HELP") == "1" {
		// When running in subprocess mode, actually run main
		oldArgs := os.Args
		os.Args = []string{"si", "--help"}
		defer func() { os.Args = oldArgs }()
		main()
		return
	}

	// Run the test in a subprocess to capture the help output
	cmd := exec.Command(os.Args[0], "-test.run=TestHelpFlag")
	cmd.Env = append(os.Environ(), "TEST_MAIN_HELP=1")
	output, err := cmd.CombinedOutput()

	// Kong's help exits with code 0, so we expect no error
	require.NoError(t, err)

	// Check output contains expected help text
	outputStr := string(output)
	assert.Contains(t, outputStr, "Usage: si")
	assert.Contains(t, outputStr, "--help")
	assert.Contains(t, outputStr, "--version")
	assert.Contains(t, outputStr, "Question to ask the LLM")
}

// TestQuestionHandling tests the question handling functionality
func TestQuestionHandling(t *testing.T) {
	// Create a mock config for testing
	cfg := &config.Config{
		LLM: config.LLMConfig{
			OpenAI: config.OpenAIConfig{
				BaseURL: "https://api.openai.com/v1",
				APIKey:  "test-api-key",
			},
		},
	}

	// Test with a simple question
	question := []string{"what", "is", "the", "capital", "of", "France?"}
	err := handleQuestion(cfg, question)

	// Since we haven't implemented the actual LLM functionality,
	// we expect an ErrNotImplemented error
	assert.Error(t, err)
	assert.Equal(t, "this feature is not yet implemented", err.Error())
}

// TestNoArgsShowsHelp tests that running without arguments shows help
func TestNoArgsShowsHelp(t *testing.T) {
	if os.Getenv("TEST_MAIN_NO_ARGS") == "1" {
		// When running in subprocess mode, actually run main
		oldArgs := os.Args
		os.Args = []string{"si"}
		defer func() { os.Args = oldArgs }()
		main()
		return
	}

	// Run the test in a subprocess to capture the output
	cmd := exec.Command(os.Args[0], "-test.run=TestNoArgsShowsHelp")
	cmd.Env = append(os.Environ(), "TEST_MAIN_NO_ARGS=1")
	output, err := cmd.CombinedOutput()

	// Kong's help exits with code 0, so we expect no error
	require.NoError(t, err)

	// Check output contains expected help text
	outputStr := string(output)
	assert.Contains(t, outputStr, "Usage: si")
}

// TestConfigNotFound tests the behavior when config file is not found
func TestConfigNotFound(t *testing.T) {
	// Save original os.Args and restore after test
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Set up test args with a non-existent config file
	tempDir := t.TempDir()
	nonExistentConfig := filepath.Join(tempDir, "non-existent-config.yaml")
	os.Args = []string{"si", "--config", nonExistentConfig, "test", "question"}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Reset after test
	defer func() { os.Stdout = oldStdout }()

	// Call main with exit handling
	exitCalled := false
	exitCode := 0
	oldOsExit := osExit
	defer func() { osExit = oldOsExit }()
	osExit = func(code int) {
		exitCalled = true
		exitCode = code
		panic("os.Exit called") // Use panic to stop execution
	}

	// Run main and recover from the expected panic
	func() {
		defer func() {
			recover() // Recover from the panic caused by our mock osExit
		}()
		main()
	}()

	// Read captured output
	w.Close()
	var buf bytes.Buffer
	_, err := buf.ReadFrom(r)
	require.NoError(t, err)

	// Check output and exit behavior
	assert.True(t, exitCalled, "os.Exit should have been called")
	assert.Equal(t, 1, exitCode, "Exit code should be 1")
	assert.Contains(t, buf.String(), "Error loading configuration")
	assert.Contains(t, buf.String(), "failed to read config file")
}

// TestInvalidConfig tests the behavior when config is invalid
func TestInvalidConfig(t *testing.T) {
	// Create a temporary config file with invalid content
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "invalid-config.yaml")
	err := os.WriteFile(configPath, []byte("invalid: yaml: content"), 0644)
	require.NoError(t, err)

	// Save original LoadConfig function and restore after test
	oldLoadConfig := loadConfigFunc
	defer func() { loadConfigFunc = oldLoadConfig }()

	// Mock LoadConfig to return a config with validation error
	loadConfigFunc = func(path string) (*config.Config, error) {
		return &config.Config{}, nil // Return empty config that will fail validation
	}

	// Save original os.Args and restore after test
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Set up test args
	os.Args = []string{"si", "--config", configPath, "test", "question"}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Reset after test
	defer func() { os.Stdout = oldStdout }()

	// Call main with exit handling
	exitCalled := false
	exitCode := 0
	oldOsExit := osExit
	defer func() { osExit = oldOsExit }()
	osExit = func(code int) {
		exitCalled = true
		exitCode = code
		panic("os.Exit called") // Use panic to stop execution
	}

	// Run main and recover from the expected panic
	func() {
		defer func() {
			recover() // Recover from the panic caused by our mock osExit
		}()
		main()
	}()

	// Read captured output
	w.Close()
	var buf bytes.Buffer
	_, err = buf.ReadFrom(r)
	require.NoError(t, err)

	// Check output and exit behavior
	assert.True(t, exitCalled, "os.Exit should have been called")
	assert.Equal(t, 1, exitCode, "Exit code should be 1")
	assert.Contains(t, buf.String(), "Invalid configuration")
}
