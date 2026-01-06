# Contributing to findata-go

Thank you for your interest in contributing to findata-go! This document provides guidelines and instructions for contributing.

## Code of Conduct

Be respectful, inclusive, and professional in all interactions.

## How to Contribute

### Reporting Bugs

1. Check if the bug has already been reported in [Issues](https://github.com/Vikramarjuna/findata-go/issues)
2. If not, create a new issue using the bug report template
3. Provide a minimal reproducible example
4. Include your environment details (Go version, OS, etc.)

### Suggesting Features

1. Check if the feature has already been requested
2. Create a new issue using the feature request template
3. Clearly describe the use case and proposed API
4. Be open to discussion and feedback

### Submitting Pull Requests

1. **Fork the repository** and create a new branch from `main`
2. **Make your changes** following the coding standards below
3. **Add tests** for any new functionality
4. **Run tests and linter** to ensure everything passes
5. **Update documentation** if needed
6. **Submit a pull request** with a clear description

## Development Setup

### Prerequisites

- Go 1.22 or higher (1.24 recommended)
- golangci-lint v2.7.2 or higher for linting

### Clone and Setup

```bash
git clone https://github.com/Vikramarjuna/findata-go.git
cd findata-go
make deps
```

### Running Tests

```bash
# Run all tests
make test

# Run tests without integration tests
make test-short

# Run tests with coverage
make coverage
```

### Running Linter

```bash
make lint
```

### Building Examples

```bash
make examples
make run-nse
make run-mf
```

## Coding Standards

### Go Style

- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `gofmt` for formatting (run `make fmt`)
- Follow the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)

### Code Organization

- Keep functions small and focused
- Use meaningful variable and function names
- Add comments for exported functions and types
- Group related functionality in packages

### Testing

- Write table-driven tests where appropriate
- Aim for high test coverage (>80%)
- Test both success and error cases
- Use meaningful test names: `TestFunctionName_Scenario`

### Error Handling

- Return errors, don't panic
- Wrap errors with context using `fmt.Errorf`
- Use sentinel errors for expected error conditions

### Logging

- Use the logger interface, not direct slog calls
- Log at appropriate levels (DEBUG, INFO, WARN, ERROR)
- Include relevant context in log messages

## Commit Messages

Follow conventional commits format:

```
<type>(<scope>): <subject>

<body>

<footer>
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `test`: Test changes
- `refactor`: Code refactoring
- `perf`: Performance improvements
- `chore`: Build/tooling changes

Examples:
```
feat(equity): add support for BSE quotes

fix(mf): handle empty search results correctly

docs(readme): update installation instructions
```

## Versioning

We use [Semantic Versioning](https://semver.org/):

- **MAJOR**: Breaking changes
- **MINOR**: New features (backward compatible)
- **PATCH**: Bug fixes (backward compatible)

## Release Process

1. Update CHANGELOG.md
2. Update VERSION file
3. Create a tag: `make tag-version VERSION=v1.0.0`
4. Push the tag: `git push origin v1.0.0`
5. GitHub Actions will automatically create a release

## Questions?

Feel free to open an issue for any questions or clarifications.

Thank you for contributing! 🎉

