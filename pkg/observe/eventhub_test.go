package observe

import (
	"testing"
	"time"
)

func TestEventHubSubscribe(t *testing.T) {
	hub := NewEventHub(10)
	runID := "test-run-1"

	ch := hub.Subscribe(runID)
	if ch == nil {
		t.Fatal("Subscribe returned nil channel")
	}

	if count := hub.SubscriberCount(runID); count != 1 {
		t.Errorf("SubscriberCount = %d, want 1", count)
	}
}

func TestEventHubPublish(t *testing.T) {
	hub := NewEventHub(10)
	runID := "test-run-2"

	ch := hub.Subscribe(runID)

	event := Event{
		Timestamp: time.Now(),
		Type:      EventInit,
		RunID:     runID,
		Agent:     "test-agent",
		Data:      map[string]any{"message": "test message"},
	}

	hub.Publish(runID, event)

	select {
	case received := <-ch:
		if received.Type != event.Type {
			t.Errorf("Received type = %s, want %s", received.Type, event.Type)
		}
		if received.RunID != runID {
			t.Errorf("Received runID = %s, want %s", received.RunID, runID)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Timeout waiting for event")
	}
}

func TestEventHubMultipleSubscribers(t *testing.T) {
	hub := NewEventHub(10)
	runID := "test-run-3"

	ch1 := hub.Subscribe(runID)
	ch2 := hub.Subscribe(runID)

	if count := hub.SubscriberCount(runID); count != 2 {
		t.Errorf("SubscriberCount = %d, want 2", count)
	}

	event := Event{
		Timestamp: time.Now(),
		Type:      EventInit,
		RunID:     runID,
		Agent:     "test-agent",
		Data:      map[string]any{"message": "broadcast"},
	}

	hub.Publish(runID, event)

	// Both subscribers should receive
	select {
	case <-ch1:
	case <-time.After(100 * time.Millisecond):
		t.Error("ch1 timeout")
	}

	select {
	case <-ch2:
	case <-time.After(100 * time.Millisecond):
		t.Error("ch2 timeout")
	}
}

func TestEventHubUnsubscribe(t *testing.T) {
	hub := NewEventHub(10)
	runID := "test-run-4"

	ch := hub.Subscribe(runID)
	hub.Unsubscribe(runID, ch)

	if count := hub.SubscriberCount(runID); count != 0 {
		t.Errorf("SubscriberCount after unsubscribe = %d, want 0", count)
	}
}

func TestEventHubUnsubscribeAll(t *testing.T) {
	hub := NewEventHub(10)
	runID := "test-run-5"

	hub.Subscribe(runID)
	hub.Subscribe(runID)
	hub.Subscribe(runID)

	hub.UnsubscribeAll(runID)

	if count := hub.SubscriberCount(runID); count != 0 {
		t.Errorf("SubscriberCount after UnsubscribeAll = %d, want 0", count)
	}
}

func TestEventHubBufferOverflow(t *testing.T) {
	hub := NewEventHub(2) // Small buffer
	runID := "test-run-6"

	ch := hub.Subscribe(runID)

	// Fill buffer + extra (should not block)
	for i := 0; i < 5; i++ {
		event := Event{
			Timestamp: time.Now(),
			Type:      EventInit,
			RunID:     runID,
			Agent:     "test-agent",
			Data:      map[string]any{"message": "overflow test"},
		}
		hub.Publish(runID, event)
	}

	// Should receive at least buffer size
	count := 0
	timeout := time.After(100 * time.Millisecond)

	for {
		select {
		case <-ch:
			count++
		case <-timeout:
			if count < 2 {
				t.Errorf("Received %d events, expected at least 2", count)
			}
			return
		}
	}
}
