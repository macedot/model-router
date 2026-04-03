---
name: refactor-core-http-server
description: Workflow command scaffold for refactor-core-http-server in model-router.
allowed_tools: ["Bash", "Read", "Write", "Grep", "Glob"]
---

# /refactor-core-http-server

Use this workflow when working on **refactor-core-http-server** in `model-router`.

## Goal

Migrates or refactors the core HTTP server implementation, including handler logic and supporting services, often replacing frameworks or changing server patterns.

## Common Files

- `main.go`
- `handlers/anthropic.go`
- `handlers/anthropic_test.go`
- `handlers/models.go`
- `handlers/models_test.go`
- `handlers/openai.go`

## Suggested Sequence

1. Understand the current state and failure mode before editing.
2. Make the smallest coherent change that satisfies the workflow goal.
3. Run the most relevant verification for touched files.
4. Summarize what changed and what still needs review.

## Typical Commit Signals

- Update main server entrypoint (main.go) to use new HTTP framework or pattern
- Refactor all route handlers (handlers/*.go) to match new framework API
- Update or rewrite handler tests (handlers/*_test.go) to match new framework's testing style
- Update supporting services (services/forwarder.go) to integrate with new handler logic

## Notes

- Treat this as a scaffold, not a hard-coded script.
- Update the command if the workflow evolves materially.