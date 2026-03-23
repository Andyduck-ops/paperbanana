package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/paperbanana/paperbanana/internal/application/persistence"
	domainworkspace "github.com/paperbanana/paperbanana/internal/domain/workspace"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// mockWorkspaceService implements WorkspaceService for testing.
type mockWorkspaceService struct {
	projects       map[string]*domainworkspace.Project
	folders        map[string]*domainworkspace.Folder
	visualizations map[string]*domainworkspace.Visualization
	err            error
}

func newMockWorkspaceService() *mockWorkspaceService {
	return &mockWorkspaceService{
		projects:       make(map[string]*domainworkspace.Project),
		folders:        make(map[string]*domainworkspace.Folder),
		visualizations: make(map[string]*domainworkspace.Visualization),
	}
}

func (m *mockWorkspaceService) CreateProject(ctx context.Context, name, description string) (*domainworkspace.Project, error) {
	if m.err != nil {
		return nil, m.err
	}
	project := &domainworkspace.Project{
		ID:          uuid.NewString(),
		Name:        name,
		Description: description,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	m.projects[project.ID] = project
	return project, nil
}

func (m *mockWorkspaceService) GetProject(ctx context.Context, id string) (*domainworkspace.Project, error) {
	if m.err != nil {
		return nil, m.err
	}
	if p, ok := m.projects[id]; ok {
		return p, nil
	}
	return nil, errors.New("project not found")
}

func (m *mockWorkspaceService) ListProjects(ctx context.Context) ([]*domainworkspace.Project, error) {
	if m.err != nil {
		return nil, m.err
	}
	var result []*domainworkspace.Project
	for _, p := range m.projects {
		result = append(result, p)
	}
	return result, nil
}

func (m *mockWorkspaceService) CreateFolder(ctx context.Context, projectID string, parentID *string, name string) (*domainworkspace.Folder, error) {
	if m.err != nil {
		return nil, m.err
	}
	if _, ok := m.projects[projectID]; !ok {
		return nil, errors.New("project not found")
	}
	folder := &domainworkspace.Folder{
		ID:        uuid.NewString(),
		ProjectID: projectID,
		ParentID:  parentID,
		Name:      name,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	m.folders[projectID+"/"+folder.ID] = folder
	return folder, nil
}

func (m *mockWorkspaceService) CreateVisualization(ctx context.Context, projectID string, folderID *string, name string) (*domainworkspace.Visualization, error) {
	if m.err != nil {
		return nil, m.err
	}
	if _, ok := m.projects[projectID]; !ok {
		return nil, errors.New("project not found")
	}
	viz := &domainworkspace.Visualization{
		ID:        uuid.NewString(),
		ProjectID: projectID,
		FolderID:  folderID,
		Name:      name,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	m.visualizations[projectID+"/"+viz.ID] = viz
	return viz, nil
}

func (m *mockWorkspaceService) MoveFolder(ctx context.Context, projectID, folderID string, newParentID *string) error {
	if m.err != nil {
		return m.err
	}
	key := projectID + "/" + folderID
	if f, ok := m.folders[key]; ok {
		f.ParentID = newParentID
		return nil
	}
	return errors.New("folder not found")
}

func (m *mockWorkspaceService) MoveVisualization(ctx context.Context, projectID, vizID string, newFolderID *string) error {
	if m.err != nil {
		return m.err
	}
	key := projectID + "/" + vizID
	if v, ok := m.visualizations[key]; ok {
		v.FolderID = newFolderID
		return nil
	}
	return errors.New("visualization not found")
}

func (m *mockWorkspaceService) TrashFolder(ctx context.Context, projectID, folderID string) error {
	if m.err != nil {
		return m.err
	}
	key := projectID + "/" + folderID
	if f, ok := m.folders[key]; ok {
		now := time.Now()
		f.DeletedAt = &now
		return nil
	}
	return errors.New("folder not found")
}

func (m *mockWorkspaceService) TrashVisualization(ctx context.Context, projectID, vizID string) error {
	if m.err != nil {
		return m.err
	}
	key := projectID + "/" + vizID
	if v, ok := m.visualizations[key]; ok {
		now := time.Now()
		v.DeletedAt = &now
		return nil
	}
	return errors.New("visualization not found")
}

func (m *mockWorkspaceService) TrashProject(ctx context.Context, projectID string) error {
	if m.err != nil {
		return m.err
	}
	if _, ok := m.projects[projectID]; ok {
		delete(m.projects, projectID)
		return nil
	}
	return errors.New("project not found")
}

func (m *mockWorkspaceService) RestoreFolder(ctx context.Context, projectID, folderID string) error {
	if m.err != nil {
		return m.err
	}
	key := projectID + "/" + folderID
	if f, ok := m.folders[key]; ok {
		f.DeletedAt = nil
		return nil
	}
	return errors.New("folder not found")
}

func (m *mockWorkspaceService) RestoreVisualization(ctx context.Context, projectID, vizID string) error {
	if m.err != nil {
		return m.err
	}
	key := projectID + "/" + vizID
	if v, ok := m.visualizations[key]; ok {
		v.DeletedAt = nil
		return nil
	}
	return errors.New("visualization not found")
}

func (m *mockWorkspaceService) ListFolderContents(ctx context.Context, projectID string, folderID *string) (*persistence.FolderContents, error) {
	if m.err != nil {
		return nil, m.err
	}
	contents := &persistence.FolderContents{}
	for _, f := range m.folders {
		if f.ProjectID != projectID {
			continue
		}
		if folderID == nil && f.ParentID == nil {
			contents.Folders = append(contents.Folders, f)
		} else if folderID != nil && f.ParentID != nil && *f.ParentID == *folderID {
			contents.Folders = append(contents.Folders, f)
		}
	}
	for _, v := range m.visualizations {
		if v.ProjectID != projectID {
			continue
		}
		if folderID == nil && v.FolderID == nil {
			contents.Visualizations = append(contents.Visualizations, v)
		} else if folderID != nil && v.FolderID != nil && *v.FolderID == *folderID {
			contents.Visualizations = append(contents.Visualizations, v)
		}
	}
	return contents, nil
}

func setupWorkspaceTest(t *testing.T, mock *mockWorkspaceService) (*gin.Engine, *WorkspaceHandler) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()
	handler := NewWorkspaceHandler(mock, logger)
	router := gin.New()
	return router, handler
}

func TestWorkspaceHandler_CreateAndList(t *testing.T) {
	mock := newMockWorkspaceService()
	router, handler := setupWorkspaceTest(t, mock)
	router.POST("/projects", handler.CreateProject)
	router.GET("/projects", handler.ListProjects)
	router.GET("/projects/:project_id", handler.GetProject)

	t.Run("creates and lists projects", func(t *testing.T) {
		// Create a project
		createReq := CreateProjectRequest{
			Name:        "Test Project",
			Description: "A test project",
		}
		body, _ := json.Marshal(createReq)
		req := httptest.NewRequest("POST", "/projects", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response ProjectResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "Test Project", response.Name)
		assert.NotEmpty(t, response.ID)

		// List projects
		req = httptest.NewRequest("GET", "/projects", nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var listResponse map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &listResponse)
		require.NoError(t, err)
		projects := listResponse["projects"].([]interface{})
		assert.GreaterOrEqual(t, len(projects), 1)
	})

	t.Run("gets a project by ID", func(t *testing.T) {
		// Create a project first
		createReq := CreateProjectRequest{Name: "Get Test Project"}
		body, _ := json.Marshal(createReq)
		req := httptest.NewRequest("POST", "/projects", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		var created ProjectResponse
		json.Unmarshal(w.Body.Bytes(), &created)

		// Get it back
		req = httptest.NewRequest("GET", "/projects/"+created.ID, nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("returns error for missing name", func(t *testing.T) {
		createReq := CreateProjectRequest{Description: "No name"}
		body, _ := json.Marshal(createReq)
		req := httptest.NewRequest("POST", "/projects", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestWorkspaceHandler_MoveItemWithinProject(t *testing.T) {
	mock := newMockWorkspaceService()
	router, handler := setupWorkspaceTest(t, mock)
	router.POST("/projects", handler.CreateProject)
	router.POST("/folders", handler.CreateFolder)
	router.POST("/workspace/move", handler.MoveItem)

	// Create a project
	createProjectReq := CreateProjectRequest{Name: "Move Test Project"}
	body, _ := json.Marshal(createProjectReq)
	req := httptest.NewRequest("POST", "/projects", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var project ProjectResponse
	json.Unmarshal(w.Body.Bytes(), &project)

	// Create two folders
	createFolderReq := CreateFolderRequest{
		ProjectID: project.ID,
		Name:      "Folder 1",
	}
	body, _ = json.Marshal(createFolderReq)
	req = httptest.NewRequest("POST", "/folders", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var folder1 FolderResponse
	json.Unmarshal(w.Body.Bytes(), &folder1)

	createFolderReq = CreateFolderRequest{
		ProjectID: project.ID,
		Name:      "Folder 2",
	}
	body, _ = json.Marshal(createFolderReq)
	req = httptest.NewRequest("POST", "/folders", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var folder2 FolderResponse
	json.Unmarshal(w.Body.Bytes(), &folder2)

	// Create a child folder under folder1
	createFolderReq = CreateFolderRequest{
		ProjectID: project.ID,
		ParentID:  &folder1.ID,
		Name:      "Child Folder",
	}
	body, _ = json.Marshal(createFolderReq)
	req = httptest.NewRequest("POST", "/folders", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var child FolderResponse
	json.Unmarshal(w.Body.Bytes(), &child)

	t.Run("moves folder to new parent", func(t *testing.T) {
		moveReq := MoveItemRequest{
			ProjectID:   project.ID,
			ItemType:    "folder",
			ItemID:      child.ID,
			NewParentID: &folder2.ID,
		}
		body, _ := json.Marshal(moveReq)
		req := httptest.NewRequest("POST", "/workspace/move", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("rejects invalid item type", func(t *testing.T) {
		moveReq := MoveItemRequest{
			ProjectID: project.ID,
			ItemType:  "invalid",
			ItemID:    child.ID,
		}
		body, _ := json.Marshal(moveReq)
		req := httptest.NewRequest("POST", "/workspace/move", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestWorkspaceHandler_RejectCrossProjectMove(t *testing.T) {
	mock := newMockWorkspaceService()
	router, handler := setupWorkspaceTest(t, mock)
	router.POST("/projects", handler.CreateProject)
	router.POST("/folders", handler.CreateFolder)
	router.POST("/workspace/move", handler.MoveItem)

	// Create two projects
	var projects []ProjectResponse
	for i := 0; i < 2; i++ {
		createProjectReq := CreateProjectRequest{Name: "Project " + string(rune('A'+i))}
		body, _ := json.Marshal(createProjectReq)
		req := httptest.NewRequest("POST", "/projects", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		var project ProjectResponse
		json.Unmarshal(w.Body.Bytes(), &project)
		projects = append(projects, project)
	}

	// Create a folder in project A
	createFolderReq := CreateFolderRequest{
		ProjectID: projects[0].ID,
		Name:      "Folder in A",
	}
	body, _ := json.Marshal(createFolderReq)
	req := httptest.NewRequest("POST", "/folders", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var folder FolderResponse
	json.Unmarshal(w.Body.Bytes(), &folder)

	t.Run("rejects cross-project move", func(t *testing.T) {
		// Try to move folder from project A using project B context
		moveReq := MoveItemRequest{
			ProjectID: projects[1].ID, // Wrong project
			ItemType:  "folder",
			ItemID:    folder.ID,
		}
		body, _ := json.Marshal(moveReq)
		req := httptest.NewRequest("POST", "/workspace/move", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Should fail because folder doesn't exist in project B
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestWorkspaceHandler_TrashAndRestore(t *testing.T) {
	mock := newMockWorkspaceService()
	router, handler := setupWorkspaceTest(t, mock)
	router.POST("/projects", handler.CreateProject)
	router.POST("/folders", handler.CreateFolder)
	router.POST("/visualizations", handler.CreateVisualization)
	router.POST("/workspace/trash", handler.TrashItem)
	router.POST("/workspace/restore", handler.RestoreItem)

	// Create a project
	createProjectReq := CreateProjectRequest{Name: "Trash Test Project"}
	body, _ := json.Marshal(createProjectReq)
	req := httptest.NewRequest("POST", "/projects", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var project ProjectResponse
	json.Unmarshal(w.Body.Bytes(), &project)

	// Create a folder
	createFolderReq := CreateFolderRequest{
		ProjectID: project.ID,
		Name:      "Folder to Trash",
	}
	body, _ = json.Marshal(createFolderReq)
	req = httptest.NewRequest("POST", "/folders", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var folder FolderResponse
	json.Unmarshal(w.Body.Bytes(), &folder)

	// Create a visualization
	createVizReq := CreateVisualizationRequest{
		ProjectID: project.ID,
		Name:      "Viz to Trash",
	}
	body, _ = json.Marshal(createVizReq)
	req = httptest.NewRequest("POST", "/visualizations", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var viz VisualizationResponse
	json.Unmarshal(w.Body.Bytes(), &viz)

	t.Run("trashes folder", func(t *testing.T) {
		trashReq := TrashItemRequest{
			ProjectID: project.ID,
			ItemType:  "folder",
			ItemID:    folder.ID,
		}
		body, _ := json.Marshal(trashReq)
		req := httptest.NewRequest("POST", "/workspace/trash", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("restores folder", func(t *testing.T) {
		restoreReq := RestoreItemRequest{
			ProjectID: project.ID,
			ItemType:  "folder",
			ItemID:    folder.ID,
		}
		body, _ := json.Marshal(restoreReq)
		req := httptest.NewRequest("POST", "/workspace/restore", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("trashes visualization", func(t *testing.T) {
		trashReq := TrashItemRequest{
			ProjectID: project.ID,
			ItemType:  "visualization",
			ItemID:    viz.ID,
		}
		body, _ := json.Marshal(trashReq)
		req := httptest.NewRequest("POST", "/workspace/trash", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("restores visualization", func(t *testing.T) {
		restoreReq := RestoreItemRequest{
			ProjectID: project.ID,
			ItemType:  "visualization",
			ItemID:    viz.ID,
		}
		body, _ := json.Marshal(restoreReq)
		req := httptest.NewRequest("POST", "/workspace/restore", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("trashes project", func(t *testing.T) {
		trashReq := TrashItemRequest{
			ProjectID: project.ID,
			ItemType:  "project",
			ItemID:    project.ID,
		}
		body, _ := json.Marshal(trashReq)
		req := httptest.NewRequest("POST", "/workspace/trash", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestWorkspaceHandler_ListFolderContents(t *testing.T) {
	mock := newMockWorkspaceService()
	router, handler := setupWorkspaceTest(t, mock)
	router.POST("/projects", handler.CreateProject)
	router.POST("/folders", handler.CreateFolder)
	router.POST("/visualizations", handler.CreateVisualization)
	router.GET("/folders/contents", handler.ListFolderContents)

	// Create a project
	createProjectReq := CreateProjectRequest{Name: "List Contents Test Project"}
	body, _ := json.Marshal(createProjectReq)
	req := httptest.NewRequest("POST", "/projects", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var project ProjectResponse
	json.Unmarshal(w.Body.Bytes(), &project)

	// Create folders
	for i := 0; i < 2; i++ {
		createFolderReq := CreateFolderRequest{
			ProjectID: project.ID,
			Name:      "Folder " + string(rune('A'+i)),
		}
		body, _ = json.Marshal(createFolderReq)
		req = httptest.NewRequest("POST", "/folders", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}

	// Create visualizations
	for i := 0; i < 2; i++ {
		createVizReq := CreateVisualizationRequest{
			ProjectID: project.ID,
			Name:      "Viz " + string(rune('A'+i)),
		}
		body, _ = json.Marshal(createVizReq)
		req = httptest.NewRequest("POST", "/visualizations", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}

	t.Run("lists mixed contents", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/folders/contents?project_id="+project.ID, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response ListFolderContentsResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, project.ID, response.ProjectID)
		assert.Len(t, response.Items, 4) // 2 folders + 2 visualizations
	})
}
