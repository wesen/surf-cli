package router

import "testing"

func TestMapToolToMessagePageRead(t *testing.T) {
	tabID := int64(17)
	msg, err := MapToolToMessage(ToolRequest{
		Params: ToolParams{Tool: "page.read", Args: map[string]any{}},
		TabID:  &tabID,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg["type"] != "READ_PAGE" {
		t.Fatalf("unexpected type: %#v", msg)
	}
	if msg["tabId"] != int64(17) {
		t.Fatalf("tabId missing from message: %#v", msg)
	}
}

func TestMapToolToMessageClickAlias(t *testing.T) {
	msg, err := MapToolToMessage(ToolRequest{
		Params: ToolParams{Tool: "click", Args: map[string]any{"ref": "e12"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg["type"] != "CLICK_REF" || msg["button"] != "left" {
		t.Fatalf("unexpected click mapping: %#v", msg)
	}
}

func TestMapToolToMessageRejectProviderTool(t *testing.T) {
	_, err := MapToolToMessage(ToolRequest{
		Params: ToolParams{Tool: "chatgpt", Args: map[string]any{"query": "hi"}},
	})
	if err == nil {
		t.Fatalf("expected unsupported error")
	}
	if _, ok := err.(*UnsupportedToolError); !ok {
		t.Fatalf("expected UnsupportedToolError, got %T", err)
	}
}

func TestMapToolToMessageRejectDeferredTool(t *testing.T) {
	_, err := MapToolToMessage(ToolRequest{Params: ToolParams{Tool: "perf.start", Args: map[string]any{}}})
	if err == nil {
		t.Fatalf("expected unsupported error")
	}
}

func TestMapToolToMessageTabSwitchNamed(t *testing.T) {
	msg, err := MapToolToMessage(ToolRequest{Params: ToolParams{Tool: "tab.switch", Args: map[string]any{"id": "docs"}}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg["type"] != "NAMED_TAB_SWITCH" {
		t.Fatalf("unexpected mapping: %#v", msg)
	}
}
