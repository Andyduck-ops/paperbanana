# Error Handling Completeness Analysis

**Task ID**: DEEP-005
**Date**: 2026-03-25
**Analyst**: analyst-agent

---

## Executive Summary

This report analyzes the error handling completeness in `paperbanana-clean` (Go) and compares it with `repo-cn` (Python). The Go implementation demonstrates a **significantly more sophisticated error handling architecture** with classification, categorization, and retry strategies, but has several gaps in failure scenario handling and user-friendly messaging.

| Aspect | Go (paperbanana-clean) | Python (repo-cn) | Gap Status |
|--------|------------------------|------------------|------------|
| Error Classification | Excellent | Basic | Go superior |
| Error Recovery | Good | Minimal | Go superior |
| User-Friendly Messages | Partial | Minimal | Go better but incomplete |
| Structured Logging | Good | None | Go superior |
| Graceful Degradation | Missing | Missing | Both deficient |
| Failure Scenario Coverage | Partial | Partial | Both need work |

---

## 1. Error Classification Analysis (errors.go)

### 1.1 Current Implementation

The `internal/domain/agent/errors.go` provides a **well-designed error classification system**:

```go
// Error codes defined with clear semantics
type ErrorCode string

const (
    // Transient errors - can be retried
    ErrCodeLLMTimeout         ErrorCode = "llm_timeout"
    ErrCodeRateLimit          ErrorCode = "rate_limit"
    ErrCodeServiceUnavailable ErrorCode = "service_unavailable"
    ErrCodeNetworkError       ErrorCode = "network_error"

    // Permanent errors - should not be retried
    ErrCodeInvalidInput     ErrorCode = "invalid_input"
    ErrCodeInvalidConfig    ErrorCode = "invalid_config"
    ErrCodeUnsupportedType  ErrorCode = "unsupported_type"
    ErrCodeResourceNotFound ErrorCode = "resource_not_found"

    // ... more codes
)

// Error categories for handling strategy
type ErrorCategory string

const (
    ErrorCategoryTransient     ErrorCategory = "transient"     // Can be retried
    ErrorCategoryPermanent     ErrorCategory = "permanent"     // Should not retry
    ErrorCategoryConfiguration ErrorCategory = "configuration" // User action needed
    ErrorCategoryInternal      ErrorCategory = "internal"      // System error
)
```

### 1.2 Classification Logic (ClassifyError)

**Strengths**:
- Pattern-based error matching for common failure modes
- Timeout detection via string patterns
- Rate limit detection (429, "rate limit")
- Network error detection (connection refused, no such host)
- Auth error detection (401, 403, "api key")

**Weaknesses**:

| Gap | Description | Impact |
|-----|-------------|--------|
| **Missing error codes** | No `ErrCodeInsufficientQuota`, `ErrCodeContentPolicyViolation`, `ErrCodeModelNotAvailable` | Cannot distinguish billing/policy failures |
| **Loose pattern matching** | Case-insensitive substring matching can misclassify errors | False positives/negatives in classification |
| **No structured error wrapping** | Relies on `err.Error()` string parsing | Brittle classification |
| **Missing retry-after extraction** | Rate limit errors don't capture `Retry-After` header | Cannot implement intelligent backoff |

**Recommended additions**:

```go
// Additional error codes needed
ErrCodeInsufficientQuota     ErrorCode = "insufficient_quota"     // Billing/quota exhausted
ErrCodeContentPolicyViolation ErrorCode = "content_policy_violation" // Content filtered
ErrCodeModelNotAvailable     ErrorCode = "model_not_available"    // Model deprecated/unavailable
ErrCodeContextLengthExceeded ErrorCode = "context_length_exceeded" // Token limit exceeded
```

---

## 2. Agent Error Handling Analysis

### 2.1 Retriever Agent

**File**: `internal/application/agents/retriever/agent.go`

