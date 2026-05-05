package providers

import (
	"testing"

	"github.com/aiservice/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func ptrString(s string) *string {
	return &s
}

func findChildByName(children []models.File, name string) (*models.File, bool) {
	for i := range children {
		if children[i].Name == name {
			return &children[i], true
		}
	}
	return nil, false
}

func expectFile(t *testing.T, got, want models.File) {
	t.Helper()

	assert.Equal(t, want.Name, got.Name, "Name mismatch for %s", want.Name)
	assert.Equal(t, want.Type, got.Type, "Type mismatch for %s", want.Name)
	assert.Len(t, got.Children, len(want.Children), "Children count mismatch for %s", want.Name)

	for _, expectedChild := range want.Children {
		actualChild, found := findChildByName(got.Children, expectedChild.Name)
		assert.True(t, found, "Child %q not found in %q", expectedChild.Name, got.Name)
		if found {
			expectFile(t, *actualChild, expectedChild)
		}
	}
}

func TestToModelFile_EmptyInput(t *testing.T) {
	t.Parallel()

	fh := FileHierarchy{
		Nodes:   []FileNode{},
		RootIDs: []string{},
	}

	result := fh.ToModelFile()

	assert.True(t, result.IsEmpty(), "пустой вход должен вернуть пустой файл")
}

func TestToModelFile_SingleRoot_NoChildren(t *testing.T) {
	t.Parallel()

	fh := FileHierarchy{
		Nodes: []FileNode{
			{ID: "root", Name: "project", Type: "section", ParentID: nil},
		},
		RootIDs: []string{"root"},
	}

	result := fh.ToModelFile()

	expectFile(t, result, models.File{
		Name:     "project",
		Type:     "section",
		Children: []models.File{},
	})
}

func TestToModelFile_RootWithOneChild(t *testing.T) {
	t.Parallel()

	fh := FileHierarchy{
		Nodes: []FileNode{
			{ID: "root", Name: "src", Type: "section", ParentID: nil},
			{ID: "child1", Name: "main.go", Type: "doc", ParentID: ptrString("root")},
		},
		RootIDs: []string{"root"},
	}

	result := fh.ToModelFile()

	expectFile(t, result, models.File{
		Name: "src",
		Type: "section",
		Children: []models.File{
			{Name: "main.go", Type: "doc", Children: []models.File{}},
		},
	})
}

func TestToModelFile_RootWithMultipleChildren(t *testing.T) {
	t.Parallel()

	fh := FileHierarchy{
		Nodes: []FileNode{
			{ID: "root", Name: "project", Type: "section", ParentID: nil},
			{ID: "f1", Name: "README.md", Type: "doc", ParentID: ptrString("root")},
			{ID: "f2", Name: "go.mod", Type: "doc", ParentID: ptrString("root")},
			{ID: "f3", Name: "main.go", Type: "doc", ParentID: ptrString("root")},
		},
		RootIDs: []string{"root"},
	}

	result := fh.ToModelFile()

	assert.Equal(t, "project", result.Name)
	assert.Equal(t, "section", result.Type)
	assert.Len(t, result.Children, 3)

	names := make(map[string]bool)
	for _, child := range result.Children {
		names[child.Name] = true
		assert.Equal(t, "doc", child.Type)
		assert.Empty(t, child.Children)
	}
	assert.True(t, names["README.md"])
	assert.True(t, names["go.mod"])
	assert.True(t, names["main.go"])
}

func TestToModelFile_MultiLevelHierarchy(t *testing.T) {
	t.Parallel()

	fh := FileHierarchy{
		Nodes: []FileNode{
			{ID: "root", Name: "myapp", Type: "section", ParentID: nil},
			{ID: "cmd", Name: "cmd", Type: "section", ParentID: ptrString("root")},
			{ID: "internal", Name: "internal", Type: "section", ParentID: ptrString("root")},
			{ID: "main", Name: "main.go", Type: "doc", ParentID: ptrString("cmd")},
			{ID: "util", Name: "util.go", Type: "doc", ParentID: ptrString("internal")},
		},
		RootIDs: []string{"root"},
	}

	result := fh.ToModelFile()

	assert.Equal(t, "myapp", result.Name)
	assert.Len(t, result.Children, 2)

	var cmd, internal models.File
	for _, child := range result.Children {
		switch child.Name {
		case "cmd":
			cmd = child
		case "internal":
			internal = child
		}
	}

	assert.Equal(t, "cmd", cmd.Name)
	require.Len(t, cmd.Children, 1)
	assert.Equal(t, "main.go", cmd.Children[0].Name)
	assert.Equal(t, "doc", cmd.Children[0].Type)

	assert.Equal(t, "internal", internal.Name)
	require.Len(t, internal.Children, 1)
	assert.Equal(t, "util.go", internal.Children[0].Name)
}

