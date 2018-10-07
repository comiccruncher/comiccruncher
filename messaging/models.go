package messaging

import "encoding/json"

// An interface for returning a json string.
type JsonMessage interface {
	Json() (string, error)
}

// A type of message for sending sync commands.
type SyncMessage struct {
	CharacterSlug      string `json:"slug"`
	CharacterSyncLogId uint   `json:"sync_log_id"`
}

// Returns the json representation of the message.
func (m *SyncMessage) Json() (string, error) {
	j, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(j), nil
}

// Constructs a new sync message from a string.
func NewSyncMessageFromString(message string) (*SyncMessage, error) {
	cm := &SyncMessage{}
	err := json.Unmarshal([]byte(message), cm)
	if err != nil {
		return nil, err
	}
	return cm, nil
}

// Returns a new sync message.
func NewSyncMessage(slug string, syncLogId uint) JsonMessage {
	return &SyncMessage{
		CharacterSlug:      slug,
		CharacterSyncLogId: syncLogId,
	}
}
