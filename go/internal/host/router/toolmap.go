package router

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var digitsOnly = regexp.MustCompile(`^\d+$`)

var providerPrefixes = []string{
	"ai",
	"chatgpt",
	"gemini",
	"perplexity",
	"grok",
	"aistudio",
	"aistudio.build",
}

var deferredTools = map[string]struct{}{
	"smoke":           {},
	"batch":           {},
	"health":          {},
	"perf.start":      {},
	"perf.stop":       {},
	"perf.metrics":    {},
	"bookmark.add":    {},
	"bookmark.remove": {},
	"bookmark.list":   {},
	"history.list":    {},
	"history.search":  {},
}

type UnsupportedToolError struct {
	Tool string
}

func (e *UnsupportedToolError) Error() string {
	return fmt.Sprintf("Command '%s' is not supported in go-core profile", e.Tool)
}

func MapToolToMessage(req ToolRequest) (map[string]any, error) {
	tool := strings.TrimSpace(req.Params.Tool)
	if tool == "" {
		return nil, fmt.Errorf("No tool specified")
	}
	if isUnsupportedProvider(tool) || isDeferredTool(tool) {
		return nil, &UnsupportedToolError{Tool: tool}
	}

	a := req.Params.Args
	if a == nil {
		a = map[string]any{}
	}

	base := func(msg map[string]any) map[string]any {
		if req.TabID != nil {
			msg["tabId"] = *req.TabID
		}
		return msg
	}

	switch tool {
	case "computer":
		return mapComputerAction(a, req.TabID), nil
	case "navigate":
		return base(map[string]any{"type": "EXECUTE_NAVIGATE", "url": a["url"]}), nil
	case "read_page":
		options := map[string]any{
			"filter":            stringOr(a["filter"], "interactive"),
			"depth":             a["depth"],
			"refId":             a["ref_id"],
			"format":            a["format"],
			"forceFullSnapshot": boolOr(a["forceFullSnapshot"], false),
			"includeScreenshot": boolOr(a["includeScreenshot"], false),
		}
		return base(map[string]any{"type": "READ_PAGE", "options": options}), nil
	case "get_page_text", "page.text":
		return base(map[string]any{"type": "GET_PAGE_TEXT"}), nil
	case "page_state", "page.state":
		return base(map[string]any{"type": "PAGE_STATE"}), nil
	case "page.read":
		options := map[string]any{
			"filter":      stringOr(a["filter"], "interactive"),
			"refId":       firstNonNil(a["ref"], a["ref_id"]),
			"includeText": !boolOr(a["no-text"], false),
			"compact":     boolOr(a["compact"], false),
		}
		if d, ok := toInt(a["depth"]); ok {
			options["depth"] = d
		}
		return base(map[string]any{"type": "READ_PAGE", "options": options}), nil
	case "form_input":
		return base(map[string]any{"type": "FORM_INPUT", "ref": a["ref"], "value": a["value"]}), nil
	case "eval":
		return base(map[string]any{"type": "EVAL_IN_PAGE", "code": a["code"]}), nil
	case "find_and_type":
		return base(map[string]any{"type": "FIND_AND_TYPE", "text": a["text"], "submit": boolOr(a["submit"], false), "submitKey": stringOr(a["submitKey"], "Enter")}), nil
	case "autocomplete":
		return base(map[string]any{
			"type":       "AUTOCOMPLETE_SELECT",
			"text":       a["text"],
			"ref":        a["ref"],
			"coordinate": a["coordinate"],
			"index":      intOr(a["index"], 0),
			"waitMs":     intOr(a["waitMs"], 500),
		}), nil
	case "set_value":
		return base(map[string]any{"type": "SET_INPUT_VALUE", "selector": a["selector"], "ref": a["ref"], "value": a["value"]}), nil
	case "smart_type":
		return base(map[string]any{"type": "SMART_TYPE", "selector": a["selector"], "text": a["text"], "clear": boolOr(a["clear"], true), "submit": boolOr(a["submit"], false)}), nil
	case "scroll_to_position":
		return base(map[string]any{"type": "SCROLL_TO_POSITION", "position": a["position"], "selector": a["selector"]}), nil
	case "get_scroll_info", "scroll.info":
		return base(map[string]any{"type": "GET_SCROLL_INFO", "selector": a["selector"]}), nil
	case "close_dialogs":
		return base(map[string]any{"type": "CLOSE_DIALOGS", "maxAttempts": intOr(a["maxAttempts"], 3)}), nil
	case "tabs_context":
		return map[string]any{"type": "GET_TABS"}, nil
	case "screenshot":
		return base(map[string]any{
			"type":      "EXECUTE_SCREENSHOT",
			"savePath":  firstNonNil(a["savePath"], a["output"]),
			"annotate":  boolOr(a["annotate"], false),
			"fullpage":  boolOr(a["fullpage"], false),
			"maxHeight": intOr(a["max-height"], 4000),
			"fullRes":   boolOr(a["full"], false),
			"maxSize":   intOr(a["max-size"], 1200),
		}), nil
	case "javascript_tool", "js":
		return base(map[string]any{"type": "EXECUTE_JAVASCRIPT", "code": a["code"]}), nil
	case "wait_for_element", "wait.element":
		timeout := intOr(a["timeout"], 20000)
		if tool == "wait.element" {
			timeout = intOr(a["timeout"], 0)
		}
		msg := map[string]any{"type": "WAIT_FOR_ELEMENT", "selector": a["selector"]}
		if timeout > 0 {
			msg["timeout"] = timeout
		}
		if tool == "wait_for_element" {
			msg["state"] = stringOr(a["state"], "visible")
		}
		return base(msg), nil
	case "wait_for_url", "wait.url":
		timeout := intOr(a["timeout"], 20000)
		if tool == "wait.url" {
			timeout = intOr(a["timeout"], 0)
		}
		msg := map[string]any{"type": "WAIT_FOR_URL", "pattern": firstNonNil(a["pattern"], a["url"], a["urlContains"])}
		if timeout > 0 {
			msg["timeout"] = timeout
		}
		return base(msg), nil
	case "wait_for_network_idle", "wait.network":
		timeout := intOr(a["timeout"], 10000)
		if tool == "wait.network" {
			timeout = intOr(a["timeout"], 0)
		}
		msg := map[string]any{"type": "WAIT_FOR_NETWORK_IDLE"}
		if timeout > 0 {
			msg["timeout"] = timeout
		}
		return base(msg), nil
	case "wait.dom":
		return base(map[string]any{"type": "WAIT_FOR_DOM_STABLE", "stable": intOr(a["stable"], 100), "timeout": intOr(a["timeout"], 5000)}), nil
	case "wait.load":
		return base(map[string]any{"type": "WAIT_FOR_LOAD", "timeout": intOr(a["timeout"], 30000)}), nil
	case "wait":
		return map[string]any{"type": "LOCAL_WAIT", "seconds": intOr(a["duration"], intOr(a["seconds"], 1))}, nil
	case "console", "read_console_messages":
		return base(map[string]any{"type": "READ_CONSOLE_MESSAGES", "onlyErrors": a["only_errors"], "pattern": a["pattern"], "limit": a["limit"], "clear": a["clear"]}), nil
	case "network", "get_network_entries":
		return base(map[string]any{
			"type":        "READ_NETWORK_REQUESTS",
			"full":        boolOr(a["v"], false) || boolOr(a["vv"], false) || stringOr(a["format"], "") == "curl" || stringOr(a["format"], "") == "verbose" || stringOr(a["format"], "") == "raw",
			"urlPattern":  firstNonNil(a["filter"], a["url_pattern"], a["origin"]),
			"method":      a["method"],
			"status":      a["status"],
			"contentType": a["type"],
			"limit":       firstNonNil(a["limit"], a["last"]),
			"format":      a["format"],
			"verbose":     verboseValue(a),
		}), nil
	case "read_network_requests":
		return base(map[string]any{"type": "READ_NETWORK_REQUESTS", "urlPattern": a["url_pattern"], "limit": a["limit"], "clear": a["clear"]}), nil
	case "network.get", "get_network_entry":
		return base(map[string]any{"type": "GET_NETWORK_ENTRY", "requestId": firstNonNil(a["id"], a["0"])}), nil
	case "network.body":
		return base(map[string]any{"type": "GET_RESPONSE_BODY", "requestId": firstNonNil(a["id"], a["0"]), "isRequest": a["request"]}), nil
	case "network.curl":
		return base(map[string]any{"type": "GET_NETWORK_ENTRY", "requestId": firstNonNil(a["id"], a["0"]), "formatAsCurl": true}), nil
	case "network.origins":
		return base(map[string]any{"type": "GET_NETWORK_ORIGINS", "byTab": firstNonNil(a["by-tab"], a["byTab"])}), nil
	case "network.clear":
		return base(map[string]any{"type": "CLEAR_NETWORK_REQUESTS", "before": a["before"], "origin": a["origin"]}), nil
	case "network.stats":
		return base(map[string]any{"type": "GET_NETWORK_STATS"}), nil
	case "network.export":
		return base(map[string]any{"type": "EXPORT_NETWORK_REQUESTS", "har": a["har"], "jsonl": a["jsonl"], "output": a["output"]}), nil
	case "network.path":
		return base(map[string]any{"type": "GET_NETWORK_PATHS", "requestId": firstNonNil(a["id"], a["0"])}), nil
	case "tabs_create":
		return base(map[string]any{"type": "TABS_CREATE", "url": a["url"]}), nil
	case "tabs_register", "tab.name":
		return base(map[string]any{"type": "TABS_REGISTER", "name": a["name"]}), nil
	case "tabs_get_by_name":
		return map[string]any{"type": "TABS_GET_BY_NAME", "name": a["name"]}, nil
	case "tabs_list_named", "tab.named":
		return map[string]any{"type": "TABS_LIST_NAMED"}, nil
	case "tabs_unregister", "tab.unname":
		return map[string]any{"type": "TABS_UNREGISTER", "name": a["name"]}, nil
	case "list_tabs", "tab.list":
		return map[string]any{"type": "LIST_TABS"}, nil
	case "new_tab", "tab.new":
		return map[string]any{"type": "NEW_TAB", "url": a["url"], "urls": a["urls"]}, nil
	case "switch_tab":
		return map[string]any{"type": "SWITCH_TAB", "tabId": firstNonNil(a["tab_id"], a["tabId"])}, nil
	case "close_tab":
		return map[string]any{"type": "CLOSE_TAB", "tabId": firstNonNil(a["tab_id"], a["tabId"]), "tabIds": firstNonNil(a["tab_ids"], a["tabIds"])}, nil
	case "tab.switch":
		id := firstNonNil(a["id"], a["tab_id"], a["tabId"])
		if s, ok := id.(string); ok && s != "" && !digitsOnly.MatchString(s) {
			return map[string]any{"type": "NAMED_TAB_SWITCH", "name": s}, nil
		}
		return map[string]any{"type": "SWITCH_TAB", "tabId": id}, nil
	case "tab.close":
		id := firstNonNil(a["id"], a["tab_id"], a["tabId"])
		ids := firstNonNil(a["ids"], a["tab_ids"], a["tabIds"])
		if s, ok := id.(string); ok && s != "" && !digitsOnly.MatchString(s) {
			return map[string]any{"type": "NAMED_TAB_CLOSE", "name": s}, nil
		}
		return map[string]any{"type": "CLOSE_TAB", "tabId": id, "tabIds": ids}, nil
	case "scroll.top":
		return base(map[string]any{"type": "SCROLL_TO_POSITION", "position": "top", "selector": a["selector"]}), nil
	case "scroll.bottom":
		return base(map[string]any{"type": "SCROLL_TO_POSITION", "position": "bottom", "selector": a["selector"]}), nil
	case "scroll.to":
		return base(map[string]any{"type": "SCROLL_TO_ELEMENT", "ref": a["ref"]}), nil
	case "frame.list":
		return base(map[string]any{"type": "GET_FRAMES"}), nil
	case "frame.switch":
		msg := map[string]any{"type": "FRAME_SWITCH", "selector": a["selector"], "name": a["name"]}
		if idx, ok := toInt(a["index"]); ok {
			msg["index"] = idx
		}
		return base(msg), nil
	case "frame.main":
		return base(map[string]any{"type": "FRAME_MAIN"}), nil
	case "frame.js":
		return base(map[string]any{"type": "EVALUATE_IN_FRAME", "frameId": a["id"], "code": a["code"]}), nil
	case "dialog.accept":
		return base(map[string]any{"type": "DIALOG_ACCEPT", "text": a["text"]}), nil
	case "dialog.dismiss":
		if boolOr(a["all"], false) {
			return base(map[string]any{"type": "CLOSE_DIALOGS", "maxAttempts": intOr(a["maxAttempts"], 3)}), nil
		}
		return base(map[string]any{"type": "DIALOG_DISMISS"}), nil
	case "dialog.info":
		return base(map[string]any{"type": "DIALOG_INFO"}), nil
	case "emulate.network":
		return base(map[string]any{"type": "EMULATE_NETWORK", "preset": a["preset"]}), nil
	case "emulate.cpu":
		rate := floatOr(a["rate"], 1)
		return base(map[string]any{"type": "EMULATE_CPU", "rate": rate}), nil
	case "emulate.geo":
		if boolOr(a["clear"], false) {
			return base(map[string]any{"type": "EMULATE_GEO", "clear": true}), nil
		}
		if a["lat"] == nil || a["lon"] == nil {
			return nil, fmt.Errorf("--lat and --lon required")
		}
		return base(map[string]any{
			"type":      "EMULATE_GEO",
			"latitude":  floatOr(a["lat"], 0),
			"longitude": floatOr(a["lon"], 0),
			"accuracy":  floatOr(a["accuracy"], 100),
		}), nil
	case "emulate.device":
		if boolOr(a["list"], false) {
			return map[string]any{"type": "EMULATE_DEVICE_LIST"}, nil
		}
		if stringOr(a["device"], "") == "" {
			return nil, fmt.Errorf("device name required")
		}
		return base(map[string]any{"type": "EMULATE_DEVICE", "device": a["device"]}), nil
	case "emulate.viewport":
		msg := map[string]any{"type": "EMULATE_VIEWPORT", "mobile": a["mobile"]}
		if w, ok := toInt(a["width"]); ok {
			msg["width"] = w
		}
		if h, ok := toInt(a["height"]); ok {
			msg["height"] = h
		}
		if s, ok := toFloat(a["scale"]); ok {
			msg["deviceScaleFactor"] = s
		}
		return base(msg), nil
	case "emulate.touch":
		return base(map[string]any{"type": "EMULATE_TOUCH", "enabled": !isFalse(a["enabled"])}), nil
	case "form.fill":
		fillData := a["data"]
		if s, ok := fillData.(string); ok && s != "" {
			var parsed any
			if err := json.Unmarshal([]byte(s), &parsed); err != nil {
				return nil, fmt.Errorf("invalid --data JSON")
			}
			fillData = parsed
		}
		return base(map[string]any{"type": "FORM_FILL", "data": fillData}), nil
	case "select":
		if stringOr(a["selector"], "") == "" {
			return nil, fmt.Errorf("selector argument required")
		}
		values := toStringArray(a["values"])
		if len(values) == 0 {
			return nil, fmt.Errorf("at least one value required")
		}
		return base(map[string]any{"type": "SELECT_OPTION", "selector": a["selector"], "values": values, "by": stringOr(a["by"], "value")}), nil
	case "type", "left_click", "right_click", "double_click", "triple_click", "key", "hover", "drag", "scroll":
		a2 := copyMap(a)
		a2["action"] = tool
		return mapComputerAction(a2, req.TabID), nil
	case "click":
		a2 := copyMap(a)
		a2["action"] = "left_click"
		return mapComputerAction(a2, req.TabID), nil
	case "cookie.list":
		return base(map[string]any{"type": "COOKIE_LIST"}), nil
	case "cookie.get":
		if stringOr(a["name"], "") == "" {
			return nil, fmt.Errorf("--name required")
		}
		return base(map[string]any{"type": "COOKIE_GET", "name": a["name"]}), nil
	case "cookie.set":
		if stringOr(a["name"], "") == "" {
			return nil, fmt.Errorf("--name required")
		}
		if _, ok := a["value"]; !ok {
			return nil, fmt.Errorf("--value required")
		}
		return base(map[string]any{"type": "COOKIE_SET", "name": a["name"], "value": a["value"], "expires": a["expires"]}), nil
	case "cookie.clear":
		if boolOr(a["all"], false) {
			return base(map[string]any{"type": "COOKIE_CLEAR_ALL"}), nil
		}
		if stringOr(a["name"], "") == "" {
			return nil, fmt.Errorf("--name or --all required")
		}
		return base(map[string]any{"type": "COOKIE_CLEAR", "name": a["name"]}), nil
	case "search":
		if stringOr(a["term"], "") == "" {
			return nil, fmt.Errorf("search term required")
		}
		return base(map[string]any{"type": "SEARCH_PAGE", "term": a["term"], "caseSensitive": boolOr(a["case-sensitive"], false), "limit": intOr(a["limit"], 10)}), nil
	case "back":
		return base(map[string]any{"type": "EXECUTE_JAVASCRIPT", "code": "history.back()"}), nil
	case "forward":
		return base(map[string]any{"type": "EXECUTE_JAVASCRIPT", "code": "history.forward()"}), nil
	case "tab.reload":
		return base(map[string]any{"type": "TAB_RELOAD", "hard": boolOr(a["hard"], false)}), nil
	case "zoom":
		if boolOr(a["reset"], false) {
			return base(map[string]any{"type": "ZOOM_RESET"}), nil
		}
		if lvl, ok := toFloat(a["level"]); ok {
			return base(map[string]any{"type": "ZOOM_SET", "level": lvl}), nil
		}
		return base(map[string]any{"type": "ZOOM_GET"}), nil
	case "resize", "resize_window":
		return base(map[string]any{"type": "RESIZE_WINDOW", "width": a["width"], "height": a["height"]}), nil
	case "window.new":
		msg := map[string]any{"type": "WINDOW_NEW", "url": a["url"], "incognito": boolOr(a["incognito"], false), "focused": !boolOr(a["unfocused"], false)}
		if w, ok := toInt(a["width"]); ok {
			msg["width"] = w
		}
		if h, ok := toInt(a["height"]); ok {
			msg["height"] = h
		}
		return msg, nil
	case "window.list":
		return map[string]any{"type": "WINDOW_LIST", "includeTabs": boolOr(a["tabs"], false)}, nil
	case "window.focus":
		if a["id"] == nil {
			return nil, fmt.Errorf("window id required")
		}
		id, ok := toInt(a["id"])
		if !ok {
			return nil, fmt.Errorf("window id required")
		}
		return map[string]any{"type": "WINDOW_FOCUS", "windowId": id}, nil
	case "window.close":
		if a["id"] == nil {
			return nil, fmt.Errorf("window id required")
		}
		id, ok := toInt(a["id"])
		if !ok {
			return nil, fmt.Errorf("window id required")
		}
		return map[string]any{"type": "WINDOW_CLOSE", "windowId": id}, nil
	case "window.resize":
		if a["id"] == nil {
			return nil, fmt.Errorf("--id required")
		}
		id, ok := toInt(a["id"])
		if !ok {
			return nil, fmt.Errorf("--id required")
		}
		msg := map[string]any{"type": "WINDOW_RESIZE", "windowId": id, "state": a["state"]}
		if w, ok := toInt(a["width"]); ok {
			msg["width"] = w
		}
		if h, ok := toInt(a["height"]); ok {
			msg["height"] = h
		}
		if left, ok := toInt(a["left"]); ok {
			msg["left"] = left
		}
		if top, ok := toInt(a["top"]); ok {
			msg["top"] = top
		}
		return msg, nil
	default:
		return nil, fmt.Errorf("Unknown tool: %s", tool)
	}
}

