package graph

import (
	"github.com/aiservice/internal/models"
)

// GraphPreprocessor converts React Flow graph to semantic format
type GraphPreprocessor struct{}

// NewGraphPreprocessor creates a new graph preprocessor
func NewGraphPreprocessor() *GraphPreprocessor {
	return &GraphPreprocessor{}
}

// Process converts React Flow graph nodes and edges to semantic format
func (p *GraphPreprocessor) Process(nodes []models.GNode, edges []models.GEdge) (*models.SemanticGraph, error) {
	semanticNodes := make([]models.SemanticNode, 0, len(nodes))

	for _, node := range nodes {
		semanticNode := p.convertNode(node)
		semanticNodes = append(semanticNodes, semanticNode)
	}

	semanticEdges := make([]models.SemanticEdge, 0, len(edges))
	for _, edge := range edges {
		semanticEdge := models.SemanticEdge{
			From:  edge.Source,
			To:    edge.Target,
			Label: edge.Label,
		}
		semanticEdges = append(semanticEdges, semanticEdge)
	}

	return &models.SemanticGraph{
		Nodes: semanticNodes,
		Edges: semanticEdges,
	}, nil
}

// convertNode maps a React Flow node to a semantic node
func (p *GraphPreprocessor) convertNode(node models.GNode) models.SemanticNode {
	semanticNode := models.SemanticNode{
		ID:    node.ID,
		Label: node.Data.Title,
	}

	// Determine kind based on node type
	switch node.Type {
	case "customNode":
		semanticNode.Kind = "component"
		semanticNode.Description = node.Data.Description
	case "externalLinkNode":
		semanticNode.Kind = "external_link"
		semanticNode.URL = node.Data.URL
		semanticNode.Description = node.Data.Description
		if semanticNode.Label == "" && node.Data.Domain != "" {
			semanticNode.Label = node.Data.Domain
		}
	case "internalLinkNode":
		semanticNode.Kind = "internal_link"
		if node.Data.Element != nil {
			semanticNode.Label = node.Data.Element.Title
		}
	case "nodeLinkNode":
		semanticNode.Kind = "navigation"
		semanticNode.URL = node.Data.URL
	default:
		semanticNode.Kind = "unknown"
	}

	// Fallback for empty label
	if semanticNode.Label == "" {
		if node.Data.Description != "" {
			semanticNode.Label = node.Data.Description
		} else if node.Data.Domain != "" {
			semanticNode.Label = node.Data.Domain
		}
	}

	return semanticNode
}
