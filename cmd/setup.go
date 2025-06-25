package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/subosity/subosity-installer/internal/container"
	"github.com/subosity/subosity-installer/internal/docker"
	"github.com/subosity/subosity-installer/internal/system"
	"github.com/subosity/subosity-installer/pkg/config"
	"github.com/subosity/subosity-installer/pkg/errors"
	"github.com/subosity/subosity-installer/shared/constants"
	"github.com/subosity/subosity-installer/shared/types"
)

var (
	environment string
	domain      string
	email       string
	sslProvider string
	configFile  string
	timeout     int
)

// setupCmd represents the setup command
var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Install and configure Subosity",
	Long: `Install and configure Subosity with all required dependencies including
Docker, Supabase, and SSL certificates. This command handles the complete 
installation process from system validation to service deployment.

Examples:
  # Production installation
  subosity-installer setup --env prod --domain myapp.com --email admin@myapp.com

  # Development installation  
  subosity-installer setup --env dev --domain myapp.local

  # Using configuration file
  subosity-installer setup --config config.yaml`,
	RunE: runSetup,
}

func init() {
	rootCmd.AddCommand(setupCmd)

	// Required flags
	setupCmd.Flags().StringVar(&environment, "env", "", 
		"Environment (dev, staging, prod) [required]")
	setupCmd.Flags().StringVar(&domain, "domain", "", 
		"Domain name for the installation [required]")
	
	// Optional flags
	setupCmd.Flags().StringVar(&email, "email", "", 
		"Email address for SSL certificates and notifications")
	setupCmd.Flags().StringVar(&sslProvider, "ssl-provider", "", 
		"SSL provider (letsencrypt, self-signed, custom)")
	setupCmd.Flags().StringVar(&configFile, "config", "", 
		"Configuration file path")
	setupCmd.Flags().IntVar(&timeout, "timeout", 15, 
		"Installation timeout in minutes")

	// Mark required flags
	setupCmd.MarkFlagRequired("env")
	setupCmd.MarkFlagRequired("domain")
}

func runSetup(cmd *cobra.Command, args []string) error {
	// Create context with cancellation
	ctx, cancel := context.WithTimeout(context.Background(), 
		time.Duration(timeout)*time.Minute)
	defer cancel()
	
	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Warn("Received shutdown signal, cancelling installation...")
		cancel()
	}()
	
	log.Info("Starting Subosity installation...")
	
	// Step 1: Build configuration
	installConfig, err := buildConfiguration()
	if err != nil {
		log.Error(errors.FormatError(err.(*types.InstallError)))
		return err
	}
	
	// Step 2: Validate system requirements
	if err := validateSystem(ctx); err != nil {
		log.Error(errors.FormatError(err.(*types.InstallError)))
		return err
	}
	
	// Step 3: Ensure Docker is available
	if err := ensureDocker(ctx); err != nil {
		log.Error(errors.FormatError(err.(*types.InstallError)))
		return err
	}
	
	// Step 4: Run container-based installation
	if err := runContainerInstallation(ctx, installConfig); err != nil {
		log.Error(errors.FormatError(err.(*types.InstallError)))
		return err
	}
	
	// Step 5: Final verification and success message
	return displaySuccess(installConfig)
}

func buildConfiguration() (*types.InstallationConfig, error) {
	log.Info("Building installation configuration...")
	
	// Create base configuration from flags
	installConfig := &types.InstallationConfig{
		Environment: types.Environment(environment),
		Domain:      domain,
		Email:       email,
		SSL: types.SSLConfig{
			Provider: types.SSLProvider(sslProvider),
		},
		CreatedAt: time.Now(),
		Version:   constants.AppVersion,
	}
	
	// TODO: Load from config file if specified
	if configFile != "" {
		log.Warnf("Configuration file support not yet implemented: %s", configFile)
	}
	
	// Apply environment-specific defaults
	config.ApplyDefaults(installConfig)
	
	// Validate configuration
	if err := config.ValidateConfig(installConfig); err != nil {
		return nil, err
	}
	
	// Sanitize domain
	sanitizedDomain, err := config.SanitizeAndValidateDomain(installConfig.Domain)
	if err != nil {
		return nil, err
	}
	installConfig.Domain = sanitizedDomain
	
	log.Infof("Configuration validated for %s environment on domain %s", 
		installConfig.Environment, installConfig.Domain)
	
	return installConfig, nil
}

func validateSystem(ctx context.Context) error {
	log.Info("Validating system requirements...")
	
	detector := system.NewDetector(log)
	
	// Detect system information
	hostMetadata, err := detector.DetectSystem(ctx)
	if err != nil {
		return err
	}
	
	log.Infof("Detected system: %s %s (%s)", 
		hostMetadata.OS, hostMetadata.Version, hostMetadata.Architecture)
	
	// Validate system requirements
	if err := detector.ValidateSystemRequirements(ctx, hostMetadata); err != nil {
		return err
	}
	
	return nil
}

func ensureDocker(ctx context.Context) error {
	log.Info("Checking Docker availability...")
	
	dockerService := docker.NewService(log)
	
	// Check if Docker is already installed
	installed, err := dockerService.IsInstalled(ctx)
	if err != nil {
		return errors.WrapError(err, types.ErrCodeDockerInstall,
			"failed to check Docker installation", "docker", "check")
	}
	
	if installed {
		version, _ := dockerService.GetVersion(ctx)
		log.Infof("Docker is already installed: %s", version)
		return nil
	}
	
	// Install Docker
	log.Info("Docker not found, installing...")
	if err := dockerService.Install(ctx); err != nil {
		return err
	}
	
	return nil
}

func runContainerInstallation(ctx context.Context, installConfig *types.InstallationConfig) error {
	log.Info("Starting container-based installation...")
	
	runner := container.NewRunner(log)
	
	if err := runner.RunInstaller(ctx, installConfig); err != nil {
		return err
	}
	
	return nil
}

func displaySuccess(installConfig *types.InstallationConfig) error {
	log.Info("")
	log.Info("🎉 Subosity installation completed successfully!")
	log.Info("")
	log.Infof("Environment: %s", installConfig.Environment)
	log.Infof("Domain: %s", installConfig.Domain)
	
	// Determine the URL based on SSL configuration
	protocol := "https"
	if installConfig.SSL.Provider == types.SSLProviderSelfSigned && 
		installConfig.Environment == types.EnvironmentDevelopment {
		protocol = "http"
	}
	
	url := fmt.Sprintf("%s://%s", protocol, installConfig.Domain)
	log.Infof("Access URL: %s", url)
	log.Info("")
	log.Info("Next steps:")
	log.Info("1. Configure your domain's DNS to point to this server")
	log.Info("2. Access the application at the URL above")
	log.Info("3. Complete the initial setup in the web interface")
	log.Info("")
	log.Info("For help and documentation, visit: https://github.com/subosity/subosity-installer")
	
	return nil
}
