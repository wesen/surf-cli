const sleep = (ms) => new Promise((resolve) => setTimeout(resolve, ms));

const options = typeof SURF_OPTIONS === 'object' && SURF_OPTIONS !== null ? SURF_OPTIONS : {};
const action = options.action || 'run';
const prompt = options.prompt || '';
const promptTimeoutMs = Number.isFinite(options.promptTimeoutMs) ? Number(options.promptTimeoutMs) : 120000;
const assistantName = options.assistant || '';
const modelName = options.model || '';
const lensName = options.lens || '';
const webSearchMode = options.webSearchMode || 'keep';
const tagNames = Array.isArray(options.tags) ? options.tags : [];
const createTags = !!options.createTags;
const listAssistants = !!options.listAssistants;
const listCustomAssistants = !!options.listCustomAssistants;
const listModels = !!options.listModels;
const listLenses = !!options.listLenses;
const listTags = !!options.listTags;
const listAllOptions = !!options.listAllOptions;

function normalizeText(value) {
  return (value || '').replace(/\s+/g, ' ').trim();
}

async function waitFor(fn, timeoutMs = 5000, intervalMs = 100) {
  const started = Date.now();
  while (Date.now() - started < timeoutMs) {
    const value = fn();
    if (value) {
      return { value, waitedMs: Date.now() - started };
    }
    await sleep(intervalMs);
  }
  return null;
}

function realClick(el) {
  if (!el) {
    return;
  }
  el.focus();
  for (const type of ['pointerdown', 'mousedown', 'pointerup', 'mouseup', 'click']) {
    el.dispatchEvent(new MouseEvent(type, { bubbles: true, cancelable: true, view: window }));
  }
}

function getProfileButton() {
  return document.querySelector('#profile-select');
}

function getProfileDialog() {
  return document.querySelector('dialog.promptOptionsSelector[open]');
}

function getLensButton() {
  return document.querySelector('#lens-select');
}

function getLensListbox() {
  return document.querySelector('ul[role="listbox"][aria-label="Lens chooser"]');
}

function getWebSearchInput() {
  return document.querySelector('#internet-access input[type="checkbox"], label#internet-access input[type="checkbox"], input[aria-label="Web access"]');
}

function getWebSearchLabel() {
  return document.querySelector('#internet-access, label#internet-access');
}

function getTextarea() {
  return document.querySelector('#promptBox, textarea[placeholder="Ask Assistant"]');
}

function getSubmitButton() {
  return document.querySelector('button#submit.submit[type="submit"], button[aria-label="Submit"]');
}

function getTagsButton() {
  return document.querySelector('#tags-add');
}

function getTagsDialog() {
  return document.querySelector('#tags dialog.promptOptionsSelector');
}

function summarizeProfileButton() {
  const btn = getProfileButton();
  if (!btn) {
    return null;
  }
  return {
    text: normalizeText(btn.textContent || ''),
    ariaLabel: btn.getAttribute('aria-label'),
    ariaExpanded: btn.getAttribute('aria-expanded'),
    title: btn.getAttribute('title'),
  };
}

async function ensureProfileDialogOpen() {
  let button = getProfileButton();
  if (!button) {
    const ready = await waitFor(() => getProfileButton(), 15000, 150);
    button = ready?.value || null;
  }
  if (!button) {
    throw new Error('Profile button not found');
  }
  let dialog = getProfileDialog();
  if (dialog) {
    return dialog;
  }
  realClick(button);
  let ready = await waitFor(() => getProfileDialog(), 5000, 100);
  if (ready) {
    return ready.value;
  }
  realClick(button);
  ready = await waitFor(() => getProfileDialog(), 5000, 100);
  if (!ready) {
    throw new Error('Profile dialog did not open');
  }
  return ready.value;
}

async function ensureLensListboxOpen() {
  let button = getLensButton();
  if (!button) {
    const readyButton = await waitFor(() => getLensButton(), 15000, 150);
    button = readyButton?.value || null;
  }
  if (!button) {
    throw new Error('Lens button not found');
  }
  let listbox = getLensListbox();
  if (listbox) {
    return listbox;
  }
  realClick(button);
  const ready = await waitFor(() => getLensListbox(), 5000, 100);
  if (!ready) {
    throw new Error('Lens chooser did not open');
  }
  return ready.value;
}

async function ensureTagsDialogOpen() {
  let button = getTagsButton();
  if (!button) {
    const readyButton = await waitFor(() => getTagsButton(), 15000, 150);
    button = readyButton?.value || null;
  }
  if (!button) {
    throw new Error('Tags button not found');
  }
  let dialog = getTagsDialog();
  if (dialog) {
    return dialog;
  }
  realClick(button);
  const ready = await waitFor(() => getTagsDialog(), 5000, 100);
  if (!ready) {
    throw new Error('Tags dialog did not open');
  }
  return ready.value;
}

