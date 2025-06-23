# 📏 Style Guide: `subosity-installer`

## 1. Code Style and Formatting

### 1.1 Go Code Formatting

**Mandatory Tools:**
- `gofmt` - Standard Go formatting
- `goimports` - Automatic import management  
- `golangci-lint` - Comprehensive linting with all recommended rules

**Editor Configuration (.editorconfig):**
```ini
root = true

[*]
charset = utf-8
end_of_line = lf
insert_final_newline = true
trim_trailing_whitespace = true

[*.go]
indent_style = tab
indent_size = 4

[*.{yaml,yml,json}]
indent_style = space
indent_size = 2

[*.md]
trim_trailing_whitespace = false
```

### 1.2 Naming Conventions

**Package Names:**
- Short, concise, lowercase
- No underscores or mixed caps
- Prefer single words

```go
// Good
package docker
package config
package logger

// Avoid
package dockerClient
package configUtils
package log_formatter
```

**Function and Method Names:**
- Use camelCase for private functions
- Use PascalCase for public functions
- Be descriptive but concise

```go
// Good
func validateConfiguration(config *Config) error { }
func (s *Service) ProcessInstallation() error { }

// Avoid
func validate(c *Config) error { }  // Too generic
func (s *Service) proc() error { }  // Unclear abbreviation
```

**Variable Names:**
- Use camelCase
- Prefer full words over abbreviations
- Keep scope-appropriate length

```go
// Good
var installationConfig *Config
var userEmail string
var ctx context.Context  // Common abbreviations OK

// Avoid
var cfg *Config          // Unclear abbreviation
var veryLongVariableName string  // Unnecessarily verbose
```

**Constants:**
- Use PascalCase for exported constants
- Use camelCase for private constants
- Group related constants

```go
// Good
const (
    DefaultTimeout = 30 * time.Second
    MaxRetries     = 3
    ConfigPath     = "/opt/subosity/config.yaml"
)

const (
    envDevelopment = "dev"
    envStaging     = "staging"
    envProduction  = "prod"
)

// Avoid
const DEFAULT_TIMEOUT = 30  // Wrong case style
const max_retries = 3       // Wrong case style
```

### 1.3 Import Organization

**Import Groups (in order):**
1. Standard library
2. Third-party packages  
3. Local packages

```go
import (
    // Standard library
    "context"
    "fmt"
    "os"
    "time"
    
    // Third-party packages
    "github.com/sirupsen/logrus"
    "github.com/spf13/cobra"
    "gopkg.in/yaml.v3"
    
    // Local packages
    "github.com/subosity/subosity-installer/internal/domain"
    "github.com/subosity/subosity-installer/pkg/config"
)
```

## 1.3 Container-First Architecture Guidelines

### 1.3.1 Component Separation

**Host Binary (Thin Client) Code:**
- Keep host-side code minimal and focused
- Handle only environment validation, Docker installation, and container orchestration
- Avoid complex business logic in the host binary

```go
// Good - Host binary responsibility
func (h *HostInstaller) Install(ctx context.Context, config *Config) error {
    // 1. Validate environment
    if err := h.validator.ValidateEnvironment(ctx); err != nil {
        return fmt.Errorf("environment validation failed: %w", err)
    }
    
    // 2. Ensure Docker availability
    if err := h.docker.EnsureInstalled(ctx); err != nil {
        return fmt.Errorf("docker setup failed: %w", err)
    }
    
    // 3. Delegate to container
    return h.container.RunInstaller(ctx, config)
}

// Avoid - Complex logic in host binary
func (h *HostInstaller) Install(ctx context.Context, config *Config) error {
    // Don't put Supabase setup, service configuration, etc. in host binary
    supabaseClient := supabase.NewClient(config.SupabaseURL)
    if err := supabaseClient.InitProject(); err != nil {
        return err
    }
    // ... this belongs in the container
}
```

**Container (Smart Logic) Code:**
- All complex installation logic goes in the container
- Handle Supabase setup, service configuration, state management
- Maintain clear interfaces for host communication

