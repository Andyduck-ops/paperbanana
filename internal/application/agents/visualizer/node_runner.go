package visualizer

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	pbconfig "github.com/paperbanana/paperbanana/internal/config"
	domainagent "github.com/paperbanana/paperbanana/internal/domain/agent"
	"github.com/paperbanana/paperbanana/internal/infrastructure/nodes/httpnode"
)

const nodeNameMetadataKey = "visualizer.node_name"

type nodeExecutor interface {
	Execute(ctx context.Context, node pbconfig.NodeDefinition) (httpnode.Result, error)
}

type nodeRunner struct {
	catalog *pbconfig.NodeCatalog
	adapter nodeExecutor
}

func newNodeRunner(catalog *pbconfig.NodeCatalog, adapter nodeExecutor) *nodeRunner {
	return &nodeRunner{
		catalog: catalog,
		adapter: adapter,
	}
}

func (r *nodeRunner) enabled(input domainagent.AgentInput) bool {
	return strings.TrimSpace(input.Metadata[nodeNameMetadataKey]) != ""
}

func (r *nodeRunner) execute(ctx context.Context, input domainagent.AgentInput, prompt domainagent.PromptMetadata) (domainagent.AgentOutput, error) {
	if r.catalog == nil {
		return domainagent.AgentOutput{}, errors.New("visualizer configured-node path requires a node catalog")
	}
	if r.adapter == nil {
		return domainagent.AgentOutput{}, errors.New("visualizer configured-node path requires a node adapter")
	}

	nodeName := strings.TrimSpace(input.Metadata[nodeNameMetadataKey])
	node, ok := r.catalog.NodeByName(nodeName)
	if !ok {
		return domainagent.AgentOutput{}, fmt.Errorf("configured visualizer node %q not found", nodeName)
	}
	resolvedTemplate, ok := resolveRequestTemplate(node.RequestTemplate, input).(map[string]any)
	if !ok {
		return domainagent.AgentOutput{}, fmt.Errorf("configured visualizer node %q resolved to an invalid request template", nodeName)
	}
	node.RequestTemplate = resolvedTemplate

	result, err := r.adapter.Execute(ctx, node)
	if err != nil {
		return domainagent.AgentOutput{}, err
	}

	output := domainagent.AgentOutput{
		Stage:               domainagent.StageVisualizer,
		Content:             input.Content,
		VisualIntent:        cloneVisualIntent(input.VisualIntent),
		RetrievedReferences: cloneReferences(input.RetrievedReferences),
		Prompt:              prompt,
		GeneratedArtifacts:  cloneArtifacts(input.GeneratedArtifacts),
		CritiqueRounds:      cloneCritiqueRounds(input.CritiqueRounds),
	}

	if code := stringField(result.Body, "code", "plot_code"); code != "" {
		output.GeneratedArtifacts = append(output.GeneratedArtifacts, promptTraceArtifact(input.VisualIntent.Mode, code))
	}

	mimeType, bytes, err := imageArtifactFromNodeBody(result.Body)
	if err != nil {
		return domainagent.AgentOutput{}, err
	}
	output.GeneratedArtifacts = append(output.GeneratedArtifacts, renderedArtifact(input.VisualIntent.Mode, mimeType, bytes))
	output.Metadata = map[string]string{
		"execution_path": "configured-node",
		"node_name":      node.Name,
		"summary":        nonEmpty(stringField(result.Body, "summary"), fmt.Sprintf("generated %s figure artifact via configured node", input.VisualIntent.Mode)),
	}
	return output, nil
}

func resolveRequestTemplate(value any, input domainagent.AgentInput) any {
	switch typed := value.(type) {
	case map[string]any:
		resolved := make(map[string]any, len(typed))
		for key, nested := range typed {
			resolved[key] = resolveRequestTemplate(nested, input)
		}
		return resolved
	case []any:
		resolved := make([]any, len(typed))
		for i, item := range typed {
			resolved[i] = resolveRequestTemplate(item, input)
		}
		return resolved
	case string:
		return resolveTemplateString(typed, input)
	default:
		return value
	}
}

func resolveTemplateString(template string, input domainagent.AgentInput) any {
	if len(template) >= 4 && strings.HasPrefix(template, "{{") && strings.HasSuffix(template, "}}") {
		key := strings.TrimSpace(template[2 : len(template)-2])
		if value, ok := templateValue(key, input); ok {
			return value
		}
	}

	resolved := template
	for {
		start := strings.Index(resolved, "{{")
		if start < 0 {
			return resolved
		}
		end := strings.Index(resolved[start:], "}}")
		if end < 0 {
			return resolved
		}

		end += start
		key := strings.TrimSpace(resolved[start+2 : end])
		value, ok := templateValue(key, input)
		if !ok {
			return resolved
		}
		resolved = resolved[:start] + fmt.Sprint(value) + resolved[end+2:]
	}
}

func templateValue(key string, input domainagent.AgentInput) (any, bool) {
	switch key {
	case "content":
		return input.Content, true
	case "request_id":
		return input.RequestID, true
	case "session_id":
		return input.SessionID, true
	case "stage":
		return input.Stage, true
	case "visual_intent.mode":
		return input.VisualIntent.Mode, true
	case "visual_intent.goal":
		return input.VisualIntent.Goal, true
	case "visual_intent.audience":
		return input.VisualIntent.Audience, true
	case "visual_intent.style":
		return input.VisualIntent.Style, true
	}

	if strings.HasPrefix(key, "metadata.") {
		value, ok := input.Metadata[strings.TrimPrefix(key, "metadata.")]
		return value, ok
	}

	return nil, false
}

func imageArtifactFromNodeBody(body map[string]any) (string, []byte, error) {
	encoded := stringField(body, "image_base64", "base64_jpg", "artifact_base64")
	if encoded == "" {
		return "", nil, errors.New("configured visualizer node returned no image payload")
	}

	bytes, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", nil, fmt.Errorf("decode configured node image payload: %w", err)
	}

	mimeType := nonEmpty(stringField(body, "mime_type"), "image/png")
	return mimeType, bytes, nil
}

func stringField(body map[string]any, keys ...string) string {
	for _, key := range keys {
		value, ok := body[key]
		if !ok {
			continue
		}
		if text, ok := value.(string); ok {
			return strings.TrimSpace(text)
		}
	}
	return ""
}

func nonEmpty(value, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}