function profileEntries() {
  const dialog = getProfileDialog();
  if (!dialog) {
    return [];
  }
  return Array.from(dialog.querySelectorAll('li.option[role="option"]')).map((el) => {
    const listbox = el.closest('[role="listbox"]');
    const labelledBy = listbox?.getAttribute('aria-labelledby') || '';
    const heading = labelledBy ? normalizeText(document.getElementById(labelledBy)?.textContent || '') : '';
    const ariaLabel = normalizeText(listbox?.getAttribute('aria-label') || '');
    const title = normalizeText(el.querySelector('.center')?.textContent || el.textContent || '');
    const subtitle = normalizeText(el.querySelector('.subtitle')?.textContent || '');
    const modelId = el.querySelector('svg.model-icon[data-model]')?.getAttribute('data-model') || '';
    let kind = 'model';
    if (ariaLabel === 'Custom Assistant') {
      kind = 'custom-assistant';
    } else if (heading === 'Kagi') {
      kind = 'assistant';
    }
    return {
      kind,
      section: heading || ariaLabel || null,
      name: title,
      subtitle: subtitle || null,
      modelId: modelId || null,
      selected: el.classList.contains('selected') || el.getAttribute('aria-selected') === 'true',
    };
  }).filter((entry) => entry.name);
}

function filterProfileEntries(entries) {
  if (listAllOptions) {
    return entries;
  }
  return entries.filter((entry) => {
    if (entry.kind === 'assistant') {
      return listAssistants;
    }
    if (entry.kind === 'custom-assistant') {
      return listCustomAssistants;
    }
    return listModels;
  });
}

async function listProfiles() {
  await ensureProfileDialogOpen();
  const entries = profileEntries();
  return {
    kind: 'options',
    href: location.href,
    title: document.title,
    profileButton: summarizeProfileButton(),
    profiles: filterProfileEntries(entries),
  };
}

async function listLensesOnly() {
  const listbox = await ensureLensListboxOpen();
  const lenses = Array.from(listbox.querySelectorAll('li.option[role="option"]')).map((el) => ({
    kind: 'lens',
    name: normalizeText(el.textContent || ''),
    selected: el.getAttribute('aria-selected') === 'true',
  })).filter((entry) => entry.name);
  return {
    kind: 'options',
    href: location.href,
    title: document.title,
    profileButton: summarizeProfileButton(),
    lenses,
  };
}

function tagEntries() {
  const dialog = getTagsDialog();
  if (!dialog) {
    return [];
  }
  return Array.from(dialog.querySelectorAll('input[type="checkbox"][name="tag"]')).map((input) => ({
    kind: 'tag',
    name: normalizeText(input.closest('label')?.textContent || ''),
    value: input.value || null,
    selected: !!input.checked,
  })).filter((entry) => entry.name);
}

async function listTagsOnly() {
  await ensureTagsDialogOpen();
  return {
    kind: 'options',
    href: location.href,
    title: document.title,
    profileButton: summarizeProfileButton(),
    tags: tagEntries(),
  };
}

function matchesProfileEntry(entry) {
  const entryName = (entry.name || '').toLowerCase();
  const entryModelID = (entry.modelId || '').toLowerCase();
  if (assistantName) {
    return entry.kind !== 'model' && entryName === assistantName.toLowerCase();
  }
  if (modelName) {
    const want = modelName.toLowerCase();
    return entry.kind === 'model' && (entryModelID === want || entryName === want);
  }
  return false;
}

async function selectProfileIfRequested() {
  if (!assistantName && !modelName) {
    return null;
  }
  const dialog = await ensureProfileDialogOpen();
  const entries = profileEntries();
  const target = entries.find((entry) => matchesProfileEntry(entry));
  if (!target) {
    throw new Error(assistantName ? `Assistant not found: ${assistantName}` : `Model not found: ${modelName}`);
  }
  const options = Array.from(dialog.querySelectorAll('li.option[role="option"]'));
  const match = options.find((el) => {
    const text = normalizeText(el.querySelector('.center')?.textContent || el.textContent || '');
    const modelId = el.querySelector('svg.model-icon[data-model]')?.getAttribute('data-model') || '';
    if (assistantName) {
      return text.toLowerCase() === assistantName.toLowerCase();
    }
    const want = modelName.toLowerCase();
    return modelId.toLowerCase() === want || text.toLowerCase() === want;
  });
  if (!match) {
    throw new Error('Matching profile option element not found');
  }
  const before = summarizeProfileButton();
  realClick(match);
  await sleep(900);
  return {
    type: assistantName ? 'assistant' : 'model',
    requested: assistantName || modelName,
    selected: normalizeText(match.textContent || ''),
    modelId: match.querySelector('svg.model-icon[data-model]')?.getAttribute('data-model') || null,
    before,
    after: summarizeProfileButton(),
  };
}