func mapComputerAction(args map[string]any, tabID *int64) map[string]any {
	a := args
	action := stringOr(a["action"], "")
	if action == "" {
		return map[string]any{"type": "UNSUPPORTED_ACTION", "action": nil, "message": "No action specified for computer tool"}
	}
	coordX, coordY := coordinateXY(a)
	base := func(msg map[string]any) map[string]any {
		if tabID != nil {
			msg["tabId"] = *tabID
		}
		return msg
	}

	switch action {
	case "screenshot":
		return base(map[string]any{"type": "EXECUTE_SCREENSHOT"})
	case "left_click":
		if a["ref"] != nil {
			return base(map[string]any{"type": "CLICK_REF", "ref": a["ref"], "button": "left"})
		}
		if a["selector"] != nil {
			return base(map[string]any{"type": "CLICK_SELECTOR", "selector": a["selector"], "index": intOr(a["index"], 0), "button": "left"})
		}
		return base(map[string]any{"type": "EXECUTE_CLICK", "x": coordX, "y": coordY, "modifiers": a["modifiers"]})
	case "right_click":
		if a["ref"] != nil {
			return base(map[string]any{"type": "CLICK_REF", "ref": a["ref"], "button": "right"})
		}
		return base(map[string]any{"type": "EXECUTE_RIGHT_CLICK", "x": coordX, "y": coordY, "modifiers": a["modifiers"]})
	case "double_click":
		if a["ref"] != nil {
			return base(map[string]any{"type": "CLICK_REF", "ref": a["ref"], "button": "double"})
		}
		return base(map[string]any{"type": "EXECUTE_DOUBLE_CLICK", "x": coordX, "y": coordY, "modifiers": a["modifiers"]})
	case "triple_click":
		if a["ref"] != nil {
			return base(map[string]any{"type": "CLICK_REF", "ref": a["ref"], "button": "triple"})
		}
		return base(map[string]any{"type": "EXECUTE_TRIPLE_CLICK", "x": coordX, "y": coordY, "modifiers": a["modifiers"]})
	case "type":
		if a["ref"] != nil {
			return base(map[string]any{"type": "FORM_FILL", "data": []map[string]any{{"ref": a["ref"], "value": a["text"]}}})
		}
		return base(map[string]any{"type": "EXECUTE_TYPE", "text": a["text"]})
	case "key":
		keyValue := firstNonNil(a["key"], a["text"])
		repeat := intOr(a["repeat"], 1)
		if repeat < 1 {
			repeat = 1
		}
		if repeat > 100 {
			repeat = 100
		}
		if repeat > 1 {
			msg := map[string]any{"type": "EXECUTE_KEY_REPEAT", "key": keyValue, "repeat": repeat}
			if tabID != nil {
				msg["tabId"] = *tabID
			}
			return msg
		}
		return base(map[string]any{"type": "EXECUTE_KEY", "key": keyValue})
	case "type_submit":
		return base(map[string]any{"type": "TYPE_SUBMIT", "text": a["text"], "submitKey": stringOr(a["submitKey"], "Enter")})
	case "click_type":
		return base(map[string]any{"type": "CLICK_TYPE", "text": a["text"], "ref": a["ref"], "coordinate": a["coordinate"]})
	case "click_type_submit":
		return base(map[string]any{"type": "CLICK_TYPE_SUBMIT", "text": a["text"], "ref": a["ref"], "coordinate": a["coordinate"], "submitKey": stringOr(a["submitKey"], "Enter")})
	case "find_and_type":
		return base(map[string]any{"type": "FIND_AND_TYPE", "text": a["text"], "submit": boolOr(a["submit"], false), "submitKey": stringOr(a["submitKey"], "Enter")})
	case "scroll":
		amount := intOr(a["scroll_amount"], 3) * 100
		dir := stringOr(a["scroll_direction"], "")
		deltaX := int64(0)
		deltaY := int64(0)
		switch dir {
		case "up":
			deltaY = -amount
		case "down":
			deltaY = amount
		case "left":
			deltaX = -amount
		case "right":
			deltaX = amount
		}
		return base(map[string]any{"type": "EXECUTE_SCROLL", "deltaX": deltaX, "deltaY": deltaY, "x": coordX, "y": coordY})
	case "scroll_to":
		return base(map[string]any{"type": "SCROLL_TO_ELEMENT", "ref": a["ref"]})
	case "hover":
		if a["ref"] != nil {
			return base(map[string]any{"type": "HOVER_REF", "ref": a["ref"]})
		}
		return base(map[string]any{"type": "EXECUTE_HOVER", "x": coordX, "y": coordY})
	case "left_click_drag", "drag":
		startX, startY := pairXY(a["start_coordinate"])
		return base(map[string]any{"type": "EXECUTE_DRAG", "startX": startX, "startY": startY, "endX": coordX, "endY": coordY, "modifiers": a["modifiers"]})
	case "wait":
		return map[string]any{"type": "LOCAL_WAIT", "seconds": intOr(a["duration"], 1)}
	case "zoom":
		if boolOr(a["reset"], false) {
			msg := map[string]any{"type": "ZOOM_RESET"}
			if tabID != nil {
				msg["tabId"] = *tabID
			}
			return msg
		}
		if lvl, ok := toFloat(a["level"]); ok {
			msg := map[string]any{"type": "ZOOM_SET", "level": lvl}
			if tabID != nil {
				msg["tabId"] = *tabID
			}
			return msg
		}
		msg := map[string]any{"type": "ZOOM_GET"}
		if tabID != nil {
			msg["tabId"] = *tabID
		}
		return msg
	default:
		return map[string]any{"type": "UNSUPPORTED_ACTION", "action": action, "message": fmt.Sprintf("Unknown computer action: %s", action)}
	}
}

