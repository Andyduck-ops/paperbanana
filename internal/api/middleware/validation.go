package middleware

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

// ValidationConfig holds input validation configuration.
type ValidationConfig struct {
	MaxBodySize         int64    // Maximum request body size in bytes
	MaxPromptLength     int      // Maximum prompt length
	AllowedContentTypes []string // Allowed content types for uploads
}

// DefaultValidationConfig returns a default validation configuration.
func DefaultValidationConfig() ValidationConfig {
	return ValidationConfig{
		MaxBodySize:     10 * 1024 * 1024, // 10MB
		MaxPromptLength: 100 * 1024,       // 100KB
		AllowedContentTypes: []string{
			"image/png",
			"image/jpeg",
			"image/jpg",
			"image/gif",
			"image/webp",
			"application/pdf",
		},
	}
}

// RequestSizeLimit returns a middleware that limits request body size.
func RequestSizeLimit(maxSize int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check Content-Length header
		if c.Request.ContentLength > maxSize {
			c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, gin.H{
				"error":          "Request body too large",
				"code":           "request_too_large",
				"max_size_bytes": maxSize,
			})
			return
		}

		// Wrap the body reader to limit size
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSize)

		c.Next()
	}
}

// ValidateContentType returns a middleware that validates content type for file uploads.
func ValidateContentType(allowedTypes []string) gin.HandlerFunc {
	allowedMap := make(map[string]bool)
	for _, t := range allowedTypes {
		allowedMap[strings.ToLower(t)] = true
	}

	return func(c *gin.Context) {
		contentType := c.GetHeader("Content-Type")

		// Only validate for multipart/form-data uploads
		if !strings.HasPrefix(contentType, "multipart/form-data") {
			c.Next()
			return
		}

		// Content type will be validated per file in the handler
		c.Set("allowed_content_types", allowedTypes)
		c.Next()
	}
}

// SanitizeInput sanitizes user input by removing control characters.
func SanitizeInput(input string) string {
	// Remove control characters except newlines and tabs
	controlChars := regexp.MustCompile(`[\x00-\x08\x0B\x0C\x0E-\x1F\x7F]`)
	return controlChars.ReplaceAllString(input, "")
}

// ValidatePromptLength validates the length of a prompt.
func ValidatePromptLength(prompt string, maxLength int) bool {
	return len(prompt) <= maxLength
}

// SanitizePrompt sanitizes and validates a prompt.
func SanitizePrompt(prompt string, maxLength int) (string, bool) {
	sanitized := SanitizeInput(prompt)
	if len(sanitized) > maxLength {
		return sanitized[:maxLength], false
	}
	return sanitized, true
}

// InputSanitizer returns a middleware that sanitizes request inputs.
func InputSanitizer() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only process JSON requests
		if !strings.Contains(c.GetHeader("Content-Type"), "application/json") {
			c.Next()
			return
		}

		// Read body with size limit
		body, err := io.ReadAll(io.LimitReader(c.Request.Body, 1024*1024))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "Failed to read request body",
				"code":  "body_read_error",
			})
			return
		}

		// Sanitize control characters from body
		sanitized := SanitizeInput(string(body))

		// Restore body
		c.Request.Body = io.NopCloser(bytes.NewBufferString(sanitized))

		c.Next()
	}
}

// RequireJSON returns a middleware that requires JSON content type.
func RequireJSON() gin.HandlerFunc {
	return func(c *gin.Context) {
		contentType := c.GetHeader("Content-Type")
		if !strings.Contains(contentType, "application/json") {
			c.AbortWithStatusJSON(http.StatusUnsupportedMediaType, gin.H{
				"error": "Content-Type must be application/json",
				"code":  "invalid_content_type",
			})
			return
		}
		c.Next()
	}
}

// ValidateFileUpload validates an uploaded file.
func ValidateFileUpload(c *gin.Context, fileHeader *multipart.FileHeader, config ValidationConfig) error {
	// Check file size
	if fileHeader.Size > config.MaxBodySize {
		return gin.Error{
			Err:  gin.Error{}.Err,
			Type: gin.ErrorTypePrivate,
			Meta: gin.H{
				"error": "File size exceeds limit",
				"code":  "file_too_large",
			},
		}
	}

	// Open file to check content type
	file, err := fileHeader.Open()
	if err != nil {
		return err
	}
	defer file.Close()

	// Read first 512 bytes to detect content type
	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil && err != io.EOF {
		return err
	}

	// Detect content type
	detectedType := http.DetectContentType(buffer)

	// Check if content type is allowed
	allowed := false
	for _, t := range config.AllowedContentTypes {
		if strings.HasPrefix(detectedType, t) {
			allowed = true
			break
		}
	}

	if !allowed {
		return gin.Error{
			Err:  gin.Error{}.Err,
			Type: gin.ErrorTypePrivate,
			Meta: gin.H{
				"error":         "File type not allowed",
				"code":          "invalid_file_type",
				"detected_type": detectedType,
				"allowed_types": config.AllowedContentTypes,
			},
		}
	}

	return nil
}
