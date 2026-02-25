package commands

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-go-golems/glazed/pkg/types"
)

func ToolResponseToRow(tool string, resp map[string]any) types.Row {
	status := "ok"
	errText := ""
	text := ""
	data := any(nil)
	content := any(nil)
	result := any(nil)
	dataKind := "none"
	dataCount := int64(0)

	if e := extractErrorText(resp); e != "" {
		status = "error"
		errText = e
	} else {
		parsed := parseResult(resp)
		text = parsed.Text
		data = parsed.Data
		content = parsed.Content
		result = resp["result"]
		dataKind, dataCount = classifyData(data)
	}

	return types.NewRow(
		types.MRP("tool", tool),
		types.MRP("status", status),
		types.MRP("id", resp["id"]),
		types.MRP("error", errText),
		types.MRP("text", text),
		types.MRP("data_kind", dataKind),
		types.MRP("data_count", dataCount),
		types.MRP("data", data),
		types.MRP("content", content),
		types.MRP("result", result),
	)
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

func classifyData(v any) (string, int64) {
	switch d := v.(type) {
	case map[string]any:
		return "object", int64(len(d))
	case []any:
		return "array", int64(len(d))
	default:
		return "none", 0
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