```go
// Good - Container responsibility
func (c *ContainerInstaller) Install(ctx context.Context, config *Config) error {
    // Complex installation pipeline
    pipeline := []InstallationStep{
        c.setupSupabase,
        c.configureServices,
        c.setupSSL,
        c.createBackup,
        c.startServices,
        c.runHealthChecks,
    }
    
    return c.executePipeline(ctx, pipeline, config)
}
```

### 1.3.2 Communication Patterns

**Configuration Passing:**
```go
// Use structured configuration objects
type InstallationConfig struct {
    Environment Environment       `json:"environment"`
    Domain      string           `json:"domain"`
    Email       string           `json:"email"`
    Supabase    SupabaseConfig   `json:"supabase"`
    Security    SecurityConfig   `json:"security"`
    
    // Host metadata (populated by host binary)
    HostInfo    HostMetadata     `json:"host_info,omitempty"`
}

// Pass via environment variables or mounted files
func (r *ContainerRunner) RunInstaller(ctx context.Context, config *Config) error {
    configJSON, _ := json.Marshal(config)
    
    return r.docker.Run(ctx, ContainerConfig{
        Image: "subosity/installer:latest",
        Env: map[string]string{
            "SUBOSITY_CONFIG": string(configJSON),
        },
        Volumes: map[string]string{
            "/opt/subosity": "/app/data",
        },
    })
}
```

**Progress Reporting:**
```go
// Container reports progress via structured output
type ProgressUpdate struct {
    Phase    string    `json:"phase"`
    Step     string    `json:"step"`
    Progress float64   `json:"progress"`
    Message  string    `json:"message"`
    Error    *string   `json:"error,omitempty"`
}

// Host binary parses and displays progress
func (r *ContainerRunner) streamProgress(ctx context.Context, reader io.Reader) error {
    scanner := bufio.NewScanner(reader)
    for scanner.Scan() {
        var update ProgressUpdate
        if err := json.Unmarshal(scanner.Bytes(), &update); err == nil {
            r.progressReporter.Update(update)
        }
    }
    return scanner.Err()
}
```

### 1.3.3 Error Handling Across Boundaries

**Container Error Reporting:**
```go
// Container should output structured errors
type InstallationError struct {
    Code      string                 `json:"code"`
    Message   string                `json:"message"`
    Phase     string                `json:"phase"`
    Context   map[string]interface{} `json:"context,omitempty"`
    Timestamp time.Time             `json:"timestamp"`
}

// Always provide actionable error messages
func (e *InstallationError) Error() string {
    return fmt.Sprintf("[%s:%s] %s", e.Phase, e.Code, e.Message)
}
```

**Host Error Processing:**
```go
// Host binary should interpret container errors
func (r *ContainerRunner) processContainerError(exitCode int, stderr string) error {
    var installErr InstallationError
    if err := json.Unmarshal([]byte(stderr), &installErr); err == nil {
        return &installErr
    }
    
    // Fallback for non-structured errors
    return fmt.Errorf("container installation failed (exit %d): %s", exitCode, stderr)
}
```

## 2. Code Quality Standards

### 2.1 Function Design

**Function Length:**
- Maximum 50 lines per function (excluding comments and whitespace)
- Extract complex logic into helper functions
- Use the "newspaper" pattern - important details first

**Function Parameters:**
- Maximum 5 parameters per function
- Use structs for complex parameter sets
- Always pass context as first parameter for operations

```go
// Good
func InstallDocker(ctx context.Context, config DockerConfig) error {
    return installDockerWithRetry(ctx, config, defaultRetryConfig)
}

// Avoid - too many parameters
func InstallDocker(ctx context.Context, version string, timeout time.Duration, 
    retries int, logLevel string, dryRun bool) error {
}

// Better - use options struct
type InstallOptions struct {
    Version   string
    Timeout   time.Duration
    Retries   int
    LogLevel  string
    DryRun    bool
}

func InstallDocker(ctx context.Context, options InstallOptions) error {
}
```

**Return Values:**
- Prefer explicit error returns over panic
- Use Result types for operations that may fail
- Return early to reduce nesting

