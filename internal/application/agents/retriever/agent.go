package retriever

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/paperbanana/paperbanana/internal/application/agents/modelselection"
	domainagent "github.com/paperbanana/paperbanana/internal/domain/agent"
	domainllm "github.com/paperbanana/paperbanana/internal/domain/llm"
)

const (
	defaultMaxOutputTokens = 50000
	retrievedSource        = "paperbanana-bench"
)

type RetrievalMode string

const (
	RetrievalModeAuto   RetrievalMode = "auto"
	RetrievalModeManual RetrievalMode = "manual"
	RetrievalModeRandom RetrievalMode = "random"
	RetrievalModeNone   RetrievalMode = "none"
)

type ReferenceExample struct {
	ID            string          `json:"id"`
	VisualIntent  string          `json:"visual_intent"`
	Content       json.RawMessage `json:"content"`
	PathToGTImage string          `json:"path_to_gt_image,omitempty"`
}

func (e ReferenceExample) ContentString() string {
	if len(e.Content) == 0 {
		return ""
	}

	var text string
	if err := json.Unmarshal(e.Content, &text); err == nil {
		return text
	}

	var payload any
	if err := json.Unmarshal(e.Content, &payload); err == nil {
		compacted, marshalErr := json.Marshal(payload)
		if marshalErr == nil {
			return string(compacted)
		}
	}

	return string(e.Content)
}

type Store interface {
	Candidates(ctx context.Context, mode domainagent.VisualMode) ([]ReferenceExample, error)
	ManualExamples(ctx context.Context, mode domainagent.VisualMode) ([]ReferenceExample, error)
}

type FileStore struct {
	Root string
}

func (s FileStore) Candidates(_ context.Context, mode domainagent.VisualMode) ([]ReferenceExample, error) {
	return s.read(filepath.Join(s.Root, modeDir(mode), "ref.json"))
}

func (s FileStore) ManualExamples(_ context.Context, mode domainagent.VisualMode) ([]ReferenceExample, error) {
	return s.read(filepath.Join(s.Root, modeDir(mode), "agent_selected_12.json"))
}

func (s FileStore) read(path string) ([]ReferenceExample, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var items []ReferenceExample
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, err
	}
	return items, nil
}

type Config struct {
	Mode            RetrievalMode
	Store           Store
	Model           string
	Temperature     float64
	MaxOutputTokens int
	Random          *rand.Rand
	Now             func() time.Time
}

type Agent struct {
	client domainllm.LLMClient
	cfg    Config
	state  domainagent.AgentState
}

func NewAgent(client domainllm.LLMClient, cfg Config) *Agent {
	if cfg.Mode == "" {
		cfg.Mode = RetrievalModeAuto
	}
	if cfg.MaxOutputTokens <= 0 {
		cfg.MaxOutputTokens = defaultMaxOutputTokens
	}
	if cfg.Now == nil {
		cfg.Now = func() time.Time { return time.Now().UTC() }
	}
	if cfg.Random == nil {
		cfg.Random = rand.New(rand.NewSource(cfg.Now().UnixNano()))
	}

	return &Agent{
		client: client,
		cfg:    cfg,
		state: domainagent.AgentState{
			Stage:  domainagent.StageRetriever,
			Status: domainagent.StatusPending,
		},
	}
}

func (a *Agent) Initialize(context.Context) error {
	a.state.Stage = domainagent.StageRetriever
	a.state.Status = domainagent.StatusRunning
	a.state.Error = nil
	return nil
}

