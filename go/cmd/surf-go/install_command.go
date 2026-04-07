package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"

	"github.com/nicobailon/surf-cli/gohost/internal/installer"
	pkgerrors "github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func newInstallCommand() *cobra.Command {
	var browsersArg string
	var hostBinary string

	cmd := &cobra.Command{
		Use:   "install <extension-id>",
		Short: "Build and install the Go native messaging host",
		Long: "Builds surf-host-go from this checkout and installs native messaging manifests for the chosen browsers. " +
			"On Linux Chromium, this also installs the snap target when ~/snap/chromium/common exists.",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			extensionID := strings.TrimSpace(args[0])
			browsers := parseBrowsersArg(browsersArg)

			hostPath := hostBinary
			if hostPath == "" {
				var err error
				hostPath, err = buildHostFromCheckout()
				if err != nil {
					return err
				}
			}

			homeDir, err := os.UserHomeDir()
			if err != nil {
				return pkgerrors.Wrap(err, "resolve home directory")
			}

			result, err := installer.InstallGoNativeHost(installer.InstallOptions{
				ExtensionID: extensionID,
				Browsers:    browsers,
				HostBinary:  hostPath,
				HomeDir:     homeDir,
				GOOS:        runtime.GOOS,
			})
			if err != nil {
				return err
			}

			fmt.Printf("Platform: %s\n", runtime.GOOS)
			fmt.Printf("Go host: %s\n\n", hostPath)
			if len(result.Installed) > 0 {
				fmt.Println("Installed for:")
				for _, target := range result.Installed {
					fmt.Printf("  %s: %s\n", target.Browser, target.Path)
				}
			}
			if len(result.Skipped) > 0 {
				fmt.Printf("\nSkipped: %s\n", strings.Join(result.Skipped, ", "))
			}
			if len(result.Hints) > 0 {
				fmt.Println("\nHints:")
				for _, hint := range result.Hints {
					fmt.Printf("  - %s\n", hint)
				}
			}
			fmt.Println("\nDone! Restart your browser for changes to take effect.")
			return nil
		},
	}

	cmd.Flags().StringVarP(&browsersArg, "browser", "b", "chrome", "Browser(s) to install for: chrome, chromium, brave, edge, all")
	cmd.Flags().StringVar(&hostBinary, "host-binary", "", "Existing surf-host-go binary to install instead of building from source")
	return cmd
}

func parseBrowsersArg(v string) []string {
	browsers := strings.TrimSpace(v)
	if browsers == "" {
		return []string{"chrome"}
	}
	if browsers == "all" {
		return []string{"chrome", "chromium", "brave", "edge"}
	}
	parts := strings.Split(browsers, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.ToLower(strings.TrimSpace(part))
		if part == "" {
			continue
		}
		if !slices.Contains(result, part) {
			result = append(result, part)
		}
	}
	if len(result) == 0 {
		return []string{"chrome"}
	}
	return result
}

func buildHostFromCheckout() (string, error) {
	goRoot, err := findGoCheckoutRoot()
	if err != nil {
		return "", err
	}
	goExe, err := exec.LookPath("go")
	if err != nil {
		return "", pkgerrors.Wrap(err, "find go toolchain in PATH")
	}

	tempDir, err := os.MkdirTemp("", "surf-host-go-build-")
	if err != nil {
		return "", pkgerrors.Wrap(err, "create temporary build directory")
	}
	outputPath := filepath.Join(tempDir, "surf-host-go")
	if runtime.GOOS == "windows" {
		outputPath += ".exe"
	}

	buildCmd := exec.Command(goExe, "build", "-o", outputPath, "./cmd/surf-host-go")
	buildCmd.Dir = goRoot
	buildCmd.Env = append(os.Environ(), "CGO_ENABLED=0")
	if output, err := buildCmd.CombinedOutput(); err != nil {
		return "", pkgerrors.Wrapf(err, "build surf-host-go from %s: %s", goRoot, strings.TrimSpace(string(output)))
	}
	return outputPath, nil
}

func findGoCheckoutRoot() (string, error) {
	candidates := []string{}
	if cwd, err := os.Getwd(); err == nil {
		candidates = append(candidates, cwd)
	}
	if exe, err := os.Executable(); err == nil {
		candidates = append(candidates, filepath.Dir(exe))
	}
	if _, file, _, ok := runtime.Caller(0); ok {
		candidates = append(candidates, filepath.Dir(file))
	}

	seen := map[string]struct{}{}
	for _, start := range candidates {
		start = filepath.Clean(start)
		if _, ok := seen[start]; ok {
			continue
		}
		seen[start] = struct{}{}
		for dir := start; dir != filepath.Dir(dir); dir = filepath.Dir(dir) {
			if fileExists(filepath.Join(dir, "go.mod")) && fileExists(filepath.Join(dir, "cmd", "surf-host-go", "main.go")) {
				return dir, nil
			}
		}
	}
	return "", fmt.Errorf("could not locate surf go checkout root; run from the repository or pass --host-binary")
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
