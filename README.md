# si - Command Line LLM Assistant

`si` is a command line tool written in Go that allows you to harness the power of Large Language Models (LLMs) directly from your terminal. Ask questions, generate commands, and get AI assistance without leaving your workflow.

![Version](https://img.shields.io/badge/version-0.1.0-blue)
![Go](https://img.shields.io/badge/go-1.23-blue)

## Features

- **Simple Querying**: Ask questions and get concise answers
- **Command Generation**: Generate complex shell commands on the fly
- **Streaming Responses**: See responses as they're generated (with option to disable)
- **Configurable**: Use different LLM providers with customizable settings
- **Pipe Support**: Pipe content into `si` for context-aware responses

## Installation

### Prerequisites

- Go 1.23 or higher

### Building from Source

```bash
git clone https://github.com/Turee/si.git
cd si
go build -o bin/si cmd/si/main.go
```

### Adding to PATH

```bash
# Add to your shell configuration file (.bashrc, .zshrc, etc.)
export PATH=$PATH:/path/to/si/bin
```

## Usage Examples

### Simple Questions

```bash
si what is the capital of France?
# Output: Capital of France is Paris
```

### Command Generation

```bash
si write a ffmpeg command that encodes all video files in current directory as h265
# Output: ffmpeg -i *.mp4 -c:v libx265 -crf 28 -c:a aac -b:a 128k output_%03d.mp4
```

### Piping Content

```bash
cat error_log.txt | si explain this error
```

## Configuration

`si` is configured via a YAML file located at `~/.config/si.yaml`.

### Sample Configuration

```yaml
llm:
  openai:
    # Base URL for the OpenAI API. You can specify:
    # - Full endpoint URL: https://api.openai.com/v1/chat/completions
    # - Base API URL: https://api.openai.com/v1
    # - For Azure, use your Azure OpenAI resource endpoint
    base_url: https://api.openai.com/v1

    # Your OpenAI API key or Azure API key
    api_key: your-api-key

    # Model name to use (default: gpt-4)
    # model_name: gpt-4

    # For Azure OpenAI, specify your deployment name
    # azure_deployment_name: optional-azure-deployment-name
```

## Command Line Options

| Flag          | Description                                      |
| ------------- | ------------------------------------------------ |
| `--config`    | Path to config file (default: ~/.config/si.yaml) |
| `--debug`     | Enable debug mode                                |
| `--version`   | Show version information                         |
| `--no-stream` | Disable streaming responses                      |

## Development

### Project Structure

- `cmd/si/` - Main application code
- `pkg/config/` - Configuration handling
- `pkg/llm/` - LLM provider implementations

### Running Tests

```bash
go test ./...
```

## License

[MIT License]

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Acknowledgments

- Built with [Kong CLI](https://github.com/alecthomas/kong)
- Supports OpenAI and Azure OpenAI APIs
