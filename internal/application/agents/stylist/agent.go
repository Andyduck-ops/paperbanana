package stylist

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/paperbanana/paperbanana/internal/application/agents/modelselection"
	domainagent "github.com/paperbanana/paperbanana/internal/domain/agent"
	domainllm "github.com/paperbanana/paperbanana/internal/domain/llm"
)

const defaultMaxOutputTokens = 4000

type Config struct {
	Model           string
	Temperature     float64
	MaxOutputTokens int
}

type Agent struct {
	client domainllm.LLMClient
	cfg    Config
	state  domainagent.AgentState
}

func NewAgent(client domainllm.LLMClient, cfg Config) *Agent {
	if cfg.MaxOutputTokens <= 0 {
		cfg.MaxOutputTokens = defaultMaxOutputTokens
	}

	return &Agent{
		client: client,
		cfg:    cfg,
		state: domainagent.AgentState{
			Stage:  domainagent.StageStylist,
			Status: domainagent.StatusPending,
		},
	}
}

func (a *Agent) Initialize(context.Context) error {
	a.state.Stage = domainagent.StageStylist
	a.state.Status = domainagent.StatusRunning
	a.state.Error = nil
	return nil
}

func (a *Agent) Execute(ctx context.Context, input domainagent.AgentInput) (domainagent.AgentOutput, error) {
	if a.client == nil {
		return domainagent.AgentOutput{}, errors.New("stylist requires an llm client")
	}

	a.state.Stage = domainagent.StageStylist
	a.state.Status = domainagent.StatusRunning
	a.state.Input = input
	a.state.Error = nil

	// Build the prompt with style guide
	prompt := buildPrompt(input.VisualIntent.Mode)

	// Build the message content
	messageContent := buildMessageContent(input)

	a.state.Output = domainagent.AgentOutput{
		Stage:               domainagent.StageStylist,
		VisualIntent:        cloneVisualIntent(input.VisualIntent),
		RetrievedReferences: cloneReferences(input.RetrievedReferences),
		Prompt:              prompt,
	}

	// Call LLM to enhance the plan
	resp, err := a.client.Generate(ctx, domainllm.GenerateRequest{
		SystemInstruction: prompt.SystemInstruction,
		Messages: []domainllm.Message{
			{
				Role:  domainllm.RoleUser,
				Parts: []domainllm.Part{domainllm.TextPart(messageContent)},
			},
		},
		Model:         modelselection.QueryModel(input.Metadata, a.cfg.Model),
		Temperature:   a.cfg.Temperature,
		MaxTokens:     a.cfg.MaxOutputTokens,
		PromptVersion: prompt.Version,
	})
	if err != nil {
		a.state.Status = domainagent.StatusFailed
		a.state.Error = &domainagent.ErrorDetail{
			Message: err.Error(),
			Stage:   domainagent.StageStylist,
		}
		return domainagent.AgentOutput{}, err
	}

	content := strings.TrimSpace(resp.Content)
	if content == "" {
		content = strings.TrimSpace(domainllm.CollectText(resp.Parts))
	}
	if content == "" {
		err = errors.New("stylist returned empty content")
		a.state.Status = domainagent.StatusFailed
		a.state.Error = &domainagent.ErrorDetail{
			Message: err.Error(),
			Stage:   domainagent.StageStylist,
		}
		return domainagent.AgentOutput{}, err
	}

	output := domainagent.AgentOutput{
		Stage:               domainagent.StageStylist,
		Content:             content,
		VisualIntent:        cloneVisualIntent(input.VisualIntent),
		RetrievedReferences: cloneReferences(input.RetrievedReferences),
		Prompt:              prompt,
		GeneratedArtifacts: append(
			append([]domainagent.Artifact(nil), cloneArtifacts(input.GeneratedArtifacts)...),
			domainagent.Artifact{
				ID:       fmt.Sprintf("stylist-%s-plan", input.VisualIntent.Mode),
				Kind:     domainagent.ArtifactKindPlan,
				MIMEType: "text/plain",
				URI:      fmt.Sprintf("memory://stylist/%s/plan", input.VisualIntent.Mode),
				Content:  content,
				Metadata: map[string]string{"mode": string(input.VisualIntent.Mode)},
			},
		),
		Metadata: cloneStringMap(input.Metadata),
	}

	a.state.Status = domainagent.StatusCompleted
	a.state.Output = output
	return output, nil
}

func (a *Agent) Cleanup(context.Context) error {
	return nil
}

func (a *Agent) GetState() domainagent.AgentState {
	return a.state
}

func (a *Agent) RestoreState(state domainagent.AgentState) error {
	a.state = state
	return nil
}

func cloneVisualIntent(intent domainagent.VisualIntent) domainagent.VisualIntent {
	cloned := intent
	cloned.Constraints = append([]string(nil), intent.Constraints...)
	cloned.PreferredOutputs = append([]string(nil), intent.PreferredOutputs...)
	return cloned
}

func cloneReferences(references []domainagent.RetrievedReference) []domainagent.RetrievedReference {
	if len(references) == 0 {
		return nil
	}

	cloned := make([]domainagent.RetrievedReference, len(references))
	for i, reference := range references {
		cloned[i] = reference
		cloned[i].Snippets = append([]string(nil), reference.Snippets...)
	}
	return cloned
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
		} else {
			cloned[i].Bytes = append([]byte(nil), artifact.Bytes...)
		}
	}
	return cloned
}

func cloneStringMap(m map[string]string) map[string]string {
	if len(m) == 0 {
		return nil
	}
	cloned := make(map[string]string, len(m))
	for k, v := range m {
		cloned[k] = v
	}
	return cloned
}
