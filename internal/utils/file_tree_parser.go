package utils

import (
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

		nodeType := p.determineNodeType(name)
		nodeID := generateNodeID(len(p.nodes))

		node := providers.FileNode{
			ID:   nodeID,
			Name: name,
			Type: nodeType,
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

func (p *FileTreeParser) determineNodeType(name string) string {
	fileExtensions := []string{
		".go", ".py", ".js", ".ts", ".java", ".c", ".cpp", ".h", ".hpp",
		".rb", ".rs", ".php", ".swift", ".kt", ".scala", ".cs",
		".md", ".txt", ".rst", ".adoc",
		".json", ".yaml", ".yml", ".toml", ".xml",
		".html", ".css", ".scss", ".less",
		".sh", ".bash", ".zsh", ".fish",
		".dockerfile", ".gitignore", ".env",
		".pdf", ".doc", ".docx", ".xls", ".xlsx",
		".png", ".jpg", ".jpeg", ".gif", ".svg", ".ico",
		".mp3", ".mp4", ".avi", ".mov",
		".zip", ".tar", ".gz", ".rar", ".7z",
	}

	nameLower := strings.ToLower(name)
	for _, ext := range fileExtensions {
		if strings.HasSuffix(nameLower, ext) {
			return "doc"
		}
	}
	if strings.Contains(name, ".") {
		return "doc"
	}
	return "section"
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
		}
		for _, childID := range childrenMap[id] {
			file.Children = append(file.Children, build(childID))
		}
		return file
	}

	return build(h.RootIDs[0])
}
