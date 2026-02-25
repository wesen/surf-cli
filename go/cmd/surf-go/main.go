package main

import (
	"fmt"
	"os"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/logging"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/help"
	help_cmd "github.com/go-go-golems/glazed/pkg/help/cmd"
	"github.com/nicobailon/surf-cli/gohost/internal/cli/commands"
	"github.com/spf13/cobra"
)

func newRootCommand(helpSystem *help.HelpSystem) (*cobra.Command, error) {
	rootCmd := &cobra.Command{
		Use:   "surf-go",
		Short: "Go CLI for surf native host",
		Long:  "Go-based CLI for core browser commands routed through the local surf native host.",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return logging.InitLoggerFromCobra(cmd)
		},
	}

	help_cmd.SetupCobraRootCommand(helpSystem, rootCmd)
	_ = logging.AddLoggingSectionToRootCommand(rootCmd, "surf-go")

	rawCmd, err := commands.NewToolRawCommand()
	if err != nil {
		return nil, err
	}
	cobraRaw, err := buildGlazedCommand(rawCmd)
	if err != nil {
		return nil, err
	}
	rootCmd.AddCommand(cobraRaw)

	if err := addPageAndInputCommands(rootCmd); err != nil {
		return nil, err
	}

	return rootCmd, nil
}

func buildGlazedCommand(cmd cmds.Command) (*cobra.Command, error) {
	return cli.BuildCobraCommandFromCommand(cmd,
		cli.WithParserConfig(cli.CobraParserConfig{
			ShortHelpSections: []string{schema.DefaultSlug},
			MiddlewaresFunc:   cli.CobraCommandDefaultMiddlewares,
		}),
	)
}

func addPageAndInputCommands(root *cobra.Command) error {
	pageCmd := &cobra.Command{Use: "page", Short: "Page inspection commands"}
	waitCmd := &cobra.Command{Use: "wait", Short: "Wait helper commands"}

	pageSpecs := []struct {
		Name     string
		Short    string
		Tool     string
		Defaults map[string]any
	}{
		{"read", "Read interactive page snapshot", "page.read", map[string]any{}},
		{"text", "Read plain page text", "page.text", map[string]any{}},
		{"state", "Read page state metadata", "page.state", map[string]any{}},
		{"search", "Search for text on the page", "search", map[string]any{}},
	}
	for _, spec := range pageSpecs {
		cmd, err := commands.NewSimpleToolCommand(spec.Name, spec.Short, spec.Tool, spec.Defaults)
		if err != nil {
			return err
		}
		cobraCmd, err := buildGlazedCommand(cmd)
		if err != nil {
			return err
		}
		pageCmd.AddCommand(cobraCmd)
	}

	waitSpecs := []struct {
		Name  string
		Short string
		Tool  string
	}{
		{"element", "Wait for element selector", "wait.element"},
		{"url", "Wait for URL pattern", "wait.url"},
		{"network", "Wait for network idle", "wait.network"},
		{"dom", "Wait for DOM stability", "wait.dom"},
	}
	for _, spec := range waitSpecs {
		cmd, err := commands.NewSimpleToolCommand(spec.Name, spec.Short, spec.Tool, map[string]any{})
		if err != nil {
			return err
		}
		cobraCmd, err := buildGlazedCommand(cmd)
		if err != nil {
			return err
		}
		waitCmd.AddCommand(cobraCmd)
	}

	root.AddCommand(pageCmd)
	root.AddCommand(waitCmd)

	inputSpecs := []struct {
		Name  string
		Short string
		Tool  string
	}{
		{"click", "Click using ref/selector/coordinates", "click"},
		{"type", "Type text into focused/input element", "type"},
		{"key", "Send keyboard key event", "key"},
		{"scroll", "Scroll page or element", "scroll"},
		{"hover", "Hover at element or coordinates", "hover"},
		{"drag", "Drag between coordinates", "drag"},
		{"select", "Select option(s) in a dropdown", "select"},
		{"screenshot", "Capture screenshot", "screenshot"},
	}
	for _, spec := range inputSpecs {
		cmd, err := commands.NewSimpleToolCommand(spec.Name, spec.Short, spec.Tool, map[string]any{})
		if err != nil {
			return err
		}
		cobraCmd, err := buildGlazedCommand(cmd)
		if err != nil {
			return err
		}
		root.AddCommand(cobraCmd)
	}

	return nil
}

func main() {
	helpSystem := help.NewHelpSystem()
	rootCmd, err := newRootCommand(helpSystem)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error constructing root command: %v\n", err)
		os.Exit(1)
	}
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
