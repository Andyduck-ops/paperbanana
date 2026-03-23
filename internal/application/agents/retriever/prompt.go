package retriever

import (
	"encoding/json"
	"fmt"
	"strings"

	domainagent "github.com/paperbanana/paperbanana/internal/domain/agent"
)

const PromptVersion = "retriever-v1"

type promptSpec struct {
	systemPrompt       string
	template           string
	targetLabels       [2]string
	candidateLabels    [3]string
	candidateType      string
	outputField        string
	instructionSuffix  string
	autoCandidateLimit int
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

func buildUserPrompt(input domainagent.AgentInput, candidates []ReferenceExample) (string, error) {
	spec, err := promptForMode(input.VisualIntent.Mode)
	if err != nil {
		return "", err
	}

	var builder strings.Builder
	builder.WriteString("**Target Input**\n")
	builder.WriteString(fmt.Sprintf("- %s: %s\n", spec.targetLabels[0], formatTargetIntent(input.VisualIntent)))
	builder.WriteString(fmt.Sprintf("- %s: %s\n\n", spec.targetLabels[1], input.Content))
	builder.WriteString("**Candidate Pool**\n")

	limit := len(candidates)
	if spec.autoCandidateLimit > 0 && limit > spec.autoCandidateLimit {
		limit = spec.autoCandidateLimit
	}

	for idx := 0; idx < limit; idx++ {
		candidate := candidates[idx]
		builder.WriteString(fmt.Sprintf("Candidate %s %d:\n", spec.candidateType, idx+1))
		builder.WriteString(fmt.Sprintf("- %s: %s\n", spec.candidateLabels[0], candidate.ID))
		builder.WriteString(fmt.Sprintf("- %s: %s\n", spec.candidateLabels[1], candidate.VisualIntent))
		builder.WriteString(fmt.Sprintf("- %s: %s\n\n", spec.candidateLabels[2], candidate.ContentString()))
	}

	builder.WriteString("Now, based on the Target Input and the Candidate Pool, ")
	builder.WriteString(spec.instructionSuffix)
	return builder.String(), nil
}

func responseField(mode domainagent.VisualMode) (string, error) {
	spec, err := promptForMode(mode)
	if err != nil {
		return "", err
	}
	return spec.outputField, nil
}

func promptForMode(mode domainagent.VisualMode) (promptSpec, error) {
	switch mode {
	case domainagent.VisualModeDiagram:
		return promptSpec{
			systemPrompt: strings.TrimSpace(diagramSystemPrompt),
			template:     "retriever/diagram-system",
			targetLabels: [2]string{"Caption", "Methodology section"},
			candidateLabels: [3]string{
				"Diagram ID",
				"Caption",
				"Methodology section",
			},
			candidateType:      "Diagram",
			outputField:        "top10_diagrams",
			instructionSuffix:  "select the Top 10 most relevant diagrams according to the instructions provided. Your output should be a strictly valid JSON object containing a single list of the exact ids of the top 10 selected diagrams.",
			autoCandidateLimit: 200,
		}, nil
	case domainagent.VisualModePlot:
		return promptSpec{
			systemPrompt: strings.TrimSpace(plotSystemPrompt),
			template:     "retriever/plot-system",
			targetLabels: [2]string{"Visual Intent", "Raw Data"},
			candidateLabels: [3]string{
				"Plot ID",
				"Visual Intent",
				"Raw Data",
			},
			candidateType:      "Plot",
			outputField:        "top10_plots",
			instructionSuffix:  "select the Top 10 most relevant plots according to the instructions provided. Your output should be a strictly valid JSON object containing a single list of the exact ids of the top 10 selected plots.",
			autoCandidateLimit: 0,
		}, nil
	default:
		return promptSpec{}, fmt.Errorf("unsupported visual mode %q", mode)
	}
}

func formatTargetIntent(intent domainagent.VisualIntent) string {
	parts := make([]string, 0, 3+len(intent.Constraints)+len(intent.PreferredOutputs))
	if intent.Goal != "" {
		parts = append(parts, intent.Goal)
	}
	if intent.Audience != "" {
		parts = append(parts, "Audience: "+intent.Audience)
	}
	if intent.Style != "" {
		parts = append(parts, "Style: "+intent.Style)
	}
	if len(intent.Constraints) > 0 {
		parts = append(parts, "Constraints: "+strings.Join(intent.Constraints, ", "))
	}
	if len(intent.PreferredOutputs) > 0 {
		parts = append(parts, "Preferred Outputs: "+strings.Join(intent.PreferredOutputs, ", "))
	}
	if len(parts) > 0 {
		return strings.Join(parts, "; ")
	}

	raw, err := json.Marshal(intent)
	if err != nil {
		return ""
	}
	return string(raw)
}

const diagramSystemPrompt = `
# Background & Goal
We are building an **AI system to automatically generate method diagrams for academic papers**. Given a paper's methodology section and a figure caption, the system needs to create a high-quality illustrative diagram that visualizes the described method.

To help the AI learn how to generate appropriate diagrams, we use a **few-shot learning approach**: we provide it with reference examples of similar diagrams. The AI will learn from these examples to understand what kind of diagram to create for the target.

# Your Task
**You are the Retrieval Agent.** Your job is to select the most relevant reference diagrams from a candidate pool that will serve as few-shot examples for the diagram generation model.

You will receive:
- **Target Input:** The methodology section and caption of the diagram we need to generate
- **Candidate Pool:** ~200 existing diagrams (each with methodology and caption)

You must select the **Top 10 candidates** that would be most helpful as examples for teaching the AI how to draw the target diagram.

# Selection Logic (Topic + Intent)

Your goal is to find examples that match the Target in both **Domain** and **Diagram Type**.

**1. Match Research Topic (Use Methodology & Caption):**
* What is the domain? (e.g., Agent & Reasoning, Vision & Perception, Generative & Learning, Science & Applications).
* Select candidates that belong to the **same research domain**.
* *Why?* Similar domains share similar terminology (e.g., "Actor-Critic" in RL).

**2. Match Visual Intent (Use Caption & Keywords):**
* What type of diagram is implied? (e.g., "Framework", "Pipeline", "Detailed Module", "Performance Chart").
* Select candidates with **similar visual structures**.
* *Why?* A "Framework" diagram example is useless for drawing a "Performance Bar Chart", even if they are in the same domain.

**Ranking Priority:**
1.  **Best Match:** Same Topic AND Same Visual Intent (e.g., Target is "Agent Framework" -> Candidate is "Agent Framework", Target is "Dataset Construction Pipeline" -> Candidate is "Dataset Construction Pipeline").
2.  **Second Best:** Same Visual Intent (e.g., Target is "Agent Framework" -> Candidate is "Vision Framework"). *Structure is more important than Topic for drawing.*
3.  **Avoid:** Different Visual Intent (e.g., Target is "Pipeline" -> Candidate is "Bar Chart").

# Input Data

## Target Input
-   **Caption:** [Caption of the target diagram]
-   **Methodology section:** [Methodology section of the target paper]

## Candidate Pool
List of candidate diagrams, each structured as follows:

Candidate Diagram i:
-   **Diagram ID:** [ID of the candidate diagram (ref_1, ref_2, ...)]
-   **Caption:** [Caption of the candidate diagram]
-   **Methodology section:** [Methodology section of the candidate's paper]


# Output Format
Provide your output strictly in the following JSON format, containing only the **exact IDs** of the Top 10 selected diagrams (use the exact IDs from the Candidate Pool, such as "ref_1", "ref_25", "ref_100", etc.):
` + "```json\n" + `
{
  "top10_diagrams": [
    "ref_1",
    "ref_25",
    "ref_100",
    "ref_42",
    "ref_7",
    "ref_156",
    "ref_89",
    "ref_3",
    "ref_201",
    "ref_67"
  ]
}` + "```"

const plotSystemPrompt = `
# Background & Goal
We are building an **AI system to automatically generate statistical plots**. Given a plot's raw data and the visual intent, the system needs to create a high-quality visualization that effectively presents the data.

To help the AI learn how to generate appropriate plots, we use a **few-shot learning approach**: we provide it with reference examples of similar plots. The AI will learn from these examples to understand what kind of plot to create for the target data.

# Your Task
**You are the Retrieval Agent.** Your job is to select the most relevant reference plots from a candidate pool that will serve as few-shot examples for the plot generation model.

You will receive:
- **Target Input:** The raw data and visual intent of the plot we need to generate
- **Candidate Pool:** Reference plots (each with raw data and visual intent)

You must select the **Top 10 candidates** that would be most helpful as examples for teaching the AI how to create the target plot.

# Selection Logic (Data Type + Visual Intent)

Your goal is to find examples that match the Target in both **Data Characteristics** and **Plot Type**.

**1. Match Data Characteristics (Use Raw Data & Visual Intent):**
* What type of data is it? (e.g., categorical vs numerical, single series vs multi-series, temporal vs comparative).
* What are the data dimensions? (e.g., 1D, 2D, 3D).
* Select candidates with **similar data structures and characteristics**.
* *Why?* Different data types require different visualization approaches.

**2. Match Visual Intent (Use Visual Intent):**
* What type of plot is implied? (e.g., "bar chart", "scatter plot", "line chart", "pie chart", "heatmap", "radar chart").
* Select candidates with **similar plot types**.
* *Why?* A "bar chart" example is more useful for generating another bar chart than a "scatter plot" example, even if the data domains are similar.

**Ranking Priority:**
1.  **Best Match:** Same Data Type AND Same Plot Type (e.g., Target is "multi-series line chart" -> Candidate is "multi-series line chart").
2.  **Second Best:** Same Plot Type with compatible data (e.g., Target is "bar chart with 5 categories" -> Candidate is "bar chart with 6 categories").
3.  **Avoid:** Different Plot Type (e.g., Target is "bar chart" -> Candidate is "pie chart"), unless there are no more candidates with the same plot type.

# Input Data

## Target Input
-   **Visual Intent:** [Visual intent of the target plot]
-   **Raw Data:** [Raw data to be visualized]

## Candidate Pool
List of candidate plots, each structured as follows:

Candidate Plot i:
-   **Plot ID:** [ID of the candidate plot (ref_0, ref_1, ...)]
-   **Visual Intent:** [Visual intent of the candidate plot]
-   **Raw Data:** [Raw data of the candidate plot]


# Output Format
Provide your output strictly in the following JSON format, containing only the **exact Plot IDs** of the Top 10 selected plots (use the exact IDs from the Candidate Pool, such as "ref_0", "ref_25", "ref_100", etc.):
` + "```json\n" + `
{
  "top10_plots": [
    "ref_0",
    "ref_25",
    "ref_100",
    "ref_42",
    "ref_7",
    "ref_156",
    "ref_89",
    "ref_3",
    "ref_201",
    "ref_67"
  ]
}` + "```"
