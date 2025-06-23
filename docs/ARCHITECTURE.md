# 🏗️ Architecture Guide: `subosity-installer`

## 1. System Architecture Overview

The `subosity-installer` follows a **container-first architecture** with a thin Go binary that validates the environment, installs Docker if needed, and delegates all complex installation logic to a specialized container image.

### 1.1 High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                  Thin Go Binary (Host)                     │
│  • Environment validation                                   │
│  • Docker installation                                      │
│  • Parameter validation                                     │
│  • Container orchestration                                  │
├─────────────────────────────────────────────────────────────┤
│                Container Boundary (Docker)                  │
├─────────────────────────────────────────────────────────────┤
│             Smart Container (subosity/installer)            │
│  • All installation logic                                   │
│  • Supabase setup                                          │
│  • Service configuration                                    │
│  • State management                                         │
│  • Backup/restore operations                               │
└─────────────────────────────────────────────────────────────┘
```

### 1.2 Component Responsibilities

#### **Thin Go Binary (Host-side)**
- **Environment Validation**: Check OS compatibility, system requirements
- **Docker Bootstrap**: Install Docker if not present
- **Parameter Validation**: Validate user inputs and configuration
- **Container Orchestration**: Pull and run the installer container
- **Progress Reporting**: Relay container status to user

#### **Smart Container (Container-side)**
- **Installation Logic**: All Supabase and service setup
- **Configuration Management**: Template processing, file generation
- **State Persistence**: Installation state and rollback data
- **Service Management**: Start/stop/update operations
- **Backup/Restore**: Data protection and recovery

### 1.3 Repository Structure

```
subosity-installer/
├── binary/                 # Thin Go binary source
│   ├── cmd/               # CLI commands for the thin binary
│   │   ├── root.go       # Root command and global flags
│   │   ├── install.go    # Main installation command
│   │   ├── update.go     # Update command
│   │   └── status.go     # Status/health check command
│   │
│   ├── internal/         # Private binary code
│   │   ├── docker/       # Docker installation and management
│   │   ├── validation/   # Environment and parameter validation
│   │   ├── container/    # Container orchestration
│   │   └── progress/     # Progress reporting and UI
│   │
│   └── pkg/             # Shared utilities for binary
│       ├── config/      # Configuration parsing
│       ├── logger/      # Logging utilities
│       └── errors/      # Error handling
│
├── container/            # Smart container source
│   ├── cmd/             # Container entry points
│   │   ├── install.go   # Installation logic
│   │   ├── backup.go    # Backup operations
│   │   ├── restore.go   # Restore operations
│   │   └── update.go    # Update operations
│   │
│   ├── internal/        # Container-specific logic
│   │   ├── supabase/    # Supabase setup and management
│   │   ├── services/    # Service configuration
│   │   ├── templates/   # Configuration templates
│   │   ├── state/       # State management
│   │   └── backup/      # Backup/restore logic
│   │
│   ├── pkg/             # Shared container utilities
│   │   ├── config/      # Configuration management
│   │   ├── filesystem/  # File operations
│   │   └── network/     # Network utilities
│   │
│   └── templates/       # Embedded templates
│       ├── docker-compose.yml.tmpl
│       ├── systemd.service.tmpl
│       └── nginx.conf.tmpl
│
├── shared/              # Shared between binary and container
│   ├── types/          # Common data structures
│   ├── constants/      # Shared constants
│   └── schemas/        # Configuration schemas
│
├── build/              # Build configurations
│   ├── binary/         # Binary build scripts
│   │   ├── goreleaser.yml    # Binary release configuration
│   │   └── build.sh          # Cross-platform build script
│   ├── container/      # Container build scripts
│   │   ├── Dockerfile        # Smart container image
│   │   └── docker-compose.yml
│   └── Makefile        # Build orchestration
│
└── docs/               # Documentation
    ├── PRD.md
    ├── ARCHITECTURE.md
    └── API.md
