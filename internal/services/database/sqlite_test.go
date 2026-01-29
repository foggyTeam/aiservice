package database

import (
	"os"
	"testing"

	"github.com/aiservice/internal/config"
	"github.com/aiservice/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestSQLiteStorage(t *testing.T) {
	// Create a temporary SQLite file for testing
	tempDBFile := "./test_db.sqlite"
	defer os.Remove(tempDBFile) // Clean up after test

	// Create SQLite storage
	cfg := config.DatabaseConfig{
		Type:     "sqlite",
		FilePath: tempDBFile,
	}
	storage, err := NewSQLiteStorage(cfg)
	assert.NoError(t, err)
	defer storage.Close()

	// Test saving a job
	job := models.Job{
		ID:        "test-job-1",
		CreatedAt: 1234567890,
		Status:    models.JobStatusPending,
		Request: models.AnalyzeRequest{
			RequestType: "test",
		},
	}

	err = storage.Save(job)
	assert.NoError(t, err)

	// Test getting the job back
	retrievedJob, err := storage.Get("test-job-1")
	assert.NoError(t, err)
	assert.Equal(t, job.ID, retrievedJob.ID)
	assert.Equal(t, job.Status, retrievedJob.Status)
	assert.Equal(t, job.Request.RequestType, retrievedJob.Request.RequestType)

	// Test updating the job
	job.Status = models.JobStatusRunning
	err = storage.Update(job)
	assert.NoError(t, err)

	updatedJob, err := storage.Get("test-job-1")
	assert.NoError(t, err)
	assert.Equal(t, models.JobStatusRunning, updatedJob.Status)

	// Test aborting the job
	err = storage.Abort(nil, "test-job-1")
	assert.NoError(t, err)

	abortedJob, err := storage.Get("test-job-1")
	assert.NoError(t, err)
	assert.Equal(t, models.JobStatusAborted, abortedJob.Status)

	// Test deleting the job
	err = storage.DeleteJobs("test-job-1")
	assert.NoError(t, err)

	// Verify the job is gone
	_, err = storage.Get("test-job-1")
	assert.Error(t, err)
}