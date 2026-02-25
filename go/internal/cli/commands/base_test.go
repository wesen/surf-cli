package commands

import "testing"

func TestBuildToolRequest(t *testing.T) {
	tabID := int64(4)
	windowID := int64(9)
	req := BuildToolRequest("page.read", map[string]any{"filter": "interactive"}, &tabID, &windowID)
	if req["type"] != "tool_request" {
		t.Fatalf("unexpected type: %#v", req)
	}
	params := req["params"].(map[string]any)
	if params["tool"] != "page.read" {
		t.Fatalf("unexpected tool: %#v", params)
	}
	if req["tabId"] != int64(4) || req["windowId"] != int64(9) {
		t.Fatalf("expected tab/window id in request: %#v", req)
	}
}
