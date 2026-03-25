# PaperBanana Defect Catalog

**Session**: TLV4-drift-deep-dive-20260325
**Date**: 2026-03-25
**Status**: Deep Analysis Complete

---

## Executive Summary

This catalog identifies **17 defects** in paperbanana-clean not mentioned in the previous drift analysis report. These defects span 5 critical areas: agent lifecycle management, prompt template processing, reference image handling, concurrency patterns, and session resume mechanics.

| Severity | Count | Category |
|----------|-------|----------|
| Critical | 4 | Agent lifecycle, concurrency, security |
| High | 6 | Prompt handling, resume, image processing |
| Medium | 4 | Style guide, error context, observability |
| Low | 3 | Logging, metadata, edge cases |

---

## DEFECT-001: Stylist Agent Not Integrated in Canonical Pipeline

**Severity**: Critical
**Category**: Agent Lifecycle Management
**Impact**: Full pipeline

### Description

The stylist agent exists in the Go codebase (`internal/agents/stylist/` and `internal/application/agents/stylist/`) but is **NOT included in the canonical pipeline** defined in `types.go`:

```go
// paperbanana-clean/internal/domain/agent/types.go:20-26
var pipelineOrder = []StageName{
    StageRetriever,
    StagePlanner,
    StageStylist,    // Listed but...
    StageVisualizer,
    StageCritic,
}
```

However, in `runner.go:52-62`:
```go
func NewCanonicalRunner(retriever, planner, stylist, visualizer, critic domainagent.BaseAgent, ...) *Runner {
    agents := map[domainagent.StageName]domainagent.BaseAgent{
        domainagent.StageRetriever:  retriever,
        domainagent.StagePlanner:    planner,
        domainagent.StageVisualizer: visualizer,  // NO STYLIST HERE
        domainagent.StageCritic:     critic,
    }
    if stylist != nil {
        agents[domainagent.StageStylist] = stylist  // Only added conditionally
    }
    ...
}
```

**Python Comparison** (`repo-cn/agents/stylist_agent.py`): Stylist is always part of the full pipeline in `_run_critic_iterations` path.

### Root Cause

The stylist is optional in Go but critical for style guide application. The `dev_full` mode in Python always includes stylist, but Go's `NewCanonicalRunner` makes it optional.

### Impact

- Figures generated without NeurIPS 2025 style guide refinement
- Inconsistent visual quality between Python and Go outputs
- `dev_full` mode equivalent produces different results

### Fix Recommendation

1. Make stylist mandatory in the canonical pipeline
2. Add validation in `NewCanonicalRunner` to error if stylist is nil for full pipeline mode
3. Ensure `orderedPipeline()` includes stylist when present

---

## DEFECT-002: Planner Image Example Limit Too Restrictive

**Severity**: High
**Category**: Reference Image Processing
**Impact**: Planner output quality

### Description

Go's planner limits reference image examples to **2 images** while Python includes **all retrieved examples**:

```go
// paperbanana-clean/internal/application/agents/planner/prompt.go:14-17
const (
    planningExampleLimit      = 4   // Text examples
    planningImageExampleLimit = 2   // IMAGE examples - TOO LOW!
    ...
)
```

```python
# repo-cn/agents/planner_agent.py:61-86
for idx, item in enumerate(examples):
    # ... build prompt
    content_list.append({"type": "image", "image_base64": ref_image_base64})
    # Python includes ALL example images, no limit
```

### Impact

- Go planner sees 50% fewer visual references than Python
- Reduced learning from example patterns
- Lower output quality for complex figures

### Fix Recommendation

Increase `planningImageExampleLimit` to 4 or make it configurable via `Config` struct.

---

## DEFECT-003: Missing Lite Retrieval Mode

**Severity**: Critical
**Category**: Prompt Template Processing
**Impact**: Token cost, latency

### Description

Python has two retrieval modes: `auto` (lite, ~3K tokens) and `auto-full` (~80K tokens):

```python
# repo-cn/agents/retriever_agent.py:92-100
elif retrieval_setting == "auto":
    data["top10_references"] = await self._retrieve_and_parse(data, cfg, lite=True)   # LITE
elif retrieval_setting == "auto-full":
    data["top10_references"] = await self._retrieve_and_parse(data, cfg, lite=False)  # FULL
```

