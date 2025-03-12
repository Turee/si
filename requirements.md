# si

## Usage

si is a command line command line application written in go that enables the user to invoke the power of LLM's from the command line.

Usage examples:

Simple questions:

```bash
user> si what is the capital of france?
Capital of france is paris
```

Quick command generation

```bash
user> Write a ffmpeg command that encodes all video files in current directory as h265
ffmpeg -i *.mp4 -c:v libx265 -crf 28 -c:a aac -b:a 128k output_%03d.mp4
```


## Configuration

si is configured by an yaml file located in ~/.config/si.yaml

example config file:

```yaml
llm:
  openai:
    base_url:
    api_key:
    azure_deployment_name: # Optional
```