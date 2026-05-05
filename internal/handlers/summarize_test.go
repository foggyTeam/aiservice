package handlers

import (
	"fmt"
	"testing"

	"github.com/aiservice/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestValidateSummarizeRequest_Valid(t *testing.T) {
	t.Parallel()

	req := models.SummarizeRequest{
		RequestID: "req-123",
		UserID:    "user-456",
		Board: models.Board{
			BoardID: "board-789",
			Elements: []models.Element{
				{Id: "e1", Type: "text", Content: "hello", Width: 100, Height: 50},
				{Id: "e2", Type: "rectangle", Width: 200, Height: 100},
			},
		},
	}

	err := validateSummarizeRequest(req)
	assert.NoError(t, err)
}

func TestValidateSummarizeRequest_EmptyBoardID(t *testing.T) {
	t.Parallel()

	req := models.SummarizeRequest{
		UserID: "user-1",
		Board:  models.Board{BoardID: ""},
	}

	err := validateSummarizeRequest(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "boardID is empty")
}

func TestValidateSummarizeRequest_EmptyUserID(t *testing.T) {
	t.Parallel()

	req := models.SummarizeRequest{
		UserID: "",
		Board:  models.Board{BoardID: "board-1"},
	}

	err := validateSummarizeRequest(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "userID is empty")
}

func TestValidateSummarizeRequest_TooManyElements(t *testing.T) {
	t.Parallel()

	req := models.SummarizeRequest{
		UserID: "user-1",
		Board: models.Board{
			BoardID:  "board-1",
			Elements: makeElements(1001, nil), // > 1000
		},
	}

	err := validateSummarizeRequest(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "too many elements")
	assert.Contains(t, err.Error(), "1000")
}

func TestValidateSummarizeRequest_MaxElements_Allowed(t *testing.T) {
	t.Parallel()

	req := models.SummarizeRequest{
		UserID: "user-1",
		Board: models.Board{
			BoardID:  "board-1",
			Elements: makeElements(1000, nil),
		},
	}

	err := validateSummarizeRequest(req)
	assert.NoError(t, err)
}

func TestValidateSummarizeRequest_EmptyElements_Allowed(t *testing.T) {
	t.Parallel()

	req := models.SummarizeRequest{
		UserID: "user-1",
		Board: models.Board{
			BoardID:  "board-1",
			Elements: []models.Element{},
		},
	}

	err := validateSummarizeRequest(req)
	assert.NoError(t, err)
}

func TestValidateSummarizeRequest_Element_EmptyID(t *testing.T) {
	t.Parallel()

	req := models.SummarizeRequest{
		UserID: "user-1",
		Board: models.Board{
			BoardID: "board-1",
			Elements: []models.Element{
				{Id: "valid-id"},
				{Id: ""},
				{Id: "another-valid"},
			},
		},
	}

	err := validateSummarizeRequest(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "element ID cannot be empty")
}

func TestValidateSummarizeRequest_Element_NegativeWidth(t *testing.T) {
	t.Parallel()

	req := models.SummarizeRequest{
		UserID: "user-1",
		Board: models.Board{
			BoardID: "board-1",
			Elements: []models.Element{
				{Id: "e1", Width: -1, Height: 50},
			},
		},
	}

	err := validateSummarizeRequest(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "width and height must be non-negative")
}

func TestValidateSummarizeRequest_Element_NegativeHeight(t *testing.T) {
	t.Parallel()

	req := models.SummarizeRequest{
		UserID: "user-1",
		Board: models.Board{
			BoardID: "board-1",
			Elements: []models.Element{
				{Id: "e1", Width: 100, Height: -5},
			},
		},
	}

	err := validateSummarizeRequest(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "width and height must be non-negative")
}

func TestValidateSummarizeRequest_Element_ZeroDimensions_Allowed(t *testing.T) {
	t.Parallel()

	req := models.SummarizeRequest{
		UserID: "user-1",
		Board: models.Board{
			BoardID: "board-1",
			Elements: []models.Element{
				{Id: "e1", Width: 0, Height: 0},
			},
		},
	}

	err := validateSummarizeRequest(req)
	assert.NoError(t, err)
}

