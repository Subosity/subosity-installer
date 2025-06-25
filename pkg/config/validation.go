package config

import (
	"fmt"
	"net/mail"
	"net/url"
	"regexp"
	"strings"

	"github.com/subosity/subosity-installer/pkg/errors"
	"github.com/subosity/subosity-installer/shared/types"
)

var (
	// Domain validation regex - basic FQDN check
	domainRegex = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`)
)

// ValidateConfig validates the installation configuration
func ValidateConfig(config *types.InstallationConfig) error {
	if config == nil {
		return errors.NewConfigError("configuration cannot be nil", "")
	}
	
	// Validate environment
	if err := validateEnvironment(config.Environment); err != nil {
		return err
	}
	
	// Validate domain
	if err := validateDomain(config.Domain); err != nil {
		return err
	}
	
	// Validate email (optional for dev environment)
	if config.Email != "" {
		if err := validateEmail(config.Email); err != nil {
			return err
		}
	} else if config.Environment == types.EnvironmentProduction {
		return errors.NewConfigError("email is required for production environment", 
			"email is needed for SSL certificate generation")
	}
	
	// Validate SSL configuration
	if err := validateSSLConfig(&config.SSL); err != nil {
		return err
	}
	
	return nil
}

// ApplyDefaults applies environment-specific defaults to the configuration
func ApplyDefaults(config *types.InstallationConfig) {
	switch config.Environment {
	case types.EnvironmentProduction:
		if config.SSL.Provider == "" {
			config.SSL.Provider = types.SSLProviderLetsEncrypt
		}
		config.SSL.AutoRenew = true
		
	case types.EnvironmentStaging:
		if config.SSL.Provider == "" {
			config.SSL.Provider = types.SSLProviderLetsEncrypt
		}
		config.SSL.AutoRenew = true
		
	case types.EnvironmentDevelopment:
		if config.SSL.Provider == "" {
			config.SSL.Provider = types.SSLProviderSelfSigned
		}
		config.SSL.AutoRenew = false
		
		// Allow localhost and .local domains for development
		if config.Domain == "" {
			config.Domain = "subosity.local"
		}
	}
}

func validateEnvironment(env types.Environment) error {
	switch env {
	case types.EnvironmentDevelopment, types.EnvironmentStaging, types.EnvironmentProduction:
		return nil
	default:
		return errors.NewConfigError(
			"invalid environment",
			fmt.Sprintf("must be one of: %s, %s, %s", 
				types.EnvironmentDevelopment, 
				types.EnvironmentStaging, 
				types.EnvironmentProduction),
		)
	}
}

func validateDomain(domain string) error {
	if domain == "" {
		return errors.NewConfigError("domain is required", "")
	}
	
	// Trim whitespace and convert to lowercase
	domain = strings.TrimSpace(strings.ToLower(domain))
	
	// Allow localhost and .local domains for development
	if domain == "localhost" || strings.HasSuffix(domain, ".local") {
		return nil
	}
	
	// Validate FQDN format
	if !domainRegex.MatchString(domain) {
		return errors.NewConfigError(
			"invalid domain format",
			fmt.Sprintf("domain '%s' is not a valid FQDN", domain),
		)
	}
	
	// Check length constraints
	if len(domain) > 253 {
		return errors.NewConfigError(
			"domain too long",
			"domain name must not exceed 253 characters",
		)
	}
	
	return nil
}

func validateEmail(email string) error {
	if email == "" {
		return nil // Email is optional in some cases
	}
	
	// Use Go's built-in email validation
	_, err := mail.ParseAddress(email)
	if err != nil {
		return errors.NewConfigError(
			"invalid email format",
			fmt.Sprintf("email '%s' is not valid", email),
		)
	}
	
	return nil
}

func validateSSLConfig(ssl *types.SSLConfig) error {
	if ssl == nil {
		return nil // SSL config is optional, defaults will be applied
	}
	
	switch ssl.Provider {
	case types.SSLProviderLetsEncrypt:
		// Let's Encrypt requires a valid email
		if ssl.Email == "" {
			return errors.NewConfigError(
				"email required for Let's Encrypt",
				"Let's Encrypt requires a valid email address for certificate registration",
			)
		}
		if err := validateEmail(ssl.Email); err != nil {
			return err
		}
		
	case types.SSLProviderCustom:
		// Custom SSL requires both cert and key
		if ssl.CustomCert == "" || ssl.CustomKey == "" {
			return errors.NewConfigError(
				"custom SSL requires both certificate and key",
				"both custom_cert and custom_key must be provided for custom SSL",
			)
		}
		
	case types.SSLProviderSelfSigned:
		// Self-signed doesn't require additional validation
		
	case "":
		// Empty provider is fine, defaults will be applied
		
	default:
		return errors.NewConfigError(
			"invalid SSL provider",
			fmt.Sprintf("SSL provider must be one of: %s, %s, %s",
				types.SSLProviderLetsEncrypt,
				types.SSLProviderSelfSigned,
				types.SSLProviderCustom),
		)
	}
	
	return nil
}

// SanitizeAndValidateDomain cleans and validates a domain name
func SanitizeAndValidateDomain(domain string) (string, error) {
	// Trim whitespace and convert to lowercase
	domain = strings.TrimSpace(strings.ToLower(domain))
	
	// Remove protocol if present
	if strings.HasPrefix(domain, "http://") || strings.HasPrefix(domain, "https://") {
		parsed, err := url.Parse(domain)
		if err != nil {
			return "", errors.NewConfigError("invalid domain URL", err.Error())
		}
		domain = parsed.Host
	}
	
	// Validate the cleaned domain
	if err := validateDomain(domain); err != nil {
		return "", err
	}
	
	return domain, nil
}
