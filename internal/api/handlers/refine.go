package handlers

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	polishagent "github.com/paperbanana/paperbanana/internal/agents/polish"
	"github.com/paperbanana/paperbanana/internal/api/dto"
	domainagent "github.com/paperbanana/paperbanana/internal/domain/agent"
	domainllm "github.com/paperbanana/paperbanana/internal/domain/llm"
	domainworkspace "github.com/paperbanana/paperbanana/internal/domain/workspace"
	"go.uber.org/zap"
)

const (
	defaultRefineProjectID     = "default"
	refineSessionSchemaVersion = "agent-session/v1"
)

// RefineSessionSaver persists refine sessions for history restore.
type RefineSessionSaver interface {
	SaveSession(ctx context.Context, session *domainworkspace.SessionRecord) error
}

// RefineHandler handles image refinement requests.
type RefineHandler struct {
	generator    domainllm.LLMClient
	sessionSaver RefineSessionSaver
	logger       *zap.Logger
}

// NewRefineHandler creates a new refine handler.
func NewRefineHandler(generator domainllm.LLMClient, sessionSaver RefineSessionSaver, logger *zap.Logger) *RefineHandler {
	return &RefineHandler{
		generator:    generator,
		sessionSaver: sessionSaver,
		logger:       logger,
	}
}

// Refine handles POST /api/v1/refine for image enhancement.
func (h *RefineHandler) Refine(c *gin.Context) {
	var req dto.RefineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.RefineResponse{
			Error: err.Error(),
		})
		return
	}

	sessionID := strings.TrimSpace(req.SessionID)
	if sessionID == "" {
		sessionID = uuid.NewString()
	}

	// Validate request
	if req.ImageData == "" {
		c.JSON(http.StatusBadRequest, dto.RefineResponse{
			SessionID: sessionID,
			Error:     "image_data is required",
		})
		return
	}
	if req.Instructions == "" {
		c.JSON(http.StatusBadRequest, dto.RefineResponse{
			SessionID: sessionID,
			Error:     "instructions is required",
		})
		return
	}

	// Set default resolution
	resolution := req.Resolution
	if resolution == "" {
		resolution = "2K"
	}
	if resolution != "2K" && resolution != "4K" {
		c.JSON(http.StatusBadRequest, dto.RefineResponse{
			SessionID: sessionID,
			Error:     "resolution must be '2K' or '4K'",
		})
		return
	}

	totalIterations, stopReason, err := resolveIterationPlan(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.RefineResponse{
			SessionID: sessionID,
			Error:     err.Error(),
		})
		return
	}

	mimeType, imageBytes, err := decodeRefineImage(req.ImageData)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.RefineResponse{
			SessionID: sessionID,
			Error:     "invalid image_data: " + err.Error(),
		})
		return
	}

	if h.generator == nil {
		c.JSON(http.StatusInternalServerError, dto.RefineResponse{
			SessionID: sessionID,
			Error:     "refine generator is not configured",
		})
		return
	}

	requestID := uuid.NewString()
	startedAt := time.Now().UTC()
	currentMimeType := mimeType
	currentImage := append([]byte(nil), imageBytes...)

	finalOutput := domainagent.AgentOutput{}
	for iteration := 1; iteration <= totalIterations; iteration++ {
		output, outputMimeType, outputBytes, runErr := h.runRefineIteration(c.Request.Context(), sessionID, requestID, req, resolution, currentMimeType, currentImage, iteration, totalIterations)
		if runErr != nil {
			h.persistRefineSession(c.Request.Context(), buildFailedRefineSession(
				sessionID,
				requestID,
				req.Instructions,
				startedAt,
				time.Now().UTC(),
				totalIterations,
				stopReason,
				runErr,
			))
			c.JSON(http.StatusInternalServerError, dto.RefineResponse{
				SessionID: sessionID,
				Status:    "failed",
				Error:     "refinement failed: " + runErr.Error(),
				Metadata: &dto.RefineResponseMetadata{
					Iterations: strconv.Itoa(iteration),
					StopReason: stopReason,
				},
			})
			return
		}

		finalOutput = output
		currentMimeType = outputMimeType
		currentImage = outputBytes
	}

	completedAt := time.Now().UTC()
	finalOutput.Metadata = mergeStringMaps(finalOutput.Metadata, map[string]string{
		"summary":            buildRefineSummary(totalIterations),
		"refine.iterations":  strconv.Itoa(totalIterations),
		"refine.stop_reason": stopReason,
	})
	h.persistRefineSession(c.Request.Context(), buildCompletedRefineSession(
		sessionID,
		requestID,
		req.Instructions,
		startedAt,
		completedAt,
		finalOutput,
		totalIterations,
		stopReason,
	))

	// Build artifacts array for the response
	artifacts := []dto.Artifact{
		{
			Type:   "image",
			Format: formatFromMIMEType(currentMimeType),
			Data:   base64.StdEncoding.EncodeToString(currentImage),
			Width:  0, // Could be extracted from image if needed
			Height: 0, // Could be extracted from image if needed
		},
	}

	c.JSON(http.StatusOK, dto.RefineResponse{
		SessionID: sessionID,
		Status:    "completed",
		Content:   finalOutput.Content,
		Image: &dto.RefineImagePayload{
			Data:     base64.StdEncoding.EncodeToString(currentImage),
			MIMEType: currentMimeType,
			Metadata: firstArtifactMetadata(finalOutput.GeneratedArtifacts),
		},
		ImageData: base64.StdEncoding.EncodeToString(currentImage),
		Metadata: &dto.RefineResponseMetadata{
			Iterations: strconv.Itoa(totalIterations),
			StopReason: stopReason,
		},
		Artifacts: artifacts,
		IterationInfo: &dto.IterationInfo{
			Enabled:         req.EnableIteration,
			RoundsCompleted: totalIterations,
			MaxRounds:       req.MaxIterations,
		},
	})
}

