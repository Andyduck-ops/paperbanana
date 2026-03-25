# PaperBanana Drift Analysis Report

**Session**: coord-20260325-drift-analysis
**Date**: 2026-03-25
**Status**: Analysis Complete

---

## Executive Summary

**Verdict**: **Moderate Drift Detected** - paperbanana-clean's operational code has architectural advantages but implementation gaps compared to repo-cn.

| Aspect | repo-cn (Python) | paperbanana-clean (Go) | Verdict |
|--------|------------------|------------------------|---------|
| **Pipeline Logic** | ✅ Simple, proven | ⚠️ Complex, incomplete | repo-cn better |
| **Error Handling** | ⚠️ Basic retry | ✅ Classified, circuit breaker | Go better |
| **State Management** | ❌ No resume | ✅ Snapshot + restore | Go better |
| **Resume Capability** | ❌ Missing | ✅ Full support | Go better |
| **Observability** | ⚠️ Print statements | ⚠️ Not connected | Both weak |
| **Production Ready** | ✅ Working end-to-end | ⚠️ Gaps identified | repo-cn better |
| **Provider Support** | ✅ 2 providers | ✅ 5+ providers | Go better |
| **Type Safety** | ❌ Runtime errors | ✅ Compile-time | Go better |

---

## Architecture Comparison

### Pipeline Implementation

**repo-cn (Python)**
```
Retriever → Planner → Stylist → Visualizer → Critic (iterative)
     │                    │           │            │
     └────────────────────┴───────────┴────────────┘
                      Dict[str, Any] (in-memory)
```
- Simple dict-based state flow
- Critic iteration with early stopping
- Direct API calls with retry wrappers

**paperbanana-clean (Go)**
```
Retriever → Planner → Stylist → Visualizer → Critic (iterative)
     │                    │           │            │
     └────────────────────┴───────────┴────────────┘
               AgentInput → AgentOutput (typed)
                      │
               SnapshotStore → SQLite
```
- Strong typing with domain types
- Session persistence for resume
- Circuit breaker + classified errors

### Key Differences

| Feature | repo-cn | paperbanana-clean |
|---------|---------|-------------------|
| **Agent Interface** | `process(data) -> data` | `Initialize/Execute/Cleanup/GetState/RestoreState` |
| **State Flow** | Mutable dict | Immutable clones |
| **Persistence** | JSON dump every 10 samples | SQLite per-stage snapshot |
| **Streaming** | AsyncGenerator | SSE via channels |
| **Configuration** | YAML + dataclass | Viper + validation |

---

## Identified Gaps in paperbanana-clean

### Critical (P0)

1. **No Graceful Degradation for LLM Failures**
   - When LLM fails mid-pipeline, entire session fails
   - repo-cn continues with text-only if image fails

2. **Hardcoded Plot Execution Security Risk**
   - `exec.Command("python3", "-c", ...)` without sandboxing
   - Arbitrary code execution vulnerability

3. **Batch Results In-Memory Only**
   - Lost on server restart
   - No persistence layer for batch results

### High (P1)

4. **Limited Observability**
   - No structured logging of LLM requests
   - Prometheus endpoints defined but not connected
   - No distributed tracing

5. **No Health Check Dependencies**
   - `/health` always returns OK
   - No database/Redis/LLM connectivity checks

6. **Session State Size Growth**
   - No pruning of large artifacts
   - Database bloat for batches

### Medium (P2)

7. **Tight Coupling in main.go**
   - 375-line bootstrap with manual DI
   - Difficult to test

8. **Missing Request Validation Schema**
   - No JSON schema validation
   - Inconsistent error messages

---

## What repo-cn Does Better

1. **End-to-End Working Pipeline**
   - Proven in production use
   - Complete critic iteration logic
   - All agent implementations functional

2. **Graceful Degradation**
   - Missing reference file → fallback to none
   - Failed image → text-only critique
   - Critic "no changes" → early stop

3. **Token Optimization**
   - Lite retrieval mode (96% token savings)
   - Smart prompt construction

4. **Developer Experience**
   - Simple Streamlit UI
   - Easy local development
   - Clear error messages

---

## What paperbanana-clean Does Better

1. **State Management**
   - Full resume capability
   - SQLite persistence
   - Session snapshots

2. **Error Handling**
   - Classified errors (transient vs permanent)
   - Circuit breaker pattern
   - Exponential backoff

3. **Provider Support**
   - OpenAI, Gemini, Anthropic, OpenRouter, OpenAI-compatible
   - Runtime provider switching
   - API key rotation

4. **Type Safety**
   - Compile-time checking
   - Interface contracts
   - Explicit error returns

5. **Architecture**
   - Clean separation (domain/application/infrastructure)
   - Repository pattern
   - Unit of Work for transactions

---

## Recommendations

### Option A: Fix paperbanana-clean (Recommended)

**Pros**: Keep Go advantages (type safety, resume, providers)
**Cons**: Significant work to close gaps

