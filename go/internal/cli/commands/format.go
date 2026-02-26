package commands

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-go-golems/glazed/pkg/types"
)

func ToolResponseToRows(resp map[string]any) []types.Row {
	if e := extractErrorText(resp); e != "" {
		return []types.Row{types.NewRow(types.MRP("error", e))}
	}

	parsed := parseResult(resp)
	switch data := parsed.Data.(type) {
	case map[string]any:
		return []types.Row{types.NewRowFromMap(data)}
	case []any:
		rows := make([]types.Row, 0, len(data))
		for _, item := range data {
			switch v := item.(type) {
			case map[string]any:
				rows = append(rows, types.NewRowFromMap(v))
			default:
				rows = append(rows, types.NewRow(types.MRP("content", v)))
			}
		}
		if len(rows) > 0 {
			return rows
		}
	}

	if strings.TrimSpace(parsed.Text) != "" {
		return []types.Row{types.NewRow(types.MRP("content", parsed.Text))}
	}
	if len(parsed.Content) > 0 {
		return []types.Row{types.NewRow(types.MRP("content", parsed.Content))}
	}

	return []types.Row{types.NewRow(types.MRP("content", nil))}
}

type parsedResult struct {
	Text    string
	Data    any
	Content []any
}

func parseResult(resp map[string]any) parsedResult {
	pr := parsedResult{}
	res, ok := resp["result"].(map[string]any)
	if !ok {
		return pr
	}
	content, ok := res["content"].([]any)
	if !ok {
		return pr
	}
	pr.Content = content

	textParts := make([]string, 0, len(content))
	for _, item := range content {
		block, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if blockType, _ := block["type"].(string); blockType == "text" {
			if s, ok := block["text"].(string); ok {
				textParts = append(textParts, s)
			}
		}
	}
	pr.Text = strings.Join(textParts, "\n")
	pr.Data = parseStructuredText(pr.Text)
	return pr
}

func parseStructuredText(text string) any {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return nil
	}
	var v any
	if err := json.Unmarshal([]byte(trimmed), &v); err != nil {
		return nil
	}
	switch v.(type) {
	case map[string]any, []any:
		return v
	default:
		return nil
	}
}

func extractErrorText(resp map[string]any) string {
	rawErr, ok := resp["error"]
	if !ok || rawErr == nil {
		return ""
	}
	if s, ok := rawErr.(string); ok {
		return s
	}
	em, ok := rawErr.(map[string]any)
	if !ok {
		return fmt.Sprintf("%v", rawErr)
	}
	content, ok := em["content"].([]any)
	if !ok || len(content) == 0 {
		b, _ := json.Marshal(em)
		return string(b)
	}
	block, ok := content[0].(map[string]any)
	if !ok {
		return "unknown error"
	}
	if text, ok := block["text"].(string); ok {
		return text
	}
	return "unknown error"
}
