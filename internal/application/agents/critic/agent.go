package critic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/paperbanana/paperbanana/internal/application/agents/modelselection"
	domainagent "github.com/paperbanana/paperbanana/internal/domain/agent"
	domainllm "github.com/paperbanana/paperbanana/internal/domain/llm"
)

const (
	defaultMaxOutputTokens = 50000
	defaultMaxRounds       = 3
	noChangesNeeded        = "No changes needed."
)

type Config struct {
	Model           string
	Temperature     float64
	MaxOutputTokens int
	MaxRounds       int
	RevisionAgent   domainagent.BaseAgent
	Now             func() time.Time
}

type Agent struct {
	client domainllm.LLMClient
	cfg    Config
	state  domainagent.AgentState
}

type critiqueResponse struct {
	Suggestions        string `json:"critic_suggestions"`
	RevisedDescription string `json:"revised_description"`
}

func NewAgent(client domainllm.LLMClient, cfg Config) *Agent {
	if cfg.MaxOutputTokens <= 0 {
		cfg.MaxOutputTokens = defaultMaxOutputTokens
	}
	if cfg.MaxRounds <= 0 {
		cfg.MaxRounds = defaultMaxRounds
	}
	if cfg.Now == nil {
		cfg.Now = func() time.Time { return time.Now().UTC() }
	}

	return &Agent{
		client: client,
		cfg:    cfg,
		state:  domainagent.AgentState{Stage: domainagent.StageCritic, Status: domainagent.StatusPending},
	}
}

func (a *Agent) Initialize(context.Context) error {
	a.state.Stage = domainagent.StageCritic
	a.state.Status = domainagent.StatusRunning
	a.state.Error = nil
	return nil
}

