package providers

import (
	"context"
	"fmt"
	"sync"

	"github.com/aiservice/internal/models"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/genkit"
)

type LLMClient interface {
	Structurize(ctx context.Context, parts []*ai.Part) (models.StructurizeResponse, error)
	Summarize(ctx context.Context, parts []*ai.Part) (models.SummarizeResponse, error)
	GetName() string // Added for provider identification
}

type SummarizeFlow struct {
	Prompt  string      `json:"userPrompt"`
	Element models.Text `json:"element"`
}

// For the recursive structure, we'll use a different approach that doesn't trigger schema generation
// We'll use a direct call to the LLM instead of Genkit's flow system for the recursive structure

var summarizeFlowInstance *core.Flow[*SummarizeFlow, *SummarizeFlow, struct{}]
var summarizeFlowOnce sync.Once

func GetSummarizeFlow(gkit *genkit.Genkit) *core.Flow[*SummarizeFlow, *SummarizeFlow, struct{}] {
	summarizeFlowOnce.Do(func() {
		summarizeFlowInstance = genkit.DefineFlow(gkit, "summarize flow", func(ctx context.Context, input *SummarizeFlow) (*SummarizeFlow, error) {
			// Note: This flow is not meant to be run directly, use GenerateData instead
			return nil, fmt.Errorf("this flow is not meant to be run directly, use GenerateData instead")
		})
	})
	return summarizeFlowInstance
}

func RunSummarizeGeneration(ctx context.Context, gkit *genkit.Genkit, parts []*ai.Part) (*SummarizeFlow, error) {
	prompt := ai.NewUserMessage(parts...)
	resp, _, err := genkit.GenerateData[SummarizeFlow](ctx, gkit, ai.WithMessages(prompt))
	if err != nil {
		return nil, fmt.Errorf("failed to generate llm request: %w", err)
	}
	return resp, nil
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

var structurizeFlowInstance *core.Flow[*SimpleStructurizeFlow, *SimpleStructurizeFlow, struct{}]
var structurizeFlowOnce sync.Once

func GetStructurizeFlow(gkit *genkit.Genkit) *core.Flow[*SimpleStructurizeFlow, *SimpleStructurizeFlow, struct{}] {
	structurizeFlowOnce.Do(func() {
		structurizeFlowInstance = genkit.DefineFlow(gkit, "structurize flow", func(ctx context.Context, input *SimpleStructurizeFlow) (*SimpleStructurizeFlow, error) {
			// Note: This flow is not meant to be run directly, use GenerateData instead
			return nil, fmt.Errorf("this flow is not meant to be run directly, use GenerateData instead")
		})
	})
	return structurizeFlowInstance
}

func RunStructurizeGeneration(ctx context.Context, gkit *genkit.Genkit, parts []*ai.Part) (*SimpleStructurizeFlow, error) {
	prompt := ai.NewUserMessage(parts...)
	resp, _, err := genkit.GenerateData[SimpleStructurizeFlow](ctx, gkit, ai.WithMessages(prompt))
	if err != nil {
		return nil, fmt.Errorf("failed to generate llm request: %w", err)
	}
	return resp, nil
}

// RunStructurizeGenerationAndConvert executes the structurize generation and converts the result to the original File model
func RunStructurizeGenerationAndConvert(ctx context.Context, gkit *genkit.Genkit, parts []*ai.Part) (models.File, string, error) {
	prompt := ai.NewUserMessage(parts...)
	resp, _, err := genkit.GenerateData[SimpleStructurizeFlow](ctx, gkit, ai.WithMessages(prompt))
	if err != nil {
		return models.File{}, "", fmt.Errorf("failed to generate llm request: %w", err)
	}

	// Convert the flat hierarchy to the original recursive File model
	modelFile := resp.File.ToModelFile()

	return modelFile, resp.AiTreeResponse, nil
}
