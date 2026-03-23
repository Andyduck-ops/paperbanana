package polish

import (
	"context"
	"sync"

	domainagent "github.com/paperbanana/paperbanana/internal/domain/agent"
	domainllm "github.com/paperbanana/paperbanana/internal/domain/llm"
	"go.uber.org/zap"
)

// Config holds the configuration for PolishAgent.
type Config struct {
	Model      string
	Resolution string // "2K" or "4K"
}

// PolishAgent enhances existing images using LLM vision capabilities.
// It accepts an image and edit instructions, then outputs a refined version
// with improvements for publication quality.
type PolishAgent struct {
	llmClient domainllm.LLMClient
	config    Config
	logger    *zap.Logger
	state     domainagent.AgentState
	mu        sync.RWMutex
}

// NewAgent creates a new PolishAgent instance.
func NewAgent(llmClient domainllm.LLMClient, config Config, logger *zap.Logger) *PolishAgent {
	return &PolishAgent{
		llmClient: llmClient,
		config:    config,
		logger:    logger,
	}
}

// Initialize sets up the agent's initial state.
func (a *PolishAgent) Initialize(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.state = domainagent.AgentState{
		Stage:  domainagent.StagePolish,
		Status: domainagent.StatusPending,
	}
	return nil
}

// Execute performs image enhancement based on the input.
// It accepts image data and refinement instructions, then returns
// enhanced content suitable for the target resolution.
func (a *PolishAgent) Execute(ctx context.Context, input domainagent.AgentInput) (domainagent.AgentOutput, error) {
	a.mu.Lock()
	a.state.Status = domainagent.StatusRunning
	a.state.Input = input
	a.mu.Unlock()

	// Build prompt with image and instructions
	prompt := buildPolishPrompt(input, a.config.Resolution)

	// Create request with image if present
	var parts []domainllm.Part
	parts = append(parts, domainllm.TextPart(prompt))

	// Append any existing image parts from input
	if len(input.Messages) > 0 {
		for _, part := range input.Messages[0].Parts {
			if part.Type == domainllm.PartTypeImage {
				parts = append(parts, part)
			}
		}
	}

	req := domainllm.GenerateRequest{
		Messages: []domainllm.Message{
			{
				Role:  domainllm.RoleUser,
				Parts: parts,
			},
		},
		Model: a.config.Model,
	}

	response, err := a.llmClient.Generate(ctx, req)
	if err != nil {
		a.mu.Lock()
		a.state.Status = domainagent.StatusFailed
		a.mu.Unlock()
		return domainagent.AgentOutput{}, err
	}

	output := domainagent.AgentOutput{
		Stage:   domainagent.StagePolish,
		Content: response.Content,
	}

	a.mu.Lock()
	a.state.Status = domainagent.StatusCompleted
	a.state.Output = output
	a.mu.Unlock()

	return output, nil
}

// Cleanup releases any resources held by the agent.
func (a *PolishAgent) Cleanup(ctx context.Context) error {
	return nil
}

// GetState returns the current agent state.
func (a *PolishAgent) GetState() domainagent.AgentState {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.state
}

// RestoreState restores the agent to a previous state.
func (a *PolishAgent) RestoreState(state domainagent.AgentState) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.state = state
	return nil
}
