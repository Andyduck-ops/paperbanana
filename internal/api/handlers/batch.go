package handlers

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/paperbanana/paperbanana/internal/api/dto"
	"github.com/paperbanana/paperbanana/internal/application/orchestrator"
	domainagent "github.com/paperbanana/paperbanana/internal/domain/agent"
	"go.uber.org/zap"
)

// MaxBatchCandidates is the maximum number of candidates allowed in a batch.
const MaxBatchCandidates = 50

// BatchHandler handles batch generation requests.
type BatchHandler struct {
	batchRunner  *orchestrator.BatchRunner
	agentFactory orchestrator.AgentFactory
	logger       *zap.Logger
}

// NewBatchHandler creates a new batch handler.
func NewBatchHandler(batchRunner *orchestrator.BatchRunner, agentFactory orchestrator.AgentFactory, logger *zap.Logger) *BatchHandler {
	return &BatchHandler{
		batchRunner:  batchRunner,
		agentFactory: agentFactory,
		logger:       logger,
	}
}

// StreamBatchGenerate handles POST /api/v1/generate/batch with SSE streaming.
func (h *BatchHandler) StreamBatchGenerate(c *gin.Context) {
	var req dto.BatchGenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set defaults
	if req.NumCandidates <= 0 {
		req.NumCandidates = 1
	}

	// Validate
	if err := validateBatchRequest(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	// Build inputs
	inputs, err := buildBatchInputs(req)
	if err != nil {
		c.SSEvent("error", gin.H{"error": err.Error()})
		c.Writer.Flush()
		return
	}

	// Start batch
	handle, err := h.batchRunner.StartBatch(c.Request.Context(), inputs)
	if err != nil {
		c.SSEvent("error", gin.H{"error": err.Error()})
		c.Writer.Flush()
		return
	}

	// Stream events
	for event := range handle.Events() {
		c.SSEvent(string(event.Type), event)
		c.Writer.Flush()
	}

	// Wait for result
	result, err := handle.Wait()
	if err != nil {
		c.SSEvent("error", gin.H{"error": err.Error()})
		c.Writer.Flush()
		return
	}

	// Emit final result
	c.SSEvent("batch_result", dto.FromBatchResult(result))
	c.Writer.Flush()
}

func validateBatchRequest(req dto.BatchGenerateRequest) error {
	if strings.TrimSpace(req.Prompt) == "" {
		return errors.New("prompt is required")
	}
	if req.NumCandidates < 1 {
		return errors.New("num_candidates must be at least 1")
	}
	if req.NumCandidates > MaxBatchCandidates {
		return fmt.Errorf("num_candidates must be at most %d", MaxBatchCandidates)
	}
	// Validate critic rounds
	if req.CriticRounds < 0 || req.CriticRounds > 5 {
		return errors.New("critic_rounds must be between 0 and 5")
	}
	return nil
}

func buildBatchInputs(req dto.BatchGenerateRequest) ([]domainagent.AgentInput, error) {
	// Build base input from request
	mode, err := parseVisualMode(req.Mode)
	if err != nil {
		return nil, err
	}

	sessionID := strings.TrimSpace(req.SessionID)
	if sessionID == "" {
		sessionID = buildID("batch")
	}

	inputs := make([]domainagent.AgentInput, req.NumCandidates)
	for i := 0; i < req.NumCandidates; i++ {
		candidateSessionID := fmt.Sprintf("%s-candidate-%d", sessionID, i)
		requestID := buildID("request")

		metadata := map[string]string{
			"http.prompt":        req.Prompt,
			"http.mode":          string(mode),
			"http.model":         req.Model,
			"http.session_id":    candidateSessionID,
			"http.project_id":    req.ProjectID,
			"batch.candidate_id": strconv.Itoa(i),
		}
		if nodeName := strings.TrimSpace(req.VisualizerNode); nodeName != "" {
			metadata["http.visualizer_node"] = nodeName
			metadata["visualizer.node_name"] = nodeName
		}
		if req.FolderID != nil {
			metadata["http.folder_id"] = *req.FolderID
		}

		// Add config fields to metadata
		if req.AspectRatio != "" {
			metadata["config.aspect_ratio"] = req.AspectRatio
		}
		if req.CriticRounds > 0 {
			metadata["config.critic_rounds"] = strconv.Itoa(req.CriticRounds)
		}
		if req.RetrievalMode != "" {
			metadata["config.retrieval_mode"] = req.RetrievalMode
			metadata["retrieval_setting"] = req.RetrievalMode
		}
		if req.PipelineMode != "" {
			metadata["config.pipeline_mode"] = req.PipelineMode
		}
		if req.QueryModel != "" {
			metadata["config.query_model"] = req.QueryModel
		}
		if req.GenModel != "" {
			metadata["config.gen_model"] = req.GenModel
		}

		inputs[i] = domainagent.AgentInput{
			SessionID: candidateSessionID,
			RequestID: requestID,
			Content:   req.Prompt,
			VisualIntent: domainagent.VisualIntent{
				Mode:             mode,
				Goal:             req.Prompt,
				Style:            "academic",
				PreferredOutputs: []string{"png"},
			},
			Metadata: metadata,
		}
	}

	return inputs, nil
}

// DownloadBatchZip handles POST /api/v1/batch/download and returns a ZIP file with all successful candidate artifacts.
func (h *BatchHandler) DownloadBatchZip(c *gin.Context) {
	var req dto.BatchDownloadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "batch_id is required"})
		return
	}

	// Get batch result from runner's result store
	result, err := h.batchRunner.GetBatchResult(req.BatchID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "batch not found or expired"})
		return
	}

	// Set response headers for ZIP download
	filename := fmt.Sprintf("paperviz_candidates_%s.zip", time.Now().Format("20060102_150405"))
	c.Header("Content-Type", "application/zip")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Transfer-Encoding", "chunked")

	// Stream ZIP directly to response
	zipWriter := zip.NewWriter(c.Writer)
	defer zipWriter.Close()

	// Add successful candidate artifacts
	for _, candidate := range result.Results {
		if candidate.Status != domainagent.StatusCompleted {
			continue
		}
		for i, artifact := range candidate.Artifacts {
			var content []byte
			if artifact.Bytes != nil {
				content = artifact.Bytes
			} else if artifact.Content != "" {
				content = []byte(artifact.Content)
			}
			if len(content) == 0 {
				continue
			}

			name := fmt.Sprintf("candidate_%d_artifact_%d.png", candidate.CandidateID, i)
			w, err := zipWriter.Create(name)
			if err != nil {
				continue
			}
			w.Write(content)
		}
	}

	// Add metadata.json
	meta := map[string]interface{}{
		"batch_id":     result.BatchID,
		"generated_at": time.Now().Format(time.RFC3339),
		"successful":   result.Successful,
		"failed":       result.Failed,
		"total":        len(result.Results),
		"timing":       result.Timing,
		"candidates":   buildCandidateMetadata(result.Results),
	}
	metaW, _ := zipWriter.Create("metadata.json")
	json.NewEncoder(metaW).Encode(meta)
}

// CandidateMeta represents metadata for a single candidate in the ZIP.
type CandidateMeta struct {
	CandidateID int    `json:"candidate_id"`
	SessionID   string `json:"session_id"`
	Status      string `json:"status"`
	Artifacts   int    `json:"artifacts"`
	Error       string `json:"error,omitempty"`
}

func buildCandidateMetadata(results []domainagent.CandidateResult) []CandidateMeta {
	metas := make([]CandidateMeta, len(results))
	for i, r := range results {
		metas[i] = CandidateMeta{
			CandidateID: r.CandidateID,
			SessionID:   r.SessionID,
			Status:      string(r.Status),
			Artifacts:   len(r.Artifacts),
		}
		if r.Error != nil {
			metas[i].Error = r.Error.Message
		}
	}
	return metas
}
