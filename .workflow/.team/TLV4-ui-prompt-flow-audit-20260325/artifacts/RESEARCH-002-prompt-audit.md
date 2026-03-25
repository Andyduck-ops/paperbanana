# RESEARCH-002: Prompt Template Audit Report

**Session**: TLV4-ui-prompt-flow-audit-20260325
**Role**: prompt-auditor (analyst)
**Generated**: 2026-03-25

---

## Executive Summary

This audit evaluates the prompt templates across the 5-agent pipeline (Retriever, Planner, Stylist, Visualizer, Critic) in the Go implementation (`paperbanana-clean`) against the Python reference implementation (`repo-cn`). The audit covers prompt content, few-shot examples, language consistency, and token efficiency.

**Overall Assessment**: The Go implementation faithfully preserves the core prompt logic from Python, but has several notable differences in style guide handling, system prompt structure, and user prompt construction.

---

## 1. Agent-by-Agent Prompt Analysis

### 1.1 Retriever Agent

| Aspect | Go Implementation | Python Reference | Assessment |
|--------|-------------------|------------------|------------|
| **System Prompt** | Identical to Python | Same content | PASS - 100% match |
| **Version** | `retriever-v1` | No versioning | Improvement in Go |
| **Candidate Limit** | 12 (auto), 220K char budget | 200 (diagram), no budget | Go has better limits |
| **Output Key** | `top10_diagrams` / `top10_plots` | `top10_references` | Minor naming difference |
| **Lite Mode** | No equivalent | Yes (`lite=True/False`) | Missing feature in Go |

**System Prompt Comparison (Diagram Mode)**:
```
Go:     diagramSystemPrompt (lines 282-346)
Python: DIAGRAM_RETRIEVER_AGENT_SYSTEM_PROMPT (lines 207-273)
Status: IDENTICAL - Character-for-character match
```

**Key Difference - Scoring Algorithm**:
- Go adds sophisticated pre-filtering: `shortlistCandidates()` with token overlap scoring (lines 116-182)
- Python relies purely on LLM selection without pre-scoring
- Go's approach reduces token consumption by limiting candidate pool

**Issue Found**:
- Go uses `top10_diagrams` and `top10_plots` as output keys
- Python uses `top10_references` for both modes
- This could cause integration issues if downstream expects Python naming

---

### 1.2 Planner Agent

| Aspect | Go Implementation | Python Reference | Assessment |
|--------|-------------------|------------------|------------|
| **System Prompt** | Identical | Same content | PASS |
| **Version** | `planner-v2` | No versioning | Improvement in Go |
| **Example Limit** | 4 examples, 2 images | No explicit limit | Go has safeguards |
| **Char Limits** | 3000 (target), 1600 (example) | No explicit limits | Good practice |
| **Image Loading** | `loadImage()` callback | Inline base64 | Better abstraction |

**System Prompt Comparison (Diagram Mode)**:
```
Go:     diagramSystemPrompt (lines 182-189)
Python: DIAGRAM_PLANNER_AGENT_SYSTEM_PROMPT (lines 136-143)
Status: IDENTICAL - Both use the same prompt structure
```

**User Prompt Construction Difference**:

Go approach (lines 74-104):
```go
// Builds example prompts + final prompt in structured format
buildExamplePrompt() -> "Example N:\nContent: ...\nIntent: ...\nReference: "
buildFinalPrompt() -> "Now, based on the following... provide a detailed description..."
```

Python approach (lines 71-96):
```python
# Similar structure but concatenates differently
user_prompt += f"Example {idx+1}:\n"
user_prompt += f"{cfg['content_label']}: {item_content}\n"
# Then appends image as base64 directly
content_list.append({"type": "image", "image_base64": ref_image_base64})
```

**Assessment**: Go's approach is cleaner with better separation of concerns. Both produce equivalent prompts.

---

### 1.3 Stylist Agent

| Aspect | Go Implementation | Python Reference | Assessment |
|--------|-------------------|------------------|------------|
| **System Prompt** | Embedded style guide | External file reference | DIFFERENT |
| **Version** | `1.1.0` | No versioning | Improvement in Go |
| **Style Guide** | Inline constant (hardcoded) | File read at runtime | Trade-off |

**Critical Difference - Style Guide Handling**:

Go (lines 11-27):
```go
const styleGuideReference = `## NeurIPS 2025 Style Guidelines
### Diagram Style
- Use clean, minimalist layouts...
### Plot Style
- Axis labels with units...`
```

