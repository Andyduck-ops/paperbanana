package middleware

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthConfig holds authentication configuration.
type AuthConfig struct {
	Enabled    bool
	APIKeys    []string
	HeaderName string
}

// Auth returns a middleware that validates API keys.
func Auth(config AuthConfig) gin.HandlerFunc {
	if !config.Enabled {
		// Authentication disabled, pass through
		return func(c *gin.Context) {
			c.Next()
		}
	}

	if config.HeaderName == "" {
		config.HeaderName = "X-API-Key"
	}

	return func(c *gin.Context) {
		apiKey := c.GetHeader(config.HeaderName)

		// Check if API key is provided
		if apiKey == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "API key required",
				"code":  "auth_missing_key",
			})
			return
		}

		// Validate API key using constant-time comparison
		valid := false
		for _, key := range config.APIKeys {
			if subtle.ConstantTimeCompare([]byte(apiKey), []byte(key)) == 1 {
				valid = true
				break
			}
		}

		if !valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid API key",
				"code":  "auth_invalid_key",
			})
			return
		}

		// Store hashed key identifier for logging
		keyID := hashKeyID(apiKey)
		c.Set("api_key_id", keyID)

		c.Next()
	}
}

// hashKeyID returns a truncated hash of the API key for identification.
func hashKeyID(key string) string {
	if len(key) < 8 {
		return "****"
	}
	// Show first 4 and last 4 characters
	return key[:4] + "****" + key[len(key)-4:]
}

// OptionalAuth returns a middleware that validates API keys if present but doesn't require them.
func OptionalAuth(config AuthConfig) gin.HandlerFunc {
	if config.HeaderName == "" {
		config.HeaderName = "X-API-Key"
	}

	return func(c *gin.Context) {
		apiKey := c.GetHeader(config.HeaderName)

		if apiKey == "" {
			// No API key provided, continue without auth
			c.Set("authenticated", false)
			c.Next()
			return
		}

		// Validate API key
		valid := false
		for _, key := range config.APIKeys {
			if subtle.ConstantTimeCompare([]byte(apiKey), []byte(key)) == 1 {
				valid = true
				break
			}
		}

		if valid {
			c.Set("authenticated", true)
			c.Set("api_key_id", hashKeyID(apiKey))
		} else {
			c.Set("authenticated", false)
		}

		c.Next()
	}
}

// GetAPIKeyID returns the API key identifier from the context.
func GetAPIKeyID(c *gin.Context) string {
	if keyID, exists := c.Get("api_key_id"); exists {
		if id, ok := keyID.(string); ok {
			return id
		}
	}
	return "anonymous"
}

// IsAuthenticated returns true if the request has a valid API key.
func IsAuthenticated(c *gin.Context) bool {
	auth, exists := c.Get("authenticated")
	if !exists {
		return false
	}
	return auth.(bool)
}

// BearerAuth returns a middleware that validates Bearer tokens.
func BearerAuth(validTokens []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header required",
				"code":  "auth_missing_header",
			})
			return
		}

		// Extract Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization header format",
				"code":  "auth_invalid_format",
			})
			return
		}

		token := parts[1]

		// Validate token
		valid := false
		for _, t := range validTokens {
			if subtle.ConstantTimeCompare([]byte(token), []byte(t)) == 1 {
				valid = true
				break
			}
		}

		if !valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token",
				"code":  "auth_invalid_token",
			})
			return
		}

		c.Set("authenticated", true)
		c.Next()
	}
}
