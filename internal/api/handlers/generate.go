package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/paperbanana/paperbanana/internal/application/orchestrator"
	domainagent "github.com/paperbanana/paperbanana/internal/domain/agent"
	"go.uber.org/zap"
)

type Handler struct {
	runner       Runner
	logger       *zap.Logger
	registry     *SessionRegistry // Optional: for session cancellation
	assetService *AssetServiceAdapter // Optional: for persisting artifacts
}

type RunHandle interface {
	Events() <-chan domainagent.Event
	Wait() (orchestrator.RunResult, error)
}

type Runner interface {
	Start(ctx context.Context, input domainagent.AgentInput) (RunHandle, error)
	Resume(ctx context.Context, input domainagent.AgentInput) (RunHandle, error)
}

type orchestratorRunnerAdapter struct {
	runner *orchestrator.Runner
}

func NewRunnerAdapter(runner *orchestrator.Runner) Runner {
	return orchestratorRunnerAdapter{runner: runner}
}

func (a orchestratorRunnerAdapter) Start(ctx context.Context, input domainagent.AgentInput) (RunHandle, error) {
	return a.runner.Start(ctx, input)
}

func (a orchestratorRunnerAdapter) Resume(ctx context.Context, input domainagent.AgentInput) (RunHandle, error) {
	return a.runner.Resume(ctx, input)
}

func NewHandler(runner Runner, logger *zap.Logger) *Handler {
	return &Handler{runner: runner, logger: logger}
}

// NewHandlerWithRegistry creates a Handler with session registry for cancellation support.
func NewHandlerWithRegistry(runner Runner, registry *SessionRegistry, logger *zap.Logger) *Handler {
	return &Handler{runner: runner, logger: logger, registry: registry}
}

// NewHandlerWithAssetService creates a Handler with asset persistence support.
func NewHandlerWithAssetService(runner Runner, registry *SessionRegistry, assetService *AssetServiceAdapter, logger *zap.Logger) *Handler {
	return &Handler{runner: runner, logger: logger, registry: registry, assetService: assetService}
}

type GenerateRequest struct {
	Prompt         string  `json:"prompt"`
	Content        string  `json:"content"`
	VisualIntent   string  `json:"visual_intent"`
	Mode           string  `json:"mode"`
	Model          string  `json:"model"`
	Temperature    float64 `json:"temperature"`
	MaxTokens      int     `json:"max_tokens"`
	SessionID      string  `json:"session_id"`
	Resume         bool    `json:"resume"`
	VisualizerNode string  `json:"visualizer_node"`
	AspectRatio    string  `json:"aspect_ratio"`
	CriticRounds   int     `json:"critic_rounds"`
	RetrievalMode  string  `json:"retrieval_mode"`
	PipelineMode   string  `json:"pipeline_mode"`
	QueryModel     string  `json:"query_model"`
	GenModel       string  `json:"gen_model"`
	// Project ownership for persistence
	ProjectID       string  `json:"project_id"`
	FolderID        *string `json:"folder_id"`
	VisualizationID *string `json:"visualization_id"`
}

type GenerateResponse struct {
	SessionID          string                 `json:"session_id"`
	RequestID          string                 `json:"request_id"`
	ProjectID          string                 `json:"project_id,omitempty"`
	Content            string                 `json:"content"`
	GeneratedArtifacts []domainagent.Artifact `json:"generated_artifacts,omitempty"`
	TokensUsed         int                    `json:"tokens_used"`
	FinishReason       string                 `json:"finish_reason"`
}

