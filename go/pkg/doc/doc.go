package doc

import (
	"embed"
	_ "embed"

	"github.com/go-go-golems/glazed/pkg/help"
)

//go:embed tutorials/*.md
var docFS embed.FS

func AddDocToHelpSystem(helpSystem *help.HelpSystem) error {
	return helpSystem.LoadSectionsFromFS(docFS, ".")
}
