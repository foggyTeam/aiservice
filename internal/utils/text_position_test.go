package utils

import (
	"strings"
	"testing"

	"github.com/aiservice/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestCalculateTextSize(t *testing.T) {
	tests := []struct {
		name         string
		content      string
		minWidth     float32
		minHeight    float32
		maxWidth     float32
		maxHeight    float32
	}{
		{
			name:      "short text",
			content:   "Short summary",
			minWidth:  40,
			minHeight: 20,
			maxWidth:  500,
			maxHeight: 100,
		},
		{
			name:      "medium text",
			content:   strings.Repeat("This is a medium length summary text. ", 5),
			minWidth:  40,
			minHeight: 20,
			maxWidth:  600,
			maxHeight: 300,
		},
		{
			name:      "long text",
			content:   strings.Repeat("This is a long summary text that should require multiple lines and more height. ", 20),
			minWidth:  40,
			minHeight: 20,
			maxWidth:  600,
			maxHeight: 1000,
		},
		{
			name:      "empty text",
			content:   "",
			minWidth:  40,
			minHeight: 20,
			maxWidth:  500,
			maxHeight: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			width, height := CalculateTextSize(tt.content)

			assert.GreaterOrEqual(t, width, tt.minWidth, "width should be at least minimum")
			assert.GreaterOrEqual(t, height, tt.minHeight, "height should be at least minimum")
			assert.LessOrEqual(t, width, tt.maxWidth, "width should not exceed maximum")
			assert.LessOrEqual(t, height, tt.maxHeight, "height should not exceed maximum")
		})
	}
}

func TestCalculateTextSize_IncreasesWithContent(t *testing.T) {
	shortText := "Short"
	longText := strings.Repeat("This is a much longer text that should require more space. ", 10)

	_, shortHeight := CalculateTextSize(shortText)
	_, longHeight := CalculateTextSize(longText)

	// Height should increase with more lines
	assert.Less(t, shortHeight, longHeight, "longer text should have greater height")
	// Width may be the same (fixed per line), but height should differ
	assert.Greater(t, longHeight, shortHeight, "much longer text should have much greater height")
}

func TestFindFreeSpace_NoOverlap(t *testing.T) {
	elements := []models.Element{
		{
			Id:     "elem1",
			Type:   "rect",
			X:      100,
			Y:      100,
			Width:  200,
			Height: 150,
		},
		{
			Id:     "elem2",
			Type:   "text",
			X:      400,
			Y:      300,
			Width:  150,
			Height: 100,
		},
	}

	textWidth := float32(400)
	textHeight := float32(100)

	position := FindFreeSpace(elements, textWidth, textHeight)

	// Verify the position doesn't overlap with existing elements
	assert.False(t, overlapsWithAny(position.X, position.Y, textWidth, textHeight, elements),
		"found position should not overlap with existing elements")
}

func TestFindFreeSpace_EmptyBoard(t *testing.T) {
	elements := []models.Element{}

	textWidth := float32(400)
	textHeight := float32(100)

	position := FindFreeSpace(elements, textWidth, textHeight)

	// Should return a valid position within board boundaries
	assert.GreaterOrEqual(t, position.X, float32(0), "X should be within board")
	assert.GreaterOrEqual(t, position.Y, float32(0), "Y should be within board")
	assert.LessOrEqual(t, position.X+position.Width, BoardMaxX, "should fit within board width")
	assert.LessOrEqual(t, position.Y+position.Height, BoardMaxY, "should fit within board height")
}

