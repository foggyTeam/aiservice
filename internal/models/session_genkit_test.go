package models

import (
	"testing"

	"github.com/firebase/genkit/go/ai"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToGenkitMessages_WithEmptyEntries(t *testing.T) {
	entries := []MessageEntry{}

	result := ToGenkitMessages(entries)

	require.NotNil(t, result)
	assert.Len(t, result, 0)
}

func TestToGenkitMessages_WithNilEntries(t *testing.T) {
	var entries []MessageEntry

	result := ToGenkitMessages(entries)

	require.NotNil(t, result)
	assert.Len(t, result, 0)
}

func TestToGenkitMessages_WithSingleUserMessage(t *testing.T) {
	entries := []MessageEntry{
		{Role: "user", Content: "Hello"},
	}

	result := ToGenkitMessages(entries)

	require.NotNil(t, result)
	require.Len(t, result, 1)
	assert.NotNil(t, result[0])
}

func TestToGenkitMessages_WithSingleModelMessage(t *testing.T) {
	entries := []MessageEntry{
		{Role: "model", Content: "Hi there"},
	}

	result := ToGenkitMessages(entries)

	require.NotNil(t, result)
	require.Len(t, result, 1)
	assert.NotNil(t, result[0])
}

func TestToGenkitMessages_WithMultipleMessages(t *testing.T) {
	entries := []MessageEntry{
		{Role: "user", Content: "Hello"},
		{Role: "model", Content: "Hi"},
		{Role: "user", Content: "How are you?"},
		{Role: "model", Content: "I'm great!"},
	}

	result := ToGenkitMessages(entries)

	require.NotNil(t, result)
	require.Len(t, result, 4)
	for _, msg := range result {
		assert.NotNil(t, msg)
	}
}

func TestToGenkitMessages_WithAlternatingRoles(t *testing.T) {
	entries := []MessageEntry{
		{Role: "user", Content: "Message 1"},
		{Role: "model", Content: "Message 2"},
		{Role: "user", Content: "Message 3"},
		{Role: "model", Content: "Message 4"},
		{Role: "user", Content: "Message 5"},
	}

	result := ToGenkitMessages(entries)

	require.NotNil(t, result)
	require.Len(t, result, 5)
}

func TestToGenkitMessages_PreservesContentOrder(t *testing.T) {
	entries := []MessageEntry{
		{Role: "user", Content: "first"},
		{Role: "user", Content: "second"},
		{Role: "user", Content: "third"},
	}

	result := ToGenkitMessages(entries)

	require.NotNil(t, result)
	require.Len(t, result, 3)
	for _, msg := range result {
		assert.NotNil(t, msg)
		assert.NotZero(t, msg)
	}
}

func TestToGenkitMessages_WithEmptyContent(t *testing.T) {
	entries := []MessageEntry{
		{Role: "user", Content: ""},
		{Role: "model", Content: ""},
	}

	result := ToGenkitMessages(entries)

	require.NotNil(t, result)
	require.Len(t, result, 2)
	for _, msg := range result {
		assert.NotNil(t, msg)
	}
}

func TestToGenkitMessages_WithSpecialCharacters(t *testing.T) {
	entries := []MessageEntry{
		{Role: "user", Content: "Hello @#$%^&*()"},
		{Role: "model", Content: "Special: !<>?"},
	}

	result := ToGenkitMessages(entries)

	require.NotNil(t, result)
	require.Len(t, result, 2)
	for _, msg := range result {
		assert.NotNil(t, msg)
	}
}

func TestToGenkitMessages_WithUnicodeContent(t *testing.T) {
	entries := []MessageEntry{
		{Role: "user", Content: "你好"},
		{Role: "model", Content: "Привет"},
		{Role: "user", Content: "مرحبا"},
		{Role: "model", Content: "😀 👍 🎉"},
	}

	result := ToGenkitMessages(entries)

	require.NotNil(t, result)
	require.Len(t, result, 4)
	for _, msg := range result {
		assert.NotNil(t, msg)
	}
}

func TestToGenkitMessages_WithMultilineContent(t *testing.T) {
	entries := []MessageEntry{
		{Role: "user", Content: "Line 1\nLine 2\nLine 3"},
		{Role: "model", Content: "Response\nwith\nmultiple\nlines"},
	}

	result := ToGenkitMessages(entries)

	require.NotNil(t, result)
	require.Len(t, result, 2)
	for _, msg := range result {
		assert.NotNil(t, msg)
	}
}

func TestToGenkitMessages_WithVeryLongContent(t *testing.T) {
	longContent := ""
	for i := 0; i < 10000; i++ {
		longContent += "x"
	}

	entries := []MessageEntry{
		{Role: "user", Content: longContent},
		{Role: "model", Content: longContent},
	}

	result := ToGenkitMessages(entries)

	require.NotNil(t, result)
	require.Len(t, result, 2)
	for _, msg := range result {
		assert.NotNil(t, msg)
	}
}

func TestToGenkitMessages_UserRoleCorrectly(t *testing.T) {
	entries := []MessageEntry{
		{Role: "user", Content: "User message"},
	}

	result := ToGenkitMessages(entries)

	require.NotNil(t, result)
	require.Len(t, result, 1)
	assert.NotNil(t, result[0])
	userMsg := ai.NewUserMessage(ai.NewTextPart("test"))
	assert.Equal(t, len(userMsg.Content), len(result[0].Content))
}

func TestToGenkitMessages_ModelRoleCorrectly(t *testing.T) {
	entries := []MessageEntry{
		{Role: "model", Content: "Model message"},
	}

	result := ToGenkitMessages(entries)

	require.NotNil(t, result)
	require.Len(t, result, 1)
	assert.NotNil(t, result[0])
}

func TestToGenkitMessages_UnknownRoleDefaultsToModel(t *testing.T) {
	entries := []MessageEntry{
		{Role: "assistant", Content: "Unknown role"},
		{Role: "ai", Content: "Another unknown"},
		{Role: "", Content: "Empty role"},
	}

	result := ToGenkitMessages(entries)

	require.NotNil(t, result)
	require.Len(t, result, 3)
	for _, msg := range result {
		assert.NotNil(t, msg)
	}
}

func TestToGenkitMessages_CaseSensitiveRoles(t *testing.T) {
	entries := []MessageEntry{
		{Role: "User", Content: "Capital U"},
		{Role: "USER", Content: "All caps"},
		{Role: "Model", Content: "Capital M"},
		{Role: "MODEL", Content: "All caps model"},
	}

	result := ToGenkitMessages(entries)

	require.NotNil(t, result)
	require.Len(t, result, 4)
	for _, msg := range result {
		assert.NotNil(t, msg)
	}
}

func TestToGenkitMessages_WithSpaceInContent(t *testing.T) {
	entries := []MessageEntry{
		{Role: "user", Content: "   leading spaces"},
		{Role: "model", Content: "trailing spaces   "},
		{Role: "user", Content: "   both   "},
	}

	result := ToGenkitMessages(entries)

	require.NotNil(t, result)
	require.Len(t, result, 3)
	for _, msg := range result {
		assert.NotNil(t, msg)
	}
}

func TestToGenkitMessages_WithTabsAndNewlines(t *testing.T) {
	entries := []MessageEntry{
		{Role: "user", Content: "Content\twith\ttabs"},
		{Role: "model", Content: "Content\r\nwith\r\nCRLF"},
		{Role: "user", Content: "\t\n  mixed  \n\t"},
	}

	result := ToGenkitMessages(entries)

	require.NotNil(t, result)
	require.Len(t, result, 3)
	for _, msg := range result {
		assert.NotNil(t, msg)
	}
}

func TestToGenkitMessages_ManyMessages(t *testing.T) {
	entries := make([]MessageEntry, 100)
	for i := 0; i < 100; i++ {
		if i%2 == 0 {
			entries[i] = MessageEntry{Role: "user", Content: "Message"}
		} else {
			entries[i] = MessageEntry{Role: "model", Content: "Response"}
		}
	}

	result := ToGenkitMessages(entries)

	require.NotNil(t, result)
	require.Len(t, result, 100)
	for _, msg := range result {
		assert.NotNil(t, msg)
	}
}

func TestToGenkitMessages_ConsecutiveUserMessages(t *testing.T) {
	entries := []MessageEntry{
		{Role: "user", Content: "First"},
		{Role: "user", Content: "Second"},
		{Role: "user", Content: "Third"},
	}

	result := ToGenkitMessages(entries)

	require.NotNil(t, result)
	require.Len(t, result, 3)
}

func TestToGenkitMessages_ConsecutiveModelMessages(t *testing.T) {
	entries := []MessageEntry{
		{Role: "model", Content: "First"},
		{Role: "model", Content: "Second"},
		{Role: "model", Content: "Third"},
	}

	result := ToGenkitMessages(entries)

	require.NotNil(t, result)
	require.Len(t, result, 3)
}

func TestToGenkitMessages_OutputIsSlice(t *testing.T) {
	entries := []MessageEntry{
		{Role: "user", Content: "Test 1"},
		{Role: "model", Content: "Test 2"},
	}

	result := ToGenkitMessages(entries)

	require.NotNil(t, result)
	assert.IsType(t, []*ai.Message{}, result)
}

func TestToGenkitMessages_OutputCanBeIterated(t *testing.T) {
	entries := []MessageEntry{
		{Role: "user", Content: "Msg1"},
		{Role: "model", Content: "Msg2"},
		{Role: "user", Content: "Msg3"},
	}

	result := ToGenkitMessages(entries)

	count := 0
	for range result {
		count++
	}
	assert.Equal(t, 3, count)
}

func TestToGenkitMessages_OutputCanBeIndexed(t *testing.T) {
	entries := []MessageEntry{
		{Role: "user", Content: "Msg1"},
		{Role: "model", Content: "Msg2"},
	}

	result := ToGenkitMessages(entries)

	assert.NotNil(t, result[0])
	assert.NotNil(t, result[1])
}

func TestToGenkitMessages_LargeConversation(t *testing.T) {
	entries := make([]MessageEntry, 200)
	for i := 0; i < 200; i++ {
		role := "user"
		if i%2 == 1 {
			role = "model"
		}
		entries[i] = MessageEntry{
			Role:    role,
			Content: "Message content",
		}
	}

	result := ToGenkitMessages(entries)

	require.NotNil(t, result)
	require.Len(t, result, 200)
	for i, msg := range result {
		assert.NotNil(t, msg, "Message at index %d is nil", i)
	}
}

func TestToGenkitMessages_RoleWithWhitespace(t *testing.T) {
	entries := []MessageEntry{
		{Role: " user", Content: "Space before"},
		{Role: "user ", Content: "Space after"},
		{Role: " user ", Content: "Space both sides"},
	}

	result := ToGenkitMessages(entries)

	require.NotNil(t, result)
	require.Len(t, result, 3)
	for _, msg := range result {
		assert.NotNil(t, msg)
	}
}

func TestToGenkitMessages_IntegrationWithShiftAndAppendLimited(t *testing.T) {
	messages := []MessageEntry{
		{Role: "user", Content: "msg1"},
		{Role: "model", Content: "msg2"},
	}

	newMsgs := []MessageEntry{
		{Role: "user", Content: "msg3"},
	}

	combined := ShiftAndAppendLimited(messages, newMsgs, 5)
	result := ToGenkitMessages(combined)

	require.NotNil(t, result)
	require.Len(t, result, 3)
}
