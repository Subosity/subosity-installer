package types

import "time"

// Environment represents the deployment environment
type Environment string

const (
	EnvironmentDevelopment Environment = "dev"
	EnvironmentStaging     Environment = "staging"
	EnvironmentProduction  Environment = "prod"
)

// InstallationConfig represents the complete configuration for a Subosity installation
type InstallationConfig struct {
	Environment Environment    `yaml:"environment" json:"environment"`
	Domain      string        `yaml:"domain" json:"domain"`
	Email       string        `yaml:"email" json:"email"`
	SSL         SSLConfig     `yaml:"ssl" json:"ssl"`
	CreatedAt   time.Time     `yaml:"created_at" json:"created_at"`
	Version     string        `yaml:"version" json:"version"`
}

// SSLConfig defines SSL/TLS configuration
type SSLConfig struct {
	Provider    SSLProvider `yaml:"provider" json:"provider"`
	Email       string      `yaml:"email,omitempty" json:"email,omitempty"`
	CustomCert  string      `yaml:"custom_cert,omitempty" json:"custom_cert,omitempty"`
	CustomKey   string      `yaml:"custom_key,omitempty" json:"custom_key,omitempty"`
	AutoRenew   bool        `yaml:"auto_renew" json:"auto_renew"`
}

// SSLProvider represents SSL certificate providers
type SSLProvider string

const (
	SSLProviderLetsEncrypt SSLProvider = "letsencrypt"
	SSLProviderSelfSigned  SSLProvider = "self-signed"
	SSLProviderCustom      SSLProvider = "custom"
)

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

// ProgressUpdate represents a progress update from the container
type ProgressUpdate struct {
	Phase       string                 `json:"phase"`
	Step        string                 `json:"step"`
	Progress    float64                `json:"progress"` // 0.0 to 1.0
	Message     string                 `json:"message"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
}

// InstallationResult represents the final result of an installation
type InstallationResult struct {
	Success     bool                   `json:"success"`
	Phase       string                 `json:"phase"`
	Error       *ErrorDetails          `json:"error,omitempty"`
	Services    map[string]ServiceInfo `json:"services"`
	URLs        map[string]string      `json:"urls"`
	Credentials map[string]string      `json:"credentials,omitempty"`
	Duration    time.Duration          `json:"duration"`
}

// ServiceInfo represents information about a running service
type ServiceInfo struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Port    int    `json:"port,omitempty"`
	URL     string `json:"url,omitempty"`
	Healthy bool   `json:"healthy"`
}

// ErrorDetails provides detailed error information
type ErrorDetails struct {
	Code        string                 `json:"code"`
	Message     string                 `json:"message"`
	Details     string                 `json:"details,omitempty"`
	Suggestions []string               `json:"suggestions,omitempty"`
	Context     map[string]interface{} `json:"context,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
}

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
	ErrCodeUnsupportedOS     ErrorCode = "UNSUPPORTED_OS"
	ErrCodeInvalidFormat     ErrorCode = "INVALID_FORMAT"
)

// ErrorContext provides additional context for errors
type ErrorContext struct {
	Component   string            `json:"component"`
	Operation   string            `json:"operation"`
	Phase       InstallationPhase `json:"phase,omitempty"`
	Environment Environment       `json:"environment,omitempty"`
	Metadata    map[string]any    `json:"metadata,omitempty"`
}

// HostMetadata contains information about the host system
type HostMetadata struct {
	OS             string    `json:"os"`
	Version        string    `json:"version"`
	Architecture   string    `json:"architecture"`
	DockerVersion  string    `json:"docker_version,omitempty"`
	AvailableRAM   int64     `json:"available_ram"`
	AvailableDisk  int64     `json:"available_disk"`
	Timestamp      time.Time `json:"timestamp"`
}

func (e *InstallError) Error() string {
	if e.Details != "" {
		return e.Message + ": " + e.Details
	}
	return e.Message
}

func (e *InstallError) Unwrap() error {
	return e.Cause
}
