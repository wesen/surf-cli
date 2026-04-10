package commands

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/nicobailon/surf-cli/gohost/internal/cli/transport"
	"github.com/nicobailon/surf-cli/gohost/internal/host/config"
)

//go:embed scripts/gmail_search.js
var gmailSearchScript string

type GmailSearchCommand struct {
	*cmds.CommandDescription
}

var _ cmds.GlazeCommand = (*GmailSearchCommand)(nil)
var _ cmds.WriterCommand = (*GmailSearchCommand)(nil)

type GmailSearchSettings struct {
	Query       string `glazed:"query"`
	MaxResults  int    `glazed:"max-results"`
	KeepTabOpen bool   `glazed:"keep-tab-open"`
	Socket      string `glazed:"socket-path"`
	TimeoutMS   int    `glazed:"timeout-ms"`
	TabID       int64  `glazed:"tab-id"`
	WindowID    int64  `glazed:"window-id"`
	DebugSocket bool   `glazed:"debug-socket"`
}

type gmailSearchData struct {
	Raw map[string]any
}

func NewGmailSearchCommand() (*GmailSearchCommand, error) {
	glazedSection, err := NewGlazedSchemaWithYAMLDefault()
	if err != nil {
		return nil, err
	}
	commandSection, err := cli.NewCommandSettingsSection()
	if err != nil {
		return nil, err
	}

	desc := cmds.NewCommandDescription(
		"search",
		cmds.WithShort("Run a Gmail search"),
		cmds.WithLong("Opens or reuses a Gmail tab, submits a Gmail search query through the live UI, and extracts thread summary rows. By default it renders a Markdown report. Use --with-glaze-output for structured rows."),
		cmds.WithFlags(
			fields.New("max-results", fields.TypeInteger, fields.WithDefault(25), fields.WithHelp("Maximum number of search rows to extract; 0 means all visible rows")),
			fields.New("keep-tab-open", fields.TypeBool, fields.WithDefault(false), fields.WithHelp("Keep a newly created Gmail tab open instead of closing it when the command finishes")),
			fields.New("socket-path", fields.TypeString, fields.WithDefault(config.CurrentSocketPath()), fields.WithHelp("Host socket path")),
			fields.New("timeout-ms", fields.TypeInteger, fields.WithDefault(120000), fields.WithHelp("Socket request timeout in milliseconds")),
			fields.New("tab-id", fields.TypeInteger, fields.WithDefault(int64(-1)), fields.WithHelp("Optional tab id override")),
			fields.New("window-id", fields.TypeInteger, fields.WithDefault(int64(-1)), fields.WithHelp("Optional window id override")),
			fields.New("debug-socket", fields.TypeBool, fields.WithDefault(false), fields.WithHelp("Log socket request/response frames to stderr")),
		),
		cmds.WithArguments(
			fields.New("query", fields.TypeString, fields.WithRequired(true), fields.WithHelp("Gmail search query")),
		),
		cmds.WithSections(glazedSection, commandSection),
	)

	return &GmailSearchCommand{CommandDescription: desc}, nil
}

func buildGmailSearchCode(s *GmailSearchSettings) (string, error) {
	options := map[string]any{
		"query":      s.Query,
		"maxResults": s.MaxResults,
	}
	b, err := json.Marshal(options)
	if err != nil {
		return "", fmt.Errorf("marshal gmail search options: %w", err)
	}
	return fmt.Sprintf("const SURF_OPTIONS = %s;\n%s", string(b), gmailSearchScript), nil
}

func (c *GmailSearchCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
	s := &GmailSearchSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return err
	}
	data, err := fetchGmailSearch(ctx, s)
	if err != nil {
		return err
	}
	for _, row := range gmailThreadsToRows(data.Raw, "search") {
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}
	return nil
}

func (c *GmailSearchCommand) RunIntoWriter(ctx context.Context, vals *values.Values, w io.Writer) error {
	s := &GmailSearchSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return err
	}
	data, err := fetchGmailSearch(ctx, s)
	if err != nil {
		return err
	}
	_, err = io.WriteString(w, renderGmailThreadsMarkdown(data.Raw, "Gmail Search"))
	return err
}

func fetchGmailSearch(ctx context.Context, s *GmailSearchSettings) (data *gmailSearchData, retErr error) {
	if s.Query == "" {
		return nil, fmt.Errorf("query required")
	}
	code, err := buildGmailSearchCode(s)
	if err != nil {
		return nil, err
	}

	client := transport.NewClient(s.Socket, time.Duration(s.TimeoutMS)*time.Millisecond)
	client.Debug = s.DebugSocket

	var tabID *int64
	var ownedTabID *int64
	if s.TabID >= 0 {
		tabID = &s.TabID
	}
	var windowID *int64
	if s.WindowID >= 0 {
		windowID = &s.WindowID
	}

	if tabID == nil && windowID == nil {
		resolvedTabID, err := openOwnedTab(ctx, client, gmailInboxURL, tabReadyOptions{URLPrefix: gmailURLPrefix})
		if err != nil {
			return nil, err
		}
		tabID = &resolvedTabID
		ownedTabID = &resolvedTabID
	} else {
		if _, err := ExecuteTool(ctx, client, "navigate", map[string]any{"url": gmailInboxURL}, tabID, windowID); err != nil {
			return nil, err
		}
		if tabID != nil {
			if err := waitForTabReady(ctx, client, *tabID, tabReadyOptions{URLPrefix: gmailURLPrefix}); err != nil {
				return nil, err
			}
		}
	}

	defer func() {
		if retErr != nil || s.KeepTabOpen {
			return
		}
		if err := closeOwnedTab(ctx, client, ownedTabID); err != nil {
			retErr = err
		}
	}()

	resp, err := ExecuteTool(ctx, client, "js", map[string]any{"code": code}, tabID, windowID)
	if err != nil {
		return nil, err
	}
	parsed, err := parseGmailDataResponse(resp)
	if err != nil {
		return nil, err
	}
	return &gmailSearchData{Raw: parsed}, nil
}
