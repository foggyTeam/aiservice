package database

import (
	"fmt"
	"testing"

	"github.com/aiservice/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestNewStorage(t *testing.T) {
	tt := []struct {
		dbType string
		err    error
	}{
		{dbType: "memory", err: nil},
		{dbType: "", err: fmt.Errorf("invalid storage")},
	}
	for _, test := range tt {
		_, err := NewStorage(config.DatabaseConfig{Type: test.dbType})
		assert.Equal(t, err, test.err)
	}
}
