//go:build !windows

package socketbridge

import (
	"net"
	"path/filepath"
	"testing"
)

func TestListenUnixAndAccept(t *testing.T) {
	sock := filepath.Join(t.TempDir(), "surf.sock")
	l, err := Listen(sock)
	if err != nil {
		t.Fatalf("listen failed: %v", err)
	}
	defer l.Close()
	defer l.Cleanup()

	acceptCh := make(chan error, 1)
	go func() {
		conn, err := l.Accept()
		if err != nil {
			acceptCh <- err
			return
		}
		_ = conn.Close()
		acceptCh <- nil
	}()

	conn, err := net.Dial("unix", sock)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	_ = conn.Close()

	if err := <-acceptCh; err != nil {
		t.Fatalf("accept failed: %v", err)
	}
}
