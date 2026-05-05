package models

import "github.com/firebase/genkit/go/ai"

// BoardSessionState represents the session state for a board's chat history
type BoardSessionState struct {
	Messages []MessageEntry `json:"messages"`
}

// MessageEntry represents a single message in the chat history
type MessageEntry struct {
	Role    string `json:"role"`    // "user" or "model"
	Content string `json:"content"` // Text content of the message
}

// ToGenkitMessages converts session messages to Genkit messages
func ToGenkitMessages(entries []MessageEntry) []*ai.Message {
	msgs := make([]*ai.Message, 0, len(entries))
	for _, entry := range entries {
		if entry.Role == "user" {
			msgs = append(msgs, ai.NewUserMessage(ai.NewTextPart(entry.Content)))
		} else {
			// Default to model/assistant
			msgs = append(msgs, ai.NewModelMessage(ai.NewTextPart(entry.Content)))
		}
	}
	return msgs
}

func ShiftAndAppendLimited(messages []MessageEntry, newMsgs []MessageEntry, maxSize int) []MessageEntry {
	if maxSize <= 0 {
		return []MessageEntry{}
	}
	if len(newMsgs) == 0 {
		if len(messages) > maxSize {
			return messages[len(messages)-maxSize:]
		}
		return messages
	}
	combined := make([]MessageEntry, 0, len(messages)+len(newMsgs))
	combined = append(combined, messages...)
	combined = append(combined, newMsgs...)
	if len(combined) > maxSize {
		combined = combined[len(combined)-maxSize:]
	}
	return combined
}
