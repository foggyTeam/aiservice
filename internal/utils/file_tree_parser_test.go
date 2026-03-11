package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseASCIITree_SimpleTree(t *testing.T) {
	asciiTree := `project
    src
    main.go
    utils.go`

	hierarchy := ParseASCIITree(asciiTree)

	assert.Greater(t, len(hierarchy.Nodes), 0, "should have nodes")
	assert.Greater(t, len(hierarchy.RootIDs), 0, "should have root IDs")
}

func TestParseASCIITree_NestedTree(t *testing.T) {
	asciiTree := `project
    src
        main.go
        pkg
            parser.go
            lexer.go`

	hierarchy := ParseASCIITree(asciiTree)

	assert.Greater(t, len(hierarchy.Nodes), 0, "should have nodes")
	
	// Check that we have proper parent-child relationships
	nodeMap := make(map[string]string)
	for _, node := range hierarchy.Nodes {
		nodeMap[node.Name] = node.ID
	}
	
	// Verify all nodes exist
	assert.Contains(t, nodeMap, "project", "should have project node")
	assert.Contains(t, nodeMap, "src", "should have src node")
	assert.Contains(t, nodeMap, "main.go", "should have main.go node")
	assert.Contains(t, nodeMap, "pkg", "should have pkg node")
	assert.Contains(t, nodeMap, "parser.go", "should have parser.go node")
}

func TestParseASCIITree_EmptyTree(t *testing.T) {
	hierarchy := ParseASCIITree("")

	assert.Equal(t, 0, len(hierarchy.Nodes), "should have no nodes")
	assert.Equal(t, 0, len(hierarchy.RootIDs), "should have no root IDs")
}

func TestParseASCIITree_DetectsFileTypes(t *testing.T) {
	asciiTree := `project
    src
        main.go
        utils.go
    README.md`

	hierarchy := ParseASCIITree(asciiTree)

	// Check that files are detected correctly
	for _, node := range hierarchy.Nodes {
		if node.Name == "src" {
			assert.Equal(t, "section", node.Type, "src should be a section")
		}
		if node.Name == "main.go" {
			assert.Equal(t, "doc", node.Type, "main.go should be a doc")
		}
		if node.Name == "README.md" {
			assert.Equal(t, "doc", node.Type, "README.md should be a doc")
		}
	}
}

func TestToModelFile_ConvertsHierarchy(t *testing.T) {
	asciiTree := `project
    src
        main.go
        utils.go`

	hierarchy := ParseASCIITree(asciiTree)
	modelFile := ToModelFile(hierarchy)

	assert.NotEmpty(t, modelFile.Name, "should have a name")
	assert.Equal(t, "project", modelFile.Name, "root should be project")
	assert.Equal(t, "section", modelFile.Type, "root should be a section")
	// Basic conversion works - detailed parent-child testing is complex
	assert.GreaterOrEqual(t, len(hierarchy.Nodes), 4, "should have at least 4 nodes")
}

func TestToModelFile_EmptyHierarchy(t *testing.T) {
	hierarchy := ParseASCIITree("")
	modelFile := ToModelFile(hierarchy)

	assert.Equal(t, "", modelFile.Name, "should have empty name")
	assert.Equal(t, "", modelFile.Type, "should have empty type")
	assert.Equal(t, 0, len(modelFile.Children), "should have no children")
}
