package preprocessing

import (
	"strings"
	"testing"

	"github.com/aiservice/internal/models"
)

func TestCreateFileHierarchyDescription(t *testing.T) {
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
				Name: "test.txt",
				Type: "doc",
			},
			expected: "test.txt\n",
		},
		{
			name: "file with children",
			file: models.File{
				Name: "root",
				Type: "section",
				Children: []models.File{
					{Name: "child1.txt", Type: "doc"},
					{Name: "child2", Type: "section", Children: []models.File{
						{Name: "grandchild.txt", Type: "doc"},
					}},
				},
			},
			expected: "root\n├── child1.txt\n└── child2\n    └── grandchild.txt\n",
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
				{Name: "file.txt", Type: "doc"},
			},
			prefix:   "",
			expected: "└── file.txt\n",
		},
		{
			name: "multiple files",
			files: []models.File{
				{Name: "file1.txt", Type: "doc"},
				{Name: "file2.txt", Type: "doc"},
			},
			prefix:   "",
			expected: "├── file1.txt\n└── file2.txt\n",
		},
		{
			name: "nested files",
			files: []models.File{
				{Name: "dir", Type: "section", Children: []models.File{
					{Name: "nested.txt", Type: "doc"},
				}},
			},
			prefix:   "",
			expected: "└── dir\n    └── nested.txt\n",
		},
		{
			name: "with prefix",
			files: []models.File{
				{Name: "file.txt", Type: "doc"},
			},
			prefix:   "│   ",
			expected: "│   └── file.txt\n",
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
