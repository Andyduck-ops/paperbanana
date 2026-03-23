package stylist

import (
	"fmt"

	domainagent "github.com/paperbanana/paperbanana/internal/domain/agent"
)

const PromptVersion = "1.0.0"

const styleGuideReference = `## NeurIPS 2025 Style Guidelines

### Diagram Style
- Use clean, minimalist layouts with clear visual hierarchy
- Prefer sans-serif fonts (Helvetica, Arial) for labels
- Use consistent stroke widths (1-2px for lines, 2-3px for emphasis)
- Color palette: primary blues (#1a73e8, #4285f4), accent oranges (#ff6d01), neutral grays (#5f6368)
- Maintain adequate whitespace between components
- Use rounded rectangles for containers, arrows for flow

### Plot Style
- Axis labels with units in parentheses
- Legend placed outside plot area when possible
- Grid lines in light gray (#e0e0e0)
- Markers: circles for data points, consistent size
- Line styles: solid for primary, dashed for secondary
- Error bars where applicable`

func buildPrompt(mode domainagent.VisualMode) domainagent.PromptMetadata {
	systemInstruction := fmt.Sprintf(`You are a visualization style expert specializing in academic figures.

%s

## Your Task
Enhance the visualization plan while preserving its semantic content and meaning.
Apply the appropriate style guidelines based on the visual mode (%s).
Add specific style recommendations (colors, fonts, layout) to make the plan actionable for the visualizer.`, styleGuideReference, mode)

	template := "stylist/enhance-prompt"

	return domainagent.PromptMetadata{
		SystemInstruction: systemInstruction,
		Version:           PromptVersion,
		Template:          template,
		Variables: map[string]string{
			"mode": string(mode),
		},
	}
}

func buildMessageContent(input domainagent.AgentInput) string {
	return fmt.Sprintf(`## Visual Mode: %s

## Original Plan:
%s

## Instructions:
1. Enhance the visual description to align with NeurIPS 2025 style guidelines
2. Preserve all semantic content - do not change the meaning
3. Add specific style recommendations (colors, fonts, layout)
4. Ensure the enhanced plan is clear and actionable for the visualizer

Output the enhanced visualization plan:`, input.VisualIntent.Mode, input.Content)
}