func (h *Handler) Generate(c *gin.Context) {
	var req GenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := validateGenerateRequest(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	handle, _, err := h.startRun(c.Request.Context(), req)
	if err != nil {
		h.respondRunError(c, err)
		return
	}

	go func() {
		for range handle.Events() {
		}
	}()

	result, err := handle.Wait()
	if err != nil {
		h.respondRunError(c, err)
		return
	}

	// Persist artifacts if asset service is configured
	artifacts := result.Session.FinalOutput.GeneratedArtifacts
	if h.assetService != nil && len(artifacts) > 0 && req.ProjectID != "" {
		artifacts = h.persistArtifacts(c.Request.Context(), req, artifacts)
	}

	c.JSON(http.StatusOK, buildGenerateResponseWithArtifacts(result, req.ProjectID, artifacts))
}

func (h *Handler) StreamGenerate(c *gin.Context) {
	var req GenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := validateGenerateRequest(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	// Get session ID for potential cancellation
	input, err := buildAgentInput(req)
	if err != nil {
		h.logger.Error("failed to build agent input", zap.Error(err))
		c.SSEvent("error", gin.H{"error": err.Error()})
		c.Writer.Flush()
		return
	}

	// Use registry for cancellable context if available
	ctx := c.Request.Context()
	var cancel context.CancelFunc
	if h.registry != nil {
		ctx, cancel = h.registry.Register(input.SessionID, ctx)
		defer cancel()
	}

	handle, err := h.startRunWithInput(ctx, input, req.Resume)
	if err != nil {
		h.logger.Error("stream generation failed to start", zap.Error(err))
		c.SSEvent("error", gin.H{"error": err.Error()})
		c.Writer.Flush()
		return
	}

	for event := range handle.Events() {
		c.SSEvent(string(event.Type), event)
		c.Writer.Flush()
	}

	result, err := handle.Wait()
	if err != nil {
		h.logger.Error("stream generation failed", zap.Error(err))
		c.SSEvent("error", buildStreamErrorPayload(err, input, result))
		c.Writer.Flush()
		return
	}

	// Persist artifacts if asset service is configured
	artifacts := result.Session.FinalOutput.GeneratedArtifacts
	if h.assetService != nil && len(artifacts) > 0 && req.ProjectID != "" {
		artifacts = h.persistArtifacts(c.Request.Context(), req, artifacts)
	}

	c.SSEvent("result", buildGenerateResponseWithArtifacts(result, req.ProjectID, artifacts))
	c.Writer.Flush()
}

func (h *Handler) startRun(ctx context.Context, req GenerateRequest) (RunHandle, domainagent.AgentInput, error) {
	input, err := buildAgentInput(req)
	if err != nil {
		return nil, domainagent.AgentInput{}, err
	}

	handle, err := h.startRunWithInput(ctx, input, req.Resume)
	if err != nil {
		return nil, domainagent.AgentInput{}, err
	}
	return handle, input, nil
}

func (h *Handler) startRunWithInput(ctx context.Context, input domainagent.AgentInput, resume bool) (RunHandle, error) {
	var handle RunHandle
	var err error
	if resume {
		handle, err = h.runner.Resume(ctx, input)
	} else {
		handle, err = h.runner.Start(ctx, input)
	}
	if err != nil {
		return nil, err
	}
	return handle, nil
}

func (h *Handler) respondRunError(c *gin.Context, err error) {
	h.logger.Error("generation failed", zap.Error(err))
	status := http.StatusInternalServerError
	if errors.Is(err, orchestrator.ErrResumeRequiresSession) ||
		errors.Is(err, orchestrator.ErrResumeSnapshotMissing) ||
		errors.Is(err, orchestrator.ErrResumeStoreMissing) ||
		strings.HasPrefix(err.Error(), "unsupported mode ") {
		status = http.StatusBadRequest
	}
	c.JSON(status, gin.H{"error": err.Error()})
}

// Validation constants for input validation.
const (
	MaxPromptLength      = 100000 // Maximum prompt length in characters
	MaxContentLength     = 500000 // Maximum content length in characters
	MaxVisualIntentLength = 10000 // Maximum visual_intent length in characters
	MaxTemperature       = 2.0
	MinTemperature       = 0.0
	MaxMaxTokens         = 100000
)

var validModes = map[string]bool{
	"":       true, // Default mode
	"diagram": true,
	"plot":    true,
}

func validateGenerateRequest(req GenerateRequest) error {
	if req.Resume {
		if strings.TrimSpace(req.SessionID) == "" {
			return orchestrator.ErrResumeRequiresSession
		}
		return nil
	}

	// Validate prompt length
	if len(req.Prompt) > MaxPromptLength {
		return fmt.Errorf("prompt exceeds maximum length of %d characters", MaxPromptLength)
	}

	// Validate content length
	if len(req.Content) > MaxContentLength {
		return fmt.Errorf("content exceeds maximum length of %d characters", MaxContentLength)
	}

	// Validate visual_intent length
	if len(req.VisualIntent) > MaxVisualIntentLength {
		return fmt.Errorf("visual_intent exceeds maximum length of %d characters", MaxVisualIntentLength)
	}

	// Require at least one input
	if strings.TrimSpace(req.Prompt) == "" &&
		strings.TrimSpace(req.Content) == "" &&
		strings.TrimSpace(req.VisualIntent) == "" {
		return errors.New("prompt, content, or visual_intent is required")
	}

	// Validate mode
	if !validModes[req.Mode] {
		return fmt.Errorf("invalid mode %q: must be one of diagram, plot, or empty", req.Mode)
	}

	// Validate temperature
	if req.Temperature < MinTemperature || req.Temperature > MaxTemperature {
		return fmt.Errorf("temperature must be between %.1f and %.1f", MinTemperature, MaxTemperature)
	}

	// Validate max_tokens
	if req.MaxTokens < 0 {
		return errors.New("max_tokens must be non-negative")
	}
	if req.MaxTokens > MaxMaxTokens {
		return fmt.Errorf("max_tokens exceeds maximum of %d", MaxMaxTokens)
	}

	// Validate critic_rounds
	if req.CriticRounds < 0 || req.CriticRounds > 5 {
		return errors.New("critic_rounds must be between 0 and 5")
	}

	// Validate aspect_ratio format if provided
	if req.AspectRatio != "" && !isValidAspectRatio(req.AspectRatio) {
		return fmt.Errorf("invalid aspect_ratio format: %s (expected format like '16:9' or '1.5')", req.AspectRatio)
	}

	return nil
}

// isValidAspectRatio validates aspect ratio format.
func isValidAspectRatio(ratio string) bool {
	// Allow decimal format (e.g., "1.5", "0.75")
	if _, err := strconv.ParseFloat(ratio, 64); err == nil {
		return true
	}

	// Allow colon format (e.g., "16:9", "4:3")
	parts := strings.Split(ratio, ":")
	if len(parts) == 2 {
		_, err1 := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
		_, err2 := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
		return err1 == nil && err2 == nil
	}

	return false
}

func buildGenerateResponse(result orchestrator.RunResult, projectID string) GenerateResponse {
	return GenerateResponse{
		SessionID:          result.Session.SessionID,
		RequestID:          result.Session.RequestID,
		ProjectID:          projectID,
		Content:            result.Session.FinalOutput.Content,
		GeneratedArtifacts: cloneArtifactsWithProjectID(result.Session.FinalOutput.GeneratedArtifacts, projectID),
		TokensUsed:         0,
		FinishReason:       string(result.Session.Status),
	}
}

func buildGenerateResponseWithArtifacts(result orchestrator.RunResult, projectID string, artifacts []domainagent.Artifact) GenerateResponse {
	return GenerateResponse{
		SessionID:          result.Session.SessionID,
		RequestID:          result.Session.RequestID,
		ProjectID:          projectID,
		Content:            result.Session.FinalOutput.Content,
		GeneratedArtifacts: cloneArtifactsWithProjectID(artifacts, projectID),
		TokensUsed:         0,
		FinishReason:       string(result.Session.Status),
	}
}

// persistArtifacts stores artifacts via AssetService and updates them with asset IDs.
func (h *Handler) persistArtifacts(ctx context.Context, req GenerateRequest, artifacts []domainagent.Artifact) []domainagent.Artifact {
	if h.assetService == nil || req.ProjectID == "" || len(artifacts) == 0 {
		return artifacts
	}

	// Get visualization ID from request or use session ID as fallback
	visualizationID := req.SessionID
	if req.VisualizationID != nil && *req.VisualizationID != "" {
		visualizationID = *req.VisualizationID
	}

	// Register artifacts as retained assets
	assets, err := h.assetService.RegisterRetainedAssets(ctx, req.ProjectID, visualizationID, nil, artifacts)
	if err != nil {
		h.logger.Warn("failed to persist artifacts", zap.Error(err), zap.String("session_id", req.SessionID))
		return artifacts
	}

	// Build a map from artifact kind to asset ID for updating
	assetMap := make(map[domainagent.ArtifactKind]string)
	for _, asset := range assets {
		// Find matching artifact by mime type and kind
		for _, artifact := range artifacts {
			if artifact.MIMEType == asset.MIMEType {
				assetMap[artifact.Kind] = asset.ID
				break
			}
		}
	}

	// Update artifacts with asset IDs
	updated := make([]domainagent.Artifact, len(artifacts))
	for i, artifact := range artifacts {
		updated[i] = artifact
		if assetID, ok := assetMap[artifact.Kind]; ok {
			updated[i].AssetID = assetID
		}
		if len(artifact.Bytes) > 0 {
			updated[i].Bytes = append([]byte(nil), artifact.Bytes...)
		}
		if len(artifact.Metadata) > 0 {
			updated[i].Metadata = make(map[string]string, len(artifact.Metadata))
			for key, value := range artifact.Metadata {
				updated[i].Metadata[key] = value
			}
		}
	}

	return updated
}

func cloneArtifactsWithProjectID(artifacts []domainagent.Artifact, projectID string) []domainagent.Artifact {
	if len(artifacts) == 0 {
		return nil
	}

	cloned := make([]domainagent.Artifact, len(artifacts))
	for i, artifact := range artifacts {
		// Use Artifact.Clone() which shares SharedBytes instead of deep copying
		cloned[i] = artifact.Clone()
		// Keep legacy Bytes field in sync for JSON serialization
		if artifact.Shared != nil {
			cloned[i].Bytes = artifact.Bytes
		} else if len(artifact.Bytes) > 0 {
			cloned[i].Bytes = append([]byte(nil), artifact.Bytes...)
		}
		// Set project_id on each artifact for frontend URL construction
		if cloned[i].AssetID != "" && cloned[i].Metadata == nil {
			cloned[i].Metadata = make(map[string]string)
		}
		if cloned[i].AssetID != "" {
			cloned[i].Metadata["project_id"] = projectID
		}
	}

	return cloned
}

func buildAgentInput(req GenerateRequest) (domainagent.AgentInput, error) {
	if req.Resume && strings.TrimSpace(req.SessionID) == "" {
		return domainagent.AgentInput{}, orchestrator.ErrResumeRequiresSession
	}

	mode, err := parseVisualMode(req.Mode)
	if err != nil {
		return domainagent.AgentInput{}, err
	}

	sessionID := strings.TrimSpace(req.SessionID)
	if sessionID == "" {
		sessionID = buildID("session")
	}
	requestID := buildID("request")
	prompt, content, visualIntent := resolvePromptFields(req.Prompt, req.Content, req.VisualIntent)

	metadata := map[string]string{
		"http.prompt":                   prompt,
		"http.content":                  content,
		"http.visual_intent":            visualIntent,
		"http.mode":                     string(mode),
		"http.model":                    req.Model,
		"http.temperature":              strconv.FormatFloat(req.Temperature, 'f', -1, 64),
		"http.max_tokens":               strconv.Itoa(req.MaxTokens),
		"http.session_id":               sessionID,
		"http.resume":                   strconv.FormatBool(req.Resume),
		"http.project_id":               req.ProjectID,
		"config.runtime_managed_models": "true",
	}
	if nodeName := strings.TrimSpace(req.VisualizerNode); nodeName != "" {
		metadata["http.visualizer_node"] = nodeName
		metadata["visualizer.node_name"] = nodeName
	}
	if req.FolderID != nil {
		metadata["http.folder_id"] = *req.FolderID
	}
	if req.VisualizationID != nil {
		metadata["http.visualization_id"] = *req.VisualizationID
	}
	if value := strings.TrimSpace(req.AspectRatio); value != "" {
		metadata["config.aspect_ratio"] = value
	}
	if req.CriticRounds > 0 {
		metadata["config.critic_rounds"] = strconv.Itoa(req.CriticRounds)
	}
	if value := strings.TrimSpace(req.RetrievalMode); value != "" {
		metadata["config.retrieval_mode"] = value
		metadata["retrieval_setting"] = value
	}
	if value := strings.TrimSpace(req.PipelineMode); value != "" {
		metadata["config.pipeline_mode"] = value
	}
	if value := strings.TrimSpace(req.QueryModel); value != "" {
		metadata["config.query_model"] = value
	}
	if value := strings.TrimSpace(req.GenModel); value != "" {
		metadata["config.gen_model"] = value
	}

	return domainagent.AgentInput{
		SessionID: sessionID,
		RequestID: requestID,
		Content:   content,
		VisualIntent: domainagent.VisualIntent{
			Mode:             mode,
			Goal:             visualIntent,
			Style:            "academic",
			PreferredOutputs: []string{"png"},
		},
		Metadata: metadata,
	}, nil
}

func cloneArtifacts(artifacts []domainagent.Artifact) []domainagent.Artifact {
	if len(artifacts) == 0 {
		return nil
	}

	cloned := make([]domainagent.Artifact, len(artifacts))
	for i, artifact := range artifacts {
		// Use Artifact.Clone() which shares SharedBytes instead of deep copying
		cloned[i] = artifact.Clone()
		// Keep legacy Bytes field in sync for JSON serialization
		if artifact.Shared != nil {
			cloned[i].Bytes = artifact.Bytes
		} else if len(artifact.Bytes) > 0 {
			cloned[i].Bytes = append([]byte(nil), artifact.Bytes...)
		}
	}

	return cloned
}

func parseVisualMode(raw string) (domainagent.VisualMode, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", string(domainagent.VisualModeDiagram):
		return domainagent.VisualModeDiagram, nil
	case string(domainagent.VisualModePlot):
		return domainagent.VisualModePlot, nil
	default:
		return "", fmt.Errorf("unsupported mode %q", raw)
	}
}

func buildID(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UTC().UnixNano())
}

