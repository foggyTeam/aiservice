package providers

import (
	"context"
	"fmt"

	"github.com/aiservice/internal/models"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
)

//go:generate mockgen -source=$GOFILE -destination=./mocks/mock_$GOFILE -package=mocks
type LLMClient interface {
	ImageRecognition(ctx context.Context, parts []*ai.Part) (ImageRecognitionFlow, error)
	Summarize(ctx context.Context, parts []*ai.Part) (SummarizeFlow, error)
	SummarizeWithHistory(ctx context.Context, history []*ai.Message, parts []*ai.Part) (SummarizeFlow, error)
	Structurize(ctx context.Context, parts []*ai.Part) (StructurizeFlow, error)
	StructurizeWithHistory(ctx context.Context, history []*ai.Message, parts []*ai.Part) (StructurizeFlow, error)
	GenerateTemplate(ctx context.Context, parts []*ai.Part) (TemplateGenerationFlow, error)
	GenerateText(ctx context.Context, parts []*ai.Part) (string, error)
	GetName() string // Added for provider identification
}

// DigitalInkRecognizer interface for handwriting recognition
type DigitalInkRecognizer interface {
	RecognizeInk(ctx context.Context, elements []models.Element) (string, error)
}

// ImageRecognitionFlow represents the output structure for image recognition
type ImageRecognitionFlow struct {
	ImageDescription string `json:"imageDescription" jsonschema:"description=Описание содержимого изображения"`
}

// SummarizeFlow represents the output structure for summarization
type SummarizeFlow struct {
	Summarization string `json:"summarization" jsonschema:"description=Текст суммаризации содержимого доски"`
}

type TextGenerationFlow struct {
	Content string `json:"content" jsonschema:"description=Сгенерированный текст, обернутый в HTML теги"`
}

// TemplateGenerationFlow represents the output structure for template generation
type TemplateGenerationFlow struct {
	BoardType   string            `json:"boardType" jsonschema:"description=Тип сгенерированной доски: simple или graph"`
	Title       string            `json:"title" jsonschema:"description=Заголовок доски"`
	Description string            `json:"description" jsonschema:"description=Описание содержимого доски"`
	Elements    []TemplateElement `json:"elements,omitempty" jsonschema:"description=Элементы доски для simple board"`
	GraphNodes  []TemplateNode    `json:"graphNodes,omitempty" jsonschema:"description=Узлы графа для graph board"`
	GraphEdges  []TemplateEdge    `json:"graphEdges,omitempty" jsonschema:"description=Рёбра графа для graph board"`
}

// TemplateElement represents a board element for simple board
type TemplateElement struct {
	Type         string  `json:"type,omitempty" jsonschema:"description=Тип элемента: rectangle, text, ellipse, line"`
	X            float32 `json:"x,omitempty" jsonschema:"description=Координата X левого верхнего угла"`
	Y            float32 `json:"y,omitempty" jsonschema:"description=Координата Y левого верхнего угла"`
	Width        float32 `json:"width,omitempty" jsonschema:"description=Ширина элемента"`
	Height       float32 `json:"height,omitempty" jsonschema:"description=Высота элемента"`
	Rotation     float32 `json:"rotation,omitempty" jsonschema:"description=Угол поворота в градусах"`
	Fill         string  `json:"fill,omitempty" jsonschema:"description=Цвет заливки в формате hex"`
	Stroke       string  `json:"stroke,omitempty" jsonschema:"description=Цвет обводки в формате hex"`
	StrokeWidth  int     `json:"strokeWidth,omitempty" jsonschema:"description=Толщина обводки в пикселях"`
	Content      string  `json:"content,omitempty" jsonschema:"description=Текстовое содержимое элемента"`
	CornerRadius int     `json:"cornerRadius,omitempty" jsonschema:"description=Радиус скругления углов"`
	// Additional text styling fields (optional, may be returned by LLM)
	FontSize   int    `json:"fontSize,omitempty" jsonschema:"description=Размер шрифта"`
	FontColor  string `json:"fontColor,omitempty" jsonschema:"description=Цвет текста в формате hex"`
	FontFamily string `json:"fontFamily,omitempty" jsonschema:"description=Семейство шрифтов"`
}

