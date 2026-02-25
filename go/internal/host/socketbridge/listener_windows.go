//go:build windows

package socketbridge

import "errors"

var ErrWindowsPipeUnsupported = errors.New("socketbridge: windows named pipe listener is not implemented yet")

func listenPlatform(_ string) (Listener, error) {
	return nil, ErrWindowsPipeUnsupported
}