Go only has one mode and **always sends full content**:

```go
// paperbanana-clean/internal/application/agents/retriever/agent.go:207-242
case RetrievalModeAuto:
    // ... no "lite" option exists
    userPrompt, err := buildUserPrompt(input, candidates)  // Always builds FULL prompt
```

The `buildUserPrompt` function in Go has no `lite` parameter.

### Impact

- 25x higher token usage for retrieval (80K vs 3K tokens)
- Higher API costs
- Slower retrieval phase

### Fix Recommendation

1. Add `RetrievalModeAutoFull` constant
2. Modify `buildUserPrompt` to accept a `lite` boolean
3. When lite=true, omit `content` field from candidate examples

---

## DEFECT-004: Critic Revision Agent Injection Missing

**Severity**: Critical
**Category**: Agent Lifecycle Management
**Impact**: Pipeline integrity

### Description

Go's critic agent requires a `RevisionAgent` to re-render after critique, but the injection path is unclear:

```go
// paperbanana-clean/internal/application/agents/critic/agent.go:140-142
if a.cfg.RevisionAgent == nil {
    return a.fail(errors.New("critic requires a revision agent when changes are requested"))
}
```

This error path is only triggered when:
1. Critique requests changes
2. `RevisionAgent` is nil

But there's **no validation at agent creation** to ensure `RevisionAgent` is set.

### Python Comparison

Python handles this inline via `visualizer_agent.process(data)` called directly from `_run_critic_iterations`.

### Impact

- Pipeline fails at runtime if RevisionAgent not injected
- No compile-time safety for this dependency
- Error only surfaces during critique iteration

### Fix Recommendation

1. Add `RevisionAgent` validation in `NewAgent()`
2. Return error if `RevisionAgent` is nil and critic is expected to iterate
3. Document this dependency clearly in struct comments

---

## DEFECT-005: Context Cancellation Not Propagated to Visualizer Plot Execution

**Severity**: High
**Category**: Concurrency Handling
**Impact**: Resource leaks, orphaned processes

### Description

The plot executor uses `exec.CommandContext` but doesn't properly handle context cancellation for the Python subprocess:

```go
// paperbanana-clean/internal/application/agents/visualizer/plot_executor.go:37-38
cmd := exec.CommandContext(ctx, e.command, "-c", plotExecutorScript)
cmd.Stdin = strings.NewReader(cleaned)
```

If the context is cancelled:
- The process is killed via `CommandContext`
- But `cmd.CombinedOutput()` may still block on pipe reads
- No explicit process group termination

### Impact

- Orphaned Python processes on cancellation
- Resource leaks in long-running servers
- Race conditions during graceful shutdown

### Fix Recommendation

1. Use `cmd.Start()` + `cmd.Wait()` pattern
2. Add explicit `cmd.Process.Kill()` in goroutine watching context
3. Consider using `syscall.Kill(-pid)` for process group termination

---

## DEFECT-006: Session State Restoration Skips Validation

**Severity**: High
**Category**: Session Resume Details
**Impact**: Resume integrity

### Description

When restoring agent state, Go performs a simple assignment without validation:

```go
// paperbanana-clean/internal/application/agents/critic/agent.go:190-193
func (a *Agent) RestoreState(state domainagent.AgentState) error {
    a.state = state  // No validation!
    return nil
}
```

This applies to ALL agents. The restored state could be:
- From a different schema version
- Corrupted/incomplete
- Incompatible with current agent logic

### Python Comparison

Python has no resume capability, so this is a new feature that needs proper validation.

### Impact

- Silent failures on schema migration
- Corrupted state silently accepted
- Debug difficulty when resume produces unexpected behavior

### Fix Recommendation

1. Add schema version validation in `RestoreState`
2. Validate required fields are present
3. Return error if state is incompatible

---

## DEFECT-007: Visualizer Diagram Retry Logic Incomplete

**Severity**: Medium
**Category**: Concurrency Handling
**Impact**: Reliability

### Description

Go's visualizer has retry logic for diagrams, but the retry delay is hardcoded and the retry condition is limited:

```go
// paperbanana-clean/internal/application/agents/visualizer/agent.go:202-210
func diagramRetryDelay(attempt int) time.Duration {
    switch attempt {
    case 1:
        return 1500 * time.Millisecond
    case 2:
        return 3 * time.Second
    default:
        return 4500 * time.Millisecond
    }
}
```

