package preprocessing

import (
	"fmt"
	"strings"
	"testing"

	"github.com/aiservice/internal/models"
	"github.com/google/uuid"
)

func TestCreateFileHierarchyDescription(t *testing.T) {
	uuid0 := uuid.New().String()
	uuid1 := uuid.New().String()
	uuid2 := uuid.New().String()
	uuid3 := uuid.New().String()
	tests := []struct {
		name     string
		file     models.File
		expected string
	}{
		{
			name:     "empty file",
			file:     models.File{},
			expected: "",
		},
		{
			name: "file without children",
			file: models.File{
				Name: "test",
				Type: "doc",
				Id:   uuid0,
			},
			expected: fmt.Sprintf("%s.%s.%s", "test", uuid0, "doc") + "\n",
		},
		{
			name: "file with children",
			file: models.File{
				Name: "root",
				Type: "section",
				Id:   uuid0,
				Children: []models.File{
					{Name: "child1", Type: "simple", Id: uuid1},
					{Name: "child2", Type: "section", Id: uuid2, Children: []models.File{
						{Name: "grandchild", Type: "graph", Id: uuid3},
					}},
				},
			},
			expected: fmt.Sprintf("%s.%s.%s", "root", uuid0, "section") +
				"\n├── " + fmt.Sprintf("%s.%s.%s", "child1", uuid1, "simple") +
				"\n└── " + fmt.Sprintf("%s.%s.%s", "child2", uuid2, "section") +
				"\n    └── " + fmt.Sprintf("%s.%s.%s", "grandchild", uuid3, "graph") + "\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := createFileHierarchyDescription(tt.file)
			if result != tt.expected {
				t.Errorf("createFileHierarchyDescription() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestWriteFileTree(t *testing.T) {
	uuid0 := uuid.New().String()
	uuid1 := uuid.New().String()
	tests := []struct {
		name     string
		files    []models.File
		prefix   string
		expected string
	}{
		{
			name:     "empty files",
			files:    []models.File{},
			prefix:   "",
			expected: "",
		},
		{
			name: "single file",
			files: []models.File{
				{Name: "file.txt", Type: "doc", Id: uuid0},
			},
			prefix:   "",
			expected: "└── " + fmt.Sprintf("%s.%s.%s", "file.txt", uuid0, "doc") + "\n",
		},
		{
			name: "multiple files",
			files: []models.File{
				{Name: "file1", Type: "doc", Id: uuid0},
				{Name: "file2", Type: "graph", Id: uuid1},
			},
			prefix:   "",
			expected: "├── " + fmt.Sprintf("%s.%s.%s", "file1", uuid0, "doc") + "\n└── " + fmt.Sprintf("%s.%s.%s", "file2", uuid1, "graph") + "\n",
		},
		{
			name: "nested files",
			files: []models.File{
				{Name: "dir", Type: "section", Id: uuid0, Children: []models.File{
					{Name: "nested", Type: "doc", Id: uuid1},
				}},
			},
			prefix:   "",
			expected: "└── " + fmt.Sprintf("%s.%s.%s", "dir", uuid0, "section") + "\n    └── " + fmt.Sprintf("%s.%s.%s", "nested", uuid1, "doc") + "\n",
		},
		{
			name: "with prefix",
			files: []models.File{
				{Name: "file", Type: "doc", Id: uuid0},
			},
			prefix:   "│   ",
			expected: "│   └── " + fmt.Sprintf("%s.%s.%s", "file", uuid0, "doc") + "\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var sb strings.Builder
			writeFileTree(&sb, tt.files, tt.prefix)
			result := sb.String()
			if result != tt.expected {
				t.Errorf("writeFileTree() = %q, want %q", result, tt.expected)
			}
		})
	}
}
