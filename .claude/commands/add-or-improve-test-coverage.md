---
name: add-or-improve-test-coverage
description: Workflow command scaffold for add-or-improve-test-coverage in model-router.
allowed_tools: ["Bash", "Read", "Write", "Grep", "Glob"]
---

# /add-or-improve-test-coverage

Use this workflow when working on **add-or-improve-test-coverage** in `model-router`.

## Goal

Adds or improves test coverage for handlers and services, often after a refactor or bug fix, including both unit and integration tests.

## Common Files

- `handlers/openai_test.go`
- `services/forwarder_test.go`
- `services/registry_test.go`
- `services/registry.go`

## Suggested Sequence

1. Understand the current state and failure mode before editing.
2. Make the smallest coherent change that satisfies the workflow goal.
3. Run the most relevant verification for touched files.
4. Summarize what changed and what still needs review.

## Typical Commit Signals

- Write or update test files for handlers (handlers/*_test.go)
- Write or update test files for services (services/*_test.go)
- Ensure coverage for edge cases and new logic
- Update implementation files if test coverage reveals issues

## Notes

- Treat this as a scaffold, not a hard-coded script.
- Update the command if the workflow evolves materially.