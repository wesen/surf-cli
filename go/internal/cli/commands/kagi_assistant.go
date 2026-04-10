package commands

import (
	"context"
	_ "embed"
	"encoding/json"
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

//go:embed scripts/kagi_assistant.js
var kagiAssistantScript string

const kagiAssistantURL = "https://kagi.com/assistant"
const kagiAssistantTimeoutMargin = 30 * time.Second

type KagiAssistantCommand struct {
	*cmds.CommandDescription
}

var _ cmds.GlazeCommand = (*KagiAssistantCommand)(nil)
var _ cmds.WriterCommand = (*KagiAssistantCommand)(nil)

type KagiAssistantSettings struct {
	Query                string `glazed:"query"`
	Assistant            string `glazed:"assistant"`
	Model                string `glazed:"model"`
	Lens                 string `glazed:"lens"`
	Tags                 string `glazed:"tags"`
	CreateTags           bool   `glazed:"create-tags"`
	WebSearchMode        string `glazed:"web-search-mode"`
	ListAssistants       bool   `glazed:"list-assistants"`
	ListCustomAssistants bool   `glazed:"list-custom-assistants"`
	ListModels           bool   `glazed:"list-models"`
	ListLenses           bool   `glazed:"list-lenses"`
	ListTags             bool   `glazed:"list-tags"`
	ListAllOptions       bool   `glazed:"list-all-options"`
	PromptTimeoutSec     int    `glazed:"prompt-timeout-sec"`
	KeepTabOpen          bool   `glazed:"keep-tab-open"`
	Socket               string `glazed:"socket-path"`
	TimeoutMS            int    `glazed:"timeout-ms"`
	TabID                int64  `glazed:"tab-id"`
	WindowID             int64  `glazed:"window-id"`
	DebugSocket          bool   `glazed:"debug-socket"`
}

type kagiAssistantData struct {
	Raw map[string]any
}

func NewKagiAssistantCommand() (*KagiAssistantCommand, error) {
	glazedSection, err := NewGlazedSchemaWithYAMLDefault()
	if err != nil {
		return nil, err
	}
	commandSection, err := cli.NewCommandSettingsSection()
	if err != nil {
		return nil, err
	}

	desc := cmds.NewCommandDescription(
		"kagi-assistant",
		cmds.WithShort("Run Kagi Assistant from your browser session"),
		cmds.WithLong("Uses the live Kagi Assistant browser session to list available assistants/models/lenses or submit a prompt with assistant, model, lens, and web-search options. By default it renders a Markdown report. Use --with-glaze-output for structured rows."),
		cmds.WithFlags(
			fields.New("assistant", fields.TypeString, fields.WithHelp("Built-in or custom assistant name to select before submitting the prompt")),
			fields.New("model", fields.TypeString, fields.WithHelp("Raw model to select before submitting the prompt; accepts the stable model id or visible model name")),
			fields.New("lens", fields.TypeString, fields.WithHelp("Kagi lens to select before submitting the prompt")),
			fields.New("tags", fields.TypeString, fields.WithHelp("Comma-separated tag names to apply to the created assistant thread")),
			fields.New("create-tags", fields.TypeBool, fields.WithDefault(false), fields.WithHelp("Create missing tags referenced by --tags")),
			fields.New("web-search-mode", fields.TypeChoice, fields.WithChoices("keep", "on", "off"), fields.WithDefault("keep"), fields.WithHelp("Whether to keep, enable, or disable the Web Search toggle")),
			fields.New("list-assistants", fields.TypeBool, fields.WithDefault(false), fields.WithHelp("List built-in Kagi assistants")),
			fields.New("list-custom-assistants", fields.TypeBool, fields.WithDefault(false), fields.WithHelp("List custom assistants")),
			fields.New("list-models", fields.TypeBool, fields.WithDefault(false), fields.WithHelp("List raw model options")),
			fields.New("list-lenses", fields.TypeBool, fields.WithDefault(false), fields.WithHelp("List available Kagi lenses")),
			fields.New("list-tags", fields.TypeBool, fields.WithDefault(false), fields.WithHelp("List available conversation tags")),
			fields.New("list-all-options", fields.TypeBool, fields.WithDefault(false), fields.WithHelp("List assistants, custom assistants, models, and lenses together")),
			fields.New("prompt-timeout-sec", fields.TypeInteger, fields.WithDefault(120), fields.WithHelp("Maximum time to wait for the assistant response in seconds")),
			fields.New("keep-tab-open", fields.TypeBool, fields.WithDefault(false), fields.WithHelp("Keep a newly created assistant tab open instead of closing it when the command finishes")),
			fields.New("socket-path", fields.TypeString, fields.WithDefault(config.CurrentSocketPath()), fields.WithHelp("Host socket path")),
			fields.New("timeout-ms", fields.TypeInteger, fields.WithDefault(30000), fields.WithHelp("Socket request timeout in milliseconds")),
			fields.New("tab-id", fields.TypeInteger, fields.WithDefault(int64(-1)), fields.WithHelp("Optional tab id override for the assistant page")),
			fields.New("window-id", fields.TypeInteger, fields.WithDefault(int64(-1)), fields.WithHelp("Optional window id override")),
			fields.New("debug-socket", fields.TypeBool, fields.WithDefault(false), fields.WithHelp("Log socket request/response frames to stderr")),
		),
		cmds.WithArguments(
			fields.New("query", fields.TypeString, fields.WithHelp("Prompt to send to Kagi Assistant")),
		),
		cmds.WithSections(glazedSection, commandSection),
	)

	return &KagiAssistantCommand{CommandDescription: desc}, nil
}

func (c *KagiAssistantCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
	s := &KagiAssistantSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return err
	}

	data, err := fetchKagiAssistant(ctx, s)
	if err != nil {
		return err
	}

	for _, row := range kagiAssistantDataToRows(data) {
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}
	return nil
}