```go
// Good
func ValidateConfig(config *Config) error {
    if config == nil {
        return errors.New("config cannot be nil")
    }
    
    if config.Domain == "" {
        return errors.New("domain is required")
    }
    
    if !isValidDomain(config.Domain) {
        return fmt.Errorf("invalid domain format: %s", config.Domain)
    }
    
    return nil
}

// Avoid - deeply nested
func ValidateConfig(config *Config) error {
    if config != nil {
        if config.Domain != "" {
            if isValidDomain(config.Domain) {
                return nil
            } else {
                return fmt.Errorf("invalid domain format: %s", config.Domain)
            }
        } else {
            return errors.New("domain is required")
        }
    } else {
        return errors.New("config cannot be nil")
    }
}
```

### 2.2 Error Handling

**Custom Error Types:**
```go
// Define structured error types
type InstallationError struct {
    Code        ErrorCode
    Message     string
    Details     string
    Suggestions []string
    Cause       error
    Context     map[string]interface{}
}

func (e *InstallationError) Error() string {
    return fmt.Sprintf("[%s] %s: %s", e.Code, e.Message, e.Details)
}

func (e *InstallationError) Unwrap() error {
    return e.Cause
}

// Error creation helpers
func NewConfigError(msg string, details string) *InstallationError {
    return &InstallationError{
        Code:    ErrorCodeConfigInvalid,
        Message: msg,
        Details: details,
        Suggestions: []string{
            "Check configuration file syntax",
            "Verify all required fields are present",
        },
    }
}
```

**Error Wrapping:**
```go
// Good - provide context
func (s *DockerService) Install(ctx context.Context) error {
    if err := s.downloadBinary(ctx); err != nil {
        return fmt.Errorf("failed to download Docker binary: %w", err)
    }
    
    if err := s.configureService(); err != nil {
        return fmt.Errorf("failed to configure Docker service: %w", err)
    }
    
    return nil
}

// Avoid - losing context
func (s *DockerService) Install(ctx context.Context) error {
    if err := s.downloadBinary(ctx); err != nil {
        return err  // Lost context about what failed
    }
    
    if err := s.configureService(); err != nil {
        return err  // Lost context about what failed
    }
    
    return nil
}
```

### 2.3 Concurrency and Context

**Context Usage:**
```go
// Good - always check context
func (s *InstallerService) InstallWithTimeout(ctx context.Context) error {
    installCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
    defer cancel()
    
    phases := []func(context.Context) error{
        s.validateSystem,
        s.installDocker,
        s.setupSupabase,
    }
    
    for i, phase := range phases {
        // Check for cancellation before each phase
        select {
        case <-installCtx.Done():
            return installCtx.Err()
        default:
        }
        
        if err := phase(installCtx); err != nil {
            return fmt.Errorf("installation phase %d failed: %w", i, err)
        }
    }
    
    return nil
}

// Goroutine safety
func (s *Service) processAsync(ctx context.Context, items []Item) error {
    var wg sync.WaitGroup
    errorCh := make(chan error, len(items))
    
    for _, item := range items {
        wg.Add(1)
        go func(item Item) {
            defer wg.Done()
            
            if err := s.processItem(ctx, item); err != nil {
                select {
                case errorCh <- err:
                case <-ctx.Done():
                }
            }
        }(item)
    }
    
    wg.Wait()
    close(errorCh)
    
    // Collect errors
    var errors []error
    for err := range errorCh {
        errors = append(errors, err)
    }
    
    if len(errors) > 0 {
        return fmt.Errorf("failed to process %d items: %v", len(errors), errors)
    }
    
    return nil
}
```

## 3. Documentation Standards

### 3.1 Package Documentation

**Package Comments:**
```go
// Package installer provides a comprehensive toolkit for deploying
// and managing Subosity applications. It handles Docker installation,
// Supabase setup, SSL configuration, and service management with
// enterprise-grade reliability and security.
//
// The installer follows a phased approach:
//   1. System validation and requirement checking
//   2. Dependency installation (Docker, Supabase CLI)
//   3. Application deployment and configuration
//   4. Service integration and health verification
//
// Example usage:
//   config := &Config{
//       Environment: "production",
//       Domain:      "my-app.example.com",
//       Email:       "admin@example.com",
//   }
//   
//   installer := NewInstaller(config)
//   if err := installer.Install(ctx); err != nil {
//       log.Fatal(err)
//   }
package installer
```

