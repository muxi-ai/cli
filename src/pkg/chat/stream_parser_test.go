package chat

import (
	"io"
	"strings"
	"testing"
	"time"
)

func TestProcessStreamWithEventsEmitsHeartbeatAndCompleted(t *testing.T) {
	input := strings.Join([]string{
		": keepalive",
		"",
		`data: {"token":{"type":"content","content":"Hello "}}`,
		"",
		`data: {"token":{"type":"completed","content":"Hello world","session_id":"sess-1","request_id":"req-1"}}`,
		"",
	}, "\n")

	eventChan := make(chan StreamEvent, 10)
	go processStreamWithEvents(io.NopCloser(strings.NewReader(input)), eventChan, false)

	var events []StreamEvent
	for event := range eventChan {
		events = append(events, event)
	}

	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
	if events[0].Type != "heartbeat" {
		t.Fatalf("expected first event to be heartbeat, got %q", events[0].Type)
	}
	if events[1].Type != "completed" {
		t.Fatalf("expected second event to be completed, got %q", events[1].Type)
	}
	if events[1].Content != "Hello world" {
		t.Fatalf("expected completed content %q, got %q", "Hello world", events[1].Content)
	}
	if events[1].SessionID != "sess-1" || events[1].RequestID != "req-1" {
		t.Fatalf("expected session/request IDs to be preserved, got session=%q request=%q", events[1].SessionID, events[1].RequestID)
	}
}

func TestProcessStreamWithEventsRendersUIWidgets(t *testing.T) {
	input := strings.Join([]string{
		`data: {"token":{"type":"content","content":"Pick one."}}`,
		"",
		"event: ui",
		`data: {"ui":[{"type":"options","id":"w1","prompt":"Which region?","options":[{"value":"us","label":"United States"},{"value":"eu","label":"Europe"}]},{"type":"action_link","id":"w2","label":"Open dashboard","url":"https://example.com/dash"},{"type":"hologram","id":"w3"}]}`,
		"",
		`event: done`,
		`data: {"finished":true}`,
		"",
	}, "\n")

	eventChan := make(chan StreamEvent, 10)
	go processStreamWithEvents(io.NopCloser(strings.NewReader(input)), eventChan, false)

	var events []StreamEvent
	for event := range eventChan {
		events = append(events, event)
	}

	if len(events) != 1 {
		t.Fatalf("expected 1 completed event, got %d: %+v", len(events), events)
	}
	content := events[0].Content
	if !strings.Contains(content, "Pick one.") {
		t.Fatalf("expected response text preserved, got %q", content)
	}
	if !strings.Contains(content, "Which region?") || !strings.Contains(content, "1. United States") || !strings.Contains(content, "2. Europe") {
		t.Errorf("expected rendered options widget, got %q", content)
	}
	if !strings.Contains(content, "[Open dashboard](https://example.com/dash)") {
		t.Errorf("expected rendered action_link widget, got %q", content)
	}
	if strings.Contains(content, "hologram") {
		t.Errorf("unknown widget types must be skipped, got %q", content)
	}
}

func TestProcessStreamWithEventsParsesRouteLevelError(t *testing.T) {
	input := strings.Join([]string{
		"event: error",
		`data: {"error":"backend exploded","type":"RuntimeError"}`,
		"",
	}, "\n")

	eventChan := make(chan StreamEvent, 10)
	go processStreamWithEvents(io.NopCloser(strings.NewReader(input)), eventChan, false)

	var events []StreamEvent
	for event := range eventChan {
		events = append(events, event)
	}

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Type != "error" {
		t.Fatalf("expected error event, got %q", events[0].Type)
	}
	if events[0].Content != "backend exploded" {
		t.Fatalf("expected error content %q, got %q", "backend exploded", events[0].Content)
	}
}

func TestStreamChatEventsNormalizesToolCallType(t *testing.T) {
	input := strings.Join([]string{
		`data: {"token":{"type":"tool_call","tool_name":"get-current-user","content":"Using the ms365 tool..."}}`,
		"",
	}, "\n")

	var events []StreamEvent
	err := StreamChatEvents(strings.NewReader(input), func(event StreamEvent) error {
		events = append(events, event)
		return nil
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Type != "progress" {
		t.Fatalf("expected normalized event type %q, got %q", "progress", events[0].Type)
	}
	if events[0].Stage != "tool_call" {
		t.Fatalf("expected normalized stage %q, got %q", "tool_call", events[0].Stage)
	}
	if events[0].ToolName != "get-current-user" {
		t.Fatalf("expected tool name to be preserved, got %q", events[0].ToolName)
	}
}

func TestWaitForEventFromChanTreatsHeartbeatAsActivity(t *testing.T) {
	originalTimeout := streamEventTimeout
	streamEventTimeout = 100 * time.Millisecond
	defer func() {
		streamEventTimeout = originalTimeout
	}()

	eventChan := make(chan StreamEvent, 1)
	go func() {
		time.Sleep(10 * time.Millisecond)
		eventChan <- StreamEvent{Type: "heartbeat"}
	}()

	msg := waitForEventFromChan(eventChan)
	eventMsg, ok := msg.(streamEventMsg)
	if !ok {
		t.Fatalf("expected streamEventMsg, got %T", msg)
	}
	if eventMsg.event.Type != "heartbeat" {
		t.Fatalf("expected heartbeat event, got %q", eventMsg.event.Type)
	}
}
