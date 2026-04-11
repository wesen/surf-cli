function realClick(el) {
  if (!el) return;
  for (const type of ['pointerdown', 'mousedown', 'pointerup', 'mouseup', 'click']) {
    el.dispatchEvent(new MouseEvent(type, { bubbles: true, cancelable: true, view: window }));
  }
}
const input = document.querySelector('[data-testid="chat-input"]');
if (!input || !input.editor) {
  return { ok: false, error: 'editor not found', href: location.href };
}
const prompt = 'Please just say hello.';
input.editor.chain().focus().clearContent().insertContent(prompt).run();
const send = Array.from(document.querySelectorAll('button')).find((el) => (el.getAttribute('aria-label') || '').trim() === 'Send message');
if (!send) {
  return {
    ok: false,
    error: 'send button not found',
    href: location.href,
    inputText: input.editor.getText(),
    inputHTML: input.editor.getHTML(),
  };
}
realClick(send);
return {
  ok: true,
  href: location.href,
  inputText: input.editor.getText(),
  inputHTML: input.editor.getHTML(),
  sendDisabled: !!send.disabled || send.getAttribute('aria-disabled') === 'true',
};
