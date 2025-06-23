# 🤖 GitHub Copilot Instructions: Subosity Installer

## Project Overview

You are working on the **Subosity Installer**, a production-ready, container-first deployment tool for self-hosting the Subosity application. This project follows enterprise-grade development practices with a focus on reliability, security, and maintainability.

## 📚 Required Reading - Core Documentation

Before suggesting any code, familiarize yourself with these documents:

1. **[📋 docs/PRD.md](../docs/PRD.md)** - Complete product requirements, user stories, and acceptance criteria
2. **[🏗️ docs/ARCHITECTURE.md](../docs/ARCHITECTURE.md)** - System architecture, design patterns, and component interfaces
3. **[📏 STYLE_GUIDE.md](../STYLE_GUIDE.md)** - Coding standards, conventions, and best practices
4. **[🚀 README.md](../README.md)** - Project overview, installation methods, and user documentation
5. **[📊 ROADMAP.md](../ROADMAP.md)** - Development phases, timelines, and feature priorities

## 🏛️ Architectural Principles

### Container-First Architecture
- **Thin Go Binary** (~5MB): Environment validation, Docker installation, container orchestration
- **Smart Container** (subosity/installer:latest): All complex installation logic, Supabase setup, service configuration
- **Clear Separation**: Host-side handles prerequisites, container-side handles business logic

### Supabase Integration
- **Orchestration Only**: Use official Supabase CLI commands, never reimplement Supabase functionality
- **CLI Wrapper**: All Supabase operations go through `supabase` CLI commands
- **Integration Layer**: Provide glue between Supabase services and our application
- **Respect Boundaries**: Never directly manipulate Supabase's internal configuration

## 🎯 Development Standards

### Code Quality Requirements
- **Zero Technical Debt**: Aggressively refactor and improve code quality
- **Production-Ready**: Every component must be enterprise-grade and battle-tested
- **Idempotent Operations**: All installation and configuration operations must be safely repeatable
- **Resilient Design**: Handle failures gracefully with automatic recovery and rollback capabilities
- **Component-Based**: Clear separation of concerns with well-defined interfaces
- **Easily Modifiable**: Code should be self-documenting and simple to extend

### Error Handling
```go
// ✅ Always provide structured errors with context
type InstallationError struct {
    Code        ErrorCode     `json:"code"`
    Message     string        `json:"message"`
    Context     ErrorContext  `json:"context"`
    Suggestions []string      `json:"suggestions"`
}

// ✅ Wrap errors with meaningful context
if err := dockerService.Install(ctx); err != nil {
    return fmt.Errorf("failed to install Docker CE: %w", err)
}
```

### Interface Design
```go
// ✅ Small, focused interfaces
type DockerService interface {
    IsInstalled(ctx context.Context) (bool, error)
    Install(ctx context.Context) error
    GetVersion(ctx context.Context) (string, error)
}

// ❌ Avoid large, monolithic interfaces
type MegaService interface {
    // Too many responsibilities
}
```

### Testing Standards
```go
// ✅ Table-driven tests with comprehensive coverage
func TestValidateEnvironment(t *testing.T) {
    tests := []struct {
        name    string
        env     Environment
        wantErr bool
        errCode ErrorCode
    }{
        {
            name: "valid ubuntu environment",
            env: Environment{OS: "ubuntu", Version: "20.04"},
            wantErr: false,
        },
        {
            name: "unsupported OS",
            env: Environment{OS: "windows", Version: "11"},
            wantErr: true,
            errCode: ErrCodeUnsupportedOS,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateEnvironment(ctx, tt.env)
            // Test implementation
        })
    }
}
```

## 🚫 What NOT to Do

### Anti-Patterns to Avoid
- ❌ **Monolithic Functions**: Keep functions under 50 lines
- ❌ **Global State**: Use dependency injection instead
- ❌ **Magic Numbers**: Use named constants
- ❌ **Silent Failures**: Always handle and log errors appropriately
- ❌ **Supabase Reimplementation**: Never bypass or reimplement Supabase CLI functionality
- ❌ **Container Binary**: Don't containerize the thin Go binary (it installs Docker!)
- ❌ **Technical Debt**: Don't compromise on code quality for speed

### Security Anti-Patterns
- ❌ **Hardcoded Secrets**: Use secure credential management
- ❌ **Root Execution**: Run with minimal required privileges
- ❌ **Unvalidated Input**: Always validate and sanitize user inputs
- ❌ **Insecure Defaults**: Default to secure configurations

## ✅ Best Practices to Follow

