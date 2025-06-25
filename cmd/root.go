package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/subosity/subosity-installer/pkg/logger"
	"github.com/subosity/subosity-installer/shared/constants"
)

var (
	verbose bool
	logLevel string
	log     logger.Logger
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "subosity-installer",
	Short: "A turnkey deployment tool for self-hosting Subosity",
	Long: `Subosity Installer is a production-ready, container-first deployment tool 
that eliminates the complexity of self-hosting by handling Docker installation, 
Supabase setup, SSL configuration, and service management with enterprise-grade 
reliability and security.

The installer follows a container-first architecture with a thin binary that 
validates the environment and delegates complex operations to a specialized 
container image.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Initialize logger based on flags
		log = logger.NewLogger(logLevel, verbose)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, 
		"Enable verbose output")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", 
		"Set log level (debug, info, warn, error)")
		
	// Version command
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("%s version %s\n", constants.AppName, constants.AppVersion)
		},
	})
}