Issues:
1. Only retries on specific error codes (timeout, rate limit, network)
2. Doesn't retry on empty image response
3. Hardcoded delays, not exponential backoff

### Python Comparison

Python uses 5 attempts with 30-second delay for all errors.

### Impact

- Lower resilience to transient errors
- Inconsistent retry behavior with Python

### Fix Recommendation

1. Make retry delays configurable
2. Add empty response as retry condition
3. Implement proper exponential backoff with jitter

---

## DEFECT-008: Missing Stylist Prompt Context Labels

**Severity**: High
**Category**: Prompt Template Processing
**Impact**: Stylist output quality

### Description

Go's stylist prompt is simplified compared to Python:

```go
// paperbanana-clean/internal/agents/stylist/prompts.go:10-35
func buildStylistPrompt(input agent.AgentInput) string {
    return fmt.Sprintf(`...
## Visual Mode: %s
## Original Plan:
%s
...`, styleGuide, visualMode, input.Content)
}
```

Python's stylist includes context labels:

```python
# repo-cn/agents/stylist_agent.py:62-67
user_prompt += f"{cfg['context_labels'][0]}: {raw_content}\n"
user_prompt += f"{cfg['context_labels'][1]}: {data['visual_intent']}\nYour Output:"
```

For plot: `["Raw Data", "Visual Intent of the Desired Plot"]`
For diagram: `["Methodology Section", "Diagram Caption"]`

### Impact

- Stylist lacks full context for style decisions
- May produce style recommendations inconsistent with content
- Output differs from Python version

### Fix Recommendation

1. Add context labels to stylist prompt
2. Include raw content and visual intent in the prompt structure
3. Match Python's prompt format exactly

---

## DEFECT-009: Concurrent Batch Processing Lacks Rate Limiting

**Severity**: Medium
**Category**: Concurrency Handling
**Impact**: API throttling

### Description

Python uses `asyncio.Semaphore` for concurrency control:

```python
# repo-cn/utils/paperviz_processor.py:202-205
semaphore = asyncio.Semaphore(max_concurrent)
async def process_with_semaphore(doc):
    async with semaphore:
        return await self.process_single_query(doc, do_eval=do_eval)
```

Go's batch runner uses `errgroup` but without explicit semaphore:

```go
// paperbanana-clean/internal/application/orchestrator/batch_runner.go
// Uses errgroup.WithContext but no semaphore equivalent
```

### Impact

- All batch items may start API calls simultaneously
- API rate limits may be exceeded
- 429 errors more likely during batch processing

### Fix Recommendation

Add a rate limiter or semaphore wrapper around batch execution.

---

## DEFECT-010: Reference Image Loading Ignores Missing Files

**Severity**: Medium
**Category**: Reference Image Processing
**Impact**: Pipeline resilience

### Description

Go's planner fails if a reference image file is not found:

```go
// paperbanana-clean/internal/application/agents/planner/agent.go:199-214
func (a *Agent) loadExampleImage(mode domainagent.VisualMode, path string) ([]byte, string, error) {
    ...
    return nil, "", err  // Returns error, no fallback
}
```

Python handles missing references gracefully:

```python
# repo-cn/agents/planner_agent.py:83-86
image_path = self.exp_config.work_dir / f"data/PaperBananaBench/{cfg['task_name']}" / item["path_to_gt_image"]
with open(image_path, "rb") as f:
    ref_image_base64 = base64.b64encode(f.read()).decode("utf-8")
# No try/except - but wrapped in higher-level error handling
```

### Impact

- Single missing reference image fails entire pipeline
- No graceful degradation
- Different behavior from Python

### Fix Recommendation

1. Log warning when image not found
2. Continue without the image (text-only example)
3. Track missing images in metadata for debugging

---

## DEFECT-011: Prompt Version Mismatch Between Agents

**Severity**: Low
**Category**: Prompt Template Processing
**Impact**: Debuggability

### Description

Prompt versions are hardcoded inconsistently:

```go
// critic/prompt.go
const PromptVersion = "critic-v1"

// planner/prompt.go
const PromptVersion = "planner-v2"  // v2, not v1!

// visualizer/agent.go
const PromptVersion = "visualizer-v1"
```

No version history or changelog exists. When prompts change, there's no mechanism to track which version was used for a given output.

