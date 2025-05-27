# Contributing to AIED

Thank you for your interest in contributing to AIED! This guide will help you get started with contributing to the project.

## Getting Started

1. **Fork the repository** on GitHub
2. **Clone your fork** locally:
   ```bash
   git clone https://github.com/YOUR-USERNAME/aied.git
   cd aied
   ```
3. **Add upstream remote**:
   ```bash
   git remote add upstream https://github.com/dshills/aied.git
   ```
4. **Create a feature branch**:
   ```bash
   git checkout -b feature/your-feature-name
   ```

## Development Setup

1. **Install Go 1.21+** from [golang.org](https://golang.org/dl/)
2. **Install dependencies**:
   ```bash
   go mod download
   ```
3. **Run tests**:
   ```bash
   go test ./...
   ```
4. **Build the project**:
   ```bash
   go build -o aied .
   ```

## Code Style

- Follow standard Go conventions and idioms
- Use `gofmt` to format your code
- Run `go vet` to catch common mistakes
- Add comments for exported functions and types
- Keep functions focused and small
- Write descriptive commit messages

## Testing

- Write tests for new functionality
- Ensure all tests pass before submitting PR
- Aim for good test coverage
- Use table-driven tests where appropriate

Example test structure:
```go
func TestFeature(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
    }{
        {"basic case", "input", "expected"},
        // Add more test cases
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := YourFunction(tt.input)
            if result != tt.expected {
                t.Errorf("got %v, want %v", result, tt.expected)
            }
        })
    }
}
```

## Pull Request Process

1. **Update your fork**:
   ```bash
   git fetch upstream
   git checkout main
   git merge upstream/main
   ```

2. **Rebase your feature branch**:
   ```bash
   git checkout feature/your-feature-name
   git rebase main
   ```

3. **Run tests and linters**:
   ```bash
   go test ./...
   go vet ./...
   gofmt -s -w .
   ```

4. **Push to your fork**:
   ```bash
   git push origin feature/your-feature-name
   ```

5. **Create Pull Request**:
   - Use a clear, descriptive title
   - Reference any related issues
   - Describe what changes you made and why
   - Include screenshots for UI changes

## Pull Request Template

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Tests pass locally
- [ ] Added new tests
- [ ] Updated documentation

## Related Issues
Fixes #(issue number)

## Screenshots (if applicable)
```

## Areas for Contribution

### Good First Issues
- Add more VIM commands
- Improve error messages
- Add keyboard shortcuts
- Write documentation
- Add unit tests

### Feature Ideas
- Syntax highlighting
- Multiple buffers/splits
- Search and replace
- File tree explorer
- Git integration
- Plugin system
- LSP support

### AI Provider Support
- Add new AI providers
- Improve context extraction
- Add streaming responses
- Implement caching

## Reporting Issues

When reporting issues, please include:
- OS and version
- Go version
- Steps to reproduce
- Expected vs actual behavior
- Error messages/logs
- Configuration (without API keys)

## Code of Conduct

- Be respectful and inclusive
- Welcome newcomers
- Focus on constructive criticism
- Assume good intentions
- Help others learn

## Getting Help

- Check existing issues and PRs
- Read the documentation
- Ask in GitHub Discussions
- Tag maintainers if needed

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

## Recognition

Contributors will be recognized in:
- The project README
- Release notes
- GitHub contributors page

Thank you for contributing to AIED! 🎉