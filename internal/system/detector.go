package system

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/subosity/subosity-installer/pkg/errors"
	"github.com/subosity/subosity-installer/pkg/logger"
	"github.com/subosity/subosity-installer/shared/constants"
	"github.com/subosity/subosity-installer/shared/types"
)

// Detector handles system detection and validation
type Detector struct {
	logger logger.Logger
}

// NewDetector creates a new system detector
func NewDetector(log logger.Logger) *Detector {
	return &Detector{
		logger: log,
	}
}

// DetectSystem detects the host system information
func (d *Detector) DetectSystem(ctx context.Context) (*types.HostMetadata, error) {
	d.logger.Info("Detecting system environment...")
	
	osInfo, err := d.detectOS()
	if err != nil {
		return nil, errors.WrapError(err, types.ErrCodeSystemRequirements, 
			"failed to detect operating system", "system", "detection")
	}
	
	arch := runtime.GOARCH
	if arch == "amd64" {
		arch = "x86_64"
	}
	
	ram, err := d.getAvailableRAM()
	if err != nil {
		d.logger.Warnf("Could not detect RAM: %v", err)
		ram = 0
	}
	
	disk, err := d.getAvailableDisk()
	if err != nil {
		d.logger.Warnf("Could not detect disk space: %v", err)
		disk = 0
	}
	
	dockerVersion := ""
	if version, err := d.getDockerVersion(); err == nil {
		dockerVersion = version
	}
	
	return &types.HostMetadata{
		OS:             osInfo.Name,
		Version:        osInfo.Version,
		Architecture:   arch,
		DockerVersion:  dockerVersion,
		AvailableRAM:   ram,
		AvailableDisk:  disk,
		Timestamp:      time.Now(),
	}, nil
}

// ValidateSystemRequirements validates that the system meets minimum requirements
func (d *Detector) ValidateSystemRequirements(ctx context.Context, metadata *types.HostMetadata) error {
	d.logger.Info("Validating system requirements...")
	
	// Check OS compatibility
	if err := d.validateOSCompatibility(metadata.OS, metadata.Version); err != nil {
		return err
	}
	
	// Check architecture
	if err := d.validateArchitecture(metadata.Architecture); err != nil {
		return err
	}
	
	// Check RAM
	if err := d.validateRAM(metadata.AvailableRAM); err != nil {
		return err
	}
	
	// Check disk space
	if err := d.validateDiskSpace(metadata.AvailableDisk); err != nil {
		return err
	}
	
	// Check port availability
	if err := d.validatePortAvailability(ctx); err != nil {
		return err
	}
	
	// Check permissions
	if err := d.validatePermissions(); err != nil {
		return err
	}
	
	d.logger.Info("System requirements validation passed")
	return nil
}

// OSInfo represents operating system information
type OSInfo struct {
	Name    string
	Version string
}

func (d *Detector) detectOS() (*OSInfo, error) {
	// Read /etc/os-release for modern Linux distributions
	file, err := os.Open("/etc/os-release")
	if err != nil {
		return nil, fmt.Errorf("could not read /etc/os-release: %w", err)
	}
	defer file.Close()
	
	osInfo := &OSInfo{}
	scanner := bufio.NewScanner(file)
	
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "ID=") {
			osInfo.Name = strings.Trim(strings.TrimPrefix(line, "ID="), "\"")
		} else if strings.HasPrefix(line, "VERSION_ID=") {
			osInfo.Version = strings.Trim(strings.TrimPrefix(line, "VERSION_ID="), "\"")
		}
	}
	
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading /etc/os-release: %w", err)
	}
	
	if osInfo.Name == "" {
		return nil, fmt.Errorf("could not determine OS name")
	}
	
	return osInfo, nil
}

func (d *Detector) validateOSCompatibility(osName, version string) error {
	supportedVersions, exists := constants.SupportedOSes[osName]
	if !exists {
		return errors.NewSystemError(
			"unsupported operating system",
			fmt.Sprintf("OS '%s' is not supported", osName),
			[]string{
				"Use a supported Linux distribution (Ubuntu 20.04+, Debian 11+)",
				"Check the documentation for the full list of supported systems",
			},
		)
	}
	
	// Check if version is supported
	versionSupported := false
	for _, supportedVersion := range supportedVersions {
		if version == supportedVersion || strings.HasPrefix(version, supportedVersion+".") {
			versionSupported = true
			break
		}
	}
	
	if !versionSupported {
		return errors.NewSystemError(
			"unsupported OS version",
			fmt.Sprintf("OS version '%s %s' is not supported", osName, version),
			[]string{
				fmt.Sprintf("Use a supported version of %s: %v", osName, supportedVersions),
				"Upgrade your operating system to a supported version",
			},
		)
	}
	
	return nil
}