func isUnsupportedProvider(tool string) bool {
	for _, prefix := range providerPrefixes {
		if tool == prefix || strings.HasPrefix(tool, prefix+".") {
			return true
		}
	}
	return false
}

func isDeferredTool(tool string) bool {
	_, ok := deferredTools[tool]
	return ok
}

func coordinateXY(args map[string]any) (any, any) {
	if x, y := pairXY(args["coordinate"]); x != nil || y != nil {
		return x, y
	}
	if args["x"] != nil || args["y"] != nil {
		return args["x"], args["y"]
	}
	return nil, nil
}

func pairXY(v any) (any, any) {
	arr, ok := v.([]any)
	if !ok || len(arr) < 2 {
		return nil, nil
	}
	return arr[0], arr[1]
}

func verboseValue(a map[string]any) int64 {
	if boolOr(a["v"], false) {
		return 1
	}
	if boolOr(a["vv"], false) {
		return 2
	}
	return 0
}

func copyMap(m map[string]any) map[string]any {
	out := make(map[string]any, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

func firstNonNil(values ...any) any {
	for _, v := range values {
		if v != nil {
			return v
		}
	}
	return nil
}

func stringOr(v any, d string) string {
	s, ok := v.(string)
	if !ok || s == "" {
		return d
	}
	return s
}

func boolOr(v any, d bool) bool {
	if v == nil {
		return d
	}
	switch b := v.(type) {
	case bool:
		return b
	case string:
		parsed, err := strconv.ParseBool(b)
		if err != nil {
			return d
		}
		return parsed
	default:
		return d
	}
}

func isFalse(v any) bool {
	if v == nil {
		return false
	}
	switch b := v.(type) {
	case bool:
		return !b
	case string:
		parsed, err := strconv.ParseBool(b)
		return err == nil && !parsed
	default:
		return false
	}
}

func intOr(v any, d int64) int64 {
	if n, ok := toInt(v); ok {
		return n
	}
	return d
}

func toInt(v any) (int64, bool) {
	switch n := v.(type) {
	case int:
		return int64(n), true
	case int32:
		return int64(n), true
	case int64:
		return n, true
	case float64:
		return int64(n), true
	case string:
		if n == "" {
			return 0, false
		}
		x, err := strconv.ParseInt(n, 10, 64)
		if err != nil {
			return 0, false
		}
		return x, true
	default:
		return 0, false
	}
}

func floatOr(v any, d float64) float64 {
	if n, ok := toFloat(v); ok {
		return n
	}
	return d
}

func toFloat(v any) (float64, bool) {
	switch n := v.(type) {
	case int:
		return float64(n), true
	case int32:
		return float64(n), true
	case int64:
		return float64(n), true
	case float64:
		return n, true
	case string:
		if n == "" {
			return 0, false
		}
		x, err := strconv.ParseFloat(n, 64)
		if err != nil {
			return 0, false
		}
		return x, true
	default:
		return 0, false
	}
}

func toStringArray(v any) []string {
	switch raw := v.(type) {
	case nil:
		return nil
	case string:
		if raw == "" {
			return nil
		}
		return []string{raw}
	case []string:
		return raw
	case []any:
		out := make([]string, 0, len(raw))
		for _, item := range raw {
			if s, ok := item.(string); ok && s != "" {
				out = append(out, s)
			}
		}
		return out
	default:
		return nil
	}
}
