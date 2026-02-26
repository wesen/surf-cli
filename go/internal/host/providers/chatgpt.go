package providers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	defaultChatGPTTimeout = 45 * time.Minute
	extCallTimeout        = 30 * time.Second
)

const (
	chatGPTPromptSelectors = `#prompt-textarea, [data-testid="composer-textarea"], textarea[name="prompt-textarea"], .ProseMirror, [contenteditable="true"][data-virtualkeyboard="true"]`
	chatGPTSendSelectors   = `button[data-testid="send-button"], button[data-testid*="composer-send"], form button[type="submit"]`
	chatGPTURL             = "https://chatgpt.com/"
)

// NativeCaller sends one request to the extension native-messaging side and waits for a response.
type NativeCaller interface {
	Request(ctx context.Context, msg map[string]any, timeout time.Duration) (map[string]any, error)
}

type ChatGPTRequest struct {
	Query      string
	Model      string
	ListModels bool
	WithPage   bool
	File       string
	Timeout    time.Duration
	TabID      *int64
}

func HandleChatGPTTool(ctx context.Context, caller NativeCaller, rawArgs map[string]any, tabID *int64, logf func(string, ...any)) (map[string]any, error) {
	req, err := parseChatGPTRequest(rawArgs, tabID)
	if err != nil {
		return nil, err
	}
	return runChatGPTQuery(ctx, caller, req, logf)
}

func parseChatGPTRequest(rawArgs map[string]any, tabID *int64) (ChatGPTRequest, error) {
	args := rawArgs
	if args == nil {
		args = map[string]any{}
	}

	listModels := asBool(args["list-models"]) || asBool(args["listModels"])
	query := strings.TrimSpace(asString(args["query"]))
	if query == "" && !listModels {
		return ChatGPTRequest{}, fmt.Errorf("query required")
	}

	timeout := defaultChatGPTTimeout
	if v, ok := args["timeout"]; ok && v != nil {
		seconds, ok := toInt64(v)
		if !ok || seconds < 1 {
			return ChatGPTRequest{}, fmt.Errorf("timeout must be a positive integer")
		}
		timeout = time.Duration(seconds) * time.Second
	}

	return ChatGPTRequest{
		Query:      query,
		Model:      strings.TrimSpace(asString(args["model"])),
		ListModels: listModels,
		WithPage:   asBool(args["with-page"]) || asBool(args["withPage"]),
		File:       strings.TrimSpace(asString(args["file"])),
		Timeout:    timeout,
		TabID:      tabID,
	}, nil
}

