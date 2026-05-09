package utils

import (
	"log/slog"
	"strings"

	"github.com/aiservice/internal/models"
	"github.com/aiservice/internal/providers"
)

type FileTreeParser struct {
	nodes   []providers.FileNode
	rootIDs []string
}

func ParseASCIITree(asciiTree string) providers.FileHierarchy {
	parser := &FileTreeParser{
		nodes:   make([]providers.FileNode, 0),
		rootIDs: make([]string, 0),
	}
	lines := strings.Split(asciiTree, "\n")
	parser.parseLines(lines)
	return providers.FileHierarchy{
		Nodes:   parser.nodes,
		RootIDs: parser.rootIDs,
	}
}

type stackEntry struct {
	position int // rune-позиция символа ├ или └ (-1 для корня)
	nodeIdx  int
}

func (p *FileTreeParser) parseLines(lines []string) {
	var stack []stackEntry

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		position, name := p.parseLine(line)
		if name == "" {
			continue
		}

		boardName, id, boardType := p.determineNodeType(name)
		nodeID := generateNodeID(len(p.nodes))

		node := providers.FileNode{
			ID:     nodeID,
			RealID: id,
			Name:   boardName,
			Type:   boardType,
		}

		for len(stack) > 0 && stack[len(stack)-1].position >= position {
			stack = stack[:len(stack)-1]
		}

		if len(stack) == 0 {
			p.rootIDs = append(p.rootIDs, nodeID)
			node.ParentID = nil
		} else {
			parentID := p.nodes[stack[len(stack)-1].nodeIdx].ID
			node.ParentID = &parentID
		}

		nodeIdx := len(p.nodes)
		p.nodes = append(p.nodes, node)
		stack = append(stack, stackEntry{position: position, nodeIdx: nodeIdx})
	}
}

// parseLine возвращает rune-позицию последнего ├/└ (или -1 для корня)
// и имя узла без слэша
func (p *FileTreeParser) parseLine(line string) (position int, name string) {
	runes := []rune(line)

	// Ищем последний ├ или └ 
	lastConnector := -1
	for i, r := range runes {
		if r == '├' || r == '└' {
			lastConnector = i
		}
	}

	if lastConnector == -1 {
		name = strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(line), "/"))
		return -1, name
	}

	// Пропускаем ─── и пробелы 
	nameStart := lastConnector + 1
	for nameStart < len(runes) && (runes[nameStart] == '─' || runes[nameStart] == ' ') {
		nameStart++
	}

	if nameStart < len(runes) {
		name = strings.TrimSpace(string(runes[nameStart:]))
		name = strings.TrimSuffix(name, "/")
	}

	return lastConnector, name
}

func (p *FileTreeParser) determineNodeType(name string) (boardName string, id string, boardType string) {
	parts := strings.Split(name, ".")
	switch len(parts) {
	case 0:
		slog.Warn("invalid tree parsing:", "got", name)
		return "", "", ""
	case 1:
		slog.Warn("invalid tree parsing:", "got", name)
		return name, "", ""
	case 2:
		boardName := parts[0]
		boardType := parts[1]
		return boardName, "", boardType
	case 3:
		boardName := parts[0]
		id := parts[1]
		boardType := parts[2]
		return boardName, id, boardType
	}
	return "", "", ""
}

func generateNodeID(index int) string {
	return string(rune('a'+(index%26))) + string(rune('0'+(index/26)))
}

func ToModelFile(h providers.FileHierarchy) models.File {
	if len(h.Nodes) == 0 || len(h.RootIDs) == 0 {
		return models.File{}
	}

	// Индекс узлов
	nodeMap := make(map[string]providers.FileNode, len(h.Nodes))
	for _, node := range h.Nodes {
		nodeMap[node.ID] = node
	}

	// parentID -> упорядоченный список дочерних ID
	childrenMap := make(map[string][]string)
	for _, node := range h.Nodes {
		if node.ParentID != nil {
			childrenMap[*node.ParentID] = append(childrenMap[*node.ParentID], node.ID)
		}
	}

	// Рекурсивно строим models.File
	var build func(id string) models.File
	build = func(id string) models.File {
		node := nodeMap[id]
		file := models.File{
			Name: node.Name,
			Type: node.Type,
			Id:   node.RealID,
		}
		for _, childID := range childrenMap[id] {
			file.Children = append(file.Children, build(childID))
		}
		return file
	}

	return build(h.RootIDs[0])
}
