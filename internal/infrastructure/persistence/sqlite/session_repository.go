package sqlite

import (
	"context"
	"errors"
	"fmt"

	domainagent "github.com/paperbanana/paperbanana/internal/domain/agent"
	"github.com/paperbanana/paperbanana/internal/domain/workspace"
	"gorm.io/gorm"
)

// SessionRepository implements workspace.SessionRepository using SQLite/GORM.
type SessionRepository struct {
	db *gorm.DB
}

// NewSessionRepository creates a new SessionRepository.
func NewSessionRepository(db *gorm.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

// Create persists a new session record.
func (r *SessionRepository) Create(ctx context.Context, session *workspace.SessionRecord) error {
	if session == nil {
		return errors.New("session cannot be nil")
	}

	model := sessionToModel(session)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return fmt.Errorf("create session: %w", err)
	}
	return nil
}

// GetByID retrieves a session by its ID.
func (r *SessionRepository) GetByID(ctx context.Context, id string) (*workspace.SessionRecord, error) {
	var model SessionModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("session not found: %s", id)
		}
		return nil, fmt.Errorf("get session: %w", err)
	}
	return modelToSession(&model), nil
}

// GetByProject retrieves sessions for a project, ordered by creation time descending.
func (r *SessionRepository) GetByProject(ctx context.Context, projectID string, limit int) ([]*workspace.SessionRecord, error) {
	query := r.db.WithContext(ctx).
		Where("project_id = ?", projectID).
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	var models []SessionModel
	if err := query.Find(&models).Error; err != nil {
		return nil, fmt.Errorf("get sessions by project: %w", err)
	}

	sessions := make([]*workspace.SessionRecord, len(models))
	for i, model := range models {
		sessions[i] = modelToSession(&model)
	}
	return sessions, nil
}

// GetByVisualization retrieves sessions for a visualization within a project.
func (r *SessionRepository) GetByVisualization(ctx context.Context, projectID, visualizationID string, limit int) ([]*workspace.SessionRecord, error) {
	query := r.db.WithContext(ctx).
		Where("project_id = ? AND visualization_id = ?", projectID, visualizationID).
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	var models []SessionModel
	if err := query.Find(&models).Error; err != nil {
		return nil, fmt.Errorf("get sessions by visualization: %w", err)
	}

	sessions := make([]*workspace.SessionRecord, len(models))
	for i, model := range models {
		sessions[i] = modelToSession(&model)
	}
	return sessions, nil
}

// Update modifies an existing session record.
func (r *SessionRepository) Update(ctx context.Context, session *workspace.SessionRecord) error {
	if session == nil {
		return errors.New("session cannot be nil")
	}

	model := sessionToModel(session)
	result := r.db.WithContext(ctx).Save(model)
	if result.Error != nil {
		return fmt.Errorf("update session: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("session not found: %s", session.ID)
	}
	return nil
}

// sessionToModel converts a domain SessionRecord to a GORM SessionModel.
func sessionToModel(session *workspace.SessionRecord) *SessionModel {
	model := &SessionModel{
		ID:            session.ID,
		ProjectID:     session.ProjectID,
		Status:        session.Status,
		CurrentStage:  session.CurrentStage,
		SchemaVersion: session.SchemaVersion,
		CreatedAt:     session.CreatedAt,
		UpdatedAt:     session.UpdatedAt,
		CompletedAt:   session.CompletedAt,
	}

	if session.VisualizationID != nil {
		model.VisualizationID = session.VisualizationID
	}

	if session.Snapshot != nil {
		model.SnapshotJSON = SessionSnapshotPayload{
			SchemaVersion: session.Snapshot.SchemaVersion,
			SessionID:     session.Snapshot.SessionID,
			RequestID:     session.Snapshot.RequestID,
			Status:        session.Snapshot.Status,
			CurrentStage:  session.Snapshot.CurrentStage,
			Pipeline:      session.Snapshot.Pipeline,
			InitialInput:  session.Snapshot.InitialInput,
			StageStates:   session.Snapshot.StageStates,
			FinalOutput:   session.Snapshot.FinalOutput,
			Error:         session.Snapshot.Error,
			Restore:       session.Snapshot.Restore,
			Metadata:      session.Snapshot.Metadata,
			StartedAt:     session.Snapshot.StartedAt,
			UpdatedAt:     session.Snapshot.UpdatedAt,
			CompletedAt:   session.Snapshot.CompletedAt,
		}
	}

	return model
}

// modelToSession converts a GORM SessionModel to a domain SessionRecord.
func modelToSession(model *SessionModel) *workspace.SessionRecord {
	session := &workspace.SessionRecord{
		ID:            model.ID,
		ProjectID:     model.ProjectID,
		Status:        model.Status,
		CurrentStage:  model.CurrentStage,
		SchemaVersion: model.SchemaVersion,
		CreatedAt:     model.CreatedAt,
		UpdatedAt:     model.UpdatedAt,
		CompletedAt:   model.CompletedAt,
	}

	if model.VisualizationID != nil && *model.VisualizationID != "" {
		session.VisualizationID = model.VisualizationID
	}

	if model.SnapshotJSON.SessionID != "" {
		session.Snapshot = &domainagent.SessionState{
			SchemaVersion: model.SnapshotJSON.SchemaVersion,
			SessionID:     model.SnapshotJSON.SessionID,
			RequestID:     model.SnapshotJSON.RequestID,
			Status:        model.SnapshotJSON.Status,
			CurrentStage:  model.SnapshotJSON.CurrentStage,
			Pipeline:      model.SnapshotJSON.Pipeline,
			InitialInput:  model.SnapshotJSON.InitialInput,
			StageStates:   model.SnapshotJSON.StageStates,
			FinalOutput:   model.SnapshotJSON.FinalOutput,
			Error:         model.SnapshotJSON.Error,
			Restore:       model.SnapshotJSON.Restore,
			Metadata:      model.SnapshotJSON.Metadata,
			StartedAt:     model.SnapshotJSON.StartedAt,
			UpdatedAt:     model.SnapshotJSON.UpdatedAt,
			CompletedAt:   model.SnapshotJSON.CompletedAt,
		}
	}

	return session
}

// Ensure SessionRepository implements workspace.SessionRepository.
var _ workspace.SessionRepository = (*SessionRepository)(nil)
