package preprocessing

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aiservice/internal/models"
	"github.com/firebase/genkit/go/ai"
)

const SummarizePrompt = `
Тебе предоставлены:
1. Расшифрованные DIGITAL INK DATA - это текст, распознанный из рукописных заметок пользователей на доске
2. Изображение доски - скриншот для визуальной сверки

Твоя задача:
1) Проанализируй распознанный текст из рукописных заметок (DIGITAL INK DATA)
2) Сверься с изображением доски для понимания контекста
3) При анализе изображение доски стоит в приоритете над распознанным digital ink

Требования к суммаризации:
- Заполни суммаризацию как html, который ограничен следующими тегами: <p>, <br>, <strong>, <em>, <ul>, <ol>, <li>
`

const StructurizePrompt = `
Тебе нужно следовать строго моей инструкции.
Тебе предоставлены:
1. DIGITAL INK DATA - текст, распознанный из рукописных заметок пользователей на доске
2. Изображение доски - скриншот для визуальной сверки

Твоя задача:
1) Проанализируй распознанный текст из рукописных заметок (DIGITAL INK DATA)
2) Сверься с изображением доски для понимания контекста
3) Создай файловую структуру проекта на основе содержимого доски

Мне нужно, чтобы ты предоставил ответ в следующем формате JSON:
{
  "aiTreeResponse": "ASCII дерево файлов (например: project─┬─src\n              └─main.go)",
  "file": {
    "nodes": [
      {"id": "1", "name": "project", "type": "section", "parentId": null},
      {"id": "2", "name": "src", "type": "section", "parentId": "1"},
      {"id": "3", "name": "main.go", "type": "doc", "parentId": "2"}
    ],
    "rootIds": ["1"]
  }
}

Требования:
- nodes - плоский список всех узлов файловой структуры
- rootIds - ID корневых узлов (обычно один)
- parentId - ссылается на ID родительского узла (null для корня)
- type: "section" для папок, "doc" для файлов
`

// Preprocessor transforms raw input data into structured formats for AI processing
type Preprocessor struct{}

// NewPreprocessor creates a new preprocessor instance
func NewPreprocessor() *Preprocessor {
	return &Preprocessor{}
}

// PreprocessSummarizeRequest transforms a raw summarize request into a structured format
func (p *Preprocessor) PreprocessSummarizeRequest(req models.SummarizeRequest, recognizedText string, imageURI string) ([]*ai.Part, error) {
	parts := []*ai.Part{
		ai.NewTextPart(SummarizePrompt),
	}

	// Add recognized ink text if available
	if recognizedText != "" {
		inkText := fmt.Sprintf("\n\nDIGITAL INK RECOGNIZED TEXT:\n%s", recognizedText)
		parts = append(parts, ai.NewTextPart(inkText))
	}

	// Add image if URI is provided

	parts = []*ai.Part{
		ai.NewTextPart("Что ты видишь на доске?"),
	}
	if imageURI != "" {
		parts = append(parts, ai.NewMediaPart("image/jpeg", imageURI))
	}

	return parts, nil
}

// PreprocessStructurizeRequest transforms a raw structurize request into a structured format
func (p *Preprocessor) PreprocessStructurizeRequest(req models.StructurizeRequest) ([]*ai.Part, error) {
	// Check for potential memory issues
	if len(req.Board.Elements) > 1000 {
		return nil, fmt.Errorf("too many elements in board, maximum allowed is 1000")
	}

	// Preserve raw data
	rawData, err := json.Marshal(req.Board)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal raw board data: %w", err)
	}

	// Check if the marshaled data is too large
	if len(rawData) > 10*1024*1024 { // 10MB limit
		return nil, fmt.Errorf("board data too large, maximum allowed is 10MB")
	}

	// Create a structured representation of the file hierarchy
	fileStructure := p.createFileHierarchyDescription(req.File)

	// Check if combined analysis is too large
	totalSize := len(rawData) + len(fileStructure)
	if totalSize > 20*1024*1024 { // 20MB limit for combined analysis
		return nil, fmt.Errorf("combined analysis data too large, maximum allowed is 20MB")
	}

	// Combine raw data with file hierarchy description
	structuredPrompt := fmt.Sprintf(`PROJECT STRUCTURIZATION REQUEST:
%s

FILE HIERARCHY REQUESTED:
%s

RAW BOARD DATA:
%s

%s`,
		req.RequestType, fileStructure, string(rawData), StructurizePrompt)

	parts := []*ai.Part{
		ai.NewTextPart(structuredPrompt),
	}

	// Add image if available
	if req.Board.ImageURL != "" {
		parts = append(parts, ai.NewMediaPart("image/jpeg", req.Board.ImageURL))
	}

	return parts, nil
}

// createFileHierarchyDescription generates a description of the requested file hierarchy
func (p *Preprocessor) createFileHierarchyDescription(file models.File) string {
	var sb strings.Builder

	if file.IsEmpty() {
		sb.WriteString("No specific file structure requested. Create an appropriate structure based on board content.\n")
		return sb.String()
	}

	sb.WriteString(fmt.Sprintf("Requested Structure: %s (%s)\n", file.Name, file.Type))

	if len(file.Children) > 0 {
		sb.WriteString("Child Elements:\n")
		p.writeFileTree(&sb, file.Children, 1)
	} else {
		sb.WriteString("No child elements specified.\n")
	}

	return sb.String()
}

// writeFileTree recursively writes the file tree structure
func (p *Preprocessor) writeFileTree(sb *strings.Builder, files []models.File, depth int) {
	indent := strings.Repeat("  ", depth)

	for _, file := range files {
		sb.WriteString(fmt.Sprintf("%s- %s (%s)\n", indent, file.Name, file.Type))

		if len(file.Children) > 0 {
			p.writeFileTree(sb, file.Children, depth+1)
		}
	}
}
