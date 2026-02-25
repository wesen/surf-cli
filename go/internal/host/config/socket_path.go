package config

import (
	"os"
	"runtime"
)

const (
	DefaultSocketPath = "/tmp/surf.sock"
	DefaultPipePath   = "//./pipe/surf"
)

func ResolveSocketPath(goos string) string {
	if p := os.Getenv("SURF_SOCKET_PATH"); p != "" {
		return p
	}
	if goos == "windows" {
		return DefaultPipePath
	}
	return DefaultSocketPath
}

func CurrentSocketPath() string {
	return ResolveSocketPath(runtime.GOOS)
}
