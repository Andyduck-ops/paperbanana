# Batch Persistence Recommendations

## Problem Statement

Batch processing results in `paperbanana-clean` are stored only in memory (`BatchRunner.results map[string]*domainagent.BatchResult`), which means:

1. **Data Loss**: All batch results are lost on server restart
2. **No Recovery**: Cannot resume interrupted batches
3. **ZIP Unavailable**: `DownloadBatchZip` returns 404 after restart

---

## Recommended Solutions

### Option A: Database Persistence (Recommended)

Add a `batch_results` table and implement the repository pattern.

#### 1. Database Model

```go
// internal/infrastructure/persistence/sqlite/models.go

// BatchResultModel is the GORM model for batch results.
type BatchResultModel struct {
    ID          string    `gorm:"primaryKey;type:varchar(36)"`
    ProjectID   string    `gorm:"type:varchar(36);index;constraint:OnDelete:CASCADE"`
    Status      string    `gorm:"type:varchar(50);not null;index"`
    Successful  int       `gorm:"not null"`
    Failed      int       `gorm:"not null"`
    ResultsJSON ResultsPayload `gorm:"type:json;serializer:json"`
    TimingJSON  TimingPayload  `gorm:"type:json;serializer:json"`
    CreatedAt   time.Time `gorm:"not null"`
    UpdatedAt   time.Time `gorm:"not null"`
    ExpiresAt   *time.Time `gorm:"index"` // For TTL cleanup
}

type ResultsPayload struct {
    Results []domainagent.CandidateResult `json:"results"`
}

type TimingPayload struct {
    StartedAt   time.Time `json:"started_at"`
    CompletedAt time.Time `json:"completed_at"`
    Duration    int64     `json:"duration_ms"`
}

func (BatchResultModel) TableName() string {
    return "batch_results"
}
```

#### 2. Repository Interface

```go
// internal/domain/batch/repository.go

type BatchRepository interface {
    Create(ctx context.Context, batch *BatchRecord) error
    GetByID(ctx context.Context, batchID string) (*BatchRecord, error)
    Update(ctx context.Context, batch *BatchRecord) error
    Delete(ctx context.Context, batchID string) error
    ListByProject(ctx context.Context, projectID string, limit int) ([]*BatchRecord, error)
    DeleteExpired(ctx context.Context) error // Cleanup job
}
```

#### 3. BatchRunner Integration

```go
// internal/application/orchestrator/batch_runner.go

type BatchRunner struct {
    agentFactory  AgentFactory
    maxConcurrent int
    eventBuffer   int
    batchRepo     batch.BatchRepository  // NEW: Inject repository
    results       map[string]*domainagent.BatchResult  // Keep for caching
    mu            sync.RWMutex
}

func (r *BatchRunner) storeResult(batchID string, result *domainagent.BatchResult) error {
    // 1. Store in cache (for quick access)
    r.mu.Lock()
    r.results[batchID] = result
    r.mu.Unlock()

    // 2. Persist to database
    return r.batchRepo.Create(context.Background(), &batch.BatchRecord{
        ID:         batchID,
        Status:     "completed",
        Successful: result.Successful,
        Failed:     result.Failed,
        Results:    result.Results,
        Timing:     result.Timing,
    })
}

func (r *BatchRunner) GetBatchResult(batchID string) (*domainagent.BatchResult, error) {
    // 1. Check cache first
    r.mu.RLock()
    if result, ok := r.results[batchID]; ok {
        r.mu.RUnlock()
        return result, nil
    }
    r.mu.RUnlock()

    // 2. Fall back to database
    record, err := r.batchRepo.GetByID(context.Background(), batchID)
    if err != nil {
        return nil, fmt.Errorf("batch result not found: %s", batchID)
    }

    return &domainagent.BatchResult{
        BatchID:    record.ID,
        Results:    record.Results,
        Successful: record.Successful,
        Failed:     record.Failed,
        Timing:     record.Timing,
    }, nil
}
```

#### 4. Migration Script

```sql
-- migrations/XXX_add_batch_results.sql

CREATE TABLE batch_results (
    id VARCHAR(36) PRIMARY KEY,
    project_id VARCHAR(36) REFERENCES projects(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL,
    successful INTEGER NOT NULL DEFAULT 0,
    failed INTEGER NOT NULL DEFAULT 0,
    results_json JSON NOT NULL,
    timing_json JSON NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    expires_at TIMESTAMP
);

CREATE INDEX idx_batch_results_project ON batch_results(project_id);
CREATE INDEX idx_batch_results_status ON batch_results(status);
CREATE INDEX idx_batch_results_expires ON batch_results(expires_at);
```

---

### Option B: Hybrid File + Database Approach

Combine the simplicity of file storage with database indexing.

