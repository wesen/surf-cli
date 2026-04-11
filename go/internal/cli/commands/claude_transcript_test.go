package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestClaudeTranscriptDataToRowsExpandsTranscript(t *testing.T) {
	rows := claudeTranscriptDataToRows(&claudeTranscriptData{Raw: map[string]any{
		"href":              "https://claude.ai/chat/abc",
		"title":             "Conversation - Claude",
		"conversationTitle": "Conversation",
		"currentModel":      "Sonnet 4.6",
		"turnCount":         2,
		"transcript": []any{
			map[string]any{"index": 0, "role": "user", "text": "hello"},
			map[string]any{"index": 1, "role": "assistant", "text": "world"},
		},
	}})
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
	if got, _ := rows[1].Get("currentModel"); got != "Sonnet 4.6" {
		t.Fatalf("unexpected currentModel: %#v", got)
	}
}

func TestRenderClaudeTranscriptMarkdown(t *testing.T) {
	text := renderClaudeTranscriptMarkdown(map[string]any{
		"href":                "https://claude.ai/chat/abc",
		"conversationTitle":   "Greeting",
		"currentModel":        "Sonnet 4.6",
		"currentThinkingMode": "extended",
		"transcript": []any{
			map[string]any{"index": 0, "role": "user", "text": "hello"},
			map[string]any{
				"index": 1,
				"role":  "assistant",
				"text":  "world",
				"citations": []any{
					map[string]any{"text": "MIT Press", "href": "https://mitpress.example"},
				},
				"searchWeb": map[string]any{
					"label":   "Searched the web",
					"queries": []any{"The Art of Insight Sanjoy Mahajan 10 results"},
					"results": []any{
						map[string]any{"text": "Online Textbook", "host": "ocw.mit.edu", "href": "https://ocw.example"},
					},
				},
			},
		},
	})
	if !strings.Contains(text, "# Greeting") || !strings.Contains(text, "## User 0") || !strings.Contains(text, "## Assistant 1") {
		t.Fatalf("unexpected markdown render: %s", text)
	}
	for _, needle := range []string{
		"- Thinking mode: extended",
		"### Citations",
		"MIT Press: https://mitpress.example",
		"### Searched The Web",
		"The Art of Insight Sanjoy Mahajan 10 results",
		"Online Textbook [ocw.mit.edu]: https://ocw.example",
	} {
		if !strings.Contains(text, needle) {
			t.Fatalf("expected markdown to contain %q, got: %s", needle, text)
		}
	}
}

func TestWriteClaudeTranscriptExportJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "claude.json")
	data := &claudeTranscriptData{Raw: map[string]any{"href": "https://claude.ai/chat/abc", "transcript": []any{}}}
	if err := writeClaudeTranscriptExport(path, "json", data); err != nil {
		t.Fatalf("writeClaudeTranscriptExport returned error: %v", err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read export: %v", err)
	}
	if !strings.Contains(string(b), `"href": "https://claude.ai/chat/abc"`) {
		t.Fatalf("unexpected export content: %s", string(b))
	}
}
