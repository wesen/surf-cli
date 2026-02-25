package commands

import (
	"encoding/json"
	"fmt"

	"github.com/go-go-golems/glazed/pkg/types"
)

func ToolResponseToRow(resp map[string]any) types.Row {
	status := "ok"
	message := extractResultText(resp)
	errText := ""
	if e := extractErrorText(resp); e != "" {
		status = "error"
		errText = e
		message = ""
	}

	raw := ""
	if b, err := json.Marshal(resp); err == nil {
		raw = string(b)
	}

	return types.NewRow(
		types.MRP("status", status),
		types.MRP("message", message),
		types.MRP("error", errText),
		types.MRP("response", raw),
	)
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

func extractResultText(resp map[string]any) string {
	res, ok := resp["result"].(map[string]any)
	if !ok {
		b, _ := json.Marshal(resp)
		return string(b)
	}
	content, ok := res["content"].([]any)
	if !ok || len(content) == 0 {
		b, _ := json.Marshal(res)
		return string(b)
	}
	block, ok := content[0].(map[string]any)
	if !ok {
		b, _ := json.Marshal(content)
		return string(b)
	}
	if text, ok := block["text"].(string); ok {
		return text
	}
	b, _ := json.Marshal(block)
	return string(b)
}