func TestFindFreeSpace_CrowdedBoard(t *testing.T) {
	// Create a crowded board with many elements
	elements := []models.Element{
		{Id: "e1", Type: "rect", X: 50, Y: 50, Width: 200, Height: 100},
		{Id: "e2", Type: "rect", X: 300, Y: 50, Width: 200, Height: 100},
		{Id: "e3", Type: "rect", X: 550, Y: 50, Width: 200, Height: 100},
		{Id: "e4", Type: "rect", X: 50, Y: 200, Width: 200, Height: 100},
		{Id: "e5", Type: "rect", X: 300, Y: 200, Width: 200, Height: 100},
		{Id: "e6", Type: "rect", X: 550, Y: 200, Width: 200, Height: 100},
	}

	textWidth := float32(300)
	textHeight := float32(80)

	position := FindFreeSpace(elements, textWidth, textHeight)

	// Should still find a position (might be at the bottom)
	assert.GreaterOrEqual(t, position.X, float32(0), "X should be valid")
	assert.GreaterOrEqual(t, position.Y, float32(0), "Y should be valid")
	assert.False(t, overlapsWithAny(position.X, position.Y, textWidth, textHeight, elements),
		"found position should not overlap")
}

func TestFindFreeSpace_SkipsLineElements(t *testing.T) {
	// Line elements should not block placement
	elements := []models.Element{
		{
			Id:     "line1",
			Type:   "line",
			X:      0,
			Y:      0,
			Width:  1000,
			Height: 5,
		},
	}

	textWidth := float32(400)
	textHeight := float32(100)

	position := FindFreeSpace(elements, textWidth, textHeight)

	// Should find space despite the line
	assert.GreaterOrEqual(t, position.X, float32(0), "X should be valid")
	assert.GreaterOrEqual(t, position.Y, float32(0), "Y should be valid")
}

func TestCalculateTextPosition(t *testing.T) {
	elements := []models.Element{
		{
			Id:     "elem1",
			Type:   "rect",
			X:      100,
			Y:      100,
			Width:  200,
			Height: 150,
		},
	}

	content := "This is a test summary that needs proper positioning on the board."

	position := CalculateTextPosition(elements, content)

	// Verify size is calculated based on content
	width, height := CalculateTextSize(content)
	assert.InDelta(t, position.Width, width, 1, "width should match calculated size")
	assert.InDelta(t, position.Height, height, 1, "height should match calculated size")

	// Verify position doesn't overlap
	assert.False(t, overlapsWithAny(position.X, position.Y, position.Width, position.Height, elements),
		"position should not overlap with existing elements")
}

func TestRectangle_Overlaps(t *testing.T) {
	tests := []struct {
		name     string
		r1       rectangle
		r2       rectangle
		overlaps bool
	}{
		{
			name:     "no overlap - r1 left of r2",
			r1:       rectangle{0, 0, 100, 100},
			r2:       rectangle{150, 0, 100, 100},
			overlaps: false,
		},
		{
			name:     "no overlap - r1 above r2",
			r1:       rectangle{0, 0, 100, 100},
			r2:       rectangle{0, 150, 100, 100},
			overlaps: false,
		},
		{
			name:     "overlap - partial",
			r1:       rectangle{0, 0, 100, 100},
			r2:       rectangle{50, 50, 100, 100},
			overlaps: true,
		},
		{
			name:     "overlap - r1 contains r2",
			r1:       rectangle{0, 0, 200, 200},
			r2:       rectangle{50, 50, 50, 50},
			overlaps: true,
		},
		{
			name:     "no overlap - separated by margin",
			r1:       rectangle{0, 0, 100, 100},
			r2:       rectangle{125, 0, 100, 100}, // 25px gap for margin
			overlaps: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.r1.overlaps(tt.r2)
			assert.Equal(t, tt.overlaps, result)
		})
	}
}

// Helper function to check if a rectangle overlaps with any element
func overlapsWithAny(x, y, width, height float32, elements []models.Element) bool {
	testRect := rectangle{x, y, width, height}

	for _, elem := range elements {
		if elem.Type == "line" {
			continue
		}

		elemRect := rectangle{
			x:      elem.X,
			y:      elem.Y,
			width:  elem.Width,
			height: elem.Height,
		}

		if testRect.overlaps(elemRect) {
			return true
		}
	}

	return false
}
