package commands

import (
	"strings"
	"testing"
)

func TestBuildClaudeCodeIncludesOptionsAndScript(t *testing.T) {
	code, err := buildClaudeCode(&ClaudeSettings{Model: "Sonnet 4.6", PromptTimeoutSec: 42})
	if err != nil {
		t.Fatalf("buildClaudeCode returned error: %v", err)
	}
	for _, needle := range []string{`"action":"run"`, `"model":"Sonnet 4.6"`, `"promptTimeoutMs":42000`, `"thinkingMode":""`} {
		if !strings.Contains(code, needle) {
			t.Fatalf("missing %s in SURF_OPTIONS prelude: %q", needle, code[:min(240, len(code))])
		}
	}
	if !strings.Contains(code, "const CLAUDE_URL = 'https://claude.ai/new';") {
		t.Fatalf("missing embedded Claude script body")
	}
}

func TestClaudeDataToRowsExpandsModels(t *testing.T) {
	rows := claudeDataToRows(&claudeData{Raw: map[string]any{
		"kind":                "models",
		"href":                "https://claude.ai/new",
		"title":               "Claude",
		"currentModel":        "Sonnet 4.6",
		"currentThinkingMode": "extended",
		"models": []any{
			map[string]any{"name": "Opus 4.6", "description": "Most capable", "thinkingModes": []any{"standard", "extended"}},
			map[string]any{"name": "Sonnet 4.6", "description": "Everyday", "thinkingModes": []any{"standard", "extended"}},
		},
	}})
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
	if got, _ := rows[1].Get("selected"); got != true {
		t.Fatalf("expected selected row for current model, got %#v", got)
	}
	if got, _ := rows[0].Get("currentThinkingMode"); got != "extended" {
		t.Fatalf("expected currentThinkingMode row field, got %#v", got)
	}
}

func TestClaudeDataToRowsReturnsResponse(t *testing.T) {
	rows := claudeDataToRows(&claudeData{Raw: map[string]any{
		"kind":              "response",
		"href":              "https://claude.ai/chat/abc",
		"conversationTitle": "Hello",
		"response":          "Hello!",
	}})
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if got, _ := rows[0].Get("response"); got != "Hello!" {
		t.Fatalf("unexpected response field: %#v", got)
	}
}
