```markdown
# model-router Development Patterns

> Auto-generated skill from repository analysis

## Overview

This skill teaches you the core development patterns, coding conventions, and maintenance workflows for the `model-router` Go codebase. You'll learn how to refactor the HTTP server, manage dependencies, maintain and improve test coverage, and keep deployment scripts and documentation up to date. The guide is based on real repository analysis and is designed to help contributors quickly get up to speed with established practices.

## Coding Conventions

- **File Naming:**  
  Use `camelCase` for file names.  
  *Example:*  
  ```
  handlers/openai.go
  services/forwarder.go
  ```

- **Import Style:**  
  Use relative imports for internal packages.  
  *Example:*  
  ```go
  import (
      "model-router/handlers"
      "model-router/services"
  )
  ```

- **Export Style:**  
  Use named exports for functions, types, and variables.  
  *Example:*  
  ```go
  // handlers/openai.go
  func HandleOpenAIRequest(w http.ResponseWriter, r *http.Request) { ... }
  ```

- **Commit Messages:**  
  Follow the [Conventional Commits](https://www.conventionalcommits.org/) style with these prefixes:  
    - `chore:`
    - `refactor:`
    - `fix:`
    - `test:`
  *Example:*  
  ```
  refactor: migrate handlers to new http framework
  ```

## Workflows

### Refactor Core HTTP Server
**Trigger:** When changing the underlying HTTP server framework or majorly refactoring request handling  
**Command:** `/refactor-http-server`

1. Update the main server entrypoint (`main.go`) to use the new HTTP framework or pattern.
2. Refactor all route handlers (`handlers/*.go`) to match the new framework's API.
3. Update or rewrite handler tests (`handlers/*_test.go`) to match the new framework's testing style.
4. Update supporting services (e.g., `services/forwarder.go`) to integrate with the new handler logic.

*Example:*
```go
// main.go
// Before:
http.HandleFunc("/openai", handlers.HandleOpenAIRequest)

// After (using a new router):
router.POST("/openai", handlers.HandleOpenAIRequest)
```

### Dependency Removal After Refactor
**Trigger:** When a framework or major dependency is removed from the codebase  
**Command:** `/remove-dependency`

1. Run `go mod tidy` to clean up `go.mod` and `go.sum`.
2. Commit the updated `go.mod` and `go.sum` files.

*Example:*
```sh
go mod tidy
git add go.mod go.sum
git commit -m "chore: remove obsolete dependencies after refactor"
```

### Add or Improve Test Coverage
**Trigger:** When new features are added, bugs are fixed, or after major refactors  
**Command:** `/add-tests`

1. Write or update test files for handlers (e.g., `handlers/openai_test.go`).
2. Write or update test files for services (e.g., `services/forwarder_test.go`).
3. Ensure coverage for edge cases and new logic.
4. Update implementation files if test coverage reveals issues.

*Example:*
```go
// handlers/openai_test.go
func TestHandleOpenAIRequest(t *testing.T) {
    req := httptest.NewRequest("POST", "/openai", nil)
    w := httptest.NewRecorder()
    HandleOpenAIRequest(w, req)
    // assertions...
}
```

### Remove Obsolete Scripts and Docs
**Trigger:** When deployment methods change and old scripts are no longer needed  
**Command:** `/remove-scripts`

1. Delete obsolete shell scripts (`install.sh`, `uninstall.sh`).
2. Remove related Makefile targets.
3. Update `README.md` to document the new deployment method.

*Example:*
```sh
git rm install.sh uninstall.sh
# Edit Makefile and README.md as needed
git commit -am "chore: remove obsolete deployment scripts and update docs"
```

## Testing Patterns

- **Test File Naming:**  
  Test files follow the pattern `*_test.go` and are placed alongside their implementation files.

- **Framework:**  
  The specific testing framework is not detected, but standard Go testing (`testing` package) is likely used.

- **Test Example:**
  ```go
  // services/forwarder_test.go
  func TestForwardRequest(t *testing.T) {
      // Arrange
      // Act
      // Assert
  }
  ```

## Commands

| Command                | Purpose                                                    |
|------------------------|------------------------------------------------------------|
| /refactor-http-server  | Refactor the core HTTP server and related handlers         |
| /remove-dependency     | Remove obsolete dependencies from go.mod and go.sum        |
| /add-tests             | Add or improve test coverage for handlers and services     |
| /remove-scripts        | Remove obsolete deployment scripts and update documentation|
```
