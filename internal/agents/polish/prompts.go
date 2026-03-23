package polish

import (
	"fmt"

	domainagent "github.com/paperbanana/paperbanana/internal/domain/agent"
)

// buildPolishPrompt creates a prompt for image enhancement with resolution-specific hints.
func buildPolishPrompt(input domainagent.AgentInput, resolution string) string {
	resolutionHint := "high quality"
	if resolution == "4K" {
		resolutionHint = "4K resolution (3840x2160), maximum detail"
	} else if resolution == "2K" {
		resolutionHint = "2K resolution (2560x1440), high detail"
	}

	return fmt.Sprintf(`You are an expert at enhancing and refining visualizations for academic publications.

Your task is to improve the provided image according to the following instructions.

## Target Resolution: %s

## Refinement Instructions:
%s

## Guidelines:
1. Enhance clarity and readability of all text and labels
2. Improve color contrast and visual hierarchy
3. Sharpen lines and edges for print quality
4. Maintain the original semantic content and meaning
5. Apply NeurIPS 2025 style guidelines where appropriate

Generate the enhanced visualization code:`,
		resolutionHint,
		input.Content)
}
