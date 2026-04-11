function realClick(el) {
  if (!el) return;
  for (const type of ['pointerdown', 'mousedown', 'pointerup', 'mouseup', 'click']) {
    el.dispatchEvent(new MouseEvent(type, { bubbles: true, cancelable: true, view: window }));
  }
}
function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}
function summarize() {
  return {
    href: location.href,
    bodyPreview: (document.body?.innerText || '').slice(0, 2200),
  };
}
const input = document.querySelector('[data-testid="chat-input"]');
if (!input || !input.editor) return { ok: false, error: 'editor not found', ...summarize() };
input.editor.chain().focus().clearContent().insertContent('Please just say hello.').run();
await sleep(200);
const send = Array.from(document.querySelectorAll('button')).find((el) => (el.getAttribute('aria-label') || '').trim() === 'Send message');
if (!send) return { ok: false, error: 'send button not found', inputText: input.editor.getText(), inputHTML: input.editor.getHTML(), ...summarize() };
realClick(send);
await sleep(5000);
return {
  ok: true,
  inputText: input.editor.getText(),
  inputHTML: input.editor.getHTML(),
  ...summarize(),
};