| Scenario | Handled? | Code Location |
|----------|----------|---------------|
| LLM client nil (auto mode) | Yes | Line 216-217 |
| Store not configured | Yes | Line 283-284, 293-294 |
| File not found (candidates) | Yes | Line 286-288 (returns nil, nil) |
| JSON unmarshal failure | Yes | Line 87-88 |
| Unsupported mode | Yes | Line 244 |
| Prompt loading failure | Yes | Line 249-251 |

**Issues Found**:

1. **Silent failure on missing store**: Returns `nil` without error when `os.ErrNotExist`
   ```go
   // Line 286-288
   if errors.Is(err, os.ErrNotExist) {
       return nil, nil  // Silent failure - should return empty slice with warning
   }
   ```

2. **No error wrapping with context**: Errors lack stage context
   ```go
   // Current (Line 143)
   return domainagent.AgentOutput{}, err

   // Should be
   return domainagent.AgentOutput{}, domainagent.WrapAgentError(err, domainagent.StageRetriever, domainagent.ErrCodeExecutionFailed)
   ```

### 2.2 Planner Agent

**File**: `internal/application/agents/planner/agent.go`

| Scenario | Handled? | Code Location |
|----------|----------|---------------|
| LLM client nil | Yes | Line 70-71 |
| Prompt metadata loading | Yes | Line 74-77 |
| Reference bundle decode | Yes | Line 90-98 |
| Message building | Yes | Line 101-109 |
| LLM generation failure | Yes | Line 119-126 |
| Empty response handling | Partial | Line 129-133 (fallback used) |

**Good Patterns**:
- **Graceful fallback**: When LLM returns empty content, uses `fallbackPlanningContent()`
- **State update on failure**: Sets `a.state.Error` before returning

**Missing**:
- No distinction between transient and permanent LLM failures
- Image loading errors not classified (line 200-208)

### 2.3 Visualizer Agent

**File**: `internal/application/agents/visualizer/agent.go`

| Scenario | Handled? | Code Location |
|----------|----------|---------------|
| Unsupported mode | Yes | Line 88 |
| LLM client nil | Yes | Line 118-119, 226-227 |
| Prompt metadata loading | Yes | Line 66-68 |
| Diagram generation retry | Yes | Lines 161-183 (max 3 attempts) |
| Plot code empty | Yes | Line 251-252 |
| Plot executor failure | Yes | Line 255-258 |
| Empty rendered bytes | Yes | Line 259-261 |
| Image extraction from response | Yes | Line 369-388 |

**Excellent Pattern**: Retry logic with intelligent backoff

```go
// Lines 161-183: Diagram generation with retry
func (a *Agent) generateDiagramResponse(ctx context.Context, req domainllm.GenerateRequest) (*domainllm.GenerateResponse, int, error) {
    var lastErr error
    for attempt := 1; attempt <= maxDiagramAttempts; attempt++ {
        resp, err := a.invokeDiagramGenerator(ctx, req)
        if err == nil {
            return resp, attempt, nil
        }
        lastErr = err
        if attempt >= maxDiagramAttempts || !shouldRetryDiagramError(err) {
            break
        }
        if err := waitForRetry(ctx, diagramRetryDelay(attempt)); err != nil {
            return nil, attempt, err
        }
    }
    // ...
}
```

**Issues**:

1. **Plot executor has no sandboxing** (CRITICAL SECURITY)
   ```go
   // plot_executor.go line 82
   exec(code, namespace)  // Direct exec() - RCE vulnerability
   ```

2. **No timeout on Python process**: `cmd.CombinedOutput()` uses context but no explicit process kill on timeout

### 2.4 Critic Agent

**File**: `internal/application/agents/critic/agent.go`

| Scenario | Handled? | Code Location |
|----------|----------|---------------|
| LLM client nil | Yes | Line 69-70 |
| Prompt metadata loading | Yes | Line 73-76 |
| Message building | Yes | Line 92-100 |
| LLM generation failure | Yes | Line 103-113 |
| Empty response | Yes | Line 261-263 |
| JSON parse failure | Yes | Line 270-273 |
| Revision agent missing | Yes | Line 140-141 |
| Revision execution failure | Yes | Line 144-147 |

