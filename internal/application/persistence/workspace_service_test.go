package persistence

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/paperbanana/paperbanana/internal/domain/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockProjectRepository implements workspace.ProjectRepository for testing.
type mockProjectRepository struct {
	projects map[string]*workspace.Project
}

func newMockProjectRepository() *mockProjectRepository {
	return &mockProjectRepository{projects: make(map[string]*workspace.Project)}
}

func (m *mockProjectRepository) Create(ctx context.Context, project *workspace.Project) error {
	m.projects[project.ID] = project
	return nil
}

func (m *mockProjectRepository) GetByID(ctx context.Context, id string) (*workspace.Project, error) {
	p, ok := m.projects[id]
	if !ok {
		return nil, errors.New("project not found")
	}
	return p, nil
}

func (m *mockProjectRepository) List(ctx context.Context) ([]*workspace.Project, error) {
	var result []*workspace.Project
	for _, p := range m.projects {
		result = append(result, p)
	}
	return result, nil
}

func (m *mockProjectRepository) Update(ctx context.Context, project *workspace.Project) error {
	m.projects[project.ID] = project
	return nil
}

func (m *mockProjectRepository) Delete(ctx context.Context, id string) error {
	delete(m.projects, id)
	return nil
}

// mockFolderRepository implements workspace.FolderRepository for testing.
type mockFolderRepository struct {
	folders          map[string]*workspace.Folder
	descendantIDs    map[string][]string
}

func newMockFolderRepository() *mockFolderRepository {
	return &mockFolderRepository{
		folders:       make(map[string]*workspace.Folder),
		descendantIDs: make(map[string][]string),
	}
}

func folderKey(projectID, id string) string {
	return projectID + "/" + id
}

func (m *mockFolderRepository) Create(ctx context.Context, folder *workspace.Folder) error {
	m.folders[folderKey(folder.ProjectID, folder.ID)] = folder
	return nil
}

func (m *mockFolderRepository) GetByID(ctx context.Context, projectID, id string) (*workspace.Folder, error) {
	f, ok := m.folders[folderKey(projectID, id)]
	if !ok {
		return nil, errors.New("folder not found")
	}
	if f.DeletedAt != nil {
		return nil, errors.New("folder not found")
	}
	return f, nil
}

func (m *mockFolderRepository) ListByParent(ctx context.Context, projectID string, parentID *string) ([]*workspace.Folder, error) {
	var result []*workspace.Folder
	for _, f := range m.folders {
		if f.ProjectID != projectID {
			continue
		}
		if parentID == nil && f.ParentID == nil {
			result = append(result, f)
		} else if parentID != nil && f.ParentID != nil && *f.ParentID == *parentID {
			result = append(result, f)
		}
	}
	return result, nil
}

func (m *mockFolderRepository) ListByProject(ctx context.Context, projectID string) ([]*workspace.Folder, error) {
	var result []*workspace.Folder
	for _, f := range m.folders {
		if f.ProjectID == projectID {
			result = append(result, f)
		}
	}
	return result, nil
}

func (m *mockFolderRepository) Update(ctx context.Context, folder *workspace.Folder) error {
	m.folders[folderKey(folder.ProjectID, folder.ID)] = folder
	return nil
}

func (m *mockFolderRepository) Delete(ctx context.Context, projectID, id string) error {
	key := folderKey(projectID, id)
	if f, ok := m.folders[key]; ok {
		now := time.Now()
		f.DeletedAt = &now
	}
	return nil
}

func (m *mockFolderRepository) GetDescendantIDs(ctx context.Context, projectID, folderID string) ([]string, error) {
	key := projectID + "/" + folderID
	if ids, ok := m.descendantIDs[key]; ok {
		return ids, nil
	}
	return []string{folderID}, nil
}

func (m *mockFolderRepository) Restore(ctx context.Context, projectID, id string) error {
	key := folderKey(projectID, id)
	if f, ok := m.folders[key]; ok {
		f.DeletedAt = nil
	}
	return nil
}

// mockVisualizationRepository implements workspace.VisualizationRepository for testing.
type mockVisualizationRepository struct {
	visualizations map[string]*workspace.Visualization
}

func newMockVisualizationRepository() *mockVisualizationRepository {
	return &mockVisualizationRepository{visualizations: make(map[string]*workspace.Visualization)}
}

func vizKey(projectID, id string) string {
	return projectID + "/" + id
}

func (m *mockVisualizationRepository) Create(ctx context.Context, viz *workspace.Visualization) error {
	m.visualizations[vizKey(viz.ProjectID, viz.ID)] = viz
	return nil
}

