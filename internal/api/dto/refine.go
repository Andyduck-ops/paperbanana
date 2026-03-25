// Package dto provides Data Transfer Objects for API requests and responses.
package dto

// RefineRequest is the request body for image refinement.
type RefineRequest struct {
	SessionID       string `json:"session_id"`
	ImageData       string `json:"image_data"` // Base64 encoded image or data URL
	Instructions    string `json:"instructions"`
	Resolution      string `json:"resolution"` // "2K" or "4K"
	Model           string `json:"model"`      // Optional model override
	ProviderID      string `json:"provider_id"`
	MaxIterations   int    `json:"max_iterations,omitempty"`
	EnableIteration bool   `json:"enable_iteration,omitempty"`
}

// RefineImagePayload contains the refined image payload.
type RefineImagePayload struct {
	Data     string            `json:"data,omitempty"`
	MIMEType string            `json:"mime_type,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// RefineResponseMetadata describes refinement loop metadata.
type RefineResponseMetadata struct {
	Iterations   string `json:"iterations,omitempty"`
	StopReason   string `json:"stop_reason,omitempty"`
	QualityScore string `json:"quality_score,omitempty"`
}

// RefineResponse is the response for image refinement.
type RefineResponse struct {
	SessionID string                  `json:"session_id"`
	Status    string                  `json:"status"`
	Image     *RefineImagePayload     `json:"image,omitempty"`
	ImageData string                  `json:"image_data,omitempty"` // Compatibility field for older callers
	Content   string                  `json:"content,omitempty"`
	Error     string                  `json:"error,omitempty"`
	Metadata  *RefineResponseMetadata `json:"metadata,omitempty"`
}
