package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORSConfig holds CORS configuration.
type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}

// DefaultCORSConfig returns a secure default CORS configuration.
// By default, only localhost origins are allowed for development safety.
// For production, configure allowed origins via PAPERBANANA_SECURITY_CORS_ALLOWED_ORIGINS.
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		// Secure default: only allow localhost origins
		// For production, explicitly configure allowed origins
		AllowedOrigins: []string{
			"http://localhost",
			"http://localhost:*",
			"http://127.0.0.1",
			"http://127.0.0.1:*",
		},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Origin", "Content-Type", "Accept", "Authorization", "X-API-Key", "X-Request-ID"},
		ExposedHeaders:   []string{"X-Request-ID", "X-RateLimit-Limit", "X-RateLimit-Remaining", "X-RateLimit-Reset"},
		AllowCredentials: false,
		MaxAge:           86400, // 24 hours
	}
}

// CORS returns a middleware that handles CORS.
func CORS(config CORSConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin == "" {
			c.Next()
			return
		}

		// Check if origin is allowed
		allowed := false
		for _, o := range config.AllowedOrigins {
			if o == "*" || o == origin || matchWildcard(o, origin) {
				allowed = true
				break
			}
		}

		if !allowed {
			c.Next()
			return
		}

		// Set CORS headers
		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Methods", strings.Join(config.AllowedMethods, ", "))
		c.Header("Access-Control-Allow-Headers", strings.Join(config.AllowedHeaders, ", "))
		c.Header("Access-Control-Expose-Headers", strings.Join(config.ExposedHeaders, ", "))

		if config.AllowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		if config.MaxAge > 0 {
			c.Header("Access-Control-Max-Age", intToStr(config.MaxAge))
		}

		// Handle preflight request
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// matchWildcard checks if the origin matches a wildcard pattern.
func matchWildcard(pattern, origin string) bool {
	if !strings.Contains(pattern, "*") {
		return pattern == origin
	}

	// Simple wildcard matching for subdomains
	parts := strings.SplitN(pattern, "*", 2)
	if len(parts) != 2 {
		return pattern == origin
	}

	prefix, suffix := parts[0], parts[1]

	if !strings.HasPrefix(origin, prefix) {
		return false
	}

	if !strings.HasSuffix(origin, suffix) {
		return false
	}

	return true
}

// CORSWithConfig returns a CORS middleware with the specified configuration.
func CORSWithConfig(allowedOrigins []string, allowCredentials bool) gin.HandlerFunc {
	config := DefaultCORSConfig()
	config.AllowedOrigins = allowedOrigins
	config.AllowCredentials = allowCredentials
	return CORS(config)
}

// StrictCORS returns a CORS middleware that only allows specific origins.
func StrictCORS(allowedOrigins []string) gin.HandlerFunc {
	config := DefaultCORSConfig()
	config.AllowedOrigins = allowedOrigins
	config.AllowCredentials = true
	return CORS(config)
}

// DevelopmentCORS returns a permissive CORS middleware for development only.
// WARNING: This allows any origin and should NEVER be used in production.
func DevelopmentCORS() gin.HandlerFunc {
	config := CORSConfig{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Origin", "Content-Type", "Accept", "Authorization", "X-API-Key", "X-Request-ID"},
		ExposedHeaders:   []string{"X-Request-ID", "X-RateLimit-Limit", "X-RateLimit-Remaining", "X-RateLimit-Reset"},
		AllowCredentials: false,
		MaxAge:           86400,
	}
	return CORS(config)
}

// ProductionCORS returns a CORS middleware configured for production use.
// Only explicitly listed origins are allowed.
func ProductionCORS(allowedOrigins []string, allowCredentials bool) gin.HandlerFunc {
	config := CORSConfig{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Origin", "Content-Type", "Accept", "Authorization", "X-API-Key", "X-Request-ID"},
		ExposedHeaders:   []string{"X-Request-ID", "X-RateLimit-Limit", "X-RateLimit-Remaining", "X-RateLimit-Reset"},
		AllowCredentials: allowCredentials,
		MaxAge:           86400,
	}
	return CORS(config)
}
