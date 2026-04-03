package preprocessing

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aiservice/internal/models"
	"github.com/firebase/genkit/go/ai"
)

const GenerateTemplatePrompt = `
Ты ассистент для генерации шаблонов досок на основе текстового описания.

Тебе предоставлен промпт от пользователя с описанием того, какую доску он хочет создать.

Твоя задача:
1) Проанализируй промпт пользователя
2) Определи тип доски (simple или graph) на основе контекста
3) Сгенерируй структуру базову структуру доски

Типы досок:
- simple: простая доска с элементами (rectangle, text, ellipse, line). Подходит для brainstorming, заметок, идей.
- graph: граф с узлами и рёбрами (React Flow). Подходит для архитектуры, зависимостей, workflow.

Требования:
- Для simple board: заполни elements, оставь graphNodes и graphEdges пустыми
- Для graph board: заполни graphNodes и graphEdges, оставь elements пустым
- Позиции элементов должны быть разумными (не выходить за пределы 2000x2000)
- Цвета в формате hex (#RRGGBB)
- Content может содержать HTML теги: <p>, <br>, <strong>, <em>, <ul>, <ol>, <li>
- Title и description должны кратко описывать сгенерированную доску
`

const ImageRecognitionPrompt = `
Что ты видишь на доске?
Опиши всё, что видишь: текст, графы, схемы, рисунки и другие элементы.

Ответ верни в формате JSON:
{
  "imageDescription": "Всё, что ты видишь на доске"
}
`

// SummarizeSystemPrompt contains fixed instructions for summarization
const SummarizeSystemPrompt = `
Тебе предоставлено описание содержимого доски:
1. IMAGE DESCRIPTION - описание того, что видно на изображении доски
2. DIGITAL INK DATA - распознанный текст из рукописных заметок
3. GRAPH STRUCTURE - семантическое представление графа (узлы и связи между ними)

GRAPH STRUCTURE формат:
{
  "nodes": [
    {"id": "...", "label": "...", "kind": "component|service|database|external_link|internal_link|navigation", "description": "...", "url": "..."}
  ],
  "edges": [
    {"from": "...", "to": "...", "label": "..."}
  ]
}

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

// StructurizeSystemPrompt contains fixed instructions for structurization
const StructurizeSystemPrompt = `
Тебе предоставлено описание содержимого доски:
1. IMAGE DESCRIPTION - описание того, что видно на изображении доски (текст, графы, схемы, рисунки)
2. DIGITAL INK DATA - текст из рукописных заметок
3. GRAPH STRUCTURE - семантическое представление графа (узлы и связи между ними)

GRAPH STRUCTURE формат:
{
  "nodes": [
    {"id": "...", "label": "...", "kind": "component|service|database|external_link|internal_link|navigation", "description": "...", "url": "..."}
  ],
  "edges": [
    {"from": "...", "to": "...", "label": "..."}
  ]
}

Твоя задача:
1) Проанализируй всё предоставленное описание
2) Создай файловую структуру проекта на основе содержимого доски

Мне нужно, чтобы ты предоставил ответ в формате ASCII TREE дерева. Вот пример:
main/
├── test
│   ├── doc.x
│   └── file.txt
└── test.go
`

// Preprocessor transforms raw input data into structured formats for AI processing
type Preprocessor struct{}

// NewPreprocessor creates a new preprocessor instance
func NewPreprocessor() *Preprocessor {
	return &Preprocessor{}
}

// PreprocessGenerateTemplateRequest prepares data for template generation
func PreprocessGenerateTemplateRequest(prompt string, boardType models.BoardType) []*ai.Part {
	var userData strings.Builder

	userData.WriteString("USER PROMPT:\n")
	userData.WriteString(prompt)
	userData.WriteString("\n\n")

	userData.WriteString("REQUESTED BOARD TYPE: ")
	userData.WriteString(string(boardType))
	userData.WriteString("\n\n")

	userData.WriteString("Generate a board template based on the user's prompt.")

	return []*ai.Part{
		ai.NewTextPart(userData.String()),
	}
}

// PreprocessSummarizeData prepares dynamic data for user message (data only, no instructions)
func PreprocessSummarizeData(imageDescription string, recognizedText string, semanticGraph *models.SemanticGraph) string {
	var userData strings.Builder

	// Добавляем семантический граф если есть
	if semanticGraph != nil {
		graphJSON, _ := json.MarshalIndent(semanticGraph, "", "  ")
		userData.WriteString("GRAPH STRUCTURE:\n")
		userData.WriteString(string(graphJSON))
		userData.WriteString("\n\n")
	}

	// Добавляем image description если есть
	if imageDescription != "" {
		userData.WriteString("IMAGE DESCRIPTION:\n")
		userData.WriteString(imageDescription)
		userData.WriteString("\n\n")
	}

	// Добавляем digital ink data если есть
	if recognizedText != "" {
		userData.WriteString("DIGITAL INK DATA:\n")
		userData.WriteString(recognizedText)
	}

	return userData.String()
}

// PreprocessStructurizeData prepares dynamic data for user message (data only, no instructions)
func PreprocessStructurizeData(imageDescription string, inkData string, semanticGraph *models.SemanticGraph) string {
	var userData strings.Builder

	// Добавляем семантический граф если есть
	if semanticGraph != nil {
		graphJSON, _ := json.MarshalIndent(semanticGraph, "", "  ")
		userData.WriteString("GRAPH STRUCTURE:\n")
		userData.WriteString(string(graphJSON))
		userData.WriteString("\n\n")
	}

	// Добавляем image description если есть
	if imageDescription != "" {
		userData.WriteString("IMAGE DESCRIPTION:\n")
		userData.WriteString(imageDescription)
		userData.WriteString("\n\n")
	}

	// Добавляем digital ink data если есть
	if inkData != "" {
		userData.WriteString("DIGITAL INK DATA:\n")
		userData.WriteString(inkData)
	}

	return userData.String()
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
