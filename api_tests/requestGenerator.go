package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/aiservice/internal/models"
)

type RawBoard struct {
	BoardID string `json:"boardId"`
	Type    string `json:"type"`
	State   struct {
		Layers [][]RawShape `json:"layers"`
	} `json:"state"`
}

type RawShape struct {
	ID           string    `json:"id"`
	Type         string    `json:"type"` // rect, line, text, ellipse
	X            float64   `json:"x"`
	Y            float64   `json:"y"`
	Width        float64   `json:"width"`
	Height       float64   `json:"height"`
	Rotation     float64   `json:"rotation"`
	Fill         string    `json:"fill,omitempty"`
	Stroke       string    `json:"stroke,omitempty"`
	StrokeWidth  int       `json:"strokeWidth,omitempty"`
	CornerRadius int       `json:"cornerRadius,omitempty"`
	Content      string    `json:"content,omitempty"`
	Points       []float64 `json:"points,omitempty"` // digital ink: [x1,y1,x2,y2,...]
	Tension      float64   `json:"tension,omitempty"`
	LineCap      string    `json:"lineCap,omitempty"`
	LineJoin     string    `json:"lineJoin,omitempty"`
}

type Generator struct {
	baseRawBoard RawBoard
	rng          *rand.Rand
}

func NewGenerator(rawJSON []byte) (*Generator, error) {
	var raw RawBoard
	if err := json.Unmarshal(rawJSON, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse raw board: %w", err)
	}

	return &Generator{
		baseRawBoard: raw,
		rng:          rand.New(rand.NewSource(time.Now().UnixNano())),
	}, nil
}

type Config struct {
	ScalePoints      int
	ExtraElements    int
	ProbabilityGraph float64
	IncludeInk       bool
	Seed             int64
}

func DefaultConfig() Config {
	return Config{
		ScalePoints:      10,
		ExtraElements:    20,
		ProbabilityGraph: 0.3,
		IncludeInk:       true,
		Seed:             time.Now().UnixNano(),
	}
}

func HeavyConfig() Config {
	return Config{
		ScalePoints:      100,
		ExtraElements:    200,
		ProbabilityGraph: 0.5,
		IncludeInk:       true,
		Seed:             42,
	}
}

func LightConfig() Config {
	return Config{
		ScalePoints:      1,
		ExtraElements:    5,
		ProbabilityGraph: 0.1,
		IncludeInk:       false,
		Seed:             time.Now().UnixNano(),
	}
}

func (g *Generator) GenerateSummarize(cfg Config) (*models.SummarizeRequest, error) {
	board, err := g.convertBoard(cfg)
	if err != nil {
		return nil, err
	}

	return &models.SummarizeRequest{
		RequestID:   g.genID("sum_"),
		UserID:      g.genID("user_"),
		RequestType: "summarize",
		Board:       board,
	}, nil
}

func (g *Generator) GenerateStructurize(cfg Config) (*models.StructurizeRequest, error) {
	board, err := g.convertBoard(cfg)
	if err != nil {
		return nil, err
	}

	return &models.StructurizeRequest{
		RequestID:   g.genID("str_"),
		UserID:      g.genID("user_"),
		RequestType: "structurize",
		Board:       board,
		File:        g.generateFileTree(3, 5),
	}, nil
}

func (g *Generator) GenerateTemplate(cfg Config) (*models.GenerateTemplateRequest, error) {
	boardType := models.BoardTypeSimple
	if g.rng.Float64() < cfg.ProbabilityGraph {
		boardType = models.BoardTypeGraph
	}

	prompts := []string{
		"Создай шаблон для планирования спринта",
		"Сгенерируй структуру для технической документации",
		"Подготовь доску для ретроспективы команды",
		"Создай шаблон mindmap для brainstorming-сессии",
		"Сгенерируй структуру для онбординга нового сотрудника",
		"Подготовь шаблон для отчета по проекту",
		"Создай доску для планирования маркетинговой кампании",
		"Сгенерируй структуру для учебного курса",
		"Подготовь шаблон для управления задачами в команде",
		"Создай шаблон для анализа конкурентов",
		"Сгенерируй структуру для planning-сессии",
		"Подготовь шаблон для проведения SWOT-анализа",
		"Создай доску для организации конференции",
		"Сгенерируй структуру для создания продукта",
		"Подготовь шаблон для управления рисками в проекте",
		"Создай шаблон для оценки рисков в проекте",
	}

	return &models.GenerateTemplateRequest{
		RequestID:   g.genID("tpl_"),
		UserID:      g.genID("user_"),
		RequestType: "generateTemplate",
		BoardID:     g.baseRawBoard.BoardID,
		Prompt:      prompts[g.rng.Intn(len(prompts))],
		BoardType:   boardType,
	}, nil
}

