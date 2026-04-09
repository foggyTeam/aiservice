package models

import "time"

// BoardAnalysisCache stores the cached analysis result for a board
type BoardAnalysisCache struct {
	BoardID       string          `json:"boardId"`
	GlobalSummary string          `json:"globalSummary"`
	KeyConcepts   []string        `json:"keyConcepts"`
	Regions       []RegionSummary `json:"regions"`
	LastFullScan  time.Time       `json:"lastFullScan"`
	ChangeCount   int             `json:"changeCount"`
	ElementHashes map[string]string `json:"elementHashes"`
	Elements      []Element       `json:"elements"` // Сохраняем элементы для diff
	ImageURL      string          `json:"imageUrl"` // URL изображения доски
	BoardSize     BoundingBox     `json:"boardSize"` // Размер доски
}

// RegionSummary stores the analysis summary for a specific region
type RegionSummary struct {
	ID          string      `json:"id"`
	BBox        BoundingBox `json:"bbox"`
	Summary     string      `json:"summary"`
	ElementIDs  []string    `json:"elementIds"`
	LastUpdated time.Time   `json:"lastUpdated"`
}

// BoundingBox represents a rectangular area on the board
type BoundingBox struct {
	X float32 `json:"x"`
	Y float32 `json:"y"`
	W float32 `json:"w"`
	H float32 `json:"h"`
}

// Area calculates the area of the bounding box
func (b BoundingBox) Area() float64 {
	return float64(b.W) * float64(b.H)
}

// Intersects checks if two bounding boxes intersect
func (b BoundingBox) Intersects(other BoundingBox) bool {
	return b.X < other.X+other.W &&
		b.X+b.W > other.X &&
		b.Y < other.Y+other.H &&
		b.Y+b.H > other.Y
}

// Union returns the smallest bounding box that contains both boxes
func (b BoundingBox) Union(other BoundingBox) BoundingBox {
	x := min(b.X, other.X)
	y := min(b.Y, other.Y)
	w := max(b.X+b.W, other.X+other.W) - x
	h := max(b.Y+b.H, other.Y+other.H) - y
	return BoundingBox{X: x, Y: y, W: w, H: h}
}

// Contains checks if a point is inside the bounding box
func (b BoundingBox) Contains(x, y float32) bool {
	return x >= b.X && x <= b.X+b.W && y >= b.Y && y <= b.Y+b.H
}

// ElementChange represents a change to a single element
type ElementChange struct {
	ElementID  string     `json:"elementId"`
	ChangeType ChangeType `json:"changeType"`
	OldElement *Element   `json:"oldElement,omitempty"`
	NewElement *Element   `json:"newElement,omitempty"`
}

// ChangeType represents the type of change
type ChangeType string

const (
	ChangeAdded    ChangeType = "added"
	ChangeModified ChangeType = "modified"
	ChangeDeleted  ChangeType = "deleted"
)

// IncrementalAnalysisRequest represents a request for incremental analysis
type IncrementalAnalysisRequest struct {
	RequestID    string          `json:"requestId"`
	UserID       string          `json:"userId"`
	RequestType  string          `json:"requestType"`
	BoardID      string          `json:"boardId"`
	Changes      []ElementChange `json:"changes"`
	FullBoard    *Board          `json:"fullBoard,omitempty"`
	IsFullRescan bool            `json:"isFullRescan"`
}

// IncrementalAnalysisResponse represents the response from incremental analysis
type IncrementalAnalysisResponse struct {
	RequestID      string   `json:"requestId"`
	UserID         string   `json:"userId"`
	RequestType    string   `json:"requestType"`
	GlobalSummary  string   `json:"globalSummary"`
	KeyConcepts    []string `json:"keyConcepts"`
	UpdatedRegions []string `json:"updatedRegions"`
	IsFullRescan   bool     `json:"isFullRescan"`
}
