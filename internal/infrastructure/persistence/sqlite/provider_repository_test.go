package sqlite

import (
	"context"
	"path/filepath"
	"testing"

	domainconfig "github.com/paperbanana/paperbanana/internal/domain/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProviderRepositoryCreateRestoresSoftDeletedProvider(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "providers.db")

	result, err := Bootstrap(ctx, BootstrapConfig{
		DatabasePath:      dbPath,
		EnableForeignKeys: true,
		BusyTimeoutMs:     5000,
		EnableWAL:         false,
	})
	require.NoError(t, err)
	defer func() {
		require.NoError(t, Close(result.DB))
	}()

	repo := NewProviderRepository(result.DB)

	original := &domainconfig.Provider{
		Name:        "grok",
		Type:        domainconfig.ProviderTypeGrok,
		DisplayName: "xAI Grok",
		APIHost:     "https://api.x.ai/v1",
		Enabled:     false,
		IsSystem:    true,
		Models: []domainconfig.ModelInfo{
			{ID: "grok-4.1-fast", Name: "Grok 4.1 Fast", Enabled: true},
		},
	}
	require.NoError(t, repo.Create(original))
	require.NotEmpty(t, original.ID)

	require.NoError(t, repo.Delete(original.ID))

	recreated := &domainconfig.Provider{
		Name:        "grok",
		Type:        domainconfig.ProviderTypeGrok,
		DisplayName: "xAI Grok",
		APIHost:     "https://lx.lxsummer.cloud/v1",
		Enabled:     true,
		IsSystem:    true,
		QueryModel:  "grok-4.1-fast",
		GenModel:    "grok-imagine-1.0",
		Models: []domainconfig.ModelInfo{
			{ID: "grok-4.1-fast", Name: "Grok 4.1 Fast", Enabled: true},
			{ID: "grok-imagine-1.0", Name: "Grok Imagine 1.0", SupportsVision: true, Enabled: true},
		},
	}
	require.NoError(t, repo.Create(recreated))

	assert.Equal(t, original.ID, recreated.ID)

	restored, err := repo.GetByName("grok")
	require.NoError(t, err)
	assert.Equal(t, original.ID, restored.ID)
	assert.Equal(t, "https://lx.lxsummer.cloud/v1", restored.APIHost)
	assert.True(t, restored.Enabled)
	assert.Equal(t, "grok-4.1-fast", restored.QueryModel)
	assert.Equal(t, "grok-imagine-1.0", restored.GenModel)
	require.Len(t, restored.Models, 2)

	var count int64
	require.NoError(t, result.DB.Table("providers").Where("name = ? AND deleted_at IS NULL", "grok").Count(&count).Error)
	assert.Equal(t, int64(1), count)
}
