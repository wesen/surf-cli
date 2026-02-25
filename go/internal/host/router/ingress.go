package router

import (
	"fmt"
)

type ToolRequest struct {
	Type   string
	Method string
	Params ToolParams
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

	return ToolRequest{
		Type:   msgType,
		Method: method,
		Params: ToolParams{Tool: tool, Args: args},
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