**Good Pattern**: fail() helper method standardizes error state

```go
func (a *Agent) fail(err error) (domainagent.AgentOutput, error) {
    a.state.Status = domainagent.StatusFailed
    a.state.Error = &domainagent.ErrorDetail{
        Message: err.Error(),
        Stage:   domainagent.StageCritic,
    }
    return domainagent.AgentOutput{}, err
}
```

**Issues**:

1. **No error code classification in fail()**: Should use `ClassifyError()` to determine retryability

2. **Revision agent requirement check is late**: Checked at runtime in the loop, not at initialization

---

## 3. Orchestrator Error Recovery (runner.go)

### 3.1 Stage Execution Error Handling

**File**: `internal/application/orchestrator/runner.go`

The `finishStageError()` method provides comprehensive error handling:

```go
func (r *Runner) finishStageError(...) (RunResult, error) {
    // 1. Classify the error
    errorCode := domainagent.ClassifyError(err)
    detail := domainagent.WrapAgentError(err, stage, errorCode)

    // 2. Add timeout-specific handling
    if errors.Is(ctx.Err(), context.DeadlineExceeded) {
        detail.Code = string(domainagent.ErrCodeStageTimeout)
        detail.Category = string(domainagent.ErrorCategoryTransient)
        detail.Suggestion = fmt.Sprintf("The %s stage took too long...", stage)
    }

    // 3. Persist state for resume capability
    if persistErr := r.persistSnapshot(tracker, stageState); persistErr != nil {
        err = errors.Join(err, persistErr)
    }

    // 4. Emit events for SSE streaming
    publisher.emit(domainagent.EventStageFailed, ...)
    publisher.emit(domainagent.EventRunFailed, ...)
}
```

### 3.2 Resume Error Handling

| Scenario | Handled? | Code Location |
|----------|----------|---------------|
| No session ID | Yes | Line 377-378 |
| No snapshot store | Yes | Line 380-381 |
| Snapshot not found | Yes | Line 390-391 |
| Snapshot invalid | Yes | Line 390-391 |
| Agent not registered | Yes | Line 428-431, 441-444 |
| State restoration failure | Yes | Line 447-451 |

### 3.3 Context Cancellation Handling

```go
// Line 315-326: Proper cancellation handling
if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) ... {
    stageState.Status = domainagent.StatusCanceled
    tracker.failStage(stageState, domainagent.StatusCanceled, detail)
    // ... persist and emit events
    publisher.emit(domainagent.EventRunCanceled, ...)
}
```

**Issues**:

1. **No retry at orchestrator level**: When a transient error occurs, the orchestrator immediately fails rather than retrying the stage
2. **No partial progress preservation**: Failed stages don't preserve intermediate results for debugging
3. **No graceful degradation**: Cannot continue with reduced functionality when optional stages fail

---

## 4. Batch Runner Error Handling

**File**: `internal/application/orchestrator/batch_runner.go`

| Scenario | Handled? | Code Location |
|----------|----------|---------------|
| Empty input list | Yes | Line 122-123 |
| Shared retriever failure | Yes | Line 154-167 |
| Individual candidate failure | Yes | Line 211-240 |
| Context cancellation | Partial | Uses errgroup but no explicit handling |

**Good Pattern**: Candidate failures don't fail the entire batch

```go
// Line 221-222
return nil // Don't fail the whole batch
```

**Issues**:

1. **No persistence on failure**: Batch results lost on server restart (already documented in ISSUE-006)
2. **No result expiration**: Results stored indefinitely in memory
3. **No partial batch recovery**: Cannot resume a partially completed batch
4. **Error detail missing classification**: `ErrorDetail` created without using `ClassifyError()`

