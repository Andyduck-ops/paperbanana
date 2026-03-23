package sqlite

import (
	"context"

	"github.com/paperbanana/paperbanana/internal/application/persistence"
	"github.com/paperbanana/paperbanana/internal/domain/workspace"
	"gorm.io/gorm"
)

// TxManager implements persistence.TxManager using GORM transactions.
type TxManager struct {
	db *gorm.DB
}

// NewTxManager creates a new TxManager.
func NewTxManager(db *gorm.DB) *TxManager {
	return &TxManager{db: db}
}

// RunInTx executes fn within a transaction. If fn returns an error,
// the transaction is rolled back; otherwise it is committed.
func (m *TxManager) RunInTx(ctx context.Context, fn func(repos persistence.Repositories) error) error {
	return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		repos := newRepositories(tx)
		return fn(repos)
	})
}

// ReadOnlyTx executes fn within a read-only transaction context.
// For SQLite, this uses a deferred transaction which allows reads but
// will fail on writes if the database is in read-only mode.
func (m *TxManager) ReadOnlyTx(ctx context.Context, fn func(repos persistence.Repositories) error) error {
	// SQLite doesn't have true read-only transactions, but we use
	// a standard transaction which provides consistent read snapshots.
	return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		repos := newRepositories(tx)
		return fn(repos)
	})
}

// repositories is the concrete implementation of Repositories interface.
type repositories struct {
	db    *gorm.DB
	proj  *ProjectRepository
	fold  *FolderRepository
	viz   *VisualizationRepository
	ver   *VersionRepository
	sess  *SessionRepository
	asset *AssetRepository
}

// newRepositories creates a new repositories instance with the given DB.
func newRepositories(db *gorm.DB) *repositories {
	return &repositories{db: db}
}

// Projects returns the project repository.
func (r *repositories) Projects() workspace.ProjectRepository {
	if r.proj == nil {
		r.proj = NewProjectRepository(r.db)
	}
	return r.proj
}

// Folders returns the folder repository.
func (r *repositories) Folders() workspace.FolderRepository {
	if r.fold == nil {
		r.fold = NewFolderRepository(r.db)
	}
	return r.fold
}

// Visualizations returns the visualization repository.
func (r *repositories) Visualizations() workspace.VisualizationRepository {
	if r.viz == nil {
		r.viz = NewVisualizationRepository(r.db)
	}
	return r.viz
}

// Versions returns the version repository.
func (r *repositories) Versions() workspace.VersionRepository {
	if r.ver == nil {
		r.ver = NewVersionRepository(r.db)
	}
	return r.ver
}

// Sessions returns the session repository.
func (r *repositories) Sessions() workspace.SessionRepository {
	if r.sess == nil {
		r.sess = NewSessionRepository(r.db)
	}
	return r.sess
}

// Assets returns the asset repository.
func (r *repositories) Assets() workspace.AssetRepository {
	if r.asset == nil {
		r.asset = NewAssetRepository(r.db)
	}
	return r.asset
}

// Ensure TxManager implements the persistence.TxManager interface.
var _ persistence.TxManager = (*TxManager)(nil)

// Ensure repositories implements persistence.Repositories interface.
var _ persistence.Repositories = (*repositories)(nil)