### 3.2 Function Documentation

**Public Function Comments:**
```go
// InstallDocker downloads and installs Docker CE on the target system.
// It automatically detects the operating system and uses the appropriate
// installation method (apt, yum, etc.).
//
// The installation process includes:
//   - Adding Docker's official repository
//   - Installing Docker CE and Docker Compose
//   - Configuring the Docker daemon with security best practices
//   - Adding the current user to the docker group
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - config: Docker configuration including version preferences
//
// Returns an error if:
//   - The operating system is not supported
//   - Network connectivity issues prevent package downloads
//   - Insufficient privileges for system modification
//   - Package installation fails
//
// Example:
//   config := DockerConfig{
//       Version: "24.0.0",
//       EnableDaemon: true,
//   }
//   
//   if err := InstallDocker(ctx, config); err != nil {
//       return fmt.Errorf("Docker installation failed: %w", err)
//   }
func InstallDocker(ctx context.Context, config DockerConfig) error {
    // Implementation...
}
```

**Complex Function Comments:**
```go
// validateSystemRequirements performs comprehensive system validation
// including OS compatibility, resource availability, and port conflicts.
//
// Validation checks performed:
//   - Operating system and version compatibility
//   - Available RAM (minimum 2GB, recommended 4GB)
//   - Available disk space (minimum 10GB, recommended 50GB)
//   - Required ports availability (80, 443, 5432, 8000, 3000)
//   - Network connectivity to package repositories
//   - User privileges (sudo access for system modifications)
//
// The function returns a ValidationResult containing:
//   - Overall validation status (pass/fail/warning)
//   - Detailed results for each check
//   - Specific error messages and remediation suggestions
//   - Resource usage information for capacity planning
func validateSystemRequirements(ctx context.Context, config *Config) (*ValidationResult, error) {
    // Implementation...
}
```

### 3.3 Type Documentation

**Struct Documentation:**
```go
// Config represents the complete configuration for a Subosity installation.
// It includes all necessary parameters for deployment, security settings,
// and operational preferences.
//
// Configuration sources (in order of precedence):
//   1. Command-line flags
//   2. Environment variables  
//   3. Configuration file
//   4. Default values
//
// Example configuration file:
//   environment: production
//   domain: my-app.example.com
//   email: admin@example.com
//   ssl:
//     provider: letsencrypt
//     auto_renew: true
//   database:
//     backup_retention_days: 30
//     backup_schedule: "0 2 * * *"
type Config struct {
    // Environment specifies the deployment environment.
    // Valid values: "dev", "staging", "prod"
    // Default: "prod"
    Environment Environment `yaml:"environment" validate:"required,oneof=dev staging prod"`
    
    // Domain is the fully qualified domain name where the application
    // will be accessible. Must be a valid FQDN.
    // Example: "my-app.example.com"
    Domain string `yaml:"domain" validate:"required,fqdn"`
    
    // Email is used for Let's Encrypt certificate registration and
    // administrative notifications. Must be a valid email address.
    Email string `yaml:"email" validate:"email"`
    
    // SSL configures TLS/SSL certificate management.
    // See SSLConfig for detailed options.
    SSL SSLConfig `yaml:"ssl"`
    
    // Database configures PostgreSQL settings including backups.
    // See DatabaseConfig for detailed options.
    Database DatabaseConfig `yaml:"database"`
}
```

## 4. Testing Standards

### 4.1 Test Organization

**Test File Structure:**
```go
// installation_test.go
package installer

import (
    "context"
    "testing"
    "time"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/stretchr/testify/mock"
)

// Test naming: Test<FunctionName>_<Scenario>
func TestInstallDocker_SuccessfulInstallation(t *testing.T) {
    // Arrange
    ctx := context.Background()
    mockSystem := &MockSystemAdapter{}
    mockNetwork := &MockNetworkAdapter{}
    
    installer := NewDockerInstaller(mockSystem, mockNetwork)
    
    // Setup expectations
    mockSystem.On("DetectOS").Return(&OSInfo{
        Name:    "ubuntu",
        Version: "22.04",
    }, nil)
    
    // Act
    err := installer.Install(ctx, DockerConfig{})
    
    // Assert
    require.NoError(t, err)
    mockSystem.AssertExpectations(t)
}

func TestInstallDocker_UnsupportedOS(t *testing.T) {
    // Arrange
    ctx := context.Background()
    mockSystem := &MockSystemAdapter{}
    
    installer := NewDockerInstaller(mockSystem, nil)
    
    mockSystem.On("DetectOS").Return(&OSInfo{
        Name:    "windows",
        Version: "10",
    }, nil)
    
    // Act
    err := installer.Install(ctx, DockerConfig{})
    
    // Assert
    require.Error(t, err)
    assert.Contains(t, err.Error(), "unsupported operating system")
    assert.Equal(t, ErrorCodeUnsupportedOS, extractErrorCode(err))
}
```

