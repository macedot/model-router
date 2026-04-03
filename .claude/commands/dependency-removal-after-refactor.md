---
name: dependency-removal-after-refactor
description: Workflow command scaffold for dependency-removal-after-refactor in model-router.
allowed_tools: ["Bash", "Read", "Write", "Grep", "Glob"]
---

# /dependency-removal-after-refactor

Use this workflow when working on **dependency-removal-after-refactor** in `model-router`.

## Goal

Removes obsolete dependencies from go.mod and go.sum after a major refactor eliminates their usage.

## Common Files

- `go.mod`
- `go.sum`

## Suggested Sequence

1. Understand the current state and failure mode before editing.
2. Make the smallest coherent change that satisfies the workflow goal.
3. Run the most relevant verification for touched files.
4. Summarize what changed and what still needs review.

## Typical Commit Signals

- Run go mod tidy to clean up go.mod and go.sum.
- Commit the updated dependency files.

## Notes

- Treat this as a scaffold, not a hard-coded script.
- Update the command if the workflow evolves materially.