Python (lines 59-60):
```python
with open(self.exp_config.work_dir / f"style_guides/neurips2025_{task_name}_style_guide.md", "r") as f:
    style_guide = f.read()
```

**Comparison of Style Guide Content**:

| Aspect | Go Inline Guide | Python External File |
|--------|-----------------|---------------------|
| Diagram guidance | 6 bullet points | 105 lines (comprehensive) |
| Plot guidance | 6 bullet points | 90 lines (comprehensive) |
| Color palettes | Basic mention | Detailed hex codes, domain-specific |
| Common pitfalls | Not included | Included |
| Domain-specific styles | Not included | Agent, Vision, Theory sections |

**ISSUE**: Go's inline style guide is a SIGNIFICANTLY ABBREVIATED version (~50 words) vs Python's comprehensive guides (~2000 words each). This could affect output quality.

**System Prompt Structure Difference**:

Go (lines 30-37):
```go
systemInstruction := fmt.Sprintf(`You are a visualization style expert...
%s
## Your Task...`, styleGuideReference, mode)
```

Python (lines 104-128):
```python
DIAGRAM_STYLIST_AGENT_SYSTEM_PROMPT = """
## ROLE
You are a Lead Visual Designer for top-tier AI conferences (e.g., NeurIPS 2025).
## TASK
Our goal is to generate high-quality, publication-ready diagrams...
**Crucial Instructions:**
1. Preserve Semantic Content...
"""
```

**Assessment**: Python's system prompt is more detailed with explicit "Crucial Instructions" for handling existing descriptions. Go's version is simpler but may lack nuance.

---

### 1.4 Visualizer Agent

| Aspect | Go Implementation | Python Reference | Assessment |
|--------|-------------------|------------------|------------|
| **System Prompt** | Very brief (1 line) | Very brief (1 line) | MATCH |
| **Version** | `visualizer-v1` | No versioning | Improvement in Go |
| **Retry Logic** | 3 attempts with backoff | 5 attempts, 30s delay | Different approach |
| **Plot Execution** | External executor interface | Inline process pool | Better separation |

**System Prompt Comparison**:

Go (line 505-507):
```go
const diagramSystemPrompt = "You are an expert scientific diagram illustrator. Generate high-quality scientific diagrams based on user requests.\n"
const plotSystemPrompt = "You are an expert statistical plot illustrator. Write code to generate high-quality statistical plots based on user requests.\n"
```

Python (lines 237-239):
```python
DIAGRAM_VISUALIZER_AGENT_SYSTEM_PROMPT = """You are an expert scientific diagram illustrator. Generate high-quality scientific diagrams based on user requests."""
PLOT_VISUALIZER_AGENT_SYSTEM_PROMPT = """You are an expert statistical plot illustrator. Write code to generate high-quality statistical plots based on user requests."""
```

**Status**: IDENTICAL (modulo whitespace)

**User Prompt Difference**:

Go (lines 342-356):
```go
// Diagram mode:
fmt.Sprintf("Render an image based on the following detailed description: %s\n Note that do not include figure titles in the image. Diagram: ", input.Content)

// Plot mode:
fmt.Sprintf("Use python matplotlib to generate a statistical plot based on the following detailed description: %s\n Only provide the code without any explanations. Code:", input.Content)
```

Python (lines 76-88):
```python
# Diagram mode:
"Render an image based on the following detailed description: {desc}\n Note that do not include figure titles in the image. Diagram: "

# Plot mode:
"Use python matplotlib to generate a statistical plot based on the following detailed description: {desc}\n Only provide the code without any explanations. Code:"
```

**Status**: IDENTICAL format

---

### 1.5 Critic Agent

| Aspect | Go Implementation | Python Reference | Assessment |
|--------|-------------------|------------------|------------|
| **System Prompt** | Identical | Same content | PASS |
| **Version** | `critic-v1` | No versioning | Improvement in Go |
| **Missing Image Msg** | Mode-specific | Mode-specific | MATCH |

**System Prompt Comparison (Diagram Mode)**:

Go (lines 102-141):
```go
const diagramSystemPrompt = `
## ROLE
You are a Lead Visual Designer for top-tier AI conferences (e.g., NeurIPS 2025).
## TASK
Your task is to conduct a sanity check...
## CRITIQUE & REVISION RULES
1. Content...
2. Presentation...
## INPUT DATA...
## OUTPUT...`
```

Python (lines 157-196):
```python
DIAGRAM_CRITIC_AGENT_SYSTEM_PROMPT = """
## ROLE
You are a Lead Visual Designer for top-tier AI conferences (e.g., NeurIPS 2025).
## TASK
Your task is to conduct a sanity check...
...
"""
```

**Status**: IDENTICAL content

**Missing Image Message**:

Go:
```go
// Diagram:
"[SYSTEM NOTICE] The diagram image could not be generated based on the current description. Please inspect the description for missing labels, layout issues, or invalid instructions and provide a corrected revision."