```go
// Line 217-218: Missing error classification
Error: &domainagent.ErrorDetail{Message: err.Error()},
// Should be:
Error: domainagent.WrapAgentError(err, stage, domainagent.ClassifyError(err)),
```

---

## 5. API Handler Error Handling

### 5.1 Generate Handler

**File**: `internal/api/handlers/generate.go`

| Scenario | HTTP Status | Code Location |
|----------|-------------|---------------|
| Invalid JSON | 400 | Line 87-88 |
| Validation failure | 400 | Line 91-93 |
| Resume without session | 400 | Line 178-181 |
| Unsupported mode | 400 | Line 181 |
| Internal errors | 500 | Line 177 |

**Good Pattern**: `buildStreamErrorPayload()` provides detailed error information

```go
func buildStreamErrorPayload(err error, input domainagent.AgentInput, result orchestrator.RunResult) gin.H {
    payload := gin.H{
        "message":    err.Error(),
        "session_id": input.SessionID,
        // ...
    }
    if detail := result.Session.Error; detail != nil {
        payload["code"] = detail.Code
        payload["category"] = detail.Category
        payload["retryable"] = detail.Retryable
        payload["suggestion"] = detail.Suggestion
    }
}
```

### 5.2 Batch Handler

**File**: `internal/api/handlers/batch.go`

| Scenario | HTTP Status | Code Location |
|----------|-------------|---------------|
| Invalid JSON | 400 | Line 42-43 |
| Validation failure | 400 | Line 53-55 |
| Batch not found/expired | 404 | Line 203-206 |
| Too many candidates | 400 | Line 108-110 |

**Issues**:

1. **No rate limiting**: No protection against batch abuse
2. **No batch status polling**: Client cannot check progress without keeping SSE connection
3. **ZIP download on missing batch**: Returns 404 but no alternative download mechanism

---

## 6. Comparison with Python (repo-cn)

### 6.1 Error Handling Approach

| Aspect | Python (repo-cn) | Go (paperbanana-clean) |
|--------|------------------|------------------------|
| Error classification | None - uses raw exceptions | Full classification system |
| Retry logic | Manual in API calls (`max_attempts`) | Structured with backoff delays |
| State persistence on error | None | Snapshot persistence |
| User-facing messages | Print statements | Structured `ErrorDetail` with suggestions |
| Cancellation handling | None | Full context propagation |
| Graceful degradation | None | None |

### 6.2 Python Code Analysis

**File**: `repo-cn/utils/generation_utils.py`

```python
# Python retry pattern (line 452-493)
for attempt in range(max_attempts):
    try:
        response = await client.aio.models.generate_content(...)
        # ...
    except Exception as e:
        current_delay = min(retry_delay * (2 ** attempt), 30)
        print(f"Attempt {attempt + 1} failed: {e}. Retrying...")
        if attempt < max_attempts - 1:
            await asyncio.sleep(current_delay)
        else:
            result_list = ["Error"] * target_candidate_count  # Returns "Error" strings
```

**Key Differences**:

1. **Python returns `"Error"` strings** instead of structured error objects
2. **No error classification** - all errors treated uniformly
3. **No state preservation** for resume
4. **No cancellation handling** - long-running requests cannot be stopped

### 6.3 Python Visualizer Error Handling

**File**: `repo-cn/agents/visualizer_agent.py`

```python
# Line 43-59: No error handling for exec()
try:
    exec_globals = {}
    exec(code_clean, exec_globals)  # CRITICAL: No sandboxing, no exception handling
    # ...
except Exception as e:
    print(f"Error executing plot code: {e}")
    return None  # Silent failure - caller cannot distinguish from "no figure"
```

**Comparison**:
- Go: Returns `ErrorDetail` with stage context and classification
- Python: Prints to console and returns `None`

---

## 7. User-Friendly Error Messages

### 7.1 Current Implementation

The `errorCodeInfo` map provides user-friendly suggestions:

