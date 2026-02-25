package commands

import "testing"

func TestExtractErrorText(t *testing.T) {
	resp := map[string]any{
		"type": "tool_response",
		"error": map[string]any{
			"content": []any{map[string]any{"type": "text", "text": "boom"}},
		},
	}
	if got := extractErrorText(resp); got != "boom" {
		t.Fatalf("unexpected error text: %q", got)
	}
}

func TestExtractResultText(t *testing.T) {
	resp := map[string]any{
		"type": "tool_response",
		"result": map[string]any{
			"content": []any{map[string]any{"type": "text", "text": "ok"}},
		},
	}
	if got := extractResultText(resp); got != "ok" {
		t.Fatalf("unexpected result text: %q", got)
	}
}
