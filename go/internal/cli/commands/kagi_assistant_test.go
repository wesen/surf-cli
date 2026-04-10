package commands

import (
	"strings"
	"testing"
)

func TestBuildKagiAssistantCodeIncludesTagsAndOptions(t *testing.T) {
	code, err := buildKagiAssistantCode(&KagiAssistantSettings{
		Assistant:        "Quick",
		Lens:             "Programming",
		Tags:             "Temporary, photo",
		CreateTags:       true,
		WebSearchMode:    "off",
		PromptTimeoutSec: 90,
		Query:            "hello",
	})
	if err != nil {
		t.Fatalf("buildKagiAssistantCode returned error: %v", err)
	}
	for _, needle := range []string{
		`"assistant":"Quick"`,
		`"lens":"Programming"`,
		`"webSearchMode":"off"`,
		`"createTags":true`,
		`"tags":["Temporary","photo"]`,
		`"prompt":"hello"`,
	} {
		if !strings.Contains(code, needle) {
			t.Fatalf("missing %s in generated code", needle)
		}
	}
	if !strings.Contains(code, "const tagNames = Array.isArray(options.tags)") {
		t.Fatalf("missing embedded kagi assistant script body")
	}
}

func TestSplitCSVValues(t *testing.T) {
	got := splitCSVValues(" Temporary,photo ,, Public ")
	want := []string{"Temporary", "photo", "Public"}
	if len(got) != len(want) {
		t.Fatalf("unexpected length: got %d want %d (%v)", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("unexpected value at %d: got %q want %q", i, got[i], want[i])
		}
	}
}

func TestRenderKagiAssistantMarkdownIncludesTags(t *testing.T) {
	text := renderKagiAssistantMarkdown(map[string]any{
		"kind":     "response",
		"href":     "https://kagi.com/assistant/abc",
		"prompt":   "hello",
		"response": "world",
		"tagSelection": map[string]any{
			"visibleTags": []any{"Temporary", "photo"},
		},
		"metadata": map[string]any{"Model": "Quick"},
	})
	for _, needle := range []string{"# Kagi Assistant", "- Tags: [Temporary photo]", "## Prompt", "## Response", "- Model: Quick"} {
		if !strings.Contains(text, needle) {
			t.Fatalf("missing %q in markdown render: %s", needle, text)
		}
	}
}
