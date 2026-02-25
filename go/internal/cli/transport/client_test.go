package transport

import (
	"bufio"
	"context"
	"encoding/json"
	"net"
	"path/filepath"
	"testing"
	"time"
)

func TestClientSendUnix(t *testing.T) {
	sock := filepath.Join(t.TempDir(), "surf.sock")
	ln, err := net.Listen("unix", sock)
	if err != nil {
		t.Fatalf("listen failed: %v", err)
	}
	defer ln.Close()

	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		reader := bufio.NewReader(conn)
		line, err := reader.ReadBytes('\n')
		if err != nil {
			return
		}
		var req map[string]any
		if err := json.Unmarshal(line, &req); err != nil {
			return
		}
		resp := map[string]any{"type": "tool_response", "id": req["id"], "result": map[string]any{"content": []map[string]any{{"type": "text", "text": "ok"}}}}
		b, _ := json.Marshal(resp)
		_, _ = conn.Write(append(b, '\n'))
	}()

	client := NewClient(sock, 2*time.Second)
	resp, err := client.Send(context.Background(), map[string]any{"type": "tool_request", "id": "x"})
	if err != nil {
		t.Fatalf("send failed: %v", err)
	}
	if resp["type"] != "tool_response" {
		t.Fatalf("unexpected response: %#v", resp)
	}
}

func TestClientStream(t *testing.T) {
	sock := filepath.Join(t.TempDir(), "surf.sock")
	ln, err := net.Listen("unix", sock)
	if err != nil {
		t.Fatalf("listen failed: %v", err)
	}
	defer ln.Close()

	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		reader := bufio.NewReader(conn)

		_, _ = reader.ReadBytes('\n') // stream_request
		started, _ := json.Marshal(map[string]any{"type": "stream_started", "streamId": 1})
		_, _ = conn.Write(append(started, '\n'))
		event, _ := json.Marshal(map[string]any{"type": "console_event", "text": "hello"})
		_, _ = conn.Write(append(event, '\n'))

		_, _ = reader.ReadBytes('\n') // stream_stop
	}()

	client := NewClient(sock, 2*time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	got := ""
	err = client.Stream(ctx, "STREAM_CONSOLE", map[string]any{}, nil, func(msg map[string]any) error {
		if text, ok := msg["text"].(string); ok {
			got = text
		}
		cancel()
		return nil
	})
	if err != nil {
		t.Fatalf("stream failed: %v", err)
	}
	if got != "hello" {
		t.Fatalf("unexpected stream message: %q", got)
	}
}
