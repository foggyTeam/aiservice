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

const GenererateTextPrompt = `
Ты ассистент для генерации текста.

Тебе предоставлен промпт от пользователя с описанием того, какой контент он хочет создать.

Твоя задача:
1) Проанализируй промпт пользователя
2) Сгенерировать текст-ответ на запрос пользователя.

Требования:
- Content(твой ответ) должен быть обернут в HTML теги: <p>, <br>, <strong>, <em>, <ul>, <ol>, <li>
`

const ImageRecognitionPrompt = `Что ты видишь на картинке?`

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
4. PROJECT STRUCTURE - иерархическое представление файловой структуры проекта (если есть)

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
2) Создай файловую структуру проекта на основе содержимого доски и предоставленной иерархии (если есть)

Мне нужно, чтобы ты предоставил ответ в формате ASCII TREE дерева. Вот пример:
задачи
├── процессы
│   ├── тестовое задание
│   └── договор
└── регламент
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

func PreprocessGenerateTextRequest(prompt string) []*ai.Part {
	var userData strings.Builder
	userData.WriteString("Generate a text content based on the user's prompt.")
	fmt.Fprintf(&userData, "User prompt: %s\n", prompt)
	return []*ai.Part{ai.NewTextPart(userData.String())}
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
func PreprocessStructurizeData(imageDescription string, inkData string, semanticGraph *models.SemanticGraph, existedStructure models.File) string {
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

	if fileHierarchy := createFileHierarchyDescription(existedStructure); fileHierarchy != "" {
		userData.WriteString("\n\n")
		userData.WriteString("PROJECT STRUCTURE:\n")
		userData.WriteString(fileHierarchy)
	}

	return userData.String()
}

// createFileHierarchyDescription generates a description of the requested file hierarchy
func createFileHierarchyDescription(file models.File) string {
	var sb strings.Builder
	if file.IsEmpty() {
		return ""
	}
	sb.WriteString(file.Name)
	if len(file.Children) > 0 {
		sb.WriteString("\n")
		writeFileTree(&sb, file.Children, "")
	} else {
		sb.WriteString("\n")
	}
	return sb.String()
}

// writeFileTree recursively writes the file tree structure
func writeFileTree(sb *strings.Builder, files []models.File, prefix string) {
	for idx, file := range files {
		isLast := idx == len(files)-1
		branch := "├── "
		childPrefix := prefix + "│   "
		if isLast {
			branch = "└── "
			childPrefix = prefix + "    "
		}
		sb.WriteString(prefix + branch + file.Name + "\n")
		if len(file.Children) > 0 {
			writeFileTree(sb, file.Children, childPrefix)
		}
	}
}
