# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

model-router is an AI model proxy server that exposes a unified API (OpenAI or Anthropic format) and transparently forwards requests to one or more external AI providers, performing automatic format conversion. It supports fallback routing — if the first provider fails, it tries the next one.

## Build & Development Commands

```bash
make build              # Build ./model-router binary (version embedded via -ldflags)
make run                # Build + run
make test               # Run all tests (go test ./...)
make clean              # Remove binary
make release VERSION=x  # Tag release, push, create GitHub release with cross-compiled binaries
```

Run a single test:
```bash
go test ./handlers/ -run TestOpenAIHandler_Success
```

No linter is configured. No external dependencies — the entire project uses only the Go standard library.

## Architecture

```
main.go                  Entry point, HTTP server, middleware, graceful shutdown
  |
  +-- config/            Config loading from JSON + env var expansion (${env:VAR}) + .env files
  +-- handlers/          HTTP handlers for each endpoint
  |     openai.go        POST /v1/chat/completions
  |     anthropic.go     POST /v1/messages
  |     models.go        GET  /models (redacts URLs and API keys from response)
  |     decoder.go       JSON body decoder with 50MB limit
  |     errors.go        Error response helper
  +-- services/          Core business logic
  |     registry.go      Model lookup by name (returns defensive copies)
  |     forwarder.go     HTTP forwarding to external providers, streaming support
  |     converter.go     OpenAI → Anthropic request format conversion
  +-- models/            Data types (InternalModel, ExternalModel, RequestFormat)
```

### Request Flow

1. Client sends request to `/v1/chat/completions` or `/v1/messages`
2. Handler decodes body, validates required fields, looks up model name in registry
3. For each external provider configured for that model:
   - If format differs from request format, converts via `converter.ToAnthropic()`
   - Sets correct auth headers (`Authorization: Bearer` for OpenAI, `x-api-key` for Anthropic)
   - Forwards to upstream, returns raw response body (pass-through, no parsing)
   - On failure with `strategy: "fallback"`, waits `retry_delay_secs` then tries next provider
4. Streaming uses a flushing `bufio.Writer` (4KB buffer) — no fallback once streaming starts

### Key Design Patterns

- **Handler closures**: `NewOpenAIHandler(registry, forwarder)` returns `http.HandlerFunc`, capturing deps as closure variables. Not struct methods.
- **Interface-based DI**: Handlers depend on `RegistryReader` interface, not concrete `ModelRegistry`.
- **Defensive copying**: `ModelRegistry.Get()`/`List()` return deep copies — tests verify mutation of returned values doesn't affect registry state.
- **Transparent proxy**: Forwarder passes raw upstream bytes to client without parsing response bodies (except status code).
- **Zero dependencies**: No third-party packages. All stdlib.

## Configuration

Config file search order: `./config.json` → `~/.config/model-router/config.json` → defaults (port 12345, no models).

API keys support `env:VAR_NAME` syntax (expanded via `os.Expand`). Optional `.env` file loaded from cwd or config file directory.

See `config.example.json` for the full schema.

## Deployment

- **Docker**: Multi-stage build (Go Alpine → distroless/static). `docker compose up -d`.
- **CI**: GitHub Actions build Docker images to ghcr.io and cross-compiled binaries for linux/darwin on amd64/arm64.
- **Dependabot**: Daily checks for Go modules and Docker base images.

## Testing

All tests use stdlib `testing` + `httptest` — no external test frameworks. Tests spin up `httptest.NewServer` mock upstreams for real HTTP round-trips. Key test areas:

- Format conversion correctness (OpenAI→Anthropic field mapping, defaults, immutability)
- Fallback logic (success on first, failover to second, all-fail → 502)
- Streaming (flushing, oversized response truncation, error after headers sent)
- Security (API keys/URLs redacted from `/models` response)
- Config (env expansion, `.env` loading, file discovery)