**Table-Driven Tests:**
```go
func TestValidateDomain(t *testing.T) {
    tests := []struct {
        name        string
        input       string
        expected    string
        expectError bool
        errorCode   ErrorCode
    }{
        {
            name:        "valid_domain",
            input:       "example.com",
            expected:    "example.com",
            expectError: false,
        },
        {
            name:        "valid_subdomain",
            input:       "app.example.com",
            expected:    "app.example.com",
            expectError: false,
        },
        {
            name:        "domain_with_whitespace",
            input:       "  example.com  ",
            expected:    "example.com",
            expectError: false,
        },
        {
            name:        "invalid_domain_no_tld",
            input:       "example",
            expected:    "",
            expectError: true,
            errorCode:   ErrorCodeInvalidFormat,
        },
        {
            name:        "invalid_domain_too_long",
            input:       strings.Repeat("a", 254) + ".com",
            expected:    "",
            expectError: true,
            errorCode:   ErrorCodeMaxLengthExceeded,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := ValidateAndSanitizeDomain(tt.input)
            
            if tt.expectError {
                require.Error(t, err)
                assert.Equal(t, tt.errorCode, extractErrorCode(err))
                assert.Empty(t, result)
            } else {
                require.NoError(t, err)
                assert.Equal(t, tt.expected, result)
            }
        })
    }
}
```

### 4.2 Mock Guidelines

**Mock Implementation:**
```go
// Mock interfaces for testing
type MockDockerAdapter struct {
    mock.Mock
}

func (m *MockDockerAdapter) Install(ctx context.Context, config DockerConfig) error {
    args := m.Called(ctx, config)
    return args.Error(0)
}

func (m *MockDockerAdapter) IsInstalled(ctx context.Context) (bool, error) {
    args := m.Called(ctx)
    return args.Bool(0), args.Error(1)
}

func (m *MockDockerAdapter) GetVersion(ctx context.Context) (string, error) {
    args := m.Called(ctx)
    return args.String(0), args.Error(1)
}

// Test helpers
func createMockDockerAdapter() *MockDockerAdapter {
    mock := &MockDockerAdapter{}
    
    // Default successful behaviors
    mock.On("IsInstalled", mock.Anything).Return(false, nil)
    mock.On("Install", mock.Anything, mock.Anything).Return(nil)
    mock.On("GetVersion", mock.Anything).Return("24.0.0", nil)
    
    return mock
}
```

### 4.3 Integration Testing

**Test Containers Usage:**
```go
func TestFullInstallation_E2E(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }
    
    ctx := context.Background()
    
    // Start test container
    container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: testcontainers.ContainerRequest{
            Image: "ubuntu:22.04",
            Cmd:   []string{"sleep", "3600"},
            Mounts: testcontainers.Mounts{
                testcontainers.BindMount("./dist/subosity-installer", "/usr/local/bin/subosity-installer"),
            },
        },
        Started: true,
    })
    require.NoError(t, err)
    defer container.Terminate(ctx)
    
    // Install required packages in container
    _, err = container.Exec(ctx, []string{"apt-get", "update"})
    require.NoError(t, err)
    
    _, err = container.Exec(ctx, []string{"apt-get", "install", "-y", "curl", "sudo"})
    require.NoError(t, err)
    
    // Run installer
    exitCode, err := container.Exec(ctx, []string{
        "/usr/local/bin/subosity-installer", "setup",
        "--env", "dev",
        "--domain", "test.localhost",
        "--email", "test@example.com",
        "--ssl-provider", "self-signed",
    })
    require.NoError(t, err)
    assert.Equal(t, 0, exitCode)
    
    // Verify installation
    exitCode, err = container.Exec(ctx, []string{
        "/usr/local/bin/subosity-installer", "status",
    })
    require.NoError(t, err)
    assert.Equal(t, 0, exitCode)
}
```

