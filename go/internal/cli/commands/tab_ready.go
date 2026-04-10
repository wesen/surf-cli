package commands

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/nicobailon/surf-cli/gohost/internal/cli/transport"
)

const (
	defaultTabReadyTimeout    = 20 * time.Second
	tabReadyRetryInterval     = 400 * time.Millisecond
	tabReadyProbeScript       = `return { href: location.href, title: document.title, readyState: document.readyState }`
	tabExecutionContextErrMsg = "Cannot find default execution context"
)

type tabReadyOptions struct {
	URLExact  string
	URLPrefix string
	Timeout   time.Duration
}

type tabReadyState struct {
	Href       string
	Title      string
	ReadyState string
}

func openOwnedTab(ctx context.Context, client *transport.Client, url string, opts tabReadyOptions) (int64, error) {
	tabResp, err := ExecuteTool(ctx, client, "tab.new", map[string]any{"url": url}, nil, nil)
	if err != nil {
		return 0, err
	}
	tabID, err := extractTabIDFromResponse(tabResp)
	if err != nil {
		return 0, err
	}
	if err := waitForTabReady(ctx, client, tabID, opts); err != nil {
		return 0, err
	}
	return tabID, nil
}

func waitForTabReady(ctx context.Context, client *transport.Client, tabID int64, opts tabReadyOptions) error {
	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = defaultTabReadyTimeout
	}
	deadline := time.Now().Add(timeout)
	var lastState *tabReadyState
	var lastErr error

	for time.Now().Before(deadline) {
		state, err := probeTabReady(ctx, client, tabID)
		if err == nil {
			lastState = state
			if state.ReadyState == "complete" && state.Href != "" && state.Href != "about:blank" && tabURLMatches(state.Href, opts) {
				return nil
			}
		} else if !strings.Contains(err.Error(), tabExecutionContextErrMsg) {
			return err
		} else {
			lastErr = err
		}
		time.Sleep(tabReadyRetryInterval)
	}

	if lastState != nil {
		return fmt.Errorf("tab %d not ready: href=%q title=%q readyState=%q", tabID, lastState.Href, lastState.Title, lastState.ReadyState)
	}
	if lastErr != nil {
		return lastErr
	}
	return fmt.Errorf("tab %d not ready before timeout", tabID)
}

func probeTabReady(ctx context.Context, client *transport.Client, tabID int64) (*tabReadyState, error) {
	resp, err := ExecuteTool(ctx, client, "js", map[string]any{"code": tabReadyProbeScript}, &tabID, nil)
	if err != nil {
		return nil, err
	}
	if e := extractErrorText(resp); e != "" {
		return nil, fmt.Errorf("%s", e)
	}
	parsed := parseResult(resp)
	dataMap, ok := parsed.Data.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("unexpected tab-ready response shape")
	}
	return &tabReadyState{
		Href:       tabReadyString(dataMap["href"]),
		Title:      tabReadyString(dataMap["title"]),
		ReadyState: tabReadyString(dataMap["readyState"]),
	}, nil
}

func tabURLMatches(href string, opts tabReadyOptions) bool {
	if opts.URLExact != "" && href != opts.URLExact {
		return false
	}
	if opts.URLPrefix != "" && !strings.HasPrefix(href, opts.URLPrefix) {
		return false
	}
	return true
}

func tabReadyString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case fmt.Stringer:
		return t.String()
	default:
		if v == nil {
			return ""
		}
		return fmt.Sprint(v)
	}
}
