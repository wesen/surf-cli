const turns = Array.from(document.querySelectorAll('[data-message-author-role], [data-turn="assistant"], [data-turn="user"]'));
const normalized = turns.map((node, index) => {
  const role = node.getAttribute('data-message-author-role') || node.getAttribute('data-turn') || 'unknown';
  const messageId = node.getAttribute('data-message-id') || node.getAttribute('data-turn-id') || null;
  const model = node.getAttribute('data-message-model-slug') || null;
  const text = (node.innerText || '').trim();
  return {
    index,
    role,
    messageId,
    model,
    textLength: text.length,
    preview: text.slice(0, 180),
  };
});
return {
  href: location.href,
  title: document.title,
  turnCount: normalized.length,
  turns: normalized,
};
