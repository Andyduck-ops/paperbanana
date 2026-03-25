package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// AuditLogger returns a middleware that logs audit information for each request.
func AuditLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Generate or extract request ID
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Set request ID in context and response
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)

		// Get API key ID if authenticated
		apiKeyID := GetAPIKeyID(c)

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start)

		// Log audit information
		logger.Info("request",
			zap.String("request_id", requestID),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("query", c.Request.URL.RawQuery),
			zap.Int("status", c.Writer.Status()),
			zap.Int("size", c.Writer.Size()),
			zap.String("ip", c.ClientIP()),
			zap.String("user_agent", c.GetHeader("User-Agent")),
			zap.String("api_key_id", apiKeyID),
			zap.Duration("duration", duration),
		)
	}
}

// RequestID returns a middleware that ensures each request has a unique ID.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)

		c.Next()
	}
}

// GetRequestID returns the request ID from the context.
func GetRequestID(c *gin.Context) string {
	if requestID, exists := c.Get("request_id"); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return ""
}

// AuditEvent represents an audit event for custom logging.
type AuditEvent struct {
	RequestID   string                 `json:"request_id"`
	EventType   string                 `json:"event_type"`
	UserID      string                 `json:"user_id,omitempty"`
	APIKeyID    string                 `json:"api_key_id,omitempty"`
	Resource    string                 `json:"resource"`
	Action      string                 `json:"action"`
	Details     map[string]interface{} `json:"details,omitempty"`
	Status      string                 `json:"status"`
	Error       string                 `json:"error,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
	Duration    time.Duration          `json:"duration"`
}

// LogAudit logs a custom audit event.
func LogAudit(c *gin.Context, logger *zap.Logger, event AuditEvent) {
	if event.RequestID == "" {
		event.RequestID = GetRequestID(c)
	}
	if event.APIKeyID == "" {
		event.APIKeyID = GetAPIKeyID(c)
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	logger.Info("audit_event",
		zap.String("request_id", event.RequestID),
		zap.String("event_type", event.EventType),
		zap.String("user_id", event.UserID),
		zap.String("api_key_id", event.APIKeyID),
		zap.String("resource", event.Resource),
		zap.String("action", event.Action),
		zap.String("status", event.Status),
		zap.String("error", event.Error),
		zap.Time("timestamp", event.Timestamp),
		zap.Duration("duration", event.Duration),
		zap.Any("details", event.Details),
	)
}