```

## 2. Design Patterns and Principles

### 2.1 Container-First Architecture

#### **Separation of Concerns**
The architecture cleanly separates host-side concerns from container-side logic:

**Host-side (Thin Binary)**:
```go
type HostInstaller struct {
    dockerService    DockerService
    validator       EnvironmentValidator
    containerRunner ContainerRunner
    progressReporter ProgressReporter
}

func (h *HostInstaller) Install(ctx context.Context, config *Config) error {
    // 1. Validate environment
    if err := h.validator.ValidateEnvironment(ctx); err != nil {
        return fmt.Errorf("environment validation failed: %w", err)
    }
    
    // 2. Ensure Docker is available
    if err := h.dockerService.EnsureInstalled(ctx); err != nil {
        return fmt.Errorf("docker setup failed: %w", err)
    }
    
    // 3. Delegate to container
    return h.containerRunner.RunInstaller(ctx, config)
}
```

**Container-side (Smart Container)**:
```go
type ContainerInstaller struct {
    supabaseService  SupabaseService
    configManager   ConfigManager
    stateManager    StateManager
    serviceManager  ServiceManager
}

func (c *ContainerInstaller) Install(ctx context.Context, config *Config) error {
    // All complex installation logic happens here
    return c.executeInstallationPipeline(ctx, config)
}
```

#### **Command Pattern for Container Operations**
Each container operation is implemented as a command:

```go
type ContainerCommand interface {
    Execute(ctx context.Context, params CommandParams) error
    Validate(params CommandParams) error
    GetDescription() string
}

type InstallCommand struct {
    supabase SupabaseAdapter
    services ServiceAdapter
    state    StateAdapter
}

func (c *InstallCommand) Execute(ctx context.Context, params CommandParams) error {
    // Install Supabase, configure services, manage state
    state := &InstallationState{
        Phase: "starting",
        StartTime: time.Now(),
    }
    
    return c.executeWithStateTracking(ctx, params, state)
}
```

#### **Factory Pattern**
Create service instances with proper dependency injection:

```go
type ServiceFactory interface {
    CreateInstaller(ctx context.Context) (InstallerService, error)
    CreateUpdater(ctx context.Context) (UpdaterService, error)
    CreateBackupService(ctx context.Context) (BackupService, error)
}

type DefaultServiceFactory struct {
    config *Config
    logger Logger
}

func (f *DefaultServiceFactory) CreateInstaller(ctx context.Context) (InstallerService, error) {
    dockerAdapter := docker.NewClient(f.config.Docker)
    supabaseAdapter := supabase.NewCLI(f.config.Supabase)
    
    return services.NewInstaller(dockerAdapter, supabaseAdapter, f.logger), nil
}
```

#### **Strategy Pattern**
Different installation strategies per environment:

```go
type InstallationStrategy interface {
    Install(ctx context.Context, config *Config) error
    Validate(config *Config) error
}

type ProductionStrategy struct {
    // Production-specific dependencies
}

type DevelopmentStrategy struct {
    // Development-specific dependencies
}

func (s *InstallationService) Install(ctx context.Context, config *Config) error {
    strategy := s.getStrategy(config.Environment)
    return strategy.Install(ctx, config)
}
```

#### **Observer Pattern**
Progress reporting and event handling:

```go
type ProgressObserver interface {
    OnProgress(event ProgressEvent)
    OnError(event ErrorEvent)
    OnComplete(event CompleteEvent)
}

type InstallationService struct {
    observers []ProgressObserver
}

func (s *InstallationService) notifyProgress(event ProgressEvent) {
    for _, observer := range s.observers {
        observer.OnProgress(event)
    }
}
```

### 2.2 SOLID Principles Implementation

#### **Single Responsibility Principle (SRP)**
Each service has a single, well-defined responsibility:
- `InstallerService`: Handles installation logic
- `ValidatorService`: Validates system requirements and configuration
- `BackupService`: Manages backup and restore operations
- `MonitorService`: Provides health checks and monitoring

#### **Open/Closed Principle (OCP)**
The system is open for extension but closed for modification through interfaces:

```go
type SystemDetector interface {
    DetectOS() (*OSInfo, error)
    DetectArchitecture() (Architecture, error)
    CheckRequirements() error
}

