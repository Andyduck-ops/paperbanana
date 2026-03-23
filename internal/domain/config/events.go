package config

import "time"

// ConfigEventType represents the type of configuration change.
type ConfigEventType string

const (
	// EventProviderCreated is emitted when a new provider is created.
	EventProviderCreated ConfigEventType = "provider_created"
	// EventProviderUpdated is emitted when a provider is updated.
	EventProviderUpdated ConfigEventType = "provider_updated"
	// EventProviderDeleted is emitted when a provider is deleted.
	EventProviderDeleted ConfigEventType = "provider_deleted"
	// EventKeyAdded is emitted when an API key is added.
	EventKeyAdded ConfigEventType = "key_added"
	// EventKeyDeleted is emitted when an API key is deleted.
	EventKeyDeleted ConfigEventType = "key_deleted"
	// EventKeyToggled is emitted when an API key's active status changes.
	EventKeyToggled ConfigEventType = "key_toggled"
)

// ConfigEvent represents a configuration change event.
type ConfigEvent struct {
	Type       ConfigEventType `json:"type"`
	ProviderID string          `json:"provider_id"`
	KeyID      string          `json:"key_id,omitempty"`
	Timestamp  time.Time       `json:"timestamp"`
}

// ConfigWatcher watches for configuration changes.
type ConfigWatcher interface {
	Subscribe() <-chan ConfigEvent
	Unsubscribe(ch <-chan ConfigEvent)
}
