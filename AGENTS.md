# Agent Guidelines for hourglass

## Build/Lint/Test Commands
- **Build**: `go build`
- **Run**: `go run main.go`
- **Test**: `go test ./...` (all tests) or `go test -run TestName` (single test)
- **Lint**: `gofmt -d .` (check formatting)
- **Format**: `gofmt -w .`

## Code Style Guidelines
- Use `gofmt` for consistent formatting
- Follow Go naming: camelCase for vars/functions, PascalCase for exported types
- Keep functions focused and under 50 lines when possible
- Group imports: standard library first, then third-party packages with blank lines
- Use `if err != nil` pattern consistently; return errors early with descriptive messages
- Define custom types for clarity (e.g., `appState int`)
- Use meaningful field names in structs and group related fields together
- Use Bubbletea ecosystem (bubbles, bubbletea) for TUI components
- Use beeep for cross-platform notifications
- Check go.mod before adding new dependencies
- Write table-driven tests for multiple cases and test error conditions