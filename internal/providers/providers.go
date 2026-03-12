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
	Structurize(ctx context.Context, parts []*ai.Part) (StructurizeFlow, error)
	GenerateTemplate(ctx context.Context, parts []*ai.Part) (TemplateGenerationFlow, error)
	GetName() string // Added for provider identification
}

// ImageRecognitionFlow represents the output structure for image recognition
type ImageRecognitionFlow struct {
	ImageDescription string `json:"imageDescription"`
}

// SummarizeFlow represents the output structure for summarization
type SummarizeFlow struct {
	Summarization string `json:"summarization"`
}

// TemplateGenerationFlow represents the output structure for template generation
type TemplateGenerationFlow struct {
	BoardType    string            `json:"boardType"`
	Title        string            `json:"title"`
	Description  string            `json:"description"`
	Elements     []TemplateElement `json:"elements,omitempty"`
	GraphNodes   []TemplateNode    `json:"graphNodes,omitempty"`
	GraphEdges   []TemplateEdge    `json:"graphEdges,omitempty"`
}

// TemplateElement represents a board element for simple board
type TemplateElement struct {
	Type         string  `json:"type"`
	X            float32 `json:"x"`
	Y            float32 `json:"y"`
	Width        float32 `json:"width"`
	Height       float32 `json:"height"`
	Rotation     float32 `json:"rotation"`
	Fill         string  `json:"fill,omitempty"`
	Stroke       string  `json:"stroke,omitempty"`
	StrokeWidth  int     `json:"strokeWidth,omitempty"`
	Content      string  `json:"content,omitempty"`
	CornerRadius int     `json:"cornerRadius,omitempty"`
}

// TemplateNode represents a graph node for graph board
type TemplateNode struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	PositionX   float32 `json:"positionX"`
	PositionY   float32 `json:"positionY"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Shape       string `json:"shape"`
	Color       string `json:"color,omitempty"`
	URL         string `json:"url,omitempty"`
}

// TemplateEdge represents a graph edge for graph board
type TemplateEdge struct {
	ID     string `json:"id"`
	Source string `json:"source"`
	Target string `json:"target"`
	Type   string `json:"type"`
	Label  string `json:"label,omitempty"`
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
	AiTreeResponse string `json:"aiTreeResponse"`
}
