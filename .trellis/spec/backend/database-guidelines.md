# Database Guidelines

> SQLite/GORM patterns and conventions for PaperBanana.

---

## Overview

PaperBanana uses **SQLite** as its sole relational database, accessed through **GORM** with the pure-Go `github.com/glebarez/sqlite` driver. There are no external database dependencies; the database file is created automatically on first run.

---

## Technology Stack

| Component | Library | Notes |
|-----------|---------|-------|
| Database | SQLite | Single-file, zero-config |
| ORM | GORM (`gorm.io/gorm`) | Auto-migration, hooks, associations |
| Driver | `github.com/glebarez/sqlite` | Pure Go (no CGo required for this driver) |
| Alt Driver | `github.com/mattn/go-sqlite3` | Present in go.mod but not used for primary connection |
| Encryption | AES-256-GCM + Argon2id | For API key storage at rest |

---

## Bootstrap and Migration

Database initialization is handled by `internal/infrastructure/persistence/sqlite/bootstrap.go`. There is no standalone migration tool; GORM `AutoMigrate` runs on every server start.

```go
// internal/infrastructure/persistence/sqlite/bootstrap.go
func Bootstrap(ctx context.Context, cfg BootstrapConfig) (*BootstrapResult, error) {
    db, err := gorm.Open(sqlite.Open(cfg.DatabasePath), &gorm.Config{})
    // Apply PRAGMAs (foreign_keys, busy_timeout, WAL)
    // Run AutoMigrate for all models
    if err := db.WithContext(ctx).AutoMigrate(AllModels()...); err != nil {
        return nil, fmt.Errorf("auto migrate: %w", err)
    }
    return &BootstrapResult{DB: db}, nil
}
```

**All models** are registered in `models.go` via `AllModels()`:

```go
// internal/infrastructure/persistence/sqlite/models.go
func AllModels() []interface{} {
    return []interface{}{
        &ProjectModel{},
        &FolderModel{},
        &VisualizationModel{},
        &VisualizationVersionModel{},
        &VersionArtifactModel{},
        &SessionModel{},
        &AssetModel{},
        &ProviderModel{},
        &APIKeyModel{},
        &BatchResultModel{},
    }
}
```

### PRAGMA Configuration

Controlled via `BootstrapConfig` and `internal/config/config.go`:

| PRAGMA | Default | Config Key |
|--------|---------|------------|
| `foreign_keys` | ON | `persistence.enable_foreign_keys` |
| `busy_timeout` | 5000ms | `persistence.busy_timeout_ms` |
| `journal_mode` | WAL | `persistence.enable_wal` |
| `synchronous` | NORMAL | (set when WAL enabled) |
| `cache_size` | -64000 | (set when WAL enabled) |
| `wal_autocheckpoint` | 1000 | (set when WAL enabled) |

**Known issue**: WAL mode default was changed to `true` in config defaults. Previous versions ran without WAL. If you see locking issues, verify `enable_wal` in your config.

---

## Repository Pattern

Interfaces are defined in the domain layer; implementations live in the infrastructure layer. This allows the application layer to depend on abstractions.

### Interface Definition (Domain)

```go
// internal/domain/workspace/repositories.go
type ProjectRepository interface {
    Create(ctx context.Context, project *Project) error
    GetByID(ctx context.Context, id string) (*Project, error)
    List(ctx context.Context) ([]*Project, error)
    Update(ctx context.Context, project *Project) error
    Delete(ctx context.Context, id string) error
}
```

### Implementation (Infrastructure)

```go
// internal/infrastructure/persistence/sqlite/workspace_repository.go
type ProjectRepository struct { db *gorm.DB }

func NewProjectRepository(db *gorm.DB) *ProjectRepository {
    return &ProjectRepository{db: db}
}
```

### Compile-Time Interface Checks

All implementations include compile-time assertions:

```go
var _ workspace.ProjectRepository = (*ProjectRepository)(nil)
```

---

## Transaction Management

Transactions are managed by `TxManager` in `internal/infrastructure/persistence/sqlite/tx_manager.go`. It wraps GORM's `Transaction()` method and provides a `Repositories` factory so that all repository operations within a transaction use the same `*gorm.DB` instance.

```go
// internal/infrastructure/persistence/sqlite/tx_manager.go
func (m *TxManager) RunInTx(ctx context.Context, fn func(repos persistence.Repositories) error) error {
    return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        repos := newRepositories(tx)
        return fn(repos)
    })
}
```

The `Repositories` interface provides lazy-initialized repository instances for each aggregate:

```go
type Repositories interface {
    Projects() workspace.ProjectRepository
    Folders() workspace.FolderRepository
    Visualizations() workspace.VisualizationRepository
    Versions() workspace.VersionRepository
    Sessions() workspace.SessionRepository
    Assets() workspace.AssetRepository
}
```

### Read-Only Transactions

`ReadOnlyTx` exists but is a thin wrapper; SQLite does not have true read-only transactions. It provides consistent read snapshots within a standard transaction.

---

## Session Storage

Session state is persisted as **JSON blobs** in the `sessions` table. The `SessionModel.SnapshotJSON` field uses GORM's JSON serializer to store the full `domainagent.SessionState` struct:

```go
// internal/infrastructure/persistence/sqlite/models.go
type SessionModel struct {
    ID              string                 `gorm:"primaryKey;type:varchar(36)"`
    ProjectID       string                 `gorm:"type:varchar(36);not null;index"`
    VisualizationID *string               `gorm:"type:varchar(36);index"`
    Status          string                 `gorm:"type:varchar(50);not null;index"`
    SnapshotJSON    SessionSnapshotPayload `gorm:"type:json;serializer:json;not null"`
    // ...
}
```

The `PersistentSnapshotStore` (`internal/infrastructure/persistence/sqlite/snapshot_store.go`) bridges the orchestrator's `SnapshotStore` interface with the session repository, enabling pipeline resume after restart.

---

## Asset Storage (Dual Implementation)

Assets use a **split architecture**: metadata in SQLite, bytes on the filesystem.

### Database Metadata

```go
// AssetModel tracks metadata
type AssetModel struct {
    ID              string `gorm:"primaryKey"`
    ProjectID       string `gorm:"not null;index"`
    StorageKey      string `gorm:"not null;uniqueIndex"`
    ByteSize        int64  `gorm:"not null"`
    ChecksumSHA256  string `gorm:"type:varchar(64);not null"`
    // ...
}
```

### Filesystem Bytes

Two implementations exist:

1. **`internal/infrastructure/persistence/sqlite/local_asset_store.go`** -- Simple `os.WriteFile`/`os.ReadFile` under a configurable root directory. Used by the persistence layer.

2. **`internal/infrastructure/assets/localstore/store.go`** -- More full-featured store with path traversal protection, SHA-256 verification, and UUID-based opaque keys. Used by the asset handler.

**Known issue**: These are duplicate implementations. The `localstore` package has better security (path traversal protection). The `sqlite/local_asset_store.go` is simpler and lacks path traversal checks.

---

## GORM Model Conventions

- **Table names**: Explicit `TableName()` method on every model (plural, lowercase)
- **Primary keys**: `varchar(36)` string IDs (UUIDs)
- **Soft deletes**: `gorm.DeletedAt` field on Project, Folder, Visualization, Asset models
- **Foreign keys**: `constraint:OnDelete:CASCADE` or `OnDelete:SET NULL` as appropriate
- **Timestamps**: `CreatedAt` and `UpdatedAt` on every model, `CompletedAt` as nullable pointer where needed
- **JSON columns**: GORM's `serializer:json` tag for structured data (session snapshots, batch results)
- **Indexes**: Explicit indexes on foreign keys, status fields, and unique constraints

---

## Query Patterns

### Standard CRUD

Repositories follow a consistent pattern: accept context, return domain entity or error.

```go
func (r *ProjectRepository) GetByID(ctx context.Context, id string) (*workspace.Project, error) {
    var model ProjectModel
    if err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, fmt.Errorf("project %s: %w", id, workspace.ErrNotFound)
        }
        return nil, fmt.Errorf("get project %s: %w", id, err)
    }
    return toDomainProject(&model), nil
}
```

### Recursive Queries

The `FolderRepository.GetDescendantIDs` uses a recursive CTE for efficient tree traversal:

```go
// Uses WITH RECURSIVE for subtree queries
```

---

## Common Mistakes

1. **Forgetting project-scoped queries**: All workspace queries must include `project_id` filtering. The `ProjectScopedQuery` interface exists for compile-time verification.

2. **Using the wrong asset store**: New code should prefer `internal/infrastructure/assets/localstore/` which has path traversal protection, not `internal/infrastructure/persistence/sqlite/local_asset_store.go`.

3. **Not wrapping errors with context**: Always add context when propagating GORM errors (see Error Handling spec).

4. **Ignoring `DeletedAt` in queries**: Soft-deleted records are automatically excluded by GORM unless `Unscoped()` is used.

5. **Direct GORM usage outside infrastructure**: Application services should use repository interfaces, not `*gorm.DB` directly. Only the `TxManager` and repository implementations should touch GORM.

6. **Creating new `*gorm.DB` connections in tests**: Use the `Bootstrap` function for test setup, which applies PRAGMAs and migrations correctly.