func (h *RefineHandler) runRefineIteration(
	ctx context.Context,
	sessionID string,
	requestID string,
	req dto.RefineRequest,
	resolution string,
	mimeType string,
	imageBytes []byte,
	iteration int,
	totalIterations int,
) (domainagent.AgentOutput, string, []byte, error) {
	input := domainagent.AgentInput{
		SessionID: sessionID,
		RequestID: requestID,
		Stage:     domainagent.StagePolish,
		Content:   req.Instructions,
		Messages: []domainllm.Message{
			{
				Role: domainllm.RoleUser,
				Parts: []domainllm.Part{
					domainllm.InlineImagePart(mimeType, imageBytes),
				},
			},
		},
		Metadata: map[string]string{
			"refine.iteration":        strconv.Itoa(iteration),
			"refine.total_iterations": strconv.Itoa(totalIterations),
			"refine.enable_iteration": strconv.FormatBool(req.EnableIteration),
			"http.prompt":             req.Instructions,
		},
	}

	config := polishagent.Config{
		Model:      resolvedRefineModel(req.ProviderID, req.Model),
		Resolution: resolution,
	}
	agent := polishagent.NewAgent(h.generator, config, h.logger)

	if err := agent.Initialize(ctx); err != nil {
		return domainagent.AgentOutput{}, "", nil, fmt.Errorf("failed to initialize agent: %w", err)
	}
	defer func() {
		_ = agent.Cleanup(ctx)
	}()

	output, err := agent.Execute(ctx, input)
	if err != nil {
		return domainagent.AgentOutput{}, "", nil, err
	}

	output.Metadata = mergeStringMaps(output.Metadata, map[string]string{
		"refine.iteration":        strconv.Itoa(iteration),
		"refine.total_iterations": strconv.Itoa(totalIterations),
	})

	for _, artifact := range output.GeneratedArtifacts {
		if artifact.Kind == domainagent.ArtifactKindPolishedImage && len(artifact.Bytes) > 0 {
			return output, artifact.MIMEType, append([]byte(nil), artifact.Bytes...), nil
		}
	}

	return domainagent.AgentOutput{}, "", nil, fmt.Errorf("refine iteration %d returned no image artifact", iteration)
}

func resolveIterationPlan(req dto.RefineRequest) (int, string, error) {
	if !req.EnableIteration {
		return 1, "accepted", nil
	}

	maxIterations := req.MaxIterations
	if maxIterations == 0 {
		maxIterations = 3
	}
	if maxIterations < 1 || maxIterations > 5 {
		return 0, "", fmt.Errorf("max_iterations must be between 1 and 5")
	}
	if maxIterations == 1 {
		return 1, "accepted", nil
	}
	return maxIterations, "max_iterations", nil
}

