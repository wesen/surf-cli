package commands

import (
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/settings"
)

type outputDefaultSettings struct {
	Output string `glazed:"output"`
}

// NewGlazedSchemaWithYAMLDefault keeps glazed output settings intact while
// defaulting output format to yaml for surf-go commands.
func NewGlazedSchemaWithYAMLDefault() (schema.Section, error) {
	return settings.NewGlazedSchema(
		settings.WithOutputSectionOptions(
			schema.WithDefaults(&outputDefaultSettings{
				Output: "yaml",
			}),
		),
	)
}
