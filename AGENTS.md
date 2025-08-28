# AGENTS.md

## Build/Lint/Test Commands
- **Build**: `go build`
- **Run**: `go run main.go`
- **Test**: `go test ./...` (run all tests)
- **Single test**: `go test -run TestName`
- **Lint**: `go vet` and `golangci-lint run`
- **Format**: `gofmt -w .`

## Code Style Guidelines

### General
- Follow standard Go formatting with `gofmt`
- Use `go vet` for static analysis
- Keep lines under 120 characters

### Imports
- Group imports: standard library, blank line, third-party packages
- Use blank imports only when required for side effects

### Naming
- Use `camelCase` for unexported identifiers
- Use `PascalCase` for exported identifiers
- Constants use `SCREAMING_SNAKE_CASE` when appropriate

### Error Handling
- Always check and handle errors explicitly
- Return errors rather than panicking in library code
- Use descriptive error messages

### Types
- Use `time.Duration` for time intervals
- Prefer structs over maps for complex data
- Use interfaces for abstraction when beneficial

### Bubbletea Patterns
- Follow standard Bubbletea TUI patterns
- Use proper model initialization in `Init()`
- Handle messages appropriately in `Update()`
- Keep `View()` focused on rendering