func decodeRefineImage(raw string) (string, []byte, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", nil, fmt.Errorf("image_data is empty")
	}

	if strings.HasPrefix(raw, "data:") {
		comma := strings.Index(raw, ",")
		if comma <= 5 {
			return "", nil, fmt.Errorf("invalid data URL")
		}

		metadata := raw[:comma]
		payload := raw[comma+1:]
		if !strings.Contains(metadata, ";base64") {
			return "", nil, fmt.Errorf("data URL must be base64 encoded")
		}

		decoded, err := base64.StdEncoding.DecodeString(payload)
		if err != nil {
			return "", nil, err
		}

		mimeType := strings.TrimPrefix(strings.Split(metadata, ";")[0], "data:")
		if mimeType == "" {
			mimeType = detectMIMEType(decoded)
		}
		return mimeType, decoded, nil
	}

	decoded, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return "", nil, err
	}
	return detectMIMEType(decoded), decoded, nil
}

func resolvedRefineModel(providerID, model string) string {
	providerID = strings.TrimSpace(providerID)
	model = strings.TrimSpace(model)
	if providerID == "" {
		return model
	}
	if model == "" {
		return providerID + ":"
	}
	return providerID + ":" + model
}

func buildCompletedRefineSession(
	sessionID string,
	requestID string,
	instructions string,
	startedAt time.Time,
	completedAt time.Time,
	finalOutput domainagent.AgentOutput,
	totalIterations int,
	stopReason string,
) *domainworkspace.SessionRecord {
	initialInput := domainagent.AgentInput{
		SessionID: sessionID,
		RequestID: requestID,
		Stage:     domainagent.StagePolish,
		Content:   instructions,
		Metadata: map[string]string{
			"http.prompt":             instructions,
			"refine.iterations":       strconv.Itoa(totalIterations),
			"refine.stop_reason":      stopReason,
			"refine.enable_iteration": strconv.FormatBool(totalIterations > 1),
		},
	}

	stageState := domainagent.AgentState{
		Stage:  domainagent.StagePolish,
		Status: domainagent.StatusCompleted,
		Timing: domainagent.Timing{
			StartedAt:   startedAt,
			CompletedAt: completedAt,
			Duration:    completedAt.Sub(startedAt),
		},
		Input:  initialInput,
		Output: finalOutput,
	}

	snapshot := &domainagent.SessionState{
		SchemaVersion: refineSessionSchemaVersion,
		SessionID:     sessionID,
		RequestID:     requestID,
		Status:        domainagent.StatusCompleted,
		CurrentStage:  domainagent.StagePolish,
		Pipeline:      []domainagent.StageName{domainagent.StagePolish},
		InitialInput:  initialInput,
		StageStates:   []domainagent.AgentState{stageState},
		FinalOutput:   finalOutput,
		Metadata: map[string]string{
			"http.prompt":        instructions,
			"refine.iterations":  strconv.Itoa(totalIterations),
			"refine.stop_reason": stopReason,
		},
		StartedAt:   startedAt,
		UpdatedAt:   completedAt,
		CompletedAt: completedAt,
	}

	return &domainworkspace.SessionRecord{
		ID:            sessionID,
		ProjectID:     defaultRefineProjectID,
		Status:        string(domainagent.StatusCompleted),
		CurrentStage:  string(domainagent.StagePolish),
		SchemaVersion: snapshot.SchemaVersion,
		Snapshot:      snapshot,
		CreatedAt:     startedAt,
		UpdatedAt:     completedAt,
		CompletedAt:   &completedAt,
	}
}