## 5. Performance Guidelines

### 5.1 Memory Management

**Avoid Memory Leaks:**
```go
// Good - properly close resources
func ProcessLargeFile(filename string) error {
    file, err := os.Open(filename)
    if err != nil {
        return err
    }
    defer file.Close()  // Always defer close
    
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        if err := processLine(scanner.Text()); err != nil {
            return err
        }
    }
    
    return scanner.Err()
}

// Avoid - potential memory leak
func ProcessLargeFile(filename string) error {
    data, err := os.ReadFile(filename)  // Loads entire file into memory
    if err != nil {
        return err
    }
    
    lines := strings.Split(string(data), "\n")  // Additional memory allocation
    for _, line := range lines {
        if err := processLine(line); err != nil {
            return err
        }
    }
    
    return nil
}
```

**Efficient String Operations:**
```go
// Good - use strings.Builder for concatenation
func buildConfigFile(sections []ConfigSection) string {
    var builder strings.Builder
    builder.Grow(len(sections) * 100)  // Pre-allocate capacity
    
    for _, section := range sections {
        builder.WriteString(section.Header)
        builder.WriteString("\n")
        builder.WriteString(section.Content)
        builder.WriteString("\n\n")
    }
    
    return builder.String()
}

// Avoid - inefficient string concatenation
func buildConfigFile(sections []ConfigSection) string {
    result := ""
    for _, section := range sections {
        result += section.Header    // Creates new string each time
        result += "\n"
        result += section.Content
        result += "\n\n"
    }
    return result
}
```

### 5.2 I/O Operations

**Buffered I/O:**
```go
// Good - buffered operations
func CopyFileEfficiently(src, dst string) error {
    srcFile, err := os.Open(src)
    if err != nil {
        return err
    }
    defer srcFile.Close()
    
    dstFile, err := os.Create(dst)
    if err != nil {
        return err
    }
    defer dstFile.Close()
    
    // Use buffered I/O
    reader := bufio.NewReader(srcFile)
    writer := bufio.NewWriter(dstFile)
    defer writer.Flush()
    
    buffer := make([]byte, 64*1024)  // 64KB buffer
    _, err = io.CopyBuffer(writer, reader, buffer)
    return err
}
```

## 6. Security Guidelines

### 6.1 Input Validation

**Comprehensive Validation:**
```go
func ValidateUserInput(input UserInput) error {
    // 1. Sanitize input
    input.Domain = strings.TrimSpace(strings.ToLower(input.Domain))
    input.Email = strings.TrimSpace(strings.ToLower(input.Email))
    
    // 2. Validate format
    if !domainRegex.MatchString(input.Domain) {
        return NewValidationError("invalid domain format", ErrorCodeInvalidFormat)
    }
    
    if !emailRegex.MatchString(input.Email) {
        return NewValidationError("invalid email format", ErrorCodeInvalidFormat)
    }
    
    // 3. Validate length constraints
    if len(input.Domain) > 253 {
        return NewValidationError("domain too long", ErrorCodeMaxLengthExceeded)
    }
    
    // 4. Check against prohibited patterns
    for _, prohibited := range prohibitedDomains {
        if strings.Contains(input.Domain, prohibited) {
            return NewValidationError("prohibited domain pattern", ErrorCodeProhibitedPattern)
        }
    }
    
    return nil
}
```

**Command Injection Prevention:**
```go
// Good - use parameterized commands
func RunDockerCommand(args []string) error {
    cmd := exec.Command("docker", args...)
    cmd.Env = getSecureEnvironment()
    
    output, err := cmd.CombinedOutput()
    if err != nil {
        return fmt.Errorf("docker command failed: %w\nOutput: %s", err, output)
    }
    
    return nil
}

// Avoid - vulnerable to injection
func RunDockerCommand(cmdString string) error {
    cmd := exec.Command("sh", "-c", cmdString)  // Dangerous!
    return cmd.Run()
}
```

