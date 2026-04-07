package installer

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"

	pkgerrors "github.com/pkg/errors"
)

const HostName = "surf.browser.host"

var extensionIDPattern = regexp.MustCompile(`^[a-p]{32}$`)

type browserConfig struct {
	Name    string
	Darwin  string
	Linux   string
	Windows string
}

var browsers = map[string]browserConfig{
	"chrome": {
		Name:    "Google Chrome",
		Darwin:  "Library/Application Support/Google/Chrome/NativeMessagingHosts",
		Linux:   ".config/google-chrome/NativeMessagingHosts",
		Windows: "Google\\Chrome",
	},
	"chromium": {
		Name:    "Chromium",
		Darwin:  "Library/Application Support/Chromium/NativeMessagingHosts",
		Linux:   ".config/chromium/NativeMessagingHosts",
		Windows: "Chromium",
	},
	"brave": {
		Name:    "Brave",
		Darwin:  "Library/Application Support/BraveSoftware/Brave-Browser/NativeMessagingHosts",
		Linux:   ".config/BraveSoftware/Brave-Browser/NativeMessagingHosts",
		Windows: "BraveSoftware\\Brave-Browser",
	},
	"edge": {
		Name:    "Microsoft Edge",
		Darwin:  "Library/Application Support/Microsoft Edge/NativeMessagingHosts",
		Linux:   ".config/microsoft-edge/NativeMessagingHosts",
		Windows: "Microsoft\\Edge",
	},
}

type TargetInstall struct {
	Browser string
	Path    string
}

type Result struct {
	Installed []TargetInstall
	Skipped   []string
	Hints     []string
}

type InstallOptions struct {
	ExtensionID string
	Browsers    []string
	HostBinary  string
	HomeDir     string
	GOOS        string
}

func InstallGoNativeHost(opts InstallOptions) (*Result, error) {
	if opts.ExtensionID == "" {
		return nil, fmt.Errorf("extension ID required")
	}
	if !extensionIDPattern.MatchString(opts.ExtensionID) {
		return nil, fmt.Errorf("invalid extension ID format: expected 32 lowercase letters (a-p)")
	}
	if opts.HostBinary == "" {
		return nil, fmt.Errorf("host binary path required")
	}
	if opts.HomeDir == "" {
		opts.HomeDir = os.Getenv("HOME")
	}
	if opts.HomeDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, pkgerrors.Wrap(err, "resolve user home dir")
		}
		opts.HomeDir = home
	}
	if opts.GOOS == "" {
		opts.GOOS = runtime.GOOS
	}
	if len(opts.Browsers) == 0 {
		opts.Browsers = []string{"chrome"}
	}

	r := &Result{}
	for _, browser := range opts.Browsers {
		cfg, ok := browsers[browser]
		if !ok {
			r.Skipped = append(r.Skipped, browser)
			r.Hints = append(r.Hints, fmt.Sprintf("Unknown browser: %s", browser))
			continue
		}
		targets, hints, err := installBrowser(opts, cfg, browser)
		if err != nil {
			r.Skipped = append(r.Skipped, cfg.Name)
			r.Hints = append(r.Hints, err.Error())
			continue
		}
		r.Installed = append(r.Installed, targets...)
		r.Hints = append(r.Hints, hints...)
	}
	return r, nil
}

