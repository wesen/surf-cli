package config

import "testing"

func TestResolveSocketPathDefaultLinux(t *testing.T) {
	t.Setenv("SURF_SOCKET_PATH", "")
	if got := ResolveSocketPath("linux"); got != DefaultSocketPath {
		t.Fatalf("unexpected linux socket path: %s", got)
	}
}

func TestResolveSocketPathDefaultWindows(t *testing.T) {
	t.Setenv("SURF_SOCKET_PATH", "")
	if got := ResolveSocketPath("windows"); got != DefaultPipePath {
		t.Fatalf("unexpected windows pipe path: %s", got)
	}
}

func TestResolveSocketPathEnvOverride(t *testing.T) {
	t.Setenv("SURF_SOCKET_PATH", "/custom/sock")
	if got := ResolveSocketPath("linux"); got != "/custom/sock" {
		t.Fatalf("expected env override, got %s", got)
	}
}
