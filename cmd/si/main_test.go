package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/Turee/si/pkg/config"
	"github.com/Turee/si/pkg/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

// MockProvider is a mock implementation of the llm.Provider interface for testing
type MockProvider struct {
	AskResponse     string
	AskStreamChunks []string
	AskError        error
	AskStreamError  error
	QuestionAsked   string
}

// Ask implements the Provider interface
func (m *MockProvider) Ask(ctx context.Context, question string) (string, error) {
	m.QuestionAsked = question
	return m.AskResponse, m.AskError
}

// AskStream implements the Provider interface
func (m *MockProvider) AskStream(ctx context.Context, question string, callback func(chunk string) error) error {
	m.QuestionAsked = question
	if m.AskStreamError != nil {
		return m.AskStreamError
	}

	// If AskStreamChunks is empty but AskResponse is set, use that
	if len(m.AskStreamChunks) == 0 && m.AskResponse != "" {
		return callback(m.AskResponse)
	}

	// Send each chunk through the callback
	for _, chunk := range m.AskStreamChunks {
		if err := callback(chunk); err != nil {
			return err
		}
	}
	return nil
}

// MockNewProvider returns a mock provider for testing
func MockNewProvider(cfg *config.Config) (llm.Provider, error) {
	return &MockProvider{
		AskResponse: "This is a mock response from LLM",
		AskStreamChunks: []string{
			"This ", "is ", "a ", "mock ", "streamed ", "response ",
			"from ", "LLM", ".",
		},
	}, nil
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

	// Save original NewProvider and restore after test
	oldNewProvider := llm.NewProvider
	defer func() {
		llm.NewProvider = oldNewProvider
	}()

	// Create mock provider
	mockProvider := &MockProvider{
		AskResponse: "Paris is the capital of France.",
	}

	// Replace NewProvider with a function that returns our mock
	llm.NewProvider = func(cfg *config.Config) (llm.Provider, error) {
		return mockProvider, nil
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()

	// Test with a simple question
	question := []string{"what", "is", "the", "capital", "of", "France?"}
	err := handleQuestion(cfg, question, "")
	w.Close()

	// Read captured output
	var buf bytes.Buffer
	_, err2 := buf.ReadFrom(r)
	require.NoError(t, err2)

	// Verify results
	require.NoError(t, err)
	assert.Equal(t, "what is the capital of France?", mockProvider.QuestionAsked)
	assert.Contains(t, buf.String(), "Paris is the capital of France.")
}

// TestStdinInput tests the stdin input handling functionality
func TestStdinInput(t *testing.T) {
	// Save original stdin functions and restore after test
	oldStdinStat := stdinStat
	oldStdinRead := stdinRead
	defer func() {
		stdinStat = oldStdinStat
		stdinRead = oldStdinRead
	}()

	// Mock stdin stat to simulate piped input
	stdinStat = func() (os.FileInfo, error) {
		return mockFileInfo{mode: 0}, nil
	}

	// Mock stdin read to return test content
	testStdinContent := "This is test content from stdin"
	stdinRead = func() ([]byte, error) {
		return []byte(testStdinContent), nil
	}

	// Test reading from stdin
	content, err := checkStdin()
	require.NoError(t, err)
	assert.Equal(t, testStdinContent, content)

	// Create a mock config for testing
	cfg := &config.Config{
		LLM: config.LLMConfig{
			OpenAI: config.OpenAIConfig{
				BaseURL: "https://api.openai.com/v1",
				APIKey:  "test-api-key",
			},
		},
	}

	// Save original NewProvider and restore after test
	oldNewProvider := llm.NewProvider
	defer func() {
		llm.NewProvider = oldNewProvider
	}()

	// Create mock provider
	mockProvider := &MockProvider{
		AskResponse: "Response based on stdin content",
	}

	// Replace NewProvider with a function that returns our mock
	llm.NewProvider = func(cfg *config.Config) (llm.Provider, error) {
		return mockProvider, nil
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()

	// Test handling question with stdin content
	err = handleQuestion(cfg, []string{}, content)
	w.Close()

	// Read captured output
	var buf bytes.Buffer
	_, err2 := buf.ReadFrom(r)
	require.NoError(t, err2)

	// Verify results
	require.NoError(t, err)
	assert.Equal(t, testStdinContent, mockProvider.QuestionAsked)
	assert.Contains(t, buf.String(), "Response based on stdin content")
}

// Mock FileInfo implementation for testing
type mockFileInfo struct {
	mode os.FileMode
}

func (m mockFileInfo) Name() string       { return "stdin" }
func (m mockFileInfo) Size() int64        { return 0 }
func (m mockFileInfo) Mode() os.FileMode  { return m.mode }
func (m mockFileInfo) ModTime() time.Time { return time.Time{} }
func (m mockFileInfo) IsDir() bool        { return false }
func (m mockFileInfo) Sys() interface{}   { return nil }

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

// TestMainWithStdin tests the main function with stdin input
func TestMainWithStdin(t *testing.T) {
	if os.Getenv("TEST_MAIN_STDIN") == "1" {
		// When running in subprocess mode, set up mocks and run main

		// Mock the configuration loading
		oldLoadConfig := loadConfigFunc
		defer func() { loadConfigFunc = oldLoadConfig }()
		loadConfigFunc = func(path string) (*config.Config, error) {
			return &config.Config{
				LLM: config.LLMConfig{
					OpenAI: config.OpenAIConfig{
						BaseURL:   "https://api.openai.com/v1",
						APIKey:    "test-api-key",
						ModelName: "gpt-4",
					},
				},
			}, nil
		}

		// Mock stdin functions
		oldStdinStat := stdinStat
		oldStdinRead := stdinRead
		defer func() {
			stdinStat = oldStdinStat
			stdinRead = oldStdinRead
		}()

		// Mock stdin stat to simulate piped input
		stdinStat = func() (os.FileInfo, error) {
			return mockFileInfo{mode: 0}, nil // 0 means not a character device (pipe)
		}

		// Mock stdin read to return test content
		testStdinContent := "Test content from stdin"
		stdinRead = func() ([]byte, error) {
			return []byte(testStdinContent), nil
		}

		// Mock os.Exit to prevent actual exit
		oldOsExit := osExit
		defer func() { osExit = oldOsExit }()
		osExit = func(code int) {
			// Instead of panic, just write a message to stdout that can be checked later
			fmt.Printf("EXIT_CODE_%d\n", code)
			// Don't actually exit in tests
		}

		// Save original NewProvider and restore after test
		oldNewProvider := llm.NewProvider
		defer func() {
			llm.NewProvider = oldNewProvider
		}()

		// Mock LLM provider
		mockProvider := &MockProvider{
			AskResponse: "Test successful mock response",
		}
		llm.NewProvider = func(cfg *config.Config) (llm.Provider, error) {
			return mockProvider, nil
		}

		// Set up flag to indicate test is running in subprocess environment
		fmt.Println("SUBPROCESS_STARTED")

		// Set minimal args
		oldArgs := os.Args
		os.Args = []string{"si"}
		defer func() { os.Args = oldArgs }()

		main()

		// Indicate test completed successfully
		fmt.Println("SUBPROCESS_COMPLETED")

		return
	}

	// Run the test in a subprocess to isolate it
	cmd := exec.Command(os.Args[0], "-test.run=TestMainWithStdin")
	cmd.Env = append(os.Environ(), "TEST_MAIN_STDIN=1")
	output, _ := cmd.CombinedOutput()
	outputStr := string(output)

	// Check that the subprocess started and completed
	assert.Contains(t, outputStr, "SUBPROCESS_STARTED", "Subprocess did not start properly")
	assert.Contains(t, outputStr, "SUBPROCESS_COMPLETED", "Subprocess did not complete properly")

	// Make sure exit wasn't called with error
	assert.NotContains(t, outputStr, "EXIT_CODE_1", "os.Exit(1) was called")
}

// TestQuestionHandlingNoStream tests the question handling functionality with streaming disabled
func TestQuestionHandlingNoStream(t *testing.T) {
	// Save original CLI.NoStream and restore after test
	oldNoStream := CLI.NoStream
	defer func() { CLI.NoStream = oldNoStream }()

	// Enable NoStream option
	CLI.NoStream = true

	// Create a mock config for testing
	cfg := &config.Config{
		LLM: config.LLMConfig{
			OpenAI: config.OpenAIConfig{
				BaseURL: "https://api.openai.com/v1",
				APIKey:  "test-api-key",
			},
		},
	}

	// Save original NewProvider and restore after test
	oldNewProvider := llm.NewProvider
	defer func() {
		llm.NewProvider = oldNewProvider
	}()

	// Create mock provider
	mockProvider := &MockProvider{
		AskResponse: "Paris is the capital of France.",
	}

	// Replace NewProvider with a function that returns our mock
	llm.NewProvider = func(cfg *config.Config) (llm.Provider, error) {
		return mockProvider, nil
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()

	// Test with a simple question
	question := []string{"what", "is", "the", "capital", "of", "France?"}
	err := handleQuestion(cfg, question, "")
	w.Close()

	// Read captured output
	var buf bytes.Buffer
	_, err2 := buf.ReadFrom(r)
	require.NoError(t, err2)

	// Verify results
	require.NoError(t, err)
	assert.Equal(t, "what is the capital of France?", mockProvider.QuestionAsked)
	assert.Contains(t, buf.String(), "Paris is the capital of France.")
}
