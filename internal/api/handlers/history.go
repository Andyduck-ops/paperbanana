package handlers

import (
	"context"
	"errors"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	domainagent "github.com/paperbanana/paperbanana/internal/domain/agent"
	domainworkspace "github.com/paperbanana/paperbanana/internal/domain/workspace"
	"go.uber.org/zap"
)

// HistoryHandler handles history and session-related HTTP requests.
type HistoryHandler struct {
	historyService HistoryService
	logger         *zap.Logger
}

// HistoryService provides the application logic for history operations.
type HistoryService interface {
	ListHistory(ctx context.Context, projectID, visualizationID string, limit int) ([]*domainworkspace.VisualizationVersion, error)
	ListRecentSessions(ctx context.Context, projectID string, limit int) ([]*domainworkspace.SessionRecord, error)
	GetVersion(ctx context.Context, projectID, versionID string) (*domainworkspace.VisualizationVersion, error)
	GetLatestSession(ctx context.Context, projectID, visualizationID string) (*domainworkspace.SessionRecord, error)
	GetSessionByID(ctx context.Context, sessionID string) (*domainworkspace.SessionRecord, error)
}

// NewHistoryHandler creates a new HistoryHandler.
func NewHistoryHandler(historyService HistoryService, logger *zap.Logger) *HistoryHandler {
	return &HistoryHandler{
		historyService: historyService,
		logger:         logger,
	}
}

// ListHistoryRequest represents the request for listing version history.
type ListHistoryRequest struct {
	ProjectID       string `form:"project_id"`
	VisualizationID string `form:"visualization_id"`
	Limit           int    `form:"limit"`
}

// VersionResponse represents a single version in the response.
type VersionResponse struct {
	ID              string                    `json:"id"`
	VisualizationID string                    `json:"visualization_id"`
	VersionNumber   int                       `json:"version_number"`
	Summary         string                    `json:"summary"`
	CreatedAt       string                    `json:"created_at"`
	Artifacts       []VersionArtifactResponse `json:"artifacts,omitempty"`
}

// VersionArtifactResponse represents an artifact attached to a version.
type VersionArtifactResponse struct {
	ID         string `json:"id"`
	Kind       string `json:"kind"`
	MIMEType   string `json:"mime_type"`
	StorageKey string `json:"storage_key"`
	ByteSize   int64  `json:"byte_size"`
}

// ListHistoryResponse represents the response for listing version history.
type ListHistoryResponse struct {
	ProjectID string            `json:"project_id"`
	Versions  []VersionResponse `json:"versions"`
}

// ListRecentSessionsRequest represents the request for listing recent sessions.
type ListRecentSessionsRequest struct {
	ProjectID string `form:"project_id"`
	Limit     int    `form:"limit"`
}

