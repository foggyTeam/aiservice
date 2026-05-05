package handlers

import (
	"testing"

	"github.com/aiservice/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestValidate_ValidRequest(t *testing.T) {
	req := models.GenerateTemplateRequest{
		RequestID:   "req-123",
		UserID:      "user-123",
		RequestType: models.GenerateTemplateType,
		BoardID:     "board-123",
		Prompt:      "Create a sample architecture diagram for a microservices system",
		BoardType:   models.BoardTypeSimple,
	}

	err := validateTemplate(req)
	assert.NoError(t, err, "Valid request should not return error")
}

func TestValidate_MissingRequestID(t *testing.T) {
	req := models.GenerateTemplateRequest{
		RequestID:   "",
		UserID:      "user-123",
		RequestType: models.GenerateTemplateType,
		BoardID:     "board-123",
		Prompt:      "Create a sample architecture diagram",
		BoardType:   models.BoardTypeSimple,
	}

	err := validateTemplate(req)

	assert.Error(t, err, "Missing RequestID should return error")
	assert.Contains(t, err.Error(), "requestId is required")
}

func TestValidate_MissingUserID(t *testing.T) {
	req := models.GenerateTemplateRequest{
		RequestID:   "req-123",
		UserID:      "",
		RequestType: models.GenerateTemplateType,
		BoardID:     "board-123",
		Prompt:      "Create a sample architecture diagram",
		BoardType:   models.BoardTypeSimple,
	}

	err := validateTemplate(req)

	assert.Error(t, err, "Missing UserID should return error")
	assert.Contains(t, err.Error(), "userId is required")
}

func TestValidate_InvalidRequestType(t *testing.T) {
	req := models.GenerateTemplateRequest{
		RequestID:   "req-123",
		UserID:      "user-123",
		RequestType: "invalidType",
		BoardID:     "board-123",
		Prompt:      "Create a sample architecture diagram",
		BoardType:   models.BoardTypeSimple,
	}

	err := validateTemplate(req)

	assert.Error(t, err, "Invalid RequestType should return error")
	assert.Contains(t, err.Error(), "requestType must be 'generateTemplate'")
}

func TestValidate_ValidGenerateTextType(t *testing.T) {
	req := models.GenerateTemplateRequest{
		RequestID:   "req-123",
		UserID:      "user-123",
		RequestType: models.GenerateTextType,
		BoardID:     "board-123",
		Prompt:      "Generate a detailed description of the system",
		BoardType:   models.BoardTypeSimple,
	}

	err := validateTemplate(req)

	assert.NoError(t, err, "GenerateTextType should be valid")
}

func TestValidate_MissingBoardID(t *testing.T) {
	req := models.GenerateTemplateRequest{
		RequestID:   "req-123",
		UserID:      "user-123",
		RequestType: models.GenerateTemplateType,
		BoardID:     "",
		Prompt:      "Create a sample architecture diagram",
		BoardType:   models.BoardTypeSimple,
	}

	err := validateTemplate(req)

	assert.Error(t, err, "Missing BoardID should return error")
	assert.Contains(t, err.Error(), "boardId is required")
}

func TestValidate_MissingPrompt(t *testing.T) {
	req := models.GenerateTemplateRequest{
		RequestID:   "req-123",
		UserID:      "user-123",
		RequestType: models.GenerateTemplateType,
		BoardID:     "board-123",
		Prompt:      "",
		BoardType:   models.BoardTypeSimple,
	}

	err := validateTemplate(req)

	assert.Error(t, err, "Missing Prompt should return error")
	assert.Contains(t, err.Error(), "prompt is required")
}

func TestValidate_PromptTooShort(t *testing.T) {
	req := models.GenerateTemplateRequest{
		RequestID:   "req-123",
		UserID:      "user-123",
		RequestType: models.GenerateTemplateType,
		BoardID:     "board-123",
		Prompt:      "short",
		BoardType:   models.BoardTypeSimple,
	}

	err := validateTemplate(req)

	assert.Error(t, err, "Prompt less than 10 characters should return error")
	assert.Contains(t, err.Error(), "prompt must be at least 10 characters")
}

func TestValidate_PromptExactly10Characters(t *testing.T) {
	req := models.GenerateTemplateRequest{
		RequestID:   "req-123",
		UserID:      "user-123",
		RequestType: models.GenerateTemplateType,
		BoardID:     "board-123",
		Prompt:      "1234567890",
		BoardType:   models.BoardTypeSimple,
	}

	err := validateTemplate(req)

	assert.NoError(t, err, "Prompt with exactly 10 characters should be valid")
}

