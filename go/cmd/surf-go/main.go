package main

import (
	"fmt"
	"os"

	"github.com/go-go-golems/glazed/pkg/cli"
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
	cobraRaw, err := cli.BuildCobraCommandFromCommand(rawCmd,
		cli.WithParserConfig(cli.CobraParserConfig{
			ShortHelpSections: []string{schema.DefaultSlug},
			MiddlewaresFunc:   cli.CobraCommandDefaultMiddlewares,
		}),
	)
	if err != nil {
		return nil, err
	}
	rootCmd.AddCommand(cobraRaw)

	return rootCmd, nil
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
