package docker

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/subosity/subosity-installer/pkg/errors"
	"github.com/subosity/subosity-installer/pkg/logger"
	"github.com/subosity/subosity-installer/shared/constants"
	"github.com/subosity/subosity-installer/shared/types"
)

// Service handles Docker installation and management
type Service struct {
	logger logger.Logger
}

// NewService creates a new Docker service
func NewService(log logger.Logger) *Service {
	return &Service{
		logger: log,
	}
}

// IsInstalled checks if Docker is installed and running
func (s *Service) IsInstalled(ctx context.Context) (bool, error) {
	// Check if docker command exists
	if _, err := exec.LookPath("docker"); err != nil {
		return false, nil
	}
	
	// Check if Docker daemon is running
	cmd := exec.CommandContext(ctx, "docker", "info")
	if err := cmd.Run(); err != nil {
		return false, nil
	}
	
	return true, nil
}

// Install installs Docker CE and Docker Compose
func (s *Service) Install(ctx context.Context) error {
	s.logger.Info("Installing Docker CE and Docker Compose...")
	
	// Create context with timeout
	installCtx, cancel := context.WithTimeout(ctx, constants.DefaultDockerTimeout)
	defer cancel()
	
	// Step 1: Remove potentially conflicting packages
	if err := s.removeConflictingPackages(installCtx); err != nil {
		return errors.WrapError(err, types.ErrCodeDockerInstall, 
			"failed to remove conflicting packages", "docker", "cleanup")
	}
	
	// Step 2: Setup Docker's apt repository
	if err := s.setupDockerRepository(installCtx); err != nil {
		return errors.WrapError(err, types.ErrCodeDockerInstall, 
			"failed to setup Docker repository", "docker", "repository")
	}
	
	// Step 3: Install Docker packages
	if err := s.installDockerPackages(installCtx); err != nil {
		return errors.WrapError(err, types.ErrCodeDockerInstall, 
			"failed to install Docker packages", "docker", "installation")
	}
	
	// Step 4: Configure Docker daemon and user
	if err := s.configureDocker(installCtx); err != nil {
		return errors.WrapError(err, types.ErrCodeDockerInstall, 
			"failed to configure Docker", "docker", "configuration")
	}
	
	// Step 5: Verify installation
	if err := s.verifyInstallation(installCtx); err != nil {
		return errors.WrapError(err, types.ErrCodeDockerInstall, 
			"Docker installation verification failed", "docker", "verification")
	}
	
	logger.Success(s.logger, "Docker CE and Docker Compose installed successfully")
	return nil
}

