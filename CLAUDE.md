# CLAUDE.md

Guide for Claude Code (claude.ai/code) working this repo.

## Project Overview

model-router = AI model proxy. Unified API (OpenAI/Anthropic format), forwards to external providers. Fallback routing — first provider fail, try next.

## Build & Development Commands

**Go 1.26+** required.

```bash
make build              # Build ./model-router binary (version embedded via -ldflags)
make run                # Build + run
make test               # Run all tests (go test ./...)
make clean              # Remove binary
make release VERSION=x  # Tag release, push, create GitHub release (CI handles binaries)
```

Run single test:
```bash
go test ./handlers/ -run TestOpenAIHandler_Success
```

No linter. No external deps — stdlib only. AGPL-3.0.

## Architecture

```
main.go                  Entry point, HTTP server, logging/recovery middleware, graceful shutdown
  |
  +-- config/            Config loading from JSON + env var expansion (${VAR}) + .env files
  +-- handlers/          HTTP handlers for each endpoint
  |     openai.go        POST /v1/chat/completions
  |     anthropic.go     POST /v1/messages
  |     models.go        GET  /models (redacts URLs and API keys from response)
  |     decoder.go       Reads raw body into map[string]interface{} + validation envelope
  |     errors.go        Sanitized error response helper (logs full error, sends clientMsg only)
  +-- services/          Core business logic
  |     registry.go      Model lookup by name (returns defensive copies)
  |     forwarder.go     Pure HTTP transport: Forward() and ForwardStream() accept raw []byte
  |     converter.go     Map-based format conversion (PrepareRequest)
  +-- models/            Data types (InternalModel, ExternalModel, Provider, RequestEnvelope)
```

### Request Flow

1. Client sends request to `/v1/chat/completions` or `/v1/messages`
2. `decoder.readBody()` reads raw body into `map[string]interface{}` (generic passthrough) + `RequestEnvelope` (validation)
3. Handler validates `model` + `messages` from envelope, looks up model name in registry
4. For each external provider: `converter.PrepareRequest(body, targetName, sourceFormat, targetFormat)` sets model name, applies cross-format conversion if needed
5. `forwarder.Forward()` or `ForwardStream()` sends raw bytes upstream with auth headers (`Authorization: Bearer` for OpenAI, `x-api-key` + `anthropic-version` for Anthropic)
6. Raw response bytes returned to client — no response parsing
7. On failure with `strategy: "fallback"`, waits `retry_delay_secs` then tries next provider

### Generic Passthrough Design

Router uses NO typed request structs. Decodes into `map[string]interface{}` — all client fields pass through unchanged. Only `model` name touched. Cross-format conversion in `converter.go`: known field renames (`stop` ↔ `stop_sequences`), drops source-only fields with warning (`frequency_penalty` has no Anthropic equivalent), everything else passes through as-is.

### Key Design Patterns

- **Handler closures**: `NewOpenAIHandler(registry, forwarder)` returns `http.HandlerFunc`, captures deps as closure variables. Not struct methods.
- **Interface-based DI**: Handlers depend on `RegistryReader` interface, not concrete `ModelRegistry`.
- **Defensive copying**: `ModelRegistry.Get()`/`List()` return deep copies — tests verify mutation of returned values doesn't affect registry state.
- **Raw bytes pipeline**: Forwarder accepts `[]byte`, knows nothing about request formats. All format logic in `converter.go`.
- **Zero dependencies**: No third-party packages. All stdlib.

## Configuration

Config search order: `./config.json` → `~/.config/model-router/config.json` → defaults (port 12345, no models).

Providers defined once in top-level `"providers"` section, referenced by ID in model externals. Old inline format (externals as objects) still supported.

```json
{
    "port": 12345,
    "providers": [
        {"id": "zai", "name": "glm-5.1", "url": "...", "api_key": "${API_KEY_ZAI}", "format": "openai"}
    ],
    "models": [
        {"name": "coding", "strategy": "fallback", "externals": ["zai", "minimax"]}
    ]
}
```

API keys support `${VAR_NAME}` syntax (expanded via `os.Expand`). Optional `.env` file loaded from cwd or config file directory.

## Deployment

- **Docker**: Multi-stage build (Go Alpine → distroless/static). `docker compose up -d`.
- **CI**: GitHub Actions build Docker images to ghcr.io, cross-compiled binaries for linux/darwin on amd64/arm64.
- **Dependabot**: Daily checks for Go modules + Docker base images.

## Testing

All tests use stdlib `testing` + `httptest` — no external frameworks. Tests spin up `httptest.NewServer` mock upstreams for real HTTP round-trips. Key areas:

- Converter: same-format passthrough, cross-format field renames, source-only field dropping
- Forwarder: success/error/timeout for Forward + ForwardStream, response size limits
- Handlers: validation errors, fallback logic, streaming, cross-format forwarding
- Registry: defensive copy verification, missing model lookup
- Config: env expansion, provider resolution, file discovery