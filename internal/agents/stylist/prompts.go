package stylist

import (
	"fmt"

	"github.com/paperbanana/paperbanana/internal/domain/agent"
	"github.com/paperbanana/paperbanana/internal/agents/stylist/styleguides"
)

func buildStylistPrompt(input agent.AgentInput) string {
	visualMode := string(input.VisualIntent.Mode)
	styleGuide := styleguides.GetStyleGuide(visualMode)

	return fmt.Sprintf(`You are a visualization style expert specializing in academic figures.

Your task is to enhance the following visualization plan while preserving its semantic content and meaning.

%s

## Visual Mode: %s

## Original Plan:
%s

## Instructions:
1. Enhance the visual description to align with NeurIPS 2025 style guidelines
2. Preserve all semantic content - do not change the meaning
3. Add specific style recommendations (colors, fonts, layout)
4. Ensure the enhanced plan is clear and actionable for the visualizer

Output the enhanced visualization plan:`,
		styleGuide,
		visualMode,
		input.Content)
}