### Code Organization
```go
// ✅ Clear package structure
package installer

import (
    "context"
    "fmt"
    
    "github.com/subosity/subosity-installer/internal/docker"
    "github.com/subosity/subosity-installer/pkg/config"
)

// ✅ Dependency injection
func NewInstaller(docker docker.Service, config *config.Config) *Installer {
    return &Installer{
        docker: docker,
        config: config,
    }
}
```

### Configuration Management
```go
// ✅ Structured configuration with validation
type Config struct {
    Environment Environment `yaml:"environment" validate:"required,oneof=dev staging prod"`
    Domain      string     `yaml:"domain" validate:"required,fqdn"`
    Email       string     `yaml:"email" validate:"required,email"`
}

// ✅ Environment-specific defaults
func (c *Config) ApplyDefaults(env Environment) {
    switch env {
    case EnvironmentProduction:
        c.SSL.Provider = "letsencrypt"
        c.Backup.RetentionDays = 30
    case EnvironmentDevelopment:
        c.SSL.Provider = "self-signed"
        c.Backup.RetentionDays = 7
    }
}
```

### Progress Reporting
```go
// ✅ Structured progress updates
type ProgressUpdate struct {
    Phase    string    `json:"phase"`
    Step     string    `json:"step"`
    Progress float64   `json:"progress"` // 0.0 to 1.0
    Message  string    `json:"message"`
}

// ✅ Observable operations
func (s *InstallationService) Install(ctx context.Context, config *Config) error {
    s.notifyProgress(ProgressUpdate{
        Phase:    "validation",
        Step:     "system_requirements",
        Progress: 0.1,
        Message:  "Validating system requirements...",
    })
    
    // Implementation
}
```

### Resource Management
```go
// ✅ Proper resource cleanup
func (s *Service) ProcessFiles(ctx context.Context, files []string) error {
    for _, file := range files {
        f, err := os.Open(file)
        if err != nil {
            return err
        }
        defer f.Close() // Always clean up resources
        
        // Process file
    }
    return nil
}

// ✅ Context-aware operations
func (s *Service) LongRunningOperation(ctx context.Context) error {
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
        // Continue operation
    }
}
```

## 🔍 Code Review Checklist

When suggesting code, ensure it meets these criteria:

### Functionality
- [ ] Implements requirements from PRD.md
- [ ] Follows architectural patterns from ARCHITECTURE.md
- [ ] Adheres to coding standards from STYLE_GUIDE.md
- [ ] Operations are idempotent and can be safely retried
- [ ] Proper error handling with structured error types
- [ ] Comprehensive input validation

### Quality
- [ ] Functions are focused and under 50 lines
- [ ] Clear, self-documenting variable and function names
- [ ] No magic numbers or hardcoded values
- [ ] Proper separation of concerns
- [ ] Dependencies injected, not imported globally
- [ ] Comprehensive test coverage

### Security
- [ ] Input validation and sanitization
- [ ] Secure credential handling
- [ ] Principle of least privilege
- [ ] Audit logging for security events
- [ ] No sensitive data in logs

### Performance
- [ ] Efficient resource usage
- [ ] Proper caching where appropriate
- [ ] Memory leaks prevented
- [ ] Concurrent operations are thread-safe

## 🎯 Specific Guidance

### When Working on Host Binary (thin client):
- Focus on environment validation, Docker installation, and container orchestration
- Keep business logic minimal - delegate complex operations to container
- Ensure robust error reporting back to user
- Handle Docker installation across different Linux distributions

### When Working on Container (smart installer):
- Implement all complex installation logic here
- Use Supabase CLI commands exclusively for Supabase operations
- Maintain installation state for rollback capabilities
- Provide structured progress updates
- Handle service configuration and systemd integration

### When Adding New Features:
1. Update PRD.md with requirements and acceptance criteria
2. Design interfaces following ARCHITECTURE.md patterns
3. Implement following STYLE_GUIDE.md conventions
4. Add comprehensive tests
5. Update README.md documentation

## 🚀 Remember

This is a **production-grade enterprise tool** used by system administrators and DevOps teams. Every line of code should reflect this responsibility:

- **Reliability First**: Users depend on this for critical infrastructure
- **Security Always**: Handle secrets, permissions, and access carefully
- **Documentation**: Code should be self-documenting and well-commented
- **Maintainability**: Future developers should easily understand and modify the code
- **User Experience**: Provide clear feedback, helpful error messages, and intuitive interfaces

When in doubt, prioritize **code quality**, **security**, and **user experience** over development speed.
