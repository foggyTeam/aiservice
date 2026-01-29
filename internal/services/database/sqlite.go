package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/aiservice/internal/config"
	"github.com/aiservice/internal/models"

	_ "github.com/mattn/go-sqlite3"
)

type SQLiteJobStorage struct {
	db *sql.DB
}

// NewSQLiteStorage creates a new SQLite storage instance
func NewSQLiteStorage(cfg config.DatabaseConfig) (*SQLiteJobStorage, error) {
	db, err := sql.Open("sqlite3", cfg.FilePath+"?cache=shared&_busy_timeout=10000")
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Create the jobs table if it doesn't exist
	if err := createJobsTable(db); err != nil {
		return nil, fmt.Errorf("failed to create jobs table: %w", err)
	}

	storage := &SQLiteJobStorage{
		db: db,
	}

	return storage, nil
}

// createJobsTable creates the jobs table if it doesn't exist
func createJobsTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS jobs (
		id TEXT PRIMARY KEY,
		request_type TEXT NOT NULL,
		request_data TEXT NOT NULL,
		created_at INTEGER NOT NULL,
		retries INTEGER DEFAULT 0,
		status TEXT NOT NULL DEFAULT 'pending',
		result_data TEXT
	);
	CREATE INDEX IF NOT EXISTS idx_jobs_status ON jobs(status);
	CREATE INDEX IF NOT EXISTS idx_jobs_created_at ON jobs(created_at);
	`

	_, err := db.Exec(query)
	return err
}

func (s *SQLiteJobStorage) Save(job models.Job) error {
	requestData, err := json.Marshal(job.Request)
	if err != nil {
		return fmt.Errorf("failed to marshal request data: %w", err)
	}

	query := `
	INSERT INTO jobs (id, request_type, request_data, created_at, retries, status)
	VALUES (?, ?, ?, ?, ?, ?)
	ON CONFLICT(id) DO UPDATE SET
		request_type = excluded.request_type,
		request_data = excluded.request_data,
		created_at = excluded.created_at,
		retries = excluded.retries,
		status = excluded.status
	`

	_, err = s.db.Exec(query, job.ID, job.Request.RequestType, string(requestData), job.CreatedAt, job.Retries, string(job.Status))
	if err != nil {
		return fmt.Errorf("failed to save job: %w", err)
	}

	return nil
}

func (s *SQLiteJobStorage) Get(id string) (models.Job, error) {
	query := "SELECT id, request_type, request_data, created_at, retries, status, result_data FROM jobs WHERE id = ?"
	row := s.db.QueryRow(query, id)

	var jobID, requestType, requestData, status, resultData sql.NullString
	var createdAt int64
	var retries int

	err := row.Scan(&jobID, &requestType, &requestData, &createdAt, &retries, &status, &resultData)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.Job{}, fmt.Errorf("job not found")
		}
		return models.Job{}, fmt.Errorf("failed to get job: %w", err)
	}

	var request models.AnalyzeRequest
	if err := json.Unmarshal([]byte(requestData.String), &request); err != nil {
		return models.Job{}, fmt.Errorf("failed to unmarshal request data: %w", err)
	}

	job := models.Job{
		ID:        jobID.String,
		Request:   request,
		CreatedAt: createdAt,
		Retries:   retries,
		Status:    models.JobStatus(status.String),
	}

	return job, nil
}

func (s *SQLiteJobStorage) Update(job models.Job) error {
	requestData, err := json.Marshal(job.Request)
	if err != nil {
		return fmt.Errorf("failed to marshal request data: %w", err)
	}

	var resultData *string
	if job.Status == models.JobStatusCompleted {
		// In a real implementation, we would store the result data
		// For now, we'll leave it as null
	}

	query := `
	UPDATE jobs
	SET request_type = ?, request_data = ?, created_at = ?, retries = ?, status = ?, result_data = ?
	WHERE id = ?
	`

	_, err = s.db.Exec(query,
		job.Request.RequestType,
		string(requestData),
		job.CreatedAt,
		job.Retries,
		string(job.Status),
		resultData,
		job.ID)

	if err != nil {
		return fmt.Errorf("failed to update job: %w", err)
	}

	return nil
}

func (s *SQLiteJobStorage) Abort(ctx context.Context, id string) error {
	query := "UPDATE jobs SET status = ? WHERE id = ?"
	result, err := s.db.Exec(query, string(models.JobStatusAborted), id)
	if err != nil {
		return fmt.Errorf("failed to abort job: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("job not found")
	}

	return nil
}

func (s *SQLiteJobStorage) GetAll() ([]models.Job, error) {
	query := "SELECT id, request_type, request_data, created_at, retries, status, result_data FROM jobs ORDER BY created_at DESC"
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all jobs: %w", err)
	}
	defer rows.Close()

	var jobs []models.Job
	for rows.Next() {
		var jobID, requestType, requestData, status, resultData sql.NullString
		var createdAt int64
		var retries int

		err := rows.Scan(&jobID, &requestType, &requestData, &createdAt, &retries, &status, &resultData)
		if err != nil {
			return nil, fmt.Errorf("failed to scan job row: %w", err)
		}

		var request models.AnalyzeRequest
		if err := json.Unmarshal([]byte(requestData.String), &request); err != nil {
			slog.Error("failed to unmarshal request data", "job_id", jobID.String, "error", err)
			continue
		}

		job := models.Job{
			ID:        jobID.String,
			Request:   request,
			CreatedAt: createdAt,
			Retries:   retries,
			Status:    models.JobStatus(status.String),
		}

		jobs = append(jobs, job)
	}

	return jobs, nil
}

func (s *SQLiteJobStorage) DeleteJobs(ids ...string) error {
	if len(ids) == 0 {
		return nil
	}

	// Build a parameterized query for deletion
	placeholders := make([]string, len(ids))
	args := make([]any, len(ids))

	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf("DELETE FROM jobs WHERE id IN (%s)", strings.Join(placeholders, ","))

	result, err := s.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to delete jobs: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	slog.Info("deleted jobs", "count", rowsAffected)

	return nil
}

// Close closes the database connection
func (s *SQLiteJobStorage) Close() error {
	return s.db.Close()
}
