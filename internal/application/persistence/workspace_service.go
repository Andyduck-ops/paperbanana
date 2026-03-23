package persistence

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/paperbanana/paperbanana/internal/domain/workspace"
)

// FolderContents represents mixed folder listing results.
type FolderContents struct {
	Folders        []*workspace.Folder
	Visualizations []*workspace.Visualization
}

// WorkspaceService provides workspace use cases for create, move, list, trash, and restore.
type WorkspaceService struct {
	txManager TxManager
}

// NewWorkspaceService creates a new WorkspaceService.
func NewWorkspaceService(txManager TxManager) *WorkspaceService {
	return &WorkspaceService{txManager: txManager}
}

// CreateProject creates a new project.
func (s *WorkspaceService) CreateProject(ctx context.Context, name, description string) (*workspace.Project, error) {
	project := &workspace.Project{
		ID:          uuid.New().String(),
		Name:        name,
		Description: description,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	err := s.txManager.RunInTx(ctx, func(repos Repositories) error {
		return repos.Projects().Create(ctx, project)
	})
	if err != nil {
		return nil, err
	}
	return project, nil
}

// GetProject retrieves a project by ID.
func (s *WorkspaceService) GetProject(ctx context.Context, id string) (*workspace.Project, error) {
	var project *workspace.Project
	err := s.txManager.ReadOnlyTx(ctx, func(repos Repositories) error {
		var err error
		project, err = repos.Projects().GetByID(ctx, id)
		return err
	})
	return project, err
}

// ListProjects retrieves all projects.
func (s *WorkspaceService) ListProjects(ctx context.Context) ([]*workspace.Project, error) {
	var projects []*workspace.Project
	err := s.txManager.ReadOnlyTx(ctx, func(repos Repositories) error {
		var err error
		projects, err = repos.Projects().List(ctx)
		return err
	})
	return projects, err
}

// CreateFolder creates a new folder within a project.
func (s *WorkspaceService) CreateFolder(ctx context.Context, projectID string, parentID *string, name string) (*workspace.Folder, error) {
	folder := &workspace.Folder{
		ID:        uuid.New().String(),
		ProjectID: projectID,
		ParentID:  parentID,
		Name:      name,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	err := s.txManager.RunInTx(ctx, func(repos Repositories) error {
		// Verify project exists
		_, err := repos.Projects().GetByID(ctx, projectID)
		if err != nil {
			return err
		}
		return repos.Folders().Create(ctx, folder)
	})
	if err != nil {
		return nil, err
	}
	return folder, nil
}

// CreateVisualization creates a new visualization within a project.
func (s *WorkspaceService) CreateVisualization(ctx context.Context, projectID string, folderID *string, name string) (*workspace.Visualization, error) {
	viz := &workspace.Visualization{
		ID:        uuid.New().String(),
		ProjectID: projectID,
		FolderID:  folderID,
		Name:      name,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	err := s.txManager.RunInTx(ctx, func(repos Repositories) error {
		// Verify project exists
		_, err := repos.Projects().GetByID(ctx, projectID)
		if err != nil {
			return err
		}
		return repos.Visualizations().Create(ctx, viz)
	})
	if err != nil {
		return nil, err
	}
	return viz, nil
}

// MoveFolder moves a folder to a new parent within the same project.
func (s *WorkspaceService) MoveFolder(ctx context.Context, projectID, folderID string, newParentID *string) error {
	return s.txManager.RunInTx(ctx, func(repos Repositories) error {
		folder, err := repos.Folders().GetByID(ctx, projectID, folderID)
		if err != nil {
			return err
		}
		folder.ParentID = newParentID
		return repos.Folders().Update(ctx, folder)
	})
}

// MoveVisualization moves a visualization to a new folder within the same project.
func (s *WorkspaceService) MoveVisualization(ctx context.Context, projectID, vizID string, newFolderID *string) error {
	return s.txManager.RunInTx(ctx, func(repos Repositories) error {
		viz, err := repos.Visualizations().GetByID(ctx, projectID, vizID)
		if err != nil {
			return err
		}
		viz.FolderID = newFolderID
		return repos.Visualizations().Update(ctx, viz)
	})
}

// TrashFolder soft-deletes a folder.
func (s *WorkspaceService) TrashFolder(ctx context.Context, projectID, folderID string) error {
	return s.txManager.RunInTx(ctx, func(repos Repositories) error {
		return repos.Folders().Delete(ctx, projectID, folderID)
	})
}

// TrashVisualization soft-deletes a visualization.
func (s *WorkspaceService) TrashVisualization(ctx context.Context, projectID, vizID string) error {
	return s.txManager.RunInTx(ctx, func(repos Repositories) error {
		return repos.Visualizations().Delete(ctx, projectID, vizID)
	})
}

// TrashProject soft-deletes a project.
func (s *WorkspaceService) TrashProject(ctx context.Context, projectID string) error {
	return s.txManager.RunInTx(ctx, func(repos Repositories) error {
		return repos.Projects().Delete(ctx, projectID)
	})
}

// RestoreFolder restores a soft-deleted folder.
func (s *WorkspaceService) RestoreFolder(ctx context.Context, projectID, folderID string) error {
	return s.txManager.RunInTx(ctx, func(repos Repositories) error {
		return repos.Folders().Restore(ctx, projectID, folderID)
	})
}

// RestoreVisualization restores a soft-deleted visualization.
func (s *WorkspaceService) RestoreVisualization(ctx context.Context, projectID, vizID string) error {
	return s.txManager.RunInTx(ctx, func(repos Repositories) error {
		return repos.Visualizations().Restore(ctx, projectID, vizID)
	})
}

// ListFolderContents returns mixed folder contents (folders and visualizations).
func (s *WorkspaceService) ListFolderContents(ctx context.Context, projectID string, folderID *string) (*FolderContents, error) {
	var contents FolderContents

	err := s.txManager.ReadOnlyTx(ctx, func(repos Repositories) error {
		folders, err := repos.Folders().ListByParent(ctx, projectID, folderID)
		if err != nil {
			return err
		}
		contents.Folders = folders

		vizs, err := repos.Visualizations().ListByFolder(ctx, projectID, folderID)
		if err != nil {
			return err
		}
		contents.Visualizations = vizs
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &contents, nil
}
