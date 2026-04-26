package handlers

import (
	"context"
	"errors"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// SessionRegistry tracks active sessions and provides cancellation capability.
type SessionRegistry struct {
	mu       sync.RWMutex
	sessions map[string]*sessionHandle
}

type sessionHandle struct {
	cancel context.CancelFunc
	once   sync.Once
}

// cancelOnce safely cancels the session exactly once.
func (h *sessionHandle) cancelOnce() {
	h.once.Do(func() {
		if h.cancel != nil {
			h.cancel()
		}
	})
}

// NewSessionRegistry creates a new SessionRegistry.
func NewSessionRegistry() *SessionRegistry {
	return &SessionRegistry{
		sessions: make(map[string]*sessionHandle),
	}
}

// Register adds a session to the registry with its cancel function.
// Returns a wrapped context that will be automatically cleaned up.
func (r *SessionRegistry) Register(sessionID string, ctx context.Context) (context.Context, context.CancelFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()

	ctx, cancel := context.WithCancel(ctx)
	handle := &sessionHandle{cancel: cancel}
	r.sessions[sessionID] = handle

	// Return a cancel function that also removes from registry
	return ctx, func() {
		handle.cancelOnce()
		r.mu.Lock()
		delete(r.sessions, sessionID)
		r.mu.Unlock()
	}
}

// Cancel cancels a specific session by ID.
// Returns true if the session was found and canceled.
func (r *SessionRegistry) Cancel(sessionID string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	handle, ok := r.sessions[sessionID]
	if !ok {
		return false
	}

	handle.cancelOnce()
	delete(r.sessions, sessionID)
	return true
}

// Exists checks if a session is currently active.
func (r *SessionRegistry) Exists(sessionID string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.sessions[sessionID]
	return ok
}

// ErrSessionNotFound is returned when attempting to cancel a non-existent session.
var ErrSessionNotFound = errors.New("session not found or already completed")

// CancelHandler handles session cancellation requests.
type CancelHandler struct {
	registry *SessionRegistry
	logger   *zap.Logger
}

// NewCancelHandler creates a new CancelHandler.
func NewCancelHandler(registry *SessionRegistry, logger *zap.Logger) *CancelHandler {
	return &CancelHandler{
		registry: registry,
		logger:   logger,
	}
}

// CancelRequest represents the cancel request body.
type CancelRequest struct {
	Reason string `json:"reason,omitempty"`
}

// CancelResponse represents the cancel response.
type CancelResponse struct {
	SessionID string `json:"session_id"`
	Status    string `json:"status"`
	Message   string `json:"message"`
}

// Cancel handles POST /api/v1/sessions/:session_id/cancel
func (h *CancelHandler) Cancel(c *gin.Context) {
	sessionID := c.Param("session_id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session_id is required"})
		return
	}

	var req CancelRequest
	// Request body is optional, ignore binding errors
	_ = c.ShouldBindJSON(&req)

	canceled := h.registry.Cancel(sessionID)
	if !canceled {
		h.logger.Debug("session not found for cancellation",
			zap.String("session_id", sessionID),
		)
		c.JSON(http.StatusNotFound, gin.H{
			"error":      ErrSessionNotFound.Error(),
			"session_id": sessionID,
		})
		return
	}

	h.logger.Info("session canceled",
		zap.String("session_id", sessionID),
		zap.String("reason", req.Reason),
	)

	c.JSON(http.StatusOK, CancelResponse{
		SessionID: sessionID,
		Status:    "canceled",
		Message:   "Session canceled successfully",
	})
}