func (g *Generator) convertBoard(cfg Config) (models.Board, error) {
	if cfg.Seed != 0 {
		g.rng = rand.New(rand.NewSource(cfg.Seed))
	}

	elements := make([]models.Element, 0)

	for _, layer := range g.baseRawBoard.State.Layers {
		for _, shape := range layer {
			if shape.Type == "line" && !cfg.IncludeInk {
				continue
			}

			elem := models.Element{
				Id:           shape.ID,
				Type:         g.mapShapeType(shape.Type),
				X:            float32(shape.X),
				Y:            float32(shape.Y),
				Width:        float32(shape.Width),
				Height:       float32(shape.Height),
				Rotation:     float32(shape.Rotation),
				Fill:         shape.Fill,
				Stroke:       shape.Stroke,
				StrokeWidth:  shape.StrokeWidth,
				CornerRadius: shape.CornerRadius,
				Content:      shape.Content,
				Tension:      float32(shape.Tension),
			}

			if len(shape.Points) > 0 && cfg.ScalePoints > 1 {
				elem.Points = g.scalePoints(shape.Points, cfg.ScalePoints)
			} else if len(shape.Points) > 0 {
				elem.Points = make([]float32, len(shape.Points))
				for i, v := range shape.Points {
					elem.Points[i] = float32(v)
				}
			}

			elements = append(elements, elem)
		}
	}

	elements = append(elements, g.generateExtraElements(cfg.ExtraElements)...)

	var graphNodes []models.GNode
	var graphEdges []models.GEdge
	if g.rng.Float64() < cfg.ProbabilityGraph {
		graphNodes, graphEdges = g.generateGraph(5, 8)
	}

	return models.Board{
		BoardID:    g.baseRawBoard.BoardID,
		ImageURL:   "",
		Elements:   elements,
		GraphNodes: graphNodes,
		GraphEdges: graphEdges,
	}, nil
}

func (g *Generator) mapShapeType(rawType string) string {
	switch rawType {
	case "rect":
		return "rectangle"
	case "line":
		return "line"
	case "text":
		return "text"
	case "ellipse", "circle":
		return "ellipse"
	default:
		return "rectangle" // fallback
	}
}

func (g *Generator) scalePoints(points []float64, factor int) []float32 {
	if factor <= 1 || len(points) == 0 {
		result := make([]float32, len(points))
		for i, v := range points {
			result[i] = float32(v)
		}
		return result
	}

	original := make([]float64, len(points))
	copy(original, points)

	scaled := make([]float64, 0, len(points)*factor)
	for range factor {
		for _, v := range original {
			noise := (g.rng.Float64() - 0.5) * 0.01
			scaled = append(scaled, v+noise)
		}
	}

	result := make([]float32, len(scaled))
	for i, v := range scaled {
		result[i] = float32(v)
	}
	return result
}

func (g *Generator) generateExtraElements(count int) []models.Element {
	elements := make([]models.Element, 0, count)
	types := []string{"rectangle", "text", "ellipse", "line"}

	for range count {
		elemType := types[g.rng.Intn(len(types))]
		elem := models.Element{
			Id:       g.genID(elemType + "_"),
			Type:     elemType,
			X:        float32(g.rng.Intn(2000)),
			Y:        float32(g.rng.Intn(1500)),
			Width:    float32(50 + g.rng.Intn(300)),
			Height:   float32(30 + g.rng.Intn(200)),
			Rotation: float32(g.rng.Intn(360)),
			Fill:     fmt.Sprintf("#%06X", g.rng.Intn(0xFFFFFF)),
		}

		if elemType == "text" {
			elem.Content = g.genLorem(10, 50)
		}
		if elemType == "line" {
			pointCount := 10 + g.rng.Intn(40)
			elem.Points = make([]float32, pointCount*2)
			for j := range elem.Points {
				elem.Points[j] = float32(g.rng.Intn(2000))
			}
		}

		elements = append(elements, elem)
	}
	return elements
}

func (g *Generator) generateFileTree(maxDepth, maxChildren int) models.File {
	return g.buildFileNode(0, maxDepth, maxChildren)
}

func (g *Generator) buildFileNode(depth, maxDepth, maxChildren int) models.File {
	isFolder := depth < maxDepth && g.rng.Float64() < 0.7
	nodeType := "section"
	if !isFolder {
		nodeType = "doc"
	}

	node := models.File{
		Name: g.genName(isFolder),
		Type: nodeType,
	}

	if isFolder && depth < maxDepth {
		childCount := g.rng.Intn(maxChildren) + 1
		node.Children = make([]models.File, childCount)
		for i := range node.Children {
			node.Children[i] = g.buildFileNode(depth+1, maxDepth, maxChildren)
		}
	}

	return node
}

