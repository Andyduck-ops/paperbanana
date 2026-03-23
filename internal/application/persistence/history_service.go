package persistence

import (
	"context"
	"errors"
	"fmt"
	"time"

	domainagent "github.com/paperbanana/paperbanana/internal/domain/workspace"
	"github.com/google/uuid"
)

// HistoryService provides transactional operations for session persistence
// and version history management.
type HistoryService struct {
	txManager TxManager
}

// NewHistoryService creates a new HistoryService.
func NewHistoryService(txManager TxManager) *HistoryService {
	return &HistoryService{txManager: txManager}
}

// SaveSession persists a session record for restore/audit purposes.
// This is called during pipeline execution to checkpoint progress.
func (s *HistoryService) SaveSession(ctx context.Context, session *domainagent.SessionRecord) error {
	return s.txManager.RunInTx(ctx, func(repos Repositories) error {
		// Check if session exists
		existing, err := repos.Sessions().GetByID(ctx, session.ID)
		if err != nil {
			// Session doesn't exist, create it
			return repos.Sessions().Create(ctx, session)
		}

		// Update existing session
		existing.Status = session.Status
		existing.CurrentStage = session.CurrentStage
		existing.Snapshot = session.Snapshot
		existing.UpdatedAt = time.Now().UTC()
		if session.CompletedAt != nil {
			existing.CompletedAt = session.CompletedAt
		}
		return repos.Sessions().Update(ctx, existing)
	})
}

// PublishVersion saves the session and creates a new immutable version.
// This is called after a successful pipeline run to record history.
// Returns the new version ID.
func (s *HistoryService) PublishVersion(ctx context.Context, session *domainagent.SessionRecord, summary string) (string, error) {
	var versionID string

	err := s.txManager.RunInTx(ctx, func(repos Repositories) error {
		// Save or update the session
		existing, err := repos.Sessions().GetByID(ctx, session.ID)
		if err != nil {
			if err := repos.Sessions().Create(ctx, session); err != nil {
				return fmt.Errorf("create session: %w", err)
			}
		} else {
			existing.Status = session.Status
			existing.CurrentStage = session.CurrentStage
			existing.Snapshot = session.Snapshot
			existing.UpdatedAt = time.Now().UTC()
			if session.CompletedAt != nil {
				existing.CompletedAt = session.CompletedAt
			}
			if err := repos.Sessions().Update(ctx, existing); err != nil {
				return fmt.Errorf("update session: %w", err)
			}
		}

		// Only create versions for sessions with a visualization
		if session.VisualizationID == nil {
			return errors.New("cannot publish version without visualization")
		}

		// Get the latest version number
		latestVersion, err := repos.Versions().GetLatestByVisualization(ctx, session.ProjectID, *session.VisualizationID)
		nextNumber := 1
		if err == nil {
			nextNumber = latestVersion.VersionNumber + 1
		}

		// Create the new version
		versionID = uuid.NewString()
		version := &domainagent.VisualizationVersion{
			ID:              versionID,
			VisualizationID: *session.VisualizationID,
			ProjectID:       session.ProjectID,
			VersionNumber:   nextNumber,
			SessionID:       session.ID,
			Summary:         summary,
			CreatedAt:       time.Now().UTC(),
			SessionSnapshot: session.Snapshot,
		}

		if err := repos.Versions().Create(ctx, version); err != nil {
			return fmt.Errorf("create version: %w", err)
		}

		// Update the visualization's current version pointer
		if err := repos.Visualizations().SetCurrentVersion(ctx, session.ProjectID, *session.VisualizationID, versionID); err != nil {
			return fmt.Errorf("set current version: %w", err)
		}

		return nil
	})

	if err != nil {
		return "", err
	}
	return versionID, nil
}

// ListHistory returns the version history for a visualization.
// Results are ordered by version number descending (most recent first).
func (s *HistoryService) ListHistory(ctx context.Context, projectID, visualizationID string, limit int) ([]*domainagent.VisualizationVersion, error) {
	var versions []*domainagent.VisualizationVersion

	err := s.txManager.ReadOnlyTx(ctx, func(repos Repositories) error {
		var err error
		versions, err = repos.Versions().ListByVisualization(ctx, projectID, visualizationID, limit)
		return err
	})

	return versions, err
}

// GetVersion retrieves a specific version by ID.
func (s *HistoryService) GetVersion(ctx context.Context, projectID, versionID string) (*domainagent.VisualizationVersion, error) {
	var version *domainagent.VisualizationVersion

	err := s.txManager.ReadOnlyTx(ctx, func(repos Repositories) error {
		var err error
		version, err = repos.Versions().GetByID(ctx, projectID, versionID)
		return err
	})

	return version, err
}

// LoadResumableSession finds the most recent session for a visualization
// that can be resumed (typically the last failed or incomplete session).
func (s *HistoryService) LoadResumableSession(ctx context.Context, projectID, visualizationID string) (*domainagent.SessionRecord, error) {
	var session *domainagent.SessionRecord

	err := s.txManager.ReadOnlyTx(ctx, func(repos Repositories) error {
		sessions, err := repos.Sessions().GetByVisualization(ctx, projectID, visualizationID, 1)
		if err != nil {
			return err
		}
		if len(sessions) == 0 {
			return errors.New("no resumable session found")
		}
		session = sessions[0]
		return nil
	})

	return session, err
}

// GetLatestSession retrieves the most recent session for a visualization.
func (s *HistoryService) GetLatestSession(ctx context.Context, projectID, visualizationID string) (*domainagent.SessionRecord, error) {
	return s.LoadResumableSession(ctx, projectID, visualizationID)
}

// GetSessionByID retrieves a session by its ID.
func (s *HistoryService) GetSessionByID(ctx context.Context, sessionID string) (*domainagent.SessionRecord, error) {
	var session *domainagent.SessionRecord

	err := s.txManager.ReadOnlyTx(ctx, func(repos Repositories) error {
		var err error
		session, err = repos.Sessions().GetByID(ctx, sessionID)
		return err
	})

	return session, err
}
