package handlers

import (
	"fmt"
	"strings"
	"testing"

	"github.com/aiservice/internal/models"
	"github.com/stretchr/testify/assert"
)

func makeFile(name, ftype string, children ...models.File) models.File {
	return models.File{
		Name:     name,
		Type:     ftype,
		Children: children,
	}
}

func makeNestedFile(depth int, namePrefix string) models.File {
	if depth <= 0 {
		return makeFile(fmt.Sprintf("%s-leaf", namePrefix), "doc")
	}
	return makeFile(
		fmt.Sprintf("%s-level-%d", namePrefix, depth),
		"section",
		makeNestedFile(depth-1, namePrefix),
	)
}

func makeElements(n int, modifier func(int, *models.Element)) []models.Element {
	elems := make([]models.Element, n)
	for i := range n {
		elem := &models.Element{
			Id:     fmt.Sprintf("elem-%d", i),
			Type:   "rectangle",
			Width:  100,
			Height: 50,
		}
		if modifier != nil {
			modifier(i, elem)
		}
		elems[i] = *elem
	}
	return elems
}

func TestValidateStructurizeRequest_Valid(t *testing.T) {
	t.Parallel()

	req := models.StructurizeRequest{
		RequestID: "req-123",
		UserID:    "user-456",
		File:      makeFile("root", "section", makeFile("main.go", "doc")),
		Board: models.Board{
			BoardID: "board-789",
			Elements: []models.Element{
				{Id: "e1", Type: "text", Content: "hello", Width: 100, Height: 50},
			},
		},
	}

	err := validateStructurizeRequest(req)
	assert.NoError(t, err)
}

func TestValidateStructurizeRequest_EmptyFile(t *testing.T) {
	t.Parallel()

	req := models.StructurizeRequest{
		UserID: "user-1",
		File:   models.File{},
		Board:  models.Board{BoardID: "board-1"},
	}

	err := validateStructurizeRequest(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "file data is empty")
}

func TestValidateStructurizeRequest_EmptyUserID(t *testing.T) {
	t.Parallel()

	req := models.StructurizeRequest{
		UserID: "",
		File:   makeFile("root", "section"),
		Board:  models.Board{BoardID: "board-1"},
	}

	err := validateStructurizeRequest(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "userID is empty")
}

func TestValidateStructurizeRequest_OrderOfFileVsUserID(t *testing.T) {
	t.Parallel()

	req := models.StructurizeRequest{
		UserID: "",
		File:   models.File{},
		Board:  models.Board{BoardID: "board-1"},
	}

	err := validateStructurizeRequest(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "file data is empty")
	assert.NotContains(t, err.Error(), "userID is empty")
}

func TestValidateFileStructure_MaxDepth_Allowed(t *testing.T) {
	t.Parallel()

	deepFile := makeNestedFile(10, "deep")

	err := validateFileStructure(deepFile, 0)
	assert.NoError(t, err)
}

func TestValidateFileStructure_DepthExceeded(t *testing.T) {
	t.Parallel()

	tooDeepFile := makeNestedFile(11, "too-deep")

	err := validateFileStructure(tooDeepFile, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "file structure too deep")
	assert.Contains(t, err.Error(), "10")
}

func TestValidateFileStructure_DepthWithOffset(t *testing.T) {
	t.Parallel()

	deepFile := makeNestedFile(6, "offset")
	err := validateFileStructure(deepFile, 5)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "too deep")

	deepFileOk := makeNestedFile(5, "offset")
	err = validateFileStructure(deepFileOk, 5)
	assert.NoError(t, err)
}

func TestValidateFileStructure_FlatStructure(t *testing.T) {
	t.Parallel()

	file := makeFile("root", "section")
	err := validateFileStructure(file, 0)
	assert.NoError(t, err)
}

func TestValidateFileStructure_MaxNameLength_Allowed(t *testing.T) {
	t.Parallel()

	longName := strings.Repeat("a", 255)
	file := makeFile(longName, "doc")

	err := validateFileStructure(file, 0)
	assert.NoError(t, err)
}

func TestValidateFileStructure_NameTooLong(t *testing.T) {
	t.Parallel()

	tooLongName := strings.Repeat("a", 256)
	file := makeFile(tooLongName, "doc")

	err := validateFileStructure(file, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "file name too long")
	assert.Contains(t, err.Error(), "255")
}

func TestValidateFileStructure_UnicodeName(t *testing.T) {
	t.Parallel()

	unicodeName := strings.Repeat("Пр", 85) // ~255 байт
	file := makeFile(unicodeName, "doc")

	err := validateFileStructure(file, 0)
	assert.Error(t, err)

	assert.NotPanics(t, func() {
		_ = validateFileStructure(file, 0)
	})
}

func TestValidateFileStructure_EmptyChildren_Skipped(t *testing.T) {
	t.Parallel()

	emptyChild := models.File{}
	parent := makeFile("parent", "section", emptyChild)

	err := validateFileStructure(parent, 0)
	assert.NoError(t, err, "пустые дети должны игнорироваться")
}

