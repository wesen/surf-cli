package commands

import (
	"strings"
	"testing"
)

func TestBuildKagiSearchCodeIncludesOptionsAndScript(t *testing.T) {
	code, err := buildKagiSearchCode(&KagiSearchSettings{MaxResults: 7})
	if err != nil {
		t.Fatalf("buildKagiSearchCode returned error: %v", err)
	}
	if !strings.Contains(code, `const SURF_OPTIONS = {"maxResults":7};`) {
		t.Fatalf("missing SURF_OPTIONS prelude: %q", code[:min(160, len(code))])
	}
	if !strings.Contains(code, "const RESULTS_SELECTOR = 'main div._0_SRI.search-result, main div.__srgi';") {
		t.Fatalf("missing embedded kagi search script body")
	}
}

func TestBuildKagiSearchURL(t *testing.T) {
	got := buildKagiSearchURL("llm transcript attribution")
	want := "https://kagi.com/search?q=llm+transcript+attribution"
	if got != want {
		t.Fatalf("unexpected kagi search URL: got %q want %q", got, want)
	}
}

func TestKagiSearchDataToRowsExpandsResults(t *testing.T) {
	data := &kagiSearchData{Raw: map[string]any{
		"query":       "llm transcript attribution",
		"href":        "https://kagi.com/search?q=llm+transcript+attribution",
		"title":       "llm transcript attribution - Kagi Search",
		"resultCount": 2,
		"waitedMs":    500,
		"maxResults":  10,
		"quickAnswer": map[string]any{"title": "Quick Answer", "text": "A summary"},
		"results": []any{
			map[string]any{"index": 1, "title": "Paper A", "url": "https://example.com/a", "snippet": "Snippet A"},
			map[string]any{"index": 2, "title": "Paper B", "url": "https://example.com/b", "snippet": "Snippet B"},
		},
	}}
	rows := kagiSearchDataToRows(data)
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
	if got, _ := rows[0].Get("query"); got != "llm transcript attribution" {
		t.Fatalf("unexpected query: %#v", got)
	}
	if got, _ := rows[1].Get("quickAnswerText"); got != "A summary" {
		t.Fatalf("unexpected quickAnswerText: %#v", got)
	}
}

func TestRenderKagiSearchMarkdown(t *testing.T) {
	text := renderKagiSearchMarkdown(map[string]any{
		"query":       "llm transcript attribution",
		"href":        "https://kagi.com/search?q=llm+transcript+attribution",
		"title":       "llm transcript attribution - Kagi Search",
		"resultCount": 1,
		"quickAnswer": map[string]any{"title": "Quick Answer", "text": "A summary"},
		"results": []any{
			map[string]any{
				"index":      1,
				"title":      "Paper A",
				"url":        "https://example.com/a",
				"displayUrl": "example.com/a",
				"snippet":    "Snippet A",
			},
		},
	})
	if !strings.Contains(text, "# Kagi Search") || !strings.Contains(text, "## Quick Answer") || !strings.Contains(text, "### 1. [Paper A](https://example.com/a)") {
		t.Fatalf("unexpected markdown render: %s", text)
	}
}

func TestExtractTabIDFromResponse(t *testing.T) {
	resp := map[string]any{
		"type": "tool_response",
		"result": map[string]any{
			"content": []any{map[string]any{"type": "text", "text": `{"success":true,"tabId":42,"url":"https://kagi.com/"}`}},
		},
	}
	tabID, err := extractTabIDFromResponse(resp)
	if err != nil {
		t.Fatalf("extractTabIDFromResponse returned error: %v", err)
	}
	if tabID != 42 {
		t.Fatalf("unexpected tab id: %d", tabID)
	}
}
