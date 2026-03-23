package persistence

import (
	"context"

	"github.com/paperbanana/paperbanana/internal/domain/workspace"
)

// Repositories provides access to all domain repositories within a transaction.
// Application services receive this bundle from TxManager.RunInTx.
type Repositories interface {
	Projects() workspace.ProjectRepository
	Folders() workspace.FolderRepository
	Visualizations() workspace.VisualizationRepository
	Versions() workspace.VersionRepository
	Sessions() workspace.SessionRepository
	Assets() workspace.AssetRepository
}

// TxManager provides transactional boundaries for persistence operations.
// All writes that span multiple aggregates must go through RunInTx.
type TxManager interface {
	// RunInTx executes fn within a transaction. If fn returns an error,
	// the transaction is rolled back; otherwise it is committed.
	RunInTx(ctx context.Context, fn func(Repositories) error) error

	// ReadOnlyTx executes fn within a read-only transaction context.
	// Write operations will fail; use for consistent read snapshots.
	ReadOnlyTx(ctx context.Context, fn func(Repositories) error) error
}

// AssetStore manages file bytes for assets.
// Implementations store files under an opaque key and provide read/write/delete access.
type AssetStore interface {
	// Write stores the provided bytes under the given storage key.
	Write(ctx context.Context, storageKey string, data []byte) error

	// Read retrieves the bytes for the given storage key.
	Read(ctx context.Context, storageKey string) ([]byte, error)

	// Delete removes the file for the given storage key.
	Delete(ctx context.Context, storageKey string) error

	// Exists checks whether a file exists for the storage key.
	Exists(ctx context.Context, storageKey string) (bool, error)
}
