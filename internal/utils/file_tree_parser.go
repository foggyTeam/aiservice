package utils

import (
	"strings"

	"github.com/aiservice/internal/models"
	"github.com/aiservice/internal/providers"
)

// FileTreeParser parses ASCII tree representations into FileHierarchy
type FileTreeParser struct {
	nodes   []providers.FileNode
	rootIDs []string
}

// ParseASCIITree parses an ASCII tree string and returns a FileHierarchy
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

func (p *FileTreeParser) parseLines(lines []string) {
	var parentStack []int // Stack of parent node indices

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Determine depth and extract name
		depth, name := p.parseLine(line)
		if name == "" {
			continue
		}

		// Determine if it's a directory or file
		nodeType := p.determineNodeType(name)

		// Create node ID
		nodeID := generateNodeID(len(p.nodes))

		// Create the node
		node := providers.FileNode{
			ID:   nodeID,
			Name: name,
			Type: nodeType,
		}

		// Set parent relationship
		if depth == 0 {
			// Root node
			p.rootIDs = append(p.rootIDs, nodeID)
			node.ParentID = nil
		} else {
			// Find parent from stack
			if depth <= len(parentStack) {
				// Pop stack to correct depth
				parentStack = parentStack[:depth]
			}
			if len(parentStack) > 0 {
				parentID := p.nodes[parentStack[len(parentStack)-1]].ID
				node.ParentID = &parentID
			}
		}

		// Add node and push to stack
		p.nodes = append(p.nodes, node)
		parentStack = append(parentStack, len(p.nodes)-1)
	}
}

func (p *FileTreeParser) parseLine(line string) (depth int, name string) {
	// Count leading spaces to determine depth
	// Each 4 spaces = 1 depth level
	
	leadingSpaces := 0
	for i := 0; i < len(line) && line[i] == ' '; i++ {
		leadingSpaces++
	}
	
	depth = leadingSpaces / 4
	
	// Extract name (everything after leading spaces)
	if leadingSpaces < len(line) {
		name = strings.TrimSpace(line[leadingSpaces:])
	}
	
	return depth, name
}

func (p *FileTreeParser) determineNodeType(name string) string {
	// Common file extensions
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

	// Check if it's a file with known extension
	for _, ext := range fileExtensions {
		if strings.HasSuffix(nameLower, ext) {
			return "doc"
		}
	}

	// If name contains a dot, treat as file
	if strings.Contains(name, ".") {
		return "doc"
	}

	// Default to section (directory)
	return "section"
}

func generateNodeID(index int) string {
	return string(rune('a'+(index%26))) + string(rune('0'+(index/26)))
}

// ToModelFile converts FileHierarchy to models.File
func ToModelFile(h providers.FileHierarchy) models.File {
	if len(h.Nodes) == 0 {
		return models.File{}
	}

	// Create a map of all nodes by ID for quick lookup
	nodeMap := make(map[string]providers.FileNode)
	for _, node := range h.Nodes {
		nodeMap[node.ID] = node
	}

	// Create a map to store the model files we create
	modelMap := make(map[string]models.File)

	// First pass: create all model files without children
	for _, node := range h.Nodes {
		modelMap[node.ID] = models.File{
			Name: node.Name,
			Type: node.Type,
		}
	}

	// Second pass: assign children to each parent
	childrenMap := make(map[string][]models.File)
	for _, node := range h.Nodes {
		if node.ParentID != nil {
			childrenMap[*node.ParentID] = append(childrenMap[*node.ParentID], modelMap[node.ID])
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
	if len(h.RootIDs) > 0 {
		if rootFile, exists := modelMap[h.RootIDs[0]]; exists {
			rootNode = rootFile
		}
	}

	return rootNode
}