func buildStreamErrorPayload(err error, input domainagent.AgentInput, result orchestrator.RunResult) gin.H {
	payload := gin.H{
		"message":    err.Error(),
		"error":      err.Error(),
		"session_id": input.SessionID,
		"request_id": input.RequestID,
	}

	if detail := result.Session.Error; detail != nil {
		payload["message"] = detail.Message
		payload["error"] = detail.Message
		payload["code"] = detail.Code
		payload["category"] = detail.Category
		payload["retryable"] = detail.Retryable
		payload["suggestion"] = detail.Suggestion
		if detail.Stage != "" {
			payload["stage"] = detail.Stage
		}
	}

	if result.FailedStage != "" {
		payload["failed_stage"] = result.FailedStage
	}

	return payload
}

func resolvePromptFields(prompt, content, visualIntent string) (string, string, string) {
	trimmedPrompt := strings.TrimSpace(prompt)
	trimmedContent := strings.TrimSpace(content)
	trimmedVisualIntent := strings.TrimSpace(visualIntent)

	if trimmedContent == "" && trimmedVisualIntent == "" {
		return trimmedPrompt, trimmedPrompt, trimmedPrompt
	}

	if trimmedPrompt == "" {
		trimmedPrompt = buildCombinedPrompt(trimmedContent, trimmedVisualIntent)
	}
	if trimmedContent == "" {
		trimmedContent = trimmedPrompt
	}
	if trimmedVisualIntent == "" {
		trimmedVisualIntent = trimmedPrompt
	}

	return trimmedPrompt, trimmedContent, trimmedVisualIntent
}

func buildCombinedPrompt(content, visualIntent string) string {
	switch {
	case content != "" && visualIntent != "":
		return fmt.Sprintf("Paper Context & References:\n%s\n\nTarget Figure Brief:\n%s", content, visualIntent)
	case content != "":
		return content
	default:
		return visualIntent
	}
}
