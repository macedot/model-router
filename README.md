# Model Router

AI model proxy that accepts OpenAI or Anthropic format requests and forwards them to external providers with format conversion.

## Features

- OpenAI `/v1/chat/completions` endpoint
- Anthropic `/v1/messages` endpoint
- Per-target format conversion (OpenAI ↔ Anthropic)
- Model name mapping (internal name → provider model name)
- Fallback strategy across multiple provider models
- Streaming support (OpenAI `stream: true`)
- Reusable provider definitions referenced by ID
- Configurable via JSON with `${VAR}` environment variable expansion
- Optional `.env` file loading (current dir → config dir)
- Config search order: current directory → `~/.config/model-router/`

## System Requirements

**Development:** Go 1.26+

**Runtime:** No dependencies — standalone static binary.

## Quick Start

### Run with Docker

```bash
cp config.example.json config.json  # if you don't have one yet
docker compose up -d
```

## Configuration

### Config file search order

1. `config.json` in the current working directory
2. `~/.config/model-router/config.json`

### Config structure

```json
{
  "port": 12345,
  "providers": [
    {
      "id": "my-provider",
      "name": "provider-model-name",
      "url": "https://api.provider.com/v1/chat/completions",
      "api_key": "${API_KEY}",
      "format": "openai"
    }
  ],
  "models": [
    {
      "name": "my-model",
      "request_format": "openai",
      "strategy": "fallback",
      "retry_delay_secs": 1,
      "externals": ["my-provider"]
    }
  ]
}
```

Externals can reference providers by ID (preferred) or use inline objects (legacy):

```json
"externals": ["my-provider", "another-provider"]
```

```json
"externals": [{ "name": "legacy", "url": "...", "api_key": "...", "format": "openai" }]
```

### Config fields

**Top-level:**

| Field       | Type   | Default | Description                              |
| ----------- | ------ | ------- | ---------------------------------------- |
| `port`      | uint16 | 12345   | Listen port                              |
| `providers` | array  | `[]`    | Reusable provider definitions            |
| `models`    | array  | `[]`    | Internal model definitions               |

**Provider:**

| Field    | Type   | Description                          |
| -------- | ------ | ------------------------------------ |
| `id`     | string | Unique identifier (required)         |
| `name`   | string | Model name sent to the provider      |
| `url`    | string | Provider API endpoint URL            |
| `api_key`| string | API key (supports `${VAR}` expansion)|
| `format` | string | `openai` or `anthropic`              |

**Model:**

| Field              | Type          | Default | Description                                                |
| ------------------ | ------------- | ------- | ---------------------------------------------------------- |
| `name`             | string        | —       | Internal model name used by clients                        |
| `request_format`   | string        | —       | Client-facing format: `openai` or `anthropic`              |
| `strategy`         | string        | —       | Routing strategy: `fallback`                               |
| `retry_delay_secs` | uint32        | 0       | Seconds to wait before retrying the next external provider |
| `externals`        | array         | —       | Provider IDs (strings) or inline external objects          |

### Environment variables

In config, reference them with `${VAR_NAME}`:

```json
"api_key": "${API_KEY_MINIMAX}"
```

Optionally place a `.env` file in the current directory or alongside the config file.

## API Endpoints

### `GET /models`

Returns the registered internal models with their configuration (API keys and URLs are redacted).

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

Accepts OpenAI-format chat completion requests. Routes to the internal model name specified in `model`. Supports streaming via `stream: true`.

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

## Docker

```bash
# Create .env with your API keys
cp config.example.json config.json  # if needed
# edit .env with your keys

# Start
docker compose up -d

# Stop
docker compose down
```

Pre-built images are available at `ghcr.io/macedot/model-router`.

## Development

```bash
make build            # Build binary locally
make run              # Build and run
make test             # Run go test ./...
make clean            # Remove binary
make release VERSION=1.0.0  # Create git tag and GitHub release
```

## License

AGPL-3.0
