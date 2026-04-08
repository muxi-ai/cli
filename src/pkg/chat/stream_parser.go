package chat

import (
	"bufio"
	"encoding/json"
	"io"
	"strings"
	"time"
)

var streamEventTimeout = 60 * time.Second

type sseFrame struct {
	Event   string
	Data    string
	Comment string
}

type routeErrorData struct {
	Error string `json:"error"`
	Type  string `json:"type"`
}

// StreamChatEvents parses the runtime chat SSE stream into chat StreamEvents.
func StreamChatEvents(reader io.Reader, callback func(StreamEvent) error) error {
	return streamChatEvents(reader, nil, callback)
}

func streamChatEvents(
	reader io.Reader,
	rawObserver func(string),
	callback func(StreamEvent) error,
) error {
	if callback == nil {
		return nil
	}

	err := parseSSEFrames(reader, rawObserver, func(frame sseFrame) error {
		switch {
		case frame.Comment != "":
			return callback(StreamEvent{
				Type:    "heartbeat",
				Content: frame.Comment,
			})
		case frame.Event == "done":
			return io.EOF
		case frame.Event == "error":
			return callback(StreamEvent{
				Type:    "error",
				Content: extractRouteErrorMessage(frame.Data),
			})
		}

		if frame.Data == "" {
			return nil
		}

		if isFinishedFrame(frame.Data) {
			return io.EOF
		}

		if event, ok := parseMuxiTokenEvent(frame.Data); ok {
			return callback(event)
		}

		if routeErr := extractRouteError(frame.Data); routeErr != "" {
			return callback(StreamEvent{
				Type:    "error",
				Content: routeErr,
			})
		}

		return nil
	})
	if err != nil && err != io.EOF {
		return err
	}
	return nil
}

func parseSSEFrames(
	reader io.Reader,
	rawObserver func(string),
	callback func(sseFrame) error,
) error {
	buffered := bufio.NewReader(reader)
	var currentEvent string
	var currentData strings.Builder

	flush := func() error {
		if currentEvent == "" && currentData.Len() == 0 {
			return nil
		}

		frame := sseFrame{
			Event: currentEvent,
			Data:  currentData.String(),
		}
		currentEvent = ""
		currentData.Reset()

		if callback == nil {
			return nil
		}
		return callback(frame)
	}

	for {
		line, err := buffered.ReadString('\n')
		if err != nil && err != io.EOF {
			return err
		}

		line = strings.TrimRight(line, "\r\n")
		if rawObserver != nil {
			rawObserver(line)
		}

		switch {
		case strings.HasPrefix(line, ":"):
			if callback != nil {
				if err := callback(sseFrame{
					Comment: strings.TrimSpace(strings.TrimPrefix(line, ":")),
				}); err != nil {
					return err
				}
			}
		case line == "":
			if err := flush(); err != nil {
				return err
			}
		case strings.HasPrefix(line, "event:"):
			currentEvent = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
		case strings.HasPrefix(line, "data:"):
			if currentData.Len() > 0 {
				currentData.WriteByte('\n')
			}
			currentData.WriteString(strings.TrimSpace(strings.TrimPrefix(line, "data:")))
		}

		if err == io.EOF {
			break
		}
	}

	return flush()
}

func parseMuxiTokenEvent(data string) (StreamEvent, bool) {
	var muxiToken MuxiToken
	if err := json.Unmarshal([]byte(data), &muxiToken); err != nil || muxiToken.Token.Type == "" {
		return StreamEvent{}, false
	}

	eventType := muxiToken.Token.Type
	stage := muxiToken.Token.Stage
	if eventType == "tool_call" {
		eventType = "progress"
		if stage == "" {
			stage = "tool_call"
		}
	}

	return StreamEvent{
		Type:      eventType,
		Stage:     stage,
		Content:   muxiToken.Token.Content,
		ToolName:  muxiToken.Token.ToolName,
		SessionID: muxiToken.Token.SessionID,
		RequestID: muxiToken.Token.RequestID,
		Artifacts: muxiToken.Token.Artifacts,
	}, true
}

func isFinishedFrame(data string) bool {
	var finished struct {
		Finished bool `json:"finished"`
	}
	return json.Unmarshal([]byte(data), &finished) == nil && finished.Finished
}

func extractRouteError(data string) string {
	var routeErr routeErrorData
	if json.Unmarshal([]byte(data), &routeErr) == nil && routeErr.Error != "" {
		return routeErr.Error
	}
	return ""
}

func extractRouteErrorMessage(data string) string {
	if routeErr := extractRouteError(data); routeErr != "" {
		return routeErr
	}
	return strings.TrimSpace(data)
}
