package handlers

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
	domainconfig "github.com/paperbanana/paperbanana/internal/domain/config"
)

// ConfigSSEHandler handles SSE for config changes.
type ConfigSSEHandler struct {
	watcher domainconfig.ConfigWatcher
}

// NewConfigSSEHandler creates a new config SSE handler.
func NewConfigSSEHandler(watcher domainconfig.ConfigWatcher) *ConfigSSEHandler {
	return &ConfigSSEHandler{watcher: watcher}
}

// StreamConfigChanges handles GET /api/v1/config/stream
func (h *ConfigSSEHandler) StreamConfigChanges(c *gin.Context) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	events := h.watcher.Subscribe()
	defer h.watcher.Unsubscribe(events)

	// Send initial connection message
	c.SSEvent("connected", "Config stream connected")
	c.Writer.Flush()

	for {
		select {
		case event := <-events:
			data, _ := json.Marshal(event)
			c.SSEvent("config_changed", string(data))
			c.Writer.Flush()
		case <-c.Request.Context().Done():
			return
		}
	}
}
