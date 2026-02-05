package preprocessing

import (
	"testing"

	"github.com/aiservice/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestPreprocessor(t *testing.T) {
	preprocessor := NewPreprocessor()

	t.Run("PreprocessSummarizeRequest", func(t *testing.T) {
		req := models.SummarizeRequest{
			Board: models.Board{
				BoardID: "test-board",
				Elements: []models.Element{
					{
						Id:      "elem1",
						Type:    "text",
						X:       10.0,
						Y:       20.0,
						Width:   100.0,
						Height:  50.0,
						Content: "Project Goals",
					},
					{
						Id:      "elem2",
						Type:    "text",
						X:       150.0,
						Y:       20.0,
						Width:   100.0,
						Height:  50.0,
						Content: "Increase revenue by 20%",
					},
				},
			},
		}

		parts, err := preprocessor.PreprocessSummarizeRequest(req)
		assert.NoError(t, err)
		assert.NotEmpty(t, parts)
		assert.Contains(t, parts[0].Text, "Project Goals")
		assert.Contains(t, parts[0].Text, "SPATIAL ANALYSIS")
		assert.Contains(t, parts[0].Text, "SEMANTIC ANNOTATIONS")
	})

	t.Run("PreprocessStructurizeRequest", func(t *testing.T) {
		req := models.StructurizeRequest{
			Board: models.Board{
				BoardID: "test-board",
				Elements: []models.Element{
					{
						Id:      "elem1",
						Type:    "text",
						X:       10.0,
						Y:       20.0,
						Width:   100.0,
						Height:  50.0,
						Content: "Project Structure",
					},
				},
			},
			File: models.File{
				Name: "project-root",
				Type: "section",
				Children: []*models.File{
					{
						Name: "src",
						Type: "section",
					},
					{
						Name: "README.md",
						Type: "doc",
					},
				},
			},
		}

		parts, err := preprocessor.PreprocessStructurizeRequest(req)
		assert.NoError(t, err)
		assert.NotEmpty(t, parts)
		assert.Contains(t, parts[0].Text, "project-root")
		assert.Contains(t, parts[0].Text, "src")
		assert.Contains(t, parts[0].Text, "README.md")
		assert.Contains(t, parts[0].Text, "SPATIAL ANALYSIS")
		assert.Contains(t, parts[0].Text, "SEMANTIC ANNOTATIONS")
	})

	t.Run("analyzeSpatialRelationships", func(t *testing.T) {
		elements := []models.Element{
			{
				Id:      "elem1",
				Type:    "text",
				X:       10.0,
				Y:       20.0,
				Width:   100.0,
				Height:  50.0,
				Content: "Test content",
			},
			{
				Id:      "elem2",
				Type:    "text",
				X:       150.0,
				Y:       20.0,
				Width:   100.0,
				Height:  50.0,
				Content: "Nearby content",
			},
		}

		clusters := preprocessor.analyzeSpatialRelationships(elements)
		assert.NotEmpty(t, clusters)
	})

	t.Run("annotateSemantics", func(t *testing.T) {
		elements := []models.Element{
			{
				Id:      "elem1",
				Type:    "text",
				X:       10.0,
				Y:       20.0,
				Width:   100.0,
				Height:  50.0,
				Content: "Key insight",
			},
			{
				Id:    "elem2",
				Type:  "rectangle",
				X:     150.0,
				Y:     20.0,
				Width: 200.0,
				Height: 100.0,
			},
		}

		annotations := preprocessor.annotateSemantics(elements)
		assert.Contains(t, annotations, "Element 1")
		assert.Contains(t, annotations, "Key insight")
		assert.Contains(t, annotations, "Inferred Role")
	})

	t.Run("PreprocessForSimilarity", func(t *testing.T) {
		req := models.AnalyzeRequest{
			RequestType: models.SummarizeType,
			SummarizeRequest: models.SummarizeRequest{
				Board: models.Board{
					BoardID: "test-board",
					Elements: []models.Element{
						{Id: "elem1", Type: "text", Content: "content"},
					},
				},
			},
		}

		similarityKey := preprocessor.PreprocessForSimilarity(req)
		assert.Equal(t, "summarize:test-board:1", similarityKey)
	})
}