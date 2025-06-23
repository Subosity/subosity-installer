# 🤝 Contributing to Subosity Installer

Thank you for your interest in contributing to the Subosity Installer! This document provides guidelines and information for contributors.

## 📋 Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Contribution Workflow](#contribution-workflow)
- [Code Standards](#code-standards)
- [Testing Requirements](#testing-requirements)
- [Documentation](#documentation)
- [Pull Request Process](#pull-request-process)
- [Release Process](#release-process)

## 📜 Code of Conduct

This project adheres to a [Code of Conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code. Please report unacceptable behavior to the project maintainers.

## 🚀 Getting Started

### Prerequisites

- **Go 1.21+** - Required for building the host binary
- **Docker 24.0+** - Required for container development and testing
- **Git** - Version control
- **Make** - Build automation
- **golangci-lint** - Code quality and linting

### Development Tools

```bash
# Install golangci-lint
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.54.2

# Install pre-commit hooks (optional but recommended)
pip install pre-commit
pre-commit install
```

## 🔧 Development Setup

1. **Fork and Clone**
   ```bash
   git clone https://github.com/YOUR_USERNAME/subosity-installer.git
   cd subosity-installer
   ```

2. **Set up Development Environment**
   ```bash
   # Install dependencies
   go mod download
   
   # Verify setup
   make test
   make lint
   ```

3. **Build and Test**
   ```bash
   # Build binary
   make build
   
   # Build container
   make build-container
   
   # Run integration tests
   make test-integration
   ```

## 🔄 Contribution Workflow

### 1. Create an Issue First

Before starting work:
- **Bug Reports**: Use the bug report template
- **Feature Requests**: Use the feature request template
- **Security Issues**: Email security@subosity.com privately

### 2. Development Process

1. **Create a branch** from `develop`:
   ```bash
   git checkout develop
   git pull origin develop
   git checkout -b feature/your-feature-name
   ```

2. **Follow the architecture**:
   - Read [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for design patterns
   - Review [docs/PRD.md](docs/PRD.md) for requirements
   - Follow [STYLE_GUIDE.md](STYLE_GUIDE.md) for coding standards

3. **Make atomic commits**:
   ```bash
   git commit -m "feat: add Docker installation validation
   
   - Implement OS-specific Docker installation checks
   - Add comprehensive error messages for failed installations
   - Include retry logic for network-dependent operations
   
   Closes #123"
   ```

### 3. Commit Message Format

We follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Build process or auxiliary tool changes

**Examples:**
```bash
feat(binary): add environment validation for Ubuntu 22.04
fix(container): resolve Supabase CLI installation timeout
docs(arch): update container-first architecture diagrams
test(integration): add Docker installation test scenarios
```

## 🎯 Code Standards

### Architecture Principles

- **Container-First**: Thin binary delegates to smart container
- **Supabase CLI Integration**: Never reimplement Supabase functionality
- **Production-Ready**: Enterprise-grade reliability and security
- **Idempotent Operations**: Safe to retry any operation
- **Zero Technical Debt**: Aggressive refactoring and code quality

### Code Quality Requirements

- **Function Length**: Maximum 50 lines per function
- **Interface Design**: Small, focused interfaces
- **Error Handling**: Structured errors with context and suggestions
- **Documentation**: Self-documenting code with clear comments
- **Resource Management**: Proper cleanup and context handling

### Example Code Structure

```go
// ✅ Good: Clear structure and error handling
func (s *DockerService) Install(ctx context.Context) error {
    if err := s.validateEnvironment(ctx); err != nil {
        return fmt.Errorf("environment validation failed: %w", err)
    }
    
    if err := s.downloadBinary(ctx); err != nil {
        return fmt.Errorf("failed to download Docker binary: %w", err)
    }
    
    return s.configureService(ctx)
}

// ❌ Avoid: Monolithic functions, poor error handling
func (s *DockerService) Install(ctx context.Context) error {
    // 100+ lines of mixed concerns
    // Silent failures or generic error messages
}
```

## 🧪 Testing Requirements

### Test Coverage Standards

- **Unit Tests**: 85%+ overall coverage
- **Critical Path**: 95%+ coverage for installation flows
- **Integration Tests**: Full end-to-end scenarios
- **Security Tests**: Vulnerability and penetration testing

### Test Structure

```go
func TestDockerInstallation(t *testing.T) {
    tests := []struct {
        name    string
        env     Environment
        wantErr bool
        errCode ErrorCode
    }{
        {
            name: "successful ubuntu installation",
            env: Environment{OS: "ubuntu", Version: "20.04"},
            wantErr: false,
        },
        {
            name: "unsupported operating system",
            env: Environment{OS: "windows", Version: "11"},
            wantErr: true,
            errCode: ErrCodeUnsupportedOS,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            service := NewDockerService()
            err := service.Install(context.Background(), tt.env)
            
            if (err != nil) != tt.wantErr {
                t.Errorf("Install() error = %v, wantErr %v", err, tt.wantErr)
            }
            
            if tt.wantErr && tt.errCode != "" {
                var installErr *InstallationError
                if errors.As(err, &installErr) {
                    assert.Equal(t, tt.errCode, installErr.Code)
                }
            }
        })
    }
}
```

### Running Tests

```bash
# Unit tests
go test ./...

# Integration tests
go test -tags=integration ./...

# Coverage report
make coverage

# Specific test with verbose output
go test -v ./internal/docker -run TestDockerInstallation
```

## 📚 Documentation

### Required Documentation Updates

When adding features, update:

1. **[docs/PRD.md](docs/PRD.md)** - Requirements and acceptance criteria
2. **[docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)** - Design patterns and interfaces
3. **[README.md](README.md)** - User-facing documentation
4. **[STYLE_GUIDE.md](STYLE_GUIDE.md)** - Coding standards (if applicable)

### Documentation Standards

- **Clear Examples**: Provide code examples for new APIs
- **Error Scenarios**: Document error conditions and recovery
- **Security Considerations**: Highlight security implications
- **Migration Guides**: For breaking changes

## 🔍 Pull Request Process

### Before Submitting

- [ ] All tests pass (`make test`)
- [ ] Code passes linting (`make lint`)
- [ ] Documentation is updated
- [ ] Security scan passes (`make security-scan`)
- [ ] Integration tests pass (`make test-integration`)

### PR Checklist

- [ ] **Clear Title**: Follows conventional commit format
- [ ] **Detailed Description**: Explains what, why, and how
- [ ] **Issue Reference**: Links to related issue(s)
- [ ] **Breaking Changes**: Clearly documented if any
- [ ] **Screenshots/Logs**: For UI or output changes

### Review Process

1. **Automated Checks**: CI/CD pipeline must pass
2. **Code Review**: At least one maintainer approval required
3. **Security Review**: For security-related changes
4. **Documentation Review**: For user-facing changes

### Merge Requirements

- ✅ All CI checks pass
- ✅ At least 1 maintainer approval
- ✅ No unresolved conversations
- ✅ Up-to-date with target branch
- ✅ Linear history (squash merge preferred)

## 🚀 Release Process

### Version Strategy

We follow [Semantic Versioning](https://semver.org/):
- **Major** (1.0.0): Breaking changes
- **Minor** (1.1.0): New features, backward compatible
- **Patch** (1.1.1): Bug fixes, backward compatible

### Release Workflow

1. **Feature Development**: Work on `develop` branch
2. **Release Preparation**: Create `release/v1.x.x` branch
3. **Testing**: Comprehensive testing on release branch
4. **Release**: Merge to `main` and tag
5. **Post-Release**: Merge back to `develop`

## 🆘 Getting Help

### Support Channels

- **GitHub Issues**: Bug reports and feature requests
- **GitHub Discussions**: Questions and community support
- **Discord**: Real-time community chat (link in README)
- **Documentation**: Comprehensive guides in `docs/`

### Maintainer Contact

- **General Questions**: Create a GitHub Discussion
- **Security Issues**: security@subosity.com
- **Urgent Issues**: Tag maintainers in GitHub issues

## 🙏 Recognition

Contributors will be recognized in:
- **AUTHORS.md** file
- **Release notes** for significant contributions
- **GitHub contributors** widget
- **Annual contributor highlights**

---

**Thank you for contributing to Subosity Installer! Your efforts help make self-hosting accessible to everyone.** 🚀
