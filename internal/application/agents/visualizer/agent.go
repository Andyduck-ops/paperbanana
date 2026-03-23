package visualizer

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/paperbanana/paperbanana/internal/application/agents/modelselection"
	pbconfig "github.com/paperbanana/paperbanana/internal/config"
	domainagent "github.com/paperbanana/paperbanana/internal/domain/agent"
	domainllm "github.com/paperbanana/paperbanana/internal/domain/llm"
)

const (
	defaultMaxOutputTokens = 50000
	PromptVersion          = "visualizer-v1"
)

type Config struct {
	Model           string
	Temperature     float64
	MaxOutputTokens int
	PlotExecutor    PlotExecutor
	NodeCatalog     *pbconfig.NodeCatalog
	NodeAdapter     nodeExecutor
}

type Agent struct {
	client     domainllm.LLMClient
	cfg        Config
	nodeRunner *nodeRunner
	state      domainagent.AgentState
}

func NewAgent(client domainllm.LLMClient, cfg Config) *Agent {
	if cfg.MaxOutputTokens <= 0 {
		cfg.MaxOutputTokens = defaultMaxOutputTokens
	}
	if cfg.PlotExecutor == nil {
		cfg.PlotExecutor = NewPlotExecutor()
	}

	return &Agent{
		client:     client,
		cfg:        cfg,
		nodeRunner: newNodeRunner(cfg.NodeCatalog, cfg.NodeAdapter),
		state: domainagent.AgentState{
			Stage:  domainagent.StageVisualizer,
			Status: domainagent.StatusPending,
		},
	}
}

func (a *Agent) Initialize(context.Context) error {
	a.state.Stage = domainagent.StageVisualizer
	a.state.Status = domainagent.StatusRunning
	a.state.Error = nil
	return nil
}

