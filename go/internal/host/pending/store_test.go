package pending

import (
	"net"
	"testing"

	"github.com/nicobailon/surf-cli/gohost/internal/host/socketbridge"
)

func TestIDAllocatorIncrements(t *testing.T) {
	a := NewIDAllocator(0)
	if a.Next() != 1 || a.Next() != 2 {
		t.Fatalf("unexpected id sequence")
	}
}

func TestStorePutPopAndDeleteForSession(t *testing.T) {
	aServer, aClient := net.Pipe()
	bServer, bClient := net.Pipe()
	defer aServer.Close()
	defer aClient.Close()
	defer bServer.Close()
	defer bClient.Close()

	sa := socketbridge.NewSession(aServer)
	sb := socketbridge.NewSession(bServer)

	s := NewStore()
	s.Put(1, Request{Session: sa, Kind: KindToolRequest, OriginalID: 9, HasOriginalID: true})
	s.Put(2, Request{Session: sb, Kind: KindStream})

	req, ok := s.Pop(1)
	if !ok || req.OriginalID != 9 {
		t.Fatalf("pop mismatch: ok=%v req=%+v", ok, req)
	}

	if s.Len() != 1 {
		t.Fatalf("expected len=1, got %d", s.Len())
	}

	ids := s.DeleteForSession(sb)
	if len(ids) != 1 || ids[0] != 2 {
		t.Fatalf("unexpected deleted ids: %#v", ids)
	}
	if s.Len() != 0 {
		t.Fatalf("expected empty store")
	}
}
