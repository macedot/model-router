# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build Commands

```bash
make build              # Build binary (VERSION=1.2.3 sets the version)
make run                # Build and run
make test              # Run go test ./...
make clean             # Remove binary
make install           # Build and install binary to ~/.local/bin/
make release VERSION=1.0.0  # Create git tag and GitHub release
```

Use `bash test/assert.sh` to run the integration test scripts directly.
```

## Architecture

**model-router** is an AI model proxy that accepts OpenAI or Anthropic format requests and forwards them to external providers with format conversion.

### Request Flow

```
Client → Handler → Forwarder → External Provider
         (route)    (convert)    (API call)
```

- `/v1/chat/completions` (OpenAI) → routes to OpenAIHandler
- `/v1/messages` (Anthropic) → routes to AnthropicHandler
- `/models` → lists registered internal models

### Key Concepts

1. **InternalModel**: Client-facing model name with one or more ExternalModel providers
2. **ExternalModel**: Actual provider endpoint with URL, API key, and target format
3. **Format Conversion**: Forwarder converts OpenAI ↔ Anthropic based on ExternalModel.Format
4. **Model Name Replacement**: ExternalModel.Name replaces the incoming model name before forwarding

### Config System

Config file search order: current working directory → `~/.model-router/config.json` → anywhere on `$PATH`
- `.env` loading: Loads from current directory and config directory (non-blocking)
- `env:VAR_NAME` syntax in config for environment variable substitution

### Constants

All defaults are in `config/config.go` `Defaults` struct:
- `Port`, `ReadTimeout`, `WriteTimeout`, `BodyLimit`, `ShutdownTimeout`

### Services (diagonal dependency)

```
main → ModelRegistry → []InternalModel
     → Forwarder
```

Handlers receive services via constructor injection.
