package memory

import (
	"context"
	"encoding/json"
)

// PersistentStore is the interface for persistent key-value storage.
type PersistentStore interface {
	Get(ctx context.Context, agentID, key string) ([]byte, error)
	Set(ctx context.Context, agentID, key string, value []byte) error
	Delete(ctx context.Context, agentID, key string) error
	ListKeys(ctx context.Context, agentID string) ([]string, error)
	DeleteAgent(ctx context.Context, agentID string) error
}

// PersistentMemory wraps a PersistentStore with agent scoping.
type PersistentMemory struct {
	store   PersistentStore
	agentID string
}

// NewPersistentMemory creates a persistent memory for an agent.
func NewPersistentMemory(store PersistentStore, agentID string) *PersistentMemory {
	return &PersistentMemory{store: store, agentID: agentID}
}

// Get retrieves a value by key.
func (p *PersistentMemory) Get(ctx context.Context, key string, dest any) error {
	b, err := p.store.Get(ctx, p.agentID, key)
	if err != nil {
		return err
	}
	if b == nil {
		return nil
	}
	return json.Unmarshal(b, dest)
}

// Set stores a value.
func (p *PersistentMemory) Set(ctx context.Context, key string, value any) error {
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return p.store.Set(ctx, p.agentID, key, b)
}

// Delete removes a key.
func (p *PersistentMemory) Delete(ctx context.Context, key string) error {
	return p.store.Delete(ctx, p.agentID, key)
}
