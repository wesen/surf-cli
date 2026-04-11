package commands

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
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

//go:embed scripts/claude_ask.js
var claudeAskScript string

const (
	claudeNewURL           = "https://claude.ai/new"
	claudeAskTimeoutMargin = 30 * time.Second
)

type ClaudeCommand struct {
	*cmds.CommandDescription
}

type ClaudeSettings struct {
	Query            string `glazed:"query"`
	Model            string `glazed:"model"`
	ListModels       bool   `glazed:"list-models"`
	ThinkingMode     string `glazed:"thinking-mode"`
	PromptTimeoutSec int    `glazed:"prompt-timeout-sec"`
	KeepTabOpen      bool   `glazed:"keep-tab-open"`
	Socket           string `glazed:"socket-path"`
	TimeoutMS        int    `glazed:"timeout-ms"`
	TabID            int64  `glazed:"tab-id"`
	WindowID         int64  `glazed:"window-id"`
	DebugSocket      bool   `glazed:"debug-socket"`
}

type claudeData struct {
	Raw map[string]any
}

func NewClaudeCommand() (*ClaudeCommand, error) {
	glazedSection, err := NewGlazedSchemaWithYAMLDefault()
	if err != nil {
		return nil, err
	}
	commandSection, err := cli.NewCommandSettingsSection()
	if err != nil {
		return nil, err
	}

	desc := cmds.NewCommandDescription(
		"claude",
		cmds.WithShort("Send a prompt to Claude using your browser session"),
		cmds.WithLong("Uses the live claude.ai browser session to submit prompts, optionally select a model and thinking mode, or list available model options."),
		cmds.WithFlags(
			fields.New("model", fields.TypeString, fields.WithHelp("Claude model option to select before submitting the prompt")),
			fields.New("list-models", fields.TypeBool, fields.WithDefault(false), fields.WithHelp("List available Claude model options without sending a prompt")),
			fields.New("thinking-mode", fields.TypeChoice, fields.WithChoices("default", "standard", "extended"), fields.WithDefault("default"), fields.WithHelp("Claude thinking mode to apply before sending the prompt")),
			fields.New("prompt-timeout-sec", fields.TypeInteger, fields.WithDefault(180), fields.WithHelp("Maximum time to wait for the Claude response in seconds")),
			fields.New("keep-tab-open", fields.TypeBool, fields.WithDefault(false), fields.WithHelp("Keep a newly created Claude tab open instead of closing it when the command finishes")),
			fields.New("socket-path", fields.TypeString, fields.WithDefault(config.CurrentSocketPath()), fields.WithHelp("Host socket path")),
			fields.New("timeout-ms", fields.TypeInteger, fields.WithDefault(30000), fields.WithHelp("Socket request timeout in milliseconds")),
			fields.New("tab-id", fields.TypeInteger, fields.WithDefault(int64(-1)), fields.WithHelp("Optional tab id override for the Claude page")),
			fields.New("window-id", fields.TypeInteger, fields.WithDefault(int64(-1)), fields.WithHelp("Optional window id override")),
			fields.New("debug-socket", fields.TypeBool, fields.WithDefault(false), fields.WithHelp("Log socket request/response frames to stderr")),
		),
		cmds.WithArguments(
			fields.New("query", fields.TypeString, fields.WithHelp("Prompt to send to Claude")),
		),
		cmds.WithSections(glazedSection, commandSection),
	)

	return &ClaudeCommand{CommandDescription: desc}, nil
}

func buildClaudeCode(s *ClaudeSettings) (string, error) {
	action := "run"
	if s.ListModels {
		action = "list-models"
	}
	options := map[string]any{
		"action":          action,
		"prompt":          s.Query,
		"model":           s.Model,
		"thinkingMode":    s.ThinkingMode,
		"promptTimeoutMs": s.PromptTimeoutSec * 1000,
	}
	b, err := json.Marshal(options)
	if err != nil {
		return "", fmt.Errorf("marshal Claude options: %w", err)
	}
	return fmt.Sprintf("const SURF_OPTIONS = %s;\n%s", string(b), claudeAskScript), nil
}