func runChatGPTQuery(ctx context.Context, caller NativeCaller, req ChatGPTRequest, logf func(string, ...any)) (map[string]any, error) {
	if logf == nil {
		logf = func(string, ...any) {}
	}

	started := time.Now()
	logf("[chatgpt] starting query")

	cookiesResp, err := caller.Request(ctx, map[string]any{"type": "GET_CHATGPT_COOKIES"}, extCallTimeout)
	if err != nil {
		return nil, err
	}
	if e := responseError(cookiesResp); e != "" {
		return nil, errors.New(e)
	}
	if !hasChatGPTSessionCookie(cookiesResp["cookies"]) {
		return nil, fmt.Errorf("ChatGPT login required")
	}

	fullPrompt := req.Query
	if !req.ListModels && req.WithPage {
		pageMsg := map[string]any{"type": "GET_PAGE_TEXT"}
		if req.TabID != nil {
			pageMsg["tabId"] = *req.TabID
		}
		pageResp, pageErr := caller.Request(ctx, pageMsg, 45*time.Second)
		if pageErr == nil && responseError(pageResp) == "" {
			url := asString(pageResp["url"])
			text := asString(pageResp["text"])
			if text == "" {
				text = asString(pageResp["pageContent"])
			}
			if url != "" || text != "" {
				fullPrompt = fmt.Sprintf("Page: %s\n\n%s\n\n---\n\n%s", url, text, req.Query)
			}
		}
	}

	tabResp, err := caller.Request(ctx, map[string]any{"type": "CHATGPT_NEW_TAB"}, 45*time.Second)
	if err != nil {
		return nil, err
	}
	if e := responseError(tabResp); e != "" {
		return nil, errors.New(e)
	}
	tabID, ok := toInt64(tabResp["tabId"])
	if !ok || tabID <= 0 {
		return nil, fmt.Errorf("Failed to create ChatGPT tab")
	}
	logf("[chatgpt] opened tab %d", tabID)

	bridge := &chatGPTBridge{caller: caller, tabID: tabID}
	defer func() {
		_, _ = caller.Request(context.Background(), map[string]any{
			"type":  "CHATGPT_CLOSE_TAB",
			"tabId": tabID,
		}, extCallTimeout)
	}()

	if err := bridge.waitForPageLoad(ctx, 45*time.Second); err != nil {
		return nil, err
	}
	blocked, err := bridge.isCloudflareBlocked(ctx)
	if err != nil {
		return nil, err
	}
	if blocked {
		return nil, fmt.Errorf("Cloudflare challenge detected - complete in browser")
	}

	loginStatus, err := bridge.checkLoginStatus(ctx)
	if err != nil {
		return nil, err
	}
	if loginStatus.Status != 200 || loginStatus.HasLoginCTA {
		return nil, fmt.Errorf("ChatGPT login required")
	}

	promptReady, err := bridge.waitForPromptReady(ctx, 30*time.Second)
	if err != nil {
		return nil, err
	}
	if !promptReady {
		return nil, fmt.Errorf("Prompt textarea not ready")
	}

	if req.ListModels {
		models, selected, err := bridge.listModels(ctx, 12*time.Second)
		if err != nil {
			return nil, err
		}
		return map[string]any{
			"models":   models,
			"selected": selected,
			"tookMs":   time.Since(started).Milliseconds(),
		}, nil
	}

	if req.Model != "" {
		if err := bridge.selectModel(ctx, req.Model, 8*time.Second); err != nil {
			return nil, err
		}
	}

	if req.File != "" {
		if err := bridge.uploadFiles(ctx, req.File); err != nil {
			return nil, err
		}
	}

	if err := bridge.typePrompt(ctx, fullPrompt); err != nil {
		return nil, err
	}
	if err := bridge.clickSend(ctx); err != nil {
		return nil, err
	}
	responseText, err := bridge.waitForResponse(ctx, req.Timeout)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"response": responseText,
		"model":    firstNonEmpty(req.Model, "current"),
		"tookMs":   time.Since(started).Milliseconds(),
	}, nil
}

type chatGPTBridge struct {
	caller NativeCaller
	tabID  int64
}

type chatGPTLoginStatus struct {
	Status      int64
	HasLoginCTA bool
}

func (b *chatGPTBridge) evaluate(ctx context.Context, expression string, timeout time.Duration) (any, error) {
	resp, err := b.caller.Request(ctx, map[string]any{
		"type":       "CHATGPT_EVALUATE",
		"tabId":      b.tabID,
		"expression": expression,
	}, timeout)
	if err != nil {
		return nil, err
	}
	if e := responseError(resp); e != "" {
		return nil, errors.New(e)
	}
	if exceptionDetails, ok := resp["exceptionDetails"]; ok && exceptionDetails != nil {
		return nil, errors.New(extractExceptionText(exceptionDetails))
	}
	result, _ := resp["result"].(map[string]any)
	if result == nil {
		if v, ok := resp["value"]; ok {
			return v, nil
		}
		return nil, nil
	}
	return result["value"], nil
}

func (b *chatGPTBridge) cdpCommand(ctx context.Context, method string, params map[string]any, timeout time.Duration) error {
	resp, err := b.caller.Request(ctx, map[string]any{
		"type":   "CHATGPT_CDP_COMMAND",
		"tabId":  b.tabID,
		"method": method,
		"params": params,
	}, timeout)
	if err != nil {
		return err
	}
	if e := responseError(resp); e != "" {
		return errors.New(e)
	}
	return nil
}

