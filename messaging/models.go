package messaging

import (
	"encoding/json"
)

// JSONMessage is an interface for returning a json string.
type JSONMessage interface {
	JSON() (string, error)
}

// SyncMessage is a type of message for sending sync commands.
type SyncMessage struct {
	CharacterSlug      string `json:"slug"`
	CharacterSyncLogID uint   `json:"sync_log_id"`
}

// JSON returns the json representation of the message.
func (m *SyncMessage) JSON() (string, error) {
	j, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(j), nil
}

// NewSyncMessageFromString constructs a new sync message from a string.
func NewSyncMessageFromString(message string) (*SyncMessage, error) {
	cm := &SyncMessage{}
	err := json.Unmarshal([]byte(message), cm)
	if err != nil {
		return nil, err
	}
	return cm, nil
}

// NewSyncMessage returns a new sync message.
func NewSyncMessage(slug string, syncLogID uint) JSONMessage {
	return &SyncMessage{
		CharacterSlug:      slug,
		CharacterSyncLogID: syncLogID,
	}
}
