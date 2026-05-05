package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToGenkitMessagesOrder(t *testing.T) {
	entries := []MessageEntry{
		{Role: "user", Content: "First"},
		{Role: "model", Content: "Second"},
		{Role: "user", Content: "Third"},
	}

	msgs := ToGenkitMessages(entries)

	assert.Len(t, msgs, 3)
	for i := range msgs {
		assert.NotNil(t, msgs[i])
	}
}

func TestToGenkitMessagesNilInput(t *testing.T) {
	var entries []MessageEntry
	msgs := ToGenkitMessages(entries)

	assert.NotNil(t, msgs)
	assert.Empty(t, msgs)
}

func TestToGenkitMessagesLargeConversation(t *testing.T) {
	entries := make([]MessageEntry, 100)
	for i := 0; i < 100; i++ {
		if i%2 == 0 {
			entries[i] = MessageEntry{Role: "user", Content: "Message " + string(rune(i))}
		} else {
			entries[i] = MessageEntry{Role: "model", Content: "Response " + string(rune(i))}
		}
	}

	msgs := ToGenkitMessages(entries)

	assert.Len(t, msgs, 100)
	for _, msg := range msgs {
		assert.NotNil(t, msg)
	}
}

func TestShiftAndAppendLimited(t *testing.T) {
	tests := []struct {
		name        string
		messages    []MessageEntry
		newMsgs     []MessageEntry
		maxSize     int
		expectedLen int
	}{
		{
			name:        "empty initial messages",
			messages:    []MessageEntry{},
			newMsgs:     []MessageEntry{{Role: "user", Content: "Hello"}},
			maxSize:     10,
			expectedLen: 1,
		},
		{
			name:        "append to existing messages within limit",
			messages:    []MessageEntry{{Role: "user", Content: "Hi"}},
			newMsgs:     []MessageEntry{{Role: "model", Content: "Hello"}},
			maxSize:     10,
			expectedLen: 2,
		},
		{
			name:        "exceed max size should trim old messages",
			messages:    []MessageEntry{{Role: "user", Content: "1"}, {Role: "user", Content: "2"}},
			newMsgs:     []MessageEntry{{Role: "user", Content: "3"}, {Role: "user", Content: "4"}},
			maxSize:     3,
			expectedLen: 3,
		},
		{
			name:        "empty new messages",
			messages:    []MessageEntry{{Role: "user", Content: "1"}, {Role: "user", Content: "2"}},
			newMsgs:     []MessageEntry{},
			maxSize:     10,
			expectedLen: 2,
		},
		{
			name:        "empty new messages with trim needed",
			messages:    []MessageEntry{{Role: "user", Content: "1"}, {Role: "user", Content: "2"}, {Role: "user", Content: "3"}},
			newMsgs:     []MessageEntry{},
			maxSize:     2,
			expectedLen: 2,
		},
		{
			name:        "max size zero returns empty",
			messages:    []MessageEntry{{Role: "user", Content: "1"}},
			newMsgs:     []MessageEntry{{Role: "user", Content: "2"}},
			maxSize:     0,
			expectedLen: 0,
		},
		{
			name:        "max size negative returns empty",
			messages:    []MessageEntry{{Role: "user", Content: "1"}},
			newMsgs:     []MessageEntry{{Role: "user", Content: "2"}},
			maxSize:     -1,
			expectedLen: 0,
		},
		{
			name:        "single max size",
			messages:    []MessageEntry{{Role: "user", Content: "1"}, {Role: "user", Content: "2"}},
			newMsgs:     []MessageEntry{{Role: "user", Content: "3"}},
			maxSize:     1,
			expectedLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ShiftAndAppendLimited(tt.messages, tt.newMsgs, tt.maxSize)
			assert.Len(t, result, tt.expectedLen)
		})
	}
}

func TestShiftAndAppendLimitedKeepsLatest(t *testing.T) {
	messages := []MessageEntry{
		{Role: "user", Content: "1"},
		{Role: "user", Content: "2"},
	}
	newMsgs := []MessageEntry{
		{Role: "user", Content: "3"},
		{Role: "user", Content: "4"},
	}

	result := ShiftAndAppendLimited(messages, newMsgs, 3)

	assert.Len(t, result, 3)
	assert.Equal(t, "2", result[0].Content)
	assert.Equal(t, "3", result[1].Content)
	assert.Equal(t, "4", result[2].Content)
}

func TestShiftAndAppendLimitedPreservesOrder(t *testing.T) {
	messages := []MessageEntry{
		{Role: "user", Content: "A"},
		{Role: "model", Content: "B"},
	}
	newMsgs := []MessageEntry{
		{Role: "user", Content: "C"},
		{Role: "model", Content: "D"},
	}

	result := ShiftAndAppendLimited(messages, newMsgs, 10)

	assert.Equal(t, []MessageEntry{
		{Role: "user", Content: "A"},
		{Role: "model", Content: "B"},
		{Role: "user", Content: "C"},
		{Role: "model", Content: "D"},
	}, result)
}

func TestShiftAndAppendLimitedLargeConversation(t *testing.T) {
	messages := make([]MessageEntry, 50)
	for i := 0; i < 50; i++ {
		messages[i] = MessageEntry{Role: "user", Content: "Message " + string(rune(i))}
	}

	newMsgs := make([]MessageEntry, 30)
	for i := 0; i < 30; i++ {
		newMsgs[i] = MessageEntry{Role: "model", Content: "Response " + string(rune(i))}
	}

	result := ShiftAndAppendLimited(messages, newMsgs, 60)

	assert.Len(t, result, 60)
	assert.Equal(t, "Message ", result[0].Content[:8])
}

func TestShiftAndAppendLimitedEdgeCases(t *testing.T) {
	t.Run("exact fit", func(t *testing.T) {
		messages := []MessageEntry{{Role: "user", Content: "1"}, {Role: "user", Content: "2"}}
		newMsgs := []MessageEntry{{Role: "user", Content: "3"}, {Role: "user", Content: "4"}}

		result := ShiftAndAppendLimited(messages, newMsgs, 4)

		assert.Len(t, result, 4)
	})

	t.Run("new messages exceed max by themselves", func(t *testing.T) {
		messages := []MessageEntry{{Role: "user", Content: "1"}}
		newMsgs := []MessageEntry{
			{Role: "user", Content: "2"},
			{Role: "user", Content: "3"},
			{Role: "user", Content: "4"},
			{Role: "user", Content: "5"},
		}

		result := ShiftAndAppendLimited(messages, newMsgs, 3)

		assert.Len(t, result, 3)
		assert.Equal(t, "3", result[0].Content)
		assert.Equal(t, "4", result[1].Content)
		assert.Equal(t, "5", result[2].Content)
	})

	t.Run("nil messages slice", func(t *testing.T) {
		var messages []MessageEntry
		newMsgs := []MessageEntry{{Role: "user", Content: "1"}}

		result := ShiftAndAppendLimited(messages, newMsgs, 5)

		assert.Len(t, result, 1)
	})
}
