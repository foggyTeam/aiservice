package database

import (
	"github.com/aiservice/internal/config"
	jobservice "github.com/aiservice/internal/services/jobService"
	"github.com/aiservice/internal/services/storage"
)

// NewStorage creates a new storage implementation based on the configuration
func NewStorage(cfg config.DatabaseConfig) (jobservice.JobStorage, error) {
	switch cfg.Type {
	case "memory":
		return storage.NewInMemoryJobStorage(), nil
	case "sqlite":
		return NewSQLiteStorage(cfg)
	default:
		return NewSQLiteStorage(cfg) // Default to SQLite
	}
}