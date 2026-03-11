package preprocessing

import (
	"fmt"
	"strings"

	"github.com/aiservice/internal/models"
	"github.com/firebase/genkit/go/ai"
)

const ImageRecognitionPrompt = `
Что ты видишь на доске?
Опиши всё, что видишь: текст, графы, схемы, рисунки и другие элементы.

Ответ верни в формате JSON:
{
  "imageDescription": "Всё, что ты видишь на доске"
}
`

const SummarizePrompt = `
Тебе предоставлено описание содержимого доски:
1. IMAGE DESCRIPTION - описание того, что видно на изображении доски
2. DIGITAL INK DATA - распознанный текст из рукописных заметок

Твоя задача:
1) Проанализируй всё предоставленное описание
2) Создай краткую суммаризацию ключевых выводов и заключений

Мне нужно, чтобы ты предоставил ответ в следующем формате JSON:
{
  "summarization": "Краткое содержание того, к чему пришли пользователи на доске"
}

Требования к суммаризации:
- Content - это html тип, который ограничен следующими тегами: <p>, <br>, <strong>, <em>, <ul>, <ol>, <li>
- Будь краток, но информативен
- Отражай ключевые выводы и решения пользователей
`

const StructurizePrompt = `
Тебе предоставлено описание содержимого доски:
1. IMAGE DESCRIPTION - описание того, что видно на изображении доски (текст, графы, схемы, рисунки)
2. DIGITAL INK DATA - текст из рукописных заметок

Твоя задача:
1) Проанализируй всё предоставленное описание
2) Создай файловую структуру проекта на основе содержимого доски

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
func (p *Preprocessor) PreprocessStructurizeRequest(imageDescription string, inkData string) ([]*ai.Part, error) {
	parts := []*ai.Part{
		ai.NewTextPart(StructurizePrompt),
	}

	// Add image description if available
	if imageDescription != "" {
		parts = append(parts, ai.NewTextPart("\n\nIMAGE DESCRIPTION:\n"+imageDescription))
	}

	// Add ink data if available
	if inkData != "" {
		parts = append(parts, ai.NewTextPart("\n\nDIGITAL INK DATA:\n"+inkData))
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