// Plot:
"[SYSTEM NOTICE] The plot image could not be generated based on the current description (likely due to invalid code). Please check the description for errors (e.g., syntax issues, missing data) and provide a revised version."
```

Python: Identical messages (lines 92-96)

---

## 2. Few-Shot Example Handling

### 2.1 PaperBananaBench References

| Agent | Go Implementation | Python Reference |
|-------|-------------------|------------------|
| Retriever | Loads from JSON, uses ID-based selection | Same approach |
| Planner | Loads examples, limits to 4, loads up to 2 images | Same, no limits |
| Stylist | No few-shot examples | No few-shot examples |
| Critic | Uses generated image as visual reference | Same |
| Visualizer | No few-shot examples | No few-shot examples |

**Issue**: Go hardcodes `planningExampleLimit = 4` and `planningImageExampleLimit = 2`. Python has no such limits, potentially sending more examples (could be more expensive but potentially better quality).

### 2.2 Reference Image Loading

Go approach (planner/prompt.go lines 60-64):
```go
if idx >= planningImageExampleLimit || example.PathToGTImage == "" {
    continue
}
image, mimeType, err := loadImage(input.VisualIntent.Mode, example.PathToGTImage)
```

Python approach (planner_agent.py lines 83-86):
```python
image_path = self.exp_config.work_dir / f"data/PaperBananaBench/{cfg['task_name']}" / item["path_to_gt_image"]
with open(image_path, "rb") as f:
    ref_image_base64 = base64.b64encode(f.read()).decode("utf-8")
