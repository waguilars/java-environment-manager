# Contributing to jem

Thank you for your interest in contributing to jem! This document provides guidelines for contributing to the project.

## How to Contribute

### Reporting Bugs

- Open an issue on GitHub with a clear description
- Include your OS, jem version, and steps to reproduce
- Include error messages and logs

### Suggesting Features

- Open an issue with the feature request label
- Explain the use case and benefits
- Provide examples if possible

### Contributing Code

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Run tests (`go test ./...`)
5. Commit with clear messages
6. Push to your branch
7. Open a Pull Request

## Code Style

### Go Style Guide

- Follow [Effective Go](https://go.dev/doc/effective_go) guidelines
- Use `gofmt` for formatting
- Keep functions focused and small (< 50 lines preferred)
- Use meaningful variable and function names

### File Organization

```
internal/          # Private application code
cmd/               # CLI command entry points
pkg/               # Public utilities (if needed)
```

### Naming Conventions

- **Functions/Variables**: `camelCase`
- **Types/Structs**: `PascalCase`
- **Constants**: `UPPER_SNAKE_CASE`
- **Files**: `lower_snake_case.go`

### Error Handling

```go
// Good
if err != nil {
    return fmt.Errorf("failed to install JDK: %w", err)
}

// Bad
if err != nil {
    log.Fatal(err)
}
```

### Testing

- All code must have tests
- Use table-driven tests for complex logic
- Test both success and failure cases

```go
func TestInstallJDK(t *testing.T) {
    tests := []struct {
        name    string
        version string
        wantErr bool
    }{
        {"valid version", "17", false},
        {"invalid version", "invalid", true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test implementation
        })
    }
}
```

## Testing

### Run All Tests

```bash
go test ./...
```

### Run Specific Package Tests

```bash
go test ./internal/jdk/...
go test ./internal/downloader/...
```

### Run Tests with Coverage

```bash
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Test Requirements

- All tests must pass before PR submission
- Minimum 80% code coverage
- Tests should be independent and reproducible

## Pull Request Process

### PR Template

```markdown
## Description
[Describe your changes]

## Related Issue
[Fixes #issue-number]

## Testing
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Manual testing completed

## Documentation
- [ ] README updated (if needed)
- [ ] CLI help text updated
```

### PR Review Checklist

- [ ] Code follows project style
- [ ] Tests added/updated
- [ ] Documentation updated
- [ ] No breaking changes (or documented)
- [ ] Version updated in `main.go`

### PR Labels

- `bug` - Bug fix
- `enhancement` - New feature
- `breaking` - Breaking change
- `documentation` - Docs only
- `performance` - Performance improvement

## Development Setup

### Prerequisites

- Go 1.26.1+
- Git
- Make (optional)

### Clone Repository

```bash
git clone https://github.com/user/jem.git
cd jem
```

### Build

```bash
go build -o jem
```

### Run Tests

```bash
go test ./...
```

### Format Code

```bash
gofmt -w .
```

### Vet Code

```bash
go vet ./...
```

## Community

- Join our Discord server (if available)
- Follow us on Twitter @jem_java
- Email maintainers for questions

## Recognition

Contributors will be recognized in:
- README contributors section
- Release notes
- Project documentation

## Questions?

Open an issue or contact the maintainers.
