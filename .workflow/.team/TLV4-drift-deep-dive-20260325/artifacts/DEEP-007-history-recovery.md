# DEEP-007: History Data Storage, Loading, and Recovery Mechanism Analysis

## Executive Summary

The paperbanana system implements a comprehensive history and recovery mechanism through three integrated layers:
1. **SQLite Persistence Layer** - Session records stored via GORM with JSON-serialized snapshots
2. **Application Service Layer** - Transactional history service managing session and version persistence
3. **Orchestrator Layer** - Real-time snapshot persistence during pipeline execution

---

## 1. Session Data Persistence Mechanism

### 1.1 Storage Architecture

**Primary Storage: SQLite Database via GORM**

```
SessionModel (sessions table)
├── ID              string  (primary key, varchar(36))
├── ProjectID       string  (indexed, foreign key to projects)
├── VisualizationID *string (indexed, foreign key to visualizations)
├── Status          string  (indexed: running/completed/failed/canceled)
├── CurrentStage    string  (indexed)
├── SchemaVersion   string
├── SnapshotJSON    JSON    (full SessionState serialized)
├── CreatedAt       time.Time
├── UpdatedAt       time.Time
└── CompletedAt     *time.Time
```

**Location:** `internal/infrastructure/persistence/sqlite/models.go:96-107`

### 1.2 Snapshot Payload Structure

The `SessionSnapshotPayload` captures the complete agent execution state:

```go
type SessionSnapshotPayload struct {
    SchemaVersion string
    SessionID     string
    RequestID     string
    Status        RunStatus
    CurrentStage  StageName
    Pipeline      []StageName
    InitialInput  AgentInput
    StageStates   []AgentState      // Full history of completed stages
    FinalOutput   AgentOutput
    Error         *ErrorDetail
    Restore       RestoreMetadata   // Resume tracking
    Metadata      map[string]string
    StartedAt     time.Time
    UpdatedAt     time.Time
    CompletedAt   time.Time
}
```

**Location:** `internal/infrastructure/persistence/sqlite/models.go:115-131`

### 1.3 Persistence Flow

```
Pipeline Execution
       │
       ▼
┌─────────────────────────┐
│   sessionTracker        │ (in-memory state tracking)
│   - Tracks stage states │
│   - Maintains snapshot  │
└───────────┬─────────────┘
            │ (after each stage)
            ▼
┌─────────────────────────┐
│ PersistentSnapshotStore │
│   Save() / SaveWithProject()
└───────────┬─────────────┘
            │
            ▼
┌─────────────────────────┐
│   SessionRepository     │
│   Create() / Update()   │
└───────────┬─────────────┘
            │
            ▼
    SQLite Database
```

**Key Code Locations:**
- Orchestrator snapshot persistence: `internal/application/orchestrator/runner.go:369-374`
- Persistent snapshot store: `internal/infrastructure/persistence/sqlite/snapshot_store.go:29-72`
- Session repository: `internal/infrastructure/persistence/sqlite/session_repository.go`

---

## 2. History API Implementation

### 2.1 API Endpoints

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/api/v1/history` | GET | List version history for visualization |
| `/api/v1/history/:project_id/:version_id` | GET | Get specific version |
| `/api/v1/sessions/recent` | GET | List recent sessions (with batch grouping) |
| `/api/v1/session/latest` | GET | Get latest session for visualization |
| `/api/v1/session/:session_id` | GET | Get session by ID |

**Location:** `internal/api/router.go:108-112`

### 2.2 History Service Interface

```go
type HistoryService interface {
    ListHistory(ctx, projectID, visualizationID string, limit int) ([]*VisualizationVersion, error)
    ListRecentSessions(ctx, projectID string, limit int) ([]*SessionRecord, error)
    GetVersion(ctx, projectID, versionID string) (*VisualizationVersion, error)
    GetLatestSession(ctx, projectID, visualizationID string) (*SessionRecord, error)
    GetSessionByID(ctx, sessionID string) (*SessionRecord, error)
}
```

**Location:** `internal/api/handlers/history.go:25-31`

### 2.3 Batch Session Grouping

The `ListRecentSessions` endpoint includes special logic for batch execution grouping:

- Sessions with `batch.group_id` metadata are aggregated
- Only the latest session in each batch group is returned
- Batch metadata includes:
  - `batch_id`: Group identifier
  - `candidate_session_ids`: List of all candidate session IDs
  - Aggregated status (running if any running, completed if all completed)

**Location:** `internal/api/handlers/history.go:402-450`

---

## 3. User History View Implementation

### 3.1 Frontend Hook: useHistory

```typescript
export function useHistory(projectId?: string) {
  const [state, setState] = useState<HistoryState>({
    sessions: [],
    isLoading: false,
    error: null,
  });

  const fetchHistory = useCallback(async () => {
    const response = await apiClient.listHistory(projectId);
    // Maps to HistorySession interface
  }, [projectId]);

  return { sessions, isLoading, error, refresh: fetchHistory };
}
```

**Location:** `web/src/hooks/useHistory.ts`

### 3.2 History Sidebar Component

The `HistorySidebar` component:
- Fetches sessions on mount via `useHistory` hook
- Displays sessions with thumbnail, prompt, and timestamp
- Supports selection for session resumption

**Location:** `web/src/components/HistorySidebar.tsx`

### 3.3 API Client Method

```typescript
async listHistory(projectId?: string) {
  const query = projectId ? `?project_id=${projectId}` : '';
  const response = await fetch(`${API_BASE}/history${query}`);
  // Returns { sessions: [...] }
}
```

**Location:** `web/src/lib/api.ts:79-90`

---

## 4. Session Recovery Mechanism

### 4.1 Resume Flow

```
User Request (resume: true, session_id: "xxx")
       │
       ▼