#### Implementation

```go
// internal/infrastructure/persistence/batch_file_store.go

type BatchFileStore struct {
    baseDir string
    db      *gorm.DB
}

func (s *BatchFileStore) Save(batchID string, result *domainagent.BatchResult) error {
    // 1. Save full result to file
    filePath := filepath.Join(s.baseDir, batchID+".json")
    data, err := json.Marshal(result)
    if err != nil {
        return err
    }
    if err := os.WriteFile(filePath, data, 0644); err != nil {
        return err
    }

    // 2. Save index to database
    return s.db.Create(&BatchIndexModel{
        ID:        batchID,
        FilePath:  filePath,
        CreatedAt: time.Now(),
        ExpiresAt: time.Now().Add(24 * time.Hour),
    }).Error
}

func (s *BatchFileStore) Load(batchID string) (*domainagent.BatchResult, error) {
    var index BatchIndexModel
    if err := s.db.Where("id = ?", batchID).First(&index).Error; err != nil {
        return nil, err
    }

    data, err := os.ReadFile(index.FilePath)
    if err != nil {
        return nil, err
    }

    var result domainagent.BatchResult
    return &result, json.Unmarshal(data, &result)
}
```

---

### Option C: Redis Cache (For High-Scale Deployments)

Use Redis for distributed caching with TTL.

```go
// internal/infrastructure/cache/redis_batch_store.go

type RedisBatchStore struct {
    client *redis.Client
    ttl    time.Duration
}

func (s *RedisBatchStore) Save(batchID string, result *domainagent.BatchResult) error {
    data, err := json.Marshal(result)
    if err != nil {
        return err
    }
    return s.client.Set(context.Background(),
        "batch:"+batchID, data, s.ttl).Err()
}

func (s *RedisBatchStore) Load(batchID string) (*domainagent.BatchResult, error) {
    data, err := s.client.Get(context.Background(), "batch:"+batchID).Bytes()
    if err != nil {
        return nil, err
    }

    var result domainagent.BatchResult
    return &result, json.Unmarshal(data, &result)
}
```

---

## Implementation Priority

| Option | Effort | Persistence | Scalability | Recommended For |
|--------|--------|-------------|-------------|-----------------|
| A: Database | Medium | Full | Good | Single-server deployments |
| B: Hybrid | Low | Full | Good | Quick fix, minimal changes |
| C: Redis | Medium | TTL-limited | Excellent | Multi-instance deployments |

**Recommendation**: Start with **Option A** (Database Persistence) as it:
- Integrates with existing SQLite infrastructure
- Provides full persistence without TTL limits
- Supports batch history and analytics
- Enables cleanup via `expires_at` column

---

## Additional Improvements

### 1. Batch Progress Persistence

Store progress events for restart recovery:

```go
type BatchProgressModel struct {
    ID          string    `gorm:"primaryKey"`
    BatchID     string    `gorm:"index"`
    EventType   string    `gorm:"not null"`
    CandidateID int       `gorm:"not null"`
    Status      string    `gorm:"not null"`
    OccurredAt  time.Time `gorm:"not null"`
    Metadata    JSON      `gorm:"type:json"`
}
```

### 2. Batch to Session Mapping

Track which sessions belong to which batch:

```go
type BatchSessionModel struct {
    BatchID    string `gorm:"primaryKey;index"`
    SessionID  string `gorm:"primaryKey;index"`
    CandidateID int   `gorm:"not null"`
}
```

This enables:
- Reconstructing batch results from sessions
- Querying all sessions in a batch
- Better analytics

### 3. Cleanup Job

Add scheduled cleanup for expired batch results:

```go
func (r *BatchRunner) CleanupExpired(ctx context.Context) error {
    return r.batchRepo.DeleteExpired(ctx)
}

// Run daily
ticker := time.NewTicker(24 * time.Hour)
go func() {
    for range ticker.C {
        runner.CleanupExpired(context.Background())
    }
}()
```

---

## Testing Strategy

1. **Unit Tests**: Repository CRUD operations
2. **Integration Tests**: Full batch lifecycle with persistence
3. **Restart Tests**: Simulate server restart and verify recovery
4. **Performance Tests**: Large batch sizes (50+ candidates)

```go
func TestBatchPersistence(t *testing.T) {
    // Create batch
    runner := NewBatchRunner(factory, WithBatchRepository(repo))
    handle, _ := runner.StartBatch(ctx, inputs)
    result, _ := handle.Wait()

    // Simulate restart
    newRunner := NewBatchRunner(factory, WithBatchRepository(repo))

    // Verify recovery
    recovered, err := newRunner.GetBatchResult(result.BatchID)
    assert.NoError(t, err)
    assert.Equal(t, result.BatchID, recovered.BatchID)
}
```
