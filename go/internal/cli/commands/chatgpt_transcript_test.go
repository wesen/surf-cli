package commands

import (
	"strings"
	"testing"
)

func TestBuildChatGPTTranscriptCodeIncludesOptionsAndScript(t *testing.T) {
	code, err := buildChatGPTTranscriptCode(&ChatGPTTranscriptSettings{WithActivity: true, ActivityLimit: 3})
	if err != nil {
		t.Fatalf("buildChatGPTTranscriptCode returned error: %v", err)
	}
	if !strings.Contains(code, `const SURF_OPTIONS = {"activityLimit":3,"withActivity":true};`) && !strings.Contains(code, `const SURF_OPTIONS = {"withActivity":true,"activityLimit":3};`) {
		t.Fatalf("missing SURF_OPTIONS prelude: %q", code[:min(160, len(code))])
	}
	if !strings.Contains(code, "const sections = Array.from(document.querySelectorAll('section[data-testid^=\"conversation-turn-\"]'));") {
		t.Fatalf("missing embedded transcript script body")
	}
}

func TestChatGPTTranscriptResponseToRowsExpandsTranscript(t *testing.T) {
	resp := map[string]any{
		"type": "tool_response",
		"result": map[string]any{
			"content": []any{map[string]any{"type": "text", "text": `{"href":"https://chatgpt.com/c/abc","title":"Conversation","turnCount":2,"withActivity":true,"activityLimit":3,"activityExported":1,"transcript":[{"index":0,"role":"user","text":"hello"},{"index":1,"role":"assistant","text":"world","activityFound":true,"activityText":"thoughts"}]}`}},
		},
	}
	rows, err := chatGPTTranscriptResponseToRows(resp)
	if err != nil {
		t.Fatalf("chatGPTTranscriptResponseToRows returned error: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
	if got, _ := rows[0].Get("href"); got != "https://chatgpt.com/c/abc" {
		t.Fatalf("unexpected href: %#v", got)
	}
	if got, _ := rows[1].Get("activityText"); got != "thoughts" {
		t.Fatalf("unexpected activity text: %#v", got)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
