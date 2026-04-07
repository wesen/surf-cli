package main

import "testing"

func TestDecodeToolPayloadParsesStructuredJSON(t *testing.T) {
	resp := map[string]any{
		"result": map[string]any{
			"content": []any{
				map[string]any{"type": "text", "text": `{"response":"hello","model":"gpt-5-2"}`},
			},
		},
	}

	data, err := decodeToolPayload(resp)
	if err != nil {
		t.Fatalf("decodeToolPayload returned error: %v", err)
	}
	if got := data["response"]; got != "hello" {
		t.Fatalf("expected response hello, got %#v", got)
	}
	if got := data["model"]; got != "gpt-5-2" {
		t.Fatalf("expected model gpt-5-2, got %#v", got)
	}
}

func TestDecodeToolPayloadExtractsToolError(t *testing.T) {
	resp := map[string]any{
		"error": map[string]any{
			"content": []any{
				map[string]any{"type": "text", "text": "ChatGPT login required"},
			},
		},
	}

	_, err := decodeToolPayload(resp)
	if err == nil {
		t.Fatalf("expected error")
	}
	if got := err.Error(); got != "ChatGPT login required" {
		t.Fatalf("unexpected error: %s", got)
	}
}

func TestDecodeToolPayloadFallsBackToPlainText(t *testing.T) {
	resp := map[string]any{
		"result": map[string]any{
			"content": []any{
				map[string]any{"type": "text", "text": "plain text response"},
			},
		},
	}

	data, err := decodeToolPayload(resp)
	if err != nil {
		t.Fatalf("decodeToolPayload returned error: %v", err)
	}
	if got := data["response"]; got != "plain text response" {
		t.Fatalf("expected plain text fallback, got %#v", got)
	}
}