func (c *KagiAssistantCommand) RunIntoWriter(ctx context.Context, vals *values.Values, w io.Writer) error {
	s := &KagiAssistantSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return err
	}

	data, err := fetchKagiAssistant(ctx, s)
	if err != nil {
		return err
	}

	_, err = io.WriteString(w, renderKagiAssistantMarkdown(data.Raw))
	return err
}

func hasKagiAssistantListMode(s *KagiAssistantSettings) bool {
	return s.ListAssistants || s.ListCustomAssistants || s.ListModels || s.ListLenses || s.ListTags || s.ListAllOptions
}

func validateKagiAssistantSettings(s *KagiAssistantSettings) error {
	if strings.TrimSpace(s.Assistant) != "" && strings.TrimSpace(s.Model) != "" {
		return fmt.Errorf("--assistant and --model are mutually exclusive")
	}
	if !hasKagiAssistantListMode(s) && strings.TrimSpace(s.Query) == "" {
		return fmt.Errorf("query required unless a list flag is set")
	}
	return nil
}

func buildKagiAssistantCode(s *KagiAssistantSettings) (string, error) {
	action := "run"
	switch {
	case s.ListAllOptions || (s.ListLenses && (s.ListAssistants || s.ListCustomAssistants || s.ListModels || s.ListTags)) || (s.ListTags && (s.ListAssistants || s.ListCustomAssistants || s.ListModels)):
		action = "list-all"
	case s.ListLenses && !s.ListAssistants && !s.ListCustomAssistants && !s.ListModels && !s.ListTags:
		action = "list-lenses"
	case s.ListTags && !s.ListAssistants && !s.ListCustomAssistants && !s.ListModels && !s.ListLenses:
		action = "list-tags"
	case hasKagiAssistantListMode(s):
		action = "list-profiles"
	}
	options := map[string]any{
		"action":               action,
		"prompt":               s.Query,
		"assistant":            s.Assistant,
		"model":                s.Model,
		"lens":                 s.Lens,
		"tags":                 splitCSVValues(s.Tags),
		"createTags":           s.CreateTags,
		"webSearchMode":        s.WebSearchMode,
		"listAssistants":       s.ListAssistants,
		"listCustomAssistants": s.ListCustomAssistants,
		"listModels":           s.ListModels,
		"listLenses":           s.ListLenses,
		"listTags":             s.ListTags,
		"listAllOptions":       s.ListAllOptions,
		"promptTimeoutMs":      s.PromptTimeoutSec * 1000,
	}
	b, err := json.Marshal(options)
	if err != nil {
		return "", fmt.Errorf("marshal kagi assistant options: %w", err)
	}
	return fmt.Sprintf("const SURF_OPTIONS = %s;\n%s", string(b), kagiAssistantScript), nil
}

