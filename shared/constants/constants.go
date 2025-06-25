package constants

import "time"

// Application constants
const (
	AppName    = "subosity-installer"
	AppVersion = "1.0.0-dev"
)

// Container configuration
const (
	ContainerImage    = "subosity/installer:latest"
	ContainerImageDev = "subosity/installer:dev"
)

// Default paths
const (
	DefaultInstallPath = "/opt/subosity"
	DefaultConfigPath  = "/opt/subosity/config.yaml"
	DefaultLogsPath    = "/opt/subosity/logs"
	DefaultBackupPath  = "/opt/subosity/backups"
	DefaultDataPath    = "/opt/subosity/data"
)

// Default timeouts
const (
	DefaultInstallTimeout = 15 * time.Minute
	DefaultDockerTimeout  = 5 * time.Minute
	DefaultHealthTimeout  = 30 * time.Second
)

// System requirements
const (
	MinRAMBytes  = 2 * 1024 * 1024 * 1024  // 2GB
	MinDiskBytes = 10 * 1024 * 1024 * 1024 // 10GB
)

// Required ports
var RequiredPorts = []int{80, 443, 5432, 8000, 3000}

// Supported OS distributions
var SupportedOSes = map[string][]string{
	"ubuntu": {"20.04", "22.04", "24.04"},
	"debian": {"11", "12"},
}

// Docker installation constants
const (
	DockerGPGKeyURL = "https://download.docker.com/linux/ubuntu/gpg"
	DockerGPGKeyPath = "/etc/apt/keyrings/docker.asc"
	DockerListPath = "/etc/apt/sources.list.d/docker.list"
)

// Progress indicators
const (
	ProgressValidation    = 0.1
	ProgressPreparation   = 0.2
	ProgressDependencies  = 0.4
	ProgressSupabase      = 0.6
	ProgressApplication   = 0.8
	ProgressVerification  = 0.9
	ProgressComplete      = 1.0
)
