# Security Audit Learnings

## Codebase Security Posture

### Strengths Discovered
1. **Parameterized SQL**: All database queries use GORM parameterized queries - no SQL injection vectors found
2. **Path Traversal Protection**: Robust implementation using `filepath.IsLocal()` with double-check
3. **API Key Encryption**: AES-256-GCM with Argon2id key derivation - cryptographic best practices
4. **Constant-Time Auth**: API key comparison uses `subtle.ConstantTimeCompare`

### Weaknesses Discovered
1. **Python exec()**: Direct code execution without sandboxing - critical vulnerability pattern
2. **Hardcoded Paths**: Python interpreter path hardcoded, reducing deployment flexibility
3. **Missing Headers**: Security headers not implemented in router

## OWASP Alignment Analysis

- **Strong**: A02 (Cryptographic), A07 (Auth), A03 (SQL Injection)
- **Weak**: A03 (Code Injection), A05 (Security Headers), A01 (CORS)

## Tool Patterns for Future Audits

1. Use `Grep` for pattern-based vulnerability scanning
2. Read multiple related files to understand context
3. Check both implementation and configuration
4. Verify encryption key derivation parameters (memory, iterations, parallelism)

---

# Batch Persistence Analysis Learnings

## Architecture Split Between Session and Batch
- Individual sessions ARE persisted via `SnapshotStore` and `SessionRepository`
- Batch aggregates are NOT persisted - they remain in `BatchRunner.results` map (in-memory)
- This is an architectural gap, not a missing feature - the infrastructure exists

## Existing Infrastructure Reuse Potential
- `SessionRepository` already has full JSON snapshot capability
- `HistoryService` provides transactional session persistence
- Adding batch persistence is an incremental change, not a rewrite

## Comparison with Python Reference (repo-cn)
- repo-cn uses file-based JSON persistence with incremental saves (every 10 samples)
- paperbanana-clean has more sophisticated session persistence but lacks batch aggregation
- The gap is specifically at the batch aggregation layer, not the session layer

## Implementation Notes
1. The `BatchRunner` is the central orchestration point for batch persistence changes
2. `storeResult()` currently only updates memory - needs persistence hook
3. `GetBatchResult()` needs a fallback to database/secondary store
4. SSE streaming works well - no changes needed for progress tracking

## Technical Debt Identified
- Missing `BatchResultModel` in database schema
- No cleanup mechanism for old batch results
- No batch-to-session mapping table for reconstruction

---

# Deep Defect Analysis Learnings

## Key Architectural Findings

### Agent Integration Gaps
1. **Stylist Agent**: Exists in Go code but is optional in canonical pipeline, unlike Python where it's mandatory for full mode
2. **Revision Agent**: Critic requires RevisionAgent for re-rendering but no compile-time validation exists

### Retrieval Efficiency
1. **Lite Mode Missing**: Python's "lite" retrieval mode saves 96% tokens (3K vs 80K). Go lacks this optimization entirely
2. **Image Example Limit**: Go's planner caps image examples at 2, Python has no limit - affects few-shot learning quality

### Concurrency & Lifecycle
1. **Context Propagation**: Go's plot executor doesn't properly handle process cleanup on context cancellation - risk of orphaned processes
2. **Resume Validation**: Go's `RestoreState` is a simple assignment with no validation - could cause silent failures

## Architectural Insights

### What Go Does Better
1. **Artifact Reuse**: Go added intelligent artifact reuse when critique says "no changes" - Python lacks this
2. **Error Classification**: Go has sophisticated error classification (transient vs permanent) with retry strategies
3. **Stage Timeouts**: Go has per-stage timeout configuration - Python has no timeout mechanism
4. **Event Streaming**: Go's RunHandle event channel is cleaner than Python's AsyncGenerator pattern

### What Python Does Better
1. **Retrieval Modes**: Python has `auto` (lite) and `auto-full` options - Go only has one mode
2. **Reference Image Limits**: Python includes all retrieved images - Go limits to 2
3. **Style Guide Flexibility**: Python loads from filesystem - Go has compiled-in guides

## Code Patterns to Preserve

1. **Immutable Clones**: Go's clone pattern for agent state is valuable for resume capability
2. **Stage Timeouts**: Per-agent timeout configuration is a good safety mechanism
3. **Event-Based Progress**: RunHandle pattern allows real-time progress tracking

## Tool Patterns for Deep Dives

1. Read corresponding agent files side-by-side (Python vs Go)
2. Compare prompt construction logic explicitly
3. Check constant/limit values in each codebase
4. Verify error handling paths match expected behavior
5. Look for validation gaps in state transitions

---

# Image Display Flow Analysis Learnings

## JSON Field Name Mismatches Are Silent Failures
**Context**: Backend uses `bytes` for base64 image data, frontend expects `data`.

**Key Insight**: Gin's JSON serializer converts `[]byte` to base64 automatically, but the field name mismatch means frontend never receives the image data. This is a silent failure - no error is thrown, the image simply doesn't display.

**Prevention**: Create shared TypeScript types from Go structs, or use code generation.

## Asset Persistence Is Optional But Required For Display
**Context**: Visualizer creates artifacts with `memory://` URIs and raw bytes, but no persistence step exists.

**Key Insight**: The full asset storage infrastructure exists (`localstore`, `AssetService`, asset endpoints) but is never connected to the generation pipeline. Images live only in memory during the request lifecycle.

**Implication**: For persistent/retrievable images, a persistence step must be added after visualizer execution.

## API Parameter Count Mismatches
**Context**: Asset endpoint requires `project_id` and `asset_id`, frontend only provides single `assetId`.

