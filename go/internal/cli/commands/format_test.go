package commands

import (
	"testing"

	"github.com/go-go-golems/glazed/pkg/types"
)

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

func TestToolResponseToRowsUsesParsedObjectRow(t *testing.T) {
	resp := map[string]any{
		"type": "tool_response",
		"result": map[string]any{
			"content": []any{map[string]any{"type": "text", "text": `{"title":"Example","url":"https://example.com"}`}},
		},
	}
	rows := ToolResponseToRows(resp)
	if len(rows) != 1 {
		t.Fatalf("expected one row, got %d", len(rows))
	}
	row := rows[0]
	fields := types.GetFields(row)
	if len(fields) != 2 || fields[0] != "title" || fields[1] != "url" {
		t.Fatalf("unexpected row fields: %#v", fields)
	}
}

func TestToolResponseToRowsExpandsArrayToRows(t *testing.T) {
	resp := map[string]any{
		"type": "tool_response",
		"result": map[string]any{
			"content": []any{map[string]any{"type": "text", "text": `[{"id":1},{"id":2}]`}},
		},
	}
	rows := ToolResponseToRows(resp)
	if len(rows) != 2 {
		t.Fatalf("expected two rows, got %d", len(rows))
	}
	v0, ok := rows[0].Get("id")
	if !ok {
		t.Fatalf("expected id field in first row")
	}
	v1, ok := rows[1].Get("id")
	if !ok {
		t.Fatalf("expected id field in second row")
	}
	if v0 != float64(1) || v1 != float64(2) {
		t.Fatalf("unexpected row values: %#v %#v", v0, v1)
	}
}

func TestToolResponseToRowsFallsBackToContentText(t *testing.T) {
	resp := map[string]any{
		"type": "tool_response",
		"result": map[string]any{
			"content": []any{map[string]any{"type": "text", "text": "OK"}},
		},
	}
	rows := ToolResponseToRows(resp)
	if len(rows) != 1 {
		t.Fatalf("expected one row, got %d", len(rows))
	}
	value, ok := rows[0].Get("content")
	if !ok {
		t.Fatalf("expected content field in row")
	}
	if s, ok := value.(string); !ok || s != "OK" {
		t.Fatalf("expected text fallback in content field, got %#v", value)
	}
}
