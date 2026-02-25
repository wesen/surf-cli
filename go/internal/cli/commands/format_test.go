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
	if got := parseResult(resp).Text; got != "ok" {
		t.Fatalf("unexpected result text: %q", got)
	}
}

func TestParseStructuredTextObject(t *testing.T) {
	resp := map[string]any{
		"type": "tool_response",
		"result": map[string]any{
			"content": []any{map[string]any{"type": "text", "text": `{"title":"Example","url":"https://example.com"}`}},
		},
	}
	data := parseResult(resp).Data
	m, ok := data.(map[string]any)
	if !ok {
		t.Fatalf("expected object data, got %T", data)
	}
	if m["title"] != "Example" {
		t.Fatalf("unexpected parsed data: %#v", m)
	}
}

func TestParseStructuredTextArray(t *testing.T) {
	resp := map[string]any{
		"type": "tool_response",
		"result": map[string]any{
			"content": []any{map[string]any{"type": "text", "text": `[{"id":1},{"id":2}]`}},
		},
	}
	data := parseResult(resp).Data
	a, ok := data.([]any)
	if !ok || len(a) != 2 {
		t.Fatalf("expected array data of len 2, got %#v", data)
	}
}
