package planner

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"mime"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/paperbanana/paperbanana/internal/application/agents/modelselection"
	domainagent "github.com/paperbanana/paperbanana/internal/domain/agent"
	domainllm "github.com/paperbanana/paperbanana/internal/domain/llm"
)

const (
	defaultMaxOutputTokens = 50000
	defaultExamplesRoot    = "data/PaperBananaBench"
)

type imageLoader func(mode domainagent.VisualMode, path string) ([]byte, string, error)

type Config struct {
	Model            string
	Temperature      float64
	MaxOutputTokens  int
	ExamplesRoot     string
	LoadExampleImage imageLoader
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
	if cfg.ExamplesRoot == "" {
		cfg.ExamplesRoot = defaultExamplesRoot
	}

	agent := &Agent{
		client: client,
		cfg:    cfg,
		state: domainagent.AgentState{
			Stage:  domainagent.StagePlanner,
			Status: domainagent.StatusPending,
		},
	}
	if cfg.LoadExampleImage == nil {
		agent.cfg.LoadExampleImage = agent.loadExampleImage
	}
	return agent
}

func (a *Agent) Initialize(context.Context) error {
	a.state.Stage = domainagent.StagePlanner
	a.state.Status = domainagent.StatusRunning
	a.state.Error = nil
	return nil
}

