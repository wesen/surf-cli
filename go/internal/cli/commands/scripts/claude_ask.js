const CLAUDE_URL = 'https://claude.ai/new';

function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

function normalizeText(value) {
  return String(value || '')
    .replace(/\s+/g, ' ')
    .trim();
}

function normalizeMatch(value) {
  return normalizeText(value).toLowerCase();
}

function getModelButton() {
  return document.querySelector('[data-testid="model-selector-dropdown"]');
}

function getEditorHost() {
  return document.querySelector('[data-testid="chat-input"]');
}

function getEditor() {
  const host = getEditorHost();
  return host && host.editor ? host.editor : null;
}

function getSendButton() {
  return Array.from(document.querySelectorAll('button')).find(
    (el) => normalizeText(el.getAttribute('aria-label')) === 'Send message'
  ) || null;
}

function isVisible(el) {
  return !!el && el.getClientRects().length > 0;
}

function parseCurrentModelState() {
  const button = getModelButton();
  const rawText = normalizeText(button?.innerText || button?.textContent || '');
  const thinkingMode = /\bextended\b/i.test(rawText) ? 'extended' : 'standard';
  const model = normalizeText(rawText.replace(/\bExtended\b/i, ''));
  return {
    rawText,
    model,
    thinkingMode,
  };
}

function parseMenuItem(rawText) {
  const text = normalizeText(rawText);
  if (!text) {
    return { kind: 'unknown', name: '', description: '', rawText: '' };
  }
  if (text === 'More models') {
    return { kind: 'submenu', name: 'More models', description: '', rawText: text };
  }
  if (text.startsWith('Extended thinking')) {
    return {
      kind: 'thinking-toggle',
      name: 'Extended thinking',
      description: normalizeText(text.slice('Extended thinking'.length)),
      rawText: text,
    };
  }
  const modelPrefixes = ['Opus 4.6', 'Sonnet 4.6', 'Haiku 4.5', 'Opus 4.5', 'Opus 3', 'Sonnet 4.5'];
  for (const prefix of modelPrefixes) {
    if (text.startsWith(prefix)) {
      return {
        kind: 'model',
        name: prefix,
        description: normalizeText(text.slice(prefix.length)),
        rawText: text,
      };
    }
  }
  return {
    kind: 'unknown',
    name: text,
    description: '',
    rawText: text,
  };
}

async function openModelMenu() {
  const button = getModelButton();
  if (!button) {
    throw new Error('model selector not found');
  }
  button.click();
  for (let i = 0; i < 15; i += 1) {
    await sleep(100);
    if (button.getAttribute('aria-expanded') === 'true') {
      const items = getMenuItems();
      if (items.length > 0) {
        return items;
      }
    }
  }
  return getMenuItems();
}

function getOpenMenus() {
  return Array.from(document.querySelectorAll('[role="menu"]')).filter(isVisible);
}

function getMenuItems() {
  const openMenus = getOpenMenus();
  const seen = new Set();
  const items = [];
  for (const menu of openMenus) {
    for (const el of Array.from(menu.querySelectorAll('[role="menuitem"]'))) {
      if (seen.has(el)) {
        continue;
      }
      seen.add(el);
      items.push(el);
    }
  }
  return items;
}

function collectMenuState() {
  const items = getMenuItems();
  const parsedItems = items.map((el) => ({ el, parsed: parseMenuItem(el.innerText || el.textContent || '') }));
  return {
    items: parsedItems,
    models: parsedItems.filter((item) => item.parsed.kind === 'model'),
    moreModels: parsedItems.find((item) => item.parsed.kind === 'submenu') || null,
    thinkingToggle: parsedItems.find((item) => item.parsed.kind === 'thinking-toggle') || null,
  };
}

async function ensureSubmenuOpen() {
  let state = collectMenuState();
  if (!state.moreModels) {
    return state;
  }
  if (state.models.length > 3) {
    return state;
  }
  state.moreModels.el.click();
  let previousCount = 0;
  let stableCycles = 0;
  for (let i = 0; i < 12; i += 1) {
    await sleep(120);
    state = collectMenuState();
    const count = state.models.length;
    if (count === previousCount) {
      stableCycles += 1;
    } else {
      stableCycles = 0;
    }
    previousCount = count;
    if (count > 3 && stableCycles >= 2) {
      return state;
    }
  }
  return state;
}

async function listModels() {
  await openModelMenu();
  const current = parseCurrentModelState();
  let state = collectMenuState();
  state = await ensureSubmenuOpen();
  const seen = new Set();
  const models = [];
  for (const item of state.models) {
    if (seen.has(item.parsed.name)) {
      continue;
    }
    seen.add(item.parsed.name);
    models.push({
      name: item.parsed.name,
      description: item.parsed.description,
      rawText: item.parsed.rawText,
      thinkingModes: ['standard', 'extended'],
    });
  }
  return {
    kind: 'models',
    href: location.href,
    title: document.title,
    currentModel: current.model,
    currentThinkingMode: current.thinkingMode,
    models,
    thinkingModes: ['standard', 'extended'],
  };
}

