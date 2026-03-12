package models

import "fmt"

const (
	SummarizeType        = "summarize"
	StructurizeType      = "structurize"
	GenerateTemplateType = "generateTemplate"
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
	Content string `json:"content"`
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
	Id          string  `json:"id"`
	Type        string  `json:"type"` //rect, line, text, ellipse,
	X           float32 `json:"x"`
	Y           float32 `json:"y"`
	Width       float32 `json:"width"`
	Height      float32 `json:"height"`
	Rotation    float32 `json:"rotation"`
	Fill        string  `json:"fill,omitempty"`        // цвет заливки
	Stroke      string  `json:"stroke,omitempty"`      // цвет обводки
	StrokeWidth int     `json:"strokeWidth,omitempty"` // толщина обводки
}

type Element struct {
	Id          string  `json:"id"`
	Type        string  `json:"type"` //rect, line, text, ellipse,
	X           float32 `json:"x"`
	Y           float32 `json:"y"`
	Width       float32 `json:"width"`
	Height      float32 `json:"height"`
	Rotation    float32 `json:"rotation"`
	Fill        string  `json:"fill,omitempty"`        // цвет заливки
	Stroke      string  `json:"stroke,omitempty"`      // цвет обводки
	StrokeWidth int     `json:"strokeWidth,omitempty"` // толщина обводки

	// inserted from rectangle model
	CornerRadius int `json:"cornerRadius,omitempty"`

	// inserted from text model
	Content string `json:"content,omitempty"`

	Points  []float32 `json:"points,omitempty"`  // [x, y], [x, y]
	Tension float32   `json:"tension,omitempty"` // давление
}

type Board struct {
	BoardID       string          `json:"boardId"`
	ImageURL      string          `json:"imageUrl,omitempty"`
	Elements      []Element       `json:"elements,omitempty"`
	GraphNodes    []GNode         `json:"graphNodes,omitempty"`
	GraphEdges    []GEdge         `json:"graphEdges,omitempty"`
	SemanticGraph *SemanticGraph  `json:"-"`
}

type AnalyzeRequest struct {
	RequestType        string                 `json:"requestType"`
	SummarizeRequest   SummarizeRequest
	StructurizeRequest StructurizeRequest
	TemplateRequest    GenerateTemplateRequest
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

type AnalyzeResponse struct {
	SummarizeResponse   SummarizeResponse
	StructurizeResponse StructurizeResponse
	TemplateResponse    GenerateTemplateResponse
}

type SummarizeRequest struct {
	RequestID   string `json:"requestId,omitempty"`
	UserID      string `json:"userId,omitempty"`
	RequestType string `json:"requestType"` // summarize
	Board       Board  `json:"board"`
}
type SummarizeResponse struct {
	RequestID   string `json:"requestId"`
	UserID      string `json:"userId"`
	RequestType string `json:"requestType"` // summarize
	Element     Text   `json:"text"`        // конкретный элемент - текст, который суммаризовал инфу по доске, расположенный в свободном пространстве доски
}
type StructurizeRequest struct {
	RequestID   string `json:"requestId"`
	UserID      string `json:"userId"`
	RequestType string `json:"requestType"` // structurize
	Board       Board  `json:"board"`
	File        File   `json:"file"`
}
type StructurizeResponse struct {
	RequestID      string `json:"requestId"`
	UserID         string `json:"userId"`
	RequestType    string `json:"requestType"`    // structurize
	AiTreeResponse string `json:"aiTreeResponse"` // дерево ASCII файлов
	File           File   `json:"file"`
}

type File struct {
	Name     string `json:"name" example:"main.go"`
	Type     string `json:"type" example:"doc"` //doc, simple, graph,(поле children пустое) | section (содердит детей)
	Children []File `json:"children"`
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
	RequestID   string    `json:"requestId"`
	UserID      string    `json:"userId"`
	RequestType string    `json:"requestType"`
	BoardID     string    `json:"boardId"`
	Prompt      string    `json:"prompt"`
	BoardType   BoardType `json:"boardType"`
}

// Validate validates the generate template request
func (r *GenerateTemplateRequest) Validate() error {
	if r.RequestID == "" {
		return fmt.Errorf("requestId is required")
	}

	if r.UserID == "" {
		return fmt.Errorf("userId is required")
	}

	if r.RequestType == "" {
		return fmt.Errorf("requestType is required")
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
	RequestID   string `json:"requestId"`
	UserID      string `json:"userId"`
	RequestType string `json:"requestType"`
	Board       Board  `json:"board"`
}
