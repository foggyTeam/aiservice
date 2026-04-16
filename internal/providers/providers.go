package providers

import (
	"context"
	"fmt"

	"github.com/aiservice/internal/models"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
)

type LLMClient interface {
	ImageRecognition(ctx context.Context, parts []*ai.Part) (ImageRecognitionFlow, error)
	Summarize(ctx context.Context, parts []*ai.Part) (SummarizeFlow, error)
	SummarizeWithHistory(ctx context.Context, history []*ai.Message, parts []*ai.Part) (SummarizeFlow, error)
	Structurize(ctx context.Context, parts []*ai.Part) (StructurizeFlow, error)
	GenerateTemplate(ctx context.Context, parts []*ai.Part) (TemplateGenerationFlow, error)
	GetName() string // Added for provider identification
}

// ImageRecognitionFlow represents the output structure for image recognition
type ImageRecognitionFlow struct {
	ImageDescription string `json:"imageDescription" jsonschema:"description=Описание содержимого изображения"`
}

// SummarizeFlow represents the output structure for summarization
type SummarizeFlow struct {
	Summarization string `json:"summarization" jsonschema:"description=Текст суммаризации содержимого доски"`
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
	if len(fh.Nodes) == 0 {
		return models.File{}
	}

	// Create a map of all nodes by ID for quick lookup
	nodeMap := make(map[string]FileNode)
	for _, node := range fh.Nodes {
		nodeMap[node.ID] = node
	}

	// Create a map to store the model files we create
	modelMap := make(map[string]models.File)

	// First pass: create all model files without children
	for id, node := range nodeMap {
		modelMap[id] = models.File{
			Name: node.Name,
			Type: node.Type,
		}
	}

	// Second pass: assign children to each parent
	childrenMap := make(map[string][]models.File)
	for id, node := range nodeMap {
		if node.ParentID != nil {
			// This node has a parent, add it as a child to the parent
			parentID := *node.ParentID
			childrenMap[parentID] = append(childrenMap[parentID], modelMap[id])
		}
	}

	// Assign children to parents
	for parentID, children := range childrenMap {
		if parentFile, exists := modelMap[parentID]; exists {
			parentFile.Children = children
			modelMap[parentID] = parentFile
		}
	}

	// Find the root node (the first one from rootIds that exists)
	var rootNode models.File
	if len(fh.RootIDs) > 0 {
		if rootFile, exists := modelMap[fh.RootIDs[0]]; exists {
			rootNode = rootFile
		}
	}

	return rootNode
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
