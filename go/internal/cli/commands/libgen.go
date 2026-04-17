package commands

import (
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