async function setLensIfRequested() {
  if (!lensName) {
    return null;
  }
  const listbox = await ensureLensListboxOpen();
  const options = Array.from(listbox.querySelectorAll('li.option[role="option"]'));
  const match = options.find((el) => normalizeText(el.textContent || '').toLowerCase() === lensName.toLowerCase());
  if (!match) {
    throw new Error(`Lens not found: ${lensName}`);
  }
  realClick(match);
  await sleep(300);
  return {
    requested: lensName,
    selected: normalizeText(match.textContent || ''),
  };
}

async function setWebSearchIfRequested() {
  if (webSearchMode === 'keep') {
    return null;
  }
  const input = getWebSearchInput();
  const label = getWebSearchLabel();
  if (!input || !label) {
    throw new Error('Web Search toggle not found');
  }
  const before = !!input.checked;
  const wanted = webSearchMode === 'on';
  if (before !== wanted) {
    realClick(label);
    await sleep(250);
  }
  return {
    requested: webSearchMode,
    before,
    after: !!input.checked,
  };
}

async function ensureTagExists(name) {
  const dialog = await ensureTagsDialogOpen();
  const existing = tagEntries().find((entry) => entry.name.toLowerCase() === name.toLowerCase());
  if (existing) {
    return existing;
  }
  if (!createTags) {
    throw new Error(`Tag not found: ${name}`);
  }
  const input = dialog.querySelector('input[type="search"][data-selector="input"]');
  const createButton = dialog.querySelector('button.create-new[data-selector="createNew"]');
  if (!input || !createButton) {
    throw new Error('Tag creation controls not found');
  }
  input.focus();
  input.value = name;
  input.dispatchEvent(new Event('input', { bubbles: true }));
  input.dispatchEvent(new Event('change', { bubbles: true }));
  const createReady = await waitFor(() => {
    if (!createButton.disabled) {
      return createButton;
    }
    return null;
  }, 5000, 100);
  if (!createReady) {
    throw new Error(`Could not create tag: ${name}`);
  }
  realClick(createButton);
  const created = await waitFor(() => tagEntries().find((entry) => entry.name.toLowerCase() === name.toLowerCase()), 5000, 100);
  if (!created) {
    const visible = currentVisibleTags();
    if (visible.some((tag) => tag.toLowerCase() === name.toLowerCase())) {
      return { kind: 'tag', name, value: null, selected: true, created: true };
    }
    throw new Error(`Created tag did not appear: ${name}`);
  }
  return { ...created.value, created: true };
}

function currentVisibleTags() {
  const tags = document.querySelector('#tags');
  if (!tags) {
    return [];
  }
  return Array.from(tags.querySelectorAll('.untagged, .dialog-tag, .tag, [class*="tag"]')).map((el) => normalizeText(el.textContent || '')).filter((name) => name && name !== '+');
}

async function applyTagsIfRequested() {
  if (tagNames.length === 0) {
    return null;
  }
  await ensureTagsDialogOpen();
  const applied = [];
  for (const rawName of tagNames) {
    const name = normalizeText(rawName);
    if (!name) {
      continue;
    }
    const tag = await ensureTagExists(name);
    const dialog = await ensureTagsDialogOpen();
    const input = Array.from(dialog.querySelectorAll('input[type="checkbox"][name="tag"]')).find((el) => normalizeText(el.closest('label')?.textContent || '').toLowerCase() === name.toLowerCase());
    if (input && !input.checked) {
      realClick(input.closest('label') || input);
      await sleep(150);
    }
    applied.push({
      name,
      value: tag.value || null,
      created: !!tag.created,
      selected: input ? !!input.checked : true,
    });
  }
  return {
    requested: tagNames,
    createTags,
    applied,
    visibleTags: currentVisibleTags(),
  };
}

function extractMetadataMap() {
  const info = Array.from(document.querySelectorAll('.message-info')).pop();
  const metadata = {};
  if (!info) {
    return metadata;
  }
  const items = Array.from(info.querySelectorAll('li'));
  for (const item of items) {
    const attribute = normalizeText(item.querySelector('.attribute')?.textContent || '');
    const value = normalizeText(item.querySelector('.value')?.textContent || '');
    if (attribute) {
      metadata[attribute] = value || null;
    }
  }
  return metadata;
}