async function selectModel(requested) {
  const current = parseCurrentModelState();
  if (!requested) {
    return { requested: '', applied: current.model };
  }
  const want = normalizeMatch(requested);
  await openModelMenu();
  let state = collectMenuState();
  let chosen = state.models.find((item) => {
    const name = normalizeMatch(item.parsed.name);
    const raw = normalizeMatch(item.parsed.rawText);
    return name === want || raw === want || name.includes(want) || want.includes(name);
  });
  if (!chosen) {
    state = await ensureSubmenuOpen();
    chosen = state.models.find((item) => {
      const name = normalizeMatch(item.parsed.name);
      const raw = normalizeMatch(item.parsed.rawText);
      return name === want || raw === want || name.includes(want) || want.includes(name);
    });
  }
  if (!chosen) {
    return {
      requested,
      applied: current.model,
      error: `Model not found: ${requested}`,
    };
  }
  chosen.el.scrollIntoView({ block: "nearest" });
  chosen.el.click();
  await sleep(400);
  return {
    requested,
    applied: chosen.parsed.name,
  };
}

async function ensureThinkingMode(requested) {
  const want = normalizeMatch(requested || '');
  const current = parseCurrentModelState();
  if (!want || want === 'default') {
    return { requested: requested || '', applied: current.thinkingMode };
  }
  if (want !== 'standard' && want !== 'extended') {
    return {
      requested,
      applied: current.thinkingMode,
      error: `Unsupported thinking mode: ${requested}`,
    };
  }
  if (current.thinkingMode === want) {
    return { requested, applied: current.thinkingMode };
  }
  await openModelMenu();
  const state = collectMenuState();
  if (!state.thinkingToggle) {
    return {
      requested,
      applied: current.thinkingMode,
      error: 'Extended thinking toggle not found',
    };
  }
  state.thinkingToggle.el.click();
  await sleep(400);
  return {
    requested,
    applied: parseCurrentModelState().thinkingMode,
  };
}

async function setPrompt(prompt) {
  const editor = getEditor();
  if (!editor) {
    throw new Error('Claude editor not found');
  }
  editor.chain().focus().clearContent().insertContent(String(prompt || '')).run();
  for (let i = 0; i < 10; i += 1) {
    if (normalizeText(editor.getText()) === normalizeText(prompt)) {
      return {
        text: editor.getText(),
        html: editor.getHTML(),
      };
    }
    await sleep(80);
  }
  return {
    text: editor.getText(),
    html: editor.getHTML(),
  };
}

function getAssistantNodes() {
  return Array.from(document.querySelectorAll('div.font-claude-response')).filter(
    (el) => !el.parentElement?.closest('div.font-claude-response')
  );
}

function getLatestAssistantText() {
  const nodes = getAssistantNodes();
  if (nodes.length === 0) {
    return '';
  }
  return normalizeText(nodes[nodes.length - 1].innerText || nodes[nodes.length - 1].textContent || '');
}

function getConversationTitle() {
  const titleButton = document.querySelector('[data-testid="chat-title-button"]');
  const title = normalizeText(titleButton?.innerText || titleButton?.textContent || '');
  if (title) {
    return title;
  }
  return normalizeText(document.title.replace(/\s*-\s*Claude$/, ''));
}

async function waitForResponse(timeoutMs) {
  const started = Date.now();
  let lastText = '';
  let stableCycles = 0;
  while (Date.now() - started < timeoutMs) {
    const text = getLatestAssistantText();
    const sendVisible = !!getSendButton();
    if (text) {
      if (text === lastText) {
        stableCycles += 1;
      } else {
        stableCycles = 0;
      }
      if (sendVisible && stableCycles >= 2) {
        return {
          text,
          waitedMs: Date.now() - started,
        };
      }
      lastText = text;
    }
    await sleep(700);
  }
  throw new Error('Timed out waiting for Claude response');
}

if (SURF_OPTIONS.action === 'list-models') {
  return await listModels();
}

const prompt = String(SURF_OPTIONS.prompt || '');
if (!prompt.trim()) {
  return { kind: 'error', error: 'prompt required' };
}

const modelSelection = await selectModel(SURF_OPTIONS.model || '');
if (modelSelection.error) {
  return {
    kind: 'error',
    href: location.href,
    title: document.title,
    currentModel: parseCurrentModelState().model,
    currentThinkingMode: parseCurrentModelState().thinkingMode,
    error: modelSelection.error,
  };
}

const thinkingSelection = await ensureThinkingMode(SURF_OPTIONS.thinkingMode || 'default');
if (thinkingSelection.error) {
  return {
    kind: 'error',
    href: location.href,
    title: document.title,
    currentModel: parseCurrentModelState().model,
    currentThinkingMode: parseCurrentModelState().thinkingMode,
    modelSelection,
    error: thinkingSelection.error,
  };
}

const promptState = await setPrompt(prompt);
const send = getSendButton();
if (!send) {
  return {
    kind: 'error',
    href: location.href,
    title: document.title,
    currentModel: parseCurrentModelState().model,
    currentThinkingMode: parseCurrentModelState().thinkingMode,
    promptText: promptState.text,
    promptHTML: promptState.html,
    error: 'Send button not found',
  };
}

send.click();
const response = await waitForResponse(Number(SURF_OPTIONS.promptTimeoutMs || 120000));
const current = parseCurrentModelState();

return {
  kind: 'response',
  href: location.href,
  title: document.title,
  conversationTitle: getConversationTitle(),
  currentModel: current.model,
  currentThinkingMode: current.thinkingMode,
  modelSelection,
  thinkingSelection,
  prompt,
  response: response.text,
  responseLength: response.text.length,
  waitedMs: response.waitedMs,
  createdFrom: CLAUDE_URL,
};
