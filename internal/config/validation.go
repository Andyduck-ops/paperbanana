package config

import (
	"errors"
	"net/url"
	"regexp"
	"strings"
)

var (
	// providerNamePattern matches valid provider names
	providerNamePattern = regexp.MustCompile(`^[a-z][a-z0-9_\-]*$`)

	// modelIDPattern matches model identifiers
	modelIDPattern = regexp.MustCompile(`^[A-Za-z0-9_\-\.\/]+$`)

	// disallowedPatterns for injection detection
	disallowedPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)<script`),     // XSS
		regexp.MustCompile(`(?i)javascript:`), // XSS
		regexp.MustCompile(`\$\(`),            // Command substitution
		regexp.MustCompile(`\|\s*\w+`),        // Pipe to command
		regexp.MustCompile(`;\s*\w+`),         // Command chain
		regexp.MustCompile(`\.\./`),           // Path traversal
	}
)

// ValidateProviderName validates provider name format.
func ValidateProviderName(name string) error {
	if name == "" {
		return errors.New("provider name cannot be empty")
	}
	if len(name) > 64 {
		return errors.New("provider name too long (max 64 chars)")
	}
	if !providerNamePattern.MatchString(name) {
		return errors.New("provider name must start with lowercase letter and contain only alphanumeric, underscore, or hyphen")
	}
	return nil
}

// ValidateAPIKey validates API key format.
func ValidateAPIKey(key string) error {
	if key == "" {
		return errors.New("API key cannot be empty")
	}
	if len(key) > 512 {
		return errors.New("API key too long (max 512 chars)")
	}

	// Check for injection patterns
	for _, pattern := range disallowedPatterns {
		if pattern.MatchString(key) {
			return errors.New("API key contains disallowed pattern")
		}
	}

	return nil
}

// ValidateBaseURL validates base URL format.
func ValidateBaseURL(baseURL string) error {
	if baseURL == "" {
		return nil // Optional field
	}
	if len(baseURL) > 512 {
		return errors.New("base URL too long (max 512 chars)")
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		return errors.New("invalid URL format")
	}

	if u.Scheme != "https" && u.Scheme != "http" {
		return errors.New("URL must use http or https scheme")
	}

	if u.Host == "" {
		return errors.New("URL must have a host")
	}

	// Log warning for localhost/private addresses in production
	host := strings.ToLower(u.Host)
	if strings.Contains(host, "localhost") ||
		strings.Contains(host, "127.0.0.1") ||
		strings.Contains(host, "0.0.0.0") ||
		strings.Contains(host, "::1") {
		// Allow in development - could log warning in production
	}

	return nil
}

// ValidateModelID validates model identifier format.
func ValidateModelID(model string) error {
	if model == "" {
		return errors.New("model cannot be empty")
	}
	if len(model) > 256 {
		return errors.New("model ID too long (max 256 chars)")
	}
	if !modelIDPattern.MatchString(model) {
		return errors.New("model ID contains invalid characters")
	}
	return nil
}

// ValidateProviderConfig validates a complete provider configuration.
func ValidateProviderConfig(name string, cfg ProviderConfig) error {
	if err := ValidateProviderName(name); err != nil {
		return err
	}
	if err := ValidateAPIKey(cfg.APIKey); err != nil {
		return err
	}
	if err := ValidateBaseURL(cfg.BaseURL); err != nil {
		return err
	}
	if err := ValidateModelID(cfg.Model); err != nil {
		return err
	}
	return nil
}

// MaskAPIKey masks an API key for display.
// Shows first 6 and last 4 characters with **** in between.
func MaskAPIKey(key string) string {
	if len(key) <= 10 {
		return "****"
	}
	return key[:6] + "****" + key[len(key)-4:]
}
