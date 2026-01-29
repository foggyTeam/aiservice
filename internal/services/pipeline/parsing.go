package pipeline

import (
	"math"

	"github.com/aiservice/internal/models"
)

var TestData = []models.Element{
	// 1. Заголовок
	{
		Id:      "text-1",
		Type:    "text",
		X:       50,
		Y:       40,
		Content: "Анализ функции y = sin(x)",
	},

	// 2. Прямоугольник с заливкой и закруглением
	{
		Id:           "rect-1",
		Type:         "rect",
		X:            100,
		Y:            100,
		Width:        300,
		Height:       80,
		Fill:         "#fffacd", // лимонный шёлк
		Stroke:       "#000000",
		StrokeWidth:  2,
		CornerRadius: 12,
	},

	// 3. Текст формулы внутри прямоугольника
	{
		Id:      "text-2",
		Type:    "text",
		X:       120,
		Y:       130,
		Content: "f(x) = sin(x), x ∈ [0; 2π]",
	},

	// 4. Эллипс вокруг формулы
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

	// 5. Стрелка (линия с наконечником — имитируем через line)
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

	// 6. Рукописная кривая — график sin(x)
	{
		Id:          "hand-1",
		Type:        "line", // часто так помечают
		X:           0,
		Y:           0, // игнорируются, если есть Points
		Stroke:      "#008000",
		StrokeWidth: 2,
		Points: func() []float32 {
			var pts []float32
			for i := 0; i <= 50; i++ {
				t := float64(i) / 50.0 * 2 * math.Pi
				x := 100 + float32(t/math.Pi*100)  // от 100 до 300
				y := 300 - float32(math.Sin(t)*50) // колебания ±50 от 300
				pts = append(pts, x, y)
			}
			return pts
		}(),
	},

	// 7. Короткий рукописный штрих — галочка или каракуля
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

	// 8. Пустой элемент для проверки устойчивости
	{
		Id:   "empty-1",
		Type: "unknown",
		X:    0,
		Y:    0,
	},
}