func (a *Agent) Execute(ctx context.Context, input domainagent.AgentInput) (domainagent.AgentOutput, error) {
	mode, err := a.resolveMode(input)
	if err != nil {
		return domainagent.AgentOutput{}, err
	}

	prompt, err := a.promptMetadata(input.VisualIntent.Mode)
	if err != nil {
		return domainagent.AgentOutput{}, err
	}

	a.state.Stage = domainagent.StageRetriever
	a.state.Status = domainagent.StatusRunning
	a.state.Input = input
	a.state.Error = nil
	a.state.Output = domainagent.AgentOutput{
		Stage:        domainagent.StageRetriever,
		VisualIntent: input.VisualIntent,
		Prompt:       prompt,
	}

	output, err := a.executeMode(ctx, input, mode, prompt)
	if err != nil {
		a.state.Status = domainagent.StatusFailed
		a.state.Error = &domainagent.ErrorDetail{
			Message: err.Error(),
			Stage:   domainagent.StageRetriever,
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

func (a *Agent) executeMode(ctx context.Context, input domainagent.AgentInput, mode RetrievalMode, prompt domainagent.PromptMetadata) (domainagent.AgentOutput, error) {
	switch mode {
	case RetrievalModeNone:
		return a.buildOutput(input, prompt, mode, nil, nil), nil
	case RetrievalModeManual:
		examples, err := a.loadManualExamples(ctx, input.VisualIntent.Mode)
		if err != nil {
			return domainagent.AgentOutput{}, err
		}
		examples = limitExamples(examples, 10)
		return a.buildOutput(input, prompt, mode, selectedExampleIDs(examples), examples), nil
	case RetrievalModeRandom:
		candidates, err := a.loadCandidates(ctx, input.VisualIntent.Mode)
		if err != nil {
			return domainagent.AgentOutput{}, err
		}
		selected := a.randomExamples(candidates, 10)
		return a.buildOutput(input, prompt, mode, selectedExampleIDs(selected), selected), nil
	case RetrievalModeAuto:
		candidates, err := a.loadCandidates(ctx, input.VisualIntent.Mode)
		if err != nil {
			return domainagent.AgentOutput{}, err
		}
		if len(candidates) == 0 {
			return a.buildOutput(input, prompt, RetrievalModeNone, nil, nil), nil
		}
		if a.client == nil {
			return domainagent.AgentOutput{}, errors.New("retriever requires an llm client in auto mode")
		}

		userPrompt, err := buildUserPrompt(input, candidates)
		if err != nil {
			return domainagent.AgentOutput{}, err
		}
		req := domainllm.GenerateRequest{
			SystemInstruction: prompt.SystemInstruction,
			Messages: []domainllm.Message{
				{
					Role:  domainllm.RoleUser,
					Parts: []domainllm.Part{domainllm.TextPart(userPrompt)},
				},
			},
			Model:         modelselection.QueryModel(input.Metadata, a.cfg.Model),
			Temperature:   a.cfg.Temperature,
			MaxTokens:     a.cfg.MaxOutputTokens,
			PromptVersion: prompt.Version,
		}
		resp, err := a.client.Generate(ctx, req)
		if err != nil {
			return domainagent.AgentOutput{}, err
		}

		ids := ParseTopReferences(resp.Content, input.VisualIntent.Mode)
		return a.buildOutput(input, prompt, mode, ids, selectExamples(ids, candidates)), nil
	default:
		return domainagent.AgentOutput{}, fmt.Errorf("unsupported retrieval mode %q", mode)
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

func (a *Agent) resolveMode(input domainagent.AgentInput) (RetrievalMode, error) {
	return parseMode(modelselection.RetrievalMode(input.Metadata, string(a.cfg.Mode)))
}

func parseMode(value string) (RetrievalMode, error) {
	switch RetrievalMode(value) {
	case RetrievalModeAuto, RetrievalModeManual, RetrievalModeRandom, RetrievalModeNone:
		return RetrievalMode(value), nil
	default:
		return "", fmt.Errorf("unknown retrieval_setting: %s", value)
	}
}

func (a *Agent) loadCandidates(ctx context.Context, mode domainagent.VisualMode) ([]ReferenceExample, error) {
	if a.cfg.Store == nil {
		return nil, errors.New("retriever store is not configured")
	}
	items, err := a.cfg.Store.Candidates(ctx, mode)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	return items, err
}

func (a *Agent) loadManualExamples(ctx context.Context, mode domainagent.VisualMode) ([]ReferenceExample, error) {
	if a.cfg.Store == nil {
		return nil, errors.New("retriever store is not configured")
	}
	items, err := a.cfg.Store.ManualExamples(ctx, mode)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	return items, err
}

func (a *Agent) randomExamples(items []ReferenceExample, limit int) []ReferenceExample {
	if len(items) <= limit {
		return append([]ReferenceExample(nil), items...)
	}

	indexes := a.cfg.Random.Perm(len(items))[:limit]
	sort.Ints(indexes)
	selected := make([]ReferenceExample, 0, len(indexes))
	for _, index := range indexes {
		selected = append(selected, items[index])
	}
	return selected
}

func (a *Agent) buildOutput(input domainagent.AgentInput, prompt domainagent.PromptMetadata, mode RetrievalMode, ids []string, examples []ReferenceExample) domainagent.AgentOutput {
	now := a.cfg.Now()
	references := makeRetrievedReferences(ids, examples, now)
	output := domainagent.AgentOutput{
		Stage:               domainagent.StageRetriever,
		VisualIntent:        input.VisualIntent,
		RetrievedReferences: references,
		Prompt:              prompt,
		Metadata: map[string]string{
			"retrieval_setting": string(mode),
			"retrieved_count":   fmt.Sprintf("%d", len(references)),
		},
	}
	if len(examples) > 0 {
		content, err := json.Marshal(examples)
		if err == nil {
			output.GeneratedArtifacts = []domainagent.Artifact{
				{
					ID:       fmt.Sprintf("retriever-%s-examples", input.VisualIntent.Mode),
					Kind:     domainagent.ArtifactKindReferenceBundle,
					MIMEType: "application/json",
					URI:      fmt.Sprintf("memory://retriever/%s/examples", input.VisualIntent.Mode),
					Content:  string(content),
					Metadata: map[string]string{
						"mode":  string(input.VisualIntent.Mode),
						"count": fmt.Sprintf("%d", len(examples)),
					},
				},
			}
		}
	}
	return output
}

func ParseTopReferences(raw string, mode domainagent.VisualMode) []string {
	field, err := responseField(mode)
	if err != nil {
		return nil
	}

	for _, candidate := range []string{strings.TrimSpace(raw), extractJSONObject(raw)} {
		ids := decodeStructuredResponse(candidate, field)
		if len(ids) > 0 {
			return ids
		}
	}

	ids := extractArrayFallback(raw, field)
	if len(ids) > 0 {
		return ids
	}
	return nil
}

func decodeStructuredResponse(raw, field string) []string {
	if raw == "" {
		return nil
	}

	var payload map[string]json.RawMessage
	if err := json.Unmarshal([]byte(stripCodeFence(raw)), &payload); err != nil {
		return nil
	}

	value, ok := payload[field]
	if !ok {
		return nil
	}

	var ids []string
	if err := json.Unmarshal(value, &ids); err != nil {
		return nil
	}
	return dedupeIDs(ids)
}

func extractJSONObject(raw string) string {
	cleaned := stripCodeFence(raw)
	start := strings.Index(cleaned, "{")
	end := strings.LastIndex(cleaned, "}")
	if start == -1 || end == -1 || end <= start {
		return ""
	}
	return cleaned[start : end+1]
}

func stripCodeFence(raw string) string {
	trimmed := strings.TrimSpace(raw)
	trimmed = strings.TrimPrefix(trimmed, "```json")
	trimmed = strings.TrimPrefix(trimmed, "```")
	trimmed = strings.TrimSuffix(trimmed, "```")
	return strings.TrimSpace(trimmed)
}

func extractArrayFallback(raw, field string) []string {
	marker := `"` + field + `"`
	index := strings.Index(raw, marker)
	if index == -1 {
		return nil
	}

	remainder := raw[index+len(marker):]
	start := strings.Index(remainder, "[")
	if start == -1 {
		return nil
	}
	remainder = remainder[start:]
	end := strings.Index(remainder, "]")
	if end >= 0 {
		remainder = remainder[:end+1]
	}

	matches := regexp.MustCompile(`"([^"]+)"`).FindAllStringSubmatch(remainder, -1)
	ids := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) > 1 {
			ids = append(ids, match[1])
		}
	}
	return dedupeIDs(ids)
}

func limitExamples(items []ReferenceExample, limit int) []ReferenceExample {
	if len(items) <= limit {
		return append([]ReferenceExample(nil), items...)
	}
	return append([]ReferenceExample(nil), items[:limit]...)
}

func selectExamples(ids []string, candidates []ReferenceExample) []ReferenceExample {
	if len(ids) == 0 || len(candidates) == 0 {
		return nil
	}

	index := make(map[string]ReferenceExample, len(candidates))
	for _, candidate := range candidates {
		index[candidate.ID] = candidate
	}

	selected := make([]ReferenceExample, 0, len(ids))
	for _, id := range ids {
		candidate, ok := index[id]
		if ok {
			selected = append(selected, candidate)
		}
	}
	return selected
}

func makeRetrievedReferences(ids []string, examples []ReferenceExample, retrievedAt time.Time) []domainagent.RetrievedReference {
	if len(ids) == 0 {
		return nil
	}

	index := make(map[string]ReferenceExample, len(examples))
	for _, example := range examples {
		index[example.ID] = example
	}

	references := make([]domainagent.RetrievedReference, 0, len(ids))
	for position, id := range ids {
		example, ok := index[id]
		reference := domainagent.RetrievedReference{
			ID:          id,
			Source:      retrievedSource,
			RetrievedAt: retrievedAt,
			Score:       float64(len(ids) - position),
		}
		if ok {
			reference.Title = example.VisualIntent
			reference.URI = example.PathToGTImage
			reference.Summary = truncate(example.ContentString(), 280)
			reference.Snippets = []string{example.VisualIntent}
		}
		references = append(references, reference)
	}
	return references
}

func truncate(value string, limit int) string {
	if len(value) <= limit {
		return value
	}
	return value[:limit]
}

func dedupeIDs(ids []string) []string {
	if len(ids) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(ids))
	deduped := make([]string, 0, len(ids))
	for _, id := range ids {
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		deduped = append(deduped, id)
	}
	return deduped
}

func selectedExampleIDs(examples []ReferenceExample) []string {
	ids := make([]string, 0, len(examples))
	for _, example := range examples {
		ids = append(ids, example.ID)
	}
	return ids
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