// Can be extended with new OS support without modifying existing code
type UbuntuDetector struct{}
type DebianDetector struct{}
type CentOSDetector struct{}
```

#### **Liskov Substitution Principle (LSP)**
All implementations of interfaces are fully substitutable:

```go
type FileSystem interface {
    WriteFile(path string, data []byte, perm os.FileMode) error
    ReadFile(path string) ([]byte, error)
    CreateDir(path string, perm os.FileMode) error
}

// Both implementations can be used interchangeably
type LocalFileSystem struct{}
type MockFileSystem struct{}
```

#### **Interface Segregation for Host/Container Boundary**
Separate interfaces for host-side and container-side operations:

```go
// Host-side interfaces (thin binary)
type DockerService interface {
    IsInstalled(ctx context.Context) (bool, error)
    Install(ctx context.Context) error
    GetVersion(ctx context.Context) (string, error)
}

type ContainerRunner interface {
    PullImage(ctx context.Context, image string) error
    RunInstaller(ctx context.Context, config *Config) error
    GetLogs(ctx context.Context, containerID string) ([]string, error)
}

type EnvironmentValidator interface {
    ValidateOS(ctx context.Context) error
    ValidateResources(ctx context.Context) error
    ValidateNetwork(ctx context.Context) error
}

// Container-side interfaces (smart container)
type SupabaseService interface {
    Install(ctx context.Context) error
    InitProject(ctx context.Context, config SupabaseConfig) error
    StartServices(ctx context.Context) error
    RunMigrations(ctx context.Context) error
}

type ServiceManager interface {
    ConfigureNginx(ctx context.Context, config NginxConfig) error
    SetupSystemd(ctx context.Context, services []SystemdService) error
    ConfigureFirewall(ctx context.Context, rules FirewallRules) error
}

type StateManager interface {
    SaveInstallationState(ctx context.Context, state *InstallationState) error
    LoadInstallationState(ctx context.Context) (*InstallationState, error)
    CreateCheckpoint(ctx context.Context, phase string) error
    Rollback(ctx context.Context, checkpointID string) error
}
```

#### **Dependency Injection with Context Boundaries**
Clear separation between host and container dependencies:

```go
// Host-side service composition
type HostServices struct {
    Docker     DockerService
    Validator  EnvironmentValidator
    Container  ContainerRunner
    Logger     Logger
}

// Container-side service composition  
type ContainerServices struct {
    Supabase   SupabaseService
    Services   ServiceManager
    State      StateManager
    Templates  TemplateManager
    Logger     Logger
}
```

## 3. Data Models and Communication Protocol

### 3.1 Host-Container Communication

#### **Configuration Transfer**
The host binary validates and passes configuration to the container:

```go
// Shared configuration structure
type InstallationConfig struct {
    Environment    Environment            `json:"environment" validate:"required,oneof=dev staging prod"`
    Domain         string                `json:"domain" validate:"required,fqdn"`
    Email          string                `json:"email" validate:"required,email"`
    
    // Supabase configuration
    Supabase       SupabaseConfig        `json:"supabase"`
    
    // Service configuration
    Services       ServicesConfig        `json:"services"`
    
    // Security settings
    Security       SecurityConfig        `json:"security"`
    
    // Host metadata (filled by binary)
    HostInfo       HostMetadata          `json:"host_info,omitempty"`
}

