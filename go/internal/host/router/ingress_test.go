package router

import "testing"

func TestParseToolRequestValid(t *testing.T) {
	msg := map[string]any{
		"type":   "tool_request",
		"method": "execute_tool",
		"params": map[string]any{
			"tool": "page.read",
			"args": map[string]any{"filter": "interactive"},
		},
	}
	req, err := ParseToolRequest(msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Params.Tool != "page.read" {
		t.Fatalf("unexpected tool: %q", req.Params.Tool)
	}
}

func TestParseToolRequestInvalid(t *testing.T) {
	_, err := ParseToolRequest(map[string]any{"type": "tool_request"})
	if err == nil {
		t.Fatalf("expected validation error")
	}
}

func TestParseStreamRequestValid(t *testing.T) {
	msg := map[string]any{
		"type":       "stream_request",
		"streamType": "STREAM_NETWORK",
		"options":    map[string]any{"limit": 20},
	}
	req, err := ParseStreamRequest(msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.StreamType != "STREAM_NETWORK" {
		t.Fatalf("unexpected streamType: %q", req.StreamType)
	}
}

func TestParseStreamRequestInvalidType(t *testing.T) {
	_, err := ParseStreamRequest(map[string]any{
		"type":       "stream_request",
		"streamType": "STREAM_UNKNOWN",
	})
	if err == nil {
		t.Fatalf("expected validation error")
	}
}

func TestParseStreamStop(t *testing.T) {
	if err := ParseStreamStop(map[string]any{"type": "stream_stop"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := ParseStreamStop(map[string]any{"type": "stream_request"}); err == nil {
		t.Fatalf("expected validation error")
	}
}