```go
var errorCodeInfo = map[ErrorCode]struct {
    Category   ErrorCategory
    Suggestion string
}{
    ErrCodeRateLimit: {
        Category:   ErrorCategoryTransient,
        Suggestion: "Rate limit exceeded. Please wait a moment and try again.",
    },
    ErrCodeMissingAPIKey: {
        Category:   ErrorCategoryConfiguration,
        Suggestion: "API key is missing. Please configure your API key in settings.",
    },
    // ...
}
```

### 7.2 Missing User Messages

| Error Code | Current | Recommended Addition |
|------------|---------|---------------------|
| `ErrCodeStageTimeout` | Generic timeout message | "The [stage] step took too long. Try simplifying your request or using a smaller model." |
| `ErrCodeExecutionFailed` | Generic internal message | "Code generation failed. Please try a different description." |
| `ErrCodeCancelled` | "The operation was cancelled" | "Your request was cancelled. This may be due to a timeout or manual cancellation." |
| Plot execution error | "execute plot code: [stderr]" | "Plot rendering failed. The generated code may have errors. Error: [simplified message]" |

### 7.3 Frontend Integration

The current `buildStreamErrorPayload()` provides good structure but frontend may not use all fields:

```json
{
  "message": "...",
  "error": "...",
  "code": "rate_limit",
  "category": "transient",
  "retryable": true,
  "suggestion": "Rate limit exceeded. Please wait...",
  "stage": "visualizer",
  "failed_stage": "visualizer"
}
```

**Recommendation**: Frontend should display `suggestion` when available, with `retryable` determining retry button visibility.

---

## 8. Failure Scenario Coverage Matrix

| Failure Scenario | Classified? | Recovered? | User Message? | Logged? |
|-----------------|-------------|------------|---------------|---------|
| LLM timeout | Yes | No (fails stage) | Yes | Yes |
| Rate limit | Yes | No | Yes | Yes |
| Network error | Yes | No | Yes | Yes |
| Invalid API key | Yes | No | Yes | Yes |
| Model not found | No (falls through) | No | No | Yes |
| Content policy violation | No | No | No | Yes |
| Quota exceeded | No | No | No | Yes |
| Context length exceeded | No | No | No | Yes |
| Python exec error | No | No | Partial | Yes |
| Image generation empty | No | No | No | Yes |
| Snapshot persistence fail | N/A | No | No | Yes |
| Resume from corrupt snapshot | Yes | Yes (error returned) | Partial | Yes |

---

## 9. Recommendations

### 9.1 Critical (P0)

1. **Add missing error codes for LLM failures**:
   - `ErrCodeInsufficientQuota`
   - `ErrCodeContentPolicyViolation`
   - `ErrCodeModelNotAvailable`
   - `ErrCodeContextLengthExceeded`

2. **Sandbox Python execution** (already documented in ISSUE-001)
   - Use Docker container isolation
   - Add execution timeout with process kill
   - Implement resource limits

3. **Add error classification to agent `fail()` methods**:
   ```go
   func (a *Agent) fail(err error) (domainagent.AgentOutput, error) {
       code := domainagent.ClassifyError(err)
       detail := domainagent.NewErrorDetail(code, err.Error(), false)
       detail.Stage = domainagent.StageCritic
       // ...
   }
   ```

### 9.2 High (P1)

4. **Implement stage-level retry for transient errors**:
   ```go
   func (r *Runner) executeStageWithRetry(ctx context.Context, ...) {
       maxRetries := 2
       for attempt := 0; attempt <= maxRetries; attempt++ {
           output, err := stageAgent.Execute(stageCtx, stageInput)
           if err == nil { return output, nil }
           if !isRetryable(err) { return output, err }
           // Wait with backoff
       }
   }
   ```

5. **Add graceful degradation for optional stages**:
   - Stylist failure should allow pipeline to continue
   - Critic failure should preserve rendered output
   - Add `stage.optional` flag to configuration