type HostMetadata struct {
    OS             string    `json:"os"`
    Architecture   string    `json:"architecture"`
    DockerVersion  string    `json:"docker_version"`
    AvailableRAM   int64     `json:"available_ram"`
    AvailableDisk  int64     `json:"available_disk"`
    Timestamp      time.Time `json:"timestamp"`
}
```

#### **Progress Communication**
The container reports progress back to the host:

```go
type ProgressUpdate struct {
    Phase       string                 `json:"phase"`
    Step        string                `json:"step"`
    Progress    float64               `json:"progress"` // 0.0 to 1.0
    Message     string                `json:"message"`
    Metadata    map[string]interface{} `json:"metadata,omitempty"`
    Timestamp   time.Time             `json:"timestamp"`
}

type InstallationResult struct {
    Success     bool                   `json:"success"`
    Phase       string                `json:"phase"`
    Error       *ErrorDetails         `json:"error,omitempty"`
    Services    map[string]ServiceInfo `json:"services"`
    URLs        map[string]string     `json:"urls"`
    Credentials map[string]string     `json:"credentials,omitempty"`
    Duration    time.Duration         `json:"duration"`
}
```
)

// InstallationConfig represents the complete installation configuration
type InstallationConfig struct {
    Environment Environment    `yaml:"environment" validate:"required,oneof=dev staging prod"`
    Domain      string        `yaml:"domain" validate:"required,fqdn"`
    Email       string        `yaml:"email" validate:"email"`
    SSL         SSLConfig     `yaml:"ssl"`
    Database    DatabaseConfig `yaml:"database"`
    Services    ServicesConfig `yaml:"services"`
    Backup      BackupConfig  `yaml:"backup"`
    CreatedAt   time.Time     `yaml:"created_at"`
    Version     string        `yaml:"version"`
}

// InstallationState tracks the current state of an installation
type InstallationState struct {
    Phase          InstallationPhase `json:"phase"`
    Progress       float64          `json:"progress"`
    CurrentStep    string           `json:"current_step"`
    CompletedSteps []string         `json:"completed_steps"`
    Errors         []InstallError   `json:"errors,omitempty"`
    StartedAt      time.Time        `json:"started_at"`
    UpdatedAt      time.Time        `json:"updated_at"`
    Metadata       map[string]any   `json:"metadata,omitempty"`
}

// InstallationPhase represents the current phase of installation
type InstallationPhase string

const (
    PhaseValidation    InstallationPhase = "validation"
    PhasePreparation   InstallationPhase = "preparation"
    PhaseDependencies  InstallationPhase = "dependencies"
    PhaseSupabase      InstallationPhase = "supabase"
    PhaseApplication   InstallationPhase = "application"
    PhaseConfiguration InstallationPhase = "configuration"
    PhaseVerification  InstallationPhase = "verification"
    PhaseComplete      InstallationPhase = "complete"
    PhaseFailed        InstallationPhase = "failed"
)
```

### 3.2 Configuration Structures

```go
// SSLConfig defines SSL/TLS configuration
type SSLConfig struct {
    Provider    SSLProvider `yaml:"provider" validate:"required,oneof=letsencrypt self-signed custom"`
    Email       string      `yaml:"email,omitempty" validate:"omitempty,email"`
    CustomCert  string      `yaml:"custom_cert,omitempty"`
    CustomKey   string      `yaml:"custom_key,omitempty"`
    AutoRenew   bool        `yaml:"auto_renew" default:"true"`
}

// DatabaseConfig defines database configuration
type DatabaseConfig struct {
    BackupRetentionDays int    `yaml:"backup_retention_days" validate:"min=1,max=365" default:"30"`
    BackupSchedule      string `yaml:"backup_schedule" default:"0 2 * * *"` // Daily at 2 AM
    EnableWAL          bool   `yaml:"enable_wal" default:"true"`
    MaxConnections     int    `yaml:"max_connections" validate:"min=1,max=1000" default:"100"`
}

// ServicesConfig defines service-specific configuration
type ServicesConfig struct {
    Frontend FrontendConfig `yaml:"frontend"`
    Backend  BackendConfig  `yaml:"backend"`
    Auth     AuthConfig     `yaml:"auth"`
    Storage  StorageConfig  `yaml:"storage"`
}
```