func (a *Agent) Execute(ctx context.Context, input domainagent.AgentInput) (domainagent.AgentOutput, error) {
	if a.client == nil {
		return domainagent.AgentOutput{}, errors.New("critic requires an llm client")
	}

	prompt, err := a.promptMetadata(input.VisualIntent.Mode)
	if err != nil {
		return domainagent.AgentOutput{}, err
	}

	a.state.Stage = domainagent.StageCritic
	a.state.Status = domainagent.StatusRunning
	a.state.Input = input
	a.state.Error = nil

	currentDescription := resolveDescription(input)
	currentArtifacts := cloneArtifacts(input.GeneratedArtifacts)
	currentRounds := cloneCritiqueRounds(input.CritiqueRounds)
	stopReason := "max_rounds"
	reusedArtifact := false

	maxRounds := a.resolveMaxRounds(input.Metadata)
	for round := 0; round < maxRounds; round++ {
		latestArtifact := latestRenderedArtifact(currentArtifacts)
		messages, err := buildMessages(
			input.VisualIntent.Mode,
			currentDescription,
			resolveSourceContent(input),
			input.VisualIntent.Goal,
			latestArtifact,
		)
		if err != nil {
			return a.fail(err)
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
			return a.fail(err)
		}

		critique, err := parseCritiqueResponse(resp)
		if err != nil {
			return a.fail(err)
		}

		evaluatedAt := a.cfg.Now()
		roundState := domainagent.CritiqueRound{
			Round:            len(currentRounds),
			Summary:          critique.Suggestions,
			Accepted:         critique.noChanges(),
			RequestedChanges: critique.requestedChanges(),
			EvaluatedAt:      evaluatedAt,
		}
		currentRounds = append(currentRounds, roundState)
		currentArtifacts = append(currentArtifacts, critiqueArtifact(input.VisualIntent.Mode, roundState.Round, critique, evaluatedAt))

		if critique.noChanges() {
			reusedArtifact = hasRenderedArtifact(currentArtifacts)
			if latestArtifact.Kind == domainagent.ArtifactKindRenderedFigure && len(latestArtifact.Bytes) > 0 {
				currentArtifacts = append(currentArtifacts, latestArtifact)
			}
			stopReason = "no_change"
			break
		}

		if a.cfg.RevisionAgent == nil {
			return a.fail(errors.New("critic requires a revision agent when changes are requested"))
		}

		revisionOutput, err := a.revise(ctx, input, currentDescription, critique.RevisedDescription, currentArtifacts, currentRounds)
		if err != nil {
			return a.fail(err)
		}

		if !hasRenderedArtifact(revisionOutput.GeneratedArtifacts) {
			reusedArtifact = hasRenderedArtifact(currentArtifacts)
			stopReason = "render_failed"
			break
		}

		currentDescription = nonEmpty(strings.TrimSpace(revisionOutput.Content), critique.RevisedDescription)
		currentArtifacts = cloneArtifacts(revisionOutput.GeneratedArtifacts)
		stopReason = "revised"
	}

	output := domainagent.AgentOutput{
		Stage:               domainagent.StageCritic,
		Content:             currentDescription,
		VisualIntent:        cloneVisualIntent(input.VisualIntent),
		RetrievedReferences: cloneReferences(input.RetrievedReferences),
		Prompt:              prompt,
		GeneratedArtifacts:  currentArtifacts,
		CritiqueRounds:      currentRounds,
		Metadata: map[string]string{
			"round_count":     fmt.Sprintf("%d", len(currentRounds)),
			"reused_artifact": fmt.Sprintf("%t", reusedArtifact),
			"stop_reason":     stopReason,
			"max_rounds":      fmt.Sprintf("%d", maxRounds),
			"summary":         summarizeCriticRun(len(currentRounds), reusedArtifact, stopReason),
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

func (a *Agent) resolveMaxRounds(metadata map[string]string) int {
	if value := strings.TrimSpace(metadata["config.critic_rounds"]); value != "" {
		parsed, err := strconv.Atoi(value)
		if err == nil && parsed > 0 {
			return parsed
		}
	}
	return a.cfg.MaxRounds
}

func (a *Agent) revise(ctx context.Context, input domainagent.AgentInput, currentDescription, revisedDescription string, artifacts []domainagent.Artifact, rounds []domainagent.CritiqueRound) (domainagent.AgentOutput, error) {
	revisionInput := cloneAgentInput(input)
	revisionInput.Stage = domainagent.StageVisualizer
	revisionInput.Content = nonEmpty(strings.TrimSpace(revisedDescription), currentDescription)
	revisionInput.GeneratedArtifacts = cloneArtifacts(artifacts)
	revisionInput.CritiqueRounds = cloneCritiqueRounds(rounds)

	if err := a.cfg.RevisionAgent.Initialize(ctx); err != nil {
		return domainagent.AgentOutput{}, err
	}

	output, execErr := a.cfg.RevisionAgent.Execute(ctx, revisionInput)
	cleanupErr := a.cfg.RevisionAgent.Cleanup(ctx)
	if execErr != nil {
		if cleanupErr != nil {
			return domainagent.AgentOutput{}, errors.Join(execErr, cleanupErr)
		}
		return domainagent.AgentOutput{}, execErr
	}
	if cleanupErr != nil {
		return domainagent.AgentOutput{}, cleanupErr
	}
	return output, nil
}

func (a *Agent) fail(err error) (domainagent.AgentOutput, error) {
	a.state.Status = domainagent.StatusFailed
	a.state.Error = &domainagent.ErrorDetail{
		Message: err.Error(),
		Stage:   domainagent.StageCritic,
	}
	return domainagent.AgentOutput{}, err
}

func parseCritiqueResponse(response *domainllm.GenerateResponse) (critiqueResponse, error) {
	raw := strings.TrimSpace(collectResponseText(response))
	if raw == "" {
		return critiqueResponse{}, errors.New("critic returned empty content")
	}

	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```JSON")
	raw = strings.TrimSuffix(raw, "```")
	raw = strings.TrimSpace(raw)

	var parsed critiqueResponse
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return critiqueResponse{}, fmt.Errorf("decode critic response: %w", err)
	}

	parsed.Suggestions = strings.TrimSpace(parsed.Suggestions)
	parsed.RevisedDescription = strings.TrimSpace(parsed.RevisedDescription)
	if parsed.Suggestions == "" {
		parsed.Suggestions = noChangesNeeded
	}
	if parsed.RevisedDescription == "" {
		parsed.RevisedDescription = noChangesNeeded
	}
	return parsed, nil
}

func collectResponseText(response *domainllm.GenerateResponse) string {
	if response == nil {
		return ""
	}
	if content := strings.TrimSpace(response.Content); content != "" {
		return content
	}
	return strings.TrimSpace(domainllm.CollectText(response.Parts))
}

func (c critiqueResponse) noChanges() bool {
	return strings.EqualFold(strings.TrimSpace(c.Suggestions), noChangesNeeded) ||
		strings.EqualFold(strings.TrimSpace(c.RevisedDescription), noChangesNeeded)
}

func (c critiqueResponse) requestedChanges() []string {
	suggestion := strings.TrimSpace(c.Suggestions)
	if suggestion == "" {
		return nil
	}
	return []string{suggestion}
}

func resolveDescription(input domainagent.AgentInput) string {
	for i := len(input.GeneratedArtifacts) - 1; i >= 0; i-- {
		artifact := input.GeneratedArtifacts[i]
		if artifact.Kind == domainagent.ArtifactKindPlan && strings.TrimSpace(artifact.Content) != "" {
			return artifact.Content
		}
	}
	return strings.TrimSpace(input.Content)
}

func resolveSourceContent(input domainagent.AgentInput) string {
	for _, key := range []string{"critic.source_content", "orchestrator.initial_content"} {
		if value := strings.TrimSpace(input.Metadata[key]); value != "" {
			return value
		}
	}
	return strings.TrimSpace(input.Content)
}

func latestRenderedArtifact(artifacts []domainagent.Artifact) domainagent.Artifact {
	for i := len(artifacts) - 1; i >= 0; i-- {
		if artifacts[i].Kind == domainagent.ArtifactKindRenderedFigure && len(artifacts[i].Bytes) > 0 {
			return cloneArtifact(artifacts[i])
		}
	}
	return domainagent.Artifact{}
}

func hasRenderedArtifact(artifacts []domainagent.Artifact) bool {
	for _, artifact := range artifacts {
		if artifact.Kind == domainagent.ArtifactKindRenderedFigure && len(artifact.Bytes) > 0 {
			return true
		}
	}
	return false
}

func critiqueArtifact(mode domainagent.VisualMode, round int, critique critiqueResponse, evaluatedAt time.Time) domainagent.Artifact {
	payload, _ := json.Marshal(map[string]any{
		"round":               round,
		"critic_suggestions":  critique.Suggestions,
		"revised_description": critique.RevisedDescription,
		"evaluated_at":        evaluatedAt.Format(time.RFC3339Nano),
	})

	return domainagent.Artifact{
		ID:       fmt.Sprintf("critic-%s-round-%d", mode, round),
		Kind:     domainagent.ArtifactKindCritique,
		MIMEType: "application/json",
		URI:      fmt.Sprintf("memory://critic/%s/round/%d", mode, round),
		Content:  string(payload),
		Metadata: map[string]string{
			"mode":  string(mode),
			"round": fmt.Sprintf("%d", round),
		},
	}
}

func summarizeCriticRun(rounds int, reused bool, stopReason string) string {
	switch {
	case reused && stopReason == "no_change":
		return "critic accepted the prior rendered artifact without rerendering"
	case stopReason == "render_failed":
		return "critic kept the prior rendered artifact after rerendering failed"
	default:
		return fmt.Sprintf("critic completed %d critique round(s)", rounds)
	}
}

func nonEmpty(value, fallback string) string {
	if strings.TrimSpace(value) != "" {
		return value
	}
	return fallback
}

func cloneAgentInput(input domainagent.AgentInput) domainagent.AgentInput {
	cloned := input
	cloned.Messages = cloneMessages(input.Messages)
	cloned.VisualIntent = cloneVisualIntent(input.VisualIntent)
	cloned.RetrievedReferences = cloneReferences(input.RetrievedReferences)
	cloned.Prompt = clonePrompt(input.Prompt)
	cloned.GeneratedArtifacts = cloneArtifacts(input.GeneratedArtifacts)
	cloned.CritiqueRounds = cloneCritiqueRounds(input.CritiqueRounds)
	cloned.Metadata = cloneStringMap(input.Metadata)
	return cloned
}

func cloneMessages(messages []domainllm.Message) []domainllm.Message {
	if len(messages) == 0 {
		return nil
	}

	cloned := make([]domainllm.Message, len(messages))
	for i, message := range messages {
		cloned[i] = message
		cloned[i].Parts = append([]domainllm.Part(nil), message.Parts...)
	}
	return cloned
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

func clonePrompt(prompt domainagent.PromptMetadata) domainagent.PromptMetadata {
	cloned := prompt
	cloned.Variables = cloneStringMap(prompt.Variables)
	return cloned
}

func cloneArtifacts(artifacts []domainagent.Artifact) []domainagent.Artifact {
	if len(artifacts) == 0 {
		return nil
	}

	cloned := make([]domainagent.Artifact, len(artifacts))
	for i, artifact := range artifacts {
		cloned[i] = cloneArtifact(artifact)
	}
	return cloned
}

func cloneArtifact(artifact domainagent.Artifact) domainagent.Artifact {
	cloned := artifact
	cloned.Bytes = append([]byte(nil), artifact.Bytes...)
	cloned.Metadata = cloneStringMap(artifact.Metadata)
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

func cloneStringMap(source map[string]string) map[string]string {
	if len(source) == 0 {
		return nil
	}

	cloned := make(map[string]string, len(source))
	for key, value := range source {
		cloned[key] = value
	}
	return cloned
}