func buildFailedRefineSession(
	sessionID string,
	requestID string,
	instructions string,
	startedAt time.Time,
	failedAt time.Time,
	totalIterations int,
	stopReason string,
	runErr error,
) *domainworkspace.SessionRecord {
	initialInput := domainagent.AgentInput{
		SessionID: sessionID,
		RequestID: requestID,
		Stage:     domainagent.StagePolish,
		Content:   instructions,
		Metadata: map[string]string{
			"http.prompt":        instructions,
			"refine.iterations":  strconv.Itoa(totalIterations),
			"refine.stop_reason": stopReason,
		},
	}

	errDetail := &domainagent.ErrorDetail{
		Message: runErr.Error(),
		Stage:   domainagent.StagePolish,
	}
	snapshot := &domainagent.SessionState{
		SchemaVersion: refineSessionSchemaVersion,
		SessionID:     sessionID,
		RequestID:     requestID,
		Status:        domainagent.StatusFailed,
		CurrentStage:  domainagent.StagePolish,
		FailedStage:   domainagent.StagePolish,
		Pipeline:      []domainagent.StageName{domainagent.StagePolish},
		InitialInput:  initialInput,
		Error:         errDetail,
		Metadata: map[string]string{
			"http.prompt":        instructions,
			"refine.iterations":  strconv.Itoa(totalIterations),
			"refine.stop_reason": stopReason,
		},
		StartedAt:   startedAt,
		UpdatedAt:   failedAt,
		CompletedAt: failedAt,
	}

	return &domainworkspace.SessionRecord{
		ID:            sessionID,
		ProjectID:     defaultRefineProjectID,
		Status:        string(domainagent.StatusFailed),
		CurrentStage:  string(domainagent.StagePolish),
		SchemaVersion: snapshot.SchemaVersion,
		Snapshot:      snapshot,
		CreatedAt:     startedAt,
		UpdatedAt:     failedAt,
		CompletedAt:   &failedAt,
	}
}

func (h *RefineHandler) persistRefineSession(ctx context.Context, session *domainworkspace.SessionRecord) {
	if h.sessionSaver == nil || session == nil {
		return
	}
	if err := h.sessionSaver.SaveSession(ctx, session); err != nil {
		h.logger.Warn("failed to persist refine session", zap.Error(err), zap.String("session_id", session.ID))
	}
}

func buildRefineSummary(totalIterations int) string {
	if totalIterations <= 1 {
		return "refined image artifact"
	}
	return fmt.Sprintf("refined image artifact after %d iterations", totalIterations)
}

func mergeStringMaps(base map[string]string, extra map[string]string) map[string]string {
	if len(base) == 0 && len(extra) == 0 {
		return nil
	}

	merged := cloneStringMap(base)
	if merged == nil {
		merged = make(map[string]string, len(extra))
	}
	for key, value := range extra {
		merged[key] = value
	}
	return merged
}

func cloneStringMap(values map[string]string) map[string]string {
	if len(values) == 0 {
		return nil
	}

	cloned := make(map[string]string, len(values))
	for key, value := range values {
		cloned[key] = value
	}
	return cloned
}

func firstArtifactMetadata(artifacts []domainagent.Artifact) map[string]string {
	if len(artifacts) == 0 {
		return nil
	}
	return cloneStringMap(artifacts[0].Metadata)
}

// formatFromMIMEType returns the image format string from MIME type.
func formatFromMIMEType(mimeType string) string {
	switch mimeType {
	case "image/png":
		return "png"
	case "image/jpeg", "image/jpg":
		return "jpeg"
	case "image/gif":
		return "gif"
	case "image/webp":
		return "webp"
	case "image/svg+xml":
		return "svg"
	default:
		return "png"
	}
}

// detectMIMEType attempts to detect the MIME type from image bytes.
func detectMIMEType(data []byte) string {
	if len(data) < 4 {
		return "image/png" // default
	}

	// PNG: 89 50 4E 47
	if data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47 {
		return "image/png"
	}
	// JPEG: FF D8 FF
	if data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF {
		return "image/jpeg"
	}
	// GIF: 47 49 46 38
	if data[0] == 0x47 && data[1] == 0x49 && data[2] == 0x46 && data[3] == 0x38 {
		return "image/gif"
	}
	// WebP: 52 49 46 46 (RIFF) + later WEBP
	if len(data) >= 12 && data[0] == 0x52 && data[1] == 0x49 && data[2] == 0x46 && data[3] == 0x46 {
		if data[8] == 0x57 && data[9] == 0x45 && data[10] == 0x42 && data[11] == 0x50 {
			return "image/webp"
		}
	}

	return "image/png" // default to PNG
}
