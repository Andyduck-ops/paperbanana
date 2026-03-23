package polish

import (
	"context"
	"testing"

	domainagent "github.com/paperbanana/paperbanana/internal/domain/agent"
	domainllm "github.com/paperbanana/paperbanana/internal/domain/llm"
	"go.uber.org/zap"
)

// mockLLMClient implements domainllm.LLMClient for testing
type mockLLMClient struct {
	generateFunc func(ctx context.Context, req domainllm.GenerateRequest) (*domainllm.GenerateResponse, error)
}

func (m *mockLLMClient) Generate(ctx context.Context, req domainllm.GenerateRequest) (*domainllm.GenerateResponse, error) {
	if m.generateFunc != nil {
		return m.generateFunc(ctx, req)
	}
	return &domainllm.GenerateResponse{Content: "test response"}, nil
}

func (m *mockLLMClient) GenerateStream(ctx context.Context, req domainllm.GenerateRequest) (<-chan domainllm.StreamChunk, <-chan error) {
	chunks := make(chan domainllm.StreamChunk)
	errs := make(chan error)
	go func() {
		defer close(chunks)
		defer close(errs)
		chunks <- domainllm.StreamChunk{Content: "test", Done: true}
	}()
	return chunks, errs
}

func (m *mockLLMClient) Provider() string {
	return "mock"
}

// Test 1: PolishAgent initializes correctly
func TestPolishAgent_Initialize(t *testing.T) {
	logger := zap.NewNop()
	mockClient := &mockLLMClient{}
	config := Config{
		Model:      "gemini-2.0-flash",
		Resolution: "2K",
	}
	agent := NewAgent(mockClient, config, logger)

	err := agent.Initialize(context.Background())
	if err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	state := agent.GetState()
	if state.Stage != domainagent.StagePolish {
		t.Errorf("expected stage %s, got %s", domainagent.StagePolish, state.Stage)
	}
	if state.Status != domainagent.StatusPending {
		t.Errorf("expected status %s, got %s", domainagent.StatusPending, state.Status)
	}
}

// Test 2: Execute accepts image and instructions
func TestPolishAgent_Execute_AcceptsImageAndInstructions(t *testing.T) {
	logger := zap.NewNop()

	var receivedPrompt string
	var receivedParts []domainllm.Part
	mockClient := &mockLLMClient{
		generateFunc: func(ctx context.Context, req domainllm.GenerateRequest) (*domainllm.GenerateResponse, error) {
			if len(req.Messages) > 0 {
				receivedParts = req.Messages[0].Parts
				if len(receivedParts) > 0 {
					receivedPrompt = receivedParts[0].Text
				}
			}
			return &domainllm.GenerateResponse{Content: "enhanced code"}, nil
		},
	}

	config := Config{
		Model:      "gemini-2.0-flash",
		Resolution: "2K",
	}
	agent := NewAgent(mockClient, config, logger)
	_ = agent.Initialize(context.Background())

	// Create input with image and instructions
	imageData := []byte("fake-image-data")
	input := domainagent.AgentInput{
		SessionID: "test-session",
		RequestID: "test-request",
		Content:   "Enhance the contrast and improve readability",
		Messages: []domainllm.Message{
			{
				Role: domainllm.RoleUser,
				Parts: []domainllm.Part{
					domainllm.InlineImagePart("image/png", imageData),
				},
			},
		},
	}

	output, err := agent.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Verify the prompt contains instructions
	if receivedPrompt == "" {
		t.Error("expected prompt to be sent to LLM, got empty")
	}

	// Verify image part is preserved
	if len(receivedParts) < 2 {
		t.Errorf("expected at least 2 parts (text + image), got %d", len(receivedParts))
	}

	// Verify output
	if output.Stage != domainagent.StagePolish {
		t.Errorf("expected stage %s, got %s", domainagent.StagePolish, output.Stage)
	}
	if output.Content != "enhanced code" {
		t.Errorf("expected content 'enhanced code', got %s", output.Content)
	}
}

// Test 3: Execute returns enhanced image content
func TestPolishAgent_Execute_ReturnsEnhancedContent(t *testing.T) {
	logger := zap.NewNop()

	expectedContent := "```python\nimport matplotlib.pyplot as plt\n# Enhanced visualization code\n```"
	mockClient := &mockLLMClient{
		generateFunc: func(ctx context.Context, req domainllm.GenerateRequest) (*domainllm.GenerateResponse, error) {
			return &domainllm.GenerateResponse{Content: expectedContent}, nil
		},
	}

	config := Config{
		Model:      "gemini-2.0-flash",
		Resolution: "2K",
	}
	agent := NewAgent(mockClient, config, logger)
	_ = agent.Initialize(context.Background())

	input := domainagent.AgentInput{
		SessionID: "test-session",
		Content:   "Make it publication ready",
	}

	output, err := agent.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if output.Content != expectedContent {
		t.Errorf("expected content '%s', got '%s'", expectedContent, output.Content)
	}

	// Verify state is completed
	state := agent.GetState()
	if state.Status != domainagent.StatusCompleted {
		t.Errorf("expected status %s, got %s", domainagent.StatusCompleted, state.Status)
	}
}

