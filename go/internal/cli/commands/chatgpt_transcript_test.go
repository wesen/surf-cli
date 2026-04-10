package commands

import (
	"os"
	"path/filepath"
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

func TestWriteChatGPTTranscriptExportMarkdown(t *testing.T) {
	path := filepath.Join(t.TempDir(), "transcript.md")
	data := &chatGPTTranscriptData{Raw: map[string]any{
		"href":             "https://chatgpt.com/c/abc",
		"title":            "Conversation",
		"withActivity":     true,
		"activityExported": 1,
		"transcript": []any{
			map[string]any{"index": 0, "role": "assistant", "messageId": "m1", "model": "gpt-5", "text": "hello", "thoughtButtonText": "Thought for 5s", "activityText": "reasoning"},
		},
	}}
	if err := writeChatGPTTranscriptExport(path, "markdown", data); err != nil {
		t.Fatalf("writeChatGPTTranscriptExport returned error: %v", err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read markdown export: %v", err)
	}
	text := string(b)
	if !strings.Contains(text, "# Conversation") || !strings.Contains(text, "### Activity") {
		t.Fatalf("unexpected markdown export: %s", text)
	}
}

func TestWriteChatGPTTranscriptExportJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "transcript.json")
	data := &chatGPTTranscriptData{Raw: map[string]any{"href": "https://chatgpt.com/c/abc", "transcript": []any{}}}
	if err := writeChatGPTTranscriptExport(path, "json", data); err != nil {
		t.Fatalf("writeChatGPTTranscriptExport returned error: %v", err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read json export: %v", err)
	}
	if !strings.Contains(string(b), `"href": "https://chatgpt.com/c/abc"`) {
		t.Fatalf("unexpected json export: %s", string(b))
	}
}

func TestRenderChatGPTTranscriptMarkdown(t *testing.T) {
	text := renderChatGPTTranscriptMarkdown(map[string]any{
		"href":             "https://chatgpt.com/c/abc",
		"title":            "Conversation",
		"withActivity":     true,
		"activityExported": 1,
		"transcript": []any{
			map[string]any{
				"index":             0,
				"role":              "assistant",
				"messageId":         "m1",
				"model":             "gpt-5",
				"text":              "hello",
				"thoughtButtonText": "Thought for 5s",
				"activityText":      "reasoning",
			},
		},
	})
	if !strings.Contains(text, "# Conversation") || !strings.Contains(text, "### Activity") || !strings.Contains(text, "Thought for 5s") {
		t.Fatalf("unexpected markdown render: %s", text)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
