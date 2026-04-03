```markdown
# model-router Development Patterns

> Auto-generated skill from repository analysis

## Overview

This skill teaches you how to work effectively in the `model-router` Go codebase. You'll learn the project's coding conventions, commit patterns, testing strategies, and step-by-step workflows for common and critical development tasks. Whether you're refactoring the HTTP framework, improving test coverage, or fixing critical bugs, this guide provides clear instructions and practical code examples.

## Coding Conventions

### File Naming

- Use **snake_case** for all file names.
  - Example: `model_router.go`, `forwarder_service.go`

### Import Style

- Use **relative imports** within the project.
  - Example:
    ```go
    import (
        "handlers"
        "services"
    )
    ```

### Export Style

- Use **named exports** for functions, types, and variables that need to be accessed outside their package.
  - Example:
    ```go
    // In handlers/user_handler.go
    package handlers

    func GetUserHandler(w http.ResponseWriter, r *http.Request) {
        // ...
    }
    ```

### Commit Patterns

- Follow **conventional commit** style.
- Common prefixes: `refactor`, `chore`, `fix`, `test`
- Example commit messages:
  - `fix: correct timeout handling in forwarder service`
  - `refactor: migrate handlers to net/http`
  - `test: add edge case tests for error handler`
- Average commit message length: ~57 characters

## Workflows

### Refactor Core HTTP Framework

**Trigger:** When changing the underlying HTTP framework or major server architecture  
**Command:** `/refactor-http-framework`

1. Update all handler files to use the new framework APIs.
    - Example: Migrate from Fiber to `net/http`
    ```go
    // Before (Fiber)
    func GetUser(c *fiber.Ctx) error { ... }

    // After (net/http)
    func GetUser(w http.ResponseWriter, r *http.Request) { ... }
    ```
2. Rewrite middleware and server setup in `main.go`.
    - Replace framework-specific setup with new framework's idioms.
3. Update or rewrite tests for handlers to use new test patterns.
    - Example: Use `httptest` for `net/http` handlers.
4. Update any service files that interact with HTTP responses.
    - Ensure response formatting and error handling match the new framework.

### Dependency Removal After Refactor

**Trigger:** When obsolete dependencies are removed after a major refactor  
**Command:** `/remove-dependency`

1. Run `go mod tidy` to clean up `go.mod` and `go.sum`.
    ```sh
    go mod tidy
    ```
2. Commit the updated dependency files.
    ```sh
    git add go.mod go.sum
    git commit -m "chore: remove obsolete dependencies"
    ```

### Add or Improve Test Coverage for Services and Handlers

**Trigger:** When increasing test coverage or verifying new/refactored logic  
**Command:** `/add-tests`

1. Create or update `*_test.go` files for services and handlers.
    - Example: `handlers/user_handler_test.go`
2. Write tests for success, error, and edge cases.
    ```go
    func TestGetUserHandler_Success(t *testing.T) {
        req := httptest.NewRequest("GET", "/user/1", nil)
        rr := httptest.NewRecorder()
        handlers.GetUserHandler(rr, req)
        // assertions...
    }
    ```
3. Verify coverage improvements.
    ```sh
    go test -cover ./...
    ```

### Critical Bugfix in Handler or Service

**Trigger:** When a critical runtime bug or production issue is discovered  
**Command:** `/fix-critical-bug`

1. Identify and fix the bug in the relevant handler or service file.
    - Example: Fix a race condition in `services/forwarder.go`
2. Update configuration if needed (e.g., timeouts in `config/timeout.go`).
3. Add or update tests to cover the fixed scenario.
    ```go
    func TestForwarder_Timeout(t *testing.T) {
        // Simulate timeout and assert correct handling
    }
    ```

## Testing Patterns

- Test files use the pattern `*_test.go` and are located alongside their implementation files.
- The testing framework is not explicitly specified, but standard Go `testing` and `httptest` packages are commonly used.
- Tests cover success, error, and edge cases.
- Example test file:
    ```go
    // handlers/user_handler_test.go
    package handlers

    import (
        "net/http/httptest"
        "testing"
    )

    func TestGetUserHandler(t *testing.T) {
        req := httptest.NewRequest("GET", "/user/1", nil)
        rr := httptest.NewRecorder()
        GetUserHandler(rr, req)
        // assertions...
    }
    ```

## Commands

| Command                | Purpose                                                      |
|------------------------|--------------------------------------------------------------|
| /refactor-http-framework | Migrate or refactor the core HTTP framework                |
| /remove-dependency     | Remove obsolete dependencies after a refactor                |
| /add-tests             | Add or improve test coverage for services and handlers       |
| /fix-critical-bug      | Fix a critical bug in handler or service code                |
```