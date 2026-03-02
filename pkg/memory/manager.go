package memory

import (
	"context"
	"sync"
)

// Manager coordinates working, persistent, and semantic memory tiers.
type Manager struct {
	working    *WorkingMemory
	persistent PersistentStore
	semantic   SemanticStore
	agentID    string
	mu         sync.RWMutex
}

// NewManager creates a memory manager for an agent.
func NewManager(agentID string, persistent PersistentStore, semantic SemanticStore) *Manager {
	return &Manager{
		working:    NewWorkingMemory(),
		persistent: persistent,
		semantic:   semantic,
		agentID:    agentID,
	}
}

// Working returns the working memory for this session.
func (m *Manager) Working() *WorkingMemory {
	return m.working
}

// Persistent returns a PersistentMemory scoped to the agent.
func (m *Manager) Persistent() *PersistentMemory {
	return NewPersistentMemory(m.persistent, m.agentID)
}

// Semantic returns the semantic store (may be nil).
func (m *Manager) Semantic() SemanticStore {
	return m.semantic
}

// ClearSession clears working memory at session end.
func (m *Manager) ClearSession() {
	m.working.Clear()
}

// DeleteAgent removes all non-log data for the agent (GDPR compliance).
func (m *Manager) DeleteAgent(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.working.Clear()
	if m.persistent != nil {
		if err := m.persistent.DeleteAgent(ctx, m.agentID); err != nil {
			return err
		}
	}
	if m.semantic != nil {
		if err := m.semantic.DeleteAgent(ctx, m.agentID); err != nil {
			return err
		}
	}
	return nil
}
