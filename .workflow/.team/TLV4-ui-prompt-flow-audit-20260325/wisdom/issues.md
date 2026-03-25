# Issues - UI Prompt Flow Audit Session

## 2026-03-25: Prompt Template Audit

### Issue 1: Style Guide Compression (HIGH)
- **Location**: `internal/application/agents/stylist/prompts.go` lines 11-27
- **Description**: Go's inline style guide is ~50 words vs Python's ~2000 word files
- **Impact**: May significantly affect diagram/plot aesthetic quality
- **Status**: OPEN
- **Recommendation**: Replace inline constant with file loading or expand to full guide

### Issue 2: Missing Lite Mode (MEDIUM)
- **Location**: `internal/application/agents/retriever/agent.go`
- **Description**: Go lacks Python's `lite=True/False` toggle for retrieval
- **Impact**: No option for high-precision retrieval with full methodology
- **Status**: OPEN
- **Recommendation**: Add `RetrieveMode` enum with Lite/Full options

### Issue 3: Output Key Naming (LOW)
- **Location**: `internal/application/agents/retriever/prompt.go`
- **Description**: Go uses `top10_diagrams`/`top10_plots`, Python uses `top10_references`
- **Impact**: API compatibility issues if downstream expects Python naming
- **Status**: OPEN
- **Recommendation**: Unify naming or document the difference clearly

### Issue 4: Hardcoded Token Limits (LOW)
- **Location**: `internal/application/agents/planner/prompt.go` lines 15-18
- **Description**: Token limits are hardcoded constants, not configurable
- **Impact**: Less flexibility for different use cases
- **Status**: OPEN
- **Recommendation**: Make limits configurable via Config struct

---

## 2026-03-25: Comprehensive Quality Report Synthesis

### Consolidated Issue Registry

| ID | Severity | Category | Issue | Source |
|----|----------|----------|-------|--------|
| C1 | Critical | Stop | No Cancel API Endpoint | RESEARCH-004 |
| C2 | Critical | Stop | useGenerate lacks abort capability | RESEARCH-004 |
| C3 | Critical | Prompt | Style Guide severely abbreviated | RESEARCH-002 |
| H1 | High | Memory | Image data repeated deep copy | RESEARCH-003 |
| H2 | High | i18n | EmptyState not using i18n keys | RESEARCH-001 |
| H3 | High | i18n | ErrorBoundary hardcoded messages | RESEARCH-001 |
| H4 | High | Stop | No graceful Python shutdown | RESEARCH-004 |
| H5 | High | Prompt | Missing Lite Mode in Retriever | RESEARCH-002 |
| M1 | Medium | Flow | SSE event field inconsistency | RESEARCH-003 |
| M2 | Medium | Storage | Snapshot uses JSON text | RESEARCH-003 |
| M3 | Medium | i18n | Accessibility labels not i18n | RESEARCH-001 |
| M4 | Medium | i18n | Example content not localized | RESEARCH-001 |
| M5 | Medium | Stop | Agent Cleanup methods are stubs | RESEARCH-004 |
| M6 | Medium | Stop | No temp file tracking | RESEARCH-004 |
| M7 | Medium | Stop | No session tracking for cancel | RESEARCH-004 |
| M8 | Medium | Prompt | Output key naming inconsistency | RESEARCH-002 |

### Quality Gate Status: REVIEW (72/100)
