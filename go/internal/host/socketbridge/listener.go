package socketbridge

import "net"

// Listener abstracts local IPC listeners so host runtime can stay OS-agnostic.
type Listener interface {
	Accept() (net.Conn, error)
	Close() error
	Endpoint() string
	Cleanup() error
}

func Listen(endpoint string) (Listener, error) {
	return listenPlatform(endpoint)
}
