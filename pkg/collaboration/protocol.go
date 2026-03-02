package collaboration

import (
	"encoding/json"
	"fmt"
	"time"
)

// MessageType defines the type of agent message.
type MessageType string

const (
	MessageTypeRequest   MessageType = "request"
	MessageTypeResponse  MessageType = "response"
	MessageTypeBroadcast MessageType = "broadcast"
	MessageTypeNotify    MessageType = "notify"
)

// ActionType defines the action in a message.
type ActionType string

const (
	ActionDelegateTask ActionType = "delegate_task"
	ActionShareData    ActionType = "share_data"
	ActionRequestHelp  ActionType = "request_help"
	ActionVote         ActionType = "vote"
	ActionAnnounce     ActionType = "announce"
	ActionQuery        ActionType = "query"
)

// AgentMessage represents a message between agents.
type AgentMessage struct {
	ID        string          `json:"id"`
	From      string          `json:"from"`
	To        string          `json:"to"` // Empty for broadcast
	Type      MessageType     `json:"type"`
	Action    ActionType      `json:"action"`
	Payload   json.RawMessage `json:"payload"`
	Timestamp time.Time       `json:"timestamp"`
	ReplyTo   string          `json:"reply_to,omitempty"`
	Priority  int             `json:"priority"` // 0=low, 1=normal, 2=high
}

// TaskDelegation represents a task delegation request.
type TaskDelegation struct {
	TaskID               string                 `json:"task_id"`
	Description          string                 `json:"description"`
	Goal                 string                 `json:"goal"`
	Context              map[string]interface{} `json:"context"`
	Deadline             *time.Time             `json:"deadline,omitempty"`
	RequiredCapabilities []string               `json:"required_capabilities,omitempty"`
}

// DataShare represents shared data between agents.
type DataShare struct {
	DataID      string                 `json:"data_id"`
	Description string                 `json:"description"`
	Data        map[string]interface{} `json:"data"`
	SchemaType  string                 `json:"schema_type,omitempty"`
	ExpiresAt   *time.Time             `json:"expires_at,omitempty"`
}

// VoteRequest represents a voting request.
type VoteRequest struct {
	VoteID        string    `json:"vote_id"`
	Question      string    `json:"question"`
	Options       []string  `json:"options"`
	Deadline      time.Time `json:"deadline"`
	RequiredVotes int       `json:"required_votes"`
}

// Vote represents an agent's vote.
type Vote struct {
	VoteID    string    `json:"vote_id"`
	AgentID   string    `json:"agent_id"`
	Choice    string    `json:"choice"`
	Timestamp time.Time `json:"timestamp"`
	Reason    string    `json:"reason,omitempty"`
}

// Announcement represents a broadcast announcement.
type Announcement struct {
	Subject string                 `json:"subject"`
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data,omitempty"`
	Urgent  bool                   `json:"urgent"`
}

// Query represents a query request.
type Query struct {
	QueryID  string                 `json:"query_id"`
	Question string                 `json:"question"`
	Context  map[string]interface{} `json:"context,omitempty"`
	Timeout  time.Duration          `json:"timeout"`
}

// QueryResponse represents a query response.
type QueryResponse struct {
	QueryID    string                 `json:"query_id"`
	Answer     string                 `json:"answer"`
	Confidence float64                `json:"confidence"` // 0.0 to 1.0
	Data       map[string]interface{} `json:"data,omitempty"`
}

// NewMessage creates a new agent message.
func NewMessage(from, to string, msgType MessageType, action ActionType, payload interface{}) (*AgentMessage, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	return &AgentMessage{
		ID:        generateMessageID(),
		From:      from,
		To:        to,
		Type:      msgType,
		Action:    action,
		Payload:   payloadBytes,
		Timestamp: time.Now(),
		Priority:  1, // Normal priority by default
	}, nil
}

// ParsePayload parses the message payload into the given type.
func (m *AgentMessage) ParsePayload(v interface{}) error {
	return json.Unmarshal(m.Payload, v)
}

// Reply creates a reply to this message.
func (m *AgentMessage) Reply(payload interface{}) (*AgentMessage, error) {
	reply, err := NewMessage(m.To, m.From, MessageTypeResponse, m.Action, payload)
	if err != nil {
		return nil, err
	}

	reply.ReplyTo = m.ID
	return reply, nil
}

// IsBroadcast checks if message is a broadcast.
func (m *AgentMessage) IsBroadcast() bool {
	return m.To == "" || m.To == "*"
}

// IsExpired checks if message has expired (for time-sensitive messages).
func (m *AgentMessage) IsExpired(ttl time.Duration) bool {
	return time.Since(m.Timestamp) > ttl
}

// generateMessageID generates a unique message ID.
func generateMessageID() string {
	return fmt.Sprintf("msg-%d", time.Now().UnixNano())
}

// MessagePriority sets message priority.
func (m *AgentMessage) SetPriority(priority int) {
	if priority < 0 {
		priority = 0
	}
	if priority > 2 {
		priority = 2
	}
	m.Priority = priority
}

// Validate validates the message structure.
func (m *AgentMessage) Validate() error {
	if m.From == "" {
		return fmt.Errorf("from field is required")
	}
	if m.Type == "" {
		return fmt.Errorf("type field is required")
	}
	if m.Action == "" {
		return fmt.Errorf("action field is required")
	}
	return nil
}
