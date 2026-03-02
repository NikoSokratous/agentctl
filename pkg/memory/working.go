package memory

import (
	"sync"
)

// WorkingMemory holds in-memory session state, cleared when session ends.
type WorkingMemory struct {
	mu   sync.RWMutex
	data map[string]any
}

// NewWorkingMemory creates a new working memory store.
func NewWorkingMemory() *WorkingMemory {
	return &WorkingMemory{data: make(map[string]any)}
}

// Get retrieves a value by key.
func (w *WorkingMemory) Get(key string) (any, bool) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	v, ok := w.data[key]
	return v, ok
}

// Set stores a value.
func (w *WorkingMemory) Set(key string, value any) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.data[key] = value
}

// Delete removes a key.
func (w *WorkingMemory) Delete(key string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	delete(w.data, key)
}

// Clear removes all entries (call on session end).
func (w *WorkingMemory) Clear() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.data = make(map[string]any)
}

// All returns a copy of all data.
func (w *WorkingMemory) All() map[string]any {
	w.mu.RLock()
	defer w.mu.RUnlock()
	out := make(map[string]any)
	for k, v := range w.data {
		out[k] = v
	}
	return out
}
