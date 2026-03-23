package handlers

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/paperbanana/paperbanana/internal/api/dto"
	"github.com/paperbanana/paperbanana/internal/application/orchestrator"
	domainagent "github.com/paperbanana/paperbanana/internal/domain/agent"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestBatchHandler_StreamBatchGenerate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create mock agent factory
	factory := &mockBatchAgentFactory{}

	// Create batch runner
	batchRunner := orchestrator.NewBatchRunner(factory)

	// Create handler
	logger := zap.NewNop()
	handler := NewBatchHandler(batchRunner, factory, logger)

	// Create request
	body := `{"prompt":"test prompt","num_candidates":2}`
	req := httptest.NewRequest(http.MethodPost, "/generate/batch", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Create gin context
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Execute
	handler.StreamBatchGenerate(c)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "text/event-stream", w.Header().Get("Content-Type"))

	// Verify SSE events
	bodyStr := w.Body.String()
	assert.Contains(t, bodyStr, "event:batch_start")
	assert.Contains(t, bodyStr, "event:batch_complete")
	assert.Contains(t, bodyStr, "event:batch_result")
}

func TestBatchHandler_InvalidRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logger := zap.NewNop()
	factory := &mockBatchAgentFactory{}
	batchRunner := orchestrator.NewBatchRunner(factory)
	handler := NewBatchHandler(batchRunner, factory, logger)

	tests := []struct {
		name       string
		body       string
		expectCode int
	}{
		{
			name:       "empty prompt",
			body:       `{"prompt":"","num_candidates":2}`,
			expectCode: http.StatusBadRequest,
		},
		{
			name:       "too many candidates",
			body:       `{"prompt":"test","num_candidates":100}`,
			expectCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/generate/batch", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			c, _ := gin.CreateTestContext(w)
			c.Request = req

			handler.StreamBatchGenerate(c)

			assert.Equal(t, tt.expectCode, w.Code)
		})
	}
}

func TestBatchHandler_BuildInputs(t *testing.T) {
	req := dto.BatchGenerateRequest{
		Prompt:         "test prompt",
		Mode:           "diagram",
		Model:          "gpt-4",
		SessionID:      "test-session",
		VisualizerNode: "node1",
		NumCandidates:  3,
	}

	inputs, err := buildBatchInputs(req)
	require.NoError(t, err)
	require.Len(t, inputs, 3)

	// Verify each input has unique session ID
	sessionIDs := make(map[string]bool)
	for i, input := range inputs {
		assert.Contains(t, input.SessionID, "test-session-candidate-")
		sessionIDs[input.SessionID] = true

		// Verify metadata
		assert.Equal(t, "test prompt", input.Content)
		assert.Equal(t, domainagent.VisualModeDiagram, input.VisualIntent.Mode)
		assert.Equal(t, "node1", input.Metadata["visualizer.node_name"])
		assert.Equal(t, string(rune('0'+i)), input.Metadata["batch.candidate_id"])
	}

	// All session IDs should be unique
	assert.Len(t, sessionIDs, 3)
}