### 3.3 Error Types

```go
// InstallError represents a structured error with context
type InstallError struct {
    Code        ErrorCode     `json:"code"`
    Message     string        `json:"message"`
    Details     string        `json:"details,omitempty"`
    Suggestions []string      `json:"suggestions,omitempty"`
    Cause       error         `json:"-"`
    Context     ErrorContext  `json:"context"`
    Timestamp   time.Time     `json:"timestamp"`
}

// ErrorCode represents standardized error codes
type ErrorCode string

const (
    ErrCodeSystemRequirements ErrorCode = "SYSTEM_REQUIREMENTS"
    ErrCodePortConflict      ErrorCode = "PORT_CONFLICT"
    ErrCodeDockerInstall     ErrorCode = "DOCKER_INSTALL_FAILED"
    ErrCodeSupabaseSetup     ErrorCode = "SUPABASE_SETUP_FAILED"
    ErrCodeNetworkTimeout    ErrorCode = "NETWORK_TIMEOUT"
    ErrCodePermissionDenied  ErrorCode = "PERMISSION_DENIED"
    ErrCodeConfigInvalid     ErrorCode = "CONFIG_INVALID"
    ErrCodeBackupFailed      ErrorCode = "BACKUP_FAILED"
    ErrCodeMigrationFailed   ErrorCode = "MIGRATION_FAILED"
)

// ErrorContext provides additional context for errors
type ErrorContext struct {
    Component   string            `json:"component"`
    Operation   string            `json:"operation"`
    Phase       InstallationPhase `json:"phase,omitempty"`
    Environment Environment       `json:"environment,omitempty"`
    Metadata    map[string]any    `json:"metadata,omitempty"`
}
```

## 4. Component Interfaces

### 4.1 Core Service Interfaces

```go
// InstallerService handles the complete installation process
type InstallerService interface {
    Install(ctx context.Context, config *InstallationConfig) error
    ValidatePrerequisites(ctx context.Context, config *InstallationConfig) error
    GetProgress(ctx context.Context) (*InstallationState, error)
    Cancel(ctx context.Context) error
}

// UpdaterService handles system updates
type UpdaterService interface {
    Update(ctx context.Context, options UpdateOptions) error
    CheckForUpdates(ctx context.Context) (*UpdateInfo, error)
    Rollback(ctx context.Context, version string) error
}

// BackupService handles backup and restore operations
type BackupService interface {
    CreateBackup(ctx context.Context, options BackupOptions) (*BackupInfo, error)
    RestoreBackup(ctx context.Context, backupID string) error
    ListBackups(ctx context.Context) ([]*BackupInfo, error)
    DeleteBackup(ctx context.Context, backupID string) error
}
```

### 4.2 Infrastructure Adapters

```go
// DockerAdapter handles Docker operations
type DockerAdapter interface {
    IsInstalled(ctx context.Context) (bool, error)
    Install(ctx context.Context) error
    GetVersion(ctx context.Context) (string, error)
    StartContainer(ctx context.Context, config ContainerConfig) error
```

## 5. Security Architecture

### 5.1 Host-Container Security Boundary

The container-first architecture provides important security benefits:

#### **Isolation and Sandboxing**
```go
type SecurityConfig struct {
    // Container security
    RunAsNonRoot     bool              `json:"run_as_non_root"`
    ReadOnlyRootFS   bool              `json:"readonly_rootfs"`
    DropCapabilities []string          `json:"drop_capabilities"`
    SeccompProfile   string            `json:"seccomp_profile"`
    
    // Network security
    NetworkMode      string            `json:"network_mode"`
    AllowedPorts     []int             `json:"allowed_ports"`
    
    // Volume security
    VolumePermissions map[string]string `json:"volume_permissions"`
}
```

