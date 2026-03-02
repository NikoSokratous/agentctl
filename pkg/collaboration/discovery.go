package collaboration

import (
	"context"
	"sync"
)

// MessageRouter routes messages between agents.
type MessageRouter struct {
	subscribers map[string][]chan *AgentMessage
	mu          sync.RWMutex
}

// NewMessageRouter creates a new message router.
func NewMessageRouter() *MessageRouter {
	return &MessageRouter{
		subscribers: make(map[string][]chan *AgentMessage),
	}
}

// Subscribe subscribes an agent to messages.
func (r *MessageRouter) Subscribe(agentID, pattern string) chan *AgentMessage {
	r.mu.Lock()
	defer r.mu.Unlock()

	ch := make(chan *AgentMessage, 100)
	key := agentID + ":" + pattern

	r.subscribers[key] = append(r.subscribers[key], ch)

	return ch
}

// Unsubscribe unsubscribes an agent from all messages.
func (r *MessageRouter) Unsubscribe(agentID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for key := range r.subscribers {
		if len(key) > len(agentID) && key[:len(agentID)] == agentID {
			delete(r.subscribers, key)
		}
	}
}

// Route routes a message to its destination(s).
func (r *MessageRouter) Route(ctx context.Context, message *AgentMessage) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if message.IsBroadcast() {
		// Send to all subscribers
		for _, channels := range r.subscribers {
			for _, ch := range channels {
				select {
				case ch <- message:
				default:
					// Channel full, skip
				}
			}
		}
	} else {
		// Send to specific recipient
		key := message.To + ":*"
		if channels, exists := r.subscribers[key]; exists {
			for _, ch := range channels {
				select {
				case ch <- message:
				default:
				}
			}
		}
	}

	return nil
}

// SharedWorkspace provides shared data storage for agents.
type SharedWorkspace struct {
	data map[string]interface{}
	mu   sync.RWMutex
}

// NewSharedWorkspace creates a new shared workspace.
func NewSharedWorkspace() *SharedWorkspace {
	return &SharedWorkspace{
		data: make(map[string]interface{}),
	}
}

// Put stores data in the workspace.
func (sw *SharedWorkspace) Put(key string, value interface{}) {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	sw.data[key] = value
}

// Get retrieves data from the workspace.
func (sw *SharedWorkspace) Get(key string) (interface{}, bool) {
	sw.mu.RLock()
	defer sw.mu.RUnlock()

	value, exists := sw.data[key]
	return value, exists
}

// Delete removes data from the workspace.
func (sw *SharedWorkspace) Delete(key string) {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	delete(sw.data, key)
}

// List lists all keys in the workspace.
func (sw *SharedWorkspace) List() []string {
	sw.mu.RLock()
	defer sw.mu.RUnlock()

	keys := make([]string, 0, len(sw.data))
	for key := range sw.data {
		keys = append(keys, key)
	}

	return keys
}

// Clear clears all data from the workspace.
func (sw *SharedWorkspace) Clear() {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	sw.data = make(map[string]interface{})
}
