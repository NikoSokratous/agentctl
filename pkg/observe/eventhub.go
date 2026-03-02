package observe

import (
	"sync"
)

// EventHub manages pub/sub for runtime events.
type EventHub struct {
	subscribers map[string][]chan Event
	mu          sync.RWMutex
	bufferSize  int
}

// NewEventHub creates a new event hub.
func NewEventHub(bufferSize int) *EventHub {
	if bufferSize <= 0 {
		bufferSize = 100
	}
	return &EventHub{
		subscribers: make(map[string][]chan Event),
		bufferSize:  bufferSize,
	}
}

// Subscribe creates a new subscription for a run's events.
func (h *EventHub) Subscribe(runID string) <-chan Event {
	h.mu.Lock()
	defer h.mu.Unlock()

	ch := make(chan Event, h.bufferSize)
	h.subscribers[runID] = append(h.subscribers[runID], ch)
	return ch
}

// Publish sends an event to all subscribers of a run.
func (h *EventHub) Publish(runID string, event Event) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	subs, ok := h.subscribers[runID]
	if !ok {
		return
	}

	// Send to all subscribers (non-blocking)
	for _, ch := range subs {
		select {
		case ch <- event:
		default:
			// Channel full, skip to prevent blocking
		}
	}
}

// Unsubscribe removes a subscription channel.
func (h *EventHub) Unsubscribe(runID string, ch <-chan Event) {
	h.mu.Lock()
	defer h.mu.Unlock()

	subs, ok := h.subscribers[runID]
	if !ok {
		return
	}

	// Find and remove the channel
	for i, sub := range subs {
		if sub == ch {
			h.subscribers[runID] = append(subs[:i], subs[i+1:]...)
			close(sub)
			break
		}
	}

	// Clean up if no more subscribers
	if len(h.subscribers[runID]) == 0 {
		delete(h.subscribers, runID)
	}
}

// UnsubscribeAll removes all subscriptions for a run and closes channels.
func (h *EventHub) UnsubscribeAll(runID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	subs, ok := h.subscribers[runID]
	if !ok {
		return
	}

	// Close all channels
	for _, ch := range subs {
		close(ch)
	}

	delete(h.subscribers, runID)
}

// SubscriberCount returns the number of active subscribers for a run.
func (h *EventHub) SubscriberCount(runID string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.subscribers[runID])
}
