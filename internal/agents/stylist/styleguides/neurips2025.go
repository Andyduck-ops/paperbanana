package styleguides

// NeurIPS 2025 Style Guide for Academic Visualizations
// Based on NeurIPS formatting guidelines and best practices for scientific figures

const NeurIPS2025DiagramStyle = `## NeurIPS 2025 Diagram Style Guide

### Color Palette
- Primary: #1a73e8 (Blue), #4285f4 (Light Blue)
- Accent: #ff6d01 (Orange), #34a853 (Green), #ea4335 (Red)
- Neutral: #5f6368 (Gray), #e8eaed (Light Gray), #202124 (Dark Gray)
- Background: #ffffff (White) or #f8f9fa (Off-white)

### Typography
- Title: 14-16pt, bold, sans-serif (Helvetica, Arial, Inter)
- Labels: 10-12pt, regular, sans-serif
- Annotations: 8-10pt, italic for notes
- Consistent font throughout diagram

### Shape Styles
- Containers: Rounded rectangles (rx: 4-8px), 2px stroke, light fill
- Process boxes: Rectangles with slight rounding, 1-2px stroke
- Decision points: Diamonds with 1px stroke
- Arrows: Solid lines with arrowheads, 1.5-2px stroke
- Dashed lines: 4px dash, 2px gap for optional flows

### Layout Principles
- Clear visual hierarchy: top-to-bottom or left-to-right flow
- Consistent spacing: 20-30px between major components
- Alignment: Grid-based alignment of elements
- Whitespace: 15% minimum whitespace ratio
- Margins: 20px minimum from edge

### Component Sizing
- Minimum component size: 60x40px
- Text padding: 8-12px internal
- Icon size: 16-24px
- Arrow length: Proportional to distance, minimum 30px

### Accessibility
- Contrast ratio: 4.5:1 minimum for text
- Color-blind safe: Use patterns or labels in addition to color
- Alt text: Provide descriptive alt text for all diagrams
`

const NeurIPS2025PlotStyle = `## NeurIPS 2025 Plot Style Guide

### Color Palette
- Data series: Distinct, colorblind-friendly palette
  - #1a73e8 (Blue), #ea4335 (Red), #34a853 (Green)
  - #fbbc04 (Yellow), #9c27b0 (Purple), #00bcd4 (Cyan)
- Axes: #202124 (Dark Gray)
- Grid: #e0e0e0 (Light Gray), 0.5px
- Background: #ffffff (White)

### Typography
- Title: 14pt, bold, centered above plot
- Axis labels: 12pt, with units in parentheses
- Tick labels: 10pt
- Legend: 10pt, placed outside plot area
- Annotations: 9pt, positioned to avoid overlap

### Axes and Grid
- Axis line: 1px solid dark gray
- Tick marks: 4px inward, 1px stroke
- Grid lines: Light gray, 0.5px, dashed optional
- Axis range: Include zero where appropriate

### Data Representation
- Line plots: Solid lines, 1.5-2px stroke
- Scatter plots: Circles, 6-8px diameter
- Bar charts: Solid fill, 1px border
- Error bars: Caps 2x line width
- Markers: Consistent per series, varied shapes

### Legend
- Position: Right side or bottom, outside plot area
- Border: Optional light border
- Entries: Same visual style as plot elements
- Ordering: Match visual or logical order

### Multiple Subplots
- Shared axes: Label only outer axes
- Spacing: 0.2-0.3 inches between subplots
- Labels: (a), (b), (c) for reference
- Consistent scale where comparing

### Accessibility
- Line styles: Solid, dashed, dotted for differentiation
- Markers: Different shapes (circle, square, triangle)
- High contrast: Ensure visibility when printed B&W
- Alt text: Describe trends, not just "plot of X vs Y"
`

// GetStyleGuide returns the appropriate style guide based on visual mode
func GetStyleGuide(mode string) string {
	switch mode {
	case "diagram":
		return NeurIPS2025DiagramStyle
	case "plot":
		return NeurIPS2025PlotStyle
	default:
		return NeurIPS2025DiagramStyle + "\n\n" + NeurIPS2025PlotStyle
	}
}