func TestValidateSummarizeRequest_TextContent_TooLong(t *testing.T) {
	t.Parallel()

	longText := string(make([]byte, 10001))

	req := models.SummarizeRequest{
		UserID: "user-1",
		Board: models.Board{
			BoardID: "board-1",
			Elements: []models.Element{
				{Id: "e1", Type: "text", Content: longText},
			},
		},
	}

	err := validateSummarizeRequest(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "text content too long")
	assert.Contains(t, err.Error(), "10000")
}

func TestValidateSummarizeRequest_TextContent_MaxAllowed(t *testing.T) {
	t.Parallel()

	exactMax := string(make([]byte, 10000))

	req := models.SummarizeRequest{
		UserID: "user-1",
		Board: models.Board{
			BoardID: "board-1",
			Elements: []models.Element{
				{Id: "e1", Type: "text", Content: exactMax},
			},
		},
	}

	err := validateSummarizeRequest(req)
	assert.NoError(t, err)
}

func TestValidateSummarizeRequest_NonText_LongContent_Allowed(t *testing.T) {
	t.Parallel()

	longContent := string(make([]byte, 15000))

	req := models.SummarizeRequest{
		UserID: "user-1",
		Board: models.Board{
			BoardID: "board-1",
			Elements: []models.Element{
				{Id: "e1", Type: "rectangle", Content: longContent},
				{Id: "e2", Type: "ellipse", Content: longContent},
				{Id: "e3", Type: "line", Content: longContent},
			},
		},
	}

	err := validateSummarizeRequest(req)
	assert.NoError(t, err, "не-текстовые элементы могут иметь длинный контент")
}

func TestValidateSummarizeRequest_MixedElements_OnlyTextValidated(t *testing.T) {
	t.Parallel()

	longContent := string(make([]byte, 15000))

	req := models.SummarizeRequest{
		UserID: "user-1",
		Board: models.Board{
			BoardID: "board-1",
			Elements: []models.Element{
				{Id: "e1", Type: "text", Content: "ok"},
				{Id: "e2", Type: "rectangle", Content: longContent},
				{Id: "e3", Type: "text", Content: longContent},
			},
		},
	}

	err := validateSummarizeRequest(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "text content too long")
}

func TestValidateSummarizeRequest_FirstErrorWins(t *testing.T) {
	t.Parallel()

	req := models.SummarizeRequest{
		UserID: "",
		Board: models.Board{
			BoardID: "",
			Elements: []models.Element{
				{Id: "", Width: -1},
			},
		},
	}

	err := validateSummarizeRequest(req)
	assert.Error(t, err)
	// Первая проверка - BoardID
	assert.Contains(t, err.Error(), "boardID is empty")
}

func TestValidateSummarizeRequest_UnicodeContent(t *testing.T) {
	t.Parallel()

	unicodeText := "Привет мир 🌍" + string(make([]byte, 9980))

	req := models.SummarizeRequest{
		UserID: "user-1",
		Board: models.Board{
			BoardID: "board-1",
			Elements: []models.Element{
				{Id: "e1", Type: "text", Content: unicodeText},
			},
		},
	}

	err := validateSummarizeRequest(req)
	assert.Error(t, err, "Unicode content within byte limit should be valid")
	assert.NotPanics(t, func() {
		_ = validateSummarizeRequest(req)
	})
}

func TestValidateSummarizeRequest_SpecialCharsInID(t *testing.T) {
	t.Parallel()

	req := models.SummarizeRequest{
		UserID: "user-1",
		Board: models.Board{
			BoardID: "board-1",
			Elements: []models.Element{
				{Id: "elem/with\\special:chars@2024"},
			},
		},
	}

	err := validateSummarizeRequest(req)
	assert.NoError(t, err, "специальные символы в ID допустимы")
}

func TestValidateSummarizeRequest_LargeButValidBoard(t *testing.T) {
	t.Parallel()

	elements := makeElements(1000, func(i int, e *models.Element) {
		if i%10 == 0 {
			e.Type = "text"
			e.Content = fmt.Sprintf("text-%d", i)
		}
	})

	req := models.SummarizeRequest{
		UserID: "user-1",
		Board: models.Board{
			BoardID:  "board-1",
			Elements: elements,
		},
	}

	err := validateSummarizeRequest(req)
	assert.NoError(t, err)
}
