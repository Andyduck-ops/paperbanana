package sqlite

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	domainagent "github.com/paperbanana/paperbanana/internal/domain/agent"
	"github.com/paperbanana/paperbanana/internal/domain/workspace"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVisualizationVersionRepository_PublishNextVersion(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "test.db")

	result, err := Bootstrap(ctx, BootstrapConfig{
		DatabasePath:      dbPath,
		EnableForeignKeys: true,
		BusyTimeoutMs:     5000,
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	defer Close(result.DB)

	// Create parent project and visualization
	projectID := uuid.NewString()
	vizID := uuid.NewString()

	require.NoError(t, result.DB.Create(&ProjectModel{
		ID:        projectID,
		Name:      "Test Project",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}).Error)

	require.NoError(t, result.DB.Create(&VisualizationModel{
		ID:        vizID,
		ProjectID: projectID,
		Name:      "Test Visualization",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}).Error)

	repo := NewVersionRepository(result.DB)

	now := time.Now().UTC()

	// Publish first version
	sessionID1 := uuid.NewString()
	version1 := &workspace.VisualizationVersion{
		ID:              uuid.NewString(),
		VisualizationID: vizID,
		ProjectID:       projectID,
		VersionNumber:   1,
		SessionID:       sessionID1,
		Summary:         "Initial version",
		CreatedAt:       now,
		SessionSnapshot: &domainagent.SessionState{
			SessionID:     sessionID1,
			SchemaVersion: "1.0.0",
			Status:        domainagent.StatusCompleted,
			StartedAt:     now,
		},
	}

	err = repo.Create(ctx, version1)
	require.NoError(t, err)

	// Verify version was created
	loaded, err := repo.GetByID(ctx, projectID, version1.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, loaded.VersionNumber)
	assert.Equal(t, "Initial version", loaded.Summary)
	assert.Equal(t, vizID, loaded.VisualizationID)
	assert.Equal(t, projectID, loaded.ProjectID)

	// Publish second version
	sessionID2 := uuid.NewString()
	version2 := &workspace.VisualizationVersion{
		ID:              uuid.NewString(),
		VisualizationID: vizID,
		ProjectID:       projectID,
		VersionNumber:   2,
		SessionID:       sessionID2,
		Summary:         "Updated visualization",
		CreatedAt:       now.Add(time.Hour),
		SessionSnapshot: &domainagent.SessionState{
			SessionID:     sessionID2,
			SchemaVersion: "1.0.0",
			Status:        domainagent.StatusCompleted,
			StartedAt:     now.Add(time.Hour),
		},
	}

	err = repo.Create(ctx, version2)
	require.NoError(t, err)

	// List versions
	versions, err := repo.ListByVisualization(ctx, projectID, vizID, 10)
	require.NoError(t, err)
	assert.Len(t, versions, 2)

	// Verify ordering (most recent first)
	assert.Equal(t, 2, versions[0].VersionNumber)
	assert.Equal(t, 1, versions[1].VersionNumber)

	// Get latest version
	latest, err := repo.GetLatestByVisualization(ctx, projectID, vizID)
	require.NoError(t, err)
	assert.Equal(t, 2, latest.VersionNumber)
	assert.Equal(t, "Updated visualization", latest.Summary)
}

func TestVisualizationVersionRepository_DoesNotVersionFailedRuns(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "test.db")

	result, err := Bootstrap(ctx, BootstrapConfig{
		DatabasePath:      dbPath,
		EnableForeignKeys: true,
		BusyTimeoutMs:     5000,
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	defer Close(result.DB)

	projectID := uuid.NewString()
	vizID := uuid.NewString()

	require.NoError(t, result.DB.Create(&ProjectModel{
		ID:        projectID,
		Name:      "Test Project",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}).Error)

	require.NoError(t, result.DB.Create(&VisualizationModel{
		ID:        vizID,
		ProjectID: projectID,
		Name:      "Test Visualization",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}).Error)

	sessionRepo := NewSessionRepository(result.DB)
	versionRepo := NewVersionRepository(result.DB)

	// Create a failed session (should NOT create a version)
	now := time.Now().UTC()
	failedSessionID := uuid.NewString()
	failedSession := &workspace.SessionRecord{
		ID:            failedSessionID,
		ProjectID:     projectID,
		VisualizationID: &vizID,
		Status:        string(domainagent.StatusFailed),
		CurrentStage:  string(domainagent.StagePlanner),
		SchemaVersion: "1.0.0",
		Snapshot: &domainagent.SessionState{
			SessionID:     failedSessionID,
			SchemaVersion: "1.0.0",
			Status:        domainagent.StatusFailed,
			CurrentStage:  domainagent.StagePlanner,
			StartedAt:     now,
			UpdatedAt:     now,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	require.NoError(t, sessionRepo.Create(ctx, failedSession))

	// Verify no versions exist for the visualization
	versions, err := versionRepo.ListByVisualization(ctx, projectID, vizID, 10)
	require.NoError(t, err)
	assert.Empty(t, versions, "failed runs should not create versions")

	// Verify the session exists (for audit)
	loadedSession, err := sessionRepo.GetByID(ctx, failedSessionID)
	require.NoError(t, err)
	assert.Equal(t, string(domainagent.StatusFailed), loadedSession.Status)
}

func TestVisualizationVersionRepository_VersionIsImmutable(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "test.db")

	result, err := Bootstrap(ctx, BootstrapConfig{
		DatabasePath:      dbPath,
		EnableForeignKeys: true,
		BusyTimeoutMs:     5000,
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	defer Close(result.DB)

	projectID := uuid.NewString()
	vizID := uuid.NewString()

	require.NoError(t, result.DB.Create(&ProjectModel{
		ID:        projectID,
		Name:      "Test Project",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}).Error)

	require.NoError(t, result.DB.Create(&VisualizationModel{
		ID:        vizID,
		ProjectID: projectID,
		Name:      "Test Visualization",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}).Error)

	repo := NewVersionRepository(result.DB)

	now := time.Now().UTC()
	sessionID := uuid.NewString()
	versionID := uuid.NewString()

	version := &workspace.VisualizationVersion{
		ID:              versionID,
		VisualizationID: vizID,
		ProjectID:       projectID,
		VersionNumber:   1,
		SessionID:       sessionID,
		Summary:         "Original summary",
		CreatedAt:       now,
		SessionSnapshot: &domainagent.SessionState{
			SessionID:     sessionID,
			SchemaVersion: "1.0.0",
			Status:        domainagent.StatusCompleted,
			StartedAt:     now,
		},
	}

	require.NoError(t, repo.Create(ctx, version))

	// Version repository should not have an Update method
	// (enforced by workspace.VersionRepository interface)

	// Verify we cannot create a duplicate version with same number
	duplicateVersion := &workspace.VisualizationVersion{
		ID:              uuid.NewString(),
		VisualizationID: vizID,
		ProjectID:       projectID,
		VersionNumber:   1, // Same version number
		SessionID:       uuid.NewString(),
		Summary:         "Duplicate attempt",
		CreatedAt:       now.Add(time.Hour),
	}

	err = repo.Create(ctx, duplicateVersion)
	assert.Error(t, err, "should not allow duplicate version numbers for same visualization")
}

func TestVisualizationVersionRepository_ProjectScoped(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "test.db")

	result, err := Bootstrap(ctx, BootstrapConfig{
		DatabasePath:      dbPath,
		EnableForeignKeys: true,
		BusyTimeoutMs:     5000,
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	defer Close(result.DB)

	// Create two projects
	project1ID := uuid.NewString()
	project2ID := uuid.NewString()
	viz1ID := uuid.NewString()
	viz2ID := uuid.NewString()

	require.NoError(t, result.DB.Create(&ProjectModel{
		ID:        project1ID,
		Name:      "Project 1",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}).Error)

	require.NoError(t, result.DB.Create(&ProjectModel{
		ID:        project2ID,
		Name:      "Project 2",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}).Error)

	require.NoError(t, result.DB.Create(&VisualizationModel{
		ID:        viz1ID,
		ProjectID: project1ID,
		Name:      "Viz 1",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}).Error)

	require.NoError(t, result.DB.Create(&VisualizationModel{
		ID:        viz2ID,
		ProjectID: project2ID,
		Name:      "Viz 2",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}).Error)

	repo := NewVersionRepository(result.DB)

	now := time.Now().UTC()

	// Create version for project 1
	version1 := &workspace.VisualizationVersion{
		ID:              uuid.NewString(),
		VisualizationID: viz1ID,
		ProjectID:       project1ID,
		VersionNumber:   1,
		SessionID:       uuid.NewString(),
		CreatedAt:       now,
	}
	require.NoError(t, repo.Create(ctx, version1))

	// Create version for project 2
	version2 := &workspace.VisualizationVersion{
		ID:              uuid.NewString(),
		VisualizationID: viz2ID,
		ProjectID:       project2ID,
		VersionNumber:   1,
		SessionID:       uuid.NewString(),
		CreatedAt:       now,
	}
	require.NoError(t, repo.Create(ctx, version2))

	// GetByID should enforce project scoping
	loaded, err := repo.GetByID(ctx, project1ID, version1.ID)
	require.NoError(t, err)
	assert.Equal(t, project1ID, loaded.ProjectID)

	// Attempting to get version from wrong project should fail
	_, err = repo.GetByID(ctx, project2ID, version1.ID)
	assert.Error(t, err, "should not find version from different project")

	// ListByVisualization should only return versions for the project
	versions, err := repo.ListByVisualization(ctx, project1ID, viz1ID, 10)
	require.NoError(t, err)
	assert.Len(t, versions, 1)
	assert.Equal(t, project1ID, versions[0].ProjectID)
}

func TestVisualizationVersionRepository_WithArtifacts(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "test.db")

	result, err := Bootstrap(ctx, BootstrapConfig{
		DatabasePath:      dbPath,
		EnableForeignKeys: true,
		BusyTimeoutMs:     5000,
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	defer Close(result.DB)

	projectID := uuid.NewString()
	vizID := uuid.NewString()

	require.NoError(t, result.DB.Create(&ProjectModel{
		ID:        projectID,
		Name:      "Test Project",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}).Error)

	require.NoError(t, result.DB.Create(&VisualizationModel{
		ID:        vizID,
		ProjectID: projectID,
		Name:      "Test Visualization",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}).Error)

	repo := NewVersionRepository(result.DB)

	now := time.Now().UTC()
	sessionID := uuid.NewString()

	version := &workspace.VisualizationVersion{
		ID:              uuid.NewString(),
		VisualizationID: vizID,
		ProjectID:       projectID,
		VersionNumber:   1,
		SessionID:       sessionID,
		Summary:         "Version with artifacts",
		Artifacts: []workspace.VersionArtifact{
			{
				ID:             uuid.NewString(),
				AssetID:        uuid.NewString(),
				Kind:           "final",
				MIMEType:       "image/png",
				StorageKey:     "assets/abc123.png",
				ByteSize:       12345,
				ChecksumSHA256: "sha256hash",
			},
		},
		CreatedAt: now,
		SessionSnapshot: &domainagent.SessionState{
			SessionID:     sessionID,
			SchemaVersion: "1.0.0",
			Status:        domainagent.StatusCompleted,
			StartedAt:     now,
		},
	}

	require.NoError(t, repo.Create(ctx, version))

	// Load and verify artifacts
	loaded, err := repo.GetByID(ctx, projectID, version.ID)
	require.NoError(t, err)
	require.Len(t, loaded.Artifacts, 1)
	assert.Equal(t, "final", loaded.Artifacts[0].Kind)
	assert.Equal(t, "image/png", loaded.Artifacts[0].MIMEType)
	assert.Equal(t, "assets/abc123.png", loaded.Artifacts[0].StorageKey)
	assert.Equal(t, int64(12345), loaded.Artifacts[0].ByteSize)
}