func TestBatchHandler_DefaultNumCandidates(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logger := zap.NewNop()
	factory := &mockBatchAgentFactory{}
	batchRunner := orchestrator.NewBatchRunner(factory)
	handler := NewBatchHandler(batchRunner, factory, logger)

	// Request without num_candidates should default to 1
	body := `{"prompt":"test prompt"}`
	req := httptest.NewRequest(http.MethodPost, "/generate/batch", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.StreamBatchGenerate(c)

	// Should succeed with default of 1 candidate
	assert.Equal(t, http.StatusOK, w.Code)
}

// mockBatchAgentFactory implements orchestrator.AgentFactory for testing
type mockBatchAgentFactory struct{}

func (f *mockBatchAgentFactory) CreateRetriever() domainagent.BaseAgent {
	return &mockBatchAgent{stage: domainagent.StageRetriever}
}

func (f *mockBatchAgentFactory) CreatePlanner() domainagent.BaseAgent {
	return &mockBatchAgent{stage: domainagent.StagePlanner}
}

func (f *mockBatchAgentFactory) CreateVisualizer() domainagent.BaseAgent {
	return &mockBatchAgent{stage: domainagent.StageVisualizer}
}

func (f *mockBatchAgentFactory) CreateCritic() domainagent.BaseAgent {
	return &mockBatchAgent{stage: domainagent.StageCritic}
}

type mockBatchAgent struct {
	stage domainagent.StageName
}

func (a *mockBatchAgent) Initialize(ctx context.Context) error {
	return nil
}

func (a *mockBatchAgent) Execute(ctx context.Context, input domainagent.AgentInput) (domainagent.AgentOutput, error) {
	return domainagent.AgentOutput{
		Stage:        a.stage,
		Content:      "mock output",
		VisualIntent: input.VisualIntent,
		GeneratedArtifacts: []domainagent.Artifact{
			{
				ID:       "artifact-1",
				Kind:     domainagent.ArtifactKindRenderedFigure,
				MIMEType: "image/png",
				Content:  "mock artifact",
			},
		},
	}, nil
}

func (a *mockBatchAgent) Cleanup(ctx context.Context) error {
	return nil
}

func (a *mockBatchAgent) GetState() domainagent.AgentState {
	return domainagent.AgentState{Stage: a.stage, Status: domainagent.StatusCompleted}
}

func (a *mockBatchAgent) RestoreState(state domainagent.AgentState) error {
	return nil
}

// Helper to parse SSE events
func parseSSEEvents(body string) []map[string]interface{} {
	var events []map[string]interface{}
	lines := strings.Split(body, "\n")
	var currentEvent map[string]interface{}

	for _, line := range lines {
		if strings.HasPrefix(line, "event:") {
			if currentEvent != nil {
				events = append(events, currentEvent)
			}
			currentEvent = map[string]interface{}{
				"event": strings.TrimSpace(strings.TrimPrefix(line, "event:")),
			}
		} else if strings.HasPrefix(line, "data:") {
			if currentEvent != nil {
				var data map[string]interface{}
				json.Unmarshal([]byte(strings.TrimPrefix(line, "data:")), &data)
				currentEvent["data"] = data
			}
		}
	}
	if currentEvent != nil {
		events = append(events, currentEvent)
	}

	return events
}

func TestBatchHandler_DownloadBatchZip(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		body       string
		expectCode int
		setupFunc  func(*testing.T, *mockBatchAgentFactory, *orchestrator.BatchRunner) string
		checkFunc  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:       "missing batch_id returns 400",
			body:       `{}`,
			expectCode: http.StatusBadRequest,
		},
		{
			name:       "non-existent batch returns 404",
			body:       `{"batch_id":"non-existent-id"}`,
			expectCode: http.StatusNotFound,
		},
		{
			name:       "valid batch returns ZIP with artifacts",
			body:       "", // Will be set dynamically
			expectCode: http.StatusOK,
			setupFunc: func(t *testing.T, factory *mockBatchAgentFactory, batchRunner *orchestrator.BatchRunner) string {
				// Run a batch to generate a result
				inputs := []domainagent.AgentInput{
					{
						SessionID: "test-session",
						Content:   "Generate diagram",
						VisualIntent: domainagent.VisualIntent{
							Mode:  domainagent.VisualModeDiagram,
							Goal:  "Test batch",
							Style: "academic",
						},
					},
				}

				handle, err := batchRunner.StartBatch(context.Background(), inputs)
				require.NoError(t, err)

				// Wait for completion
				batchResult, err := handle.Wait()
				require.NoError(t, err)

				return fmt.Sprintf(`{"batch_id":"%s"}`, batchResult.BatchID)
			},
			checkFunc: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, "application/zip", w.Header().Get("Content-Type"))
				assert.Contains(t, w.Header().Get("Content-Disposition"), "attachment; filename=paperviz_candidates_")
				assert.Contains(t, w.Header().Get("Content-Disposition"), ".zip")
				// Verify ZIP content is not empty
				assert.Greater(t, w.Body.Len(), 0)
			},
		},
		{
			name:       "ZIP contains candidate PNG images",
			body:       "", // Will be set dynamically
			expectCode: http.StatusOK,
			setupFunc: func(t *testing.T, factory *mockBatchAgentFactory, batchRunner *orchestrator.BatchRunner) string {
				inputs := []domainagent.AgentInput{
					{
						SessionID: "test-session",
						Content:   "Generate diagram",
						VisualIntent: domainagent.VisualIntent{
							Mode:  domainagent.VisualModeDiagram,
							Goal:  "Test batch",
							Style: "academic",
						},
					},
				}

				handle, err := batchRunner.StartBatch(context.Background(), inputs)
				require.NoError(t, err)

				batchResult, err := handle.Wait()
				require.NoError(t, err)

				return fmt.Sprintf(`{"batch_id":"%s"}`, batchResult.BatchID)
			},
			checkFunc: func(t *testing.T, w *httptest.ResponseRecorder) {
				// Parse ZIP and verify contents
				zipReader, err := zip.NewReader(bytes.NewReader(w.Body.Bytes()), int64(w.Body.Len()))
				require.NoError(t, err)

				// Find PNG files
				var pngFiles []string
				for _, f := range zipReader.File {
					if strings.HasSuffix(f.Name, ".png") {
						pngFiles = append(pngFiles, f.Name)
					}
				}
				assert.NotEmpty(t, pngFiles, "ZIP should contain PNG files")
				assert.Contains(t, pngFiles[0], "candidate_")
			},
		},
		{
			name:       "ZIP contains metadata.json with batch info",
			body:       "", // Will be set dynamically
			expectCode: http.StatusOK,
			setupFunc: func(t *testing.T, factory *mockBatchAgentFactory, batchRunner *orchestrator.BatchRunner) string {
				inputs := []domainagent.AgentInput{
					{
						SessionID: "test-session",
						Content:   "Generate diagram",
						VisualIntent: domainagent.VisualIntent{
							Mode:  domainagent.VisualModeDiagram,
							Goal:  "Test batch",
							Style: "academic",
						},
					},
				}

				handle, err := batchRunner.StartBatch(context.Background(), inputs)
				require.NoError(t, err)

				batchResult, err := handle.Wait()
				require.NoError(t, err)

				return fmt.Sprintf(`{"batch_id":"%s"}`, batchResult.BatchID)
			},
			checkFunc: func(t *testing.T, w *httptest.ResponseRecorder) {
				// Parse ZIP and verify metadata.json
				zipReader, err := zip.NewReader(bytes.NewReader(w.Body.Bytes()), int64(w.Body.Len()))
				require.NoError(t, err)

				// Find and read metadata.json
				var metaFile *zip.File
				for _, f := range zipReader.File {
					if f.Name == "metadata.json" {
						metaFile = f
						break
					}
				}
				require.NotNil(t, metaFile, "ZIP should contain metadata.json")

				rc, err := metaFile.Open()
				require.NoError(t, err)
				defer rc.Close()

				var meta map[string]interface{}
				err = json.NewDecoder(rc).Decode(&meta)
				require.NoError(t, err)

				// Verify required fields
				assert.Contains(t, meta, "batch_id")
				assert.Contains(t, meta, "generated_at")
				assert.Contains(t, meta, "successful")
				assert.Contains(t, meta, "failed")
				assert.Contains(t, meta, "total")
				assert.Contains(t, meta, "candidates")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory := &mockBatchAgentFactory{}
			batchRunner := orchestrator.NewBatchRunner(factory)
			logger := zap.NewNop()
			handler := NewBatchHandler(batchRunner, factory, logger)

			body := tt.body
			if tt.setupFunc != nil {
				body = tt.setupFunc(t, factory, batchRunner)
			}

			req := httptest.NewRequest(http.MethodPost, "/batch/download", strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			c, _ := gin.CreateTestContext(w)
			c.Request = req

			handler.DownloadBatchZip(c)

			assert.Equal(t, tt.expectCode, w.Code)
			if tt.checkFunc != nil {
				tt.checkFunc(t, w)
			}
		})
	}
}
