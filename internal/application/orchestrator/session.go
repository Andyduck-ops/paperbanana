package orchestrator

import (
	"sync"
	"time"

	domainagent "github.com/paperbanana/paperbanana/internal/domain/agent"
	domainllm "github.com/paperbanana/paperbanana/internal/domain/llm"
)

const sessionSchemaVersion = "agent-session/v1"

type RunResult struct {
	Session     domainagent.SessionState
	FailedStage domainagent.StageName
}

type RunHandle struct {
	events <-chan domainagent.Event
	done   chan struct{}

	once   sync.Once
	result RunResult
	err    error
}

func newRunHandle(events <-chan domainagent.Event) *RunHandle {
	return &RunHandle{
		events: events,
		done:   make(chan struct{}),
	}
}

func (h *RunHandle) Events() <-chan domainagent.Event {
	return h.events
}

func (h *RunHandle) Wait() (RunResult, error) {
	<-h.done
	return h.result, h.err
}

func (h *RunHandle) setOutcome(result RunResult, err error) {
	h.once.Do(func() {
		h.result = result
		h.err = err
		close(h.done)
	})
}

type sessionTracker struct {
	state        domainagent.SessionState
	currentInput domainagent.AgentInput
}

func newSessionTracker(input domainagent.AgentInput, pipeline []domainagent.StageName) *sessionTracker {
	now := time.Now().UTC()
	clonedInput := cloneAgentInput(input)
	clonedPipeline := append([]domainagent.StageName(nil), pipeline...)
	if len(clonedPipeline) == 0 {
		clonedPipeline = domainagent.CanonicalPipeline()
	}

	return &sessionTracker{
		state: domainagent.SessionState{
			SchemaVersion: sessionSchemaVersion,
			SessionID:     clonedInput.SessionID,
			RequestID:     clonedInput.RequestID,
			Status:        domainagent.StatusRunning,
			Pipeline:      clonedPipeline,
			InitialInput:  clonedInput,
			Restore:       clonedInput.Restore,
			Metadata:      cloneStringMap(clonedInput.Metadata),
			StartedAt:     now,
			UpdatedAt:     now,
		},
		currentInput: clonedInput,
	}
}

func (s *sessionTracker) stageInput(stage domainagent.StageName) domainagent.AgentInput {
	next := cloneAgentInput(s.currentInput)
	next.Stage = stage
	return next
}

func (s *sessionTracker) completeStage(state domainagent.AgentState, output domainagent.AgentOutput) {
	s.state.CurrentStage = state.Stage
	s.state.Status = domainagent.StatusRunning
	s.state.UpdatedAt = state.Timing.CompletedAt
	s.state.StageStates = append(s.state.StageStates, cloneAgentState(state))
	s.state.FinalOutput = cloneAgentOutput(output)
	s.currentInput = mergeAgentInput(state.Input, output)
}

func (s *sessionTracker) failStage(state domainagent.AgentState, status domainagent.RunStatus, errDetail *domainagent.ErrorDetail) {
	s.state.CurrentStage = state.Stage
	s.state.Status = status
	s.state.Error = cloneErrorDetail(errDetail)
	s.state.UpdatedAt = state.Timing.CompletedAt
	s.state.CompletedAt = state.Timing.CompletedAt
	s.state.StageStates = append(s.state.StageStates, cloneAgentState(state))
}

func (s *sessionTracker) completeRun(at time.Time) {
	s.state.Status = domainagent.StatusCompleted
	s.state.UpdatedAt = at
	s.state.CompletedAt = at
}

func (s *sessionTracker) snapshot() domainagent.SessionState {
	return cloneSessionState(s.state)
}

func mergeAgentInput(input domainagent.AgentInput, output domainagent.AgentOutput) domainagent.AgentInput {
	next := cloneAgentInput(input)

	if output.Content != "" {
		next.Content = output.Content
	}
	if len(output.Messages) > 0 {
		next.Messages = cloneMessages(output.Messages)
	}
	if hasVisualIntent(output.VisualIntent) {
		next.VisualIntent = cloneVisualIntent(output.VisualIntent)
	}
	if len(output.RetrievedReferences) > 0 {
		next.RetrievedReferences = cloneReferences(output.RetrievedReferences)
	}
	if hasPrompt(output.Prompt) {
		next.Prompt = clonePrompt(output.Prompt)
	}
	if len(output.GeneratedArtifacts) > 0 {
		next.GeneratedArtifacts = cloneArtifacts(output.GeneratedArtifacts)
	}
	if len(output.CritiqueRounds) > 0 {
		next.CritiqueRounds = cloneCritiqueRounds(output.CritiqueRounds)
	}
	next.Metadata = mergeStringMaps(next.Metadata, output.Metadata)

	return next
}

func cloneSessionState(state domainagent.SessionState) domainagent.SessionState {
	cloned := state
	cloned.Pipeline = append([]domainagent.StageName(nil), state.Pipeline...)
	cloned.InitialInput = cloneAgentInput(state.InitialInput)
	cloned.StageStates = make([]domainagent.AgentState, len(state.StageStates))
	for i, item := range state.StageStates {
		cloned.StageStates[i] = cloneAgentState(item)
	}
	cloned.FinalOutput = cloneAgentOutput(state.FinalOutput)
	cloned.Error = cloneErrorDetail(state.Error)
	cloned.Metadata = cloneStringMap(state.Metadata)
	return cloned
}

func cloneAgentState(state domainagent.AgentState) domainagent.AgentState {
	cloned := state
	cloned.Input = cloneAgentInput(state.Input)
	cloned.Output = cloneAgentOutput(state.Output)
	cloned.Error = cloneErrorDetail(state.Error)
	return cloned
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

func cloneAgentOutput(output domainagent.AgentOutput) domainagent.AgentOutput {
	cloned := output
	cloned.Messages = cloneMessages(output.Messages)
	cloned.VisualIntent = cloneVisualIntent(output.VisualIntent)
	cloned.RetrievedReferences = cloneReferences(output.RetrievedReferences)
	cloned.Prompt = clonePrompt(output.Prompt)
	cloned.GeneratedArtifacts = cloneArtifacts(output.GeneratedArtifacts)
	cloned.CritiqueRounds = cloneCritiqueRounds(output.CritiqueRounds)
	cloned.Error = cloneErrorDetail(output.Error)
	cloned.Metadata = cloneStringMap(output.Metadata)
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

func cloneErrorDetail(detail *domainagent.ErrorDetail) *domainagent.ErrorDetail {
	if detail == nil {
		return nil
	}

	cloned := *detail
	return &cloned
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

func mergeStringMaps(base map[string]string, incoming map[string]string) map[string]string {
	switch {
	case len(base) == 0 && len(incoming) == 0:
		return nil
	case len(incoming) == 0:
		return cloneStringMap(base)
	}

	merged := cloneStringMap(base)
	if merged == nil {
		merged = make(map[string]string, len(incoming))
	}
	for key, value := range incoming {
		merged[key] = value
	}
	return merged
}

func hasVisualIntent(intent domainagent.VisualIntent) bool {
	return intent.Mode != "" ||
		intent.Goal != "" ||
		intent.Audience != "" ||
		intent.Style != "" ||
		len(intent.Constraints) > 0 ||
		len(intent.PreferredOutputs) > 0
}

func hasPrompt(prompt domainagent.PromptMetadata) bool {
	return prompt.SystemInstruction != "" ||
		prompt.Version != "" ||
		prompt.Template != "" ||
		len(prompt.Variables) > 0
}
