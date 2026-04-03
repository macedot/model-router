# Model Router

AI model proxy that accepts OpenAI or Anthropic format requests and forwards them to external providers with format conversion.

## ✨ Features

- OpenAI `/v1/chat/completions` endpoint
- Anthropic `/v1/messages` endpoint
- Per-target format conversion (OpenAI ↔ Anthropic)
- Model name mapping (internal name → provider model name)
- Fallback strategy across multiple provider models
- Configurable via JSON with `env:VAR` environment variable support
- Optional `.env` file loading (current dir → config dir)
- Config search order: current directory → `~/.model-router/` → `$PATH`

## 🚀 Quick Start

```bash
# Build
make build

# Run
./model-router
```

## ⚙️ Configuration

Config file is searched in this order:

1. `config.json` in the current working directory
2. `~/.model-router/config.json`
3. `config.json` anywhere on `$PATH`

```json
{
  "port": 12345,
  "models": [
    {
      "name": "my-model",
      "strategy": "fallback",
      "externals": [
        {
          "name": "provider-model-name",
          "url": "https://api.provider.com/v1/chat/completions",
          "api_key": "env:API_KEY",
          "format": "openai"
        }
      ]
    }
  ]
}
```

### Environment Variables

API keys can be loaded via `.env` file (searched in current directory, then config directory) or set in the environment.

Use `env:VAR_NAME` in config to reference environment variables.

## 🌐 API Endpoints

- `GET /models` - List available models
- `POST /v1/chat/completions` - OpenAI-compatible chat
- `POST /v1/messages` - Anthropic-compatible messages

## 🛠️ Development

```bash
make test    # Run go test ./...
make build   # Build binary
make install # Install as systemd user service
```

## 📄 License

GPLv3