func (c *ClaudeCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
	s := &ClaudeSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return err
	}
	if strings.TrimSpace(s.Query) == "" && !s.ListModels {
		return fmt.Errorf("query required unless --list-models is set")
	}

	data, err := fetchClaude(ctx, s)
	if err != nil {
		return err
	}

	for _, row := range claudeDataToRows(data) {
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}
	return nil
}

func fetchClaude(ctx context.Context, s *ClaudeSettings) (data *claudeData, retErr error) {
	code, err := buildClaudeCode(s)
	if err != nil {
		return nil, err
	}

	socketTimeout := time.Duration(s.TimeoutMS) * time.Millisecond
	promptTimeout := time.Duration(s.PromptTimeoutSec) * time.Second
	if promptTimeout > 0 {
		minTimeout := promptTimeout + claudeAskTimeoutMargin
		if socketTimeout < minTimeout {
			socketTimeout = minTimeout
		}
	}
	client := transport.NewClient(s.Socket, socketTimeout)
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
		resolvedTabID, err := openOwnedTab(ctx, client, claudeNewURL, tabReadyOptions{URLExact: claudeNewURL})
		if err != nil {
			return nil, err
		}
		tabID = &resolvedTabID
		ownedTabID = &resolvedTabID
	} else {
		if _, err := ExecuteTool(ctx, client, "navigate", map[string]any{"url": claudeNewURL}, tabID, windowID); err != nil {
			return nil, err
		}
		if tabID != nil {
			if err := waitForTabReady(ctx, client, *tabID, tabReadyOptions{URLExact: claudeNewURL}); err != nil {
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
	return parseClaudeResponse(resp)
}

func parseClaudeResponse(resp map[string]any) (*claudeData, error) {
	if e := extractErrorText(resp); e != "" {
		return nil, fmt.Errorf("%s", e)
	}
	parsed := parseResult(resp)
	dataMap, ok := parsed.Data.(map[string]any)
	if !ok {
		if parsed.Text != "" {
			return &claudeData{Raw: map[string]any{"content": parsed.Text}}, nil
		}
		return &claudeData{Raw: map[string]any{"content": nil}}, nil
	}
	if kind, _ := dataMap["kind"].(string); kind == "error" {
		if msg, _ := dataMap["error"].(string); msg != "" {
			return nil, fmt.Errorf("%s", msg)
		}
	}
	return &claudeData{Raw: dataMap}, nil
}

func claudeDataToRows(data *claudeData) []types.Row {
	dataMap := data.Raw
	kind, _ := dataMap["kind"].(string)

	if kind == "models" {
		currentModel, _ := dataMap["currentModel"].(string)
		items, _ := dataMap["models"].([]any)
		rows := make([]types.Row, 0, len(items))
		for _, item := range items {
			m, ok := item.(map[string]any)
			if !ok {
				continue
			}
			name, _ := m["name"].(string)
			currentThinkingMode, _ := dataMap["currentThinkingMode"].(string)
			rowMap := map[string]any{
				"href":                dataMap["href"],
				"title":               dataMap["title"],
				"currentModel":        currentModel,
				"currentThinkingMode": currentThinkingMode,
				"model":               name,
				"selected":            currentModel == name,
				"description":         m["description"],
				"rawText":             m["rawText"],
				"thinkingModes":       m["thinkingModes"],
			}
			rows = append(rows, types.NewRowFromMap(rowMap))
		}
		if len(rows) > 0 {
			return rows
		}
	}

	if kind == "response" {
		return []types.Row{types.NewRowFromMap(dataMap)}
	}
	return []types.Row{types.NewRowFromMap(dataMap)}
}
