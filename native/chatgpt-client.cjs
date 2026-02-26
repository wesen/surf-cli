const CHATGPT_URL = "https://chatgpt.com/";

const SELECTORS = {
  promptTextarea: '#prompt-textarea, [data-testid="composer-textarea"], textarea[name="prompt-textarea"], .ProseMirror, [contenteditable="true"][data-virtualkeyboard="true"]',
  sendButton: 'button[data-testid="send-button"], button[data-testid*="composer-send"], form button[type="submit"]',
  modelButton: '[data-testid="model-switcher-dropdown-button"]',
  menuContainer: '[role="menu"], [data-radix-collection-root]',
  menuItem: 'button, [role="menuitem"], [role="menuitemradio"], [data-testid*="model-switcher-"]',
  assistantMessage: '[data-message-author-role="assistant"], [data-turn="assistant"]',
  stopButton: '[data-testid="stop-button"]',
  finishedActions: 'button[data-testid="copy-turn-action-button"], button[data-testid="good-response-turn-action-button"]',
  conversationTurn: 'article[data-testid^="conversation-turn"], div[data-testid^="conversation-turn"]',
  fileInput: 'input[type="file"]',
  cloudflareScript: 'script[src*="/challenge-platform/"]',
};

function delay(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

function buildClickDispatcher() {
  return `function dispatchClickSequence(target){
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
  }`;
}

function hasRequiredCookies(cookies) {
  if (!cookies || !Array.isArray(cookies)) return false;
  const sessionCookie = cookies.find(
    (c) => c.name === "__Secure-next-auth.session-token" && c.value
  );
  return Boolean(sessionCookie);
}

async function evaluate(cdp, expression) {
  const result = await cdp(expression);
  if (result.exceptionDetails) {
    const desc = result.exceptionDetails.exception?.description || 
                 result.exceptionDetails.text || 
                 "Evaluation failed";
    throw new Error(desc);
  }
  if (result.error) {
    throw new Error(result.error);
  }
  return result.result?.value;
}

async function waitForPageLoad(cdp, timeoutMs = 45000) {
  const deadline = Date.now() + timeoutMs;
  while (Date.now() < deadline) {
    const ready = await evaluate(cdp, "document.readyState");
    if (ready === "complete" || ready === "interactive") {
      return;
    }
    await delay(100);
  }
  throw new Error("Page did not load in time");
}

async function isCloudflareBlocked(cdp) {
  const title = await evaluate(cdp, "document.title.toLowerCase()");
  if (title && title.includes("just a moment")) return true;
  const hasScript = await evaluate(
    cdp,
    `Boolean(document.querySelector('${SELECTORS.cloudflareScript}'))`
  );
  return hasScript;
}

async function checkLoginStatus(cdp) {
  const result = await evaluate(
    cdp,
    `(async () => {
      try {
        const response = await fetch('/backend-api/me', { 
          cache: 'no-store', 
          credentials: 'include' 
        });
        const hasLoginCta = Array.from(document.querySelectorAll('a[href*="/auth/login"], button'))
          .some(el => {
            const text = (el.textContent || '').toLowerCase().trim();
            return text.startsWith('log in') || text.startsWith('sign in');
          });
        return { 
          status: response.status, 
          hasLoginCta,
          url: location.href
        };
      } catch (e) {
        return { status: 0, error: e.message, url: location.href };
      }
    })()`
  );
  return result || { status: 0 };
}

async function waitForPromptReady(cdp, timeoutMs = 30000) {
  const deadline = Date.now() + timeoutMs;
  const selectors = JSON.stringify(SELECTORS.promptTextarea.split(", "));
  while (Date.now() < deadline) {
    const found = await evaluate(
      cdp,
      `(() => {
        const selectors = ${selectors};
        for (const selector of selectors) {
          const node = document.querySelector(selector);
          if (node && !node.hasAttribute('disabled')) {
            return true;
          }
        }
        return false;
      })()`
    );
    if (found) return true;
    await delay(200);
  }
  return false;
}

async function selectModel(cdp, desiredModel, timeoutMs = 8000) {
  await openModelMenu(cdp);
  const normalizedModel = desiredModel.toLowerCase().replace(/[^a-z0-9]/g, "");
  const deadline = Date.now() + timeoutMs;
  let lastAvailable = [];
  
  while (Date.now() < deadline) {
    const result = await evaluate(
      cdp,
      `(() => {
        ${buildClickDispatcher()}
        const normalizeDisplay = (text) => {
          const cleaned = (text || '').replace(/\\s+/g, ' ').trim();
          if (!cleaned) return '';
          const spaced = cleaned.replace(/([a-z])([A-Z])/g, '$1 $2').replace(/\\s+/g, ' ').trim();
          for (const known of ['Auto', 'Instant', 'Thinking', 'Pro']) {
            if (spaced === known || spaced.startsWith(known + ' ')) return known.toLowerCase();
          }
          return spaced;
        };
        const extractModelId = (blob) => {
          const lower = (blob || '').toLowerCase();
          const gptMatch = lower.match(/\\b(gpt[-a-z0-9._]+)\\b/);
          if (gptMatch) return gptMatch[1];
          const reasoningMatch = lower.match(/\\b(o[0-9][-a-z0-9._]*)\\b/);
          if (reasoningMatch) return reasoningMatch[1];
          return '';
        };
        const targetModel = ${JSON.stringify(normalizedModel)};
        const menuSelector = '${SELECTORS.menuContainer}';
        const itemSelector = '${SELECTORS.menuItem}';
        const normalize = (text) => (text || '').toLowerCase().replace(/[^a-z0-9]/g, '');
        
        const menu = document.querySelector(menuSelector);
        if (!menu) {
          return { found: false, waiting: true };
        }
        const items = Array.from(menu.querySelectorAll(itemSelector));
        const available = [];
        let bestMatch = null;
        let bestScore = 0;
        for (const item of items) {
          const label = normalizeDisplay(item.getAttribute('aria-label') || item.textContent || '');
          const text = normalize(item.textContent || '');
          const testId = normalize(item.getAttribute('data-testid') || '');
          const canonical = extractModelId([item.getAttribute('data-testid'), item.getAttribute('aria-label'), item.textContent].filter(Boolean).join(' ')) || label;
          if (canonical && canonical !== 'legacy models' && !available.includes(canonical)) available.push(canonical);
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
          return { found: true, success: true, label: bestMatch.textContent?.trim(), available };
        }
        return { found: true, success: false, error: 'No matching model in menu', available };
      })()`
    );
    
    if (result && result.found) {
      if (Array.isArray(result.available)) {
        lastAvailable = result.available;
      }
      if (result.success) {
        await delay(200);
        return result.label;
      }
      const suffix = lastAvailable.length > 0 ? ` Available: ${lastAvailable.join(", ")}` : "";
      throw new Error(`Model not found: ${desiredModel}.${suffix}`);
    }
    
    await delay(100);
  }
  
  const suffix = lastAvailable.length > 0 ? ` Available: ${lastAvailable.join(", ")}` : "";
  throw new Error(`Model not found: ${desiredModel} (timeout).${suffix}`);
}

async function openModelMenu(cdp, timeoutMs = 8000) {
  const deadline = Date.now() + timeoutMs;
  const modelButton = await evaluate(
    cdp,
    `(() => {
      const btn = document.querySelector('${SELECTORS.modelButton}');
      return btn ? true : false;
    })()`
  );
  if (!modelButton) {
    throw new Error("Model selector button not found");
  }
  await evaluate(
    cdp,
    `(() => {
      ${buildClickDispatcher()}
      const btn = document.querySelector('${SELECTORS.modelButton}');
      if (btn) dispatchClickSequence(btn);
    })()`
  );
  while (Date.now() < deadline) {
    const menuVisible = await evaluate(
      cdp,
      `(() => Boolean(document.querySelector('${SELECTORS.menuContainer}')))()`
    );
    if (menuVisible) return;
    await delay(100);
  }
  throw new Error("Model selector menu did not open");
}

async function readModelList(cdp, timeoutMs = 8000) {
  await openModelMenu(cdp, timeoutMs);
  const deadline = Date.now() + timeoutMs;
  while (Date.now() < deadline) {
    const snapshot = await evaluate(
      cdp,
      `(() => {
        const normalizeDisplay = (text) => {
          const cleaned = (text || '').replace(/\\s+/g, ' ').trim();
          if (!cleaned) return '';
          const spaced = cleaned.replace(/([a-z])([A-Z])/g, '$1 $2').replace(/\\s+/g, ' ').trim();
          for (const known of ['Auto', 'Instant', 'Thinking', 'Pro']) {
            if (spaced === known || spaced.startsWith(known + ' ')) return known.toLowerCase();
          }
          return spaced;
        };
        const extractModelId = (blob) => {
          const lower = (blob || '').toLowerCase();
          const gptMatch = lower.match(/\\b(gpt[-a-z0-9._]+)\\b/);
          if (gptMatch) return gptMatch[1];
          const reasoningMatch = lower.match(/\\b(o[0-9][-a-z0-9._]*)\\b/);
          if (reasoningMatch) return reasoningMatch[1];
          return '';
        };
        const menu = document.querySelector('${SELECTORS.menuContainer}');
        if (!menu) return { found: false };
        const items = Array.from(menu.querySelectorAll('${SELECTORS.menuItem}'));
        const models = [];
        let selected = null;
        for (const item of items) {
          const label = normalizeDisplay(item.getAttribute('aria-label') || item.textContent || '');
          const canonical = extractModelId([item.getAttribute('data-testid'), item.getAttribute('aria-label'), item.textContent].filter(Boolean).join(' ')) || label;
          if (!canonical) continue;
          if (canonical === 'legacy models') continue;
          if (!models.includes(canonical)) models.push(canonical);
          const ariaChecked = item.getAttribute('aria-checked');
          const dataState = item.getAttribute('data-state');
          if (ariaChecked === 'true' || dataState === 'checked') {
            selected = canonical;
          }
        }
        return { found: true, models, selected };
      })()`
    );
    if (snapshot && snapshot.found) {
      return {
        models: Array.isArray(snapshot.models) ? snapshot.models : [],
        selected: typeof snapshot.selected === "string" && snapshot.selected ? snapshot.selected : null,
      };
    }
    await delay(100);
  }
  throw new Error("Failed to read ChatGPT models");
}

async function typePrompt(cdp, inputCdp, prompt) {
  const selectors = JSON.stringify(SELECTORS.promptTextarea.split(", "));
  const encodedPrompt = JSON.stringify(prompt);
  const focused = await evaluate(
    cdp,
    `(() => {
      ${buildClickDispatcher()}
      const selectors = ${selectors};
      for (const selector of selectors) {
        const node = document.querySelector(selector);
        if (!node) continue;
        dispatchClickSequence(node);
        if (typeof node.focus === 'function') node.focus();
        const doc = node.ownerDocument;
        const selection = doc?.getSelection?.();
        if (selection) {
          const range = doc.createRange();
          range.selectNodeContents(node);
          range.collapse(false);
          selection.removeAllRanges();
          selection.addRange(range);
        }
        return true;
      }
      return false;
    })()`
  );
  if (!focused) {
    throw new Error("Failed to focus prompt textarea");
  }
  await inputCdp("Input.insertText", { text: prompt });
  await delay(300);
  const verified = await evaluate(
    cdp,
    `(() => {
      const selectors = ${selectors};
      for (const selector of selectors) {
        const node = document.querySelector(selector);
        if (!node) continue;
        const text = node.innerText || node.value || node.textContent || '';
        if (text.trim().length > 0) return true;
      }
      return false;
    })()`
  );
  if (!verified) {
    await evaluate(
      cdp,
      `(() => {
        const editor = document.querySelector('#prompt-textarea');
        const fallback = document.querySelector('textarea[name="prompt-textarea"]');
        if (fallback) {
          fallback.value = ${encodedPrompt};
          fallback.dispatchEvent(new InputEvent('input', { bubbles: true, data: ${encodedPrompt}, inputType: 'insertFromPaste' }));
        }
        if (editor) {
          editor.textContent = ${encodedPrompt};
          editor.dispatchEvent(new InputEvent('input', { bubbles: true, data: ${encodedPrompt}, inputType: 'insertFromPaste' }));
        }
      })()`
    );
  }
}

async function clickSend(cdp, inputCdp) {
  const selectors = SELECTORS.sendButton.split(", ");
  const selectorsJson = JSON.stringify(selectors);
  const deadline = Date.now() + 8000;
  while (Date.now() < deadline) {
    const result = await evaluate(
      cdp,
      `(() => {
        ${buildClickDispatcher()}
        const selectors = ${selectorsJson};
        let button = null;
        for (const selector of selectors) {
          button = document.querySelector(selector);
          if (button) break;
        }
        if (!button) return 'missing';
        const disabled = button.hasAttribute('disabled') || 
                        button.getAttribute('aria-disabled') === 'true' ||
                        button.getAttribute('data-disabled') === 'true';
        if (disabled) return 'disabled';
        dispatchClickSequence(button);
        return 'clicked';
      })()`
    );
    if (result === "clicked") return true;
    if (result === "missing") break;
    await delay(100);
  }
  await inputCdp("Input.dispatchKeyEvent", {
    type: "keyDown",
    key: "Enter",
    code: "Enter",
    windowsVirtualKeyCode: 13,
    nativeVirtualKeyCode: 13,
    text: "\r",
  });
  await inputCdp("Input.dispatchKeyEvent", {
    type: "keyUp",
    key: "Enter",
    code: "Enter",
    windowsVirtualKeyCode: 13,
    nativeVirtualKeyCode: 13,
  });
  return true;
}

function normalizeFileList(file) {
  if (!file) return [];
  if (Array.isArray(file)) {
    return file
      .map((item) => (typeof item === "string" ? item.trim() : ""))
      .filter(Boolean);
  }
  if (typeof file === "string") {
    return file
      .split(",")
      .map((item) => item.trim())
      .filter(Boolean);
  }
  return [];
}

async function waitForFileInputSelector(cdp, timeoutMs = 10000) {
  const deadline = Date.now() + timeoutMs;
  while (Date.now() < deadline) {
    const selector = await evaluate(
      cdp,
      `(() => {
        ${buildClickDispatcher()}
        const attr = "data-surf-file-input-id";
        const pickInput = () => {
          const inputs = Array.from(document.querySelectorAll('${SELECTORS.fileInput}'));
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
          id = \`surf-upload-\${Date.now()}-\${Math.random().toString(36).slice(2, 8)}\`;
          input.setAttribute(attr, id);
        }
        return \`[\${attr}="\${id}"]\`;
      })()`
    );
    if (typeof selector === "string" && selector.trim() !== "") {
      return selector;
    }
    await delay(250);
  }
  return null;
}

async function uploadChatGPTFiles(cdp, uploadFile, tabId, file, log) {
  const files = normalizeFileList(file);
  if (files.length === 0) {
    throw new Error("Invalid file path");
  }
  if (typeof uploadFile !== "function") {
    throw new Error("File upload not supported by host");
  }
  const selector = await waitForFileInputSelector(cdp, 12000);
  if (!selector) {
    throw new Error("ChatGPT file input not found");
  }
  const result = await uploadFile(tabId, selector, files);
  if (!result || result.error || result.success === false) {
    throw new Error(result?.error || "File upload failed");
  }
  log(`Uploaded ${files.length} file(s)`);
  await delay(500);
}

async function waitForResponse(cdp, timeoutMs = 2700000) {
  const deadline = Date.now() + timeoutMs;
  let previousLength = 0;
  let stableCycles = 0;
  const requiredStableCycles = 6;
  const minStableMs = 1200;
  let lastChangeAt = Date.now();
  while (Date.now() < deadline) {
    const snapshot = await evaluate(
      cdp,
      `(() => {
        const CONVERSATION_SELECTOR = '${SELECTORS.conversationTurn}';
        const ASSISTANT_SELECTOR = '${SELECTORS.assistantMessage}';
        const STOP_SELECTOR = '${SELECTORS.stopButton}';
        const FINISHED_SELECTOR = '${SELECTORS.finishedActions}';
        const isAssistantTurn = (node) => {
          if (!(node instanceof HTMLElement)) return false;
          const role = (node.getAttribute('data-message-author-role') || '').toLowerCase();
          if (role === 'assistant') return true;
          const turn = (node.getAttribute('data-turn') || '').toLowerCase();
          if (turn === 'assistant') return true;
          return Boolean(node.querySelector(ASSISTANT_SELECTOR));
        };
        const turns = Array.from(document.querySelectorAll(CONVERSATION_SELECTOR));
        let lastAssistantTurn = null;
        for (let i = turns.length - 1; i >= 0; i--) {
          if (isAssistantTurn(turns[i])) {
            lastAssistantTurn = turns[i];
            break;
          }
        }
        if (!lastAssistantTurn) {
          return { text: '', stopVisible: Boolean(document.querySelector(STOP_SELECTOR)), finished: false };
        }
        const messageRoot = lastAssistantTurn.querySelector(ASSISTANT_SELECTOR) || lastAssistantTurn;
        const contentRoot = messageRoot.querySelector('.markdown') || 
                           messageRoot.querySelector('[data-message-content]') ||
                           messageRoot.querySelector('.prose') ||
                           messageRoot;
        const text = (contentRoot?.innerText || contentRoot?.textContent || '').trim();
        const stopVisible = Boolean(document.querySelector(STOP_SELECTOR));
        const finished = Boolean(lastAssistantTurn.querySelector(FINISHED_SELECTOR));
        const messageId = messageRoot.getAttribute('data-message-id') || null;
        return { text, stopVisible, finished, messageId, turnIndex: turns.length - 1 };
      })()`
    );
    if (!snapshot) {
      await delay(400);
      continue;
    }
    const currentLength = (snapshot.text || "").length;
    if (currentLength > previousLength) {
      previousLength = currentLength;
      stableCycles = 0;
      lastChangeAt = Date.now();
    } else {
      stableCycles++;
    }
    const stableMs = Date.now() - lastChangeAt;
    if (!snapshot.stopVisible) {
      const stableEnough = stableCycles >= requiredStableCycles && stableMs >= minStableMs;
      const finishedVisible = snapshot.finished;
      if ((finishedVisible || stableEnough) && currentLength > 0) {
        return {
          text: snapshot.text,
          messageId: snapshot.messageId,
          turnIndex: snapshot.turnIndex,
        };
      }
    }
    await delay(400);
  }
  throw new Error("Response timeout");
}

async function query(options) {
  const {
    prompt,
    model,
    file,
    timeout = 2700000,
    getCookies,
    createTab,
    closeTab,
    cdpEvaluate,
    cdpCommand,
    uploadFile,
    log = () => {},
  } = options;
  const startTime = Date.now();
  log("Starting ChatGPT query");
  const { cookies } = await getCookies();
  if (!hasRequiredCookies(cookies)) {
    throw new Error("ChatGPT login required");
  }
  log(`Got ${cookies.length} cookies`);
  const tabInfo = await createTab();
  const { tabId } = tabInfo;
  if (!tabId) {
    throw new Error("Failed to create ChatGPT tab");
  }
  log(`Created tab ${tabId}`);
  
  const cdp = (expr) => cdpEvaluate(tabId, expr);
  const inputCdp = (method, params) => cdpCommand(tabId, method, params);
  
  try {
    await waitForPageLoad(cdp);
    log("Page loaded");
    if (await isCloudflareBlocked(cdp)) {
      throw new Error("Cloudflare challenge detected - complete in browser");
    }
    const loginStatus = await checkLoginStatus(cdp);
    if (loginStatus.status !== 200 || loginStatus.hasLoginCta) {
      throw new Error("ChatGPT login required");
    }
    log("Login verified");
    const promptReady = await waitForPromptReady(cdp);
    if (!promptReady) {
      throw new Error("Prompt textarea not ready");
    }
    log("Prompt ready");
    if (model) {
      const selectedLabel = await selectModel(cdp, model);
      log(`Selected model: ${selectedLabel}`);
    }
    if (file) {
      await uploadChatGPTFiles(cdp, uploadFile, tabId, file, log);
    }
    await typePrompt(cdp, inputCdp, prompt);
    log("Prompt typed");
    await clickSend(cdp, inputCdp);
    log("Prompt sent, waiting for response...");
    const response = await waitForResponse(cdp, timeout);
    log(`Response received (${response.text.length} chars)`);
    return {
      response: response.text,
      model: model || "current",
      messageId: response.messageId,
      tookMs: Date.now() - startTime,
    };
  } finally {
    await closeTab(tabId).catch(() => {});
  }
}

async function listModels(options) {
  const {
    timeout = 30000,
    getCookies,
    createTab,
    closeTab,
    cdpEvaluate,
    log = () => {},
  } = options;
  const startTime = Date.now();
  log("Listing ChatGPT models");
  const { cookies } = await getCookies();
  if (!hasRequiredCookies(cookies)) {
    throw new Error("ChatGPT login required");
  }
  const tabInfo = await createTab();
  const { tabId } = tabInfo;
  if (!tabId) {
    throw new Error("Failed to create ChatGPT tab");
  }
  const cdp = (expr) => cdpEvaluate(tabId, expr);
  try {
    await waitForPageLoad(cdp);
    if (await isCloudflareBlocked(cdp)) {
      throw new Error("Cloudflare challenge detected - complete in browser");
    }
    const loginStatus = await checkLoginStatus(cdp);
    if (loginStatus.status !== 200 || loginStatus.hasLoginCta) {
      throw new Error("ChatGPT login required");
    }
    const promptReady = await waitForPromptReady(cdp);
    if (!promptReady) {
      throw new Error("Prompt textarea not ready");
    }
    const snapshot = await readModelList(cdp, Math.min(timeout, 20000));
    return {
      models: snapshot.models,
      selected: snapshot.selected,
      tookMs: Date.now() - startTime,
    };
  } finally {
    await closeTab(tabId).catch(() => {});
  }
}

module.exports = { query, listModels, hasRequiredCookies, CHATGPT_URL };
