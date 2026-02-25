//go:build !windows

package socketbridge

import (
	"errors"
	"net"
	"os"
)

type unixListener struct {
	endpoint string
	ln       net.Listener
}

func listenPlatform(endpoint string) (Listener, error) {
	if endpoint == "" {
		return nil, errors.New("socketbridge: endpoint is required")
	}
	_ = os.Remove(endpoint)

	ln, err := net.Listen("unix", endpoint)
	if err != nil {
		return nil, err
	}
	_ = os.Chmod(endpoint, 0o600)

	return &unixListener{endpoint: endpoint, ln: ln}, nil
}

func (l *unixListener) Accept() (net.Conn, error) {
	return l.ln.Accept()
}

func (l *unixListener) Close() error {
	return l.ln.Close()
}

func (l *unixListener) Endpoint() string {
	return l.endpoint
}

func (l *unixListener) Cleanup() error {
	err := os.Remove(l.endpoint)
	if err == nil || errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}