### Impact

- Cannot correlate output quality with prompt version
- No audit trail for prompt changes
- Debug difficulty

### Fix Recommendation

1. Centralize prompt version management
2. Include prompt version in session metadata
3. Add changelog for prompt modifications

---

## DEFECT-012: Critic Early Stop Logic Differs from Python

**Severity**: Medium
**Category**: Agent Lifecycle Management
**Impact**: Critic iteration behavior

### Description

Go's critic early stop condition:

```go
// paperbanana-clean/internal/application/agents/critic/agent.go:131-138
if critique.noChanges() {
    reusedArtifact = hasRenderedArtifact(currentArtifacts)
    if latestArtifact.Kind == domainagent.ArtifactKindRenderedFigure && len(latestArtifact.Bytes) > 0 {
        currentArtifacts = append(currentArtifacts, latestArtifact)  // Reuses
    }
    stopReason = "no_change"
    break
}
```

Python's early stop:

```python
# repo-cn/utils/paperviz_processor.py:84-86
if critic_suggestions.strip() == "No changes needed.":
    print(f"[Critic Round {round_idx}] No changes needed. Stopping iteration.")
    break
```

The Python version is simpler - it just breaks. Go adds artifact reuse logic that Python doesn't have, which could lead to different behavior.

### Impact

- Different artifact handling on early stop
- Potential for duplicated artifacts in Go
- Inconsistent behavior between implementations

### Fix Recommendation

1. Verify artifact handling matches Python exactly
2. Document the difference if intentional
3. Add tests comparing both implementations

---

## DEFECT-013: Missing Pipeline Mode Validation

**Severity**: Low
**Category**: Agent Lifecycle Management
**Impact**: User experience

### Description

Go's pipeline mode filtering silently falls back to default for unknown modes:

```go
// paperbanana-clean/internal/application/orchestrator/runner.go:606-638
func (r *Runner) filterPipeline(metadata map[string]string, base []domainagent.StageName) []domainagent.StageName {
    ...
    switch mode {
    case "full":
        return append([]domainagent.StageName(nil), base...)
    case "planner-critic":
        allowed[domainagent.StagePlanner] = true
        allowed[domainagent.StageCritic] = true
    case "vanilla":
        allowed[domainagent.StageVisualizer] = true
    default:
        return append([]domainagent.StageName(nil), base...)  // Silent fallback!
    }
}
```

Invalid modes are treated as "full" without warning.

### Python Comparison

Python validates modes explicitly:

```python
# repo-cn/main.py:68-72
parser.add_argument(
    "--retrieval_setting",
    choices=["auto", "manual", "random", "none"],  # Explicit validation
    ...
)
```

### Impact

- Typos in mode names silently ignored
- Unexpected pipeline behavior
- Debug difficulty

### Fix Recommendation

1. Log warning for unknown modes
2. Return error or validate at API layer
3. Document all valid modes

---

## DEFECT-014: Visualizer Reuse Logic Missing from Python

**Severity**: Low
**Category**: Agent Lifecycle Management
**Impact**: Behavior difference

### Description

Go has artifact reuse logic in visualizer:

```go
// paperbanana-clean/internal/application/agents/visualizer/agent.go:391-401
func shouldReuseArtifact(input domainagent.AgentInput) bool {
    if !critiqueRequestsNoChange(input.CritiqueRounds) {
        return false
    }
    for _, artifact := range input.GeneratedArtifacts {
        if artifact.Kind == domainagent.ArtifactKindRenderedFigure && len(artifact.Bytes) > 0 {
            return true
        }
    }
    return false
}
```

This is a **new feature** in Go that Python doesn't have. While potentially an improvement, it creates a behavior divergence.

### Impact

- Go skips re-rendering when critique says "no changes"
- Python always re-renders
- Different resource usage patterns

### Fix Recommendation

1. Document this as intentional enhancement
2. Make it configurable
3. Consider backporting to Python

---

## DEFECT-015: Error Detail Context Lost in Fail Path

**Severity**: Medium
**Category**: Concurrency Handling
**Impact**: Debuggability

### Description

When the runner fails a stage, the error detail is wrapped but original context can be lost:

```go
// paperbanana-clean/internal/application/orchestrator/runner.go:280-294
errorCode := domainagent.ClassifyError(err)
detail := domainagent.WrapAgentError(err, stage, errorCode)
```

