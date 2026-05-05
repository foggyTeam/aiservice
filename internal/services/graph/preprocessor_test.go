package graph

import (
	"fmt"
	"testing"

	"github.com/aiservice/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcess_EmptyInput(t *testing.T) {
	t.Parallel()

	p := NewGraphPreprocessor()

	result, err := p.Process([]models.GNode{}, []models.GEdge{})

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Empty(t, result.Nodes)
	assert.Empty(t, result.Edges)
}

func TestProcess_NodesOnly(t *testing.T) {
	t.Parallel()

	p := NewGraphPreprocessor()
	nodes := []models.GNode{
		{
			ID:   "node-1",
			Type: "customNode",
			Data: models.GNodeData{
				Title:       "Заголовок 1",
				Description: "Описание 1",
			},
		},
		{
			ID:   "node-2",
			Type: "externalLinkNode",
			Data: models.GNodeData{
				Title: "Ссылка",
				URL:   "https://example.com",
			},
		},
	}

	result, err := p.Process(nodes, []models.GEdge{})

	require.NoError(t, err)
	assert.Len(t, result.Nodes, 2)
	assert.Empty(t, result.Edges)

	assert.Equal(t, "node-1", result.Nodes[0].ID)
	assert.Equal(t, "Заголовок 1", result.Nodes[0].Label)
	assert.Equal(t, "component", result.Nodes[0].Kind)
	assert.Equal(t, "Описание 1", result.Nodes[0].Description)

	assert.Equal(t, "node-2", result.Nodes[1].ID)
	assert.Equal(t, "external_link", result.Nodes[1].Kind)
	assert.Equal(t, "https://example.com", result.Nodes[1].URL)
}

func TestProcess_EdgesOnly(t *testing.T) {
	t.Parallel()

	p := NewGraphPreprocessor()
	edges := []models.GEdge{
		{ID: "e1", Source: "A", Target: "B", Label: "связь"},
		{ID: "e2", Source: "B", Target: "C", Label: ""},
	}

	result, err := p.Process([]models.GNode{}, edges)

	require.NoError(t, err)
	assert.Empty(t, result.Nodes)
	assert.Len(t, result.Edges, 2)

	assert.Equal(t, "A", result.Edges[0].From)
	assert.Equal(t, "B", result.Edges[0].To)
	assert.Equal(t, "связь", result.Edges[0].Label)

	assert.Equal(t, "B", result.Edges[1].From)
	assert.Equal(t, "C", result.Edges[1].To)
	assert.Empty(t, result.Edges[1].Label)
}

func TestProcess_FullGraph(t *testing.T) {
	t.Parallel()

	p := NewGraphPreprocessor()
	nodes := []models.GNode{
		{ID: "start", Type: "customNode", Data: models.GNodeData{Title: "Старт"}},
		{ID: "end", Type: "customNode", Data: models.GNodeData{Title: "Финиш"}},
	}
	edges := []models.GEdge{
		{Source: "start", Target: "end", Label: "переход"},
	}

	result, err := p.Process(nodes, edges)

	require.NoError(t, err)
	assert.Len(t, result.Nodes, 2)
	assert.Len(t, result.Edges, 1)
	assert.Equal(t, "start", result.Edges[0].From)
	assert.Equal(t, "end", result.Edges[0].To)
}

func TestConvertNode_CustomNode(t *testing.T) {
	t.Parallel()

	p := NewGraphPreprocessor()
	node := models.GNode{
		ID:   "c1",
		Type: "customNode",
		Data: models.GNodeData{
			Title:       "Компонент",
			Description: "Это компонент",
		},
	}

	semantic := p.convertNode(node)

	assert.Equal(t, "c1", semantic.ID)
	assert.Equal(t, "Компонент", semantic.Label)
	assert.Equal(t, "component", semantic.Kind)
	assert.Equal(t, "Это компонент", semantic.Description)
	assert.Empty(t, semantic.URL)
}

func TestConvertNode_ExternalLinkNode(t *testing.T) {
	t.Parallel()

	p := NewGraphPreprocessor()
	node := models.GNode{
		ID:   "ext1",
		Type: "externalLinkNode",
		Data: models.GNodeData{
			Title:       "Google",
			Description: "Поисковик",
			URL:         "https://google.com",
		},
	}

	semantic := p.convertNode(node)

	assert.Equal(t, "ext1", semantic.ID)
	assert.Equal(t, "Google", semantic.Label)
	assert.Equal(t, "external_link", semantic.Kind)
	assert.Equal(t, "Поисковик", semantic.Description)
	assert.Equal(t, "https://google.com", semantic.URL)
}

