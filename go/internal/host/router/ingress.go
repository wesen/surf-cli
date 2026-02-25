package router

import (
	"fmt"
	"strconv"
)

type ToolRequest struct {
	Type          string
	Method        string
	Params        ToolParams
	OriginalID    any
	HasOriginalID bool
	TabID         *int64
	WindowID      *int64
}

type ToolParams struct {
	Tool string
	Args map[string]any
}

type StreamRequest struct {
	Type       string
	StreamType string
	Options    map[string]any
}

func ParseToolRequest(msg map[string]any) (ToolRequest, error) {
	msgType, ok := msg["type"].(string)
	if !ok || msgType != "tool_request" {
		return ToolRequest{}, fmt.Errorf("invalid tool_request: missing type=tool_request")
	}

	method, ok := msg["method"].(string)
	if !ok || method != "execute_tool" {
		return ToolRequest{}, fmt.Errorf("invalid tool_request: method must be execute_tool")
	}

	paramsMap, ok := msg["params"].(map[string]any)
	if !ok {
		return ToolRequest{}, fmt.Errorf("invalid tool_request: params object required")
	}
	tool, ok := paramsMap["tool"].(string)
	if !ok || tool == "" {
		return ToolRequest{}, fmt.Errorf("invalid tool_request: params.tool is required")
	}

	args := map[string]any{}
	if rawArgs, exists := paramsMap["args"]; exists && rawArgs != nil {
		decodedArgs, ok := rawArgs.(map[string]any)
		if !ok {
			return ToolRequest{}, fmt.Errorf("invalid tool_request: params.args must be object when provided")
		}
		args = decodedArgs
	}

	tabID, err := parsePreferredInt(msg["tabId"], paramsMap["tabId"], args["tabId"])
	if err != nil {
		return ToolRequest{}, fmt.Errorf("tabId must be a number")
	}

	windowID, err := parsePreferredInt(msg["windowId"], paramsMap["windowId"], args["windowId"])
	if err != nil {
		return ToolRequest{}, fmt.Errorf("windowId must be a number")
	}

	return ToolRequest{
		Type:          msgType,
		Method:        method,
		Params:        ToolParams{Tool: tool, Args: args},
		OriginalID:    msg["id"],
		HasOriginalID: msg["id"] != nil,
		TabID:         tabID,
		WindowID:      windowID,
	}, nil
}

func ParseStreamRequest(msg map[string]any) (StreamRequest, error) {
	msgType, ok := msg["type"].(string)
	if !ok || msgType != "stream_request" {
		return StreamRequest{}, fmt.Errorf("invalid stream_request: missing type=stream_request")
	}
	streamType, ok := msg["streamType"].(string)
	if !ok || streamType == "" {
		return StreamRequest{}, fmt.Errorf("invalid stream_request: streamType is required")
	}
	if streamType != "STREAM_CONSOLE" && streamType != "STREAM_NETWORK" {
		return StreamRequest{}, fmt.Errorf("invalid stream_request: unsupported streamType %q", streamType)
	}

	options := map[string]any{}
	if rawOptions, exists := msg["options"]; exists && rawOptions != nil {
		decodedOptions, ok := rawOptions.(map[string]any)
		if !ok {
			return StreamRequest{}, fmt.Errorf("invalid stream_request: options must be object when provided")
		}
		options = decodedOptions
	}

	return StreamRequest{Type: msgType, StreamType: streamType, Options: options}, nil
}

func ParseStreamStop(msg map[string]any) error {
	msgType, ok := msg["type"].(string)
	if !ok || msgType != "stream_stop" {
		return fmt.Errorf("invalid stream_stop: missing type=stream_stop")
	}
	return nil
}

func parsePreferredInt(values ...any) (*int64, error) {
	for _, v := range values {
		if v == nil {
			continue
		}
		n, ok, err := parseInt64(v)
		if err != nil {
			return nil, err
		}
		if ok {
			return &n, nil
		}
	}
	return nil, nil
}

func parseInt64(v any) (int64, bool, error) {
	switch n := v.(type) {
	case int:
		return int64(n), true, nil
	case int32:
		return int64(n), true, nil
	case int64:
		return n, true, nil
	case float64:
		return int64(n), true, nil
	case string:
		if n == "" {
			return 0, false, nil
		}
		x, err := strconv.ParseInt(n, 10, 64)
		if err != nil {
			return 0, false, err
		}
		return x, true, nil
	default:
		return 0, false, fmt.Errorf("unsupported int value")
	}
}