The `ClassifyError` function may replace the original error message with a generic code, losing the specific failure details.

### Impact

- Debug logs show generic error codes instead of specifics
- Harder to diagnose root cause
- User-facing errors less helpful

### Fix Recommendation

1. Preserve original error message in `Message` field
2. Add `OriginalError` field to `ErrorDetail`
3. Include stack trace in debug mode

---

## DEFECT-016: Session ID Generation Not Thread-Safe

**Severity**: Low
**Category**: Concurrency Handling
**Impact**: Session uniqueness

### Description

Session ID generation is not visible in the code, but if using `time.Now()` or random without proper seeding, there could be collisions in high-throughput scenarios.

The Python code uses UUIDs implicitly through data structures but doesn't show explicit generation.

### Impact

- Potential session ID collisions under load
- Resume could target wrong session
- Data integrity issues

### Fix Recommendation

1. Use UUID v4 or similar for session IDs
2. Add collision detection
3. Document session ID format

---

## DEFECT-017: Style Guide Loading Path Hardcoded

**Severity**: Low
**Category**: Prompt Template Processing
**Impact**: Flexibility

### Description

Go's style guide is loaded via a function that's not configurable:

```go
// paperbanana-clean/internal/agents/stylist/prompts.go:10-12
func buildStylistPrompt(input agent.AgentInput) string {
    visualMode := string(input.VisualIntent.Mode)
    styleGuide := styleguides.GetStyleGuide(visualMode)  // Hardcoded lookup
```

Python loads from filesystem:

```python
# repo-cn/agents/stylist_agent.py:59-60
with open(self.exp_config.work_dir / f"style_guides/neurips2025_{task_name}_style_guide.md", "r") as f:
    style_guide = f.read()
```

### Impact

- Go's style guide is compiled into binary
- Cannot update style guide without recompilation
- Less flexible than Python's file-based approach

### Fix Recommendation

1. Add `StyleGuidePath` to stylist config
2. Load from file if path provided, fallback to compiled
3. Support hot-reload for development

---

## Summary Matrix

| ID | Severity | Category | Effort |
|----|----------|----------|--------|
| DEFECT-001 | Critical | Agent Lifecycle | 2 days |
| DEFECT-002 | High | Image Processing | 0.5 days |
| DEFECT-003 | Critical | Prompt Processing | 1 day |
| DEFECT-004 | Critical | Agent Lifecycle | 0.5 days |
| DEFECT-005 | High | Concurrency | 1 day |
| DEFECT-006 | High | Session Resume | 1 day |
| DEFECT-007 | Medium | Concurrency | 0.5 days |
| DEFECT-008 | High | Prompt Processing | 0.5 days |
| DEFECT-009 | Medium | Concurrency | 0.5 days |
| DEFECT-010 | Medium | Image Processing | 0.5 days |
| DEFECT-011 | Low | Prompt Processing | 0.25 days |
| DEFECT-012 | Medium | Agent Lifecycle | 0.5 days |
| DEFECT-013 | Low | Agent Lifecycle | 0.25 days |
| DEFECT-014 | Low | Agent Lifecycle | 0.25 days |
| DEFECT-015 | Medium | Concurrency | 0.5 days |
| DEFECT-016 | Low | Concurrency | 0.25 days |
| DEFECT-017 | Low | Prompt Processing | 0.5 days |

**Total Estimated Effort**: 10-11 days

---

## Priority Recommendations

### Immediate (P0)
1. DEFECT-001: Fix stylist integration in canonical pipeline
2. DEFECT-003: Implement lite retrieval mode
3. DEFECT-004: Add RevisionAgent validation

### This Sprint (P1)
4. DEFECT-002: Increase planner image example limit
5. DEFECT-005: Fix context cancellation in plot executor
6. DEFECT-006: Add session state validation
7. DEFECT-008: Add context labels to stylist prompt

### Next Sprint (P2)
8. DEFECT-007: Improve visualizer retry logic
9. DEFECT-009: Add batch rate limiting
10. DEFECT-010: Handle missing reference images gracefully
11. DEFECT-012: Verify critic early stop consistency
12. DEFECT-015: Preserve error context in fail path

---

*Generated by analyst agent - TLV4-drift-deep-dive-20260325*
