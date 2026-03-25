package stylist

import (
	"fmt"

	domainagent "github.com/paperbanana/paperbanana/internal/domain/agent"
)

const PromptVersion = "1.1.0"

const styleGuideReference = `## NeurIPS 2025 Style Guidelines

### Diagram Style Guide

#### 1. The "NeurIPS Look"
The prevailing aesthetic for 2025 is **"Soft Tech & Scientific Pastels."**
Gone are the days of harsh primary colors and sharp black boxes. The modern NeurIPS diagram feels approachable yet precise. It utilizes high-value (light) backgrounds to organize complexity, reserving saturation for the most critical active elements. The vibe balances **clean modularity** (clear separation of parts) with **narrative flow** (clear left-to-right progression).

#### 2. Detailed Style Options

##### A. Color Palettes
*Design Philosophy: Use color to group logic, not just to decorate. Avoid fully saturated backgrounds.*

**Background Fills (The "Zone" Strategy)**
*Used to encapsulate stages (e.g., "Pre-training phase") or environments.*
*   **Most papers use:** Very light, desaturated pastels (Opacity ~10–15%).
*   **Aesthetically pleasing options include:**
    *   Cream / Beige (e.g., #F5F5DC) – Warm, academic feel.
    *   Pale Blue / Ice (e.g., #E6F3FF) – Clean, technical feel.
    *   Mint / Sage (e.g., #E0F2F1) – Soft, organic feel.
    *   Pale Lavender (e.g., #F3E5F5) – distinctive, modern feel.
*   **Alternative (~20%):** White backgrounds with colored *dashed borders* for a high-contrast, minimalist look (common in theoretical papers).

**Functional Element Colors**
*   **For "Active" Modules (Encoders, MLP, Attention):** Medium saturation is preferred.
    *   *Common pairings:* Blue/Orange, Green/Purple, or Teal/Pink.
    *   *Observation:* Colors are often used to distinguish **status** rather than component type:
        *   **Trainable Elements:** Often Warm tones (Red, Orange, Deep Pink).
        *   **Frozen/Static Elements:** Often Cool tones (Grey, Ice Blue, Cyan).
*   **For Highlights/Results:** High saturation (Primary Red, Bright Gold) is strictly reserved for "Error/Loss," "Ground Truth," or the final output.

##### B. Shapes & Containers
*Design Philosophy: "Softened Geometry." Sharp corners are for data; rounded corners are for processes.*

**Core Components**
*   **Process Nodes (The Standard):** Rounded Rectangles (Corner radius 5–10px). This is the dominant shape (~80%) for generic layers or steps.
*   **Tensors & Data:**
    *   **3D Stacks/Cuboids:** Used to imply depth/volume (e.g., B x H x W).
    *   **Flat Squares/Grids:** Used for matrices, tokens, or attention maps.
    *   **Cylinders:** Exclusively reserved for Databases, Buffers, or Memory.

**Grouping & Hierarchy**
*   **The "Macro-Micro" Pattern:** A solid, light-colored container represents the global view, with a specific module (e.g., "Attention Block") connected via lines to a "zoomed-in" detailed breakout box.
*   **Borders:**
    *   **Solid:** For physical components.
    *   **Dashed:** Highly prevalent for indicating "Logical Stages," "Optional Paths," or "Scopes."

##### C. Lines & Arrows
*Design Philosophy: Line style dictates flow type.*

**Connector Styles**
*   **Orthogonal / Elbow (Right Angles):** Most papers use this for **Network Architectures** (implies precision, matrices, and tensors).
*   **Curved / Bezier:** Common choices include this for **System Logic, Feedback Loops, or High-Level Data Flow** (implies narrative and connection).

**Line Semantics**
*   **Solid Black/Grey:** Standard data flow (Forward pass).
*   **Dashed Lines:** Universally recognized as "Auxiliary Flow."
    *   *Used for:* Gradient updates, Skip connections, or Loss calculations.
*   **Integrated Math:** Standard operators (⊕ for Add, ⊗ for Concat/Multiply) are frequently placed *directly* on the line or intersection.

##### D. Typography & Icons
*Design Philosophy: Strict separation between "Labeling" and "Math."*

**Typography**
*   **Labels (Module Names):** **Sans-Serif** (Arial, Roboto, Helvetica).
    *   *Style:* Bold for headers, Regular for details.
*   **Variables (Math):** **Serif** (Times New Roman, LaTeX default).
    *   *Rule:* If it is a variable in your equation (e.g., x, θ, L), it **must** be Serif and Italicized in the diagram.

**Iconography Options**
*   **For Model State:**
    *   *Trainable:* Fire, Lightning icons.
    *   *Frozen:* Snowflake, Padlock, Stop Sign (Greyed out).
*   **For Operations:**
    *   *Inspection:* Magnifying Glass.
    *   *Processing/Computation:* Gear, Monitor.
*   **For Content:**
    *   *Text/Prompt:* Document, Chat Bubble icons.
    *   *Image:* Actual thumbnail of an image (not just a square).

#### 3. Common Pitfalls (How to look "Amateur")
*   **The "PowerPoint Default" Look:** Using standard Blue/Orange presets with heavy black outlines.
*   **Font Mixing:** Using Times New Roman for "Encoder" labels (makes the paper look dated to the 1990s).
*   **Inconsistent Dimension:** Mixing flat 2D boxes and 3D isometric cubes without a clear reason (e.g., 2D for logic, 3D for tensors is fine; random mixing is not).
*   **Primary Backgrounds:** Using saturated Yellow or Blue backgrounds for grouping (distracts from the content).
*   **Ambiguous Arrows:** Using the same line style for "Data Flow" and "Gradient Flow."

#### 4. Domain-Specific Styles

**If you are writing an AGENT / LLM Paper:**
*   **Vibe:** Illustrative, Narrative, "Friendly.", Cartoony.
*   **Key Elements:** Use "User Interface" aesthetics. Chat bubbles for prompts, document icons for retrieval.
*   **Characters:** It is common to use cute 2D vector robots, human avatars, or emojis to humanize the agent's reasoning steps.

**If you are writing a COMPUTER VISION / 3D Paper:**
*   **Vibe:** Spatial, Dense, Geometric.
*   **Key Elements:** Frustums (camera cones), Ray lines, and Point Clouds.
*   **Color:** Often uses RGB color coding to denote axes or channel correspondence. Use heatmaps (Rainbow/Viridis) to show activation.

**If you are writing a THEORETICAL / OPTIMIZATION Paper:**
*   **Vibe:** Minimalist, Abstract, "Textbook."
*   **Key Elements:** Focus on graph nodes (circles) and manifolds (planes/surfaces).
*   **Color:** Restrained. mostly Grayscale/Black/White with one highlight color (e.g., Gold or Blue). Avoid "cartoony" elements.

---

### Plot Style Guide

#### 1. The "NeurIPS Look": A High-Level Overview
The prevailing aesthetic for 2025 is defined by **precision, accessibility, and high contrast**. The "default" academic look has shifted away from bare-bones styling toward a more graphic, publication-ready presentation.

*   **Vibe:** Professional, clean, and information-dense.
*   **Backgrounds:** There is a heavy bias toward **stark white backgrounds** for maximum contrast in print and PDF reading, though the "Seaborn-style" light grey background remains an accepted variant.
*   **Accessibility:** A strong emphasis on distinguishing data not just by color, but by texture (patterns) and shape (markers) to support black-and-white printing and colorblind readers.

#### 2. Detailed Style Options

##### Color Palettes
*   **Categorical Data:**
    *   **Soft Pastels:** Matte, low-saturation colors (salmon, sky blue, mint, lavender) are frequently used to prevent visual fatigue.
    *   **Muted Earth Tones:** "Academic" palettes using olive, beige, slate grey, and navy.
    *   **High-Contrast Primaries:** Used sparingly when categories must be distinct (e.g., deep orange vs. vivid purple).
    *   **Accessibility Mode:** A growing trend involves combining color with **geometric patterns** (hatches, dots, stripes) to differentiate categories.
*   **Sequential & Heatmaps:**
    *   **Perceptually Uniform:** "Viridis" (blue-to-yellow) and "Magma/Plasma" (purple-to-orange) are the standard.
    *   **Diverging:** "Coolwarm" (blue-to-red) is used for positive/negative value splits.
    *   **Avoid:** The traditional "Jet/Rainbow" scale is almost entirely absent.

##### Axes & Grids
*   **Grid Style:**
    *   **Visibility:** Grid lines are almost rarely solid. Common choices include **fine dashed (--)** or **dotted (:)** lines in light gray.
    *   **Placement:** Grids are consistently rendered *behind* data elements (low Z-order).
*   **Spines (Borders):**
    *   **The "Boxed" Look:** A full enclosure (black spines on all 4 sides) is very common.
    *   **The "Open" Look:** Removing the top and right spines for a minimalist appearance.
*   **Ticks:**
    *   **Style:** Ticks are generally subtle, facing inward, or removed entirely in favor of grid alignment.

##### Layout & Typography
*   **Typography:**
    *   **Font Family:** Exclusively **Sans-Serif** (resembling Helvetica, Arial, or DejaVu Sans). Serif fonts are rarely used for labels.
    *   **Label Rotation:** X-axis labels are rotated **45 degrees** only when necessary to prevent overlap; otherwise, horizontal orientation is preferred.
*   **Legends:**
    *   **Internal Placement:** Floating the legend *inside* the plot area (top-left or top-right) to maximize the "data-ink ratio."
    *   **Top Horizontal:** Placing the legend in a single row above the plot title.
*   **Annotations:**
    *   **Direct Labeling:** Instead of forcing readers to reference a legend, text is often placed directly next to lines or on top of bars.

#### 3. Type-Specific Guidelines

##### Bar Charts & Histograms
*   **Borders:** Two distinct styles are accepted:
    *   **High-Definition:** Using **black outlines** around colored bars for a "comic-book" or high-contrast look.
    *   **Borderless:** Solid color fills with no outline (often used with light grey backgrounds).
*   **Grouping:** Bars are grouped tightly, with significant whitespace between categorical groups.
*   **Error Bars:** Consistently styled with **black, flat caps**.

##### Line Charts
*   **Markers:** A critical observation: Lines almost always include **geometric markers** (circles, squares, diamonds) at data points, rather than just being smooth strokes.
*   **Line Styles:** Use **dashed lines (--)** for theoretical limits, baselines, or secondary data, and **solid lines** for primary experimental data.
*   **Uncertainty:** Represented by semi-transparent **shaded bands** (confidence intervals) rather than simple vertical error bars.

##### Scatter Plots
*   **Shape Coding:** Use different marker shapes (e.g., circles vs. triangles) to encode a categorical dimension alongside color.
*   **Fills:** Markers are typically solid and fully opaque.
*   **3D Plots:** Depth is emphasized by drawing "walls" with grids or using drop-lines to the "floor" of the plot.

##### Heatmaps
*   **Aspect Ratio:** Cells are almost strictly **square**.
*   **Annotation:** Writing the exact value (in white or black text) **inside the cell** is highly preferred over relying solely on a color bar.
*   **Borders:** Cells are often borderless (smooth gradient look) or separated by very thin white lines.

##### Radar Charts
*   **Fills:** The polygon area uses **translucent fills** (alpha ~0.2) to show grid lines underneath.
*   **Perimeter:** The outer boundary is marked by a solid, darker line.

#### 4. Common Pitfalls (What to Avoid)
*   **The "Excel Default" Look:** Avoid heavy 3D effects on bars, shadow drops, or serif fonts (Times New Roman) on axes.
*   **The "Rainbow" Map:** Avoid the Jet/Rainbow colormap; it is considered outdated and perceptually misleading.
*   **Ambiguous Lines:** A line chart *without* markers can look ambiguous if data points are sparse; always add markers.
*   **Over-reliance on Color:** Failing to use patterns or shapes to distinguish groups makes the plot inaccessible to colorblind readers.
*   **Cluttered Grids:** Avoid solid black grid lines; they compete with the data. Always use light grey/dashed grids.`

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
4. Keep the enhanced plan concise and image-model-friendly
5. Avoid repeating the same requirement in multiple ways
6. Do not exceed roughly 500 words unless the source plan is unusually complex

Output the enhanced visualization plan:`, input.VisualIntent.Mode, input.Content)
}
