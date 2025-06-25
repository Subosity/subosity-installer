package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/spf13/cobra"
	"github.com/subosity/subosity-installer/shared/constants"
	"github.com/subosity/subosity-installer/shared/types"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check the status of Subosity installation",
	Long: `Check the status of the Subosity installation including all services,
health checks, and system information.

This command provides a comprehensive overview of:
- Installation state
- Service status (running, stopped, error)
- System health checks
- Resource usage
- Recent logs`,
	RunE: runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	log.Info("Checking Subosity installation status...")
	
	// Check if installation directory exists
	if !isInstalled() {
		log.Error("Subosity is not installed")
		log.Info("Run 'subosity-installer setup' to install Subosity")
		return fmt.Errorf("installation not found")
	}
	
	// Get status from container (if available)
	status, err := getContainerStatus(ctx)
	if err != nil {
		log.Warnf("Could not get detailed status: %v", err)
		return showBasicStatus()
	}
	
	return displayStatus(status)
}

func isInstalled() bool {
	_, err := os.Stat(constants.DefaultInstallPath)
	return err == nil
}

func getContainerStatus(ctx context.Context) (*types.InstallationResult, error) {
	// Try to get status from the installer container
	args := []string{
		"run", "--rm",
		"-v", fmt.Sprintf("%s:/app/data", constants.DefaultInstallPath),
		"-v", "/var/run/docker.sock:/var/run/docker.sock",
		constants.ContainerImageDev,
		"status",
	}
	
	cmd := exec.CommandContext(ctx, "docker", args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	
	var status types.InstallationResult
	if err := json.Unmarshal(output, &status); err != nil {
		return nil, err
	}
	
	return &status, nil
}

func showBasicStatus() error {
	log.Info("Installation directory found at " + constants.DefaultInstallPath)
	
	// Check if systemd service exists
	if _, err := os.Stat("/etc/systemd/system/subosity.service"); err == nil {
		log.Info("Systemd service is configured")
		
		// Check service status
		cmd := exec.Command("systemctl", "is-active", "subosity")
		if err := cmd.Run(); err == nil {
			log.Info("✅ Subosity service is running")
		} else {
			log.Warn("❌ Subosity service is not running")
		}
	} else {
		log.Warn("Systemd service not configured")
	}
	
	// Check Docker containers
	cmd := exec.Command("docker", "ps", "--filter", "name=subosity", "--format", "table {{.Names}}\t{{.Status}}")
	output, err := cmd.Output()
	if err == nil && len(output) > 0 {
		log.Info("Docker containers:")
		fmt.Print(string(output))
	} else {
		log.Warn("No Subosity containers found")
	}
	
	return nil
}

func displayStatus(status *types.InstallationResult) error {
	log.Info("📊 Subosity Status Report")
	log.Info("========================")
	
	if status.Success {
		log.Info("✅ Overall Status: Healthy")
	} else {
		log.Warn("❌ Overall Status: Issues Detected")
	}
	
	// Display services
	if len(status.Services) > 0 {
		log.Info("")
		log.Info("Services:")
		for name, service := range status.Services {
			statusIcon := "❌"
			if service.Healthy {
				statusIcon = "✅"
			}
			
			serviceInfo := fmt.Sprintf("  %s %s (%s)", statusIcon, name, service.Status)
			if service.URL != "" {
				serviceInfo += fmt.Sprintf(" - %s", service.URL)
			}
			log.Info(serviceInfo)
		}
	}
	
	// Display URLs
	if len(status.URLs) > 0 {
		log.Info("")
		log.Info("Access URLs:")
		for name, url := range status.URLs {
			log.Infof("  %s: %s", name, url)
		}
	}
	
	// Display any errors
	if status.Error != nil {
		log.Info("")
		log.Error("Recent Issues:")
		log.Errorf("  %s: %s", status.Error.Code, status.Error.Message)
		if status.Error.Details != "" {
			log.Errorf("  Details: %s", status.Error.Details)
		}
	}
	
	return nil
}