func TestValidate_PromptTooLong(t *testing.T) {
	// Create a prompt longer than 2000 characters
	longPrompt := ""
	for i := 0; i < 201; i++ {
		longPrompt += "0123456789"
	}

	req := models.GenerateTemplateRequest{
		RequestID:   "req-123",
		UserID:      "user-123",
		RequestType: models.GenerateTemplateType,
		BoardID:     "board-123",
		Prompt:      longPrompt,
		BoardType:   models.BoardTypeSimple,
	}

	err := validateTemplate(req)
	assert.Error(t, err, "Prompt longer than 2000 characters should return error")
	assert.Contains(t, err.Error(), "prompt must be at most 2000 characters")
}

func TestValidate_PromptExactly2000Characters(t *testing.T) {
	// Create a prompt with exactly 2000 characters
	prompt := ""
	for i := 0; i < 200; i++ {
		prompt += "0123456789"
	}

	req := models.GenerateTemplateRequest{
		RequestID:   "req-123",
		UserID:      "user-123",
		RequestType: models.GenerateTemplateType,
		BoardID:     "board-123",
		Prompt:      prompt,
		BoardType:   models.BoardTypeSimple,
	}

	err := validateTemplate(req)

	assert.NoError(t, err, "Prompt with exactly 2000 characters should be valid")
}

func TestValidate_MissingBoardType(t *testing.T) {
	req := models.GenerateTemplateRequest{
		RequestID:   "req-123",
		UserID:      "user-123",
		RequestType: models.GenerateTemplateType,
		BoardID:     "board-123",
		Prompt:      "Create a sample architecture diagram",
		BoardType:   "",
	}

	err := validateTemplate(req)

	assert.Error(t, err, "Missing BoardType should return error")
	assert.Contains(t, err.Error(), "boardType is required")
}

func TestValidate_InvalidBoardType(t *testing.T) {
	req := models.GenerateTemplateRequest{
		RequestID:   "req-123",
		UserID:      "user-123",
		RequestType: models.GenerateTemplateType,
		BoardID:     "board-123",
		Prompt:      "Create a sample architecture diagram",
		BoardType:   "invalidBoardType",
	}

	err := validateTemplate(req)

	assert.Error(t, err, "Invalid BoardType should return error")
	assert.Contains(t, err.Error(), "boardType must be 'simple', 'graph', or 'doc'")
}

func TestValidate_ValidBoardTypeGraph(t *testing.T) {
	req := models.GenerateTemplateRequest{
		RequestID:   "req-123",
		UserID:      "user-123",
		RequestType: models.GenerateTemplateType,
		BoardID:     "board-123",
		Prompt:      "Create a flowchart diagram",
		BoardType:   models.BoardTypeGraph,
	}

	err := validateTemplate(req)

	assert.NoError(t, err, "BoardTypeGraph should be valid")
}

func TestValidate_ValidBoardTypeDoc(t *testing.T) {
	req := models.GenerateTemplateRequest{
		RequestID:   "req-123",
		UserID:      "user-123",
		RequestType: models.GenerateTemplateType,
		BoardID:     "board-123",
		Prompt:      "Create a document template",
		BoardType:   models.BoardTypeDOC,
	}

	err := validateTemplate(req)

	assert.NoError(t, err, "BoardTypeDOC should be valid")
}

func TestValidate_MultipleErrors_FirstErrorReturned(t *testing.T) {
	req := models.GenerateTemplateRequest{
		RequestID:   "",
		UserID:      "",
		RequestType: models.GenerateTemplateType,
		BoardID:     "board-123",
		Prompt:      "Create a sample architecture diagram",
		BoardType:   models.BoardTypeSimple,
	}

	err := validateTemplate(req)

	assert.Error(t, err, "Missing RequestID should return error")
	assert.Contains(t, err.Error(), "requestId is required")
}

func TestValidate_EdgeCase_AllFieldsEmpty(t *testing.T) {
	req := models.GenerateTemplateRequest{
		RequestID:   "",
		UserID:      "",
		RequestType: "",
		BoardID:     "",
		Prompt:      "",
		BoardType:   "",
	}

	err := validateTemplate(req)

	assert.Error(t, err, "All empty fields should return error")
	assert.Contains(t, err.Error(), "requestId is required")
}
