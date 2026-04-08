package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/nicobailon/surf-cli/gohost/internal/host/pending"
	"github.com/nicobailon/surf-cli/gohost/internal/host/router"
	"github.com/nicobailon/surf-cli/gohost/internal/host/socketbridge"
)

func TestHandleSessionLineDispatchesChatGPTProvider(t *testing.T) {
	server, client := net.Pipe()
	defer client.Close()
	defer server.Close()

	session := socketbridge.NewSession(server)
	h := &hostRuntime{
		log:     log.New(io.Discard, "", 0),
		pending: pending.NewStore(),
		runChatGPTTool: func(_ context.Context, args map[string]any, tabID *int64, _ func(string, ...any)) (map[string]any, error) {
			if asString(args["query"]) != "hello" {
				t.Fatalf("unexpected query: %v", args["query"])
			}
			if tabID == nil || *tabID != 5 {
				t.Fatalf("unexpected tabID: %v", tabID)
			}
			return map[string]any{"response": "ok", "model": "current", "tookMs": int64(1)}, nil
		},
	}

	line := []byte(`{"type":"tool_request","method":"execute_tool","params":{"tool":"chatgpt","args":{"query":"hello"}},"id":"req-1","tabId":5}` + "\n")
	done := make(chan struct{})
	go func() {
		defer close(done)
		h.handleSessionLine(context.Background(), session, line)
	}()

	resp := readLineJSON(t, client)
	<-done
	if asString(resp["type"]) != "tool_response" {
		t.Fatalf("unexpected type: %#v", resp)
	}
	if asString(resp["id"]) != "req-1" {
		t.Fatalf("unexpected id: %#v", resp["id"])
	}
	result, _ := resp["result"].(map[string]any)
	content, _ := result["content"].([]any)
	first, _ := content[0].(map[string]any)
	text := asString(first["text"])
	if !strings.Contains(text, `"response": "ok"`) {
		t.Fatalf("unexpected tool text: %s", text)
	}
}

func TestHandleSessionLineChatGPTProviderError(t *testing.T) {
	server, client := net.Pipe()
	defer client.Close()
	defer server.Close()

	session := socketbridge.NewSession(server)
	h := &hostRuntime{
		log:     log.New(io.Discard, "", 0),
		pending: pending.NewStore(),
		runChatGPTTool: func(_ context.Context, _ map[string]any, _ *int64, _ func(string, ...any)) (map[string]any, error) {
			return nil, errors.New("boom")
		},
	}

	line := []byte(`{"type":"tool_request","method":"execute_tool","params":{"tool":"chatgpt","args":{"query":"hello"}},"id":"req-2"}` + "\n")
	done := make(chan struct{})
	go func() {
		defer close(done)
		h.handleSessionLine(context.Background(), session, line)
	}()

	resp := readLineJSON(t, client)
	<-done
	errPayload, _ := resp["error"].(map[string]any)
	content, _ := errPayload["content"].([]any)
	first, _ := content[0].(map[string]any)
	if asString(first["text"]) != "boom" {
		t.Fatalf("unexpected error payload: %#v", resp)
	}
}

func TestHandleSessionLineGeminiStillBlocked(t *testing.T) {
	server, client := net.Pipe()
	defer client.Close()
	defer server.Close()

	session := socketbridge.NewSession(server)
	h := &hostRuntime{
		log:     log.New(io.Discard, "", 0),
		pending: pending.NewStore(),
		runChatGPTTool: func(_ context.Context, _ map[string]any, _ *int64, _ func(string, ...any)) (map[string]any, error) {
			return map[string]any{"response": "unexpected"}, nil
		},
	}

	line := []byte(`{"type":"tool_request","method":"execute_tool","params":{"tool":"gemini","args":{"query":"hi"}},"id":"req-3"}` + "\n")
	done := make(chan struct{})
	go func() {
		defer close(done)
		h.handleSessionLine(context.Background(), session, line)
	}()

	resp := readLineJSON(t, client)
	<-done
	errPayload, _ := resp["error"].(map[string]any)
	content, _ := errPayload["content"].([]any)
	first, _ := content[0].(map[string]any)
	errText := asString(first["text"])
	if !strings.Contains(errText, "not supported") {
		t.Fatalf("expected unsupported error, got: %s", errText)
	}
}

func readLineJSON(t *testing.T, conn net.Conn) map[string]any {
	t.Helper()
	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	line, err := bufio.NewReader(conn).ReadBytes('\n')
	if err != nil {
		t.Fatalf("failed to read response: %v", err)
	}
	var msg map[string]any
	if err := json.Unmarshal(line, &msg); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	return msg
}

func TestHandleSessionDisconnectCancelsChatGPTProvider(t *testing.T) {
	server, client := net.Pipe()
	session := socketbridge.NewSession(server)

	canceled := make(chan struct{}, 1)
	h := &hostRuntime{
		log:      log.New(io.Discard, "", 0),
		sessions: socketbridge.NewSessionManager(),
		pending:  pending.NewStore(),
		streams:  router.NewStreamRegistry(),
		runChatGPTTool: func(ctx context.Context, _ map[string]any, _ *int64, _ func(string, ...any)) (map[string]any, error) {
			<-ctx.Done()
			canceled <- struct{}{}
			return nil, ctx.Err()
		},
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		h.handleSession(context.Background(), session)
	}()

	_, err := client.Write([]byte(`{"type":"tool_request","method":"execute_tool","params":{"tool":"chatgpt","args":{"query":"hello"}},"id":"req-cancel"}` + "\n"))
	if err != nil {
		t.Fatalf("failed to write request: %v", err)
	}
	_ = client.Close()

	select {
	case <-canceled:
	case <-time.After(2 * time.Second):
		t.Fatal("expected provider context to be canceled when client disconnected")
	}

	<-done
}
