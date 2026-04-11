function normalizeText(value) {
  return String(value || '')
    .replace(/\r\n/g, '\n')
    .replace(/\n{3,}/g, '\n\n')
    .replace(/[ \t]+\n/g, '\n')
    .replace(/\s+$/g, '')
    .trim();
}

function parseCurrentModel() {
  const button = document.querySelector('[data-testid="model-selector-dropdown"]');
  const text = String(button?.innerText || button?.textContent || '').replace(/\s+/g, ' ').trim();
  for (const prefix of ['Opus 4.6', 'Sonnet 4.6', 'Haiku 4.5', 'Extended thinking']) {
    if (text.includes(prefix)) {
      return prefix;
    }
  }
  return text;
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

function sortByDocumentOrder(items) {
  return items.sort((a, b) => {
    if (a.node === b.node) return 0;
    const pos = a.node.compareDocumentPosition(b.node);
    if (pos & Node.DOCUMENT_POSITION_FOLLOWING) return -1;
    if (pos & Node.DOCUMENT_POSITION_PRECEDING) return 1;
    return 0;
  });
}

const userTurns = Array.from(document.querySelectorAll('div[data-testid="user-message"]'))
  .map((node) => ({
    role: 'user',
    node,
    text: normalizeText(node.innerText || node.textContent || ''),
  }))
  .filter((item) => item.text);

const assistantTurns = outermostAssistantNodes()
  .map((node) => ({
    role: 'assistant',
    node,
    text: normalizeText(node.innerText || node.textContent || ''),
  }))
  .filter((item) => item.text);

const transcript = sortByDocumentOrder(userTurns.concat(assistantTurns)).map((item, index) => ({
  index,
  role: item.role,
  text: item.text,
  textLength: item.text.length,
}));

return {
  href: location.href,
  title: document.title,
  conversationTitle: getConversationTitle(),
  currentModel: parseCurrentModel(),
  turnCount: transcript.length,
  transcript,
};