func installBrowser(opts InstallOptions, cfg browserConfig, browserKey string) ([]TargetInstall, []string, error) {
	if opts.GOOS == "windows" {
		return nil, nil, fmt.Errorf("%s install via surf-go is not implemented on windows yet", cfg.Name)
	}

	manifestRel := ""
	switch opts.GOOS {
	case "darwin":
		manifestRel = cfg.Darwin
	case "linux":
		manifestRel = cfg.Linux
	default:
		return nil, nil, fmt.Errorf("unsupported platform: %s", opts.GOOS)
	}
	if manifestRel == "" {
		return nil, nil, fmt.Errorf("%s is not supported on %s", cfg.Name, opts.GOOS)
	}

	var installed []TargetInstall
	var hints []string

	wrapperDir, err := wrapperDir(opts.HomeDir, opts.GOOS)
	if err != nil {
		return nil, nil, err
	}
	hostPath := filepath.Join(wrapperDir, hostBinaryName(opts.GOOS))
	if err := copyExecutable(opts.HostBinary, hostPath); err != nil {
		return nil, nil, pkgerrors.Wrap(err, "copy standard host binary")
	}
	wrapperPath, err := writeWrapper(wrapperDir, hostPath, "")
	if err != nil {
		return nil, nil, pkgerrors.Wrap(err, "write standard wrapper")
	}
	manifestPath, err := writeManifest(filepath.Join(opts.HomeDir, manifestRel), opts.ExtensionID, wrapperPath)
	if err != nil {
		return nil, nil, pkgerrors.Wrap(err, "write standard manifest")
	}
	installed = append(installed, TargetInstall{Browser: cfg.Name, Path: manifestPath})

	if opts.GOOS == "linux" && browserKey == "chromium" {
		snapRoot := chromiumSnapRoot(opts.HomeDir)
		if snapRoot != "" {
			snapWrapperDir := filepath.Join(snapRoot, "surf-cli")
			snapHostPath := filepath.Join(snapWrapperDir, hostBinaryName(opts.GOOS))
			if err := copyExecutable(opts.HostBinary, snapHostPath); err != nil {
				return nil, nil, pkgerrors.Wrap(err, "copy snap host binary")
			}
			snapSocketPath := filepath.Join(snapWrapperDir, "surf.sock")
			snapWrapperPath, err := writeWrapper(snapWrapperDir, snapHostPath, snapSocketPath)
			if err != nil {
				return nil, nil, pkgerrors.Wrap(err, "write snap wrapper")
			}
			snapManifestPath, err := writeManifest(
				filepath.Join(snapRoot, "chromium", "NativeMessagingHosts"),
				opts.ExtensionID,
				snapWrapperPath,
			)
			if err != nil {
				return nil, nil, pkgerrors.Wrap(err, "write snap manifest")
			}
			installed = append(installed, TargetInstall{Browser: cfg.Name + " (snap)", Path: snapManifestPath})
			hints = append(hints, fmt.Sprintf("Snap target installed. Set SURF_SOCKET_PATH=%s in your shell when using surf CLI outside snap.", snapSocketPath))
		}
	}

	hints = append(hints, "Browser launches the Go host directly; SURF_HOST_PROFILE is not required.")
	return installed, hints, nil
}

func wrapperDir(homeDir, goos string) (string, error) {
	switch goos {
	case "darwin":
		return filepath.Join(homeDir, "Library/Application Support", "surf-cli"), nil
	case "linux":
		return filepath.Join(homeDir, ".local", "share", "surf-cli"), nil
	default:
		return "", fmt.Errorf("unsupported platform: %s", goos)
	}
}

func chromiumSnapRoot(homeDir string) string {
	root := filepath.Join(homeDir, "snap", "chromium", "common")
	if _, err := os.Stat(root); err == nil {
		return root
	}
	return ""
}

func hostBinaryName(goos string) string {
	if goos == "windows" {
		return "surf-host-go.exe"
	}
	return "surf-host-go"
}

func copyExecutable(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return pkgerrors.Wrap(err, "create destination directory")
	}

	in, err := os.Open(src)
	if err != nil {
		return pkgerrors.Wrap(err, "open source binary")
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return pkgerrors.Wrap(err, "open destination binary")
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return pkgerrors.Wrap(err, "copy binary bytes")
	}
	if err := out.Chmod(0o755); err != nil {
		return pkgerrors.Wrap(err, "chmod destination binary")
	}
	return nil
}

func writeWrapper(dir, hostPath, socketPath string) (string, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", pkgerrors.Wrap(err, "create wrapper directory")
	}
	wrapperPath := filepath.Join(dir, "host-wrapper.sh")
	content := "#!/bin/bash\n"
	if socketPath != "" {
		content += fmt.Sprintf("export SURF_SOCKET_PATH=%q\n", socketPath)
	}
	content += fmt.Sprintf("exec %q\n", hostPath)
	if err := os.WriteFile(wrapperPath, []byte(content), 0o755); err != nil {
		return "", pkgerrors.Wrap(err, "write wrapper")
	}
	return wrapperPath, nil
}

func writeManifest(dir, extensionID, wrapperPath string) (string, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", pkgerrors.Wrap(err, "create manifest directory")
	}
	manifestPath := filepath.Join(dir, HostName+".json")
	content := fmt.Sprintf("{\n  \"name\": %q,\n  \"description\": %q,\n  \"path\": %q,\n  \"type\": \"stdio\",\n  \"allowed_origins\": [%q]\n}\n",
		HostName,
		"Surf CLI Native Host",
		wrapperPath,
		"chrome-extension://"+extensionID+"/",
	)
	if err := os.WriteFile(manifestPath, []byte(content), 0o644); err != nil {
		return "", pkgerrors.Wrap(err, "write manifest")
	}
	return manifestPath, nil
}
