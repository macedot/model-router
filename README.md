# Model Router

AI model proxy that accepts OpenAI or Anthropic format requests and forwards them to external providers with format conversion.

## ✨ Features

- OpenAI `/v1/chat/completions` endpoint
- Anthropic `/v1/messages` endpoint
- Per-target format conversion (OpenAI ↔ Anthropic)
- Model name mapping (internal name → provider model name)
- Fallback strategy across multiple provider models
- Streaming support (OpenAI `stream: true`)
- Configurable via JSON with `env:VAR` environment variable support
- Optional `.env` file loading (current dir → config dir)
- Config search order: current directory → `~/.config/model-router/`

## 🔧 System Requirements

**Development:** Go 1.26+

**Runtime:** No dependencies — standalone static binary.

## 🚀 Quick Start

### Run with Docker

```bash
cp config.example.json config.json  # if you don't have one yet
docker compose up -d
```

## ⚙️ Configuration

### Config file search order

1. `config.json` in the current working directory
2. `~/.config/model-router/config.json`

### Config structure

```json
{
  "port": 12345,
  "models": [
    {
      "name": "my-model",
      "request_format": "openai",
      "strategy": "fallback",
      "retry_delay_secs": 1,
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

| Field              | Type   | Default | Description                                                |
| ------------------ | ------ | ------- | ---------------------------------------------------------- |
| `port`             | uint16 | 12345   | Listen port                                                |
| `request_format`   | string | —       | Client-facing format: `openai` or `anthropic`              |
| `strategy`         | string | —       | Routing strategy: `fallback`                               |
| `retry_delay_secs` | uint32 | 0       | Seconds to wait before retrying the next external provider |
| `externals`        | array  | —       | List of external providers                                 |

### Environment variables

In config, reference them with `env:VAR_NAME`:

```json
"api_key": "env:API_KEY_MINIMAX"
```

## 🌐 API Endpoints

### `GET /models`

Returns the registered internal models with their configuration.

**Response:**

```json
{
  "models": [
    {
      "name": "coding",
      "request_format": "",
      "strategy": "fallback",
      "retry_delay_secs": 0,
      "externals": [
        { "name": "glm-5.1", "format": "openai" },
        { "name": "opencode-go/minimax-m2.7", "format": "anthropic" },
        { "name": "MiniMax-M2.7", "format": "anthropic" }
      ]
    }
  ]
}
```

### `POST /v1/chat/completions` (OpenAI-compatible)

Accepts OpenAI-format chat completion requests. Routes to the internal model name specified in `model`.

**Request:**

```json
{
  "model": "my-model",
  "messages": [
    { "role": "user", "content": "Hello" }
  ],
  "temperature": 0.7,
  "max_tokens": 1024,
  "stream": false
}
```

**Response** (non-streaming):

```json
{
  "id": "chatcmpl-xxx",
  "model": "my-model",
  "choices": [
    {
      "message": { "role": "assistant", "content": "Hi" },
      "finish_reason": "stop"
    }
  ]
}
```

### `POST /v1/messages` (Anthropic-compatible)

Accepts Anthropic-format messages requests.

**Request:**

```json
{
  "model": "my-model",
  "messages": [
    { "role": "user", "content": "Hello" }
  ],
  "max_tokens": 1024,
  "temperature": 0.7
}
```

**Response:**

```json
{
  "id": "msg_xxx",
  "model": "my-model",
  "type": "message",
  "role": "assistant",
  "content": [
    { "type": "text", "text": "Hi" }
  ],
  "stop_reason": "end_turn"
}
```

## 🐳 Docker

```bash
# Create .env with your API keys
cp config.example.json config.json  # if needed
# edit .env with your keys

# Start
docker compose up -d

# Stop
docker compose down
```

## 🛠️ Development

```bash
make build            # Build binary locally
make run              # Build and run
make test             # Run go test ./...
make clean            # Remove binary
make release VERSION=1.0.0  # Create git tag and GitHub release
```

## 📄 License

AGPL-3.0
