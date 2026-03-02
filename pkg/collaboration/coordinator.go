package collaboration

import (
	"context"
	"fmt"
	"sync"
)

// Coordinator manages agent collaboration.
type Coordinator struct {
	agents    map[string]*AgentInfo
	workspace *SharedWorkspace
	router    *MessageRouter
	mu        sync.RWMutex
}

// AgentInfo contains agent metadata.
type AgentInfo struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Capabilities []string               `json:"capabilities"`
	Status       string                 `json:"status"` // active, idle, busy
	Metadata     map[string]interface{} `json:"metadata"`
}

// NewCoordinator creates a new collaboration coordinator.
func NewCoordinator() *Coordinator {
	return &Coordinator{
		agents:    make(map[string]*AgentInfo),
		workspace: NewSharedWorkspace(),
		router:    NewMessageRouter(),
	}
}

// RegisterAgent registers an agent with the coordinator.
func (c *Coordinator) RegisterAgent(ctx context.Context, agent *AgentInfo) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.agents[agent.ID]; exists {
		return fmt.Errorf("agent %s already registered", agent.ID)
	}

	c.agents[agent.ID] = agent

	// Subscribe to broadcast messages
	c.router.Subscribe(agent.ID, "*")

	return nil
}

// UnregisterAgent unregisters an agent.
func (c *Coordinator) UnregisterAgent(ctx context.Context, agentID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.agents, agentID)
	c.router.Unsubscribe(agentID)

	return nil
}

// SendMessage sends a message between agents.
func (c *Coordinator) SendMessage(ctx context.Context, message *AgentMessage) error {
	if err := message.Validate(); err != nil {
		return fmt.Errorf("invalid message: %w", err)
	}

	return c.router.Route(ctx, message)
}

// BroadcastMessage broadcasts a message to all agents.
func (c *Coordinator) BroadcastMessage(ctx context.Context, from string, action ActionType, payload interface{}) error {
	message, err := NewMessage(from, "*", MessageTypeBroadcast, action, payload)
	if err != nil {
		return err
	}

	return c.router.Route(ctx, message)
}

// DelegateTask delegates a task to another agent.
func (c *Coordinator) DelegateTask(ctx context.Context, from, to string, task *TaskDelegation) error {
	// Find suitable agent if 'to' is empty
	if to == "" {
		agent, err := c.findCapableAgent(task.RequiredCapabilities)
		if err != nil {
			return err
		}
		to = agent.ID
	}

	message, err := NewMessage(from, to, MessageTypeRequest, ActionDelegateTask, task)
	if err != nil {
		return err
	}

	return c.router.Route(ctx, message)
}

// ShareData shares data with specific agent or all agents.
func (c *Coordinator) ShareData(ctx context.Context, from, to string, data *DataShare) error {
	// Store in shared workspace
	c.workspace.Put(data.DataID, data.Data)

	// Notify recipient(s)
	msgType := MessageTypeNotify
	if to == "" || to == "*" {
		msgType = MessageTypeBroadcast
		to = "*"
	}

	message, err := NewMessage(from, to, msgType, ActionShareData, data)
	if err != nil {
		return err
	}

	return c.router.Route(ctx, message)
}

// RequestVote initiates a voting process.
func (c *Coordinator) RequestVote(ctx context.Context, from string, voteReq *VoteRequest) (*VoteResult, error) {
	// Broadcast vote request
	message, err := NewMessage(from, "*", MessageTypeBroadcast, ActionVote, voteReq)
	if err != nil {
		return nil, err
	}

	if err := c.router.Route(ctx, message); err != nil {
		return nil, err
	}

	// Collect votes (placeholder - would wait for responses)
	result := &VoteResult{
		VoteID: voteReq.VoteID,
		Votes:  make([]Vote, 0),
	}

	return result, nil
}

// VoteResult represents voting results.
type VoteResult struct {
	VoteID  string
	Votes   []Vote
	Winner  string
	Decided bool
}

// QueryAgents queries all agents with a question.
func (c *Coordinator) QueryAgents(ctx context.Context, from string, query *Query) ([]QueryResponse, error) {
	message, err := NewMessage(from, "*", MessageTypeBroadcast, ActionQuery, query)
	if err != nil {
		return nil, err
	}

	if err := c.router.Route(ctx, message); err != nil {
		return nil, err
	}

	// Placeholder: would collect responses
	return []QueryResponse{}, nil
}

// GetAgent retrieves agent information.
func (c *Coordinator) GetAgent(agentID string) (*AgentInfo, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	agent, exists := c.agents[agentID]
	if !exists {
		return nil, fmt.Errorf("agent not found: %s", agentID)
	}

	return agent, nil
}

// ListAgents lists all registered agents.
func (c *Coordinator) ListAgents() []AgentInfo {
	c.mu.RLock()
	defer c.mu.RUnlock()

	agents := make([]AgentInfo, 0, len(c.agents))
	for _, agent := range c.agents {
		agents = append(agents, *agent)
	}

	return agents
}

// findCapableAgent finds an agent with required capabilities.
func (c *Coordinator) findCapableAgent(requiredCapabilities []string) (*AgentInfo, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, agent := range c.agents {
		if agent.Status == "busy" {
			continue
		}

		if hasCapabilities(agent.Capabilities, requiredCapabilities) {
			return agent, nil
		}
	}

	return nil, fmt.Errorf("no capable agent found")
}

// hasCapabilities checks if agent has all required capabilities.
func hasCapabilities(agentCaps, required []string) bool {
	if len(required) == 0 {
		return true
	}

	capSet := make(map[string]bool)
	for _, cap := range agentCaps {
		capSet[cap] = true
	}

	for _, req := range required {
		if !capSet[req] {
			return false
		}
	}

	return true
}

// GetWorkspace returns the shared workspace.
func (c *Coordinator) GetWorkspace() *SharedWorkspace {
	return c.workspace
}

// GetRouter returns the message router.
func (c *Coordinator) GetRouter() *MessageRouter {
	return c.router
}
