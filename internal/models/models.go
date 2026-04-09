package models

import "fmt"

const (
	SummarizeType        = "summarize"
	StructurizeType      = "structurize"
	GenerateTemplateType = "generateTemplate"
	IncrementalType      = "incremental"
)

const (
	RectangeType = "rectangle"
	TextType     = "text"
	EllipseType  = "ellipse"
	LineTypeType = "line"
)

// BoardType represents the type of board
type BoardType string

const (
	BoardTypeSimple BoardType = "simple"
	BoardTypeGraph  BoardType = "graph"
)

// type Rectangle struct {
// 	BaseElement
// 	CornerRadius int `json:"cornerRadius"`
// }

type Text struct {
	BaseElement
	Content string `json:"content" jsonschema:"description=Текстовое содержимое элемента"`
}

// type Ellipse struct {
// 	BaseElement
// }

// type Line struct {
// 	BaseElement
// 	Points  []float32 `json:"points"`  // [x, y], [x, y]
// 	Tension float32   `json:"tension"` // давление
// }

// type Elements struct {
// 	Ellipse
// 	Rectangle

// 	Line
// 	Text
// }

type BaseElement struct {
	Id          string  `json:"id" jsonschema:"description=Уникальный идентификатор элемента"`
	Type        string  `json:"type" jsonschema:"description=Тип элемента: rectangle, text, ellipse, line"`
	X           float32 `json:"x" jsonschema:"description=Координата X левого верхнего угла"`
	Y           float32 `json:"y" jsonschema:"description=Координата Y левого верхнего угла"`
	Width       float32 `json:"width" jsonschema:"description=Ширина элемента"`
	Height      float32 `json:"height" jsonschema:"description=Высота элемента"`
	Rotation    float32 `json:"rotation" jsonschema:"description=Угол поворота в градусах"`
	Fill        string  `json:"fill,omitempty" jsonschema:"description=Цвет заливки в формате hex"`
	Stroke      string  `json:"stroke,omitempty" jsonschema:"description=Цвет обводки в формате hex"`
	StrokeWidth int     `json:"strokeWidth,omitempty" jsonschema:"description=Толщина обводки в пикселях"`
}

type Element struct {
	Id          string  `json:"id" jsonschema:"description=Уникальный идентификатор элемента"`
	Type        string  `json:"type" jsonschema:"description=Тип элемента: rectangle, text, ellipse, line"`
	X           float32 `json:"x" jsonschema:"description=Координата X левого верхнего угла"`
	Y           float32 `json:"y" jsonschema:"description=Координата Y левого верхнего угла"`
	Width       float32 `json:"width" jsonschema:"description=Ширина элемента"`
	Height      float32 `json:"height" jsonschema:"description=Высота элемента"`
	Rotation    float32 `json:"rotation" jsonschema:"description=Угол поворота в градусах"`
	Fill        string  `json:"fill,omitempty" jsonschema:"description=Цвет заливки в формате hex"`
	Stroke      string  `json:"stroke,omitempty" jsonschema:"description=Цвет обводки в формате hex"`
	StrokeWidth int     `json:"strokeWidth,omitempty" jsonschema:"description=Толщина обводки в пикселях"`

	// inserted from rectangle model
	CornerRadius int `json:"cornerRadius,omitempty" jsonschema:"description=Радиус скругления углов для прямоугольника"`

	// inserted from text model
	Content string `json:"content,omitempty" jsonschema:"description=Текстовое содержимое элемента"`

	Points  []float32 `json:"points,omitempty" jsonschema:"description=Массив точек для линий в формате [x1,y1,x2,y2,...]"`
	Tension float32   `json:"tension,omitempty" jsonschema:"description=Давление/натяжение для кривых линий"`
}

type Board struct {
	BoardID       string         `json:"boardId" jsonschema:"description=Уникальный идентификатор доски"`
	ImageURL      string         `json:"imageUrl,omitempty" jsonschema:"description=URL изображения доски"`
	Elements      []Element      `json:"elements,omitempty" jsonschema:"description=Элементы доски (прямоугольники, текст, линии)"`
	GraphNodes    []GNode        `json:"graphNodes,omitempty" jsonschema:"description=Узлы графа для React Flow досок"`
	GraphEdges    []GEdge        `json:"graphEdges,omitempty" jsonschema:"description=Рёбра графа для React Flow досок"`
	SemanticGraph *SemanticGraph `json:"-"`
}

type AnalyzeRequest struct {
	RequestType          string                      `json:"requestType"`
	SummarizeRequest     SummarizeRequest
	StructurizeRequest   StructurizeRequest
	TemplateRequest      GenerateTemplateRequest
	IncrementalRequest   IncrementalAnalysisRequest
}

func NewSumAnalyzeReq(req SummarizeRequest) AnalyzeRequest {
	return AnalyzeRequest{RequestType: SummarizeType, SummarizeRequest: req}
}

func NewStructAnalyzeReq(req StructurizeRequest) AnalyzeRequest {
	return AnalyzeRequest{RequestType: StructurizeType, StructurizeRequest: req}
}

func NewTemplateAnalyzeReq(req GenerateTemplateRequest) AnalyzeRequest {
	return AnalyzeRequest{RequestType: GenerateTemplateType, TemplateRequest: req}
}

func NewIncrementalAnalyzeReq(req IncrementalAnalysisRequest) AnalyzeRequest {
	return AnalyzeRequest{RequestType: IncrementalType, IncrementalRequest: req}
}

