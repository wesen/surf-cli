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
	rootCmd.AddCommand(newInstallCommand())
	rootCmd.AddCommand(newChatGPTCommand())

	rawCmd, err := commands.NewToolRawCommand()
	if err != nil {
		return nil, err
	}
	cobraRaw, err := buildGlazedCommand(rawCmd)
	if err != nil {
		return nil, err
	}
	rootCmd.AddCommand(cobraRaw)

	navigateCmd, err := commands.NewNavigateCommand()
	if err != nil {
		return nil, err
	}
	cobraNavigate, err := buildGlazedCommand(navigateCmd)
	if err != nil {
		return nil, err
	}
	rootCmd.AddCommand(cobraNavigate)

	if err := addPageAndInputCommands(rootCmd); err != nil {
		return nil, err
	}
	if err := addRemainingCoreCommands(rootCmd); err != nil {
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

type simpleSpec struct {
	Name  string
	Short string
	Tool  string
}

func addSimpleCommands(parent *cobra.Command, specs []simpleSpec) error {
	for _, spec := range specs {
		cmd, err := commands.NewSimpleToolCommand(spec.Name, spec.Short, spec.Tool, map[string]any{})
		if err != nil {
			return err
		}
		cobraCmd, err := buildGlazedCommand(cmd)
		if err != nil {
			return err
		}
		parent.AddCommand(cobraCmd)
	}
	return nil
}

func addPageAndInputCommands(root *cobra.Command) error {
	pageCmd := &cobra.Command{Use: "page", Short: "Page inspection commands"}
	waitCmd := &cobra.Command{Use: "wait", Short: "Wait helper commands"}

	pageSpecs := []simpleSpec{
		{"read", "Read interactive page snapshot", "page.read"},
		{"text", "Read plain page text", "page.text"},
		{"state", "Read page state metadata", "page.state"},
		{"search", "Search for text on the page", "search"},
	}
	if err := addSimpleCommands(pageCmd, pageSpecs); err != nil {
		return err
	}

	waitSpecs := []simpleSpec{
		{"element", "Wait for element selector", "wait.element"},
		{"url", "Wait for URL pattern", "wait.url"},
		{"network", "Wait for network idle", "wait.network"},
		{"dom", "Wait for DOM stability", "wait.dom"},
	}
	if err := addSimpleCommands(waitCmd, waitSpecs); err != nil {
		return err
	}

	root.AddCommand(pageCmd)
	root.AddCommand(waitCmd)

	inputSpecs := []simpleSpec{
		{"click", "Click using ref/selector/coordinates", "click"},
		{"type", "Type text into focused/input element", "type"},
		{"key", "Send keyboard key event", "key"},
		{"scroll", "Scroll page or element", "scroll"},
		{"hover", "Hover at element or coordinates", "hover"},
		{"drag", "Drag between coordinates", "drag"},
		{"select", "Select option(s) in a dropdown", "select"},
		{"screenshot", "Capture screenshot", "screenshot"},
	}
	if err := addSimpleCommands(root, inputSpecs); err != nil {
		return err
	}
	if err := addSimpleCommands(root, []simpleSpec{
		{"back", "Navigate back in history", "back"},
		{"forward", "Navigate forward in history", "forward"},
		{"reload", "Reload current tab", "tab.reload"},
	}); err != nil {
		return err
	}

	return nil
}

func addRemainingCoreCommands(root *cobra.Command) error {
	tabCmd := &cobra.Command{Use: "tab", Short: "Tab management commands"}
	if err := addSimpleCommands(tabCmd, []simpleSpec{
		{"list", "List tabs", "tab.list"},
		{"new", "Create a new tab", "tab.new"},
		{"switch", "Switch active tab", "tab.switch"},
		{"close", "Close one or more tabs", "tab.close"},
		{"name", "Assign name to active tab", "tab.name"},
		{"named", "List named tabs", "tab.named"},
	}); err != nil {
		return err
	}
	root.AddCommand(tabCmd)

	windowCmd := &cobra.Command{Use: "window", Short: "Browser window commands"}
	if err := addSimpleCommands(windowCmd, []simpleSpec{
		{"list", "List windows", "window.list"},
		{"new", "Create a new window", "window.new"},
		{"focus", "Focus a window", "window.focus"},
		{"close", "Close a window", "window.close"},
		{"resize", "Resize/move a window", "window.resize"},
	}); err != nil {
		return err
	}
	root.AddCommand(windowCmd)

	frameCmd := &cobra.Command{Use: "frame", Short: "Frame commands"}
	if err := addSimpleCommands(frameCmd, []simpleSpec{
		{"list", "List page frames", "frame.list"},
		{"switch", "Switch to a frame", "frame.switch"},
		{"main", "Switch to main frame", "frame.main"},
		{"eval", "Evaluate JS in current frame", "frame.js"},
	}); err != nil {
		return err
	}
	root.AddCommand(frameCmd)

	dialogCmd := &cobra.Command{Use: "dialog", Short: "Dialog commands"}
	if err := addSimpleCommands(dialogCmd, []simpleSpec{
		{"accept", "Accept the current dialog", "dialog.accept"},
		{"dismiss", "Dismiss the current dialog", "dialog.dismiss"},
		{"info", "Get current dialog info", "dialog.info"},
	}); err != nil {
		return err
	}
	root.AddCommand(dialogCmd)

	networkCmd := &cobra.Command{Use: "network", Short: "Network request commands"}
	if err := addSimpleCommands(networkCmd, []simpleSpec{
		{"list", "List captured network requests", "network"},
		{"get", "Get one captured request by id", "network.get"},
		{"body", "Get request/response body by id", "network.body"},
		{"origins", "List captured origins", "network.origins"},
		{"stats", "Show capture statistics", "network.stats"},
		{"clear", "Clear captured requests", "network.clear"},
		{"export", "Export captured requests", "network.export"},
	}); err != nil {
		return err
	}
	networkStream, err := commands.NewStreamCommand("stream", "Stream network events", "STREAM_NETWORK")
	if err != nil {
		return err
	}
	networkStreamCobra, err := buildGlazedCommand(networkStream)
	if err != nil {
		return err
	}
	networkCmd.AddCommand(networkStreamCobra)
	root.AddCommand(networkCmd)

	consoleCmd := &cobra.Command{Use: "console", Short: "Console log commands"}
	if err := addSimpleCommands(consoleCmd, []simpleSpec{
		{"read", "Read captured console messages", "console"},
	}); err != nil {
		return err
	}
	consoleStream, err := commands.NewStreamCommand("stream", "Stream console events", "STREAM_CONSOLE")
	if err != nil {
		return err
	}
	consoleStreamCobra, err := buildGlazedCommand(consoleStream)
	if err != nil {
		return err
	}
	consoleCmd.AddCommand(consoleStreamCobra)
	root.AddCommand(consoleCmd)

	cookieCmd := &cobra.Command{Use: "cookie", Short: "Cookie commands"}
	if err := addSimpleCommands(cookieCmd, []simpleSpec{
		{"list", "List cookies", "cookie.list"},
		{"get", "Get cookie by name", "cookie.get"},
		{"set", "Set cookie value", "cookie.set"},
		{"clear", "Clear cookie(s)", "cookie.clear"},
	}); err != nil {
		return err
	}
	root.AddCommand(cookieCmd)

	emulateCmd := &cobra.Command{Use: "emulate", Short: "Browser emulation commands"}
	if err := addSimpleCommands(emulateCmd, []simpleSpec{
		{"network", "Set network emulation profile", "emulate.network"},
		{"cpu", "Set CPU slowdown factor", "emulate.cpu"},
		{"geo", "Set geolocation", "emulate.geo"},
		{"device", "Set device emulation", "emulate.device"},
		{"viewport", "Set viewport emulation", "emulate.viewport"},
		{"touch", "Enable/disable touch emulation", "emulate.touch"},
	}); err != nil {
		return err
	}
	root.AddCommand(emulateCmd)

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
