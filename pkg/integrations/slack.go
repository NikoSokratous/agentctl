package integrations

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// SlackConnector sends messages to Slack via Incoming Webhooks.
type SlackConnector struct {
	webhookURL string
	client     *http.Client
}

// NewSlackConnector creates a Slack connector.
func NewSlackConnector(cfg *Config) *SlackConnector {
	return &SlackConnector{
		webhookURL: cfg.WebhookURL,
		client:     &http.Client{Timeout: 10 * time.Second},
	}
}

// slackPayload is the Slack Incoming Webhook format.
type slackPayload struct {
	Text   string       `json:"text,omitempty"`
	Blocks []slackBlock `json:"blocks,omitempty"`
}

type slackBlock struct {
	Type string     `json:"type"`
	Text *slackText `json:"text,omitempty"`
}

type slackText struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// Send posts a message to Slack.
func (s *SlackConnector) Send(ctx context.Context, msg *Message) error {
	text := "*" + msg.Title + "*\n" + msg.Body
	for _, a := range msg.Actions {
		text += "\n" + a.Label + ": " + a.URL
	}
	payload := slackPayload{
		Blocks: []slackBlock{{Type: "section", Text: &slackText{Type: "mrkdwn", Text: text}}},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", s.webhookURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack webhook returned %d", resp.StatusCode)
	}
	return nil
}