func (m *mockVisualizationRepository) GetByID(ctx context.Context, projectID, id string) (*workspace.Visualization, error) {
	v, ok := m.visualizations[vizKey(projectID, id)]
	if !ok {
		return nil, errors.New("visualization not found")
	}
	if v.DeletedAt != nil {
		return nil, errors.New("visualization not found")
	}
	return v, nil
}

func (m *mockVisualizationRepository) ListByFolder(ctx context.Context, projectID string, folderID *string) ([]*workspace.Visualization, error) {
	var result []*workspace.Visualization
	for _, v := range m.visualizations {
		if v.ProjectID != projectID {
			continue
		}
		if folderID == nil && v.FolderID == nil {
			result = append(result, v)
		} else if folderID != nil && v.FolderID != nil && *v.FolderID == *folderID {
			result = append(result, v)
		}
	}
	return result, nil
}

func (m *mockVisualizationRepository) ListByProject(ctx context.Context, projectID string) ([]*workspace.Visualization, error) {
	var result []*workspace.Visualization
	for _, v := range m.visualizations {
		if v.ProjectID == projectID {
			result = append(result, v)
		}
	}
	return result, nil
}

func (m *mockVisualizationRepository) Update(ctx context.Context, viz *workspace.Visualization) error {
	m.visualizations[vizKey(viz.ProjectID, viz.ID)] = viz
	return nil
}

func (m *mockVisualizationRepository) Delete(ctx context.Context, projectID, id string) error {
	key := vizKey(projectID, id)
	if v, ok := m.visualizations[key]; ok {
		now := time.Now()
		v.DeletedAt = &now
	}
	return nil
}

func (m *mockVisualizationRepository) SetCurrentVersion(ctx context.Context, projectID, visualizationID, versionID string) error {
	key := vizKey(projectID, visualizationID)
	if v, ok := m.visualizations[key]; ok {
		v.CurrentVersionID = &versionID
	}
	return nil
}

func (m *mockVisualizationRepository) Restore(ctx context.Context, projectID, id string) error {
	key := vizKey(projectID, id)
	if v, ok := m.visualizations[key]; ok {
		v.DeletedAt = nil
	}
	return nil
}

// mockTxManager implements TxManager for testing.
type mockTxManager struct {
	repos *mockRepositories
}

type mockRepositories struct {
	projects       *mockProjectRepository
	folders        *mockFolderRepository
	visualizations *mockVisualizationRepository
}

func (m *mockTxManager) RunInTx(ctx context.Context, fn func(Repositories) error) error {
	return fn(m.repos)
}

func (m *mockTxManager) ReadOnlyTx(ctx context.Context, fn func(Repositories) error) error {
	return fn(m.repos)
}

func (m *mockRepositories) Projects() workspace.ProjectRepository {
	return m.projects
}

func (m *mockRepositories) Folders() workspace.FolderRepository {
	return m.folders
}

func (m *mockRepositories) Visualizations() workspace.VisualizationRepository {
	return m.visualizations
}

func (m *mockRepositories) Versions() workspace.VersionRepository {
	return nil
}

func (m *mockRepositories) Sessions() workspace.SessionRepository {
	return nil
}

func (m *mockRepositories) Assets() workspace.AssetRepository {
	return nil
}

func setupMockWorkspace() (*mockTxManager, *mockRepositories) {
	repos := &mockRepositories{
		projects:       newMockProjectRepository(),
		folders:        newMockFolderRepository(),
		visualizations: newMockVisualizationRepository(),
	}
	return &mockTxManager{repos: repos}, repos
}

