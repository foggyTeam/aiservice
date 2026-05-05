package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSumAnalyzeReq_BasicCreation(t *testing.T) {
	req := SummarizeRequest{
		RequestID:   "req-123",
		UserID:      "user-456",
		RequestType: SummarizeType,
		Board: Board{
			BoardID: "board-789",
		},
	}

	result := NewSumAnalyzeReq(req)

	assert.Equal(t, SummarizeType, result.RequestType, "RequestType should be SummarizeType")
	assert.Equal(t, req, result.SummarizeRequest, "SummarizeRequest should be set correctly")
	assert.Empty(t, result.StructurizeRequest, "StructurizeRequest should be empty")
	assert.Empty(t, result.TemplateRequest, "TemplateRequest should be empty")
}

func TestNewSumAnalyzeReq_WithComplexBoard(t *testing.T) {
	board := Board{
		BoardID:  "board-1",
		ImageURL: "https://example.com/image.png",
		Elements: []Element{
			{
				Id:   "elem-1",
				Type: RectangeType,
				X:    10,
				Y:    20,
			},
			{
				Id:      "elem-2",
				Type:    TextType,
				Content: "Sample text",
			},
		},
	}

	req := SummarizeRequest{
		RequestID:   "req-001",
		UserID:      "user-001",
		RequestType: SummarizeType,
		Board:       board,
	}

	result := NewSumAnalyzeReq(req)

	assert.Equal(t, SummarizeType, result.RequestType)
	assert.Equal(t, board, result.SummarizeRequest.Board)
	assert.Equal(t, 2, len(result.SummarizeRequest.Board.Elements))
}

func TestNewSumAnalyzeReq_EmptySummarizeRequest(t *testing.T) {
	req := SummarizeRequest{}

	result := NewSumAnalyzeReq(req)

	assert.Equal(t, SummarizeType, result.RequestType)
	assert.Equal(t, req, result.SummarizeRequest)
}

func TestNewStructAnalyzeReq_BasicCreation(t *testing.T) {
	req := StructurizeRequest{
		RequestID:   "req-123",
		UserID:      "user-456",
		RequestType: StructurizeType,
		Board: Board{
			BoardID: "board-789",
		},
		File: File{
			Name: "root",
			Type: "section",
		},
	}

	result := NewStructAnalyzeReq(req)

	assert.Equal(t, StructurizeType, result.RequestType, "RequestType should be StructurizeType")
	assert.Equal(t, req, result.StructurizeRequest, "StructurizeRequest should be set correctly")
	assert.Empty(t, result.SummarizeRequest, "SummarizeRequest should be empty")
	assert.Empty(t, result.TemplateRequest, "TemplateRequest should be empty")
}

func TestNewStructAnalyzeReq_WithFileStructure(t *testing.T) {
	file := File{
		Name: "project",
		Type: "section",
		Children: []File{
			{
				Name: "src",
				Type: "section",
				Children: []File{
					{
						Name: "main.go",
						Type: "doc",
					},
				},
			},
			{
				Name: "README.md",
				Type: "doc",
			},
		},
	}

	req := StructurizeRequest{
		RequestID:   "req-struct-1",
		UserID:      "user-struct-1",
		RequestType: StructurizeType,
		File:        file,
	}

	result := NewStructAnalyzeReq(req)

	assert.Equal(t, StructurizeType, result.RequestType)
	assert.Equal(t, file, result.StructurizeRequest.File)
	assert.Equal(t, 2, len(result.StructurizeRequest.File.Children))
}

func TestNewTemplateAnalyzeReq_BasicCreation(t *testing.T) {
	req := GenerateTemplateRequest{
		RequestID:   "req-template-1",
		UserID:      "user-template-1",
		RequestType: GenerateTemplateType,
		BoardID:     "board-template-1",
		Prompt:      "Create a simple project structure diagram",
		BoardType:   BoardTypeSimple,
	}

	result := NewTemplateAnalyzeReq(req)

	assert.Equal(t, GenerateTemplateType, result.RequestType, "RequestType should be GenerateTemplateType")
	assert.Equal(t, req, result.TemplateRequest, "TemplateRequest should be set correctly")
	assert.Empty(t, result.SummarizeRequest, "SummarizeRequest should be empty")
	assert.Empty(t, result.StructurizeRequest, "StructurizeRequest should be empty")
}

func TestNewTemplateAnalyzeReq_WithGraphBoardType(t *testing.T) {
	req := GenerateTemplateRequest{
		RequestID:   "req-graph-1",
		UserID:      "user-graph-1",
		RequestType: GenerateTemplateType,
		BoardID:     "board-graph-1",
		Prompt:      "Create a flowchart diagram showing data flow",
		BoardType:   BoardTypeGraph,
	}

	result := NewTemplateAnalyzeReq(req)

	assert.Equal(t, GenerateTemplateType, result.RequestType)
	assert.Equal(t, BoardTypeGraph, result.TemplateRequest.BoardType)
}