**Priority Tasks**:
1. Implement graceful degradation in runner.go
2. Add sandboxing for plot execution
3. Persist batch results to SQLite
4. Connect observability endpoints
5. Add health check dependencies

**Estimated Effort**: 2-3 weeks

### Option B: Port repo-cn Logic to Go

**Pros**: Proven logic, clear requirements
**Cons**: Re-implement working Python code

**Priority Tasks**:
1. Port critic iteration logic exactly
2. Implement lite retrieval mode
3. Add graceful degradation patterns
4. Create integration tests

**Estimated Effort**: 3-4 weeks

### Option C: Hybrid Approach

**Pros**: Best of both worlds
**Cons**: Architecture complexity

**Strategy**:
1. Keep Go backend for infrastructure (state, persistence, providers)
2. Create Python worker for pipeline logic
3. Communication via gRPC or HTTP

**Estimated Effort**: 4-5 weeks

---

## Decision Matrix

| Criteria | Fix Go | Port to Go | Hybrid |
|----------|--------|------------|--------|
| Time to Production | ⭐⭐⭐ | ⭐⭐ | ⭐ |
| Risk | ⭐⭐⭐ | ⭐⭐ | ⭐ |
| Maintainability | ⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐ |
| Type Safety | ⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐ |
| Resume Support | ⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐ |

**Recommendation**: **Option A - Fix paperbanana-clean**
- Preserves architectural investments
- Closes gaps systematically
- Maintains Go advantages

---

## Next Steps

1. **Immediate**: Review this analysis with team
2. **This Week**: Decide on approach (A/B/C)
3. **Next Sprint**: Begin implementation

---

*Generated by maestro-coordinate parallel agent exploration*

---

## UI & Channel Configuration Analysis

### Provider Architecture Comparison

| Feature | repo-cn | paperbanana-clean |
|---------|---------|-------------------|
| **Providers** | 2 (Evolink, Gemini) | **16+** (OpenAI, Anthropic, Gemini, DeepSeek, Zhipu, Moonshot, Qwen, Doubao, etc.) |
| **API Key Storage** | YAML file | SQLite + encryption (argon2id) |
| **API Key Rotation** | Single key | **Multi-key rotation** |
| **Config Reload** | Manual restart | **Real-time SSE** |
| **Model Discovery** | None | **Dynamic via API** |

### UI Technology Comparison

| Feature | repo-cn (Streamlit) | paperbanana-clean (React) |
|---------|---------------------|---------------------------|
| **Architecture** | Monolithic 900-line `demo.py` | Go backend + React SPA |
| **Theming** | None | 4 themes (swiss, bauhaus, neo-minimal, art-deco) |
| **Localization** | Chinese only | EN/ZH toggle |
| **History** | JSON files | Database + panel UI |
| **Setup** | Zero-config scripts | Requires build |

### What Works in paperbanana-clean (per user feedback ✅)

1. **UI Design** - React SPA with 4 themes, proper state management
2. **Provider Configuration** - 16+ presets with encrypted key storage
3. **Real-time Updates** - SSE for config changes
4. **API Key Management** - Multi-key rotation, masked display

### What's Missing in paperbanana-clean

1. **Cost Transparency** - No token cost estimates per operation
2. **Retrieval Mode Selection** - Missing `auto` vs `auto-full` toggle
3. **Auto-install Scripts** - No zero-config onboarding

---

## Final Synthesis

Your assessment is correct: **paperbanana-clean's UI and channel configuration are solid, but operational code needs work.**

### The Drift

| Layer | repo-cn | paperbanana-clean | Status |
|-------|---------|-------------------|--------|
| **UI** | Simple | ✅ Better | No drift |
| **Provider Config** | Basic | ✅ Much better | No drift |
| **Pipeline Logic** | ✅ Proven | ⚠️ Gaps | **DRIFT** |
| **Error Handling** | Basic | ⚠️ No degradation | **DRIFT** |
| **State Management** | ❌ None | ✅ Resume support | No drift |
| **Security** | ⚠️ Basic | ⚠️ Plot execution risk | Both need work |

### Root Cause of Drift

The Go implementation focused on **architecture and infrastructure** (providers, persistence, SSE) but didn't port the **operational wisdom** from Python:

1. **Graceful degradation patterns** - repo-cn's "continue on failure" logic
2. **Critic iteration refinements** - early stopping, rollback mechanism
3. **Token optimization** - lite retrieval mode (96% savings)

### Recommended Fix Priority

| Priority | Task | Effort |
|----------|------|--------|
| **P0** | Add graceful degradation to `runner.go` | 2-3 days |
| **P0** | Sandbox plot execution | 1-2 days |
| **P1** | Port critic iteration refinements | 2-3 days |
| **P1** | Add lite retrieval mode | 1 day |
| **P2** | Add cost transparency to UI | 1 day |
| **P2** | Persist batch results | 1 day |

**Total Estimated**: 8-11 days
