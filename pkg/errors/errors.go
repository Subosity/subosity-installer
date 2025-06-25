package errors

import (
	"fmt"
	"time"

	"github.com/subosity/subosity-installer/shared/types"
)

// NewInstallationError creates a new structured installation error
func NewInstallationError(code types.ErrorCode, message string, component string, operation string) *types.InstallError {
	return &types.InstallError{
		Code:    code,
		Message: message,
		Context: types.ErrorContext{
			Component: component,
			Operation: operation,
		},
		Timestamp: time.Now(),
	}
}

// WrapError wraps an existing error with installation context
func WrapError(err error, code types.ErrorCode, message string, component string, operation string) *types.InstallError {
	installErr := NewInstallationError(code, message, component, operation)
	installErr.Cause = err
	if err != nil {
		installErr.Details = err.Error()
	}
	return installErr
}

// NewConfigError creates a configuration validation error
func NewConfigError(message string, details string) *types.InstallError {
	return &types.InstallError{
		Code:    types.ErrCodeConfigInvalid,
		Message: message,
		Details: details,
		Suggestions: []string{
			"Check configuration file syntax",
			"Verify all required fields are present",
			"Validate field formats (email, domain, etc.)",
		},
		Context: types.ErrorContext{
			Component: "config",
			Operation: "validation",
		},
		Timestamp: time.Now(),
	}
}

// NewSystemError creates a system requirements error
func NewSystemError(message string, details string, suggestions []string) *types.InstallError {
	if suggestions == nil {
		suggestions = []string{
			"Check system requirements documentation",
			"Ensure sufficient resources are available",
			"Verify operating system compatibility",
		}
	}
	
	return &types.InstallError{
		Code:        types.ErrCodeSystemRequirements,
		Message:     message,
		Details:     details,
		Suggestions: suggestions,
		Context: types.ErrorContext{
			Component: "system",
			Operation: "validation",
		},
		Timestamp: time.Now(),
	}
}

// NewDockerError creates a Docker-related error
func NewDockerError(message string, details string) *types.InstallError {
	return &types.InstallError{
		Code:    types.ErrCodeDockerInstall,
		Message: message,
		Details: details,
		Suggestions: []string{
			"Verify internet connectivity",
			"Check if running with sufficient privileges (sudo)",
			"Ensure package repositories are accessible",
			"Try manual Docker installation: https://docs.docker.com/install/",
		},
		Context: types.ErrorContext{
			Component: "docker",
			Operation: "installation",
		},
		Timestamp: time.Now(),
	}
}

// FormatError formats an installation error for user display
func FormatError(err *types.InstallError) string {
	result := fmt.Sprintf("Error [%s]: %s", err.Code, err.Message)
	
	if err.Details != "" {
		result += fmt.Sprintf("\nDetails: %s", err.Details)
	}
	
	if len(err.Suggestions) > 0 {
		result += "\n\nSuggestions:"
		for _, suggestion := range err.Suggestions {
			result += fmt.Sprintf("\n  • %s", suggestion)
		}
	}
	
	return result
}