function extractAssistantResponse() {
  const bubbles = Array.from(document.querySelectorAll('#chat_box .chat_bubble'));
  if (bubbles.length < 2) {
    return null;
  }
  const assistantBubble = bubbles[bubbles.length - 1];
  const userBubble = bubbles[bubbles.length - 2];
  const assistantContent = assistantBubble.querySelector('.content') || assistantBubble;
  const assistantClone = assistantContent.cloneNode(true);
  let thinkingText = '';
  const details = assistantClone.querySelector('details');
  if (details) {
    const thinkingClone = details.cloneNode(true);
    const summary = thinkingClone.querySelector('summary');
    if (summary) {
      summary.remove();
    }
    thinkingText = normalizeText(thinkingClone.textContent || '');
    details.remove();
  }
  const answerText = normalizeText(assistantClone.textContent || '');
  const metadata = extractMetadataMap();
  return {
    href: location.href,
    title: document.title,
    conversationTitle: normalizeText(document.querySelector('h1')?.textContent || document.title || ''),
    userPrompt: normalizeText(userBubble.querySelector('.content')?.textContent || userBubble.textContent || prompt),
    profileText: normalizeText(assistantBubble.querySelector('.model')?.textContent || ''),
    response: answerText,
    responseLength: answerText.length,
    thinking: thinkingText || null,
    thinkingLength: thinkingText.length,
    metadata,
    bubbleCount: bubbles.length,
    readOnly: /read-only/i.test(normalizeText(document.querySelector('#prompt-box')?.textContent || '')),
  };
}

async function submitPromptAndWait() {
  let textarea = getTextarea();
  let submit = getSubmitButton();
  if (!textarea || !submit) {
    const ready = await waitFor(() => {
      const nextTextarea = getTextarea();
      const nextSubmit = getSubmitButton();
      if (nextTextarea && nextSubmit) {
        return { textarea: nextTextarea, submit: nextSubmit };
      }
      return null;
    }, 15000, 150);
    textarea = ready?.value?.textarea || null;
    submit = ready?.value?.submit || null;
  }
  if (!textarea || !submit) {
    throw new Error('Prompt controls not found');
  }
  textarea.focus();
  textarea.value = prompt;
  textarea.dispatchEvent(new Event('input', { bubbles: true }));
  textarea.dispatchEvent(new Event('change', { bubbles: true }));
  await sleep(100);
  realClick(submit);

  const started = Date.now();
  let lastText = '';
  let stableCycles = 0;
  while (Date.now() - started < promptTimeoutMs) {
    const snapshot = extractAssistantResponse();
    if (snapshot && snapshot.response) {
      if (snapshot.response === lastText) {
        stableCycles += 1;
      } else {
        lastText = snapshot.response;
        stableCycles = 0;
      }
      const metadataCount = Object.keys(snapshot.metadata || {}).length;
      if ((metadataCount > 0 && stableCycles >= 2) || stableCycles >= 5) {
        snapshot.waitedMs = Date.now() - started;
        return snapshot;
      }
    }
    await sleep(400);
  }
  throw new Error('Timed out waiting for Kagi Assistant response');
}

if (action === 'list-profiles') {
  return await listProfiles();
}
if (action === 'list-lenses') {
  return await listLensesOnly();
}
if (action === 'list-tags') {
  return await listTagsOnly();
}
if (action === 'list-all') {
  const profilesResult = await listProfiles();
  const lensesResult = await listLensesOnly();
  const tagsResult = await listTagsOnly();
  return {
    kind: 'options',
    href: location.href,
    title: document.title,
    profileButton: profilesResult.profileButton,
    profiles: profilesResult.profiles || [],
    lenses: lensesResult.lenses || [],
    tags: tagsResult.tags || [],
  };
}

if (!prompt) {
  throw new Error('Prompt is required unless a list mode is requested');
}

const profileSelection = await selectProfileIfRequested();
const lensSelection = await setLensIfRequested();
const webSearchSelection = await setWebSearchIfRequested();
const response = await submitPromptAndWait();
const tagSelection = await applyTagsIfRequested();

return {
  kind: 'response',
  href: response.href,
  title: response.title,
  conversationTitle: response.conversationTitle,
  prompt: response.userPrompt,
  profileButton: summarizeProfileButton(),
  profileSelection,
  lensSelection,
  webSearchSelection,
  tagSelection,
  response: response.response,
  responseLength: response.responseLength,
  thinking: response.thinking,
  thinkingLength: response.thinkingLength,
  metadata: response.metadata,
  waitedMs: response.waitedMs,
  readOnly: response.readOnly,
};