#### **Credential Management**
Sensitive data is handled securely across the host-container boundary:

```go
type CredentialManager struct {
    vault SecretVault
    crypto CryptoService
}

func (cm *CredentialManager) SecurelyPassCredentials(ctx context.Context, creds *Credentials) error {
    // Encrypt credentials before passing to container
    encrypted, err := cm.crypto.Encrypt(creds)
    if err != nil {
        return err
    }
    
    // Use secure environment variable or mounted secret
    return cm.vault.StoreTemporary(ctx, encrypted)
}
```

### 5.2 Runtime Security

#### **Container Resource Limits**
```go
type ResourceLimits struct {
    Memory    string `json:"memory"`     // e.g., "2g"
    CPU       string `json:"cpu"`        // e.g., "1.5"
    DiskSpace string `json:"disk_space"` // e.g., "10g"
    PIDs      int    `json:"pids"`       // Process limit
}
```

#### **Audit and Logging**
All operations are logged for security audit:

```go
type SecurityLogger struct {
    auditLog Logger
    eventCollector EventCollector
}

func (sl *SecurityLogger) LogSecurityEvent(event SecurityEvent) {
    auditEntry := AuditEntry{
        Timestamp:   time.Now(),
        EventType:   event.Type,
        Component:   event.Component,
        User:        event.User,
        Action:      event.Action,
        Resource:    event.Resource,
        Result:      event.Result,
        IPAddress:   event.SourceIP,
        UserAgent:   event.UserAgent,
    }
    
    sl.auditLog.Info("security_event", auditEntry)
    sl.eventCollector.Collect(auditEntry)
}
```
## 6. Performance and Optimization

### 6.1 Container Optimization

#### **Image Optimization**
```dockerfile
# Multi-stage build for minimal container size
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o installer ./cmd/installer

FROM alpine:3.18
RUN apk --no-cache add ca-certificates curl docker-cli
WORKDIR /root/
COPY --from=builder /app/installer .
COPY --from=builder /app/templates ./templates
CMD ["./installer"]
```

#### **Caching Strategy**
```go
type CacheManager struct {
    local  LocalCache
    shared SharedCache
}

func (cm *CacheManager) GetOrCompute(key string, computeFn func() (interface{}, error)) (interface{}, error) {
    // Check local cache first
    if value, exists := cm.local.Get(key); exists {
        return value, nil
    }
    
    // Check shared cache
    if value, exists := cm.shared.Get(key); exists {
        cm.local.Set(key, value, time.Hour) // Cache locally
        return value, nil
    }
    
    // Compute and cache
    value, err := computeFn()
    if err != nil {
        return nil, err
    }
    
    cm.local.Set(key, value, time.Hour)
    cm.shared.Set(key, value, 24*time.Hour)
    return value, nil
}
```

### 6.2 Resource Management

#### **Memory Efficiency**
```go
type ResourceManager struct {
    memoryPool sync.Pool
    bufferPool sync.Pool
}

func (rm *ResourceManager) GetBuffer() *bytes.Buffer {
    if buf := rm.bufferPool.Get(); buf != nil {
        return buf.(*bytes.Buffer)
    }
    return &bytes.Buffer{}
}

func (rm *ResourceManager) PutBuffer(buf *bytes.Buffer) {
    buf.Reset()
    rm.bufferPool.Put(buf)
}
```

## 7. Testing Strategy

### 7.1 Multi-Layer Testing

