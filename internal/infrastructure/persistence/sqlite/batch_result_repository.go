package sqlite

import (
	"time"

	domainagent "github.com/paperbanana/paperbanana/internal/domain/agent"
	"gorm.io/gorm"
)

// BatchResultRepository persists batch execution results.
type BatchResultRepository struct {
	db *gorm.DB
}

// NewBatchResultRepository creates a new batch result repository.
func NewBatchResultRepository(db *gorm.DB) *BatchResultRepository {
	return &BatchResultRepository{db: db}
}

// Save persists a batch result to the database.
func (r *BatchResultRepository) Save(result *domainagent.BatchResult) error {
	model := &BatchResultModel{
		ID:          result.BatchID,
		Status:      "completed",
		Successful:  result.Successful,
		Failed:      result.Failed,
		StartedAt:   result.Timing.StartedAt,
		CompletedAt: result.Timing.CompletedAt,
		DurationMs:  result.Timing.Duration.Milliseconds(),
		ResultsJSON: BatchResultsPayload{Results: result.Results},
		CreatedAt:   time.Now().UTC(),
	}
	return r.db.Create(model).Error
}

// Get retrieves a batch result by ID.
func (r *BatchResultRepository) Get(batchID string) (*domainagent.BatchResult, error) {
	var model BatchResultModel
	if err := r.db.First(&model, "id = ?", batchID).Error; err != nil {
		return nil, err
	}

	return &domainagent.BatchResult{
		BatchID:    model.ID,
		Results:    model.ResultsJSON.Results,
		Successful: model.Successful,
		Failed:     model.Failed,
		Timing: domainagent.BatchTiming{
			StartedAt:   model.StartedAt,
			CompletedAt: model.CompletedAt,
			Duration:    time.Duration(model.DurationMs) * time.Millisecond,
		},
	}, nil
}

// Delete removes a batch result by ID.
func (r *BatchResultRepository) Delete(batchID string) error {
	return r.db.Delete(&BatchResultModel{}, "id = ?", batchID).Error
}

// DeleteOlderThan removes batch results older than the specified duration.
func (r *BatchResultRepository) DeleteOlderThan(olderThan time.Duration) (int64, error) {
	cutoff := time.Now().UTC().Add(-olderThan)
	result := r.db.Where("created_at < ?", cutoff).Delete(&BatchResultModel{})
	return result.RowsAffected, result.Error
}