func TestConvertNode_InternalLinkNode_WithElement(t *testing.T) {
	t.Parallel()

	p := NewGraphPreprocessor()
	node := models.GNode{
		ID:   "int1",
		Type: "internalLinkNode",
		Data: models.GNodeData{
			Title:   "Игнорируемый заголовок",
			Element: &models.GraphElement{},
		},
	}

	semantic := p.convertNode(node)

	assert.Equal(t, "int1", semantic.ID)
	assert.Equal(t, "internal_link", semantic.Kind)
}

func TestConvertNode_InternalLinkNode_FallbackToDescription(t *testing.T) {
	t.Parallel()

	p := NewGraphPreprocessor()
	node := models.GNode{
		ID:   "int2",
		Type: "internalLinkNode",
		Data: models.GNodeData{
			Description: "Описание внутренней ссылки",
			Element:     &models.GraphElement{},
		},
	}

	semantic := p.convertNode(node)

	assert.Equal(t, "internal_link", semantic.Kind)
	assert.Equal(t, "Описание внутренней ссылки", semantic.Label)
}

func TestConvertNode_NodeLinkNode(t *testing.T) {
	t.Parallel()

	p := NewGraphPreprocessor()
	node := models.GNode{
		ID:   "nav1",
		Type: "nodeLinkNode",
		Data: models.GNodeData{
			Title: "Навигация",
			URL:   "/dashboard",
		},
	}

	semantic := p.convertNode(node)

	assert.Equal(t, "nav1", semantic.ID)
	assert.Equal(t, "Навигация", semantic.Label)
	assert.Equal(t, "navigation", semantic.Kind)
	assert.Equal(t, "/dashboard", semantic.URL)
}

func TestConvertNode_UnknownType(t *testing.T) {
	t.Parallel()

	p := NewGraphPreprocessor()
	node := models.GNode{
		ID:   "unknown1",
		Type: "someUnknownType",
		Data: models.GNodeData{
			Title: "Неизвестный узел",
		},
	}

	semantic := p.convertNode(node)

	assert.Equal(t, "unknown1", semantic.ID)
	assert.Equal(t, "Неизвестный узел", semantic.Label)
	assert.Equal(t, "unknown", semantic.Kind)
}

func TestConvertNode_EmptyLabel_FallbackToDescription(t *testing.T) {
	t.Parallel()

	p := NewGraphPreprocessor()
	node := models.GNode{
		ID:   "fallback1",
		Type: "customNode",
		Data: models.GNodeData{
			Title:       "",
			Description: "Описание вместо заголовка",
		},
	}

	semantic := p.convertNode(node)

	assert.Equal(t, "Описание вместо заголовка", semantic.Label)
	assert.Equal(t, "component", semantic.Kind)
}

func TestConvertNode_EmptyLabelAndDescription(t *testing.T) {
	t.Parallel()

	p := NewGraphPreprocessor()
	node := models.GNode{
		ID:   "empty1",
		Type: "customNode",
		Data: models.GNodeData{
			Title:       "",
			Description: "",
		},
	}

	semantic := p.convertNode(node)

	assert.Equal(t, "empty1", semantic.ID)
	assert.Empty(t, semantic.Label)
	assert.Equal(t, "component", semantic.Kind)
}

func TestProcess_LargeGraph_NoPanic(t *testing.T) {
	t.Parallel()

	p := NewGraphPreprocessor()

	nodes := make([]models.GNode, 1000)
	for i := range nodes {
		nodes[i] = models.GNode{
			ID:   fmt.Sprintf("node-%d", i),
			Type: "customNode",
			Data: models.GNodeData{Title: fmt.Sprintf("Title %d", i)},
		}
	}
	edges := make([]models.GEdge, 2000)
	for i := range edges {
		edges[i] = models.GEdge{
			Source: fmt.Sprintf("node-%d", i%1000),
			Target: fmt.Sprintf("node-%d", (i+1)%1000),
			Label:  "link",
		}
	}

	result, err := p.Process(nodes, edges)

	require.NoError(t, err)
	assert.Len(t, result.Nodes, 1000)
	assert.Len(t, result.Edges, 2000)
}

func TestProcess_NilSlices_TreatedAsEmpty(t *testing.T) {
	t.Parallel()

	p := NewGraphPreprocessor()

	result, err := p.Process(nil, nil)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Empty(t, result.Nodes)
	assert.Empty(t, result.Edges)
}
