package sqlite

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestBootstrapSQLiteWithGORM(t *testing.T) {
	t.Run("creates database file", func(t *testing.T) {
		ctx := context.Background()
		dbPath := filepath.Join(t.TempDir(), "test.db")

		result, err := Bootstrap(ctx, BootstrapConfig{
			DatabasePath:      dbPath,
			EnableForeignKeys: true,
			BusyTimeoutMs:     5000,
			EnableWAL:         false,
		})
		require.NoError(t, err)
		require.NotNil(t, result)
		require.NotNil(t, result.DB)

		// Verify file exists
		_, err = os.Stat(dbPath)
		require.NoError(t, err)

		// Cleanup
		require.NoError(t, Close(result.DB))
	})

	t.Run("enables foreign keys pragma", func(t *testing.T) {
		ctx := context.Background()
		dbPath := filepath.Join(t.TempDir(), "test.db")

		result, err := Bootstrap(ctx, BootstrapConfig{
			DatabasePath:      dbPath,
			EnableForeignKeys: true,
			BusyTimeoutMs:     5000,
			EnableWAL:         false,
		})
		require.NoError(t, err)

		var foreignKeysEnabled int
		err = result.DB.Raw("PRAGMA foreign_keys;").Scan(&foreignKeysEnabled).Error
		require.NoError(t, err)
		assert.Equal(t, 1, foreignKeysEnabled, "foreign_keys pragma should be enabled")

		require.NoError(t, Close(result.DB))
	})

	t.Run("sets busy timeout pragma", func(t *testing.T) {
		ctx := context.Background()
		dbPath := filepath.Join(t.TempDir(), "test.db")

		result, err := Bootstrap(ctx, BootstrapConfig{
			DatabasePath:      dbPath,
			EnableForeignKeys: true,
			BusyTimeoutMs:     5000,
			EnableWAL:         false,
		})
		require.NoError(t, err)

		var busyTimeout int
		err = result.DB.Raw("PRAGMA busy_timeout;").Scan(&busyTimeout).Error
		require.NoError(t, err)
		assert.Equal(t, 5000, busyTimeout, "busy_timeout should be set to 5000ms")

		require.NoError(t, Close(result.DB))
	})

	t.Run("creates parent directory if needed", func(t *testing.T) {
		ctx := context.Background()
		tmpDir := t.TempDir()
		dbPath := filepath.Join(tmpDir, "nested", "dir", "test.db")

		result, err := Bootstrap(ctx, BootstrapConfig{
			DatabasePath:      dbPath,
			EnableForeignKeys: true,
			BusyTimeoutMs:     5000,
			EnableWAL:         false,
		})
		require.NoError(t, err)
		require.NotNil(t, result)

		_, err = os.Stat(dbPath)
		require.NoError(t, err)

		require.NoError(t, Close(result.DB))
	})

	t.Run("can enable WAL mode", func(t *testing.T) {
		ctx := context.Background()
		dbPath := filepath.Join(t.TempDir(), "test.db")

		result, err := Bootstrap(ctx, BootstrapConfig{
			DatabasePath:      dbPath,
			EnableForeignKeys: true,
			BusyTimeoutMs:     5000,
			EnableWAL:         true,
		})
		require.NoError(t, err)

		var journalMode string
		err = result.DB.Raw("PRAGMA journal_mode;").Scan(&journalMode).Error
		require.NoError(t, err)
		assert.Equal(t, "wal", journalMode, "journal_mode should be wal")

		require.NoError(t, Close(result.DB))
	})
}

func TestAutoMigratedSchemaContainsPhase3Tables(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "test.db")

	result, err := Bootstrap(ctx, BootstrapConfig{
		DatabasePath:      dbPath,
		EnableForeignKeys: true,
		BusyTimeoutMs:     5000,
		EnableWAL:         false,
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	defer Close(result.DB)

	// Check all Phase 3 tables exist
	expectedTables := []string{
		"projects",
		"folders",
		"visualizations",
		"visualization_versions",
		"version_artifacts",
		"sessions",
		"assets",
	}

	for _, table := range expectedTables {
		var count int64
		err := result.DB.Raw(
			"SELECT count(*) FROM sqlite_master WHERE type = 'table' AND name = ?",
			table,
		).Scan(&count).Error
		require.NoError(t, err)
		assert.Equal(t, int64(1), count, "table %s should exist", table)
	}
}

func TestSchemaHasExpectedColumns(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "test.db")

	result, err := Bootstrap(ctx, BootstrapConfig{
		DatabasePath:      dbPath,
		EnableForeignKeys: true,
		BusyTimeoutMs:     5000,
		EnableWAL:         false,
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	defer Close(result.DB)

	t.Run("projects table has expected columns", func(t *testing.T) {
		columns := getTableColumns(t, result.DB, "projects")
		expectedColumns := []string{"id", "name", "description", "created_at", "updated_at", "deleted_at"}
		for _, col := range expectedColumns {
			assert.Contains(t, columns, col, "projects table should have column %s", col)
		}
	})

	t.Run("folders table has expected columns", func(t *testing.T) {
		columns := getTableColumns(t, result.DB, "folders")
		expectedColumns := []string{"id", "project_id", "parent_id", "name", "created_at", "updated_at", "deleted_at"}
		for _, col := range expectedColumns {
			assert.Contains(t, columns, col, "folders table should have column %s", col)
		}
	})

	t.Run("visualizations table has expected columns", func(t *testing.T) {
		columns := getTableColumns(t, result.DB, "visualizations")
		expectedColumns := []string{"id", "project_id", "folder_id", "name", "current_version_id", "created_at", "updated_at", "deleted_at"}
		for _, col := range expectedColumns {
			assert.Contains(t, columns, col, "visualizations table should have column %s", col)
		}
	})

	t.Run("sessions table has expected columns", func(t *testing.T) {
		columns := getTableColumns(t, result.DB, "sessions")
		expectedColumns := []string{"id", "project_id", "visualization_id", "status", "current_stage", "schema_version", "snapshot_json", "created_at", "updated_at", "completed_at"}
		for _, col := range expectedColumns {
			assert.Contains(t, columns, col, "sessions table should have column %s", col)
		}
	})

	t.Run("assets table has expected columns", func(t *testing.T) {
		columns := getTableColumns(t, result.DB, "assets")
		expectedColumns := []string{"id", "project_id", "visualization_id", "version_id", "mime_type", "storage_key", "byte_size", "checksum_sha256", "created_at", "deleted_at"}
		for _, col := range expectedColumns {
			assert.Contains(t, columns, col, "assets table should have column %s", col)
		}
	})
}

func TestSchemaHasForeignKeys(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "test.db")

	result, err := Bootstrap(ctx, BootstrapConfig{
		DatabasePath:      dbPath,
		EnableForeignKeys: true,
		BusyTimeoutMs:     5000,
		EnableWAL:         false,
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	defer Close(result.DB)

	// Verify foreign keys enforcement is enabled
	var fkEnabled int
	err = result.DB.Raw("PRAGMA foreign_keys;").Scan(&fkEnabled).Error
	require.NoError(t, err)
	assert.Equal(t, 1, fkEnabled, "foreign_keys pragma should be enabled")
}

func getTableColumns(t *testing.T, db *gorm.DB, tableName string) []string {
	t.Helper()
	var columns []string
	rows, err := db.Raw("PRAGMA table_info(" + tableName + ")").Rows()
	require.NoError(t, err)
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dfltValue interface{}
		err := rows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk)
		require.NoError(t, err)
		columns = append(columns, name)
	}
	return columns
}
