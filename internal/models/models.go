package models

const (
	SummarizeType   = "summarize"
	StructurizeType = "structurize"
)

const (
	RectangeType = "rectangle"
	TextType     = "text"
	EllipseType  = "ellipse"
	LineTypeType = "line"
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
	BoardID  string    `json:"boardId"`
	ImageURL string    `json:"imageUrl,omitempty"`
	Elements []Element `json:"elements"`
}

type AnalyzeRequest struct {
	RequestType        string `json:"requestType"` // summarize, structurize
	SummarizeRequest   SummarizeRequest
	StructurizeRequest StructurizeRequest
}

func NewSumAnalyzeReq(req SummarizeRequest) AnalyzeRequest {
	return AnalyzeRequest{RequestType: SummarizeType, SummarizeRequest: req}
}

func NewStructAnalyzeReq(req StructurizeRequest) AnalyzeRequest {
	return AnalyzeRequest{RequestType: StructurizeType, StructurizeRequest: req}
}

type AnalyzeResponse struct {
	SummarizeResponse   SummarizeResponse
	StructurizeResponse StructurizeResponse
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
	Name     string `json:"name"`
	Type     string `json:"type"` //doc, simple, graph,(поле children пустое) | section (содердит детей)
	Children []File `json:"children"`
}

func (f File) IsEmpty() bool {
	return f.Name == "" && f.Type == ""
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
