package config

import (
	"sync"
	"time"

	domainconfig "github.com/paperbanana/paperbanana/internal/domain/config"
)

// Watcher implements ConfigWatcher using in-memory event broadcasting.
type Watcher struct {
	mu          sync.RWMutex
	subscribers []chan domainconfig.ConfigEvent
}

// NewWatcher creates a new config watcher.
func NewWatcher() *Watcher {
	return &Watcher{
		subscribers: make([]chan domainconfig.ConfigEvent, 0),
	}
}

// Subscribe returns a channel for receiving config events.
func (w *Watcher) Subscribe() <-chan domainconfig.ConfigEvent {
	w.mu.Lock()
	defer w.mu.Unlock()

	ch := make(chan domainconfig.ConfigEvent, 10)
	w.subscribers = append(w.subscribers, ch)
	return ch
}

// Unsubscribe removes a subscriber.
func (w *Watcher) Unsubscribe(ch <-chan domainconfig.ConfigEvent) {
	w.mu.Lock()
	defer w.mu.Unlock()

	for i, sub := range w.subscribers {
		if sub == ch {
			w.subscribers = append(w.subscribers[:i], w.subscribers[i+1:]...)
			close(sub)
			break
		}
	}
}

// Emit broadcasts a config event to all subscribers.
func (w *Watcher) Emit(eventType domainconfig.ConfigEventType, providerID, keyID string) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	event := domainconfig.ConfigEvent{
		Type:       eventType,
		ProviderID: providerID,
		KeyID:      keyID,
		Timestamp:  time.Now(),
	}

	for _, sub := range w.subscribers {
		select {
		case sub <- event:
		default:
			// Channel full, skip
		}
	}
}
