package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/nicobailon/surf-cli/gohost/internal/cli/transport"
)

func BuildToolRequest(tool string, args map[string]any, tabID *int64, windowID *int64) map[string]any {
	if args == nil {
		args = map[string]any{}
	}
	req := map[string]any{
		"type":   "tool_request",
		"method": "execute_tool",
		"params": map[string]any{
			"tool": tool,
			"args": args,
		},
		"id": fmt.Sprintf("go-%d", time.Now().UnixNano()),
	}
	if tabID != nil {
		req["tabId"] = *tabID
	}
	if windowID != nil {
		req["windowId"] = *windowID
	}
	return req
}

func ExecuteTool(
	ctx context.Context,
	client *transport.Client,
	tool string,
	args map[string]any,
	tabID *int64,
	windowID *int64,
) (map[string]any, error) {
	req := BuildToolRequest(tool, args, tabID, windowID)
	return client.Send(ctx, req)
}
