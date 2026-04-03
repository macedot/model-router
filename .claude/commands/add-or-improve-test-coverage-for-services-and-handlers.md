---
name: add-or-improve-test-coverage-for-services-and-handlers
description: Workflow command scaffold for add-or-improve-test-coverage-for-services-and-handlers in model-router.
allowed_tools: ["Bash", "Read", "Write", "Grep", "Glob"]
---

# /add-or-improve-test-coverage-for-services-and-handlers

Use this workflow when working on **add-or-improve-test-coverage-for-services-and-handlers** in `model-router`.

## Goal

Adds or improves test coverage for service and handler logic, often after a refactor or bugfix.

## Common Files

- `services/*_test.go`
- `handlers/*_test.go`

## Suggested Sequence

1. Understand the current state and failure mode before editing.
2. Make the smallest coherent change that satisfies the workflow goal.
3. Run the most relevant verification for touched files.
4. Summarize what changed and what still needs review.

## Typical Commit Signals

- Create or update *_test.go files for services and handlers.
- Write tests for success, error, and edge cases.
- Verify coverage improvements.

## Notes

- Treat this as a scaffold, not a hard-coded script.
- Update the command if the workflow evolves materially.