func TestWorkspaceService_CreateHierarchy(t *testing.T) {
	ctx := context.Background()
	txManager, repos := setupMockWorkspace()
	service := NewWorkspaceService(txManager)

	t.Run("creates project", func(t *testing.T) {
		project, err := service.CreateProject(ctx, "Test Project", "A test project")
		require.NoError(t, err)
		assert.NotEmpty(t, project.ID)
		assert.Equal(t, "Test Project", project.Name)

		// Verify it was stored
		stored, err := repos.projects.GetByID(ctx, project.ID)
		require.NoError(t, err)
		assert.Equal(t, project.ID, stored.ID)
	})

	t.Run("creates folder in project", func(t *testing.T) {
		// Create project first
		project, err := service.CreateProject(ctx, "Project for Folder", "")
		require.NoError(t, err)

		// Create folder
		folder, err := service.CreateFolder(ctx, project.ID, nil, "My Folder")
		require.NoError(t, err)
		assert.NotEmpty(t, folder.ID)
		assert.Equal(t, project.ID, folder.ProjectID)
		assert.Equal(t, "My Folder", folder.Name)
		assert.Nil(t, folder.ParentID)
	})

	t.Run("creates nested folder", func(t *testing.T) {
		project, err := service.CreateProject(ctx, "Project for Nested Folder", "")
		require.NoError(t, err)

		// Create parent folder
		parent, err := service.CreateFolder(ctx, project.ID, nil, "Parent")
		require.NoError(t, err)

		// Create child folder
		child, err := service.CreateFolder(ctx, project.ID, &parent.ID, "Child")
		require.NoError(t, err)
		assert.Equal(t, parent.ID, *child.ParentID)
	})

	t.Run("creates visualization", func(t *testing.T) {
		project, err := service.CreateProject(ctx, "Project for Viz", "")
		require.NoError(t, err)

		viz, err := service.CreateVisualization(ctx, project.ID, nil, "My Chart")
		require.NoError(t, err)
		assert.NotEmpty(t, viz.ID)
		assert.Equal(t, project.ID, viz.ProjectID)
		assert.Equal(t, "My Chart", viz.Name)
	})

	t.Run("rejects folder in non-existent project", func(t *testing.T) {
		_, err := service.CreateFolder(ctx, "non-existent-project", nil, "Orphan Folder")
		assert.Error(t, err)
	})

	t.Run("creates visualization in folder", func(t *testing.T) {
		project, err := service.CreateProject(ctx, "Project for Viz in Folder", "")
		require.NoError(t, err)

		folder, err := service.CreateFolder(ctx, project.ID, nil, "Container")
		require.NoError(t, err)

		viz, err := service.CreateVisualization(ctx, project.ID, &folder.ID, "Chart in Folder")
		require.NoError(t, err)
		assert.Equal(t, folder.ID, *viz.FolderID)
	})
}

func TestWorkspaceService_MoveItem(t *testing.T) {
	ctx := context.Background()
	txManager, repos := setupMockWorkspace()
	service := NewWorkspaceService(txManager)

	t.Run("moves folder within same project", func(t *testing.T) {
		project, err := service.CreateProject(ctx, "Move Folder Project", "")
		require.NoError(t, err)

		// Create parent folders
		parent1, err := service.CreateFolder(ctx, project.ID, nil, "Parent 1")
		require.NoError(t, err)

		parent2, err := service.CreateFolder(ctx, project.ID, nil, "Parent 2")
		require.NoError(t, err)

		// Create child folder under parent1
		child, err := service.CreateFolder(ctx, project.ID, &parent1.ID, "Child")
		require.NoError(t, err)
		assert.Equal(t, parent1.ID, *child.ParentID)

		// Move to parent2
		err = service.MoveFolder(ctx, project.ID, child.ID, &parent2.ID)
		require.NoError(t, err)

		// Verify
		moved, err := repos.folders.GetByID(ctx, project.ID, child.ID)
		require.NoError(t, err)
		assert.Equal(t, parent2.ID, *moved.ParentID)
	})

	t.Run("moves visualization within same project", func(t *testing.T) {
		project, err := service.CreateProject(ctx, "Move Viz Project", "")
		require.NoError(t, err)

		folder1, err := service.CreateFolder(ctx, project.ID, nil, "Folder 1")
		require.NoError(t, err)

		folder2, err := service.CreateFolder(ctx, project.ID, nil, "Folder 2")
		require.NoError(t, err)

		viz, err := service.CreateVisualization(ctx, project.ID, &folder1.ID, "Chart")
		require.NoError(t, err)
		assert.Equal(t, folder1.ID, *viz.FolderID)

		// Move to folder2
		err = service.MoveVisualization(ctx, project.ID, viz.ID, &folder2.ID)
		require.NoError(t, err)

		// Verify
		moved, err := repos.visualizations.GetByID(ctx, project.ID, viz.ID)
		require.NoError(t, err)
		assert.Equal(t, folder2.ID, *moved.FolderID)
	})

	t.Run("moves visualization to root", func(t *testing.T) {
		project, err := service.CreateProject(ctx, "Move to Root Project", "")
		require.NoError(t, err)

		folder, err := service.CreateFolder(ctx, project.ID, nil, "Folder")
		require.NoError(t, err)

		viz, err := service.CreateVisualization(ctx, project.ID, &folder.ID, "Chart")
		require.NoError(t, err)

		// Move to root
		err = service.MoveVisualization(ctx, project.ID, viz.ID, nil)
		require.NoError(t, err)

		moved, err := repos.visualizations.GetByID(ctx, project.ID, viz.ID)
		require.NoError(t, err)
		assert.Nil(t, moved.FolderID)
	})

	t.Run("rejects cross-project move", func(t *testing.T) {
		project1, err := service.CreateProject(ctx, "Project 1", "")
		require.NoError(t, err)

		project2, err := service.CreateProject(ctx, "Project 2", "")
		require.NoError(t, err)

		folder, err := service.CreateFolder(ctx, project1.ID, nil, "Folder")
		require.NoError(t, err)

		// Try to move to another project's folder
		err = service.MoveFolder(ctx, project2.ID, folder.ID, nil)
		assert.Error(t, err, "should reject cross-project folder access")
	})
}

