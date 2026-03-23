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
	runner Runner
	logger *zap.Logger
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

type GenerateRequest struct {
	Prompt         string  `json:"prompt"`
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

	c.JSON(http.StatusOK, buildGenerateResponse(result))
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

	handle, input, err := h.startRun(c.Request.Context(), req)
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
		c.SSEvent("error", gin.H{"error": err.Error(), "session_id": input.SessionID, "request_id": input.RequestID})
		c.Writer.Flush()
		return
	}

	c.SSEvent("result", buildGenerateResponse(result))
	c.Writer.Flush()
}

func (h *Handler) startRun(ctx context.Context, req GenerateRequest) (RunHandle, domainagent.AgentInput, error) {
	input, err := buildAgentInput(req)
	if err != nil {
		return nil, domainagent.AgentInput{}, err
	}

	var handle RunHandle
	if req.Resume {
		handle, err = h.runner.Resume(ctx, input)
	} else {
		handle, err = h.runner.Start(ctx, input)
	}
	if err != nil {
		return nil, domainagent.AgentInput{}, err
	}
	return handle, input, nil
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

func validateGenerateRequest(req GenerateRequest) error {
	if req.Resume {
		if strings.TrimSpace(req.SessionID) == "" {
			return orchestrator.ErrResumeRequiresSession
		}
		return nil
	}

	if strings.TrimSpace(req.Prompt) == "" {
		return errors.New("prompt is required")
	}
	if req.CriticRounds < 0 || req.CriticRounds > 5 {
		return errors.New("critic_rounds must be between 0 and 5")
	}

	return nil
}

func buildGenerateResponse(result orchestrator.RunResult) GenerateResponse {
	return GenerateResponse{
		SessionID:          result.Session.SessionID,
		RequestID:          result.Session.RequestID,
		Content:            result.Session.FinalOutput.Content,
		GeneratedArtifacts: cloneArtifacts(result.Session.FinalOutput.GeneratedArtifacts),
		TokensUsed:         0,
		FinishReason:       string(result.Session.Status),
	}
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

	metadata := map[string]string{
		"http.prompt":      req.Prompt,
		"http.mode":        string(mode),
		"http.model":       req.Model,
		"http.temperature": strconv.FormatFloat(req.Temperature, 'f', -1, 64),
		"http.max_tokens":  strconv.Itoa(req.MaxTokens),
		"http.session_id":  sessionID,
		"http.resume":      strconv.FormatBool(req.Resume),
		"http.project_id":  req.ProjectID,
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
		Content:   req.Prompt,
		VisualIntent: domainagent.VisualIntent{
			Mode:             mode,
			Goal:             req.Prompt,
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
		cloned[i] = artifact
		if len(artifact.Bytes) > 0 {
			cloned[i].Bytes = append([]byte(nil), artifact.Bytes...)
		}
		if len(artifact.Metadata) > 0 {
			cloned[i].Metadata = make(map[string]string, len(artifact.Metadata))
			for key, value := range artifact.Metadata {
				cloned[i].Metadata[key] = value
			}
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
