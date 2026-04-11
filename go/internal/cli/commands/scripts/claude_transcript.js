function normalizeText(value) {
  return String(value || '')
    .replace(/\r\n/g, '\n')
    .replace(/\n{3,}/g, '\n\n')
    .replace(/[ \t]+\n/g, '\n')
    .replace(/\s+$/g, '')
    .trim();
}

function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

function dedupeLinks(items) {
  const seen = new Set();
  const out = [];
  for (const item of items) {
    const key = `${item.href}::${item.text}`;
    if (seen.has(key)) {
      continue;
    }
    seen.add(key);
    out.push(item);
  }
  return out;
}

function parseCurrentModelState() {
  const button = document.querySelector('[data-testid="model-selector-dropdown"]');
  const rawText = String(button?.innerText || button?.textContent || '').replace(/\s+/g, ' ').trim();
  return {
    rawText,
    model: rawText.replace(/\bExtended\b/i, '').replace(/\s+/g, ' ').trim(),
    thinkingMode: /\bextended\b/i.test(rawText) ? 'extended' : 'standard',
  };
}

function getConversationTitle() {
  const titleButton = document.querySelector('[data-testid="chat-title-button"]');
  const title = String(titleButton?.innerText || titleButton?.textContent || '').replace(/\s+/g, ' ').trim();
  if (title) {
    return title;
  }
  return String(document.title || '').replace(/\s*-\s*Claude$/, '').trim();
}

function outermostAssistantNodes() {
  return Array.from(document.querySelectorAll('div.font-claude-response')).filter(
    (el) => !el.parentElement?.closest('div.font-claude-response')
  );
}

async function expandSearchWebSection(node) {
  if (!node) {
    return null;
  }
  const button = Array.from(node.querySelectorAll('button')).find((el) =>
    /searched the web/i.test(
      `${normalizeText(el.innerText || el.textContent)} ${normalizeText(el.getAttribute('aria-label'))}`
    )
  );
  if (!button) {
    return null;
  }
  if (button.getAttribute('aria-expanded') !== 'true') {
    button.click();
    await sleep(500);
  }
  let container = button.parentElement;
  while (container && container !== node) {
    if (
      container.querySelector('div[class*="transition-[grid-template-rows]"]') ||
      container.querySelector('div.border-\\[0\\.5px\\]') ||
      container.querySelector('a[href]')
    ) {
      return { button, container };
    }
    container = container.parentElement;
  }
  return { button, container: node };
}

async function extractSearchWeb(node) {
  const expanded = await expandSearchWebSection(node);
  if (!expanded) {
    return null;
  }
  const { button, container } = expanded;
  const results = dedupeLinks(
    Array.from(container.querySelectorAll('a[href]')).map((a) => ({
      href: a.href,
      text: normalizeText(a.innerText || a.textContent),
      host: normalizeText(
        a.querySelector('.text-xs, .text-text-400')?.innerText ||
          a.querySelector('.text-text-400')?.textContent ||
          ''
      ),
    }))
  ).filter((item) => item.text);
  const queries = Array.from(container.querySelectorAll('button'))
    .map((el) => normalizeText(el.innerText || el.textContent))
    .filter((text, index, items) => text && text !== normalizeText(button.innerText || button.textContent) && items.indexOf(text) === index);
  return {
    label: normalizeText(button.innerText || button.textContent),
    expanded: button.getAttribute('aria-expanded') === 'true',
    text: normalizeText(container.innerText || container.textContent),
    results,
    queries,
  };
}

async function extractCitations(node) {
  const searchWeb = await extractSearchWeb(node);
  const citations = dedupeLinks(
    Array.from(node.querySelectorAll('a[href]')).map((a) => ({
      href: a.href,
      text: normalizeText(a.innerText || a.textContent),
      parentText: normalizeText(a.parentElement?.innerText || a.parentElement?.textContent),
    }))
  ).filter((item) => item.text);
  return { citations, searchWeb };
}

function sortByDocumentOrder(items) {
  return items.sort((a, b) => {
    if (a.node === b.node) return 0;
    const pos = a.node.compareDocumentPosition(b.node);
    if (pos & Node.DOCUMENT_POSITION_FOLLOWING) return -1;
    if (pos & Node.DOCUMENT_POSITION_PRECEDING) return 1;
    return 0;
  });
}

const current = parseCurrentModelState();

const userTurns = Array.from(document.querySelectorAll('div[data-testid="user-message"]'))
  .map((node) => ({
    role: 'user',
    node,
    text: normalizeText(node.innerText || node.textContent || ''),
  }))
  .filter((item) => item.text);

const assistantTurns = [];
for (const node of outermostAssistantNodes()) {
  const text = normalizeText(node.innerText || node.textContent || '');
  if (!text) {
    continue;
  }
  const extracted = await extractCitations(node);
  assistantTurns.push({
    role: 'assistant',
    node,
    text,
    citations: extracted.citations,
    searchWeb: extracted.searchWeb,
  });
}

const transcript = sortByDocumentOrder(userTurns.concat(assistantTurns)).map((item, index) => ({
  index,
  role: item.role,
  text: item.text,
  textLength: item.text.length,
  citations: item.citations || [],
  searchWeb: item.searchWeb || null,
}));

return {
  href: location.href,
  title: document.title,
  conversationTitle: getConversationTitle(),
  currentModel: current.model,
  currentThinkingMode: current.thinkingMode,
  turnCount: transcript.length,
  transcript,
};
