# Contributing to Secrets Snapshot CLI

Thank you for your interest in contributing to Secrets Snapshot CLI! This document provides guidelines and information for contributors.

## 🤝 How to Contribute

### Types of Contributions

We welcome contributions in the following areas:

- **Bug fixes** - Fix issues and improve reliability
- **Feature enhancements** - Add new functionality
- **Documentation** - Improve docs, examples, and guides
- **Testing** - Add tests and improve test coverage
- **Performance** - Optimize code and reduce resource usage
- **Security** - Improve security practices and implementations

### Before You Start

1. **Check existing issues** - Search for existing issues or discussions
2. **Discuss major changes** - Open an issue to discuss significant changes
3. **Follow the code style** - Maintain consistency with existing code
4. **Write tests** - Ensure new code is properly tested

## 🛠️ Development Setup

### Prerequisites

- **Go 1.22+** - [Download here](https://golang.org/dl/)
- **Git** - For version control
- **Make** - For build automation (optional)

### Local Development

```bash
# Clone the repository
git clone https://github.com/secretsnap/cli.git
cd cli

# Install dependencies
go mod download

# Build the CLI
make build

# Run tests
make test

# Development mode
make dev
```

### Project Structure

```
cli/
├── cmd/                   # Command implementations
│   ├── init.go           # Initialize configuration
│   ├── bundle.go         # Encrypt/decrypt bundles
│   ├── run.go            # Run commands with env vars
│   ├── login.go          # Authentication
│   ├── project.go        # Project management
│   ├── pull.go           # Download bundles
│   ├── share.go          # Team sharing
│   ├── audit.go          # Audit logs
│   └── commands.go       # Command registration
├── internal/             # Internal packages
│   ├── api/             # HTTP client for API
│   ├── config/          # Configuration management
│   ├── crypto/          # Age encryption helpers
│   └── run/             # Process runner
├── examples/            # Usage examples
├── scripts/             # Build and release scripts
├── main.go              # CLI entry point
└── go.mod               # Go dependencies
```

## 📝 Code Style Guidelines

### Go Code Style

- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `gofmt` for formatting
- Follow standard Go naming conventions
- Write clear, descriptive comments
- Keep functions focused and concise

### Command Structure

- Use Cobra for command structure
- Follow the pattern established in existing commands
- Include proper help text and examples
- Use consistent flag naming

### Error Handling

- Return meaningful error messages
- Use `fmt.Errorf` for error wrapping
- Log errors appropriately
- Provide user-friendly error messages

## 🧪 Testing

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
go test -cover ./...

# Run specific test
go test ./internal/crypto

# Run tests with verbose output
go test -v ./...
```

### Writing Tests

- Write tests for new functionality
- Aim for good test coverage
- Use table-driven tests where appropriate
- Mock external dependencies
- Test both success and error cases

### Test Structure

```go
func TestFunctionName(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
        wantErr  bool
    }{
        {
            name:     "valid input",
            input:    "test",
            expected: "expected",
            wantErr:  false,
        },
        // Add more test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := FunctionName(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("FunctionName() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if result != tt.expected {
                t.Errorf("FunctionName() = %v, want %v", result, tt.expected)
            }
        })
    }
}
```

## 🔧 Building and Releasing

### Building

```bash
# Build for current platform
make build

# Build for all platforms
make release

# Install locally
make install
```

### Release Process

1. **Update version** - Update version in relevant files
2. **Run tests** - Ensure all tests pass
3. **Build releases** - Run `make release`
4. **Create release** - Tag and create GitHub release
5. **Update install script** - Update version in install script

## 🐛 Bug Reports

### Before Reporting

1. **Search existing issues** - Check if the bug is already reported
2. **Reproduce the issue** - Ensure you can consistently reproduce it
3. **Check environment** - Note your OS, Go version, etc.

### Bug Report Template

```markdown
**Bug Description**
A clear description of what the bug is.

**Steps to Reproduce**

1. Run command: `secretsnap bundle file.env`
2. Enter passphrase: `test`
3. See error: `failed to encrypt: invalid input`

**Expected Behavior**
What you expected to happen.

**Actual Behavior**
What actually happened.

**Environment**

- OS: macOS 12.0
- Go version: 1.22.0
- Secretsnap version: v0.1.0

**Additional Context**
Any other context about the problem.
```

## 💡 Feature Requests

### Before Requesting

1. **Check existing features** - Ensure the feature doesn't already exist
2. **Consider use case** - Think about the broader use case
3. **Check roadmap** - See if it's already planned

### Feature Request Template

```markdown
**Feature Description**
A clear description of the feature you'd like to see.

**Use Case**
Why this feature would be useful and how you'd use it.

**Proposed Implementation**
Any thoughts on how this could be implemented.

**Alternatives Considered**
Other approaches you've considered.

**Additional Context**
Any other relevant information.
```

## 🔒 Security

### Security Issues

If you discover a security vulnerability, please:

1. **Do not open a public issue**
2. **Email security@secretsnap.dev** with details
3. **Include "SECURITY" in the subject line**
4. **Provide a detailed description** of the vulnerability

### Security Guidelines

- Never commit secrets or sensitive data
- Follow secure coding practices
- Validate all inputs
- Use secure random number generation
- Implement proper error handling

## 📋 Pull Request Process

### Before Submitting

1. **Fork the repository**
2. **Create a feature branch** - `git checkout -b feature/amazing-feature`
3. **Make your changes** - Follow the coding guidelines
4. **Write tests** - Ensure new code is tested
5. **Update documentation** - Update relevant docs
6. **Run tests** - Ensure all tests pass
7. **Commit your changes** - Use clear commit messages

### Commit Message Format

```
type(scope): description

[optional body]

[optional footer]
```

Examples:

- `feat(bundle): add support for custom output formats`
- `fix(crypto): handle empty passphrase gracefully`
- `docs(readme): update installation instructions`

### Pull Request Template

```markdown
**Description**
Brief description of the changes.

**Type of Change**

- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

**Testing**

- [ ] Added tests for new functionality
- [ ] All tests pass
- [ ] Manual testing completed

**Checklist**

- [ ] Code follows style guidelines
- [ ] Self-review completed
- [ ] Documentation updated
- [ ] No breaking changes (or documented)

**Screenshots** (if applicable)
```

### Review Process

1. **Automated checks** - CI/CD will run tests
2. **Code review** - Maintainers will review your code
3. **Feedback** - Address any feedback or requested changes
4. **Merge** - Once approved, your PR will be merged

## 🏷️ Versioning

We use [Semantic Versioning](https://semver.org/) for versioning:

- **MAJOR** - Incompatible API changes
- **MINOR** - New functionality in a backwards-compatible manner
- **PATCH** - Backwards-compatible bug fixes

## 📄 License

By contributing to Secrets Snapshot CLI, you agree that your contributions will be licensed under the MIT License.

## 🆘 Getting Help

### Questions and Discussion

- **GitHub Issues** - For bugs and feature requests
- **GitHub Discussions** - For questions and general discussion
- **Email** - support@secretsnap.dev for private matters

### Resources

- [Go Documentation](https://golang.org/doc/)
- [Cobra Documentation](https://github.com/spf13/cobra)
- [Age Encryption](https://age-encryption.org/)
- [Effective Go](https://golang.org/doc/effective_go.html)

## 🙏 Recognition

Contributors will be recognized in:

- **README.md** - For significant contributions
- **Release notes** - For each release
- **GitHub contributors** - Automatic recognition

---

**Thank you for contributing to Secrets Snapshot CLI! 🎉**

Your contributions help make this tool better for everyone in the developer community.
