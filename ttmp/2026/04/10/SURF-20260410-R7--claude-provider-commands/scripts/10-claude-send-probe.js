function realClick(el) {
  if (!el) return;
  for (const type of ['pointerdown', 'mousedown', 'pointerup', 'mouseup', 'click']) {
    el.dispatchEvent(new MouseEvent(type, { bubbles: true, cancelable: true, view: window }));
  }
}
function setPrompt(input, text) {
  input.focus();
  const sel = window.getSelection();
  const range = document.createRange();
  range.selectNodeContents(input);
  sel.removeAllRanges();
  sel.addRange(range);
  document.execCommand('delete');
  document.execCommand('insertText', false, text);
  input.dispatchEvent(new InputEvent('input', { bubbles: true, data: text, inputType: 'insertText' }));
  input.dispatchEvent(new Event('change', { bubbles: true }));
}
const input = document.querySelector('[data-testid="chat-input"]');
if (!input) return { ok: false, error: 'chat input not found', href: location.href };
setPrompt(input, 'Please just say hello.');
const send = Array.from(document.querySelectorAll('button')).find((el) => (el.getAttribute('aria-label') || '').trim() === 'Send message');
if (!send) {
  return { ok: false, error: 'send button not found', href: location.href, inputText: input.textContent || '', inputHTML: input.innerHTML || '' };
}
realClick(send);
return {
  ok: true,
  href: location.href,
  inputText: input.textContent || '',
  inputHTML: input.innerHTML || '',
  sendDisabled: !!send.disabled || send.getAttribute('aria-disabled') === 'true',
};
