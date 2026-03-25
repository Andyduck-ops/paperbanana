package handlers

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/paperbanana/paperbanana/internal/api/dto"
	domainllm "github.com/paperbanana/paperbanana/internal/domain/llm"
	domainworkspace "github.com/paperbanana/paperbanana/internal/domain/workspace"
)

type fakeRefineGenerator struct {
	requests           []domainllm.GenerateRequest
	generateCalls      int
	generateImageCalls int
	generateImageFunc  func(ctx context.Context, req domainllm.GenerateRequest) (*domainllm.GenerateResponse, error)
}

func (f *fakeRefineGenerator) Generate(_ context.Context, _ domainllm.GenerateRequest) (*domainllm.GenerateResponse, error) {
	f.generateCalls++
	return &domainllm.GenerateResponse{Content: "text-response"}, nil
}

func (f *fakeRefineGenerator) GenerateImage(ctx context.Context, req domainllm.GenerateRequest) (*domainllm.GenerateResponse, error) {
	f.generateImageCalls++
	f.requests = append(f.requests, req)
	if f.generateImageFunc != nil {
		return f.generateImageFunc(ctx, req)
	}
	return &domainllm.GenerateResponse{
		Content: "refined image",
		Parts: []domainllm.Part{
			domainllm.InlineImagePart("image/png", []byte("refined-image")),
		},
	}, nil
}

func (f *fakeRefineGenerator) GenerateStream(context.Context, domainllm.GenerateRequest) (<-chan domainllm.StreamChunk, <-chan error) {
	chunks := make(chan domainllm.StreamChunk)
	errs := make(chan error)
	close(chunks)
	close(errs)
	return chunks, errs
}

func (f *fakeRefineGenerator) Provider() string {
	return "fake-refine"
}

type fakeRefineSessionSaver struct {
	sessions []*domainworkspace.SessionRecord
}

func (f *fakeRefineSessionSaver) SaveSession(_ context.Context, session *domainworkspace.SessionRecord) error {
	f.sessions = append(f.sessions, session)
	return nil
}

func setupRefineHandlerTest(t *testing.T, generator domainllm.LLMClient, saver RefineSessionSaver) *gin.Engine {
	t.Helper()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/refine", NewRefineHandler(generator, saver, zap.NewNop()).Refine)
	return router
}

func TestRefineHandler_ReturnsRefinedImageArtifactAndPersistsSession(t *testing.T) {
	generator := &fakeRefineGenerator{}
	saver := &fakeRefineSessionSaver{}
	router := setupRefineHandlerTest(t, generator, saver)

	sourceImage := []byte{0x89, 0x50, 0x4E, 0x47, 0x01}
	reqBody := dto.RefineRequest{
		ImageData:    "data:image/png;base64," + base64.StdEncoding.EncodeToString(sourceImage),
		Instructions: "Sharpen labels and improve contrast",
		Resolution:   "2K",
	}

	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/refine", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var resp dto.RefineResponse
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)

	require.NotEmpty(t, resp.SessionID)
	require.NotNil(t, resp.Image)
	assert.Equal(t, "completed", resp.Status)
	assert.Equal(t, "image/png", resp.Image.MIMEType)
	assert.Equal(t, base64.StdEncoding.EncodeToString([]byte("refined-image")), resp.Image.Data)
	assert.Equal(t, "1", resp.Metadata.Iterations)
	assert.Equal(t, "accepted", resp.Metadata.StopReason)
	assert.Equal(t, 0, generator.generateCalls)
	assert.Equal(t, 1, generator.generateImageCalls)

	require.Len(t, generator.requests, 1)
	require.Len(t, generator.requests[0].Messages, 1)
	require.Len(t, generator.requests[0].Messages[0].Parts, 2)
	assert.Equal(t, domainllm.PartTypeImage, generator.requests[0].Messages[0].Parts[1].Type)
	assert.Equal(t, sourceImage, generator.requests[0].Messages[0].Parts[1].Data)

	require.Len(t, saver.sessions, 1)
	assert.Equal(t, resp.SessionID, saver.sessions[0].ID)
	assert.Equal(t, "polish", saver.sessions[0].CurrentStage)
	require.NotNil(t, saver.sessions[0].Snapshot)
	require.Len(t, saver.sessions[0].Snapshot.FinalOutput.GeneratedArtifacts, 1)
	assert.Equal(t, "polished_image", string(saver.sessions[0].Snapshot.FinalOutput.GeneratedArtifacts[0].Kind))
}

func TestRefineHandler_AcceptsPlainBase64ImageData(t *testing.T) {
	generator := &fakeRefineGenerator{}
	router := setupRefineHandlerTest(t, generator, nil)

	sourceImage := []byte{0xFF, 0xD8, 0xFF, 0x00}
	reqBody := dto.RefineRequest{
		ImageData:    base64.StdEncoding.EncodeToString(sourceImage),
		Instructions: "Clean up the image",
		Resolution:   "4K",
	}

	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/refine", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Len(t, generator.requests, 1)
	assert.Equal(t, sourceImage, generator.requests[0].Messages[0].Parts[1].Data)
}

func TestRefineHandler_HonorsIterationPlan(t *testing.T) {
	iterationImages := [][]byte{
		[]byte("refined-round-1"),
		[]byte("refined-round-2"),
		[]byte("refined-round-3"),
	}
	generator := &fakeRefineGenerator{}
	generator.generateImageFunc = func(_ context.Context, req domainllm.GenerateRequest) (*domainllm.GenerateResponse, error) {
		index := len(generator.requests) - 1
		return &domainllm.GenerateResponse{
			Content: "refined image",
			Parts: []domainllm.Part{
				domainllm.InlineImagePart("image/png", iterationImages[index]),
			},
		}, nil
	}
	router := setupRefineHandlerTest(t, generator, nil)

	sourceImage := []byte{0x89, 0x50, 0x4E, 0x47}
	reqBody := dto.RefineRequest{
		ImageData:       "data:image/png;base64," + base64.StdEncoding.EncodeToString(sourceImage),
		Instructions:    "Iteratively improve spacing and readability",
		Resolution:      "2K",
		EnableIteration: true,
		MaxIterations:   3,
	}

	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/refine", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var resp dto.RefineResponse
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, 3, generator.generateImageCalls)
	assert.Equal(t, "3", resp.Metadata.Iterations)
	assert.Equal(t, "max_iterations", resp.Metadata.StopReason)
	require.NotNil(t, resp.Image)
	assert.Equal(t, base64.StdEncoding.EncodeToString(iterationImages[2]), resp.Image.Data)

	require.Len(t, generator.requests, 3)
	assert.Equal(t, sourceImage, generator.requests[0].Messages[0].Parts[1].Data)
	assert.Equal(t, iterationImages[0], generator.requests[1].Messages[0].Parts[1].Data)
	assert.Equal(t, iterationImages[1], generator.requests[2].Messages[0].Parts[1].Data)
}
