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
