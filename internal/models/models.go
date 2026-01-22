package models

// CanvasMeta describes canvas physical dimensions (pixels).
type CanvasMeta struct {
	Width  int     `json:"width"`
	Height int     `json:"height"`
	DPI    *int    `json:"dpi,omitempty"`
	Unit   *string `json:"unit,omitempty"`
}

// Point used in strokes. Coordinates normalized [0..1].
type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	T int64   `json:"t,omitempty"`
}

// Stroke: unified type for handwritten/ink input (coordinates normalized).
type Stroke struct {
	ID       string    `json:"strokeId"`
	Points   []Point   `json:"points"`
	Pressure []float32 `json:"pressure,omitempty"`
}

// Rectangle: normalized bounding rectangle (x,y top-left, width,height).
type Rectangle struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

// File: hierarchical user file (HTML/text etc.).
type File struct {
	Name     string `json:"name"`
	Content  string `json:"content"`
	Children []File `json:"children,omitempty"`
}

// Element: minimal unified board element.
// Supported types include: rect, ellipse, text, line, marker, image, file, strokes, shape
type Element struct {
	ID        string     `json:"id"`
	Type      string     `json:"type"`
	Rectangle *Rectangle `json:"rectangle,omitempty"`
	Text      string     `json:"text,omitempty"`
	StrokeIDs []string   `json:"strokeIds,omitempty"`
	FileURL   string     `json:"fileUrl,omitempty"`
}

// Relationship: user-provided or inferred relations between elements.
type Relationship struct {
	From string `json:"from"`
	To   string `json:"to"`
	Type string `json:"type"` // arrow|connector|group|parent|link
}

// AnalyzeRequest: unified request for structure|graph|complex analyzes.
type AnalyzeRequest struct {
	RequestID     string         `json:"requestId,omitempty"`
	RequestType   string         `json:"requestType"` // structure|graph|complex
	BoardID       string         `json:"boardId,omitempty"`
	UserID        string         `json:"userId,omitempty"`
	UserPrompt    string         `json:"userPrompt,omitempty"`
	Canvas        CanvasMeta     `json:"canvas"`
	Elements      []Element      `json:"elements,omitempty"`
	Relationships []Relationship `json:"relationships,omitempty"`
	Strokes       []Stroke       `json:"strokes,omitempty"`
	Files         []File         `json:"files,omitempty"`
	ImageURL      string         `json:"imageUrl,omitempty"`
	Graph         string         `json:"graph,omitempty"` // optional graph payload (json/string)
}

// AnalyzeResponse: unified response from pipeline/LLM.
type AnalyzeResponse struct {
	RequestID     string         `json:"requestId,omitempty"`
	LlmAnswer     string         `json:"llmAnswer,omitempty"`
	Elements      []Element      `json:"elements,omitempty"`
	Relationships []Relationship `json:"relationships,omitempty"`
	Files         []File         `json:"files,omitempty"`
	Graph         string         `json:"graph,omitempty"`
}

// Action represents a single action in the system.
type Action struct {
	Type    string         `json:"type"`
	Payload map[string]any `json:"payload,omitempty"`
	Params  map[string]any `json:"params,omitempty"`
}

// AcceptedResponse represents the response when a job is accepted.
type AcceptedResponse struct {
	JobID     string `json:"job_id"`
	Status    string `json:"status"`
	CreatedAt int64  `json:"created_at"`
	ExpiresAt int64  `json:"expires_at"`
}

func (a AcceptedResponse) Error() string {
	return "job accepted with ID: " + a.JobID
}

// Job represents a unit of work in the system.
type Job struct {
	ID        string
	Request   AnalyzeRequest
	CreatedAt int64
	Retries   int
	Status    JobStatus
}

// JobStatus represents the status of a job.
type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
)

// TranscriptionResult represents the result of a transcription.
type TranscriptionResult struct {
	Text     string
	Language string
	Metadata map[string]any
}
https://habr.com/ru/companies/ruvds/articles/985050/