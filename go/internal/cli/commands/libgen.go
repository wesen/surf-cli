package commands

import (
	"fmt"
	"os"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
)

type LibgenCommand struct {
	*cmds.CommandDescription
}

var _ cmds.Command = (*LibgenCommand)(nil)

func NewLibgenCommand() (*LibgenCommand, error) {
	commandSection, err := cli.NewCommandSettingsSection()
	if err != nil {
		return nil, err
	}

	desc := cmds.NewCommandDescription(
		"libgen",
		cmds.WithShort("1lib.sk (Z-Library) commands"),
		cmds.WithLong("Commands for 1lib.sk - a Z-Library mirror. Search, download, and explore book collections."),
		cmds.WithSections(commandSection),
	)

	return &LibgenCommand{CommandDescription: desc}, nil
}

func (c *LibgenCommand) Run(_ interface{}, _ []string) error {
	fmt.Fprintf(os.Stderr, "# 1lib.sk (Z-Library)\n\nUse `surf-go libgen search`, `libgen download`, `libgen suggestions`, `libgen collections`, or `libgen collection`\n")
	return nil
}
