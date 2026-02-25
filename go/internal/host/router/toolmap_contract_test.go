package router

import "testing"

func TestToolMapContractCoreSamples(t *testing.T) {
	tests := []struct {
		name     string
		tool     string
		args     map[string]any
		typeWant string
		key      string
		valWant  any
	}{
		{name: "page.read", tool: "page.read", args: map[string]any{}, typeWant: "READ_PAGE"},
		{name: "page.text", tool: "page.text", args: map[string]any{}, typeWant: "GET_PAGE_TEXT"},
		{name: "navigate", tool: "navigate", args: map[string]any{"url": "https://example.com"}, typeWant: "EXECUTE_NAVIGATE", key: "url", valWant: "https://example.com"},
		{name: "click_ref", tool: "click", args: map[string]any{"ref": "e12"}, typeWant: "CLICK_REF", key: "button", valWant: "left"},
		{name: "type", tool: "type", args: map[string]any{"text": "hello"}, typeWant: "EXECUTE_TYPE", key: "text", valWant: "hello"},
		{name: "network.get", tool: "network.get", args: map[string]any{"id": "req-abc"}, typeWant: "GET_NETWORK_ENTRY", key: "requestId", valWant: "req-abc"},
		{name: "cookie.set", tool: "cookie.set", args: map[string]any{"name": "session", "value": "abc"}, typeWant: "COOKIE_SET", key: "name", valWant: "session"},
		{name: "emulate.device", tool: "emulate.device", args: map[string]any{"device": "iPhone 14 Pro"}, typeWant: "EMULATE_DEVICE", key: "device", valWant: "iPhone 14 Pro"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, err := MapToolToMessage(ToolRequest{Params: ToolParams{Tool: tt.tool, Args: tt.args}})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := msg["type"]; got != tt.typeWant {
				t.Fatalf("unexpected type: got=%v want=%v msg=%#v", got, tt.typeWant, msg)
			}
			if tt.key != "" {
				if got := msg[tt.key]; got != tt.valWant {
					t.Fatalf("unexpected %s: got=%v want=%v", tt.key, got, tt.valWant)
				}
			}
		})
	}
}