func fetchKagiAssistant(ctx context.Context, s *KagiAssistantSettings) (data *kagiAssistantData, retErr error) {
	if err := validateKagiAssistantSettings(s); err != nil {
		return nil, err
	}
	code, err := buildKagiAssistantCode(s)
	if err != nil {
		return nil, err
	}

	socketTimeout := time.Duration(s.TimeoutMS) * time.Millisecond
	promptTimeout := time.Duration(s.PromptTimeoutSec) * time.Second
	if promptTimeout > 0 {
		minTimeout := promptTimeout + kagiAssistantTimeoutMargin
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
		tabResp, err := ExecuteTool(ctx, client, "tab.new", map[string]any{"url": kagiAssistantURL}, nil, nil)
		if err != nil {
			return nil, err
		}
		resolvedTabID, err := extractTabIDFromResponse(tabResp)
		if err != nil {
			return nil, err
		}
		tabID = &resolvedTabID
		ownedTabID = &resolvedTabID
	} else {
		if _, err := ExecuteTool(ctx, client, "navigate", map[string]any{"url": kagiAssistantURL}, tabID, windowID); err != nil {
			return nil, err
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

	time.Sleep(1200 * time.Millisecond)

	var resp map[string]any
	var respErr error
	for attempt := 0; attempt < 4; attempt++ {
		resp, respErr = ExecuteTool(ctx, client, "js", map[string]any{"code": code}, tabID, windowID)
		if respErr == nil {
			break
		}
		if !strings.Contains(respErr.Error(), "Cannot find default execution context") {
			return nil, respErr
		}
		time.Sleep(750 * time.Millisecond)
	}
	if respErr != nil {
		return nil, respErr
	}
	return parseKagiAssistantResponse(resp)
}

func parseKagiAssistantResponse(resp map[string]any) (*kagiAssistantData, error) {
	if e := extractErrorText(resp); e != "" {
		return nil, fmt.Errorf("%s", e)
	}
	parsed := parseResult(resp)
	dataMap, ok := parsed.Data.(map[string]any)
	if !ok {
		if parsed.Text != "" {
			return &kagiAssistantData{Raw: map[string]any{"content": parsed.Text}}, nil
		}
		return &kagiAssistantData{Raw: map[string]any{"content": nil}}, nil
	}
	return &kagiAssistantData{Raw: dataMap}, nil
}

func kagiAssistantDataToRows(data *kagiAssistantData) []types.Row {
	dataMap := data.Raw
	kind, _ := dataMap["kind"].(string)
	href, _ := dataMap["href"].(string)
	title, _ := dataMap["title"].(string)

	if kind == "options" {
		rows := make([]types.Row, 0)
		for _, key := range []string{"profiles", "lenses", "tags"} {
			items, _ := dataMap[key].([]any)
			for _, item := range items {
				m, ok := item.(map[string]any)
				if !ok {
					continue
				}
				rowMap := map[string]any{"href": href, "title": title}
				for k, v := range m {
					rowMap[k] = v
				}
				rows = append(rows, types.NewRowFromMap(rowMap))
			}
		}
		if len(rows) > 0 {
			return rows
		}
	}

	if kind == "response" {
		rowMap := map[string]any{
			"href":               href,
			"title":              title,
			"conversationTitle":  dataMap["conversationTitle"],
			"prompt":             dataMap["prompt"],
			"response":           dataMap["response"],
			"responseLength":     dataMap["responseLength"],
			"thinking":           dataMap["thinking"],
			"thinkingLength":     dataMap["thinkingLength"],
			"waitedMs":           dataMap["waitedMs"],
			"readOnly":           dataMap["readOnly"],
			"profileSelection":   dataMap["profileSelection"],
			"lensSelection":      dataMap["lensSelection"],
			"webSearchSelection": dataMap["webSearchSelection"],
			"tagSelection":       dataMap["tagSelection"],
		}
		if meta, ok := dataMap["metadata"].(map[string]any); ok {
			for _, pair := range []struct{ src, dst string }{
				{"Model", "modelUsed"},
				{"Version", "version"},
				{"Speed (tok/s)", "speedTokSec"},
				{"Tokens", "tokens"},
				{"Cost / Total ($)", "costTotal"},
				{"End to end time (s)", "endToEndTimeSec"},
				{"Submitted", "submittedAt"},
			} {
				if v, ok := meta[pair.src]; ok {
					rowMap[pair.dst] = v
				}
			}
			rowMap["metadata"] = meta
		}
		return []types.Row{types.NewRowFromMap(rowMap)}
	}

	return []types.Row{types.NewRowFromMap(dataMap)}
}

func renderKagiAssistantMarkdown(data map[string]any) string {
	var b strings.Builder
	kind, _ := data["kind"].(string)
	if kind == "options" {
		b.WriteString("# Kagi Assistant Options\n\n")
		if href, _ := data["href"].(string); href != "" {
			b.WriteString("- URL: ")
			b.WriteString(href)
			b.WriteString("\n")
		}
		for _, section := range []struct{ title, key string }{
			{"Assistants and Models", "profiles"},
			{"Lenses", "lenses"},
			{"Tags", "tags"},
		} {
			items, _ := data[section.key].([]any)
			if len(items) == 0 {
				continue
			}
			b.WriteString("\n## ")
			b.WriteString(section.title)
			b.WriteString("\n\n")
			for _, item := range items {
				m, ok := item.(map[string]any)
				if !ok {
					continue
				}
				name, _ := m["name"].(string)
				kind, _ := m["kind"].(string)
				sectionName, _ := m["section"].(string)
				modelID, _ := m["modelId"].(string)
				subtitle, _ := m["subtitle"].(string)
				selected, _ := m["selected"].(bool)
				b.WriteString("- ")
				b.WriteString(name)
				if kind != "" {
					b.WriteString(" (`")
					b.WriteString(kind)
					b.WriteString("`)")
				}
				if sectionName != "" {
					b.WriteString(" in ")
					b.WriteString(sectionName)
				}
				if modelID != "" {
					b.WriteString(" [id: `")
					b.WriteString(modelID)
					b.WriteString("`]")
				}
				if selected {
					b.WriteString(" [selected]")
				}
				b.WriteString("\n")
				if subtitle != "" {
					b.WriteString("  ")
					b.WriteString(subtitle)
					b.WriteString("\n")
				}
			}
		}
		return b.String()
	}

	b.WriteString("# Kagi Assistant\n\n")
	if href, _ := data["href"].(string); href != "" {
		b.WriteString("- URL: ")
		b.WriteString(href)
		b.WriteString("\n")
	}
	if waitedMs := data["waitedMs"]; waitedMs != nil {
		b.WriteString("- Waited: ")
		b.WriteString(fmt.Sprintf("%v ms", waitedMs))
		b.WriteString("\n")
	}
	if profileSelection, ok := data["profileSelection"].(map[string]any); ok {
		b.WriteString("- Profile: ")
		b.WriteString(fmt.Sprintf("%v", profileSelection["selected"]))
		b.WriteString("\n")
	}
	if lensSelection, ok := data["lensSelection"].(map[string]any); ok {
		b.WriteString("- Lens: ")
		b.WriteString(fmt.Sprintf("%v", lensSelection["selected"]))
		b.WriteString("\n")
	}
	if webSearchSelection, ok := data["webSearchSelection"].(map[string]any); ok {
		b.WriteString("- Web Search: ")
		b.WriteString(fmt.Sprintf("%v", webSearchSelection["after"]))
		b.WriteString("\n")
	}
	if tagSelection, ok := data["tagSelection"].(map[string]any); ok {
		b.WriteString("- Tags: ")
		b.WriteString(fmt.Sprintf("%v", tagSelection["visibleTags"]))
		b.WriteString("\n")
	}
	if prompt, _ := data["prompt"].(string); prompt != "" {
		b.WriteString("\n## Prompt\n\n")
		b.WriteString(prompt)
		b.WriteString("\n")
	}
	if response, _ := data["response"].(string); response != "" {
		b.WriteString("\n## Response\n\n")
		b.WriteString(response)
		b.WriteString("\n")
	}
	if thinking, _ := data["thinking"].(string); strings.TrimSpace(thinking) != "" {
		b.WriteString("\n## Thinking\n\n")
		b.WriteString(thinking)
		b.WriteString("\n")
	}
	if meta, ok := data["metadata"].(map[string]any); ok && len(meta) > 0 {
		b.WriteString("\n## Metadata\n\n")
		for _, key := range []string{"Model", "Version", "Speed (tok/s)", "Tokens", "Cost / Total ($)", "End to end time (s)", "Submitted"} {
			if v, ok := meta[key]; ok {
				b.WriteString("- ")
				b.WriteString(key)
				b.WriteString(": ")
				b.WriteString(fmt.Sprintf("%v", v))
				b.WriteString("\n")
			}
		}
	}
	return b.String()
}

func splitCSVValues(s string) []string {
	parts := strings.Split(s, ",")
	ret := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			ret = append(ret, part)
		}
	}
	return ret
}