#### **Host Binary Testing**
```go
func TestHostBinaryValidation(t *testing.T) {
    tests := []struct {
        name    string
        env     HostEnvironment
        wantErr bool
    }{
        {
            name: "valid ubuntu environment",
            env: HostEnvironment{
                OS: "ubuntu",
                Version: "20.04",
                RAM: 4 * 1024 * 1024 * 1024, // 4GB
                Disk: 20 * 1024 * 1024 * 1024, // 20GB
            },
            wantErr: false,
        },
        {
            name: "insufficient resources",
            env: HostEnvironment{
                OS: "ubuntu",
                Version: "20.04",
                RAM: 1 * 1024 * 1024 * 1024, // 1GB - too low
                Disk: 5 * 1024 * 1024 * 1024, // 5GB - too low
            },
            wantErr: true,
        },
    }
    
    validator := NewEnvironmentValidator()
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validator.ValidateEnvironment(context.Background(), tt.env)
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidateEnvironment() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

#### **Container Integration Testing**
```go
func TestContainerInstallation(t *testing.T) {
    // Use testcontainers for integration testing
    ctx := context.Background()
    
    req := testcontainers.ContainerRequest{
        Image: "subosity/installer:test",
        Env: map[string]string{
            "SUBOSITY_ENVIRONMENT": "test",
            "SUBOSITY_DOMAIN": "test.example.com",
        },
        ExposedPorts: []string{"3000/tcp", "5432/tcp"},
    }
    
    container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: req,
        Started: true,
    })
    require.NoError(t, err)
    defer container.Terminate(ctx)
    
    // Test installation completion
    assertInstallationSuccess(t, container)
    
    // Test service availability
    assertServicesRunning(t, container)
}
```

### 7.2 Testing Utilities

#### **Mock Implementations**
```go
type MockDockerService struct {
    InstallCalled bool
    ShouldFail    bool
}

func (m *MockDockerService) IsInstalled(ctx context.Context) (bool, error) {
    return !m.ShouldFail, nil
}

func (m *MockDockerService) Install(ctx context.Context) error {
    m.InstallCalled = true
    if m.ShouldFail {
        return errors.New("mock docker install failure")
    }
    return nil
}
```

#### **Test Data Management**
```go
type TestDataManager struct {
    fixtures map[string]interface{}
    cleanup  []func()
}

func (tdm *TestDataManager) LoadFixture(name string) interface{} {
    return tdm.fixtures[name]
}

func (tdm *TestDataManager) Cleanup() {
    for _, fn := range tdm.cleanup {
        fn()
    }
}
```

---

*This container-first architecture ensures clean separation of concerns, enhanced security through containerization, simplified deployment, and robust testing capabilities. The thin binary handles host-specific concerns while the smart container manages all complex installation logic.*

### 2.3 Supabase Integration Principles

#### **Supabase CLI as the Primary Interface**
The installer acts as an orchestrator around the official Supabase CLI, never replacing or hijacking its functionality:

```go
// ✅ Good - Use Supabase CLI commands
type SupabaseService interface {
    // Wrapper around `supabase init`
    InitProject(ctx context.Context, projectDir string) error
    
    // Wrapper around `supabase start`
    StartServices(ctx context.Context) error
    
    // Wrapper around `supabase db push`
    PushDatabase(ctx context.Context) error
    
    // Wrapper around `supabase gen types`
    GenerateTypes(ctx context.Context) error
    
    // Wrapper around `supabase status`
    GetStatus(ctx context.Context) (*SupabaseStatus, error)
}

// ❌ Avoid - Don't reimplement Supabase functionality
type SupabaseService interface {
    StartPostgres(ctx context.Context) error      // Let supabase CLI handle this
    ConfigureAuth(ctx context.Context) error      // Let supabase CLI handle this
    SetupStorage(ctx context.Context) error       // Let supabase CLI handle this
}
```

#### **Installation Flow with Supabase CLI**
```go
func (s *SupabaseInstaller) Install(ctx context.Context, config *Config) error {
    // 1. Prepare environment for Supabase
    if err := s.prepareDirectory("/opt/subosity/supabase"); err != nil {
        return err
    }
    
    // 2. Install Supabase CLI (official binary)
    if err := s.installSupabaseCLI(ctx); err != nil {
        return err
    }
    
    // 3. Use Supabase CLI to initialize project
    if err := s.runCommand("supabase", "init"); err != nil {
        return err
    }
    
    // 4. Configure project using Supabase CLI
    if err := s.configureSupabaseProject(ctx, config); err != nil {
        return err
    }
    
    // 5. Start services using Supabase CLI
    if err := s.runCommand("supabase", "start"); err != nil {
        return err
    }
    
    // 6. Integrate with our systemd service
    return s.setupSystemdIntegration(ctx)
}
```

#### **Docker Compose Integration Strategy**
The installer creates a unified Docker Compose that includes both Supabase and application services:

```go
type DockerComposeManager struct {
    supabaseCompose string // Path to Supabase's docker-compose.yml
    appCompose      string // Our application services
}

