package persistence

import (
	"context"
	"testing"
	"time"

	domainagent "github.com/paperbanana/paperbanana/internal/domain/agent"
	"github.com/paperbanana/paperbanana/internal/domain/workspace"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// historyTestRepositories implements Repositories for testing history service.
type historyTestRepositories struct {
	sessions  *historyTestSessionRepository
	versions  *historyTestVersionRepository
	vizs      *historyTestVisualizationRepository
	projects  *historyTestProjectRepository
	folders   *historyTestFolderRepository
	assets    *historyTestAssetRepository
}

func (m *historyTestRepositories) Projects() workspace.ProjectRepository    { return m.projects }
func (m *historyTestRepositories) Folders() workspace.FolderRepository      { return m.folders }
func (m *historyTestRepositories) Visualizations() workspace.VisualizationRepository { return m.vizs }
func (m *historyTestRepositories) Versions() workspace.VersionRepository    { return m.versions }
func (m *historyTestRepositories) Sessions() workspace.SessionRepository    { return m.sessions }
func (m *historyTestRepositories) Assets() workspace.AssetRepository        { return m.assets }

type historyTestSessionRepository struct {
	sessions map[string]*workspace.SessionRecord
}

func (m *historyTestSessionRepository) Create(ctx context.Context, session *workspace.SessionRecord) error {
	m.sessions[session.ID] = session
	return nil
}

func (m *historyTestSessionRepository) GetByID(ctx context.Context, id string) (*workspace.SessionRecord, error) {
	if s, ok := m.sessions[id]; ok {
		return s, nil
	}
	return nil, assert.AnError
}

func (m *historyTestSessionRepository) GetByProject(ctx context.Context, projectID string, limit int) ([]*workspace.SessionRecord, error) {
	var result []*workspace.SessionRecord
	for _, s := range m.sessions {
		if s.ProjectID == projectID {
			result = append(result, s)
		}
	}
	return result, nil
}

func (m *historyTestSessionRepository) GetByVisualization(ctx context.Context, projectID, visualizationID string, limit int) ([]*workspace.SessionRecord, error) {
	var result []*workspace.SessionRecord
	for _, s := range m.sessions {
		if s.ProjectID == projectID && s.VisualizationID != nil && *s.VisualizationID == visualizationID {
			result = append(result, s)
		}
	}
	return result, nil
}

func (m *historyTestSessionRepository) Update(ctx context.Context, session *workspace.SessionRecord) error {
	m.sessions[session.ID] = session
	return nil
}

type historyTestVersionRepository struct {
	versions         map[string]*workspace.VisualizationVersion
	byVisualization  map[string][]*workspace.VisualizationVersion
}

func (m *historyTestVersionRepository) Create(ctx context.Context, version *workspace.VisualizationVersion) error {
	m.versions[version.ID] = version
	m.byVisualization[version.VisualizationID] = append(m.byVisualization[version.VisualizationID], version)
	return nil
}

func (m *historyTestVersionRepository) GetByID(ctx context.Context, projectID, id string) (*workspace.VisualizationVersion, error) {
	if v, ok := m.versions[id]; ok && v.ProjectID == projectID {
		return v, nil
	}
	return nil, assert.AnError
}

func (m *historyTestVersionRepository) ListByVisualization(ctx context.Context, projectID, visualizationID string, limit int) ([]*workspace.VisualizationVersion, error) {
	versions := m.byVisualization[visualizationID]
	// Return in descending order (most recent first)
	result := make([]*workspace.VisualizationVersion, len(versions))
	for i, v := range versions {
		result[len(versions)-1-i] = v
	}
	return result, nil
}

func (m *historyTestVersionRepository) GetLatestByVisualization(ctx context.Context, projectID, visualizationID string) (*workspace.VisualizationVersion, error) {
	versions := m.byVisualization[visualizationID]
	if len(versions) == 0 {
		return nil, assert.AnError
	}
	return versions[len(versions)-1], nil
}

type historyTestVisualizationRepository struct {
	visualizations map[string]*workspace.Visualization
}

func (m *historyTestVisualizationRepository) Create(ctx context.Context, viz *workspace.Visualization) error {
	m.visualizations[viz.ID] = viz
	return nil
}

func (m *historyTestVisualizationRepository) GetByID(ctx context.Context, projectID, id string) (*workspace.Visualization, error) {
	if v, ok := m.visualizations[id]; ok && v.ProjectID == projectID {
		return v, nil
	}
	return nil, assert.AnError
}

func (m *historyTestVisualizationRepository) ListByFolder(ctx context.Context, projectID string, folderID *string) ([]*workspace.Visualization, error) {
	return nil, nil
}

func (m *historyTestVisualizationRepository) ListByProject(ctx context.Context, projectID string) ([]*workspace.Visualization, error) {
	return nil, nil
}

func (m *historyTestVisualizationRepository) Update(ctx context.Context, viz *workspace.Visualization) error {
	m.visualizations[viz.ID] = viz
	return nil
}

func (m *historyTestVisualizationRepository) Delete(ctx context.Context, projectID, id string) error {
	return nil
}

func (m *historyTestVisualizationRepository) SetCurrentVersion(ctx context.Context, projectID, visualizationID, versionID string) error {
	if v, ok := m.visualizations[visualizationID]; ok {
		v.CurrentVersionID = &versionID
	}
	return nil
}

func (m *historyTestVisualizationRepository) Restore(ctx context.Context, projectID, id string) error {
	if v, ok := m.visualizations[id]; ok {
		v.DeletedAt = nil
	}
	return nil
}

type historyTestProjectRepository struct{}

func (m *historyTestProjectRepository) Create(ctx context.Context, project *workspace.Project) error { return nil }
func (m *historyTestProjectRepository) GetByID(ctx context.Context, id string) (*workspace.Project, error) {
	return &workspace.Project{ID: id}, nil
}
func (m *historyTestProjectRepository) List(ctx context.Context) ([]*workspace.Project, error) { return nil, nil }
func (m *historyTestProjectRepository) Update(ctx context.Context, project *workspace.Project) error { return nil }
func (m *historyTestProjectRepository) Delete(ctx context.Context, id string) error            { return nil }

type historyTestFolderRepository struct{}

func (m *historyTestFolderRepository) Create(ctx context.Context, folder *workspace.Folder) error { return nil }
func (m *historyTestFolderRepository) GetByID(ctx context.Context, projectID, id string) (*workspace.Folder, error) {
	return nil, nil
}
func (m *historyTestFolderRepository) ListByParent(ctx context.Context, projectID string, parentID *string) ([]*workspace.Folder, error) {
	return nil, nil
}
func (m *historyTestFolderRepository) ListByProject(ctx context.Context, projectID string) ([]*workspace.Folder, error) {
	return nil, nil
}
func (m *historyTestFolderRepository) Update(ctx context.Context, folder *workspace.Folder) error { return nil }
func (m *historyTestFolderRepository) Delete(ctx context.Context, projectID, id string) error    { return nil }
func (m *historyTestFolderRepository) GetDescendantIDs(ctx context.Context, projectID, folderID string) ([]string, error) {
	return nil, nil
}
func (m *historyTestFolderRepository) Restore(ctx context.Context, projectID, id string) error { return nil }

type historyTestAssetRepository struct{}

func (m *historyTestAssetRepository) Create(ctx context.Context, asset *workspace.Asset) error { return nil }
func (m *historyTestAssetRepository) GetByID(ctx context.Context, projectID, id string) (*workspace.Asset, error) {
	return nil, nil
}
func (m *historyTestAssetRepository) GetByStorageKey(ctx context.Context, storageKey string) (*workspace.Asset, error) {
	return nil, nil
}
func (m *historyTestAssetRepository) ListByVisualization(ctx context.Context, projectID, visualizationID string) ([]*workspace.Asset, error) {
	return nil, nil
}
func (m *historyTestAssetRepository) ListByVersion(ctx context.Context, projectID, versionID string) ([]*workspace.Asset, error) {
	return nil, nil
}
func (m *historyTestAssetRepository) Delete(ctx context.Context, projectID, id string) error { return nil }

// historyTestTxManager implements TxManager for testing.
type historyTestTxManager struct {
	repos *historyTestRepositories
}

func (m *historyTestTxManager) RunInTx(ctx context.Context, fn func(Repositories) error) error {
	return fn(m.repos)
}

func (m *historyTestTxManager) ReadOnlyTx(ctx context.Context, fn func(Repositories) error) error {
	return fn(m.repos)
}

func newTestHistoryService() (*HistoryService, *historyTestRepositories) {
	repos := &historyTestRepositories{
		sessions: &historyTestSessionRepository{sessions: make(map[string]*workspace.SessionRecord)},
		versions: &historyTestVersionRepository{
			versions:        make(map[string]*workspace.VisualizationVersion),
			byVisualization: make(map[string][]*workspace.VisualizationVersion),
		},
		vizs:     &historyTestVisualizationRepository{visualizations: make(map[string]*workspace.Visualization)},
		projects: &historyTestProjectRepository{},
		folders:  &historyTestFolderRepository{},
		assets:   &historyTestAssetRepository{},
	}
	txManager := &historyTestTxManager{repos: repos}
	return NewHistoryService(txManager), repos
}

func TestHistoryService_SaveSession(t *testing.T) {
	ctx := context.Background()
	service, repos := newTestHistoryService()

	projectID := uuid.NewString()
	vizID := uuid.NewString()
	sessionID := uuid.NewString()

	now := time.Now().UTC()
	session := &workspace.SessionRecord{
		ID:              sessionID,
		ProjectID:       projectID,
		VisualizationID: &vizID,
		Status:          string(domainagent.StatusRunning),
		CurrentStage:    string(domainagent.StagePlanner),
		SchemaVersion:   "1.0.0",
		Snapshot: &domainagent.SessionState{
			SessionID:     sessionID,
			SchemaVersion: "1.0.0",
			Status:        domainagent.StatusRunning,
			CurrentStage:  domainagent.StagePlanner,
			StartedAt:     now,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := service.SaveSession(ctx, session)
	require.NoError(t, err)

	// Verify session was saved
	loaded, err := repos.sessions.GetByID(ctx, sessionID)
	require.NoError(t, err)
	assert.Equal(t, sessionID, loaded.ID)
	assert.Equal(t, string(domainagent.StatusRunning), loaded.Status)
}

func TestHistoryService_PublishVersion(t *testing.T) {
	ctx := context.Background()
	service, repos := newTestHistoryService()

	projectID := uuid.NewString()
	vizID := uuid.NewString()
	sessionID := uuid.NewString()

	// Create visualization first
	repos.vizs.visualizations[vizID] = &workspace.Visualization{
		ID:        vizID,
		ProjectID: projectID,
		Name:      "Test Visualization",
	}

	now := time.Now().UTC()
	session := &workspace.SessionRecord{
		ID:              sessionID,
		ProjectID:       projectID,
		VisualizationID: &vizID,
		Status:          string(domainagent.StatusCompleted),
		CurrentStage:    string(domainagent.StageCritic),
		SchemaVersion:   "1.0.0",
		Snapshot: &domainagent.SessionState{
			SessionID:     sessionID,
			SchemaVersion: "1.0.0",
			Status:        domainagent.StatusCompleted,
			CurrentStage:  domainagent.StageCritic,
			StartedAt:     now,
			CompletedAt:   now,
		},
		CreatedAt:   now,
		UpdatedAt:   now,
		CompletedAt: &now,
	}

	versionID, err := service.PublishVersion(ctx, session, "First version")
	require.NoError(t, err)
	assert.NotEmpty(t, versionID)

	// Verify version was created
	versions, err := repos.versions.ListByVisualization(ctx, projectID, vizID, 10)
	require.NoError(t, err)
	assert.Len(t, versions, 1)
	assert.Equal(t, 1, versions[0].VersionNumber)
	assert.Equal(t, "First version", versions[0].Summary)

	// Verify current version pointer was updated
	viz, err := repos.vizs.GetByID(ctx, projectID, vizID)
	require.NoError(t, err)
	assert.Equal(t, versionID, *viz.CurrentVersionID)

	// Verify session was saved
	_, err = repos.sessions.GetByID(ctx, sessionID)
	require.NoError(t, err)
}

func TestHistoryService_ListHistory(t *testing.T) {
	ctx := context.Background()
	service, repos := newTestHistoryService()

	projectID := uuid.NewString()
	vizID := uuid.NewString()

	// Create visualization
	repos.vizs.visualizations[vizID] = &workspace.Visualization{
		ID:        vizID,
		ProjectID: projectID,
		Name:      "Test Visualization",
	}

	now := time.Now().UTC()

	// Create multiple versions
	for i := 1; i <= 3; i++ {
		versionID := uuid.NewString()
		sessionID := uuid.NewString()
		version := &workspace.VisualizationVersion{
			ID:              versionID,
			VisualizationID: vizID,
			ProjectID:       projectID,
			VersionNumber:   i,
			SessionID:       sessionID,
			Summary:         "Version " + string(rune('A'+i-1)),
			CreatedAt:       now.Add(time.Duration(i) * time.Hour),
		}
		require.NoError(t, repos.versions.Create(ctx, version))
	}

	// List history
	versions, err := service.ListHistory(ctx, projectID, vizID, 10)
	require.NoError(t, err)
	assert.Len(t, versions, 3)

	// Verify ordering (most recent first)
	assert.Equal(t, 3, versions[0].VersionNumber)
	assert.Equal(t, 2, versions[1].VersionNumber)
	assert.Equal(t, 1, versions[2].VersionNumber)
}

func TestHistoryService_LoadResumableSession(t *testing.T) {
	ctx := context.Background()
	service, repos := newTestHistoryService()

	projectID := uuid.NewString()
	vizID := uuid.NewString()
	sessionID := uuid.NewString()

	now := time.Now().UTC()

	// Create a resumable session (failed at planner stage)
	session := &workspace.SessionRecord{
		ID:              sessionID,
		ProjectID:       projectID,
		VisualizationID: &vizID,
		Status:          string(domainagent.StatusFailed),
		CurrentStage:    string(domainagent.StagePlanner),
		SchemaVersion:   "1.0.0",
		Snapshot: &domainagent.SessionState{
			SessionID:     sessionID,
			SchemaVersion: "1.0.0",
			Status:        domainagent.StatusFailed,
			CurrentStage:  domainagent.StagePlanner,
			StartedAt:     now,
			UpdatedAt:     now,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
	require.NoError(t, repos.sessions.Create(ctx, session))

	// Load the session
	loaded, err := service.LoadResumableSession(ctx, projectID, vizID)
	require.NoError(t, err)
	assert.Equal(t, sessionID, loaded.ID)
	assert.Equal(t, string(domainagent.StatusFailed), loaded.Status)
	assert.Equal(t, string(domainagent.StagePlanner), loaded.CurrentStage)
}

func TestHistoryService_GetVersion(t *testing.T) {
	ctx := context.Background()
	service, repos := newTestHistoryService()

	projectID := uuid.NewString()
	vizID := uuid.NewString()
	versionID := uuid.NewString()

	// Create version
	version := &workspace.VisualizationVersion{
		ID:              versionID,
		VisualizationID: vizID,
		ProjectID:       projectID,
		VersionNumber:   1,
		SessionID:       uuid.NewString(),
		Summary:         "Test version",
		CreatedAt:       time.Now().UTC(),
	}
	require.NoError(t, repos.versions.Create(ctx, version))

	// Get the version
	loaded, err := service.GetVersion(ctx, projectID, versionID)
	require.NoError(t, err)
	assert.Equal(t, versionID, loaded.ID)
	assert.Equal(t, 1, loaded.VersionNumber)
	assert.Equal(t, "Test version", loaded.Summary)
}

func TestHistoryService_PublishVersionOnlyForSuccessful(t *testing.T) {
	ctx := context.Background()
	service, repos := newTestHistoryService()

	projectID := uuid.NewString()
	vizID := uuid.NewString()
	sessionID := uuid.NewString()

	// Create visualization
	repos.vizs.visualizations[vizID] = &workspace.Visualization{
		ID:        vizID,
		ProjectID: projectID,
		Name:      "Test Visualization",
	}

	now := time.Now().UTC()

	// Create a FAILED session
	failedSession := &workspace.SessionRecord{
		ID:              sessionID,
		ProjectID:       projectID,
		VisualizationID: &vizID,
		Status:          string(domainagent.StatusFailed),
		CurrentStage:    string(domainagent.StagePlanner),
		SchemaVersion:   "1.0.0",
		Snapshot: &domainagent.SessionState{
			SessionID:     sessionID,
			SchemaVersion: "1.0.0",
			Status:        domainagent.StatusFailed,
			CurrentStage:  domainagent.StagePlanner,
			StartedAt:     now,
			UpdatedAt:     now,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	// PublishVersion should still save the session but create a version
	versionID, err := service.PublishVersion(ctx, failedSession, "Failed version")
	require.NoError(t, err)
	assert.NotEmpty(t, versionID, "version should still be created for audit purposes")

	// Session should be saved
	_, err = repos.sessions.GetByID(ctx, sessionID)
	require.NoError(t, err)
}
