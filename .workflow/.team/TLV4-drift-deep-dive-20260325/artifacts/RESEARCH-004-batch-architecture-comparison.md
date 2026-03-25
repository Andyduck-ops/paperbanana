# Batch Processing Architecture Comparison

## Executive Summary

This analysis compares the batch processing implementations between **paperbanana-clean** (Go) and **repo-cn** (Python), with a focus on persistence characteristics and restart recovery capabilities.

---

## 1. Batch Result Storage Analysis

### 1.1 paperbanana-clean (Go Implementation)

**Storage Location**: In-memory only

```go
// internal/application/orchestrator/batch_runner.go
type BatchRunner struct {
    agentFactory  AgentFactory
    maxConcurrent int
    eventBuffer   int
    results       map[string]*domainagent.BatchResult  // <-- In-memory map
    mu            sync.RWMutex
}
```

**Key Findings**:
- Batch results are stored in a `map[string]*domainagent.BatchResult` within `BatchRunner.results`
- No database table exists for batch results (confirmed in `internal/infrastructure/persistence/sqlite/models.go`)
- Results are only accessible via `GetBatchResult(batchID string)` method
- **No persistence mechanism**: Results are lost on server restart

**Storage Flow**:
```
HTTP Request -> BatchRunner.StartBatch()
             -> executeBatch() (goroutine)
             -> storeResult(batchID, &batchResult)  // Memory only
             -> DownloadBatchZip() retrieves from memory
```

### 1.2 repo-cn (Python Implementation)

**Storage Location**: File-based JSON persistence

```python
# main.py
async def save_results_and_scores(current_results):
    async with aiofiles.open(output_filename, "w", encoding="utf-8") as f:
        json_string = json.dumps(current_results, ensure_ascii=False, indent=4)
        await f.write(json_string)

# Incremental saves every 10 results
if idx % 10 == 0:
    await save_results_and_scores(all_result_list)
```

**Key Findings**:
- Results are written to JSON files incrementally (every 10 samples)
- Final save ensures all results are persisted
- Uses `aiofiles` for async file I/O
- Results directory configurable via `ExpConfig.result_dir`

**Storage Flow**:
```
Input Data -> PaperVizProcessor.process_queries_batch()
           -> AsyncGenerator yields results
           -> Incremental JSON save every 10 results
           -> Final JSON file contains all results
```

---

## 2. Database Persistence Comparison

| Feature | paperbanana-clean | repo-cn |
|---------|------------------|---------|
| Batch Results Table | **None** | N/A (file-based) |
| Session Persistence | Yes (SessionModel) | No |
| Version History | Yes (VisualizationVersionModel) | No |
| Asset Metadata | Yes (AssetModel) | No |
| Progress Tracking | In-memory events | Progress bar + incremental saves |

### 2.1 Session Persistence in paperbanana-clean

The system has robust session persistence for individual sessions:

```go
// internal/infrastructure/persistence/sqlite/session_repository.go
type SessionRepository struct {
    db *gorm.DB
}

func (r *SessionRepository) Create(ctx context.Context, session *workspace.SessionRecord) error {
    model := sessionToModel(session)
    return r.db.WithContext(ctx).Create(model).Error
}
```

**SessionRecord includes**:
- Full `SessionState` as JSON snapshot
- Status tracking
- Project/Visualization associations
- Timestamps for lifecycle tracking

**However**: Batch-level metadata is NOT persisted separately. Each candidate session IS persisted individually (via `SnapshotStore.Save()`), but the aggregate `BatchResult` is not.

---

## 3. Restart Recovery Analysis

### 3.1 paperbanana-clean

**Recovery Capability**: **Partial** - Individual sessions can be resumed

```go
// internal/application/orchestrator/runner.go
func (r *Runner) Resume(ctx context.Context, input domainagent.AgentInput) (*RunHandle, error) {
    tracker, remaining, err := r.resumeTracker(input)
    // ...
}

func (r *Runner) resumeTracker(input domainagent.AgentInput) (*sessionTracker, []domainagent.StageName, error) {
    // Searches backwards through completed stages
    for index := len(searchStages) - 1; index >= 0; index-- {
        snapshot, err := r.snapshotStore.Restore(input.SessionID, stage)
        // ...
    }
}
```

**What CAN be recovered**:
- Individual session states via `SnapshotStore`
- Stage-level restore points
- Agent state rehydration

**What CANNOT be recovered**:
- Batch ID and its aggregate results
- Batch progress tracking
- Download ZIP availability

