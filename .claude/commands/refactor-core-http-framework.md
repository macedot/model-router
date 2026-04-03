---
name: refactor-core-http-framework
description: Workflow command scaffold for refactor-core-http-framework in model-router.
allowed_tools: ["Bash", "Read", "Write", "Grep", "Glob"]
---

# /refactor-core-http-framework

Use this workflow when working on **refactor-core-http-framework** in `model-router`.

## Goal

Migrates the core HTTP framework (e.g., Fiber to net/http), updating handlers, middleware, and tests to match new patterns.

## Common Files

- `handlers/*.go`
- `handlers/*_test.go`
- `main.go`
- `services/forwarder.go`

## Suggested Sequence

1. Understand the current state and failure mode before editing.
2. Make the smallest coherent change that satisfies the workflow goal.
3. Run the most relevant verification for touched files.
4. Summarize what changed and what still needs review.

## Typical Commit Signals

- Update all handler files to use new framework APIs.
- Rewrite middleware and server setup in main.go.
- Update or rewrite tests for handlers to use new test patterns.
- Update any service files that interact with HTTP responses.

## Notes

- Treat this as a scaffold, not a hard-coded script.
- Update the command if the workflow evolves materially.