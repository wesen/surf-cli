package providers

import (
	"context"
	"strings"
	"testing"
	"time"
)

type fakeNativeCaller struct {
	handler func(msg map[string]any) (map[string]any, error)
}

func (f *fakeNativeCaller) Request(_ context.Context, msg map[string]any, _ time.Duration) (map[string]any, error) {
	return f.handler(msg)
}

func TestParseChatGPTRequest(t *testing.T) {
	t.Run("requires query", func(t *testing.T) {
		_, err := parseChatGPTRequest(map[string]any{}, nil)
		if err == nil {
			t.Fatalf("expected query error")
		}
	})

	t.Run("allows list-models without query", func(t *testing.T) {
		req, err := parseChatGPTRequest(map[string]any{
			"list-models": true,
		}, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !req.ListModels {
			t.Fatalf("expected list models flag")
		}
	})

	t.Run("parses options", func(t *testing.T) {
		id := int64(7)
		req, err := parseChatGPTRequest(map[string]any{
			"query":     "hello",
			"model":     "gpt-4o",
			"with-page": "true",
			"timeout":   "30",
		}, &id)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !req.WithPage {
			t.Fatalf("expected with-page true")
		}
		if req.Timeout != 30*time.Second {
			t.Fatalf("unexpected timeout: %v", req.Timeout)
		}
		if req.Model != "gpt-4o" {
			t.Fatalf("unexpected model: %s", req.Model)
		}
	})
}

func TestHandleChatGPTToolSuccess(t *testing.T) {
	var insertedPrompt string
	var closeCalled bool
	polls := 0

	caller := &fakeNativeCaller{handler: func(msg map[string]any) (map[string]any, error) {
		switch msg["type"] {
		case "GET_CHATGPT_COOKIES":
			return map[string]any{"cookies": []any{map[string]any{"name": "__Secure-next-auth.session-token", "value": "abc"}}}, nil
		case "CHATGPT_NEW_TAB":
			return map[string]any{"tabId": int64(42)}, nil
		case "CHATGPT_CLOSE_TAB":
			closeCalled = true
			return map[string]any{"success": true}, nil
		case "CHATGPT_CDP_COMMAND":
			if msg["method"] == "Input.insertText" {
				params, _ := msg["params"].(map[string]any)
				insertedPrompt = asString(params["text"])
			}
			return map[string]any{"ok": true}, nil
		case "CHATGPT_EVALUATE":
			expr := asString(msg["expression"])
			switch {
			case strings.Contains(expr, "document.readyState"):
				return map[string]any{"result": map[string]any{"value": "complete"}}, nil
			case strings.Contains(expr, "document.title.toLowerCase"):
				return map[string]any{"result": map[string]any{"value": "chatgpt"}}, nil
			case strings.Contains(expr, "challenge-platform"):
				return map[string]any{"result": map[string]any{"value": false}}, nil
			case strings.Contains(expr, "backend-api/me"):
				return map[string]any{"result": map[string]any{"value": map[string]any{"status": float64(200), "hasLoginCta": false}}}, nil
			case strings.Contains(expr, "!node.hasAttribute('disabled')"):
				return map[string]any{"result": map[string]any{"value": true}}, nil
			case strings.Contains(expr, "text.trim().length > 0"):
				return map[string]any{"result": map[string]any{"value": true}}, nil
			case strings.Contains(expr, "return 'clicked'"):
				return map[string]any{"result": map[string]any{"value": "clicked"}}, nil
			case strings.Contains(expr, "lastAssistantTurn"):
				polls++
				return map[string]any{"result": map[string]any{"value": map[string]any{"text": "Hello from ChatGPT", "stopVisible": false, "finished": true}}}, nil
			default:
				return map[string]any{"result": map[string]any{"value": true}}, nil
			}
		default:
			return map[string]any{"ok": true}, nil
		}
	}}

	resp, err := HandleChatGPTTool(context.Background(), caller, map[string]any{"query": "hello world"}, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := asString(resp["response"]); got != "Hello from ChatGPT" {
		t.Fatalf("unexpected response: %q", got)
	}
	if insertedPrompt != "hello world" {
		t.Fatalf("unexpected inserted prompt: %q", insertedPrompt)
	}
	if !closeCalled {
		t.Fatalf("expected close tab call")
	}
	if polls == 0 {
		t.Fatalf("expected response polling")
	}
}

func TestHandleChatGPTToolWithPageContext(t *testing.T) {
	sourceTabID := int64(9)
	var insertedPrompt string
	var gotPageTabID int64

	caller := &fakeNativeCaller{handler: func(msg map[string]any) (map[string]any, error) {
		switch msg["type"] {
		case "GET_CHATGPT_COOKIES":
			return map[string]any{"cookies": []any{map[string]any{"name": "__Secure-next-auth.session-token", "value": "abc"}}}, nil
		case "GET_PAGE_TEXT":
			if v, ok := toInt64(msg["tabId"]); ok {
				gotPageTabID = v
			}
			return map[string]any{"url": "https://example.com", "text": "Page body"}, nil
		case "CHATGPT_NEW_TAB":
			return map[string]any{"tabId": int64(11)}, nil
		case "CHATGPT_CLOSE_TAB":
			return map[string]any{"success": true}, nil
		case "CHATGPT_CDP_COMMAND":
			if msg["method"] == "Input.insertText" {
				params, _ := msg["params"].(map[string]any)
				insertedPrompt = asString(params["text"])
			}
			return map[string]any{"ok": true}, nil
		case "CHATGPT_EVALUATE":
			expr := asString(msg["expression"])
			switch {
			case strings.Contains(expr, "document.readyState"):
				return map[string]any{"result": map[string]any{"value": "complete"}}, nil
			case strings.Contains(expr, "document.title.toLowerCase"):
				return map[string]any{"result": map[string]any{"value": "chatgpt"}}, nil
			case strings.Contains(expr, "challenge-platform"):
				return map[string]any{"result": map[string]any{"value": false}}, nil
			case strings.Contains(expr, "backend-api/me"):
				return map[string]any{"result": map[string]any{"value": map[string]any{"status": float64(200), "hasLoginCta": false}}}, nil
			case strings.Contains(expr, "!node.hasAttribute('disabled')"):
				return map[string]any{"result": map[string]any{"value": true}}, nil
			case strings.Contains(expr, "text.trim().length > 0"):
				return map[string]any{"result": map[string]any{"value": true}}, nil
			case strings.Contains(expr, "return 'clicked'"):
				return map[string]any{"result": map[string]any{"value": "clicked"}}, nil
			case strings.Contains(expr, "lastAssistantTurn"):
				return map[string]any{"result": map[string]any{"value": map[string]any{"text": "ok", "stopVisible": false, "finished": true}}}, nil
			default:
				return map[string]any{"result": map[string]any{"value": true}}, nil
			}
		default:
			return map[string]any{"ok": true}, nil
		}
	}}

	_, err := HandleChatGPTTool(context.Background(), caller, map[string]any{"query": "summarize", "with-page": true}, &sourceTabID, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPageTabID != sourceTabID {
		t.Fatalf("expected page text request on tab %d, got %d", sourceTabID, gotPageTabID)
	}
	if !strings.Contains(insertedPrompt, "Page: https://example.com") || !strings.Contains(insertedPrompt, "Page body") || !strings.Contains(insertedPrompt, "summarize") {
		t.Fatalf("inserted prompt missing page context: %q", insertedPrompt)
	}
}

func TestHandleChatGPTToolWithFileUpload(t *testing.T) {
	var uploadedSelector string
	var uploadedFiles []string

	caller := &fakeNativeCaller{handler: func(msg map[string]any) (map[string]any, error) {
		switch msg["type"] {
		case "GET_CHATGPT_COOKIES":
			return map[string]any{"cookies": []any{map[string]any{"name": "__Secure-next-auth.session-token", "value": "abc"}}}, nil
		case "CHATGPT_NEW_TAB":
			return map[string]any{"tabId": int64(42)}, nil
		case "CHATGPT_CLOSE_TAB":
			return map[string]any{"success": true}, nil
		case "UPLOAD_FILE":
			uploadedSelector = asString(msg["selector"])
			switch files := msg["files"].(type) {
			case []string:
				uploadedFiles = append([]string{}, files...)
			case []any:
				for _, f := range files {
					uploadedFiles = append(uploadedFiles, asString(f))
				}
			}
			return map[string]any{"success": true, "filesSet": float64(1)}, nil
		case "CHATGPT_CDP_COMMAND":
			return map[string]any{"ok": true}, nil
		case "CHATGPT_EVALUATE":
			expr := asString(msg["expression"])
			switch {
			case strings.Contains(expr, "document.readyState"):
				return map[string]any{"result": map[string]any{"value": "complete"}}, nil
			case strings.Contains(expr, "document.title.toLowerCase"):
				return map[string]any{"result": map[string]any{"value": "chatgpt"}}, nil
			case strings.Contains(expr, "challenge-platform"):
				return map[string]any{"result": map[string]any{"value": false}}, nil
			case strings.Contains(expr, "backend-api/me"):
				return map[string]any{"result": map[string]any{"value": map[string]any{"status": float64(200), "hasLoginCta": false}}}, nil
			case strings.Contains(expr, "!node.hasAttribute('disabled')"):
				return map[string]any{"result": map[string]any{"value": true}}, nil
			case strings.Contains(expr, "data-surf-file-input-id"):
				return map[string]any{"result": map[string]any{"value": `[data-surf-file-input-id="surf-upload-1"]`}}, nil
			case strings.Contains(expr, "text.trim().length > 0"):
				return map[string]any{"result": map[string]any{"value": true}}, nil
			case strings.Contains(expr, "return 'clicked'"):
				return map[string]any{"result": map[string]any{"value": "clicked"}}, nil
			case strings.Contains(expr, "lastAssistantTurn"):
				return map[string]any{"result": map[string]any{"value": map[string]any{"text": "ok", "stopVisible": false, "finished": true}}}, nil
			default:
				return map[string]any{"result": map[string]any{"value": true}}, nil
			}
		default:
			return map[string]any{"ok": true}, nil
		}
	}}

	_, err := HandleChatGPTTool(context.Background(), caller, map[string]any{
		"query": "review this file",
		"file":  "/tmp/demo.txt",
	}, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if uploadedSelector == "" {
		t.Fatalf("expected upload selector to be set")
	}
	if len(uploadedFiles) != 1 || uploadedFiles[0] != "/tmp/demo.txt" {
		t.Fatalf("unexpected uploaded files: %#v", uploadedFiles)
	}
}

func TestHandleChatGPTToolListModels(t *testing.T) {
	caller := &fakeNativeCaller{handler: func(msg map[string]any) (map[string]any, error) {
		switch msg["type"] {
		case "GET_CHATGPT_COOKIES":
			return map[string]any{"cookies": []any{map[string]any{"name": "__Secure-next-auth.session-token", "value": "abc"}}}, nil
		case "CHATGPT_NEW_TAB":
			return map[string]any{"tabId": int64(17)}, nil
		case "CHATGPT_CLOSE_TAB":
			return map[string]any{"success": true}, nil
		case "CHATGPT_EVALUATE":
			expr := asString(msg["expression"])
			switch {
			case strings.Contains(expr, "document.readyState"):
				return map[string]any{"result": map[string]any{"value": "complete"}}, nil
			case strings.Contains(expr, "document.title.toLowerCase"):
				return map[string]any{"result": map[string]any{"value": "chatgpt"}}, nil
			case strings.Contains(expr, "challenge-platform"):
				return map[string]any{"result": map[string]any{"value": false}}, nil
			case strings.Contains(expr, "backend-api/me"):
				return map[string]any{"result": map[string]any{"value": map[string]any{"status": float64(200), "hasLoginCta": false}}}, nil
			case strings.Contains(expr, "!node.hasAttribute('disabled')"):
				return map[string]any{"result": map[string]any{"value": true}}, nil
			case strings.Contains(expr, "model-switcher-dropdown-button"):
				return map[string]any{"result": map[string]any{"value": true}}, nil
			case strings.Contains(expr, "const menu = document.querySelector"):
				return map[string]any{"result": map[string]any{"value": map[string]any{
					"found":    true,
					"models":   []any{"GPT-4o", "o1"},
					"selected": "GPT-4o",
				}}}, nil
			default:
				return map[string]any{"result": map[string]any{"value": true}}, nil
			}
		default:
			return map[string]any{"ok": true}, nil
		}
	}}

	resp, err := HandleChatGPTTool(context.Background(), caller, map[string]any{
		"list-models": true,
	}, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	models, ok := resp["models"].([]string)
	if !ok {
		t.Fatalf("expected string models slice, got %#v", resp["models"])
	}
	if len(models) != 2 || models[0] != "GPT-4o" || models[1] != "o1" {
		t.Fatalf("unexpected models: %#v", models)
	}
	if asString(resp["selected"]) != "GPT-4o" {
		t.Fatalf("unexpected selected: %#v", resp["selected"])
	}
}
