package commands

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/nicobailon/surf-cli/gohost/internal/cli/transport"
	"github.com/nicobailon/surf-cli/gohost/internal/host/config"
)

type FrameDiagnoseCommand struct {
	*cmds.CommandDescription
}

var _ cmds.GlazeCommand = (*FrameDiagnoseCommand)(nil)
var _ cmds.WriterCommand = (*FrameDiagnoseCommand)(nil)

type FrameDiagnoseSettings struct {
	Socket      string `glazed:"socket-path"`
	TimeoutMS   int    `glazed:"timeout-ms"`
	TabID       int64  `glazed:"tab-id"`
	WindowID    int64  `glazed:"window-id"`
	DebugSocket bool   `glazed:"debug-socket"`
}

type frameDiagnoseData struct {
	Raw map[string]any
}

func NewFrameDiagnoseCommand() (*FrameDiagnoseCommand, error) {
	glazedSection, err := NewGlazedSchemaWithYAMLDefault()
	if err != nil {
		return nil, err
	}
	commandSection, err := cli.NewCommandSettingsSection()
	if err != nil {
		return nil, err
	}

	desc := cmds.NewCommandDescription(
		"diagnose",
		cmds.WithShort("Diagnose frame discovery and reachability for the current page"),
		cmds.WithLong("Collects main-page iframe inventory, extension frame inventory, CDP frame inventory, and child-frame content-script reachability. By default it renders a human-readable Markdown report. Use --with-glaze-output for structured rows."),
		cmds.WithFlags(
			fields.New("socket-path", fields.TypeString, fields.WithDefault(config.CurrentSocketPath()), fields.WithHelp("Host socket path")),
			fields.New("timeout-ms", fields.TypeInteger, fields.WithDefault(60000), fields.WithHelp("Socket request timeout in milliseconds")),
			fields.New("tab-id", fields.TypeInteger, fields.WithDefault(int64(-1)), fields.WithHelp("Optional tab id override")),
			fields.New("window-id", fields.TypeInteger, fields.WithDefault(int64(-1)), fields.WithHelp("Optional window id override")),
			fields.New("debug-socket", fields.TypeBool, fields.WithDefault(false), fields.WithHelp("Log socket request/response frames to stderr")),
		),
		cmds.WithSections(glazedSection, commandSection),
	)

	return &FrameDiagnoseCommand{CommandDescription: desc}, nil
}

func (c *FrameDiagnoseCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
	s := &FrameDiagnoseSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return err
	}
	data, err := fetchFrameDiagnose(ctx, s)
	if err != nil {
		return err
	}
	for _, row := range frameDiagnoseToRows(data.Raw) {
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}
	return nil
}

func (c *FrameDiagnoseCommand) RunIntoWriter(ctx context.Context, vals *values.Values, w io.Writer) error {
	s := &FrameDiagnoseSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return err
	}
	data, err := fetchFrameDiagnose(ctx, s)
	if err != nil {
		return err
	}
	_, err = io.WriteString(w, renderFrameDiagnoseMarkdown(data.Raw))
	return err
}

func fetchFrameDiagnose(ctx context.Context, s *FrameDiagnoseSettings) (*frameDiagnoseData, error) {
	client := transport.NewClient(s.Socket, time.Duration(s.TimeoutMS)*time.Millisecond)
	client.Debug = s.DebugSocket

	var tabID *int64
	if s.TabID >= 0 {
		tabID = &s.TabID
	}
	var windowID *int64
	if s.WindowID >= 0 {
		windowID = &s.WindowID
	}

	resp, err := ExecuteTool(ctx, client, "frame.diagnose", map[string]any{}, tabID, windowID)
	if err != nil {
		return nil, err
	}
	if e := extractErrorText(resp); e != "" {
		return nil, fmt.Errorf("%s", e)
	}
	parsed := parseResult(resp)
	dataMap, ok := parsed.Data.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("unexpected frame diagnose response shape")
	}
	return &frameDiagnoseData{Raw: dataMap}, nil
}

func frameDiagnoseToRows(data map[string]any) []types.Row {
	rows := []types.Row{}
	if mainPage, ok := data["mainPage"].(map[string]any); ok {
		row := map[string]any{"kind": "summary"}
		for k, v := range mainPage {
			row[k] = v
		}
		rows = append(rows, types.NewRowFromMap(row))
	}

	appendRows := func(kind string, key string) {
		items, _ := data[key].([]any)
		for _, item := range items {
			m, ok := item.(map[string]any)
			if !ok {
				continue
			}
			row := map[string]any{"kind": kind}
			for k, v := range m {
				row[k] = v
			}
			rows = append(rows, types.NewRowFromMap(row))
		}
	}

	appendRows("dom_iframe", "domIframes")
	appendRows("extension_frame", "extensionFrames")
	appendRows("cdp_frame", "cdpFrames")

	if warnings, ok := data["warnings"].([]any); ok {
		for _, item := range warnings {
			rows = append(rows, types.NewRow(
				types.MRP("kind", "warning"),
				types.MRP("message", item),
			))
		}
	}
	return rows
}

func renderFrameDiagnoseMarkdown(data map[string]any) string {
	var b strings.Builder
	b.WriteString("# Frame Diagnose\n\n")
	if mainPage, ok := data["mainPage"].(map[string]any); ok {
		b.WriteString("## Main Page\n\n")
		fmt.Fprintf(&b, "- href: %v\n", mainPage["href"])
		fmt.Fprintf(&b, "- title: %v\n", mainPage["title"])
		fmt.Fprintf(&b, "- iframeCount: %v\n\n", mainPage["iframeCount"])
	}

	renderList := func(title string, items []any, fields []string) {
		b.WriteString("## " + title + "\n\n")
		if len(items) == 0 {
			b.WriteString("_None._\n\n")
			return
		}
		for i, item := range items {
			m, ok := item.(map[string]any)
			if !ok {
				continue
			}
			fmt.Fprintf(&b, "%d.\n", i+1)
			for _, field := range fields {
				if v, exists := m[field]; exists {
					fmt.Fprintf(&b, "   - %s: %v\n", field, v)
				}
			}
		}
		b.WriteString("\n")
	}

	domIframes, _ := data["domIframes"].([]any)
	renderList("DOM Iframes", domIframes, []string{"domIndex", "src", "title", "name", "id", "sandbox", "allow", "rect"})

	extensionFrames, _ := data["extensionFrames"].([]any)
	renderList("Extension Frames", extensionFrames, []string{"extensionFrameId", "parentFrameId", "url", "errorOccurred", "contentScriptReachable", "contentScriptError", "contentScript"})

	cdpFrames, _ := data["cdpFrames"].([]any)
	renderList("CDP Frames", cdpFrames, []string{"cdpFrameId", "parentId", "name", "url"})

	warnings, _ := data["warnings"].([]any)
	b.WriteString("## Warnings\n\n")
	if len(warnings) == 0 {
		b.WriteString("_None._\n")
	} else {
		for _, warning := range warnings {
			fmt.Fprintf(&b, "- %v\n", warning)
		}
	}
	b.WriteString("\n")

	return b.String()
}