func TestWorkspaceService_TrashAndRestore(t *testing.T) {
	ctx := context.Background()
	txManager, repos := setupMockWorkspace()
	service := NewWorkspaceService(txManager)

	t.Run("trashes folder", func(t *testing.T) {
		project, err := service.CreateProject(ctx, "Trash Folder Project", "")
		require.NoError(t, err)

		folder, err := service.CreateFolder(ctx, project.ID, nil, "To Trash")
		require.NoError(t, err)

		// Trash it
		err = service.TrashFolder(ctx, project.ID, folder.ID)
		require.NoError(t, err)

		// Should not be found in normal queries
		_, err = repos.folders.GetByID(ctx, project.ID, folder.ID)
		assert.Error(t, err, "trashed folder should not be found")
	})

	t.Run("trashes visualization", func(t *testing.T) {
		project, err := service.CreateProject(ctx, "Trash Viz Project", "")
		require.NoError(t, err)

		viz, err := service.CreateVisualization(ctx, project.ID, nil, "To Trash")
		require.NoError(t, err)

		// Trash it
		err = service.TrashVisualization(ctx, project.ID, viz.ID)
		require.NoError(t, err)

		// Should not be found
		_, err = repos.visualizations.GetByID(ctx, project.ID, viz.ID)
		assert.Error(t, err, "trashed visualization should not be found")
	})

	t.Run("restores folder", func(t *testing.T) {
		project, err := service.CreateProject(ctx, "Restore Folder Project", "")
		require.NoError(t, err)

		folder, err := service.CreateFolder(ctx, project.ID, nil, "To Restore")
		require.NoError(t, err)

		// Trash
		require.NoError(t, service.TrashFolder(ctx, project.ID, folder.ID))

		// Restore
		err = service.RestoreFolder(ctx, project.ID, folder.ID)
		require.NoError(t, err)

		// Should be found again
		restored, err := repos.folders.GetByID(ctx, project.ID, folder.ID)
		require.NoError(t, err)
		assert.Nil(t, restored.DeletedAt)
	})

	t.Run("restores visualization", func(t *testing.T) {
		project, err := service.CreateProject(ctx, "Restore Viz Project", "")
		require.NoError(t, err)

		viz, err := service.CreateVisualization(ctx, project.ID, nil, "To Restore")
		require.NoError(t, err)

		// Trash
		require.NoError(t, service.TrashVisualization(ctx, project.ID, viz.ID))

		// Restore
		err = service.RestoreVisualization(ctx, project.ID, viz.ID)
		require.NoError(t, err)

		// Should be found again
		restored, err := repos.visualizations.GetByID(ctx, project.ID, viz.ID)
		require.NoError(t, err)
		assert.Nil(t, restored.DeletedAt)
	})

	t.Run("trashes project", func(t *testing.T) {
		project, err := service.CreateProject(ctx, "To Trash Project", "")
		require.NoError(t, err)

		// Trash it
		err = service.TrashProject(ctx, project.ID)
		require.NoError(t, err)

		// Should not be found
		_, err = repos.projects.GetByID(ctx, project.ID)
		assert.Error(t, err, "trashed project should not be found")
	})
}

func TestWorkspaceService_ListFolderContents(t *testing.T) {
	ctx := context.Background()
	txManager, _ := setupMockWorkspace()
	service := NewWorkspaceService(txManager)

	t.Run("lists mixed contents", func(t *testing.T) {
		project, err := service.CreateProject(ctx, "List Contents Project", "")
		require.NoError(t, err)

		// Create root folders
		_, err = service.CreateFolder(ctx, project.ID, nil, "Root Folder A")
		require.NoError(t, err)

		_, err = service.CreateFolder(ctx, project.ID, nil, "Root Folder B")
		require.NoError(t, err)

		// Create root visualizations
		_, err = service.CreateVisualization(ctx, project.ID, nil, "Root Viz 1")
		require.NoError(t, err)

		_, err = service.CreateVisualization(ctx, project.ID, nil, "Root Viz 2")
		require.NoError(t, err)

		// List root contents
		contents, err := service.ListFolderContents(ctx, project.ID, nil)
		require.NoError(t, err)

		assert.Len(t, contents.Folders, 2)
		assert.Len(t, contents.Visualizations, 2)
	})
}
