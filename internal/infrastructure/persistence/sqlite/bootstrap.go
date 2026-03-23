package sqlite

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// BootstrapConfig holds the configuration for SQLite bootstrap.
type BootstrapConfig struct {
	// DatabasePath is the path to the SQLite database file.
	DatabasePath string

	// EnableForeignKeys enables foreign key enforcement (recommended: true).
	EnableForeignKeys bool

	// BusyTimeoutMs sets the busy timeout in milliseconds for SQLite locks.
	BusyTimeoutMs int

	// EnableWAL enables write-ahead logging mode.
	// Deferred until SQLite version verification is complete.
	EnableWAL bool
}

// BootstrapResult holds the result of a successful bootstrap.
type BootstrapResult struct {
	DB *gorm.DB
}

// Bootstrap opens a SQLite database through GORM, applies required PRAGMAs,
// and runs AutoMigrate over all Phase 3 models.
func Bootstrap(ctx context.Context, cfg BootstrapConfig) (*BootstrapResult, error) {
	if err := ensureDatabaseDir(cfg.DatabasePath); err != nil {
		return nil, fmt.Errorf("ensure database directory: %w", err)
	}

	db, err := gorm.Open(sqlite.Open(cfg.DatabasePath), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("open sqlite database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get underlying sql.DB: %w", err)
	}

	// Apply PRAGMAs
	if cfg.EnableForeignKeys {
		if _, err := sqlDB.ExecContext(ctx, "PRAGMA foreign_keys = ON;"); err != nil {
			return nil, fmt.Errorf("enable foreign_keys pragma: %w", err)
		}
	}

	if cfg.BusyTimeoutMs > 0 {
		pragma := fmt.Sprintf("PRAGMA busy_timeout = %d;", cfg.BusyTimeoutMs)
		if _, err := sqlDB.ExecContext(ctx, pragma); err != nil {
			return nil, fmt.Errorf("set busy_timeout pragma: %w", err)
		}
	}

	// WAL mode is gated behind explicit configuration
	if cfg.EnableWAL {
		if _, err := sqlDB.ExecContext(ctx, "PRAGMA journal_mode = WAL;"); err != nil {
			return nil, fmt.Errorf("enable WAL mode: %w", err)
		}
	}

	// Run AutoMigrate for all Phase 3 models
	if err := db.WithContext(ctx).AutoMigrate(AllModels()...); err != nil {
		return nil, fmt.Errorf("auto migrate: %w", err)
	}

	return &BootstrapResult{DB: db}, nil
}

// ensureDatabaseDir creates the parent directory for the database file if it doesn't exist.
func ensureDatabaseDir(dbPath string) error {
	dir := filepath.Dir(dbPath)
	if dir == "" || dir == "." {
		return nil
	}
	return os.MkdirAll(dir, 0o755)
}

// Close closes the database connection.
func Close(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("get underlying sql.DB: %w", err)
	}
	return sqlDB.Close()
}