### 6.2 Secret Management

**Secure Secret Handling:**
```go
// SecureString prevents secrets from being accidentally logged or exposed
type SecureString struct {
    value []byte
}

func NewSecureString(value string) *SecureString {
    return &SecureString{value: []byte(value)}
}

func (s *SecureString) String() string {
    return "[REDACTED]"  // Never expose the actual value
}

func (s *SecureString) Bytes() []byte {
    return s.value
}

func (s *SecureString) Clear() {
    for i := range s.value {
        s.value[i] = 0  // Zero out memory
    }
    s.value = nil
}

// Usage example
func storePassword(password string) error {
    securePassword := NewSecureString(password)
    defer securePassword.Clear()  // Always clear secrets
    
    return secretManager.Store("db_password", securePassword)
}
```

## 7. Build and CI/CD Standards

### 7.1 Build Configuration

**Makefile for Container-First Architecture:**
```makefile
# Build configuration
BINARY_NAME=subosity-installer
CONTAINER_IMAGE=subosity/installer
VERSION?=$(shell git describe --tags --always --dirty)
COMMIT?=$(shell git rev-parse HEAD)
BUILD_TIME?=$(shell date -u '+%Y-%m-%d_%H:%M:%S')

# Go build flags
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.Commit=${COMMIT} -X main.BuildTime=${BUILD_TIME}"
BUILD_FLAGS=-trimpath -mod=readonly

# Default target
.PHONY: all
all: clean lint test build build-container

# Clean build artifacts
.PHONY: clean
clean:
	rm -rf dist/
	go clean -cache
	docker rmi $(CONTAINER_IMAGE):latest || true

# Run linting
.PHONY: lint
lint:
	golangci-lint run --config .golangci.yml

# Run tests
.PHONY: test
test:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Build host binary (thin client)
.PHONY: build
build:
	CGO_ENABLED=0 go build ${BUILD_FLAGS} ${LDFLAGS} -o dist/${BINARY_NAME} ./binary/cmd

# Build container image (smart installer)
.PHONY: build-container
build-container:
	docker build -f container/Dockerfile -t $(CONTAINER_IMAGE):latest .
	docker tag $(CONTAINER_IMAGE):latest $(CONTAINER_IMAGE):$(VERSION)

# Cross-compilation for host binary
.PHONY: build-all
build-all:
	GOOS=linux GOARCH=amd64 go build ${BUILD_FLAGS} ${LDFLAGS} -o dist/${BINARY_NAME}-linux-amd64 ./binary/cmd
	GOOS=linux GOARCH=arm64 go build ${BUILD_FLAGS} ${LDFLAGS} -o dist/${BINARY_NAME}-linux-arm64 ./binary/cmd

# Integration tests with container
.PHONY: test-integration
test-integration: build build-container
	go test -tags=integration -v ./tests/integration/...

# Publish container image
.PHONY: publish
publish: build-container
	docker push $(CONTAINER_IMAGE):$(VERSION)
	docker push $(CONTAINER_IMAGE):latest
```

### 7.2 CI/CD Pipeline

**GitHub Actions Workflow:**
```yaml
name: CI/CD

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4
    
    - name: Setup Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.21'
        
    - name: Cache dependencies
      uses: actions/cache@v3
      with:
        path: ~/go/pkg/mod
        key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
        
    - name: Download dependencies
      run: go mod download
      
    - name: Run linting
      uses: golangci/golangci-lint-action@v3
      with:
        version: latest
        
    - name: Run tests
      run: |
        go test -v -race -coverprofile=coverage.out ./...
        go tool cover -func=coverage.out
        
    - name: Check coverage
      run: |
        COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
        if (( $(echo "$COVERAGE < 85" | bc -l) )); then
          echo "Coverage $COVERAGE% is below minimum 85%"
          exit 1
        fi
        
    - name: Run security scan
      uses: securecodewarrior/github-action-gosec@master
      with:
        args: ./...
```

---

*This style guide ensures consistent, maintainable, and high-quality code across the entire project. All code must adhere to these standards before being merged.*
