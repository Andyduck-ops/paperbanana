package persistence

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"

	domainagent "github.com/paperbanana/paperbanana/internal/domain/agent"
)

// BatchResultStore 提供批处理结果的持久化存储
type BatchResultStore struct {
	db *sql.DB
}

// NewBatchResultStore 创建一个新的批处理结果存储
func NewBatchResultStore(dbPath string) (*BatchResultStore, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// 设置连接池参数
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	store := &BatchResultStore{db: db}
	if err := store.createTable(); err != nil {
		db.Close()
		return nil, fmt.Errorf("create table: %w", err)
	}

	return store, nil
}

// createTable 创建 batch_results 表
func (s *BatchResultStore) createTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS batch_results (
		batch_id TEXT PRIMARY KEY,
		status TEXT NOT NULL,
		successful INTEGER NOT NULL DEFAULT 0,
		failed INTEGER NOT NULL DEFAULT 0,
		results_json TEXT NOT NULL,
		started_at DATETIME NOT NULL,
		completed_at DATETIME,
		duration_ms INTEGER,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_batch_results_status ON batch_results(status);
	CREATE INDEX IF NOT EXISTS idx_batch_results_started_at ON batch_results(started_at);

	-- 创建触发器自动更新 updated_at
	CREATE TRIGGER IF NOT EXISTS batch_results_updated_at 
	AFTER UPDATE ON batch_results
	BEGIN
		UPDATE batch_results SET updated_at = CURRENT_TIMESTAMP WHERE batch_id = NEW.batch_id;
	END;
	`

	if _, err := s.db.Exec(query); err != nil {
		return fmt.Errorf("create batch_results table: %w", err)
	}

	return nil
}

// Save 保存批处理结果
func (s *BatchResultStore) Save(result *domainagent.BatchResult) error {
	if result == nil {
		return fmt.Errorf("batch result is nil")
	}

	// 序列化结果
	resultsJSON, err := json.Marshal(result.Results)
	if err != nil {
		return fmt.Errorf("marshal results: %w", err)
	}

	// 确定状态
	status := "running"
	if result.Successful+result.Failed == len(result.Results) && len(result.Results) > 0 {
		if result.Failed == 0 {
			status = "completed"
		} else if result.Successful == 0 {
			status = "failed"
		} else {
			status = "partial"
		}
	}

	// 计算持续时间（毫秒）
	var durationMs int64
	if result.Timing.Duration > 0 {
		durationMs = result.Timing.Duration.Milliseconds()
	}

	// 使用 UPSERT 语义
	query := `
		INSERT INTO batch_results (
			batch_id, status, successful, failed, results_json, 
			started_at, completed_at, duration_ms
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(batch_id) DO UPDATE SET
			status = excluded.status,
			successful = excluded.successful,
			failed = excluded.failed,
			results_json = excluded.results_json,
			completed_at = excluded.completed_at,
			duration_ms = excluded.duration_ms
	`

	var completedAt interface{}
	if !result.Timing.CompletedAt.IsZero() {
		completedAt = result.Timing.CompletedAt
	} else {
		completedAt = nil
	}

	_, err = s.db.Exec(
		query,
		result.BatchID,
		status,
		result.Successful,
		result.Failed,
		string(resultsJSON),
		result.Timing.StartedAt,
		completedAt,
		durationMs,
	)
	if err != nil {
		return fmt.Errorf("save batch result: %w", err)
	}

	return nil
}

// Get 获取批处理结果
func (s *BatchResultStore) Get(batchID string) (*domainagent.BatchResult, error) {
	query := `
		SELECT batch_id, successful, failed, results_json, 
		       started_at, completed_at, duration_ms
		FROM batch_results
		WHERE batch_id = ?
	`

	var result domainagent.BatchResult
	var resultsJSON string
	var startedAt time.Time
	var completedAt sql.NullTime
	var durationMs int64

	err := s.db.QueryRow(query, batchID).Scan(
		&result.BatchID,
		&result.Successful,
		&result.Failed,
		&resultsJSON,
		&startedAt,
		&completedAt,
		&durationMs,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("batch result not found: %s", batchID)
		}
		return nil, fmt.Errorf("get batch result: %w", err)
	}

	// 反序列化结果
	if err := json.Unmarshal([]byte(resultsJSON), &result.Results); err != nil {
		return nil, fmt.Errorf("unmarshal results: %w", err)
	}

	// 设置时间信息
	result.Timing.StartedAt = startedAt
	if completedAt.Valid {
		result.Timing.CompletedAt = completedAt.Time
		result.Timing.Duration = time.Duration(durationMs) * time.Millisecond
	}

	return &result, nil
}

// List 列出批处理结果（支持分页）
func (s *BatchResultStore) List(limit, offset int) ([]*domainagent.BatchResult, error) {
	query := `
		SELECT batch_id, successful, failed, results_json, 
		       started_at, completed_at, duration_ms
		FROM batch_results
		ORDER BY started_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := s.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list batch results: %w", err)
	}
	defer rows.Close()

	var results []*domainagent.BatchResult
	for rows.Next() {
		var result domainagent.BatchResult
		var resultsJSON string
		var startedAt time.Time
		var completedAt sql.NullTime
		var durationMs int64

		if err := rows.Scan(
			&result.BatchID,
			&result.Successful,
			&result.Failed,
			&resultsJSON,
			&startedAt,
			&completedAt,
			&durationMs,
		); err != nil {
			return nil, fmt.Errorf("scan batch result: %w", err)
		}

		if err := json.Unmarshal([]byte(resultsJSON), &result.Results); err != nil {
			return nil, fmt.Errorf("unmarshal results: %w", err)
		}

		result.Timing.StartedAt = startedAt
		if completedAt.Valid {
			result.Timing.CompletedAt = completedAt.Time
			result.Timing.Duration = time.Duration(durationMs) * time.Millisecond
		}

		results = append(results, &result)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate batch results: %w", err)
	}

	return results, nil
}

// Delete 删除批处理结果
func (s *BatchResultStore) Delete(batchID string) error {
	query := `DELETE FROM batch_results WHERE batch_id = ?`
	result, err := s.db.Exec(query, batchID)
	if err != nil {
		return fmt.Errorf("delete batch result: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("batch result not found: %s", batchID)
	}

	return nil
}

// CleanupOldResults 清理旧的批处理结果
func (s *BatchResultStore) CleanupOldResults(before time.Time) (int64, error) {
	query := `DELETE FROM batch_results WHERE started_at < ?`
	result, err := s.db.Exec(query, before)
	if err != nil {
		return 0, fmt.Errorf("cleanup old results: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("get rows affected: %w", err)
	}

	return rowsAffected, nil
}

// Close 关闭数据库连接
func (s *BatchResultStore) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}
