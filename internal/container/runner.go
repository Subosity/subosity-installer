package container

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/subosity/subosity-installer/pkg/errors"
	"github.com/subosity/subosity-installer/pkg/logger"
	"github.com/subosity/subosity-installer/shared/constants"
	"github.com/subosity/subosity-installer/shared/types"
)

// Runner handles container orchestration for the installer
type Runner struct {
	logger logger.Logger
}

// NewRunner creates a new container runner
func NewRunner(log logger.Logger) *Runner {
	return &Runner{
		logger: log,
	}
}

// RunInstaller executes the installer container with the given configuration
func (r *Runner) RunInstaller(ctx context.Context, config *types.InstallationConfig) error {
	r.logger.Info("Starting container-based installation...")
	
	// Ensure the data directory exists
	if err := r.ensureDataDirectory(); err != nil {
		return errors.WrapError(err, types.ErrCodePermissionDenied,
			"failed to create data directory", "container", "preparation")
	}
	
	// Pull the installer image
	if err := r.pullImage(ctx); err != nil {
		return errors.WrapError(err, types.ErrCodeNetworkTimeout,
			"failed to pull installer image", "container", "pull")
	}
	
	// Run the container
	if err := r.runContainer(ctx, config); err != nil {
		return err
	}
	
	logger.Success(r.logger, "Container-based installation completed successfully")
	return nil
}

func (r *Runner) ensureDataDirectory() error {
	dataDir := constants.DefaultInstallPath
	
	// Create the directory with proper permissions
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dataDir, err)
	}
	
	// Create subdirectories
	subdirs := []string{"data", "logs", "configs", "backups", "docker"}
	for _, subdir := range subdirs {
		path := fmt.Sprintf("%s/%s", dataDir, subdir)
		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", path, err)
		}
	}
	
	r.logger.Debugf("Created data directory structure at %s", dataDir)
	return nil
}

func (r *Runner) pullImage(ctx context.Context) error {
	r.logger.Info("Pulling installer container image...")
	
	// Use development image for now (Phase 1)
	image := constants.ContainerImageDev
	
	cmd := exec.CommandContext(ctx, "docker", "pull", image)
	
	// Create a progress reader
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}
	
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start docker pull: %w", err)
	}
	
	// Stream output
	go r.streamOutput(stdout, "[PULL]")
	go r.streamOutput(stderr, "[PULL]")
	
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("docker pull failed: %w", err)
	}
	
	r.logger.Info("Successfully pulled installer container image")
	return nil
}

func (r *Runner) runContainer(ctx context.Context, config *types.InstallationConfig) error {
	r.logger.Info("Running installer container...")
	
	// Serialize configuration to JSON
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to serialize configuration: %w", err)
	}
	
	// Build docker run command
	args := []string{
		"run", "--rm",
		"-v", fmt.Sprintf("%s:/app/data", constants.DefaultInstallPath),
		"-v", "/var/run/docker.sock:/var/run/docker.sock",
		"-v", "/etc:/app/host-etc:ro",
		"-v", "/usr/bin/systemctl:/usr/bin/systemctl:ro",
		"-e", fmt.Sprintf("SUBOSITY_CONFIG=%s", string(configJSON)),
		"-e", fmt.Sprintf("SUBOSITY_ENVIRONMENT=%s", config.Environment),
		"-e", fmt.Sprintf("SUBOSITY_DOMAIN=%s", config.Domain),
		"-e", fmt.Sprintf("SUBOSITY_EMAIL=%s", config.Email),
		"--network", "host", // Needed for systemd and service management
		constants.ContainerImageDev,
		"install", // Command to run inside container
	}
	
	cmd := exec.CommandContext(ctx, "docker", args...)
	
	// Create pipes for streaming output
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}
	
	// Start the container
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start installer container: %w", err)
	}
	
	// Stream output with progress parsing
	progressChan := make(chan types.ProgressUpdate, 10)
	errorChan := make(chan error, 2)
	
	go r.streamProgressOutput(stdout, progressChan, errorChan)
	go r.streamErrorOutput(stderr, errorChan)
	
	// Handle progress updates
	go r.handleProgressUpdates(progressChan)
	
	// Wait for container to complete
	waitChan := make(chan error, 1)
	go func() {
		waitChan <- cmd.Wait()
	}()
	
	select {
	case err := <-waitChan:
		close(progressChan)
		if err != nil {
			return r.processContainerError(err)
		}
		return nil
		
	case err := <-errorChan:
		if err != nil {
			cmd.Process.Kill()
			return fmt.Errorf("container stream error: %w", err)
		}
		return nil
		
	case <-ctx.Done():
		cmd.Process.Kill()
		return ctx.Err()
	}
}

func (r *Runner) streamOutput(reader io.Reader, prefix string) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		r.logger.Debugf("%s %s", prefix, line)
	}
}

func (r *Runner) streamProgressOutput(reader io.Reader, progressChan chan<- types.ProgressUpdate, errorChan chan<- error) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		
		// Try to parse as JSON progress update
		var update types.ProgressUpdate
		if err := json.Unmarshal([]byte(line), &update); err == nil {
			progressChan <- update
		} else {
			// Regular log line
			r.logger.Info(line)
		}
	}
	
	if err := scanner.Err(); err != nil {
		errorChan <- err
	}
}

func (r *Runner) streamErrorOutput(reader io.Reader, errorChan chan<- error) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		
		// Try to parse as JSON error
		var installErr types.InstallError
		if err := json.Unmarshal([]byte(line), &installErr); err == nil {
			errorChan <- &installErr
		} else {
			// Regular error line
			r.logger.Error(line)
		}
	}
	
	if err := scanner.Err(); err != nil {
		errorChan <- err
	}
}

func (r *Runner) handleProgressUpdates(progressChan <-chan types.ProgressUpdate) {
	for update := range progressChan {
		// Display progress with colored output
		percentage := int(update.Progress * 100)
		
		// Create a simple progress bar
		barLength := 20
		filled := int(float64(barLength) * update.Progress)
		bar := strings.Repeat("▓", filled) + strings.Repeat("░", barLength-filled)
		
		r.logger.Infof("%s %s %d%% - %s", update.Phase, bar, percentage, update.Message)
	}
}

func (r *Runner) processContainerError(err error) error {
	// Extract exit code if possible
	if exitError, ok := err.(*exec.ExitError); ok {
		exitCode := exitError.ExitCode()
		
		switch exitCode {
		case 1:
			return errors.NewInstallationError(types.ErrCodeConfigInvalid,
				"container installation failed with configuration error",
				"container", "execution")
		case 2:
			return errors.NewInstallationError(types.ErrCodeSystemRequirements,
				"container installation failed with system requirements error",
				"container", "execution")
		case 3:
			return errors.NewInstallationError(types.ErrCodeDockerInstall,
				"container installation failed with Docker setup error",
				"container", "execution")
		case 4:
			return errors.NewInstallationError(types.ErrCodeSupabaseSetup,
				"container installation failed with Supabase setup error",
				"container", "execution")
		default:
			return errors.NewInstallationError(types.ErrCodeSystemRequirements,
				fmt.Sprintf("container installation failed with exit code %d", exitCode),
				"container", "execution")
		}
	}
	
	return errors.WrapError(err, types.ErrCodeSystemRequirements,
		"container installation failed", "container", "execution")
}
