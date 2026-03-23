package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/paperbanana/paperbanana/internal/application/orchestrator"
	domainagent "github.com/paperbanana/paperbanana/internal/domain/agent"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type mockRunner struct {
	startFn  func(context.Context, domainagent.AgentInput) (RunHandle, error)
	resumeFn func(context.Context, domainagent.AgentInput) (RunHandle, error)
}

func (m *mockRunner) Start(ctx context.Context, input domainagent.AgentInput) (RunHandle, error) {
	return m.startFn(ctx, input)
}

func (m *mockRunner) Resume(ctx context.Context, input domainagent.AgentInput) (RunHandle, error) {
	return m.resumeFn(ctx, input)
}

type mockRunHandle struct {
	events chan domainagent.Event
	result orchestrator.RunResult
	err    error
}

func (h *mockRunHandle) Events() <-chan domainagent.Event {
	return h.events
}

func (h *mockRunHandle) Wait() (orchestrator.RunResult, error) {
	return h.result, h.err
}

func TestGenerateHandlerRunsPipeline(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()
	var captured domainagent.AgentInput
	handler := NewHandler(&mockRunner{
		startFn: func(_ context.Context, input domainagent.AgentInput) (RunHandle, error) {
			captured = input
			return &mockRunHandle{
				events: closedEvents(),
				result: orchestrator.RunResult{
					Session: domainagent.SessionState{
						SessionID:   input.SessionID,
						RequestID:   input.RequestID,
						Status:      domainagent.StatusCompleted,
						FinalOutput: domainagent.AgentOutput{Content: "pipeline response"},
					},
				},
			}, nil
		},
		resumeFn: func(context.Context, domainagent.AgentInput) (RunHandle, error) {
			return nil, errors.New("unexpected resume")
		},
	}, logger)

	router := gin.New()
	router.POST("/generate", handler.Generate)

	body, err := json.Marshal(GenerateRequest{Prompt: "test prompt", Temperature: 0.7})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/generate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "test prompt", captured.Content)
	assert.Equal(t, domainagent.VisualModeDiagram, captured.VisualIntent.Mode)
	assert.Equal(t, "test prompt", captured.VisualIntent.Goal)

	var resp GenerateResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "pipeline response", resp.Content)
	assert.NotEmpty(t, resp.SessionID)
	assert.NotEmpty(t, resp.RequestID)
	assert.Equal(t, string(domainagent.StatusCompleted), resp.FinishReason)
}

func TestGenerateHandlerReturnsFinalArtifacts(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()
	artifacts := []domainagent.Artifact{
		{
			ID:       "figure-1",
			Kind:     domainagent.ArtifactKindRenderedFigure,
			MIMEType: "image/png",
			URI:      "memory://figure-1",
		},
	}

	handler := NewHandler(&mockRunner{
		startFn: func(_ context.Context, input domainagent.AgentInput) (RunHandle, error) {
			return &mockRunHandle{
				events: closedEvents(),
				result: orchestrator.RunResult{
					Session: domainagent.SessionState{
						SessionID: input.SessionID,
						RequestID: input.RequestID,
						Status:    domainagent.StatusCompleted,
						FinalOutput: domainagent.AgentOutput{
							Content:            "pipeline response",
							GeneratedArtifacts: artifacts,
						},
					},
				},
			}, nil
		},
		resumeFn: func(context.Context, domainagent.AgentInput) (RunHandle, error) {
			return nil, errors.New("unexpected resume")
		},
	}, logger)

	router := gin.New()
	router.POST("/generate", handler.Generate)

	body, err := json.Marshal(GenerateRequest{Prompt: "test prompt"})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/generate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var resp GenerateResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Len(t, resp.GeneratedArtifacts, 1)
	assert.Equal(t, domainagent.ArtifactKindRenderedFigure, resp.GeneratedArtifacts[0].Kind)
	assert.Equal(t, "memory://figure-1", resp.GeneratedArtifacts[0].URI)
}