func (a *Agent) Execute(ctx context.Context, input domainagent.AgentInput) (domainagent.AgentOutput, error) {
	if a.client == nil {
		return domainagent.AgentOutput{}, errors.New("planner requires an llm client")
	}

	prompt, err := a.promptMetadata(input.VisualIntent.Mode)
	if err != nil {
		return domainagent.AgentOutput{}, err
	}

	a.state.Stage = domainagent.StagePlanner
	a.state.Status = domainagent.StatusRunning
	a.state.Input = input
	a.state.Error = nil
	a.state.Output = domainagent.AgentOutput{
		Stage:               domainagent.StagePlanner,
		VisualIntent:        input.VisualIntent,
		RetrievedReferences: cloneReferences(input.RetrievedReferences),
		Prompt:              prompt,
	}

	examples, err := decodeReferenceBundle(input.GeneratedArtifacts)
	if err != nil {
		a.state.Status = domainagent.StatusFailed
		a.state.Error = &domainagent.ErrorDetail{
			Message: err.Error(),
			Stage:   domainagent.StagePlanner,
		}
		return domainagent.AgentOutput{}, err
	}

	messages, err := buildMessages(input, examples, a.cfg.LoadExampleImage)
	if err != nil {
		a.state.Status = domainagent.StatusFailed
		a.state.Error = &domainagent.ErrorDetail{
			Message: err.Error(),
			Stage:   domainagent.StagePlanner,
		}
		return domainagent.AgentOutput{}, err
	}

	resp, err := a.client.Generate(ctx, domainllm.GenerateRequest{
		SystemInstruction: prompt.SystemInstruction,
		Messages:          messages,
		Model:             modelselection.QueryModel(input.Metadata, a.cfg.Model),
		Temperature:       a.cfg.Temperature,
		MaxTokens:         a.cfg.MaxOutputTokens,
		PromptVersion:     prompt.Version,
	})
	if err != nil {
		a.state.Status = domainagent.StatusFailed
		a.state.Error = &domainagent.ErrorDetail{
			Message: err.Error(),
			Stage:   domainagent.StagePlanner,
		}
		return domainagent.AgentOutput{}, err
	}

	content := collectPlanningContent(resp)
	if content == "" {
		err = errors.New("planner returned empty content")
		a.state.Status = domainagent.StatusFailed
		a.state.Error = &domainagent.ErrorDetail{
			Message: err.Error(),
			Stage:   domainagent.StagePlanner,
		}
		return domainagent.AgentOutput{}, err
	}

	output := domainagent.AgentOutput{
		Stage:               domainagent.StagePlanner,
		Content:             content,
		VisualIntent:        cloneVisualIntent(input.VisualIntent),
		RetrievedReferences: cloneReferences(input.RetrievedReferences),
		Prompt:              prompt,
		GeneratedArtifacts: append(
			append([]domainagent.Artifact(nil), cloneArtifacts(input.GeneratedArtifacts)...),
			domainagent.Artifact{
				ID:       fmt.Sprintf("planner-%s-plan", input.VisualIntent.Mode),
				Kind:     domainagent.ArtifactKindPlan,
				MIMEType: "text/plain",
				URI:      fmt.Sprintf("memory://planner/%s/plan", input.VisualIntent.Mode),
				Content:  content,
				Metadata: map[string]string{"mode": string(input.VisualIntent.Mode)},
			},
		),
		Metadata: map[string]string{
			"mode":            string(input.VisualIntent.Mode),
			"summary":         summarize(content),
			"example_count":   strconv.Itoa(len(examples)),
			"retrieved_count": strconv.Itoa(len(input.RetrievedReferences)),
		},
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

func (a *Agent) promptMetadata(mode domainagent.VisualMode) (domainagent.PromptMetadata, error) {
	systemPrompt, err := SystemPrompt(mode)
	if err != nil {
		return domainagent.PromptMetadata{}, err
	}
	template, err := promptTemplate(mode)
	if err != nil {
		return domainagent.PromptMetadata{}, err
	}

	return domainagent.PromptMetadata{
		SystemInstruction: systemPrompt,
		Version:           PromptVersion,
		Template:          template,
		Variables: map[string]string{
			"mode": string(mode),
		},
	}, nil
}

func (a *Agent) loadExampleImage(mode domainagent.VisualMode, path string) ([]byte, string, error) {
	resolved, err := a.resolveExamplePath(mode, path)
	if err != nil {
		return nil, "", err
	}

	data, err := os.ReadFile(resolved)
	if err != nil {
		return nil, "", err
	}

	mimeType := mime.TypeByExtension(filepath.Ext(resolved))
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	return data, mimeType, nil
}

func (a *Agent) resolveExamplePath(mode domainagent.VisualMode, path string) (string, error) {
	candidates := candidatePaths(a.cfg.ExamplesRoot, mode, path)
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("planner example image not found for %q", path)
}

func candidatePaths(root string, mode domainagent.VisualMode, path string) []string {
	if filepath.IsAbs(path) {
		return []string{path}
	}

	modeRoot := filepath.Join(root, modeDir(mode))
	return []string{
		filepath.Join(root, path),
		filepath.Join(modeRoot, path),
		path,
	}
}

func decodeReferenceBundle(artifacts []domainagent.Artifact) ([]referenceExample, error) {
	for _, artifact := range artifacts {
		if artifact.Kind != domainagent.ArtifactKindReferenceBundle || strings.TrimSpace(artifact.Content) == "" {
			continue
		}

		var examples []referenceExample
		if err := json.Unmarshal([]byte(artifact.Content), &examples); err != nil {
			return nil, fmt.Errorf("decode planner reference bundle: %w", err)
		}
		return examples, nil
	}
	return nil, nil
}

func summarize(content string) string {
	const limit = 160
	content = strings.TrimSpace(content)
	if len(content) <= limit {
		return content
	}
	return content[:limit]
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
		cloned[i] = artifact
		cloned[i].Bytes = append([]byte(nil), artifact.Bytes...)
		if len(artifact.Metadata) > 0 {
			cloned[i].Metadata = make(map[string]string, len(artifact.Metadata))
			for key, value := range artifact.Metadata {
				cloned[i].Metadata[key] = value
			}
		}
	}
	return cloned
}

func modeDir(mode domainagent.VisualMode) string {
	switch mode {
	case domainagent.VisualModeDiagram:
		return "diagram"
	case domainagent.VisualModePlot:
		return "plot"
	default:
		return string(mode)
	}
}
