const nodes = Array.from(document.querySelectorAll('[data-message-author-role]'));
const byMessageId = new Map();
for (const node of nodes) {
  const role = node.getAttribute('data-message-author-role') || 'unknown';
  const messageId = node.getAttribute('data-message-id') || role + ':' + byMessageId.size;
  const model = node.getAttribute('data-message-model-slug') || null;
  const text = (node.innerText || '').trim();
  if (!text) {
    continue;
  }
  const existing = byMessageId.get(messageId);
  if (!existing || text.length > existing.text.length) {
    byMessageId.set(messageId, {
      messageId,
      role,
      model,
      text,
      textLength: text.length,
    });
  }
}
const transcript = Array.from(byMessageId.values()).map((item, index) => ({
  index,
  role: item.role,
  model: item.model,
  messageId: item.messageId,
  textLength: item.textLength,
  text: item.text,
}));
return {
  href: location.href,
  title: document.title,
  turnCount: transcript.length,
  transcript,
};