// GetVersion returns the installed Docker version
func (s *Service) GetVersion(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "docker", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func (s *Service) removeConflictingPackages(ctx context.Context) error {
	s.logger.Info("Removing potentially conflicting packages...")
	
	packages := []string{
		"docker.io", "docker-doc", "docker-compose", "docker-compose-v2",
		"podman-docker", "containerd", "runc",
	}
	
	args := append([]string{"apt-get", "remove", "-y"}, packages...)
	cmd := exec.CommandContext(ctx, "sudo", args...)
	
	// Capture both stdout and stderr
	output, err := cmd.CombinedOutput()
	if err != nil {
		// It's okay if some packages aren't installed
		s.logger.Debugf("Package removal output: %s", output)
	}
	
	return nil
}

func (s *Service) setupDockerRepository(ctx context.Context) error {
	s.logger.Info("Setting up Docker repository...")
	
	// Update package index
	if err := s.runCommand(ctx, "sudo", "apt-get", "update"); err != nil {
		return fmt.Errorf("failed to update package index: %w", err)
	}
	
	// Install required packages
	if err := s.runCommand(ctx, "sudo", "apt-get", "install", "-y", 
		"ca-certificates", "curl"); err != nil {
		return fmt.Errorf("failed to install required packages: %w", err)
	}
	
	// Create keyrings directory
	if err := s.runCommand(ctx, "sudo", "install", "-m", "0755", "-d", 
		"/etc/apt/keyrings"); err != nil {
		return fmt.Errorf("failed to create keyrings directory: %w", err)
	}
	
	// Download Docker's GPG key
	if err := s.runCommand(ctx, "sudo", "curl", "-fsSL", 
		constants.DockerGPGKeyURL, "-o", constants.DockerGPGKeyPath); err != nil {
		return fmt.Errorf("failed to download Docker GPG key: %w", err)
	}
	
	// Set permissions on GPG key
	if err := s.runCommand(ctx, "sudo", "chmod", "a+r", 
		constants.DockerGPGKeyPath); err != nil {
		return fmt.Errorf("failed to set GPG key permissions: %w", err)
	}
	
	// Add Docker repository
	repoLine := fmt.Sprintf(
		"deb [arch=$(dpkg --print-architecture) signed-by=%s] "+
		"https://download.docker.com/linux/ubuntu "+
		"$(. /etc/os-release && echo \"${UBUNTU_CODENAME:-$VERSION_CODENAME}\") stable",
		constants.DockerGPGKeyPath,
	)
	
	if err := s.runCommand(ctx, "bash", "-c", 
		fmt.Sprintf("echo '%s' | sudo tee %s > /dev/null", repoLine, constants.DockerListPath)); err != nil {
		return fmt.Errorf("failed to add Docker repository: %w", err)
	}
	
	// Update package index again
	if err := s.runCommand(ctx, "sudo", "apt-get", "update"); err != nil {
		return fmt.Errorf("failed to update package index after adding repository: %w", err)
	}
	
	return nil
}

func (s *Service) installDockerPackages(ctx context.Context) error {
	s.logger.Info("Installing Docker packages...")
	
	packages := []string{
		"docker-ce", "docker-ce-cli", "containerd.io",
		"docker-buildx-plugin", "docker-compose-plugin",
	}
	
	args := append([]string{"apt-get", "install", "-y"}, packages...)
	if err := s.runCommand(ctx, "sudo", args...); err != nil {
		return fmt.Errorf("failed to install Docker packages: %w", err)
	}
	
	return nil
}

func (s *Service) configureDocker(ctx context.Context) error {
	s.logger.Info("Configuring Docker...")
	
	// Start and enable Docker service
	if err := s.runCommand(ctx, "sudo", "systemctl", "start", "docker"); err != nil {
		return fmt.Errorf("failed to start Docker service: %w", err)
	}
	
	if err := s.runCommand(ctx, "sudo", "systemctl", "enable", "docker"); err != nil {
		return fmt.Errorf("failed to enable Docker service: %w", err)
	}
	
	// Add current user to docker group
	currentUser := os.Getenv("USER")
	if currentUser == "" {
		currentUser = os.Getenv("LOGNAME")
	}
	
	if currentUser != "" && currentUser != "root" {
		if err := s.runCommand(ctx, "sudo", "usermod", "-aG", "docker", currentUser); err != nil {
			s.logger.Warnf("Failed to add user %s to docker group: %v", currentUser, err)
		} else {
			s.logger.Infof("Added user %s to docker group", currentUser)
		}
	}
	
	return nil
}

func (s *Service) verifyInstallation(ctx context.Context) error {
	s.logger.Info("Verifying Docker installation...")
	
	// Wait a moment for Docker to fully start
	time.Sleep(2 * time.Second)
	
	// Test Docker
	if err := s.runCommand(ctx, "sudo", "docker", "run", "--rm", "hello-world"); err != nil {
		return fmt.Errorf("Docker test run failed: %w", err)
	}
	
	// Test Docker Compose
	cmd := exec.CommandContext(ctx, "docker", "compose", "version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Docker Compose verification failed: %w", err)
	}
	
	return nil
}

func (s *Service) runCommand(ctx context.Context, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	
	// Log the command being run (without sensitive information)
	s.logger.Debugf("Running: %s %s", name, strings.Join(args, " "))
	
	// Capture output for debugging
	output, err := cmd.CombinedOutput()
	if err != nil {
		s.logger.Debugf("Command failed with output: %s", output)
		return fmt.Errorf("command '%s %s' failed: %w", name, strings.Join(args, " "), err)
	}
	
	return nil
}