// TemplateNode represents a graph node for graph board (simplified - content only)
type TemplateNode struct {
	ID          string `json:"id" jsonschema:"description=Уникальный идентификатор узла"`
	Type        string `json:"type" jsonschema:"description=Тип узла: customNode, externalLinkNode, internalLinkNode"`
	Title       string `json:"title" jsonschema:"description=Заголовок узла"`
	Description string `json:"description,omitempty" jsonschema:"description=Описание узла"`
	URL         string `json:"url,omitempty" jsonschema:"description=URL для внешних ссылок"`
}

// TemplateEdge represents a graph edge for graph board (simplified - connections only)
type TemplateEdge struct {
	ID     string `json:"id" jsonschema:"description=Уникальный идентификатор ребра"`
	Source string `json:"source" jsonschema:"description=ID исходного узла"`
	Target string `json:"target" jsonschema:"description=ID целевого узла"`
	Label  string `json:"label,omitempty" jsonschema:"description=Метка/описание связи"`
}

// For structurize, we'll define a flow that doesn't use the recursive File structure in its definition
// to avoid schema generation issues, but the LLM will be instructed to return the proper structure

// FileNode represents a single node in the file hierarchy for schema generation
// This avoids infinite recursion during JSON schema generation
type FileNode struct {
	ID       string  `json:"id"`
	Name     string  `json:"name" example:"main.go"`
	Type     string  `json:"type" example:"doc"` //doc, simple, graph,(поле children пустое) | section (содердит детей)
	ParentID *string `json:"parentId,omitempty"` // Points to parent node ID, nil for root
}

// FileHierarchy represents the complete file structure as a flat list with parent-child relationships
type FileHierarchy struct {
	Nodes   []FileNode `json:"nodes"`
	RootIDs []string   `json:"rootIds"` // IDs of root-level nodes
}

// ToModelFile converts the flat FileHierarchy to the original recursive File model
func (fh FileHierarchy) ToModelFile() models.File {
	if len(fh.RootIDs) == 0 {
		return models.File{}
	}
	childrenByParent := map[string][]FileNode{}
	nodeByID := map[string]FileNode{}
	for _, n := range fh.Nodes {
		nodeByID[n.ID] = n
		if n.ParentID != nil {
			childrenByParent[*n.ParentID] =
				append(childrenByParent[*n.ParentID], n)
		}
	}
	var build func(id string) models.File
	build = func(id string) models.File {
		n := nodeByID[id]
		file := models.File{
			Name: n.Name,
			Type: n.Type,
		}
		for _, child := range childrenByParent[id] {
			file.Children =
				append(file.Children, build(child.ID))
		}
		return file
	}
	return build(fh.RootIDs[0])
}

type SimpleStructurizeFlow struct {
	Prompt         string        `json:"userPrompt"`
	Answer         string        `json:"answer"`
	AiTreeResponse string        `json:"aiTreeResponse"`
	File           FileHierarchy `json:"children"`
}

func RunStructurizeGeneration(ctx context.Context, gkit *genkit.Genkit, parts []*ai.Part) (*SimpleStructurizeFlow, error) {
	prompt := ai.NewUserMessage(parts...)
	resp, err := genkit.Generate(ctx, gkit,
		ai.WithMessages(prompt),
		ai.WithOutputType(&SimpleStructurizeFlow{}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate llm request: %w", err)
	}

	var flow SimpleStructurizeFlow
	if err := resp.Output(&flow); err != nil {
		return nil, fmt.Errorf("failed to parse output: %w", err)
	}
	return &flow, nil
}

// RunStructurizeGenerationAndConvert executes the structurize generation and converts the result to the original File model
func RunStructurizeGenerationAndConvert(ctx context.Context, gkit *genkit.Genkit, parts []*ai.Part) (models.File, string, error) {
	resp, err := RunStructurizeGeneration(ctx, gkit, parts)
	if err != nil {
		return models.File{}, "", fmt.Errorf("failed to generate llm request: %w", err)
	}

	// Convert the flat hierarchy to the original recursive File model
	modelFile := resp.File.ToModelFile()

	return modelFile, resp.AiTreeResponse, nil
}

// StructurizeFlow represents the output structure for structurization (used in pipeline)
type StructurizeFlow struct {
	AiTreeResponse string `json:"aiTreeResponse" jsonschema:"description=ASCII tree representation of file structure"`
}