func (dcm *DockerComposeManager) CreateUnifiedCompose(ctx context.Context) error {
    // 1. Let Supabase CLI generate its docker-compose.yml
    if err := dcm.runCommand("supabase", "start", "--debug"); err != nil {
        return err
    }
    
    // 2. Parse Supabase's generated compose file
    supabaseServices, err := dcm.parseSupabaseCompose()
    if err != nil {
        return err
    }
    
    // 3. Add our application services
    unifiedCompose := &DockerCompose{
        Version:  "3.8",
        Services: make(map[string]Service),
        Networks: supabaseServices.Networks, // Reuse Supabase networks
        Volumes:  supabaseServices.Volumes,  // Reuse Supabase volumes
    }
    
    // 4. Copy Supabase services as-is
    for name, service := range supabaseServices.Services {
        unifiedCompose.Services[name] = service
    }
    
    // 5. Add our application services that depend on Supabase
    unifiedCompose.Services["subosity-app"] = Service{
        Image: "subosity/app:latest",
        DependsOn: []string{"supabase-db", "supabase-auth"},
        Environment: []string{
            "SUPABASE_URL=http://supabase-kong:8000",
            "SUPABASE_ANON_KEY=${SUPABASE_ANON_KEY}",
        },
    }
    
    return dcm.writeUnifiedCompose(unifiedCompose)
}
```

#### **Backup and Restore via Supabase CLI**
```go
func (s *SupabaseBackupService) CreateBackup(ctx context.Context) (*BackupInfo, error) {
    // Use Supabase CLI for database backup
    backupFile := fmt.Sprintf("/opt/subosity/backups/supabase_%s.sql", time.Now().Format("20060102_150405"))
    
    // Supabase CLI command: supabase db dump
    cmd := exec.CommandContext(ctx, "supabase", "db", "dump", 
        "--file", backupFile,
        "--exclude-table-data", "auth.sessions") // Exclude sensitive session data
    
    if err := cmd.Run(); err != nil {
        return nil, fmt.Errorf("supabase backup failed: %w", err)
    }
    
    // Also backup our application-specific data
    return s.createApplicationBackup(ctx, backupFile)
}
```

#### **Key Architectural Principles**

1. **Supabase CLI Ownership**: All Supabase operations go through the official CLI
2. **No Supabase Internals**: We never directly manipulate Supabase's internal configuration
3. **Orchestration Only**: Our installer orchestrates the environment, Supabase CLI does the work
4. **Integration Layer**: We provide the glue between Supabase services and our application
5. **Systemd Wrapper**: Our systemd service manages the unified stack, but delegates Supabase operations to CLI

```bash
# Example systemd service that manages the unified stack
[Unit]
Description=Subosity Application Stack
After=docker.service
Requires=docker.service

[Service]
Type=forking
ExecStart=/opt/subosity/bin/start-stack.sh
ExecStop=/opt/subosity/bin/stop-stack.sh
ExecReload=/opt/subosity/bin/reload-stack.sh
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

Where `start-stack.sh` orchestrates both Supabase CLI and our application:
```bash
#!/bin/bash
# Start Supabase services via CLI
cd /opt/subosity/supabase
supabase start

# Start our application services
cd /opt/subosity
docker-compose up -d subosity-app nginx-proxy
```