type AnalyzeResponse struct {
	SummarizeResponse     SummarizeResponse
	StructurizeResponse   StructurizeResponse
	TemplateResponse      GenerateTemplateResponse
	IncrementalResponse   IncrementalAnalysisResponse
}

type SummarizeRequest struct {
	RequestID   string `json:"requestId,omitempty" jsonschema:"description=Уникальный идентификатор запроса"`
	UserID      string `json:"userId,omitempty" jsonschema:"description=Уникальный идентификатор пользователя"`
	RequestType string `json:"requestType" jsonschema:"description=Тип запроса: summarize"`
	Board       Board  `json:"board" jsonschema:"description=Данные доски для анализа"`
}

type SummarizeResponse struct {
	RequestID   string `json:"requestId" jsonschema:"description=Уникальный идентификатор запроса"`
	UserID      string `json:"userId" jsonschema:"description=Уникальный идентификатор пользователя"`
	RequestType string `json:"requestType" jsonschema:"description=Тип запроса: summarize"`
	Element     Text   `json:"text" jsonschema:"description=Сгенерированный текст суммаризации"`
}

type StructurizeRequest struct {
	RequestID   string `json:"requestId" jsonschema:"description=Уникальный идентификатор запроса"`
	UserID      string `json:"userId" jsonschema:"description=Уникальный идентификатор пользователя"`
	RequestType string `json:"requestType" jsonschema:"description=Тип запроса: structurize"`
	Board       Board  `json:"board" jsonschema:"description=Данные доски для анализа"`
	File        File   `json:"file" jsonschema:"description=Исходная файловая структура"`
}

type StructurizeResponse struct {
	RequestID      string `json:"requestId" jsonschema:"description=Уникальный идентификатор запроса"`
	UserID         string `json:"userId" jsonschema:"description=Уникальный идентификатор пользователя"`
	RequestType    string `json:"requestType" jsonschema:"description=Тип запроса: structurize"`
	AiTreeResponse string `json:"aiTreeResponse" jsonschema:"description=ASCII представление файловой структуры"`
	File           File   `json:"file" jsonschema:"description=Сгенерированная файловая структура"`
}

type File struct {
	Name     string `json:"name" jsonschema:"description=Имя файла или папки"`
	Type     string `json:"type" jsonschema:"description=Тип: doc (файл), section (папка)"`
	Children []File `json:"children" jsonschema:"description=Дочерние элементы (для папок)"`
}

func (f File) IsEmpty() bool {
	return f.Name == "" && f.Type == ""
}

// IsPopulated checks if the file has meaningful content
func (f File) IsPopulated() bool {
	return f.Name != "" || f.Type != "" || len(f.Children) > 0
}

type Abort struct {
	RequestID string `json:"requestId"`
}

// Job represents a unit of work in the system.
type Job struct {
	ID        string
	Request   AnalyzeRequest
	CreatedAt int64
	Retries   int
	Status    JobStatus
}

type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
	JobStatusAborted   JobStatus = "aborted"
)

type TranscriptionResult struct {
	Text     string
	Language string
	Metadata map[string]any
}

// GenerateTemplateRequest represents a request to generate a board template
type GenerateTemplateRequest struct {
	RequestID   string    `json:"requestId" jsonschema:"description=Уникальный идентификатор запроса"`
	UserID      string    `json:"userId" jsonschema:"description=Уникальный идентификатор пользователя"`
	RequestType string    `json:"requestType" jsonschema:"description=Тип запроса: generateTemplate"`
	BoardID     string    `json:"boardId" jsonschema:"description=Уникальный идентификатор доски"`
	Prompt      string    `json:"prompt" jsonschema:"description=Текстовый промпт для генерации шаблона"`
	BoardType   BoardType `json:"boardType" jsonschema:"description=Тип доски: simple или graph"`
}

// Validate validates the generate template request
func (r *GenerateTemplateRequest) Validate() error {
	if r.RequestID == "" {
		return fmt.Errorf("requestId is required")
	}

	if r.UserID == "" {
		return fmt.Errorf("userId is required")
	}

	if r.RequestType != GenerateTemplateType {
		return fmt.Errorf("requestType must be 'generateTemplate'")
	}

	if r.BoardID == "" {
		return fmt.Errorf("boardId is required")
	}

	if r.Prompt == "" {
		return fmt.Errorf("prompt is required")
	}

	if len(r.Prompt) < 10 {
		return fmt.Errorf("prompt must be at least 10 characters")
	}

	if len(r.Prompt) > 2000 {
		return fmt.Errorf("prompt must be at most 2000 characters")
	}

	if r.BoardType == "" {
		return fmt.Errorf("boardType is required")
	}

	if r.BoardType != BoardTypeSimple && r.BoardType != BoardTypeGraph {
		return fmt.Errorf("boardType must be 'simple' or 'graph'")
	}

	return nil
}

// GenerateTemplateResponse represents the response from template generation
type GenerateTemplateResponse struct {
	RequestID   string `json:"requestId" jsonschema:"description=Уникальный идентификатор запроса"`
	UserID      string `json:"userId" jsonschema:"description=Уникальный идентификатор пользователя"`
	RequestType string `json:"requestType" jsonschema:"description=Тип запроса: generateTemplate"`
	Board       Board  `json:"board" jsonschema:"description=Сгенерированная доска с элементами"`
}