┌─────────────────────────┐
│   Generate Handler      │
│   h.runner.Resume()     │
└───────────┬─────────────┘
            │
            ▼
┌─────────────────────────┐
│   Orchestrator Runner   │
│   resumeTracker()       │
└───────────┬─────────────┘
            │
            ▼
┌─────────────────────────┐
│ PersistentSnapshotStore │
│   Restore(sessionID, stage)
└───────────┬─────────────┘
            │
            ▼
┌─────────────────────────┐
│   SessionRepository     │
│   GetByID(sessionID)    │
└───────────┬─────────────┘
            │
            ▼
    Reconstruct SessionState
    Find last completed stage
    Restore agent states
    Resume from next stage
```

### 4.2 Resume Logic Detail

**Orchestrator Resume Process** (`internal/application/orchestrator/runner.go:376-409`):

1. **Validate session ID** - Return `ErrResumeRequiresSession` if empty
2. **Check snapshot store** - Return `ErrResumeStoreMissing` if not configured
3. **Search for restore point** - Iterate pipeline stages in reverse order
4. **Find completed stage** - Look for stage with `StatusCompleted`
5. **Reconstruct tracker** - Build `sessionTracker` from snapshot
6. **Restore agent states** - Call `RestoreState()` on each completed stage agent
7. **Return remaining pipeline** - Execute from next stage after restore point

### 4.3 Restore Metadata

```go
type RestoreMetadata struct {
    SnapshotVersion string    // Schema version of the snapshot
    RestoredFrom    StageName // Stage from which resume started
    RestoredAt      time.Time // Timestamp of resume operation
    ResumeToken     string    // Optional token for external tracking
}
```

**Location:** `internal/domain/agent/types.go:105-110`

### 4.4 Event Notification

When resuming, the system emits `EventResumeStarted` with metadata:
- `resumed_from_stage`: The stage being resumed from
- `session_id`: The session being resumed

**Location:** `internal/application/orchestrator/runner.go:174-181`

---

## 5. Potential Data Loss Scenarios

### 5.1 Identified Risk Areas

#### Risk 1: No Transaction on Initial Snapshot Creation
**Severity:** Medium

The `Save()` method in `PersistentSnapshotStore` creates sessions without explicit transaction handling for the initial creation:

```go
// snapshot_store.go:55-63
existing, err := s.sessionRepo.GetByID(ctx, session.SessionID)
if err != nil {
    // Session doesn't exist, create it
    if projectID == "" {
        record.ProjectID = "default"  // Fallback default project
    }
    return s.sessionRepo.Create(ctx, record)
}
```

**Risk:** If the database operation fails after partial writes, the session may be partially persisted.

**Mitigation:** The `HistoryService.SaveSession()` wraps this in a transaction, but the direct `Save()` method bypasses it.

---

#### Risk 2: Missing Project Association
**Severity:** Medium

Sessions without project scope fall back to `"default"` project ID:

```go
// snapshot_store.go:58-62
if projectID == "" {
    // For sessions without project scope (legacy/Phase 2 compatibility)
    record.ProjectID = "default"
}
```

**Risk:** Sessions created outside the project flow become associated with a synthetic "default" project, potentially making them harder to discover and manage.

---

#### Risk 3: In-Memory State Loss Before First Persistence
**Severity:** High

The `sessionTracker` maintains state in memory and persists snapshots via `persistSnapshot()`:

```go
// runner.go:249-251
tracker.completeStage(stageState, output)
if err := r.persistSnapshot(tracker, stageState); err != nil {
    return r.finishPersistenceError(...)
}
```

**Risk:** If the process crashes before the first stage completes, no snapshot is persisted. The session exists only in memory.

**Impact:** Users cannot resume sessions that fail during the first stage.

---

#### Risk 4: Cascade Delete on Project Deletion
**Severity:** High (by design)

The `SessionModel` uses `OnDelete:CASCADE` for `ProjectID`:

```go
// models.go:98
ProjectID string `gorm:"type:varchar(36);not null;index;constraint:OnDelete:CASCADE"`
```

**Risk:** Deleting a project will cascade delete all associated sessions, including failed ones that might be needed for recovery or audit.

**Mitigation:** Consider soft-delete or archive strategy for audit/recovery purposes.

---

#### Risk 5: VisualizationID Set NULL on Visualization Delete
**Severity:** Medium

```go
// models.go:99
VisualizationID *string `gorm:"type:varchar(36);index;constraint:OnDelete:SET NULL"`
```

**Risk:** When a visualization is deleted, sessions lose their association, making it harder to trace which visualization a session belonged to.

---

#### Risk 6: No Automatic Cleanup of Orphaned Sessions
**Severity:** Low

Sessions with `"default"` project ID or with `VisualizationID` set to NULL after cascade remain in the database indefinitely.

**Impact:** Database bloat over time, potential privacy concerns if sessions contain sensitive input data.

---

#### Risk 7: Frontend Mismatch with Backend API
**Severity:** Low

The frontend `apiClient.listHistory()` calls `/api/v1/history`, but the backend `ListHistory` expects both `project_id` AND `visualization_id`:

```go
// history.go:120-126
if req.ProjectID == "" || req.VisualizationID == "" {
    c.JSON(http.StatusOK, ListHistoryResponse{
        ProjectID: req.ProjectID,
        Versions:  []VersionResponse{},
    })
    return
}
```

**Issue:** The frontend hook does not pass `visualization_id`, resulting in empty responses. The hook should likely use `/api/v1/sessions/recent` instead.

---

### 5.2 Recommendations

1. **Add periodic checkpointing**: Persist snapshots at regular intervals during long-running stages, not just after completion.

2. **Implement session archival**: Before project deletion, archive sessions to a separate table or external storage.

3. **Fix frontend API mismatch**: Update `useHistory` to use `/api/v1/sessions/recent` endpoint with proper parameters.

4. **Add transaction boundaries**: Wrap all persistence operations in the `HistoryService` transaction manager.

5. **Implement session TTL**: Add automatic cleanup for orphaned sessions older than a configurable threshold.

6. **Add resume-from-any-stage capability**: Currently resume only finds the last completed stage; consider allowing resume from any checkpointed stage.

---

## 6. Summary

The history and recovery mechanism in paperbanana is well-architected with clear separation of concerns:

- **Persistence Layer**: Reliable SQLite storage with JSON serialization for complex state
- **Service Layer**: Transactional operations via `HistoryService` and `PersistentSnapshotStore`
- **API Layer**: RESTful endpoints for history access and session retrieval
- **Orchestrator Layer**: Real-time snapshot persistence and resume capability

The main areas for improvement are:
1. Handling crashes before first stage completion
2. Preventing data loss during project/visualization deletion
3. Fixing the frontend/backend API mismatch for history listing
4. Implementing cleanup for orphaned sessions

---

## Appendix: Key File Locations

| Component | File Path |
|-----------|-----------|
| Session Model | `internal/infrastructure/persistence/sqlite/models.go:96-131` |
| Session Repository | `internal/infrastructure/persistence/sqlite/session_repository.go` |
| Snapshot Store | `internal/infrastructure/persistence/sqlite/snapshot_store.go` |
| History Service | `internal/application/persistence/history_service.go` |
| History Handler | `internal/api/handlers/history.go` |
| Session Tracker | `internal/application/orchestrator/session.go` |
| Pipeline Runner | `internal/application/orchestrator/runner.go` |
| API Router | `internal/api/router.go:108-112` |
| Frontend Hook | `web/src/hooks/useHistory.ts` |
| Frontend Component | `web/src/components/HistorySidebar.tsx` |
| API Client | `web/src/lib/api.ts` |