**Key Insight**: The URL pattern `/api/v1/assets/${artifact.assetId}` is missing the required `project_id` parameter. This would cause a 400/404 even if `assetId` were populated.

**Pattern**: Always verify route definitions match frontend URL construction.

## Batch Types Diverge From Single Generation
**Context**: `BatchArtifact` type has different fields than single generation artifact type.

**Key Insight**: Type definitions have drifted - single generation uses `data`/`assetId` pattern, batch uses `id`/`kind`/`mimeType` only. This inconsistency suggests missing type sharing.

## MIME Type Chain Is Correct
**Context**: Verified MIME type propagation from LLM response through artifact creation.

**Key Insight**: The MIME type flow is correctly implemented:
1. LLM returns `part.MIMEType` or defaults to `image/png`
2. Artifact stores `MIMEType`
3. Asset store can persist to `.mime` sidecar file
4. Frontend receives and uses `mimeType`

This is one part of the flow that works correctly.

---

# Data Dependencies Deep Dive Learnings

## Empty JSON Files != Missing Files
**Context**: `ref.json` files exist but contain `[]` (empty arrays).

**Key Insight**: The code checks for `os.ErrNotExist` to handle missing files, but empty JSON files parse successfully as empty slices. This bypasses the fallback logic and returns silently with no data.

**Impact**: Few-shot learning degrades to zero-shot without any warning to the user.

## Data Flow Dependencies Are Not Always Obvious
**Context**: Retriever reads `ref.json`, Planner loads images based on `path_to_gt_image`.

**Key Insight**: The data dependency chain spans multiple agents:
1. Retriever loads JSON metadata (candidates)
2. Retriever passes IDs to Planner
3. Planner loads images from paths in JSON
4. Planner includes images in LLM prompt

If any step fails silently, quality degrades without visibility.

## Python Has Better Missing Data Visibility
**Context**: Python's `retriever_agent.py:66-68` explicitly warns and falls back.

**Key Insight**: The Python implementation checks `if not ref_file.exists()` and prints a warning before falling back to `none` mode. Go's implementation silently returns empty slices.

**Pattern**: When handling external dependencies, always provide user-visible feedback when the dependency is missing or empty.

## Lite Mode vs Full Mode Retrieval
**Context**: Python has `auto` (lite, ~3K tokens) and `auto-full` (~80K tokens) modes.

**Key Insight**: Lite mode retrieves only IDs, not full content. This reduces token usage by 96% while still providing retrieval benefits. Go only has one mode equivalent to Python's `auto-full`.

**Optimization**: Implementing lite mode in Go would significantly reduce costs.

## Image Example Limits Affect Quality
**Context**: Go limits planner to 2 reference images, Python has no limit.

**Key Insight**: Few-shot learning benefits from more examples. The arbitrary limit of 2 images in Go may reduce quality for complex diagrams.

**Trade-off**: More examples = higher token cost, but potentially better quality.

## .gitignore Patterns Can Mask Data Gaps
**Context**: Images are in .gitignore, but empty JSON placeholders were committed.

**Key Insight**: The .gitignore excludes `images/` directories and `test.json`, but not `ref.json`. Someone committed empty `ref.json` files as placeholders without documenting this.

**Prevention**: Either exclude all benchmark data or include a download script. Placeholder files without documentation cause confusion.

---

# UI/UX Review Learnings

## Feature Existence vs Discoverability
**Context**: Header component exists but lacks History and Settings buttons.

**Key Insight**: Features implemented in parent components (App.tsx has History panel, Settings drawer) are useless if users cannot discover them. The disconnect between props passed (`onHistoryClick`, `onSettingsClick`) and props used (Header only uses theme/language) represents a gap in component integration.

**Pattern**: Always audit component props vs usage to find missing integrations.

## Custom Events Require Listeners
**Context**: EmptyState dispatches `workspace:loadExample` event but no one listens.

**Key Insight**: Using CustomEvent for component communication requires both:
1. Event dispatch (done in EmptyState)
2. Event listener (missing in App.tsx/Workspace.tsx)

**Alternative**: Callback props are more explicit and easier to trace than custom events.

## Error Messages Need Actionability
**Context**: `HTTP 401` vs `API Key invalid, please check settings`.

**Key Insight**: Users cannot recover from errors they don't understand. Raw HTTP codes provide no guidance. Every error should include:
1. What went wrong
2. Why it matters
3. What to do next

## Token Cost Transparency
**Context**: Streamlit shows "auto-full: ~80K tokens per candidate", React UI shows nothing.

**Key Insight**: Token costs are a critical UX concern for AI products. Users need to understand cost implications before submitting requests. The React UI is missing this entirely.

**Pattern**: For metered/cost-based features, always display cost estimates prominently.

## Streamlit vs React Trade-offs
**Context**: Comparing Streamlit demo.py with React web app.

**Key Insight**: Each framework has strengths:
- **Streamlit**: Fast iteration, built-in state management, clear configuration layout, cost transparency
- **React**: Responsive design, rich interactions, theme support, modern UX patterns

**Gap Analysis**:
- Streamlit has: Token warnings, ZIP download, detailed evolution timeline
- React has: History persistence, theme switching, responsive layout, candidate comparison grid

**Action**: Consider porting Streamlit's strengths (warnings, downloads, timeline) to React without losing React's advantages.

## Progress Visibility Matters
**Context**: Generate mode has stage progress, Refine mode only has spinner.

**Key Insight**: Multi-step processes need visible progress indicators. The 5-stage pipeline (Retriever, Planner, Stylist, Visualizer, Critic) is well visualized, but Refinement is a black box.

**Pattern**: Any process taking more than a few seconds should show intermediate progress.