// Test 4: Resolution parameter affects output
func TestPolishAgent_Execute_ResolutionAffectsPrompt(t *testing.T) {
	logger := zap.NewNop()

	var receivedPrompt string
	mockClient := &mockLLMClient{
		generateFunc: func(ctx context.Context, req domainllm.GenerateRequest) (*domainllm.GenerateResponse, error) {
			if len(req.Messages) > 0 && len(req.Messages[0].Parts) > 0 {
				receivedPrompt = req.Messages[0].Parts[0].Text
			}
			return &domainllm.GenerateResponse{Content: "enhanced"}, nil
		},
	}

	// Test 4K resolution
	config4K := Config{
		Model:      "gemini-2.0-flash",
		Resolution: "4K",
	}
	agent4K := NewAgent(mockClient, config4K, logger)
	_ = agent4K.Initialize(context.Background())

	input := domainagent.AgentInput{
		Content: "Enhance this",
	}
	_, _ = agent4K.Execute(context.Background(), input)

	if receivedPrompt == "" {
		t.Fatal("expected prompt to be generated")
	}
	// Check that 4K resolution hint is in the prompt
	if !containsSubstring(receivedPrompt, "4K") || !containsSubstring(receivedPrompt, "3840x2160") {
		t.Errorf("expected 4K resolution hint in prompt, got: %s", receivedPrompt)
	}

	// Test 2K resolution
	config2K := Config{
		Model:      "gemini-2.0-flash",
		Resolution: "2K",
	}
	agent2K := NewAgent(mockClient, config2K, logger)
	_ = agent2K.Initialize(context.Background())

	_, _ = agent2K.Execute(context.Background(), input)

	if !containsSubstring(receivedPrompt, "2K") || !containsSubstring(receivedPrompt, "2560x1440") {
		t.Errorf("expected 2K resolution hint in prompt, got: %s", receivedPrompt)
	}
}

// Test 5: Execute returns error when LLM fails
func TestPolishAgent_Execute_LLMError(t *testing.T) {
	logger := zap.NewNop()

	mockClient := &mockLLMClient{
		generateFunc: func(ctx context.Context, req domainllm.GenerateRequest) (*domainllm.GenerateResponse, error) {
			return nil, context.DeadlineExceeded
		},
	}

	config := Config{
		Model:      "gemini-2.0-flash",
		Resolution: "2K",
	}
	agent := NewAgent(mockClient, config, logger)
	_ = agent.Initialize(context.Background())

	input := domainagent.AgentInput{
		Content: "Enhance this",
	}

	_, err := agent.Execute(context.Background(), input)
	if err == nil {
		t.Error("expected error when LLM fails, got nil")
	}

	// Verify state is failed
	state := agent.GetState()
	if state.Status != domainagent.StatusFailed {
		t.Errorf("expected status %s, got %s", domainagent.StatusFailed, state.Status)
	}
}

// Test 6: RestoreState restores agent state correctly
func TestPolishAgent_RestoreState(t *testing.T) {
	logger := zap.NewNop()
	mockClient := &mockLLMClient{}

	config := Config{
		Model:      "gemini-2.0-flash",
		Resolution: "2K",
	}
	agent := NewAgent(mockClient, config, logger)

	// Create a state to restore
	state := domainagent.AgentState{
		Stage:  domainagent.StagePolish,
		Status: domainagent.StatusCompleted,
		Input: domainagent.AgentInput{
			SessionID: "restored-session",
			Content:   "restored content",
		},
	}

	err := agent.RestoreState(state)
	if err != nil {
		t.Fatalf("RestoreState() error = %v", err)
	}

	restoredState := agent.GetState()
	if restoredState.Stage != state.Stage {
		t.Errorf("expected stage %s, got %s", state.Stage, restoredState.Stage)
	}
	if restoredState.Status != state.Status {
		t.Errorf("expected status %s, got %s", state.Status, restoredState.Status)
	}
	if restoredState.Input.SessionID != state.Input.SessionID {
		t.Errorf("expected session ID %s, got %s", state.Input.SessionID, restoredState.Input.SessionID)
	}
}

// Test 7: Cleanup returns nil (no resources to clean)
func TestPolishAgent_Cleanup(t *testing.T) {
	logger := zap.NewNop()
	mockClient := &mockLLMClient{}

	config := Config{
		Model:      "gemini-2.0-flash",
		Resolution: "2K",
	}
	agent := NewAgent(mockClient, config, logger)
	_ = agent.Initialize(context.Background())

	err := agent.Cleanup(context.Background())
	if err != nil {
		t.Errorf("Cleanup() error = %v", err)
	}
}

// Helper function
func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstringHelper(s, substr))
}

func containsSubstringHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
