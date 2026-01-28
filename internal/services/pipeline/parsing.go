package pipeline

import (
	"math"

	"github.com/aiservice/internal/models"
)

// TODO
// После теста стало ясно, что не стоит отдельно парсить,
// можно передовать полные структуры и вроде ок работает

var TestData = []models.Element{
	{
		Id:      "text-1",
		Type:    "text",
		X:       50,
		Y:       40,
		Content: "Анализ функции y = sin(x)",
	},

	{
		Id:           "rect-1",
		Type:         "rect",
		X:            100,
		Y:            100,
		Width:        300,
		Height:       80,
		Fill:         "#fffacd", 
		Stroke:       "#000000",
		StrokeWidth:  2,
		CornerRadius: 12,
	},

	{
		Id:      "text-2",
		Type:    "text",
		X:       120,
		Y:       130,
		Content: "f(x) = sin(x), x ∈ [0; 2π]",
	},

	{
		Id:          "ellipse-1",
		Type:        "ellipse",
		X:           250, // центр
		Y:           140,
		Width:       100,
		Height:      40,
		Stroke:      "#ff0000",
		StrokeWidth: 2,
	},

	{
		Id:          "line-1",
		Type:        "line",
		X:           350,
		Y:           200,
		Width:       50,  // dx
		Height:      -30, // dy
		Stroke:      "#0000ff",
		StrokeWidth: 3,
	},

	{
		Id:          "hand-1",
		Type:        "line", 
		X:           0,
		Y:           0, // игнорируются, если есть Points
		Stroke:      "#008000",
		StrokeWidth: 2,
		Points: func() []float32 {
			var pts []float32
			for i := 0; i <= 50; i++ {
				t := float64(i) / 50.0 * 2 * math.Pi
				x := 100 + float32(t/math.Pi*100)  // от 100 до 300
				y := 300 - float32(math.Sin(t)*50) 
				pts = append(pts, x, y)
			}
			return pts
		}(),
	},

	{
		Id:          "hand-2",
		Type:        "line",
		Stroke:      "#8b0000",
		StrokeWidth: 3,
		Points: []float32{
			400, 400,
			410, 415,
			425, 395,
		},
	},

	{
		Id:   "empty-1",
		Type: "unknown",
		X:    0,
		Y:    0,
	},
}