func (a *Agent) Execute(ctx context.Context, input domainagent.AgentInput) (domainagent.AgentOutput, error) {
	prompt, err := a.promptMetadata(input.VisualIntent.Mode)
	if err != nil {
		return domainagent.AgentOutput{}, err
	}

	a.state.Stage = domainagent.StageVisualizer
	a.state.Status = domainagent.StatusRunning
	a.state.Input = input
	a.state.Error = nil
	a.state.Output = a.baseOutput(input, prompt)

	var output domainagent.AgentOutput
	switch {
	case shouldReuseArtifact(input):
		output = a.reuseOutput(input, prompt)
	case a.nodeRunner.enabled(input):
		output, err = a.nodeRunner.execute(ctx, input, prompt)
	case input.VisualIntent.Mode == domainagent.VisualModeDiagram:
		output, err = a.executeDiagram(ctx, input, prompt)
	case input.VisualIntent.Mode == domainagent.VisualModePlot:
		output, err = a.executePlot(ctx, input, prompt)
	default:
		err = fmt.Errorf("unsupported visual mode %q", input.VisualIntent.Mode)
	}
	if err != nil {
		a.state.Status = domainagent.StatusFailed
		a.state.Error = &domainagent.ErrorDetail{
			Message: err.Error(),
			Stage:   domainagent.StageVisualizer,
		}
		return domainagent.AgentOutput{}, err
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

func (a *Agent) executeDiagram(ctx context.Context, input domainagent.AgentInput, prompt domainagent.PromptMetadata) (domainagent.AgentOutput, error) {
	if a.client == nil {
		return domainagent.AgentOutput{}, errors.New("visualizer diagram mode requires an llm client")
	}

	userPrompt, err := buildUserPrompt(input)
	if err != nil {
		return domainagent.AgentOutput{}, err
	}

	resp, err := a.client.Generate(ctx, domainllm.GenerateRequest{
		SystemInstruction: prompt.SystemInstruction,
		Messages: []domainllm.Message{{
			Role:  domainllm.RoleUser,
			Parts: []domainllm.Part{domainllm.TextPart(userPrompt)},
		}},
		Model:         modelselection.GenerationModel(input.Metadata, a.cfg.Model),
		Temperature:   a.cfg.Temperature,
		MaxTokens:     a.cfg.MaxOutputTokens,
		PromptVersion: prompt.Version,
	})
	if err != nil {
		return domainagent.AgentOutput{}, err
	}

	mimeType, bytes, err := imageArtifactFromResponse(resp)
	if err != nil {
		return domainagent.AgentOutput{}, err
	}

	output := a.baseOutput(input, prompt)
	output.GeneratedArtifacts = append(output.GeneratedArtifacts, renderedArtifact(input.VisualIntent.Mode, mimeType, bytes))
	output.Metadata = map[string]string{
		"execution_path": "llm-image",
		"summary":        fmt.Sprintf("generated %s figure artifact via llm image path", input.VisualIntent.Mode),
	}
	return output, nil
}

func (a *Agent) executePlot(ctx context.Context, input domainagent.AgentInput, prompt domainagent.PromptMetadata) (domainagent.AgentOutput, error) {
	if a.client == nil {
		return domainagent.AgentOutput{}, errors.New("visualizer plot mode requires an llm client")
	}

	userPrompt, err := buildUserPrompt(input)
	if err != nil {
		return domainagent.AgentOutput{}, err
	}

	resp, err := a.client.Generate(ctx, domainllm.GenerateRequest{
		SystemInstruction: prompt.SystemInstruction,
		Messages: []domainllm.Message{{
			Role:  domainllm.RoleUser,
			Parts: []domainllm.Part{domainllm.TextPart(userPrompt)},
		}},
		Model:         modelselection.GenerationModel(input.Metadata, a.cfg.Model),
		Temperature:   a.cfg.Temperature,
		MaxTokens:     a.cfg.MaxOutputTokens,
		PromptVersion: prompt.Version,
	})
	if err != nil {
		return domainagent.AgentOutput{}, err
	}

	code := collectPlotCode(resp)
	if code == "" {
		return domainagent.AgentOutput{}, errors.New("visualizer plot mode returned empty code")
	}

	result, err := a.cfg.PlotExecutor.Execute(ctx, code)
	if err != nil {
		return domainagent.AgentOutput{}, err
	}
	if len(result.Bytes) == 0 {
		return domainagent.AgentOutput{}, errors.New("plot executor returned no rendered bytes")
	}
	if result.MIMEType == "" {
		result.MIMEType = "image/jpeg"
	}

	output := a.baseOutput(input, prompt)
	output.GeneratedArtifacts = append(output.GeneratedArtifacts,
		promptTraceArtifact(input.VisualIntent.Mode, code),
		renderedArtifact(input.VisualIntent.Mode, result.MIMEType, result.Bytes),
	)
	output.Metadata = map[string]string{
		"execution_path": "llm-plot",
		"summary":        "generated plot code and rendered figure artifact",
	}
	return output, nil
}

func (a *Agent) reuseOutput(input domainagent.AgentInput, prompt domainagent.PromptMetadata) domainagent.AgentOutput {
	output := a.baseOutput(input, prompt)
	output.Metadata = map[string]string{
		"execution_path":  "reuse",
		"reused_artifact": "true",
		"summary":         "reused prior rendered artifact after no-change critique",
	}
	return output
}

func (a *Agent) baseOutput(input domainagent.AgentInput, prompt domainagent.PromptMetadata) domainagent.AgentOutput {
	return domainagent.AgentOutput{
		Stage:               domainagent.StageVisualizer,
		Content:             input.Content,
		VisualIntent:        cloneVisualIntent(input.VisualIntent),
		RetrievedReferences: cloneReferences(input.RetrievedReferences),
		Prompt:              prompt,
		GeneratedArtifacts:  cloneArtifacts(input.GeneratedArtifacts),
		CritiqueRounds:      cloneCritiqueRounds(input.CritiqueRounds),
	}
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

func SystemPrompt(mode domainagent.VisualMode) (string, error) {
	switch mode {
	case domainagent.VisualModeDiagram:
		return diagramSystemPrompt, nil
	case domainagent.VisualModePlot:
		return plotSystemPrompt, nil
	default:
		return "", fmt.Errorf("unsupported visual mode %q", mode)
	}
}

func promptTemplate(mode domainagent.VisualMode) (string, error) {
	switch mode {
	case domainagent.VisualModeDiagram:
		return "visualizer/diagram-system", nil
	case domainagent.VisualModePlot:
		return "visualizer/plot-system", nil
	default:
		return "", fmt.Errorf("unsupported visual mode %q", mode)
	}
}

func buildUserPrompt(input domainagent.AgentInput) (string, error) {
	switch input.VisualIntent.Mode {
	case domainagent.VisualModeDiagram:
		return fmt.Sprintf(
			"Render an image based on the following detailed description: %s\n Note that do not include figure titles in the image. Diagram: ",
			input.Content,
		), nil
	case domainagent.VisualModePlot:
		return fmt.Sprintf(
			"Use python matplotlib to generate a statistical plot based on the following detailed description: %s\n Only provide the code without any explanations. Code:",
			input.Content,
		), nil
	default:
		return "", fmt.Errorf("unsupported visual mode %q", input.VisualIntent.Mode)
	}
}

func collectPlotCode(response *domainllm.GenerateResponse) string {
	if response == nil {
		return ""
	}
	if content := strings.TrimSpace(response.Content); content != "" {
		return content
	}
	return strings.TrimSpace(domainllm.CollectText(response.Parts))
}

func imageArtifactFromResponse(response *domainllm.GenerateResponse) (string, []byte, error) {
	if response == nil {
		return "", nil, errors.New("visualizer returned no response")
	}

	for _, part := range response.Parts {
		if part.Type != domainllm.PartTypeImage {
			continue
		}
		if len(part.Data) == 0 {
			continue
		}
		mimeType := part.MIMEType
		if mimeType == "" {
			mimeType = "image/png"
		}
		return mimeType, append([]byte(nil), part.Data...), nil
	}

	return "", nil, errors.New("visualizer diagram mode returned no image artifact")
}

func shouldReuseArtifact(input domainagent.AgentInput) bool {
	if !critiqueRequestsNoChange(input.CritiqueRounds) {
		return false
	}
	for _, artifact := range input.GeneratedArtifacts {
		if artifact.Kind == domainagent.ArtifactKindRenderedFigure && len(artifact.Bytes) > 0 {
			return true
		}
	}
	return false
}

func critiqueRequestsNoChange(rounds []domainagent.CritiqueRound) bool {
	if len(rounds) == 0 {
		return false
	}

	last := rounds[len(rounds)-1]
	if strings.EqualFold(strings.TrimSpace(last.Summary), "No changes needed.") {
		return true
	}
	for _, change := range last.RequestedChanges {
		if strings.EqualFold(strings.TrimSpace(change), "No changes needed.") {
			return true
		}
	}
	return false
}

func renderedArtifact(mode domainagent.VisualMode, mimeType string, bytes []byte) domainagent.Artifact {
	return domainagent.Artifact{
		ID:       fmt.Sprintf("visualizer-%s-rendered", mode),
		Kind:     domainagent.ArtifactKindRenderedFigure,
		MIMEType: mimeType,
		URI:      fmt.Sprintf("memory://visualizer/%s/rendered", mode),
		Bytes:    append([]byte(nil), bytes...),
		Metadata: map[string]string{
			"mode": string(mode),
		},
	}
}

func promptTraceArtifact(mode domainagent.VisualMode, code string) domainagent.Artifact {
	return domainagent.Artifact{
		ID:       fmt.Sprintf("visualizer-%s-code", mode),
		Kind:     domainagent.ArtifactKindPromptTrace,
		MIMEType: "text/x-python",
		URI:      fmt.Sprintf("memory://visualizer/%s/code", mode),
		Content:  code,
		Metadata: map[string]string{
			"mode": string(mode),
		},
	}
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
		cloned[i].Metadata = cloneStringMap(artifact.Metadata)
	}
	return cloned
}

func cloneCritiqueRounds(rounds []domainagent.CritiqueRound) []domainagent.CritiqueRound {
	if len(rounds) == 0 {
		return nil
	}

	cloned := make([]domainagent.CritiqueRound, len(rounds))
	for i, round := range rounds {
		cloned[i] = round
		cloned[i].RequestedChanges = append([]string(nil), round.RequestedChanges...)
	}
	return cloned
}

func cloneStringMap(values map[string]string) map[string]string {
	if len(values) == 0 {
		return nil
	}

	cloned := make(map[string]string, len(values))
	for key, value := range values {
		cloned[key] = value
	}
	return cloned
}

const diagramSystemPrompt = "You are an expert scientific diagram illustrator. Generate high-quality scientific diagrams based on user requests.\n"

const plotSystemPrompt = "You are an expert statistical plot illustrator. Write code to generate high-quality statistical plots based on user requests.\n"