func TestNewTemplateAnalyzeReq_WithGenerateTextType(t *testing.T) {
	req := GenerateTemplateRequest{
		RequestID:   "req-text-1",
		UserID:      "user-text-1",
		RequestType: GenerateTextType,
		BoardID:     "board-text-1",
		Prompt:      "Generate a detailed description",
		BoardType:   BoardTypeSimple,
	}

	result := NewTemplateAnalyzeReq(req)

	assert.Equal(t, GenerateTextType, result.RequestType, "RequestType should be GenerateTextType")
	assert.Equal(t, req, result.TemplateRequest)
}

func TestFile_IsEmpty_WithBothFieldsEmpty(t *testing.T) {
	file := File{
		Name: "",
		Type: "",
	}

	assert.True(t, file.IsEmpty(), "File with empty Name and Type should be empty")
}

func TestFile_IsEmpty_WithNameEmpty(t *testing.T) {
	file := File{
		Name: "",
		Type: "doc",
	}

	assert.False(t, file.IsEmpty(), "File with non-empty Type should not be empty")
}

func TestFile_IsEmpty_WithTypeEmpty(t *testing.T) {
	file := File{
		Name: "file.txt",
		Type: "",
	}

	assert.False(t, file.IsEmpty(), "File with non-empty Name should not be empty")
}

func TestFile_IsEmpty_WithBothFieldsPopulated(t *testing.T) {
	file := File{
		Name: "readme.md",
		Type: "doc",
	}

	assert.False(t, file.IsEmpty(), "File with Name and Type should not be empty")
}

func TestFile_IsEmpty_IgnoresChildren(t *testing.T) {
	file := File{
		Name: "",
		Type: "",
		Children: []File{
			{Name: "child", Type: "doc"},
		},
	}

	assert.True(t, file.IsEmpty(), "IsEmpty should ignore Children field")
}

func TestFile_IsPopulated_WithBothFieldsEmpty(t *testing.T) {
	file := File{
		Name: "",
		Type: "",
	}

	assert.False(t, file.IsPopulated(), "File with empty Name and Type should not be populated")
}

func TestFile_IsPopulated_WithNamePopulated(t *testing.T) {
	file := File{
		Name: "file.txt",
		Type: "",
	}

	assert.True(t, file.IsPopulated(), "File with non-empty Name should be populated")
}

func TestFile_IsPopulated_WithTypePopulated(t *testing.T) {
	file := File{
		Name: "",
		Type: "section",
	}

	assert.True(t, file.IsPopulated(), "File with non-empty Type should be populated")
}

func TestFile_IsPopulated_WithBothFieldsPopulated(t *testing.T) {
	file := File{
		Name: "project",
		Type: "section",
	}

	assert.True(t, file.IsPopulated(), "File with Name and Type should be populated")
}

func TestFile_IsPopulated_WithChildrenOnly(t *testing.T) {
	file := File{
		Name: "",
		Type: "",
		Children: []File{
			{Name: "child1", Type: "doc"},
			{Name: "child2", Type: "section"},
		},
	}

	assert.True(t, file.IsPopulated(), "File with Children should be populated even if Name and Type are empty")
}

func TestFile_IsPopulated_WithAllFieldsPopulated(t *testing.T) {
	file := File{
		Name: "root",
		Type: "section",
		Children: []File{
			{Name: "src", Type: "section"},
			{Name: "main.go", Type: "doc"},
		},
	}

	assert.True(t, file.IsPopulated(), "File with all fields populated should be populated")
}

func TestFile_IsPopulated_WithEmptyChildren(t *testing.T) {
	file := File{
		Name:     "project",
		Type:     "section",
		Children: []File{},
	}

	assert.True(t, file.IsPopulated(), "File with Name and Type and empty Children should be populated")
}

func TestIsEmptyAndIsPopulated_Consistency(t *testing.T) {
	emptyFile := File{Name: "", Type: ""}
	populatedFile := File{Name: "file", Type: "doc"}

	assert.True(t, emptyFile.IsEmpty())
	assert.False(t, emptyFile.IsPopulated())

	assert.False(t, populatedFile.IsEmpty())
	assert.True(t, populatedFile.IsPopulated())
}

func TestNewAnalyzeReqs_ConsistentBehavior(t *testing.T) {
	sumReq := NewSumAnalyzeReq(SummarizeRequest{RequestID: "1"})
	assert.Equal(t, SummarizeType, sumReq.RequestType)
	assert.NotEmpty(t, sumReq.SummarizeRequest)
	assert.Empty(t, sumReq.StructurizeRequest)

	structReq := NewStructAnalyzeReq(StructurizeRequest{RequestID: "2"})
	assert.Equal(t, StructurizeType, structReq.RequestType)
	assert.NotEmpty(t, structReq.StructurizeRequest)
	assert.Empty(t, structReq.SummarizeRequest)
}