6. **Preserve intermediate results on failure**:
   - Store partial artifacts even when stage fails
   - Add `preserve_on_failure` option to agent config

### 9.3 Medium (P2)

7. **Enhance user messages for common failures**:
   - Add specific messages for plot execution errors
   - Add context-length suggestions
   - Add billing/quota exhaustion messages

8. **Extract retry-after from rate limit responses**:
   ```go
   type RateLimitError struct {
       Base      error
       RetryAfter time.Duration
   }
   ```

9. **Add error telemetry/aggregation**:
   - Track error frequency by code
   - Alert on unusual error patterns
   - Dashboard for error rates

### 9.4 Low (P3)

10. **Add structured error logging with context**:
    - Include session_id, stage, error_code in all log entries
    - Add correlation IDs for distributed tracing

---

## 10. Summary

The Go implementation (`paperbanana-clean`) provides a **robust error handling foundation** that significantly exceeds the Python reference (`repo-cn`) in:

- **Error classification and categorization**
- **Retry strategies with intelligent backoff**
- **State persistence for resume capability**
- **Context cancellation propagation**
- **Structured error details for API responses**

However, several gaps remain:

1. **Missing error codes** for common LLM failures (quota, content policy, context length)
2. **No stage-level retry** for transient errors at orchestrator level
3. **No graceful degradation** when optional stages fail
4. **Incomplete user-friendly messages** for technical errors
5. **Python executor security** remains a critical vulnerability

The architecture is well-designed for extensibility. Adding the recommended error codes and retry logic would bring the error handling to production-ready status.

---

## Appendix A: Error Code Decision Tree

```
Error occurs
    |
    v
[LLM Error?]
    |-- timeout/deadline --> ErrCodeLLMTimeout (transient)
    |-- rate limit/429 --> ErrCodeRateLimit (transient)
    |-- 502/503/504 --> ErrCodeServiceUnavailable (transient)
    |-- network error --> ErrCodeNetworkError (transient)
    |-- 401/403/api key --> ErrCodeMissingAPIKey (configuration)
    |-- 404/not found --> ErrCodeResourceNotFound (permanent)
    |-- content filtered --> ErrCodeContentPolicyViolation (permanent)
    |-- quota exceeded --> ErrCodeInsufficientQuota (configuration)
    |-- context length --> ErrCodeContextLengthExceeded (permanent)
    |-- model unavailable --> ErrCodeModelNotAvailable (configuration)
    |-- other --> ErrCodeInternalError (internal)
    |
[Execution Error?]
    |-- Python exec fail --> ErrCodeExecutionFailed (internal)
    |-- Empty output --> ErrCodeExecutionFailed (internal)
    |-- Invalid input --> ErrCodeInvalidInput (permanent)
    |
[Context Error?]
    |-- cancelled --> ErrCodeCancelled (permanent)
    |-- deadline exceeded --> ErrCodeStageTimeout (transient)
    |
[Unknown] --> ErrCodeUnknown (internal)
```

---

## Appendix B: Files Analyzed

| File | Lines Analyzed |
|------|----------------|
| `internal/domain/agent/errors.go` | 261 |
| `internal/application/orchestrator/runner.go` | 715 |
| `internal/application/orchestrator/batch_runner.go` | 341 |
| `internal/application/agents/retriever/agent.go` | 541 |
| `internal/application/agents/planner/agent.go` | 333 |
| `internal/application/agents/visualizer/agent.go` | 508 |
| `internal/application/agents/visualizer/plot_executor.go` | 92 |
| `internal/application/agents/critic/agent.go` | 479 |
| `internal/api/handlers/generate.go` | 388 |
| `internal/api/handlers/batch.go` | 282 |
| `repo-cn/utils/generation_utils.py` | 689 |
| `repo-cn/agents/visualizer_agent.py` | 240 |
| `repo-cn/agents/critic_agent.py` | 241 |
