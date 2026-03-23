package planner

import (
	"encoding/json"
	"fmt"
	"strings"

	domainagent "github.com/paperbanana/paperbanana/internal/domain/agent"
	domainllm "github.com/paperbanana/paperbanana/internal/domain/llm"
)

const PromptVersion = "planner-v1"

type promptSpec struct {
	systemPrompt   string
	template       string
	contentLabel   string
	intentLabel    string
	referenceLabel string
	noTitleClause  string
}

func SystemPrompt(mode domainagent.VisualMode) (string, error) {
	spec, err := promptForMode(mode)
	if err != nil {
		return "", err
	}
	return spec.systemPrompt, nil
}

func promptTemplate(mode domainagent.VisualMode) (string, error) {
	spec, err := promptForMode(mode)
	if err != nil {
		return "", err
	}
	return spec.template, nil
}

func buildMessages(input domainagent.AgentInput, examples []referenceExample, loadImage imageLoader) ([]domainllm.Message, error) {
	spec, err := promptForMode(input.VisualIntent.Mode)
	if err != nil {
		return nil, err
	}

	parts := make([]domainllm.Part, 0, len(examples)*2+1)
	for idx, example := range examples {
		parts = append(parts, domainllm.TextPart(buildExamplePrompt(spec, idx+1, example)))

		if example.PathToGTImage == "" {
			continue
		}

		image, mimeType, err := loadImage(input.VisualIntent.Mode, example.PathToGTImage)
		if err != nil {
			return nil, err
		}
		parts = append(parts, domainllm.InlineImagePart(mimeType, image))
	}

	parts = append(parts, domainllm.TextPart(buildFinalPrompt(spec, input.Content, input.VisualIntent.Goal)))
	return []domainllm.Message{{
		Role:  domainllm.RoleUser,
		Parts: parts,
	}}, nil
}

func buildExamplePrompt(spec promptSpec, index int, example referenceExample) string {
	return fmt.Sprintf(
		"Example %d:\n%s: %s\n%s: %s\n%s: ",
		index,
		spec.contentLabel,
		example.ContentString(),
		spec.intentLabel,
		example.VisualIntent,
		spec.referenceLabel,
	)
}

func buildFinalPrompt(spec promptSpec, content, goal string) string {
	var builder strings.Builder
	builder.WriteString("Now, based on the following ")
	builder.WriteString(strings.ToLower(spec.contentLabel))
	builder.WriteString(" and ")
	builder.WriteString(strings.ToLower(spec.intentLabel))
	builder.WriteString(", provide a detailed description for the figure to be generated.\n")
	builder.WriteString(spec.contentLabel)
	builder.WriteString(": ")
	builder.WriteString(content)
	builder.WriteString("\n")
	builder.WriteString(spec.intentLabel)
	builder.WriteString(": ")
	builder.WriteString(goal)
	builder.WriteString("\nDetailed description of the target figure to be generated")
	builder.WriteString(spec.noTitleClause)
	builder.WriteString(":")
	return builder.String()
}

func promptForMode(mode domainagent.VisualMode) (promptSpec, error) {
	switch mode {
	case domainagent.VisualModeDiagram:
		return promptSpec{
			systemPrompt:   strings.TrimSpace(diagramSystemPrompt),
			template:       "planner/diagram-system",
			contentLabel:   "Methodology Section",
			intentLabel:    "Diagram Caption",
			referenceLabel: "Reference Diagram",
			noTitleClause:  " (do not include figure titles)",
		}, nil
	case domainagent.VisualModePlot:
		return promptSpec{
			systemPrompt:   strings.TrimSpace(plotSystemPrompt),
			template:       "planner/plot-system",
			contentLabel:   "Plot Raw Data",
			intentLabel:    "Visual Intent of the Desired Plot",
			referenceLabel: "Reference Plot",
		}, nil
	default:
		return promptSpec{}, fmt.Errorf("unsupported visual mode %q", mode)
	}
}

func collectPlanningContent(response *domainllm.GenerateResponse) string {
	if response == nil {
		return ""
	}
	if content := strings.TrimSpace(response.Content); content != "" {
		return content
	}
	return strings.TrimSpace(domainllm.CollectText(response.Parts))
}

type referenceExample struct {
	ID            string          `json:"id"`
	VisualIntent  string          `json:"visual_intent"`
	Content       json.RawMessage `json:"content"`
	PathToGTImage string          `json:"path_to_gt_image,omitempty"`
}

func (e referenceExample) ContentString() string {
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

const diagramSystemPrompt = `
I am working on a task: given the 'Methodology' section of a paper, and the caption of the desired figure, automatically generate a corresponding illustrative diagram. I will input the text of the 'Methodology' section, the figure caption, and your output should be a detailed description of an illustrative figure that effectively represents the methods described in the text.

To help you understand the task better, and grasp the principles for generating such figures, I will also provide you with several examples. You should learn from these examples to provide your figure description.

** IMPORTANT: **
Your description should be as detailed as possible. Semantically, clearly describe each element and their connections. Formally, include various details such as background style (typically pure white or very light pastel), colors, line thickness, icon styles, etc. Remember: vague or unclear specifications will only make the generated figure worse, not better.
`

const plotSystemPrompt = `
I am working on a task: given the raw data (typically in tabular or json format) and a visual intent of the desired plot, automatically generate a corresponding statistical plot that are both accurate and aesthetically pleasing. I will input the raw data and the plot visual intent, and your output should be a detailed description of an illustrative plot that effectively represents the data.  Note that your description should include all the raw data points to be plotted.

To help you understand the task better, and grasp the principles for generating such plots, I will also provide you with several examples. You should learn from these examples to provide your plot description.

** IMPORTANT: **
Your description should be as detailed as possible. For content, explain the precise mapping of variables to visual channels (x, y, hue) and explicitly enumerate every raw data point's coordinate to be drawn to ensure accuracy. For presentation, specify the exact aesthetic parameters, including specific HEX color codes, font sizes for all labels, line widths, marker dimensions, legend placement, and grid styles. You should learn from the examples' content presentation and aesthetic design (e.g., color schemes).
`