func (g *Generator) generateGraph(minNodes, maxNodes int) ([]models.GNode, []models.GEdge) {
	nodeCount := minNodes + g.rng.Intn(maxNodes-minNodes+1)
	nodes := make([]models.GNode, nodeCount)
	edges := make([]models.GEdge, 0)

	for i := range nodeCount {
		nodes[i] = models.GNode{
			ID:   g.genID("node_"),
			Type: []string{"customNode", "externalLinkNode", "internalLinkNode"}[g.rng.Intn(3)],
			Data: models.GNodeData{
				Title:       g.genLorem(3, 8),
				Description: g.genLorem(10, 30),
				URL:         "",
				Element:     nil,
			},
		}
	}

	for i := range nodes {
		edgeCount := g.rng.Intn(4)
		for range edgeCount {
			targetIdx := g.rng.Intn(nodeCount)
			if targetIdx == i {
				continue
			}
			label := ""
			if g.rng.Float64() < 0.3 {
				label = g.genLorem(2, 5)
			}
			edges = append(edges, models.GEdge{
				ID:     g.genID("edge_"),
				Source: nodes[i].ID,
				Target: nodes[targetIdx].ID,
				Label:  label,
			})
		}
	}

	return nodes, edges
}

func (g *Generator) genID(prefix string) string {
	return fmt.Sprintf("%s%x", prefix, g.rng.Uint64())
}

func (g *Generator) genName(isFolder bool) string {
	if isFolder {
		folders := []string{"Documents", "Projects", "Archive", "Drafts", "Templates", "Resources"}
		return folders[g.rng.Intn(len(folders))] + fmt.Sprintf("_%d", g.rng.Intn(100))
	}
	exts := []string{".md", ".txt", ".pdf", ".docx", ".json"}
	return "file_" + g.genID("") + exts[g.rng.Intn(len(exts))]
}

func (g *Generator) genLorem(minWords, maxWords int) string {
	words := []string{"lorem", "ipsum", "dolor", "sit", "amet", "consectetur", "adipiscing", "elit", "sed", "do", "eiusmod", "tempor", "incididunt", "ut", "labore", "et", "dolore", "magna", "aliqua"}
	count := minWords + g.rng.Intn(maxWords-minWords+1)
	parts := make([]string, count)
	for i := range parts {
		parts[i] = words[g.rng.Intn(len(words))]
	}
	return strings.Join(parts, " ")
}

func runRequestGenerator() {
	raw, err := os.ReadFile("data.txt")
	if err != nil {
		raw, err = os.ReadFile("../data.txt")
		if err != nil {
			panic(fmt.Errorf("failed to read data.txt: %w", err))
		}
	}

	gen, err := NewGenerator(raw)
	if err != nil {
		panic(fmt.Errorf("failed to create generator: %w", err))
	}

	if err := os.MkdirAll("request_fixtures", 0755); err != nil {
		panic(fmt.Errorf("failed to create fixtures directory: %w", err))
	}

	generators := []func() error{
		// generator(gen.GenerateSummarize, LightConfig(), 10, "summarize_light"),
		// generator(gen.GenerateStructurize, LightConfig(), 10, "structurize_light"),
		// generator(gen.GenerateTemplate, LightConfig(), 10, "template_light"),

		// generator(gen.GenerateSummarize, DefaultConfig(), 5, "summarize_default"),
		// generator(gen.GenerateStructurize, DefaultConfig(), 5, "structurize_default"),
		// generator(gen.GenerateTemplate, DefaultConfig(), 5, "template_default"),

		generator(gen.GenerateSummarize, HeavyConfig(), 200, "summarize_heavy"),
		generator(gen.GenerateStructurize, HeavyConfig(), 200, "structurize_heavy"),
		generator(gen.GenerateTemplate, HeavyConfig(), 200, "template_heavy"),
	}

	for _, generator := range generators {
		if err := generator(); err != nil {
			panic(err)
		}
	}
}

func generator[T any](generate func(cfg Config) (T, error), cfg Config, count int, prefix string) func() error {
	return func() error {
		for i := range count {
			req, err := generate(cfg)
			if err != nil {
				return fmt.Errorf("failed to generate %s %d: %v\n", prefix, i, err)
			}
			if err := SaveToFile(req, fmt.Sprintf("request_fixtures/%s_%d.json", prefix, i)); err != nil {
				return fmt.Errorf("failed to save %s %d: %v\n", prefix, i, err)
			}
		}
		return nil
	}
}