// RecentSessionResponse represents a lightweight session item for history panels.
type RecentSessionResponse struct {
	ID              string  `json:"id"`
	ProjectID       string  `json:"project_id"`
	VisualizationID *string `json:"visualization_id,omitempty"`
	Status          string  `json:"status"`
	CurrentStage    string  `json:"current_stage"`
	SchemaVersion   string  `json:"schema_version"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
	CompletedAt     *string `json:"completed_at,omitempty"`
	Prompt          string  `json:"prompt,omitempty"`
	Summary         string  `json:"summary,omitempty"`
	Mode            string   `json:"mode,omitempty"`
	BatchID         string   `json:"batch_id,omitempty"`
	CandidateSessionIDs []string `json:"candidate_session_ids,omitempty"`
}

// ListRecentSessionsResponse represents the response for recent sessions.
type ListRecentSessionsResponse struct {
	ProjectID string                  `json:"project_id"`
	Sessions  []RecentSessionResponse `json:"sessions"`
}

// ListHistory lists the version history for a visualization.
// GET /api/v1/history?project_id=xxx&visualization_id=yyy
func (h *HistoryHandler) ListHistory(c *gin.Context) {
	var req ListHistoryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Limit <= 0 {
		req.Limit = 10
	}
	if req.Limit > 100 {
		req.Limit = 100
	}

	// If project_id or visualization_id is not provided, return empty list (graceful degradation)
	if req.ProjectID == "" || req.VisualizationID == "" {
		c.JSON(http.StatusOK, ListHistoryResponse{
			ProjectID: req.ProjectID,
			Versions:  []VersionResponse{},
		})
		return
	}

	versions, err := h.historyService.ListHistory(c.Request.Context(), req.ProjectID, req.VisualizationID, req.Limit)
	if err != nil {
		h.logger.Error("failed to list history", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list history"})
		return
	}

	response := ListHistoryResponse{
		ProjectID: req.ProjectID,
		Versions:  make([]VersionResponse, len(versions)),
	}

	for i, v := range versions {
		response.Versions[i] = VersionResponse{
			ID:              v.ID,
			VisualizationID: v.VisualizationID,
			VersionNumber:   v.VersionNumber,
			Summary:         v.Summary,
			CreatedAt:       v.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
		for _, a := range v.Artifacts {
			response.Versions[i].Artifacts = append(response.Versions[i].Artifacts, VersionArtifactResponse{
				ID:         a.ID,
				Kind:       a.Kind,
				MIMEType:   a.MIMEType,
				StorageKey: a.StorageKey,
				ByteSize:   a.ByteSize,
			})
		}
	}

	c.JSON(http.StatusOK, response)
}

// GetVersionRequest represents the request for getting a specific version.
type GetVersionRequest struct {
	ProjectID string `uri:"project_id" binding:"required"`
	VersionID string `uri:"version_id" binding:"required"`
}

// GetVersionResponse represents the response for a single version.
type GetVersionResponse struct {
	ID              string                    `json:"id"`
	VisualizationID string                    `json:"visualization_id"`
	ProjectID       string                    `json:"project_id"`
	VersionNumber   int                       `json:"version_number"`
	Summary         string                    `json:"summary"`
	CreatedAt       string                    `json:"created_at"`
	Artifacts       []VersionArtifactResponse `json:"artifacts,omitempty"`
	SessionSnapshot *domainagent.SessionState `json:"session_snapshot,omitempty"`
}

// GetVersion retrieves a specific version by ID.
// GET /api/v1/history/:project_id/:version_id
func (h *HistoryHandler) GetVersion(c *gin.Context) {
	projectID := c.Param("project_id")
	versionID := c.Param("version_id")

	if projectID == "" || versionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project_id and version_id are required"})
		return
	}

	version, err := h.historyService.GetVersion(c.Request.Context(), projectID, versionID)
	if err != nil {
		h.logger.Error("failed to get version", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "version not found"})
		return
	}

	response := GetVersionResponse{
		ID:              version.ID,
		VisualizationID: version.VisualizationID,
		ProjectID:       version.ProjectID,
		VersionNumber:   version.VersionNumber,
		Summary:         version.Summary,
		CreatedAt:       version.CreatedAt.Format("2006-01-02T15:04:05Z"),
		SessionSnapshot: version.SessionSnapshot,
	}

	for _, a := range version.Artifacts {
		response.Artifacts = append(response.Artifacts, VersionArtifactResponse{
			ID:         a.ID,
			Kind:       a.Kind,
			MIMEType:   a.MIMEType,
			StorageKey: a.StorageKey,
			ByteSize:   a.ByteSize,
		})
	}

	c.JSON(http.StatusOK, response)
}

// GetLatestSessionRequest represents the request for getting the latest session.
type GetLatestSessionRequest struct {
	ProjectID       string `form:"project_id" binding:"required"`
	VisualizationID string `form:"visualization_id" binding:"required"`
}

// SessionResponse represents a session in the response.
type SessionResponse struct {
	ID              string                    `json:"id"`
	ProjectID       string                    `json:"project_id"`
	VisualizationID *string                   `json:"visualization_id,omitempty"`
	Status          string                    `json:"status"`
	CurrentStage    string                    `json:"current_stage"`
	SchemaVersion   string                    `json:"schema_version"`
	CreatedAt       string                    `json:"created_at"`
	UpdatedAt       string                    `json:"updated_at"`
	CompletedAt     *string                   `json:"completed_at,omitempty"`
	Snapshot        *domainagent.SessionState `json:"snapshot,omitempty"`
}

// GetLatestSession retrieves the latest session for a visualization.
// GET /api/v1/session/latest?project_id=xxx&visualization_id=yyy
func (h *HistoryHandler) GetLatestSession(c *gin.Context) {
	var req GetLatestSessionRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	session, err := h.historyService.GetLatestSession(c.Request.Context(), req.ProjectID, req.VisualizationID)
	if err != nil {
		h.logger.Error("failed to get latest session", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	response := SessionResponse{
		ID:            session.ID,
		ProjectID:     session.ProjectID,
		Status:        session.Status,
		CurrentStage:  session.CurrentStage,
		SchemaVersion: session.SchemaVersion,
		CreatedAt:     session.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:     session.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		Snapshot:      session.Snapshot,
	}

	if session.VisualizationID != nil {
		response.VisualizationID = session.VisualizationID
	}
	if session.CompletedAt != nil {
		completedAt := session.CompletedAt.Format("2006-01-02T15:04:05Z")
		response.CompletedAt = &completedAt
	}

	c.JSON(http.StatusOK, response)
}

// ListRecentSessions lists recent persisted sessions.
// GET /api/v1/sessions/recent?project_id=xxx&limit=20
func (h *HistoryHandler) ListRecentSessions(c *gin.Context) {
	var req ListRecentSessionsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.Limit > 100 {
		req.Limit = 100
	}

	fetchLimit := req.Limit
	if fetchLimit < 100 {
		fetchLimit = minInt(100, req.Limit*5)
	}

	sessions, err := h.historyService.ListRecentSessions(c.Request.Context(), req.ProjectID, fetchLimit)
	if err != nil {
		h.logger.Error("failed to list recent sessions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list recent sessions"})
		return
	}

	groupedSessions := groupRecentSessions(sessions)
	if req.Limit > 0 && len(groupedSessions) > req.Limit {
		groupedSessions = groupedSessions[:req.Limit]
	}

	response := ListRecentSessionsResponse{
		ProjectID: req.ProjectID,
		Sessions:  make([]RecentSessionResponse, len(groupedSessions)),
	}

	for i, session := range groupedSessions {
		response.Sessions[i] = mapRecentSessionResponse(session)
	}

	c.JSON(http.StatusOK, response)
}

// GetSessionRequest represents the request for getting a session by ID.
type GetSessionRequest struct {
	SessionID string `uri:"session_id" binding:"required"`
}

// GetSession retrieves a session by ID.
// GET /api/v1/session/:session_id
func (h *HistoryHandler) GetSession(c *gin.Context) {
	sessionID := c.Param("session_id")

	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session_id is required"})
		return
	}

	session, err := h.historyService.GetSessionByID(c.Request.Context(), sessionID)
	if err != nil {
		h.logger.Error("failed to get session", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	response := SessionResponse{
		ID:            session.ID,
		ProjectID:     session.ProjectID,
		Status:        session.Status,
		CurrentStage:  session.CurrentStage,
		SchemaVersion: session.SchemaVersion,
		CreatedAt:     session.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:     session.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		Snapshot:      session.Snapshot,
	}

	if session.VisualizationID != nil {
		response.VisualizationID = session.VisualizationID
	}
	if session.CompletedAt != nil {
		completedAt := session.CompletedAt.Format("2006-01-02T15:04:05Z")
		response.CompletedAt = &completedAt
	}

	c.JSON(http.StatusOK, response)
}

func mapRecentSessionResponse(session *domainworkspace.SessionRecord) RecentSessionResponse {
	response := RecentSessionResponse{
		ID:            session.ID,
		ProjectID:     session.ProjectID,
		Status:        session.Status,
		CurrentStage:  session.CurrentStage,
		SchemaVersion: session.SchemaVersion,
		CreatedAt:     session.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:     session.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		Prompt:        sessionPrompt(session),
		Summary:       sessionSummary(session),
		Mode:          sessionMode(session),
	}

	if session.VisualizationID != nil {
		response.VisualizationID = session.VisualizationID
	}
	if session.CompletedAt != nil {
		completedAt := session.CompletedAt.Format("2006-01-02T15:04:05Z")
		response.CompletedAt = &completedAt
	}
	if response.Mode == "batch" {
		response.BatchID = sessionBatchGroupID(session)
		response.CandidateSessionIDs = sessionBatchSessionIDs(session)
	}

	return response
}

type recentSessionAggregate struct {
	latest    *domainworkspace.SessionRecord
	candidates []*domainworkspace.SessionRecord
}

func groupRecentSessions(sessions []*domainworkspace.SessionRecord) []*domainworkspace.SessionRecord {
	if len(sessions) == 0 {
		return nil
	}

	batches := make(map[string]*recentSessionAggregate)

	for _, session := range sessions {
		groupID := sessionBatchGroupID(session)
		if groupID == "" {
			continue
		}

		aggregate, ok := batches[groupID]
		if !ok {
			aggregate = &recentSessionAggregate{latest: session}
			batches[groupID] = aggregate
		}
		if session.CreatedAt.After(aggregate.latest.CreatedAt) {
			aggregate.latest = session
		}
		aggregate.candidates = append(aggregate.candidates, session)
	}

	grouped := make([]*domainworkspace.SessionRecord, 0, len(sessions))
	seenGroups := make(map[string]struct{})

	for _, session := range sessions {
		groupID := sessionBatchGroupID(session)
		if groupID == "" {
			grouped = append(grouped, session)
			continue
		}

		if _, seen := seenGroups[groupID]; seen {
			continue
		}

		aggregate := batches[groupID]
		if aggregate == nil || aggregate.latest == nil {
			continue
		}

		grouped = append(grouped, buildBatchRecentSession(aggregate, groupID))
		seenGroups[groupID] = struct{}{}
	}

	return grouped
}

func buildBatchRecentSession(aggregate *recentSessionAggregate, groupID string) *domainworkspace.SessionRecord {
	latest := aggregate.latest
	session := *latest
	session.ID = groupID
	session.CurrentStage = "batch"
	session.Status = aggregateBatchStatus(aggregate.candidates)
	session.Snapshot = cloneBatchSnapshot(latest.Snapshot, aggregate.candidates, groupID)
	session.CreatedAt = latest.CreatedAt
	session.UpdatedAt = latest.UpdatedAt
	session.CompletedAt = aggregateBatchCompletedAt(aggregate.candidates)
	return &session
}

func cloneBatchSnapshot(
	snapshot *domainagent.SessionState,
	candidates []*domainworkspace.SessionRecord,
	groupID string,
) *domainagent.SessionState {
	if snapshot == nil {
		return &domainagent.SessionState{
			Metadata: map[string]string{
				"history.mode":    "batch",
				"batch.group_id":  groupID,
				"batch.session_ids": strings.Join(candidateSessionIDs(candidates), ","),
			},
		}
	}

	cloned := *snapshot
	cloned.Metadata = cloneHistoryStringMap(snapshot.Metadata)
	if cloned.Metadata == nil {
		cloned.Metadata = make(map[string]string)
	}
	cloned.Metadata["history.mode"] = "batch"
	cloned.Metadata["batch.group_id"] = groupID
	cloned.Metadata["batch.session_ids"] = strings.Join(candidateSessionIDs(candidates), ",")
	return &cloned
}

func cloneHistoryStringMap(source map[string]string) map[string]string {
	if len(source) == 0 {
		return nil
	}

	cloned := make(map[string]string, len(source))
	for key, value := range source {
		cloned[key] = value
	}
	return cloned
}

func aggregateBatchStatus(candidates []*domainworkspace.SessionRecord) string {
	if len(candidates) == 0 {
		return ""
	}

	hasRunning := false
	hasCompleted := false
	for _, candidate := range candidates {
		switch candidate.Status {
		case string(domainagent.StatusRunning):
			hasRunning = true
		case string(domainagent.StatusCompleted):
			hasCompleted = true
		}
	}

	if hasRunning {
		return string(domainagent.StatusRunning)
	}
	if hasCompleted {
		return string(domainagent.StatusCompleted)
	}
	return string(domainagent.StatusFailed)
}

func aggregateBatchCompletedAt(candidates []*domainworkspace.SessionRecord) *time.Time {
	var latest *time.Time
	for _, candidate := range candidates {
		if candidate.CompletedAt == nil {
			continue
		}
		if latest == nil || candidate.CompletedAt.After(*latest) {
			value := *candidate.CompletedAt
			latest = &value
		}
	}
	return latest
}

func candidateSessionIDs(candidates []*domainworkspace.SessionRecord) []string {
	type indexedSession struct {
		index int
		id    string
	}

	indexed := make([]indexedSession, 0, len(candidates))
	for _, candidate := range candidates {
		indexed = append(indexed, indexedSession{
			index: batchCandidateIndex(candidate),
			id:    candidate.ID,
		})
	}

	sort.SliceStable(indexed, func(i, j int) bool {
		if indexed[i].index == indexed[j].index {
			return indexed[i].id < indexed[j].id
		}
		return indexed[i].index < indexed[j].index
	})

	result := make([]string, 0, len(indexed))
	seen := make(map[string]struct{}, len(indexed))
	for _, item := range indexed {
		if _, ok := seen[item.id]; ok {
			continue
		}
		seen[item.id] = struct{}{}
		result = append(result, item.id)
	}
	return result
}

func batchCandidateIndex(session *domainworkspace.SessionRecord) int {
	if session == nil || session.Snapshot == nil {
		return 1 << 30
	}

	for _, value := range []string{
		session.Snapshot.InitialInput.Metadata["batch.candidate_id"],
		session.Snapshot.Metadata["batch.candidate_id"],
	} {
		if index, err := strconv.Atoi(strings.TrimSpace(value)); err == nil {
			return index
		}
	}

	if idx := strings.LastIndex(session.ID, "-candidate-"); idx >= 0 {
		if index, err := strconv.Atoi(session.ID[idx+len("-candidate-"):]); err == nil {
			return index
		}
	}

	return 1 << 30
}

func sessionBatchGroupID(session *domainworkspace.SessionRecord) string {
	if session == nil || session.Snapshot == nil {
		return ""
	}

	for _, value := range []string{
		session.Snapshot.InitialInput.Metadata["batch.group_id"],
		session.Snapshot.Metadata["batch.group_id"],
	} {
		if groupID := strings.TrimSpace(value); groupID != "" {
			return groupID
		}
	}

	return ""
}

func sessionMode(session *domainworkspace.SessionRecord) string {
	if session == nil {
		return ""
	}
	if groupID := sessionBatchGroupID(session); groupID != "" || session.CurrentStage == "batch" {
		return "batch"
	}
	if session.Snapshot != nil {
		if strings.EqualFold(strings.TrimSpace(session.Snapshot.Metadata["history.mode"]), "batch") {
			return "batch"
		}
		if session.CurrentStage == string(domainagent.StagePolish) {
			return "refine"
		}
		for _, stage := range session.Snapshot.StageStates {
			if stage.Stage == domainagent.StagePolish {
				return "refine"
			}
		}
	}
	return "generate"
}

func sessionBatchSessionIDs(session *domainworkspace.SessionRecord) []string {
	if session == nil || session.Snapshot == nil {
		return nil
	}

	raw := strings.TrimSpace(session.Snapshot.Metadata["batch.session_ids"])
	if raw == "" {
		return nil
	}

	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if value := strings.TrimSpace(part); value != "" {
			result = append(result, value)
		}
	}
	return result
}

func minInt(left, right int) int {
	if left < right {
		return left
	}
	return right
}

func sessionPrompt(session *domainworkspace.SessionRecord) string {
	if session == nil || session.Snapshot == nil {
		return ""
	}

	snapshot := session.Snapshot
	return normalizeSessionText(
		180,
		snapshot.InitialInput.VisualIntent.Goal,
		snapshot.InitialInput.Metadata["http.visual_intent"],
		snapshot.InitialInput.Content,
		snapshot.Metadata["http.visual_intent"],
		snapshot.Metadata["http.prompt"],
	)
}

func sessionSummary(session *domainworkspace.SessionRecord) string {
	if session == nil {
		return ""
	}

	prompt := sessionPrompt(session)
	if session.Snapshot == nil {
		return prompt
	}

	snapshot := session.Snapshot
	return normalizeSessionText(
		220,
		snapshot.FinalOutput.Metadata["summary"],
		latestStageSummary(snapshot),
		snapshot.FinalOutput.Content,
		sessionErrorMessage(session),
		prompt,
	)
}

func latestStageSummary(snapshot *domainagent.SessionState) string {
	if snapshot == nil {
		return ""
	}

	for i := len(snapshot.StageStates) - 1; i >= 0; i-- {
		stage := snapshot.StageStates[i]
		if summary := normalizeSessionText(220, stage.Output.Metadata["summary"], stage.Output.Content); summary != "" {
			return summary
		}
		if stage.Error != nil && stage.Error.Message != "" {
			return stage.Error.Message
		}
	}

	return ""
}

func sessionErrorMessage(session *domainworkspace.SessionRecord) string {
	if session == nil || session.Snapshot == nil {
		return ""
	}
	if session.Snapshot.Error != nil {
		return session.Snapshot.Error.Message
	}
	return ""
}

func normalizeSessionText(limit int, values ...string) string {
	for _, value := range values {
		normalized := strings.Join(strings.Fields(value), " ")
		if normalized == "" {
			continue
		}
		if len(normalized) <= limit {
			return normalized
		}
		if limit <= 3 {
			return normalized[:limit]
		}
		return normalized[:limit-3] + "..."
	}

	return ""
}

// isNotFoundError checks if an error indicates a not found condition.
func isNotFoundError(err error) bool {
	return errors.Is(err, errors.New("not found")) ||
		errors.Is(err, errors.New("no resumable session found")) ||
		err.Error() == "not found" ||
		err.Error() == "no resumable session found"
}
