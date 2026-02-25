package socketbridge

import (
	"bufio"
	"encoding/json"
	"net"
	"testing"
)

func TestSessionWriteJSONLine(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	s := NewSession(server)
	go func() {
		if err := s.WriteJSONLine(map[string]any{"type": "ok", "id": 7}); err != nil {
			t.Errorf("write failed: %v", err)
		}
	}()

	r := bufio.NewReader(client)
	line, err := r.ReadBytes('\n')
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(line, &got); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if got["type"] != "ok" {
		t.Fatalf("unexpected type: %#v", got)
	}
}

func TestSessionManagerNotifyExtensionDisconnected(t *testing.T) {
	aServer, aClient := net.Pipe()
	bServer, bClient := net.Pipe()
	defer aServer.Close()
	defer aClient.Close()
	defer bServer.Close()
	defer bClient.Close()

	mgr := NewSessionManager()
	mgr.Add(aServer)
	mgr.Add(bServer)

	errCh := make(chan error, 2)
	readOne := func(conn net.Conn) {
		r := bufio.NewReader(conn)
		line, err := r.ReadBytes('\n')
		if err != nil {
			errCh <- err
			return
		}
		var msg map[string]any
		if err := json.Unmarshal(line, &msg); err != nil {
			errCh <- err
			return
		}
		if msg["type"] != "extension_disconnected" {
			errCh <- errUnexpectedMessage
			return
		}
		errCh <- nil
	}
	go readOne(aClient)
	go readOne(bClient)

	mgr.NotifyExtensionDisconnected("restart")

	for i := 0; i < 2; i++ {
		if err := <-errCh; err != nil {
			t.Fatalf("unexpected read result: %v", err)
		}
	}
	if mgr.Count() != 0 {
		t.Fatalf("expected empty manager")
	}
}

var errUnexpectedMessage = &net.AddrError{Err: "unexpected message", Addr: ""}
