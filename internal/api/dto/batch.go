// Package dto provides Data Transfer Objects for API requests and responses.
package dto

import domainagent "github.com/paperbanana/paperbanana/internal/domain/agent"

// BatchGenerateRequest is the request body for batch generation.
type BatchGenerateRequest struct {
	Prompt         string  `json:"prompt"`
	Mode           string  `json:"mode"`
	Model          string  `json:"model"`
	SessionID      string  `json:"session_id"`
	Resume         bool    `json:"resume"`
	VisualizerNode string  `json:"visualizer_node"`
	ProjectID      string  `json:"project_id"`
	FolderID       *string `json:"folder_id"`
	// NumCandidates is the number of parallel candidates to generate (default: 1, max: 50).
	NumCandidates int `json:"num_candidates"`

	// Config fields for generation parameters
	AspectRatio   string `json:"aspect_ratio"`   // "21:9", "16:9", "3:2"
	CriticRounds  int    `json:"critic_rounds"`  // 1-5
	RetrievalMode string `json:"retrieval_mode"` // "auto", "manual", "random", "none"
	PipelineMode  string `json:"pipeline_mode"`  // "full", "planner-critic", "vanilla"
	QueryModel    string `json:"query_model"`
	GenModel      string `json:"gen_model"`
}

// BatchGenerateResponse is the final response for batch generation.
type BatchGenerateResponse struct {
	BatchID    string                   `json:"batch_id"`
	Results    []CandidateResultDTO     `json:"results"`
	Successful int                      `json:"successful"`
	Failed     int                      `json:"failed"`
	Timing     domainagent.BatchTiming  `json:"timing"`
}

// CandidateResultDTO represents the result of a single candidate.
type CandidateResultDTO struct {
	CandidateID int                      `json:"candidate_id"`
	SessionID   string                   `json:"session_id"`
	Status      string                   `json:"status"`
	Artifacts   []domainagent.Artifact   `json:"artifacts,omitempty"`
	Error       *domainagent.ErrorDetail `json:"error,omitempty"`
}

// FromCandidateResult converts a domain CandidateResult to DTO.
func FromCandidateResult(cr domainagent.CandidateResult) CandidateResultDTO {
	return CandidateResultDTO{
		CandidateID: cr.CandidateID,
		SessionID:   cr.SessionID,
		Status:      string(cr.Status),
		Artifacts:   cr.Artifacts,
		Error:       cr.Error,
	}
}

// FromBatchResult converts a domain BatchResult to response DTO.
func FromBatchResult(br domainagent.BatchResult) BatchGenerateResponse {
	results := make([]CandidateResultDTO, len(br.Results))
	for i, r := range br.Results {
		results[i] = FromCandidateResult(r)
	}
	return BatchGenerateResponse{
		BatchID:    br.BatchID,
		Results:    results,
		Successful: br.Successful,
		Failed:     br.Failed,
		Timing:     br.Timing,
	}
}

// BatchDownloadRequest is the request body for batch ZIP download.
type BatchDownloadRequest struct {
	BatchID string `json:"batch_id" binding:"required"`
}