**Recovery Flow**:
```
Server Restart -> BatchRunner.results = {} (empty)
              -> Previous batch IDs invalid
              -> DownloadBatchZip returns 404 "batch not found or expired"

But:
              -> Individual sessions CAN be queried from DB
              -> Resume API can restart from last checkpoint
```

### 3.2 repo-cn

**Recovery Capability**: **Full** via file persistence

```python
# Results persisted incrementally
if idx % 10 == 0:
    await save_results_and_scores(all_result_list)
```

**Recovery Approach**:
- JSON file contains all completed results
- Can resume by reading existing results file
- No built-in checkpoint recovery for in-progress items

---

## 4. Batch Progress Tracking

### 4.1 paperbanana-clean

**Tracking Mechanism**: SSE (Server-Sent Events)

```go
// internal/api/handlers/batch.go
func (h *BatchHandler) StreamBatchGenerate(c *gin.Context) {
    // ...
    for event := range handle.Events() {
        c.SSEvent(string(event.Type), event)
        c.Writer.Flush()
    }
}
```

**Event Types** (from `internal/domain/agent/events.go`):
- `EventBatchStarted`
- `EventCandidateStart`
- `EventCandidateComplete`
- `EventBatchCompleted`

**Progress Status**: Real-time streaming, but **not persisted**.

### 4.2 repo-cn

**Tracking Mechanism**: Progress bar with metrics

```python
# utils/paperviz_processor.py
with tqdm(total=len(tasks), desc="Processing concurrently") as pbar:
    for future in asyncio.as_completed(tasks):
        result_data = await future
        all_result_list.append(result_data)
        # Display metrics
        pbar.set_postfix(postfix_dict)
        pbar.update(1)
        yield result_data
```

**Progress Status**: Console progress bar, no web-based streaming.

---

## 5. Architecture Diagram

```
paperbanana-clean Batch Architecture
====================================

┌─────────────────────────────────────────────────────────────────┐
│                         HTTP Layer                               │
│  ┌───────────────────┐   ┌───────────────────────────────────┐  │
│  │ StreamBatchGenerate│   │      DownloadBatchZip            │  │
│  │   (SSE Stream)     │   │  (reads from memory)             │  │
│  └─────────┬─────────┘   └──────────────┬────────────────────┘  │
└────────────┼───────────────────────────┼────────────────────────┘
             │                           │
             ▼                           │
┌────────────────────────────────────┐   │
│         BatchRunner                │   │
│  ┌──────────────────────────────┐  │   │
│  │  results map[string]         │◄──┼───┘
│  │  *BatchResult                │  │  (in-memory only)
│  │                              │  │
│  │  ┌──────────────────────┐    │  │
│  │  │ batch_id → result    │    │  │
│  │  └──────────────────────┘    │  │
│  └──────────────────────────────┘  │
└────────────────┬───────────────────┘
                 │ creates
                 ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Individual Sessions                           │
│  ┌────────────────┐  ┌────────────────┐  ┌────────────────┐    │
│  │ Session-cand-0 │  │ Session-cand-1 │  │ Session-cand-N │    │
│  │   (SQLite)     │  │   (SQLite)     │  │   (SQLite)     │    │
│  └────────────────┘  └────────────────┘  └────────────────┘    │
│                                                                  │
│  Each session IS persisted via SnapshotStore                    │
│  But BatchResult aggregation is NOT                             │
└─────────────────────────────────────────────────────────────────┘
```

---

## 6. Key Findings Summary

### 6.1 Critical Issues in paperbanana-clean

| Issue | Severity | Impact |
|-------|----------|--------|
| Batch results in-memory only | **High** | Lost on restart, cannot resume batch |
| No batch ID to sessions mapping | Medium | Cannot reconstruct batch from sessions |
| No batch progress persistence | Medium | Real-time only, no recovery |
| Download ZIP unavailable after restart | High | User experience degradation |

### 6.2 Strengths of paperbanana-clean

- Robust session-level persistence with `SnapshotStore`
- Full restore capability for individual sessions
- SSE-based real-time progress streaming
- Transactional version history management

### 6.3 Comparison Summary

| Capability | paperbanana-clean | repo-cn |
|------------|-------------------|---------|
| Batch result persistence | **No** | Yes (JSON file) |
| Session persistence | Yes | No |
| Resume from failure | Yes (session-level) | Partial (file-based) |
| Progress streaming | Yes (SSE) | No (console only) |
| ZIP download | Yes (memory-bound) | No |
| Restart recovery | Session-level only | File-level |

---

## 7. Recommendations

See `persistence-recommendations.md` for detailed implementation guidance.
