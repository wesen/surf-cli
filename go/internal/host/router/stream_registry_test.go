package router

import (
	"net"
	"testing"

	"github.com/nicobailon/surf-cli/gohost/internal/host/socketbridge"
)

func TestStreamRegistrySetGetDelete(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	session := socketbridge.NewSession(server)
	r := NewStreamRegistry()
	r.Set(11, session)

	got, ok := r.Get(11)
	if !ok || got != session {
		t.Fatalf("unexpected registry lookup")
	}

	r.Delete(11)
	if _, ok := r.Get(11); ok {
		t.Fatalf("expected stream to be removed")
	}
}

func TestStreamRegistryDeleteBySession(t *testing.T) {
	aServer, aClient := net.Pipe()
	bServer, bClient := net.Pipe()
	defer aServer.Close()
	defer aClient.Close()
	defer bServer.Close()
	defer bClient.Close()

	sa := socketbridge.NewSession(aServer)
	sb := socketbridge.NewSession(bServer)

	r := NewStreamRegistry()
	r.Set(1, sa)
	r.Set(2, sb)
	r.Set(3, sb)

	ids := r.DeleteBySession(sb)
	if len(ids) != 2 || ids[0] != 2 || ids[1] != 3 {
		t.Fatalf("unexpected deleted ids: %#v", ids)
	}
	if r.Len() != 1 {
		t.Fatalf("expected one remaining stream, got %d", r.Len())
	}
}
