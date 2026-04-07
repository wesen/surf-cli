package installer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallGoNativeHostChromiumLinuxWithSnap(t *testing.T) {
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(home, "snap", "chromium", "common"), 0o755); err != nil {
		t.Fatalf("mkdir snap root: %v", err)
	}

	srcBinary := filepath.Join(home, "surf-host-go")
	if err := os.WriteFile(srcBinary, []byte("#!/bin/bash\necho host\n"), 0o755); err != nil {
		t.Fatalf("write fake binary: %v", err)
	}

	result, err := InstallGoNativeHost(InstallOptions{
		ExtensionID: "abcdefghijklmnopabcdefghijklmnop",
		Browsers:    []string{"chromium"},
		HostBinary:  srcBinary,
		HomeDir:     home,
		GOOS:        "linux",
	})
	if err != nil {
		t.Fatalf("install failed: %v", err)
	}

	if len(result.Installed) != 2 {
		t.Fatalf("expected standard and snap installs, got %#v", result.Installed)
	}

	standardManifest := filepath.Join(home, ".config", "chromium", "NativeMessagingHosts", HostName+".json")
	snapManifest := filepath.Join(home, "snap", "chromium", "common", "chromium", "NativeMessagingHosts", HostName+".json")
	for _, p := range []string{standardManifest, snapManifest} {
		if _, err := os.Stat(p); err != nil {
			t.Fatalf("expected manifest %s: %v", p, err)
		}
	}

	standardWrapper := filepath.Join(home, ".local", "share", "surf-cli", "host-wrapper.sh")
	snapWrapper := filepath.Join(home, "snap", "chromium", "common", "surf-cli", "host-wrapper.sh")
	for _, p := range []string{standardWrapper, snapWrapper} {
		if _, err := os.Stat(p); err != nil {
			t.Fatalf("expected wrapper %s: %v", p, err)
		}
	}

	snapWrapperContent, err := os.ReadFile(snapWrapper)
	if err != nil {
		t.Fatalf("read snap wrapper: %v", err)
	}
	if !strings.Contains(string(snapWrapperContent), "SURF_SOCKET_PATH") {
		t.Fatalf("snap wrapper should export SURF_SOCKET_PATH, got %s", string(snapWrapperContent))
	}
}

func TestInstallGoNativeHostRejectsInvalidExtensionID(t *testing.T) {
	_, err := InstallGoNativeHost(InstallOptions{
		ExtensionID: "not-valid",
		Browsers:    []string{"chromium"},
		HostBinary:  "/tmp/does-not-matter",
		HomeDir:     t.TempDir(),
		GOOS:        "linux",
	})
	if err == nil {
		t.Fatalf("expected invalid extension ID error")
	}
}
