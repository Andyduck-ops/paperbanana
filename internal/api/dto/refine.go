// Package dto provides Data Transfer Objects for API requests and responses.
package dto

// RefineRequest is the request body for image refinement.
type RefineRequest struct {
	SessionID    string `json:"session_id"`
	ImageData    string `json:"image_data"`    // Base64 encoded image
	Instructions string `json:"instructions"`  // Refinement instructions
	Resolution   string `json:"resolution"`    // "2K" or "4K"
	Model        string `json:"model"`         // Optional model override
	ProviderID   string `json:"provider_id"`   // Optional provider ID
}

// RefineResponse is the response for image refinement.
type RefineResponse struct {
	SessionID string `json:"session_id"`
	Status    string `json:"status"`
	Content   string `json:"content,omitempty"` // Enhanced code/content
	Error     string `json:"error,omitempty"`
}
