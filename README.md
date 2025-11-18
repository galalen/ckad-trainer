# CKAD Trainer

A CLI tool for generating Certified Kubernetes Application Developer (CKAD) practice scenarios using LLMs.

## Features

- Generate CKAD practice scenarios complete with validation scripts
- Validate your solutions against the generated validation scripts
- Supports multiple LLM providers (OpenAI and OpenRouter)

## Installation

### Prerequisites

- Go 1.21 or higher
- A local Kubernetes cluster — any of the following works:
  - **Minikube**
  - **kind** (Kubernetes in Docker)
  - **Docker Desktop Kubernetes**
  - **Any standard local Kubernetes cluster**
- API key for **OpenAI** or **OpenRouter**

### Building from Source

```bash
git clone https://github.com/galalen/ckad-trainer.git
cd ckad-trainer
go build -o ckad-trainer main.go
```

## Usage

### Environment Variables

You can set the following environment variables instead of using command-line flags:

```bash
export OPENAI_API_KEY="your-openai-api-key"
export OPENROUTER_API_KEY="your-openrouter-api-key"
```

### Generating Scenarios

```bash
# Generate scenarios using OpenRouter
./ckad-trainer generate --provider openrouter --api-key your_key --count 3

# Generate 5 scenarios using OpenAI
./ckad-trainer generate --provider openai --api-key your_key --count 5

# Customize model and parameters
./ckad-trainer generate --provider openai --model gpt-4 --temperature 0.8 --count 10
```

### Validating Scenarios

```bash
# Validate a specific scenario
./ckad-trainer validate --scenario 1

# Validate all scenarios
./ckad-trainer validate --all

# Validate scenarios in a custom directory
./ckad-trainer validate --all --dir ./my-scenarios
```

## License

MIT