func TestStreamGenerateHandlerEmitsOrderedStageEvents(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()
	events := make(chan domainagent.Event, 5)
	events <- domainagent.Event{Type: domainagent.EventStageStarted, Stage: domainagent.StageRetriever, Status: domainagent.StatusRunning, OccurredAt: time.Now().UTC()}
	events <- domainagent.Event{Type: domainagent.EventStageCompleted, Stage: domainagent.StageRetriever, Status: domainagent.StatusCompleted, OccurredAt: time.Now().UTC()}
	events <- domainagent.Event{Type: domainagent.EventStageStarted, Stage: domainagent.StagePlanner, Status: domainagent.StatusRunning, OccurredAt: time.Now().UTC()}
	events <- domainagent.Event{Type: domainagent.EventStageCompleted, Stage: domainagent.StagePlanner, Status: domainagent.StatusCompleted, OccurredAt: time.Now().UTC()}
	close(events)

	handler := NewHandler(&mockRunner{
		startFn: func(_ context.Context, input domainagent.AgentInput) (RunHandle, error) {
			return &mockRunHandle{
				events: events,
				result: orchestrator.RunResult{
					Session: domainagent.SessionState{
						SessionID:   input.SessionID,
						RequestID:   input.RequestID,
						Status:      domainagent.StatusCompleted,
						FinalOutput: domainagent.AgentOutput{Content: "stream response"},
					},
				},
			}, nil
		},
		resumeFn: func(context.Context, domainagent.AgentInput) (RunHandle, error) {
			return nil, errors.New("unexpected resume")
		},
	}, logger)

	router := gin.New()
	router.POST("/generate/stream", handler.StreamGenerate)

	payload, err := json.Marshal(GenerateRequest{Prompt: "stream"})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/generate/stream", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "text/event-stream", rec.Header().Get("Content-Type"))
	streamBody := rec.Body.String()
	assert.Contains(t, streamBody, "event:stage_started")
	assert.Contains(t, streamBody, "\"stage\":\"retriever\"")
	assert.Contains(t, streamBody, "\"stage\":\"planner\"")
	assert.Contains(t, streamBody, "event:result")
	assert.Less(t, strings.Index(streamBody, "\"stage\":\"retriever\""), strings.Index(streamBody, "\"stage\":\"planner\""))
}

func TestStreamGenerateHandlerResultCarriesArtifacts(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()
	artifacts := []domainagent.Artifact{
		{
			ID:       "figure-1",
			Kind:     domainagent.ArtifactKindRenderedFigure,
			MIMEType: "image/png",
			URI:      "memory://figure-1",
		},
	}

	handler := NewHandler(&mockRunner{
		startFn: func(_ context.Context, input domainagent.AgentInput) (RunHandle, error) {
			return &mockRunHandle{
				events: closedEvents(),
				result: orchestrator.RunResult{
					Session: domainagent.SessionState{
						SessionID: input.SessionID,
						RequestID: input.RequestID,
						Status:    domainagent.StatusCompleted,
						FinalOutput: domainagent.AgentOutput{
							Content:            "stream response",
							GeneratedArtifacts: artifacts,
						},
					},
				},
			}, nil
		},
		resumeFn: func(context.Context, domainagent.AgentInput) (RunHandle, error) {
			return nil, errors.New("unexpected resume")
		},
	}, logger)

	router := gin.New()
	router.POST("/generate/stream", handler.StreamGenerate)

	body, err := json.Marshal(GenerateRequest{Prompt: "stream"})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/generate/stream", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "event:result")
	assert.Contains(t, rec.Body.String(), "\"generated_artifacts\":[")
	assert.Contains(t, rec.Body.String(), "\"kind\":\"rendered_figure\"")
	assert.Contains(t, rec.Body.String(), "\"uri\":\"memory://figure-1\"")
}

func TestBuildAgentInputAllowsResumeWithoutPrompt(t *testing.T) {
	input, err := buildAgentInput(GenerateRequest{
		SessionID: "resume-session",
		Resume:    true,
	})
	require.NoError(t, err)
	assert.Equal(t, "resume-session", input.SessionID)
	assert.Empty(t, input.Content)
	assert.Equal(t, "resume-session", input.Metadata["http.session_id"])
}

func TestBuildAgentInputPropagatesVisualizerNodeSelection(t *testing.T) {
	input, err := buildAgentInput(GenerateRequest{
		Prompt:         "plot this",
		VisualizerNode: "prod-visualizer",
	})
	require.NoError(t, err)
	assert.Equal(t, "prod-visualizer", input.Metadata["visualizer.node_name"])
}

func TestGenerateHandlerFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()
	handler := NewHandler(&mockRunner{
		startFn: func(_ context.Context, _ domainagent.AgentInput) (RunHandle, error) {
			return nil, errors.New("boom")
		},
		resumeFn: func(_ context.Context, _ domainagent.AgentInput) (RunHandle, error) {
			return nil, errors.New("boom")
		},
	}, logger)

	router := gin.New()
	router.POST("/generate", handler.Generate)

	req := httptest.NewRequest(http.MethodPost, "/generate", strings.NewReader(`{"prompt":"test"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Contains(t, rec.Body.String(), "boom")
}

func closedEvents() chan domainagent.Event {
	events := make(chan domainagent.Event)
	close(events)
	return events
}