func TestValidateFileStructure_MixedChildren(t *testing.T) {
	t.Parallel()

	// Смесь пустых и непустых детей
	empty := models.File{}
	valid := makeFile("child", "doc")
	deep := makeNestedFile(5, "deep-child")

	parent := makeFile("root", "section", empty, valid, deep)
	err := validateFileStructure(parent, 0)
	assert.NoError(t, err)
}

func TestValidateFileStructure_ErrorInChild(t *testing.T) {
	t.Parallel()

	invalidChild := makeFile(strings.Repeat("x", 256), "doc") // имя слишком длинное
	parent := makeFile("root", "section", invalidChild)

	err := validateFileStructure(parent, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "file name too long")
}

func TestValidateFileStructure_MultipleLevels_WithErrors(t *testing.T) {
	t.Parallel()

	level2 := makeFile(strings.Repeat("y", 256), "doc")
	level1 := makeFile("level1", "section", level2)
	root := makeFile("root", "section", level1)

	err := validateFileStructure(root, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "too long")
}

func TestValidateStructurizeRequest_BoardElements_TooMany(t *testing.T) {
	t.Parallel()

	req := models.StructurizeRequest{
		UserID: "user-1",
		File:   makeFile("root", "section"),
		Board: models.Board{
			BoardID:  "board-1",
			Elements: makeElements(1001, nil), // > 1000
		},
	}

	err := validateStructurizeRequest(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "too many elements")
}

func TestValidateStructurizeRequest_BoardElements_EmptyID(t *testing.T) {
	t.Parallel()

	req := models.StructurizeRequest{
		UserID: "user-1",
		File:   makeFile("root", "section"),
		Board: models.Board{
			BoardID: "board-1",
			Elements: []models.Element{
				{Id: "valid"},
				{Id: ""},
			},
		},
	}

	err := validateStructurizeRequest(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "element ID cannot be empty")
}

func TestValidateStructurizeRequest_BoardElements_NegativeDimensions(t *testing.T) {
	t.Parallel()

	req := models.StructurizeRequest{
		UserID: "user-1",
		File:   makeFile("root", "section"),
		Board: models.Board{
			BoardID: "board-1",
			Elements: []models.Element{
				{Id: "e1", Width: -10, Height: 50},
			},
		},
	}

	err := validateStructurizeRequest(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-negative")
}

func TestValidateStructurizeRequest_BoardElements_TextContentLimit(t *testing.T) {
	t.Parallel()

	longText := strings.Repeat("x", 10001)
	req := models.StructurizeRequest{
		UserID: "user-1",
		File:   makeFile("root", "section"),
		Board: models.Board{
			BoardID: "board-1",
			Elements: []models.Element{
				{Id: "e1", Type: "text", Content: longText},
			},
		},
	}

	err := validateStructurizeRequest(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "text content too long")
}

func TestValidateStructurizeRequest_FullValidRequest(t *testing.T) {
	t.Parallel()

	deepButValid := makeNestedFile(9, "deep")
	elements := makeElements(500, func(i int, e *models.Element) {
		if i%10 == 0 {
			e.Type = "text"
			e.Content = fmt.Sprintf("text-%d", i)
		}
	})

	req := models.StructurizeRequest{
		RequestID: "req-full",
		UserID:    "user-full",
		File:      makeFile("project", "section", deepButValid),
		Board: models.Board{
			BoardID:  "board-full",
			Elements: elements,
		},
	}

	err := validateStructurizeRequest(req)
	assert.NoError(t, err)
}

func TestValidateStructurizeRequest_MultipleErrors_FirstWins(t *testing.T) {
	t.Parallel()

	req := models.StructurizeRequest{
		UserID: "",
		File:   models.File{},
		Board: models.Board{
			BoardID:  "board-1",
			Elements: makeElements(1001, nil),
		},
	}

	err := validateStructurizeRequest(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "file data is empty")
}

func TestValidateFileStructure_RecursionSafety(t *testing.T) {
	t.Parallel()

	deepFile := makeNestedFile(10, "safe")

	assert.NotPanics(t, func() {
		_ = validateFileStructure(deepFile, 0)
	})
}

func TestValidateFileStructure_EmptyName_Allowed(t *testing.T) {
	t.Parallel()

	file := makeFile("", "section", makeFile("child.txt", "doc"))
	err := validateFileStructure(file, 0)
	assert.NoError(t, err)
}

func TestValidateStructurizeRequest_SpecialCharsInFileNames(t *testing.T) {
	t.Parallel()

	specialName := "file/with\\special:chars@2024#test.txt"
	req := models.StructurizeRequest{
		UserID: "user-1",
		File:   makeFile(specialName, "doc"),
		Board:  models.Board{BoardID: "board-1"},
	}

	err := validateStructurizeRequest(req)
	assert.NoError(t, err)
}