func TestToModelFile_MultipleRootIDs_UsesFirst(t *testing.T) {
	t.Parallel()

	fh := FileHierarchy{
		Nodes: []FileNode{
			{ID: "root1", Name: "First", Type: "section", ParentID: nil},
			{ID: "root2", Name: "Second", Type: "section", ParentID: nil},
			{ID: "child", Name: "file.txt", Type: "doc", ParentID: ptrString("root2")},
		},
		RootIDs: []string{"root1", "root2"},
	}

	result := fh.ToModelFile()

	assert.Equal(t, "First", result.Name)
	assert.Empty(t, result.Children)
}

func TestToModelFile_RootID_NotFound(t *testing.T) {
	t.Parallel()

	fh := FileHierarchy{
		Nodes: []FileNode{
			{ID: "actual", Name: "RealRoot", Type: "section", ParentID: nil},
		},
		RootIDs: []string{"nonexistent"},
	}

	result := fh.ToModelFile()

	assert.True(t, result.IsEmpty())
}

func TestToModelFile_OrphanNode_ParentID_NotFound(t *testing.T) {
	t.Parallel()

	fh := FileHierarchy{
		Nodes: []FileNode{
			{ID: "root", Name: "root", Type: "section", ParentID: nil},
			{ID: "orphan", Name: "lost.txt", Type: "doc", ParentID: ptrString("missing-parent")},
		},
		RootIDs: []string{"root"},
	}

	result := fh.ToModelFile()

	assert.Equal(t, "root", result.Name)
	assert.Empty(t, result.Children)
}

func TestToModelFile_MixedTypes(t *testing.T) {
	t.Parallel()

	fh := FileHierarchy{
		Nodes: []FileNode{
			{ID: "r", Name: "repo", Type: "section", ParentID: nil},
			{ID: "f1", Name: "doc.pdf", Type: "doc", ParentID: ptrString("r")},
			{ID: "f2", Name: "board.simple", Type: "simple", ParentID: ptrString("r")},
			{ID: "f3", Name: "graph.board", Type: "graph", ParentID: ptrString("r")},
		},
		RootIDs: []string{"r"},
	}

	result := fh.ToModelFile()

	assert.Equal(t, "repo", result.Name)
	assert.Len(t, result.Children, 3)

	types := make(map[string]string)
	for _, child := range result.Children {
		types[child.Name] = child.Type
	}
	assert.Equal(t, "doc", types["doc.pdf"])
	assert.Equal(t, "simple", types["board.simple"])
	assert.Equal(t, "graph", types["graph.board"])
}

func TestToModelFile_EmptyRootIDs(t *testing.T) {
	t.Parallel()

	fh := FileHierarchy{
		Nodes: []FileNode{
			{ID: "orphan", Name: "no-root.txt", Type: "doc", ParentID: nil},
		},
		RootIDs: []string{},
	}

	result := fh.ToModelFile()

	assert.True(t, result.IsEmpty())
}

func TestToModelFile_Node_WithSpecialChars(t *testing.T) {
	t.Parallel()

	fh := FileHierarchy{
		Nodes: []FileNode{
			{ID: "r", Name: "my-project_2024", Type: "section", ParentID: nil},
			{ID: "f", Name: "file with spaces & (parentheses).go", Type: "doc", ParentID: ptrString("r")},
		},
		RootIDs: []string{"r"},
	}

	result := fh.ToModelFile()

	assert.Equal(t, "my-project_2024", result.Name)
	require.Len(t, result.Children, 1)
	assert.Equal(t, "file with spaces & (parentheses).go", result.Children[0].Name)
}

func TestToModelFile_Deterministic(t *testing.T) {
	t.Parallel()

	fh := FileHierarchy{
		Nodes: []FileNode{
			{ID: "a", Name: "A", Type: "section", ParentID: nil},
			{ID: "b", Name: "B", Type: "doc", ParentID: ptrString("a")},
			{ID: "c", Name: "C", Type: "doc", ParentID: ptrString("a")},
		},
		RootIDs: []string{"a"},
	}

	var results []models.File
	for range 10 {
		results = append(results, fh.ToModelFile())
	}

	first := results[0]
	for i, res := range results[1:] {
		expectFile(t, res, first)
		_ = i
	}
}
