package utils

import (
	"math"

	"github.com/aiservice/internal/models"
)

const (
	// Default text element dimensions
	DefaultTextWidth  float32 = 400
	DefaultTextHeight float32 = 100

	// Margin around elements
	ElementMargin float32 = 50

	// Board boundaries
	BoardMinX float32 = 0
	BoardMinY float32 = 0
	BoardMaxX float32 = 3000
	BoardMaxY float32 = 3000
)

// TextPosition represents the calculated position and size for a text element
type TextPosition struct {
	X      float32
	Y      float32
	Width  float32
	Height float32
}

func CalculateTextSize(content string) (width, height float32) {
	charsPerLine := 50
	lineHeight := float32(20.0)

	numLines := int(math.Ceil(float64(len(content)) / float64(charsPerLine)))
	if numLines < 1 {
		numLines = 1
	}

	width = float32(charsPerLine * 8)
	height = float32(numLines) * lineHeight

	width += 40
	height += 20

	return width, height
}

// FindFreeSpace finds a free space on the board for placing a text element
func FindFreeSpace(elements []models.Element, textWidth, textHeight float32) TextPosition {
	// Create a grid of potential positions
	gridSize := float32(50.0)

	// Try positions from top-left to bottom-right
	for y := BoardMinY + ElementMargin; y < BoardMaxY-textHeight-ElementMargin; y += gridSize {
		for x := BoardMinX + ElementMargin; x < BoardMaxX-textWidth-ElementMargin; x += gridSize {
			if !isOverlapping(x, y, textWidth, textHeight, elements) {
				return TextPosition{
					X:      x,
					Y:      y,
					Width:  textWidth,
					Height: textHeight,
				}
			}
		}
	}

	// If no free space found, return default position at bottom
	return TextPosition{
		X:      BoardMinX + ElementMargin,
		Y:      BoardMaxY - textHeight - ElementMargin,
		Width:  textWidth,
		Height: textHeight,
	}
}

// isOverlapping checks if a rectangle overlaps with any existing element
func isOverlapping(x, y, width, height float32, elements []models.Element) bool {
	newRect := rectangle{x, y, width, height}

	for _, elem := range elements {
		// Skip line elements (they're thin and shouldn't block placement)
		if elem.Type == "line" {
			continue
		}

		elemRect := rectangle{
			x:      elem.X,
			y:      elem.Y,
			width:  elem.Width,
			height: elem.Height,
		}

		if newRect.overlaps(elemRect) {
			return true
		}
	}

	return false
}

// rectangle represents a rectangular area
type rectangle struct {
	x      float32
	y      float32
	width  float32
	height float32
}

// overlaps checks if two rectangles overlap
func (r1 rectangle) overlaps(r2 rectangle) bool {
	// Add margin to avoid placing too close
	margin := float32(20.0)

	return !(r1.x+r1.width+margin < r2.x ||
		r2.x+r2.width+margin < r1.x ||
		r1.y+r1.height+margin < r2.y ||
		r2.y+r2.height+margin < r1.y)
}

// CalculateTextPosition calculates the optimal position for a text element
func CalculateTextPosition(elements []models.Element, content string) TextPosition {
	// Calculate required size based on content
	width, height := CalculateTextSize(content)

	// Find free space on the board
	return FindFreeSpace(elements, width, height)
}