func (b *chatGPTBridge) waitForPageLoad(ctx context.Context, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		v, err := b.evaluate(ctx, "document.readyState", extCallTimeout)
		if err != nil {
			return err
		}
		if ready, _ := v.(string); ready == "complete" || ready == "interactive" {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("Page did not load in time")
}

func (b *chatGPTBridge) isCloudflareBlocked(ctx context.Context) (bool, error) {
	titleValue, err := b.evaluate(ctx, "document.title.toLowerCase()", extCallTimeout)
	if err != nil {
		return false, err
	}
	if title, _ := titleValue.(string); strings.Contains(title, "just a moment") {
		return true, nil
	}
	v, err := b.evaluate(ctx, `Boolean(document.querySelector('script[src*="/challenge-platform/"]'))`, extCallTimeout)
	if err != nil {
		return false, err
	}
	return asBool(v), nil
}

func (b *chatGPTBridge) checkLoginStatus(ctx context.Context) (chatGPTLoginStatus, error) {
	v, err := b.evaluate(ctx, `(async () => {
	  try {
	    const response = await fetch('/backend-api/me', { cache: 'no-store', credentials: 'include' });
	    const hasLoginCta = Array.from(document.querySelectorAll('a[href*="/auth/login"], button'))
	      .some(el => {
	        const text = (el.textContent || '').toLowerCase().trim();
	        return text.startsWith('log in') || text.startsWith('sign in');
	      });
	    return { status: response.status, hasLoginCta };
	  } catch (e) {
	    return { status: 0, hasLoginCta: false };
	  }
	})()`, extCallTimeout)
	if err != nil {
		return chatGPTLoginStatus{}, err
	}
	m, _ := v.(map[string]any)
	status, _ := toInt64(m["status"])
	return chatGPTLoginStatus{Status: status, HasLoginCTA: asBool(m["hasLoginCta"])}, nil
}

func (b *chatGPTBridge) waitForPromptReady(ctx context.Context, timeout time.Duration) (bool, error) {
	deadline := time.Now().Add(timeout)
	sel := toJSONString(strings.Split(chatGPTPromptSelectors, ", "))
	expr := fmt.Sprintf(`(() => {
	  const selectors = %s;
	  for (const selector of selectors) {
	    const node = document.querySelector(selector);
	    if (node && !node.hasAttribute('disabled')) return true;
	  }
	  return false;
	})()`, sel)
	for time.Now().Before(deadline) {
		v, err := b.evaluate(ctx, expr, extCallTimeout)
		if err != nil {
			return false, err
		}
		if asBool(v) {
			return true, nil
		}
		time.Sleep(200 * time.Millisecond)
	}
	return false, nil
}

func (b *chatGPTBridge) selectModel(ctx context.Context, model string, timeout time.Duration) error {
	normalized := normalizeModel(model)
	target := toJSONString(normalized)

	if err := b.openModelMenu(ctx, timeout); err != nil {
		return err
	}

	deadline := time.Now().Add(timeout)
	lastAvailable := []string{}
	expr := fmt.Sprintf(`(() => {
	  function dispatchClickSequence(target){
	    if(!target || !(target instanceof EventTarget)) return false;
	    const types = ['pointerdown', 'mousedown', 'pointerup', 'mouseup', 'click'];
	    for (const type of types) {
	      const common = { bubbles: true, cancelable: true, view: window };
	      let event;
	      if (type.startsWith('pointer') && 'PointerEvent' in window) {
	        event = new PointerEvent(type, { ...common, pointerId: 1, pointerType: 'mouse' });
	      } else {
	        event = new MouseEvent(type, common);
	      }
	      target.dispatchEvent(event);
	    }
	    return true;
	  }
	  const normalizeDisplay = (text) => {
	    const cleaned = (text || '').replace(/\s+/g, ' ').trim();
	    if (!cleaned) return '';
	    const spaced = cleaned.replace(/([a-z])([A-Z])/g, '$1 $2').replace(/\s+/g, ' ').trim();
	    for (const known of ['Auto', 'Instant', 'Thinking', 'Pro']) {
	      if (spaced === known || spaced.startsWith(known + ' ')) return known.toLowerCase();
	    }
	    return spaced;
	  };
	  const extractModelId = (blob) => {
	    const lower = (blob || '').toLowerCase();
	    const gptMatch = lower.match(/\b(gpt[-a-z0-9._]+)\b/);
	    if (gptMatch) return gptMatch[1];
	    const reasoningMatch = lower.match(/\b(o[0-9][-a-z0-9._]*)\b/);
	    if (reasoningMatch) return reasoningMatch[1];
	    return '';
	  };
	  const normalize = (text) => (text || '').toLowerCase().replace(/[^a-z0-9]/g, '');
	  const targetModel = %s;
	  const menu = document.querySelector('[role="menu"], [data-radix-collection-root]');
	  if (!menu) return { found: false, waiting: true };
	  const items = Array.from(menu.querySelectorAll('button, [role="menuitem"], [role="menuitemradio"], [data-testid*="model-switcher-"]'));
	  const legacyToggle = items.find((item) => {
	    const labelNorm = normalize(item.getAttribute('aria-label') || item.textContent || '');
	    const testIdNorm = normalize(item.getAttribute('data-testid') || '');
	    return labelNorm.startsWith('legacymodels') || testIdNorm.includes('legacymodel');
	  });
	  if (legacyToggle) {
	    const expanded = legacyToggle.getAttribute('aria-expanded') === 'true' ||
	                     legacyToggle.getAttribute('data-state') === 'open';
	    const attempts = Number(menu.getAttribute('data-surf-legacy-open-attempts') || '0');
	    if (!expanded && attempts < 2) {
	      menu.setAttribute('data-surf-legacy-open-attempts', String(attempts + 1));
	      dispatchClickSequence(legacyToggle);
	      return { found: false, waiting: true };
	    }
	  }
	  const available = [];
	  let bestMatch = null;
	  let bestScore = 0;
	  for (const item of items) {
	    const label = normalizeDisplay(item.getAttribute('aria-label') || item.textContent || '');
	    const text = normalize(item.textContent || '');
	    const testId = normalize(item.getAttribute('data-testid') || '');
	    const canonical = extractModelId([item.getAttribute('data-testid'), item.getAttribute('aria-label'), item.textContent].filter(Boolean).join(' ')) || label;
	    if (canonical && canonical.toLowerCase() !== 'legacy models' && !available.includes(canonical)) available.push(canonical);
	    let score = 0;
	    const canonicalNorm = normalize(canonical);
	    if (canonicalNorm === targetModel || text === targetModel || testId === targetModel) score = 140;
	    else if (canonicalNorm.includes(targetModel) || text.includes(targetModel) || testId.includes(targetModel)) score = 100;
	    else if (targetModel.includes(canonicalNorm) || targetModel.includes(text) || targetModel.includes(testId)) score = 50;
	    if (score > bestScore) {
	      bestScore = score;
	      bestMatch = item;
	    }
	  }
	  if (bestMatch) {
	    dispatchClickSequence(bestMatch);
	    return { found: true, success: true, available, label: (bestMatch.textContent || '').trim() };
	  }
	  return { found: true, success: false, available };
	})()`, target)

	for time.Now().Before(deadline) {
		v, err := b.evaluate(ctx, expr, extCallTimeout)
		if err != nil {
			return err
		}
		m, _ := v.(map[string]any)
		if asBool(m["found"]) {
			lastAvailable = toStringSlice(m["available"])
			if asBool(m["success"]) {
				return nil
			}
			if len(lastAvailable) > 0 {
				return fmt.Errorf("Model not found: %s. Available: %s", model, strings.Join(lastAvailable, ", "))
			}
			return fmt.Errorf("Model not found: %s", model)
		}
		time.Sleep(100 * time.Millisecond)
	}

	if len(lastAvailable) > 0 {
		return fmt.Errorf("Model not found: %s (timeout). Available: %s", model, strings.Join(lastAvailable, ", "))
	}
	return fmt.Errorf("Model not found: %s (timeout)", model)
}

func (b *chatGPTBridge) openModelMenu(ctx context.Context, timeout time.Duration) error {
	buttonExists, err := b.evaluate(ctx, `(() => Boolean(document.querySelector('[data-testid="model-switcher-dropdown-button"]')))()`, extCallTimeout)
	if err != nil {
		return err
	}
	if !asBool(buttonExists) {
		return fmt.Errorf("Model selector button not found")
	}

	openExpr := `(() => {
	  function dispatchClickSequence(target){
	    if(!target || !(target instanceof EventTarget)) return false;
	    const types = ['pointerdown', 'mousedown', 'pointerup', 'mouseup', 'click'];
	    for (const type of types) {
	      const common = { bubbles: true, cancelable: true, view: window };
	      let event;
	      if (type.startsWith('pointer') && 'PointerEvent' in window) {
	        event = new PointerEvent(type, { ...common, pointerId: 1, pointerType: 'mouse' });
	      } else {
	        event = new MouseEvent(type, common);
	      }
	      target.dispatchEvent(event);
	    }
	    return true;
	  }
	  const btn = document.querySelector('[data-testid="model-switcher-dropdown-button"]');
	  if (!btn) return false;
	  dispatchClickSequence(btn);
	  return true;
	})()`
	if _, err := b.evaluate(ctx, openExpr, extCallTimeout); err != nil {
		return err
	}

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		menuVisible, err := b.evaluate(ctx, `(() => Boolean(document.querySelector('[role="menu"], [data-radix-collection-root]')))()`, extCallTimeout)
		if err != nil {
			return err
		}
		if asBool(menuVisible) {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("Model selector menu did not open")
}

func (b *chatGPTBridge) listModels(ctx context.Context, timeout time.Duration) ([]string, string, error) {
	if err := b.openModelMenu(ctx, timeout); err != nil {
		return nil, "", err
	}
	deadline := time.Now().Add(timeout)
	expr := `(() => {
	  function dispatchClickSequence(target){
	    if(!target || !(target instanceof EventTarget)) return false;
	    const types = ['pointerdown', 'mousedown', 'pointerup', 'mouseup', 'click'];
	    for (const type of types) {
	      const common = { bubbles: true, cancelable: true, view: window };
	      let event;
	      if (type.startsWith('pointer') && 'PointerEvent' in window) {
	        event = new PointerEvent(type, { ...common, pointerId: 1, pointerType: 'mouse' });
	      } else {
	        event = new MouseEvent(type, common);
	      }
	      target.dispatchEvent(event);
	    }
	    return true;
	  }
	  const normalizeDisplay = (text) => {
	    const cleaned = (text || '').replace(/\s+/g, ' ').trim();
	    if (!cleaned) return '';
	    const spaced = cleaned.replace(/([a-z])([A-Z])/g, '$1 $2').replace(/\s+/g, ' ').trim();
	    for (const known of ['Auto', 'Instant', 'Thinking', 'Pro']) {
	      if (spaced === known || spaced.startsWith(known + ' ')) return known.toLowerCase();
	    }
	    return spaced;
	  };
	  const extractModelId = (blob) => {
	    const lower = (blob || '').toLowerCase();
	    const gptMatch = lower.match(/\b(gpt[-a-z0-9._]+)\b/);
	    if (gptMatch) return gptMatch[1];
	    const reasoningMatch = lower.match(/\b(o[0-9][-a-z0-9._]*)\b/);
	    if (reasoningMatch) return reasoningMatch[1];
	    return '';
	  };
	  const menu = document.querySelector('[role="menu"], [data-radix-collection-root]');
	  if (!menu) return { found: false };
	  const items = Array.from(menu.querySelectorAll('button, [role="menuitem"], [role="menuitemradio"], [data-testid*="model-switcher-"]'));
	  const legacyToggle = items.find((item) => {
	    const labelNorm = (item.getAttribute('aria-label') || item.textContent || '').toLowerCase().replace(/[^a-z0-9]/g, '');
	    const testIdNorm = (item.getAttribute('data-testid') || '').toLowerCase().replace(/[^a-z0-9]/g, '');
	    return labelNorm.startsWith('legacymodels') || testIdNorm.includes('legacymodel');
	  });
	  if (legacyToggle) {
	    const expanded = legacyToggle.getAttribute('aria-expanded') === 'true' ||
	                     legacyToggle.getAttribute('data-state') === 'open';
	    const attempts = Number(menu.getAttribute('data-surf-legacy-open-attempts') || '0');
	    if (!expanded && attempts < 2) {
	      menu.setAttribute('data-surf-legacy-open-attempts', String(attempts + 1));
	      dispatchClickSequence(legacyToggle);
	      return { found: false };
	    }
	  }
	  const models = [];
	  let selected = null;
	  for (const item of items) {
	    const label = normalizeDisplay(item.getAttribute('aria-label') || item.textContent || '');
	    const canonical = extractModelId([item.getAttribute('data-testid'), item.getAttribute('aria-label'), item.textContent].filter(Boolean).join(' ')) || label;
	    if (!canonical) continue;
	    if (canonical.toLowerCase() === 'legacy models') continue;
	    if (!models.includes(canonical)) models.push(canonical);
	    const ariaChecked = item.getAttribute('aria-checked');
	    const dataState = item.getAttribute('data-state');
	    if (ariaChecked === 'true' || dataState === 'checked') {
	      selected = canonical;
	    }
	  }
	  return { found: true, models, selected };
	})()`
	for time.Now().Before(deadline) {
		v, err := b.evaluate(ctx, expr, extCallTimeout)
		if err != nil {
			return nil, "", err
		}
		m, _ := v.(map[string]any)
		if asBool(m["found"]) {
			models := toStringSlice(m["models"])
			return models, strings.TrimSpace(asString(m["selected"])), nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return nil, "", fmt.Errorf("Failed to read ChatGPT models")
}

func (b *chatGPTBridge) typePrompt(ctx context.Context, prompt string) error {
	sel := toJSONString(strings.Split(chatGPTPromptSelectors, ", "))
	focusExpr := fmt.Sprintf(`(() => {
	  const selectors = %s;
	  for (const selector of selectors) {
	    const node = document.querySelector(selector);
	    if (!node) continue;
	    if (typeof node.focus === 'function') node.focus();
	    node.click?.();
	    return true;
	  }
	  return false;
	})()`, sel)
	focused, err := b.evaluate(ctx, focusExpr, extCallTimeout)
	if err != nil {
		return err
	}
	if !asBool(focused) {
		return fmt.Errorf("Failed to focus prompt textarea")
	}
	if err := b.cdpCommand(ctx, "Input.insertText", map[string]any{"text": prompt}, extCallTimeout); err != nil {
		return err
	}
	time.Sleep(300 * time.Millisecond)

	verifyExpr := fmt.Sprintf(`(() => {
	  const selectors = %s;
	  for (const selector of selectors) {
	    const node = document.querySelector(selector);
	    if (!node) continue;
	    const text = node.innerText || node.value || node.textContent || '';
	    if (text.trim().length > 0) return true;
	  }
	  return false;
	})()`, sel)
	verified, err := b.evaluate(ctx, verifyExpr, extCallTimeout)
	if err != nil {
		return err
	}
	if !asBool(verified) {
		return fmt.Errorf("Failed to type prompt")
	}
	return nil
}

func (b *chatGPTBridge) clickSend(ctx context.Context) error {
	selectors := toJSONString(strings.Split(chatGPTSendSelectors, ", "))
	expr := fmt.Sprintf(`(() => {
	  const selectors = %s;
	  let button = null;
	  for (const selector of selectors) {
	    button = document.querySelector(selector);
	    if (button) break;
	  }
	  if (!button) return 'missing';
	  const disabled = button.hasAttribute('disabled') || button.getAttribute('aria-disabled') === 'true' || button.getAttribute('data-disabled') === 'true';
	  if (disabled) return 'disabled';
	  button.click();
	  return 'clicked';
	})()`, selectors)

	deadline := time.Now().Add(8 * time.Second)
	for time.Now().Before(deadline) {
		v, err := b.evaluate(ctx, expr, extCallTimeout)
		if err != nil {
			return err
		}
		status, _ := v.(string)
		if status == "clicked" {
			return nil
		}
		if status == "missing" {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	if err := b.cdpCommand(ctx, "Input.dispatchKeyEvent", map[string]any{
		"type":                  "keyDown",
		"key":                   "Enter",
		"code":                  "Enter",
		"windowsVirtualKeyCode": 13,
		"nativeVirtualKeyCode":  13,
		"text":                  "\r",
	}, extCallTimeout); err != nil {
		return err
	}
	return b.cdpCommand(ctx, "Input.dispatchKeyEvent", map[string]any{
		"type":                  "keyUp",
		"key":                   "Enter",
		"code":                  "Enter",
		"windowsVirtualKeyCode": 13,
		"nativeVirtualKeyCode":  13,
	}, extCallTimeout)
}

func (b *chatGPTBridge) waitForFileInputSelector(ctx context.Context, timeout time.Duration) (string, error) {
	script := `(function() {
	  function dispatchClickSequence(target){
	    if(!target || !(target instanceof EventTarget)) return false;
	    const types = ['pointerdown', 'mousedown', 'pointerup', 'mouseup', 'click'];
	    for (const type of types) {
	      const common = { bubbles: true, cancelable: true, view: window };
	      let event;
	      if (type.startsWith('pointer') && 'PointerEvent' in window) {
	        event = new PointerEvent(type, { ...common, pointerId: 1, pointerType: 'mouse' });
	      } else {
	        event = new MouseEvent(type, common);
	      }
	      target.dispatchEvent(event);
	    }
	    return true;
	  }
	  const attr = 'data-surf-file-input-id';
	  const pickInput = () => {
	    const inputs = Array.from(document.querySelectorAll('input[type="file"]'));
	    return inputs.find((input) => !input.disabled) || inputs[0] || null;
	  };
	  let input = pickInput();
	  if (!input) {
	    const attachSelectors = [
	      'button[data-testid*="composer-plus"]',
	      'button[data-testid*="attach"]',
	      'button[aria-label*="Attach"]',
	      'button[aria-label*="attach"]',
	      'button[aria-label*="Upload"]',
	      'button[aria-label*="upload"]',
	    ];
	    for (const selector of attachSelectors) {
	      const button = document.querySelector(selector);
	      if (button) {
	        dispatchClickSequence(button);
	        break;
	      }
	    }
	    input = pickInput();
	  }
	  if (!input) return null;
	  let id = input.getAttribute(attr);
	  if (!id) {
	    id = 'surf-upload-' + Date.now() + '-' + Math.random().toString(36).slice(2, 8);
	    input.setAttribute(attr, id);
	  }
	  return '[' + attr + '="' + id + '"]';
	})()`

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		value, err := b.evaluate(ctx, script, extCallTimeout)
		if err != nil {
			return "", err
		}
		selector := strings.TrimSpace(asString(value))
		if selector != "" {
			return selector, nil
		}
		time.Sleep(250 * time.Millisecond)
	}
	return "", fmt.Errorf("ChatGPT file input not found")
}

func splitFileList(raw string) []string {
	parts := strings.Split(raw, ",")
	files := make([]string, 0, len(parts))
	for _, part := range parts {
		clean := strings.TrimSpace(part)
		if clean != "" {
			files = append(files, clean)
		}
	}
	return files
}

func (b *chatGPTBridge) uploadFiles(ctx context.Context, rawFiles string) error {
	files := splitFileList(rawFiles)
	if len(files) == 0 {
		return fmt.Errorf("Invalid file path")
	}
	selector, err := b.waitForFileInputSelector(ctx, 12*time.Second)
	if err != nil {
		return err
	}
	resp, err := b.caller.Request(ctx, map[string]any{
		"type":     "UPLOAD_FILE",
		"tabId":    b.tabID,
		"selector": selector,
		"files":    files,
	}, 45*time.Second)
	if err != nil {
		return err
	}
	if e := responseError(resp); e != "" {
		return errors.New(e)
	}
	if success, ok := resp["success"]; ok && !asBool(success) {
		return fmt.Errorf("File upload failed")
	}
	return nil
}

func (b *chatGPTBridge) waitForResponse(ctx context.Context, timeout time.Duration) (string, error) {
	deadline := time.Now().Add(timeout)
	expr := `(() => {
	  const turns = Array.from(document.querySelectorAll('article[data-testid^="conversation-turn"], div[data-testid^="conversation-turn"]'));
	  let lastAssistantTurn = null;
	  for (let i = turns.length - 1; i >= 0; i--) {
	    const node = turns[i];
	    const role = (node.getAttribute('data-message-author-role') || '').toLowerCase();
	    const turn = (node.getAttribute('data-turn') || '').toLowerCase();
	    if (role === 'assistant' || turn === 'assistant' || node.querySelector('[data-message-author-role="assistant"], [data-turn="assistant"]')) {
	      lastAssistantTurn = node;
	      break;
	    }
	  }
	  if (!lastAssistantTurn) {
	    return { text: '', stopVisible: Boolean(document.querySelector('[data-testid="stop-button"]')), finished: false };
	  }
	  const messageRoot = lastAssistantTurn.querySelector('[data-message-author-role="assistant"], [data-turn="assistant"]') || lastAssistantTurn;
	  const contentRoot = messageRoot.querySelector('.markdown') || messageRoot.querySelector('[data-message-content]') || messageRoot.querySelector('.prose') || messageRoot;
	  const text = (contentRoot?.innerText || contentRoot?.textContent || '').trim();
	  const stopVisible = Boolean(document.querySelector('[data-testid="stop-button"]'));
	  const finished = Boolean(lastAssistantTurn.querySelector('button[data-testid="copy-turn-action-button"], button[data-testid="good-response-turn-action-button"]'));
	  return { text, stopVisible, finished };
	})()`

	for time.Now().Before(deadline) {
		v, err := b.evaluate(ctx, expr, extCallTimeout)
		if err != nil {
			return "", err
		}
		m, _ := v.(map[string]any)
		text := strings.TrimSpace(asString(m["text"]))
		if text != "" && !asBool(m["stopVisible"]) && asBool(m["finished"]) {
			return text, nil
		}
		time.Sleep(400 * time.Millisecond)
	}
	return "", fmt.Errorf("Response timeout")
}

func hasChatGPTSessionCookie(raw any) bool {
	items, ok := raw.([]any)
	if !ok {
		return false
	}
	for _, item := range items {
		m, _ := item.(map[string]any)
		if m == nil {
			continue
		}
		if asString(m["name"]) == "__Secure-next-auth.session-token" && strings.TrimSpace(asString(m["value"])) != "" {
			return true
		}
	}
	return false
}

func responseError(resp map[string]any) string {
	if resp == nil {
		return ""
	}
	if s := strings.TrimSpace(asString(resp["error"])); s != "" {
		return s
	}
	return ""
}

func extractExceptionText(raw any) string {
	m, _ := raw.(map[string]any)
	if m == nil {
		return "Evaluation failed"
	}
	if ex, ok := m["exception"].(map[string]any); ok {
		if desc := strings.TrimSpace(asString(ex["description"])); desc != "" {
			return desc
		}
	}
	if text := strings.TrimSpace(asString(m["text"])); text != "" {
		return text
	}
	return "Evaluation failed"
}

func asString(v any) string {
	s, _ := v.(string)
	return s
}

func asBool(v any) bool {
	switch x := v.(type) {
	case bool:
		return x
	case string:
		n := strings.TrimSpace(strings.ToLower(x))
		return n == "1" || n == "true" || n == "yes" || n == "y"
	default:
		return false
	}
}

func toInt64(v any) (int64, bool) {
	switch x := v.(type) {
	case int:
		return int64(x), true
	case int32:
		return int64(x), true
	case int64:
		return x, true
	case float64:
		return int64(x), true
	case json.Number:
		n, err := x.Int64()
		if err != nil {
			return 0, false
		}
		return n, true
	case string:
		if strings.TrimSpace(x) == "" {
			return 0, false
		}
		var n int64
		if _, err := fmt.Sscan(strings.TrimSpace(x), &n); err != nil {
			return 0, false
		}
		return n, true
	default:
		return 0, false
	}
}

func toJSONString(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return "null"
	}
	return string(b)
}

func normalizeModel(model string) string {
	clean := strings.ToLower(strings.TrimSpace(model))
	if clean == "" {
		return ""
	}
	return strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			return r
		}
		return -1
	}, clean)
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func toStringSlice(v any) []string {
	if direct, ok := v.([]string); ok {
		return append([]string{}, direct...)
	}
	items, ok := v.([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(items))
	seen := map[string]struct{}{}
	for _, item := range items {
		s := strings.TrimSpace(asString(item))
		if s == "" {
			continue
		}
		if _, exists := seen[s]; exists {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}
