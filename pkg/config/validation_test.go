package config

import (
	"testing"

	"github.com/subosity/subosity-installer/shared/types"
)

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      *types.InstallationConfig
		expectError bool
		errorCode   types.ErrorCode
	}{
		{
			name: "valid production config",
			config: &types.InstallationConfig{
				Environment: types.EnvironmentProduction,
				Domain:      "example.com",
				Email:       "admin@example.com",
			},
			expectError: false,
		},
		{
			name: "valid development config",
			config: &types.InstallationConfig{
				Environment: types.EnvironmentDevelopment,
				Domain:      "app.local",
				Email:       "",
			},
			expectError: false,
		},
		{
			name:        "nil config",
			config:      nil,
			expectError: true,
			errorCode:   types.ErrCodeConfigInvalid,
		},
		{
			name: "invalid environment",
			config: &types.InstallationConfig{
				Environment: "invalid",
				Domain:      "example.com",
				Email:       "admin@example.com",
			},
			expectError: true,
			errorCode:   types.ErrCodeConfigInvalid,
		},
		{
			name: "missing domain",
			config: &types.InstallationConfig{
				Environment: types.EnvironmentProduction,
				Domain:      "",
				Email:       "admin@example.com",
			},
			expectError: true,
			errorCode:   types.ErrCodeConfigInvalid,
		},
		{
			name: "invalid domain format",
			config: &types.InstallationConfig{
				Environment: types.EnvironmentProduction,
				Domain:      "not-a-domain",
				Email:       "admin@example.com",
			},
			expectError: true,
			errorCode:   types.ErrCodeConfigInvalid,
		},
		{
			name: "invalid email format",
			config: &types.InstallationConfig{
				Environment: types.EnvironmentProduction,
				Domain:      "example.com",
				Email:       "not-an-email",
			},
			expectError: true,
			errorCode:   types.ErrCodeConfigInvalid,
		},
		{
			name: "production without email",
			config: &types.InstallationConfig{
				Environment: types.EnvironmentProduction,
				Domain:      "example.com",
				Email:       "",
			},
			expectError: true,
			errorCode:   types.ErrCodeConfigInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfig(tt.config)

			if tt.expectError {
				if err == nil {
					t.Errorf("ValidateConfig() expected error but got none")
					return
				}

				if installErr, ok := err.(*types.InstallError); ok {
					if installErr.Code != tt.errorCode {
						t.Errorf("ValidateConfig() error code = %v, want %v", installErr.Code, tt.errorCode)
					}
				} else {
					t.Errorf("ValidateConfig() error type = %T, want *types.InstallError", err)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateConfig() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestSanitizeAndValidateDomain(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "valid domain",
			input:    "example.com",
			expected: "example.com",
			wantErr:  false,
		},
		{
			name:     "domain with whitespace",
			input:    "  example.com  ",
			expected: "example.com",
			wantErr:  false,
		},
		{
			name:     "uppercase domain",
			input:    "EXAMPLE.COM",
			expected: "example.com",
			wantErr:  false,
		},
		{
			name:     "domain with protocol",
			input:    "https://example.com",
			expected: "example.com",
			wantErr:  false,
		},
		{
			name:     "localhost for development",
			input:    "localhost",
			expected: "localhost",
			wantErr:  false,
		},
		{
			name:     "local domain",
			input:    "app.local",
			expected: "app.local",
			wantErr:  false,
		},
		{
			name:    "empty domain",
			input:   "",
			wantErr: true,
		},
		{
			name:    "invalid domain",
			input:   "not-a-domain",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := SanitizeAndValidateDomain(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("SanitizeAndValidateDomain() expected error but got none")
				}
				if result != "" {
					t.Errorf("SanitizeAndValidateDomain() expected empty result on error, got %s", result)
				}
			} else {
				if err != nil {
					t.Errorf("SanitizeAndValidateDomain() unexpected error = %v", err)
				}
				if result != tt.expected {
					t.Errorf("SanitizeAndValidateDomain() = %v, want %v", result, tt.expected)
				}
			}
		})
	}
}

func TestApplyDefaults(t *testing.T) {
	tests := []struct {
		name               string
		config             *types.InstallationConfig
		expectedSSLProvider types.SSLProvider
		expectedAutoRenew  bool
	}{
		{
			name: "production defaults",
			config: &types.InstallationConfig{
				Environment: types.EnvironmentProduction,
				Domain:      "example.com",
				Email:       "admin@example.com",
			},
			expectedSSLProvider: types.SSLProviderLetsEncrypt,
			expectedAutoRenew:   true,
		},
		{
			name: "development defaults",
			config: &types.InstallationConfig{
				Environment: types.EnvironmentDevelopment,
				Domain:      "",
				Email:       "",
			},
			expectedSSLProvider: types.SSLProviderSelfSigned,
			expectedAutoRenew:   false,
		},
		{
			name: "staging defaults",
			config: &types.InstallationConfig{
				Environment: types.EnvironmentStaging,
				Domain:      "staging.example.com",
				Email:       "admin@example.com",
			},
			expectedSSLProvider: types.SSLProviderLetsEncrypt,
			expectedAutoRenew:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ApplyDefaults(tt.config)

			if tt.config.SSL.Provider != tt.expectedSSLProvider {
				t.Errorf("ApplyDefaults() SSL provider = %v, want %v", 
					tt.config.SSL.Provider, tt.expectedSSLProvider)
			}

			if tt.config.SSL.AutoRenew != tt.expectedAutoRenew {
				t.Errorf("ApplyDefaults() SSL auto-renew = %v, want %v", 
					tt.config.SSL.AutoRenew, tt.expectedAutoRenew)
			}

			// Check development-specific defaults
			if tt.config.Environment == types.EnvironmentDevelopment && tt.config.Domain == "" {
				if tt.config.Domain != "subosity.local" {
					t.Errorf("ApplyDefaults() development domain = %v, want subosity.local", tt.config.Domain)
				}
			}
		})
	}
}