```

**Assessment**: Go's abstraction with `loadImage` callback is cleaner for testing and mocking.

---

## 3. Language Consistency Analysis

### 3.1 Language Distribution

| Component | Language | Notes |
|-----------|----------|-------|
| All System Prompts | English | Consistent |
| All User Prompts | English | Consistent |
| Debug Messages (Go) | English | Good |
| Debug Messages (Python) | Chinese mixed | e.g., "[DEBUG] [RetrieverAgent] 开始处理" |
| Error Messages (Go) | English | Consistent |
| Error Messages (Python) | English | Consistent |
| Style Guides | English | Consistent |

### 3.2 Terminology Consistency

| Term | Go | Python | Status |
|------|-----|--------|--------|
| Visual Intent | `VisualIntent` | `visual_intent` | Different casing (language convention) |
| Methodology Section | `Methodology Section` | Same | MATCH |
| Raw Data | `Raw Data` | Same | MATCH |
| Caption | `Caption` | Same | MATCH |
| Target Input | `Target Input` | Same | MATCH |
| Candidate Pool | `Candidate Pool` | Same | MATCH |

**Assessment**: Terminology is consistent between implementations.

---

## 4. Token Efficiency Analysis

### 4.1 Retriever Token Budget

| Metric | Go | Python |
|--------|-----|--------|
| Char Budget | 220,000 | No limit (can hit 800K+ chars in full mode) |
| Candidate Limit | 12 auto-selected | 200 (diagram) |
| Pre-scoring | Yes (token overlap + keyword) | No |

**Go Optimization**: The `shortlistCandidates()` function (lines 116-182) uses:
- Token overlap scoring (4x weight for intent)
- Visual keyword matching (6x weight)
- Character budget enforcement

This significantly reduces token consumption while maintaining relevance.

### 4.2 Planner Token Limits

| Metric | Go | Python |
|--------|-----|--------|
| Target Content Limit | 3,000 chars | No limit |
| Example Content Limit | 1,600 chars | No limit |
| Example Count Limit | 4 | No limit |
| Image Count Limit | 2 | All examples |

**Go Optimization**: Enforces limits to prevent token overflow. Python can potentially exceed context windows.

### 4.3 Stylist Prompt Size

| Metric | Go | Python |
|--------|-----|--------|
| Inline Style Guide | ~50 words | ~2000 words (file) |
| System Prompt | ~150 words | ~300 words |

**Trade-off**: Go is more token-efficient but potentially sacrifices quality. Python loads comprehensive style guides at runtime cost.

### 4.4 Visualizer Prompt

| Metric | Go | Python |
|--------|-----|--------|
| System Prompt | ~20 words | ~20 words |
| User Prompt Template | ~30 words | ~30 words |

**Assessment**: Both are minimal and efficient.

### 4.5 Critic Prompt

| Metric | Go | Python |
|--------|-----|--------|
| System Prompt | ~400 words | ~400 words |
| User Prompt | Variable | Variable |

**Assessment**: Identical efficiency.

---

## 5. Key Differences Summary

### 5.1 Critical Differences

1. **Style Guide Compression (HIGH PRIORITY)**
   - Go: ~50 word inline summary
   - Python: ~2000 word comprehensive guides
   - Impact: May affect diagram/plot aesthetic quality

2. **Lite Mode Missing in Retriever (MEDIUM PRIORITY)**
   - Python has `lite=True/False` toggle for sending methodology vs just captions
   - Go always sends truncated methodology (3K chars)
   - Impact: Different retrieval behavior, no high-precision option

3. **Output Key Naming (LOW PRIORITY)**
   - Go: `top10_diagrams` / `top10_plots`
   - Python: `top10_references`
   - Impact: API compatibility issues

### 5.2 Improvements in Go

1. **Version Tagging**: All agents have prompt versions (e.g., `retriever-v1`, `planner-v2`)
2. **Token Budgets**: Explicit limits prevent overflow
3. **Pre-scoring**: Retriever has sophisticated candidate scoring
4. **Retry Backoff**: Visualizer has exponential backoff for retries
5. **Abstraction**: Better separation of concerns (loadImage callback, PlotExecutor interface)

### 5.3 Missing Features in Go

1. **Comprehensive Style Guides**: Only summary included
2. **Lite Mode**: No equivalent to Python's `lite=True/False`
3. **Dynamic Style Guide Loading**: Hardcoded vs file-based

---

## 6. Recommendations

### 6.1 High Priority

1. **Expand Style Guide in Stylist Agent**
   - Replace inline `styleGuideReference` with full NeurIPS 2025 guidelines
   - Consider loading from file like Python for maintainability
   - Location: `internal/application/agents/stylist/prompts.go` lines 11-27

### 6.2 Medium Priority

2. **Add Lite Mode to Retriever**
   - Implement `RetrieveMode` enum: `ModeLite`, `ModeFull`
   - Allow skipping methodology content for faster retrieval
   - Location: `internal/application/agents/retriever/agent.go`

3. **Unify Output Key Naming**
   - Either use `top10_references` consistently or document the difference
   - Update downstream code if needed

### 6.3 Low Priority

4. **Make Token Limits Configurable**
   - Currently hardcoded in `planner/prompt.go`:
     - `planningExampleLimit = 4`
     - `planningImageExampleLimit = 2`
     - `planningTargetCharLimit = 3000`
   - Consider making these configurable via Config struct

5. **Add Prompt Template Versioning to Python**
   - Go has `PromptVersion` constants
   - Python lacks versioning, making prompt evolution tracking harder

---

## 7. Appendix: Prompt Version Matrix

| Agent | Go Version | Python Version |
|-------|------------|----------------|
| Retriever | `retriever-v1` | N/A |
| Planner | `planner-v2` | N/A |
| Stylist | `1.1.0` | N/A |
| Visualizer | `visualizer-v1` | N/A |
| Critic | `critic-v1` | N/A |

---

## 8. Files Analyzed

### Go Implementation
- `internal/application/agents/retriever/prompt.go` (414 lines)
- `internal/application/agents/planner/prompt.go` (199 lines)
- `internal/application/agents/stylist/prompts.go` (67 lines)
- `internal/application/agents/visualizer/agent.go` (508 lines)
- `internal/application/agents/critic/prompt.go` (186 lines)

### Python Reference
- `repo-cn/agents/retriever_agent.py` (343 lines)
- `repo-cn/agents/planner_agent.py` (153 lines)
- `repo-cn/agents/stylist_agent.py` (153 lines)
- `repo-cn/agents/visualizer_agent.py` (240 lines)
- `repo-cn/agents/critic_agent.py` (241 lines)
- `repo-cn/style_guides/neurips2025_diagram_style_guide.md` (105 lines)
- `repo-cn/style_guides/neurips2025_plot_style_guide.md` (90 lines)
- `repo-cn/prompts/diagram_eval_prompts.py` (209 lines)
- `repo-cn/prompts/plot_eval_prompts.py` (247 lines)

---

*Report generated by prompt-auditor agent (analyst role)*