func (d *Detector) validateArchitecture(arch string) error {
	if arch != "x86_64" && arch != "aarch64" && arch != "arm64" {
		return errors.NewSystemError(
			"unsupported architecture",
			fmt.Sprintf("Architecture '%s' is not supported", arch),
			[]string{
				"Use a system with x86_64 (amd64) or ARM64 (aarch64) architecture",
				"Check your system architecture with: uname -m",
			},
		)
	}
	return nil
}

func (d *Detector) validateRAM(availableRAM int64) error {
	if availableRAM > 0 && availableRAM < constants.MinRAMBytes {
		return errors.NewSystemError(
			"insufficient RAM",
			fmt.Sprintf("Available RAM: %d MB, Required: %d MB", 
				availableRAM/(1024*1024), constants.MinRAMBytes/(1024*1024)),
			[]string{
				"Add more RAM to your system",
				"Close other applications to free up memory",
				"Consider using a system with at least 2GB RAM",
			},
		)
	}
	return nil
}

func (d *Detector) validateDiskSpace(availableDisk int64) error {
	if availableDisk > 0 && availableDisk < constants.MinDiskBytes {
		return errors.NewSystemError(
			"insufficient disk space",
			fmt.Sprintf("Available disk space: %d GB, Required: %d GB", 
				availableDisk/(1024*1024*1024), constants.MinDiskBytes/(1024*1024*1024)),
			[]string{
				"Free up disk space by removing unnecessary files",
				"Consider using a system with at least 10GB free space",
				"Move the installation to a different partition with more space",
			},
		)
	}
	return nil
}

func (d *Detector) validatePortAvailability(ctx context.Context) error {
	for _, port := range constants.RequiredPorts {
		if !d.isPortAvailable(port) {
			return errors.NewSystemError(
				"port conflict",
				fmt.Sprintf("Port %d is already in use", port),
				[]string{
					fmt.Sprintf("Stop the service using port %d", port),
					fmt.Sprintf("Use 'sudo netstat -tlnp | grep :%d' to identify the process", port),
					"Consider changing the configuration to use different ports",
				},
			)
		}
	}
	return nil
}

func (d *Detector) validatePermissions() error {
	// Check if we can run sudo
	cmd := exec.Command("sudo", "-n", "true")
	if err := cmd.Run(); err != nil {
		return errors.NewSystemError(
			"insufficient privileges",
			"sudo access is required for system modifications",
			[]string{
				"Run the installer with sudo privileges",
				"Ensure your user is in the sudo group",
				"Configure passwordless sudo for your user",
			},
		)
	}
	return nil
}

func (d *Detector) getAvailableRAM() (int64, error) {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return 0, err
	}
	defer file.Close()
	
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "MemAvailable:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				kb, err := strconv.ParseInt(fields[1], 10, 64)
				if err != nil {
					return 0, err
				}
				return kb * 1024, nil // Convert KB to bytes
			}
		}
	}
	
	return 0, fmt.Errorf("could not find MemAvailable in /proc/meminfo")
}

func (d *Detector) getAvailableDisk() (int64, error) {
	var stat syscall.Statfs_t
	err := syscall.Statfs("/", &stat)
	if err != nil {
		return 0, err
	}
	
	// Available space = block size * available blocks
	return int64(stat.Bavail) * int64(stat.Bsize), nil
}

func (d *Detector) isPortAvailable(port int) bool {
	timeout := time.Second
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("localhost:%d", port), timeout)
	if err != nil {
		return true // Port is available if we can't connect
	}
	conn.Close()
	return false // Port is in use
}

func (d *Detector) getDockerVersion() (string, error) {
	cmd := exec.Command("docker", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	
	// Parse Docker version from output like "Docker version 24.0.0, build ..."
	re := regexp.MustCompile(`Docker version ([^,\s]+)`)
	matches := re.FindStringSubmatch(string(output))
	if len(matches) > 1 {
		return matches[1], nil
	}
	
	return "", fmt.Errorf("could not parse Docker version")
}
