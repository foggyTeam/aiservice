package preprocessing

import (
	"testing"

	"github.com/aiservice/internal/models"
	"github.com/stretchr/testify/assert"
)

// generateCatDrawingTestData creates simulated data from 10 people collaborating on drawing a cat
func generateCatDrawingTestData() models.Board {
	board := models.Board{
		BoardID:  "cat-drawing-session-001",
		ImageURL: "https://example.com/cat-drawing-board.jpg",
		Elements: []models.Element{},
	}

	// Person 1: Initial cat outline
	board.Elements = append(board.Elements, models.Element{
		Id:      "person1-outline",
		Type:    "line",
		X:       100.0,
		Y:       100.0,
		Width:   300.0,
		Height:  200.0,
		Content: "",
		Points:  []float32{100, 100, 200, 80, 300, 100, 350, 150, 300, 200, 200, 220, 100, 200, 80, 150, 100, 100},
		Stroke:  "#000000",
	})

	board.Elements = append(board.Elements, models.Element{
		Id:      "person1-title",
		Type:    "text",
		X:       150.0,
		Y:       50.0,
		Width:   200.0,
		Height:  30.0,
		Content: "Initial Cat Outline",
		Fill:    "#333333",
	})

	// Person 2: Eye suggestions
	board.Elements = append(board.Elements, models.Element{
		Id:          "person2-eyes",
		Type:        "ellipse",
		X:           180.0,
		Y:           130.0,
		Width:       20.0,
		Height:      20.0,
		Stroke:      "#FF0000",
		StrokeWidth: 2,
	})

	board.Elements = append(board.Elements, models.Element{
		Id:      "person2-eye-label",
		Type:    "text",
		X:       160.0,
		Y:       160.0,
		Width:   100.0,
		Height:  20.0,
		Content: "Big round eyes?",
	})

	// Person 3: Alternative eye style
	board.Elements = append(board.Elements, models.Element{
		Id:          "person3-eyes-alt",
		Type:        "ellipse",
		X:           220.0,
		Y:           130.0,
		Width:       15.0,
		Height:      25.0,
		Stroke:      "#00AA00",
		StrokeWidth: 2,
	})

	board.Elements = append(board.Elements, models.Element{
		Id:      "person3-eye-label",
		Type:    "text",
		X:       200.0,
		Y:       160.0,
		Width:   100.0,
		Height:  20.0,
		Content: "Almond shaped?",
	})

	// Person 4: Nose and mouth
	board.Elements = append(board.Elements, models.Element{
		Id:     "person4-nose",
		Type:   "line",
		X:      200.0,
		Y:      170.0,
		Width:  0.0,
		Height: 0.0,
		Points: []float32{200, 170, 210, 180, 200, 180},
		Stroke: "#FF0000",
	})

	board.Elements = append(board.Elements, models.Element{
		Id:     "person4-mouth",
		Type:   "line",
		X:      190.0,
		Y:      180.0,
		Width:  0.0,
		Height: 0.0,
		Points: []float32{190, 180, 200, 190, 210, 180},
		Stroke: "#FF0000",
	})

	board.Elements = append(board.Elements, models.Element{
		Id:      "person4-comment",
		Type:    "text",
		X:       180.0,
		Y:       200.0,
		Width:   150.0,
		Height:  40.0,
		Content: "Triangular nose,\nwavy mouth?",
	})

	// Person 5: Ear suggestions
	board.Elements = append(board.Elements, models.Element{
		Id:     "person5-ear1",
		Type:   "line",
		X:      160.0,
		Y:      110.0,
		Width:  0.0,
		Height: 0.0,
		Points: []float32{160, 110, 150, 80, 170, 90, 160, 110},
		Stroke: "#0000FF",
	})

	board.Elements = append(board.Elements, models.Element{
		Id:     "person5-ear2",
		Type:   "line",
		X:      240.0,
		Y:      110.0,
		Width:  0.0,
		Height: 0.0,
		Points: []float32{240, 110, 250, 80, 230, 90, 240, 110},
		Stroke: "#0000FF",
	})

	board.Elements = append(board.Elements, models.Element{
		Id:      "person5-ear-comment",
		Type:    "text",
		X:       140.0,
		Y:       70.0,
		Width:   120.0,
		Height:  20.0,
		Content: "Pointy ears?",
	})

	// Person 6: Whiskers
	board.Elements = append(board.Elements, models.Element{
		Id:     "person6-whiskers",
		Type:   "line",
		X:      190.0,
		Y:      175.0,
		Width:  0.0,
		Height: 0.0,
		Points: []float32{
			190, 175, 160, 170, // Left whiskers
			190, 175, 160, 175,
			190, 175, 160, 180,
			210, 175, 240, 170, // Right whiskers
			210, 175, 240, 175,
			210, 175, 240, 180,
		},
		Stroke: "#888888",
	})

	board.Elements = append(board.Elements, models.Element{
		Id:      "person6-whisk-comment",
		Type:    "text",
		X:       150.0,
		Y:       150.0,
		Width:   100.0,
		Height:  20.0,
		Content: "Whiskers!",
	})

	// Person 7: Tail suggestions
	board.Elements = append(board.Elements, models.Element{
		Id:          "person7-tail",
		Type:        "line",
		X:           350.0,
		Y:           150.0,
		Width:       0.0,
		Height:      0.0,
		Points:      []float32{350, 150, 400, 120, 450, 140, 420, 180, 380, 170},
		Stroke:      "#FFA500",
		StrokeWidth: 3,
	})

	board.Elements = append(board.Elements, models.Element{
		Id:      "person7-tail-comment",
		Type:    "text",
		X:       400.0,
		Y:       100.0,
		Width:   100.0,
		Height:  30.0,
		Content: "Curvy tail\nlike this?",
	})

	// Person 8: Alternative tail
	board.Elements = append(board.Elements, models.Element{
		Id:          "person8-tail",
		Type:        "line",
		X:           350.0,
		Y:           160.0,
		Width:       0.0,
		Height:      0.0,
		Points:      []float32{350, 160, 420, 160, 450, 130, 470, 160},
		Stroke:      "#008000",
		StrokeWidth: 3,
	})

	board.Elements = append(board.Elements, models.Element{
		Id:      "person8-tail-comment",
		Type:    "text",
		X:       420.0,
		Y:       180.0,
		Width:   100.0,
		Height:  30.0,
		Content: "Thick & curly\ntail?",
	})

	// Person 9: Paws
	board.Elements = append(board.Elements, models.Element{
		Id:     "person9-paw1",
		Type:   "ellipse",
		X:      150.0,
		Y:      200.0,
		Width:  25.0,
		Height: 15.0,
		Stroke: "#000000",
		Fill:   "#FFFFFF",
	})

	board.Elements = append(board.Elements, models.Element{
		Id:     "person9-paw2",
		Type:   "ellipse",
		X:      200.0,
		Y:      200.0,
		Width:  25.0,
		Height: 15.0,
		Stroke: "#000000",
		Fill:   "#FFFFFF",
	})

	board.Elements = append(board.Elements, models.Element{
		Id:     "person9-paw3",
		Type:   "ellipse",
		X:      250.0,
		Y:      200.0,
		Width:  25.0,
		Height: 15.0,
		Stroke: "#000000",
		Fill:   "#FFFFFF",
	})

	board.Elements = append(board.Elements, models.Element{
		Id:     "person9-paw4",
		Type:   "ellipse",
		X:      300.0,
		Y:      200.0,
		Width:  25.0,
		Height: 15.0,
		Stroke: "#000000",
		Fill:   "#FFFFFF",
	})

	board.Elements = append(board.Elements, models.Element{
		Id:      "person9-paw-comment",
		Type:    "text",
		X:       180.0,
		Y:       220.0,
		Width:   100.0,
		Height:  20.0,
		Content: "Paws!",
	})

	// Person 10: Final comment and decision box
	board.Elements = append(board.Elements, models.Element{
		Id:           "person10-decision-box",
		Type:         "rect",
		X:            50.0,
		Y:            300.0,
		Width:        400.0,
		Height:       100.0,
		Fill:         "#FFFFCC",
		Stroke:       "#333333",
		StrokeWidth:  2,
		CornerRadius: 10,
	})

	board.Elements = append(board.Elements, models.Element{
		Id:      "person10-decision-text",
		Type:    "text",
		X:       70.0,
		Y:       320.0,
		Width:   360.0,
		Height:  60.0,
		Content: "DECISION:\nRound eyes ✓\nTriangular nose ✓\nCurvy tail ✓\nAll paws ✓\nWhiskers ✓",
		Fill:    "#000000",
	})

	return board
}

func TestProcessingCatDrawingData(t *testing.T) {
	board := generateCatDrawingTestData()

	preprocessor := NewPreprocessor()

	req := models.SummarizeRequest{
		Board: board,
	}

	parts, err := preprocessor.PreprocessSummarizeRequest(req)
	assert.NoError(t, err)
	assert.NotEmpty(t, parts)

	// Verify that the output contains expected sections
	output := parts[0].Text
	assert.Contains(t, output, "RAW DATA:")
	assert.Contains(t, output, "SPATIAL ANALYSIS:")
	assert.Contains(t, output, "SEMANTIC ANNOTATIONS:")
	assert.Contains(t, output, "cat")
	assert.Contains(t, output, "eye")
	assert.Contains(t, output, "tail")
	assert.Contains(t, output, "paw")
	assert.Contains(t, output, "whisker")
}
