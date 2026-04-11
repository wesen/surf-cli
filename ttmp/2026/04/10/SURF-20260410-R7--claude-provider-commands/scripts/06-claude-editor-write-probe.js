const input = document.querySelector('[data-testid="chat-input"]');
if (!input) {
  return { ok: false, error: 'chat input not found', href: location.href };
}
input.focus();
input.textContent = 'Please just say hello.';
input.dispatchEvent(new InputEvent('beforeinput', { bubbles: true, cancelable: true, inputType: 'insertText', data: 'Please just say hello.' }));
input.dispatchEvent(new InputEvent('input', { bubbles: true, data: 'Please just say hello.', inputType: 'insertText' }));
input.dispatchEvent(new Event('change', { bubbles: true }));
const controls = Array.from(document.querySelectorAll('button, [role="button"], [data-testid], [aria-label]')).map((el, i) => ({
  i,
  tag: el.tagName.toLowerCase(),
  ariaLabel: el.getAttribute('aria-label'),
  dataTestid: el.getAttribute('data-testid'),
  disabled: !!el.disabled || el.getAttribute('aria-disabled') === 'true',
  text: (el.textContent || '').trim().replace(/\s+/g, ' ').slice(0, 200),
  className: (el.className || '').toString().slice(0, 180),
})).filter((x) => /send|submit|stop|claude|artifact|chat-input|model|extended|mic|voice|upload/i.test([x.ariaLabel, x.dataTestid, x.text, x.className].join(' ')));
return {
  ok: true,
  href: location.href,
  inputText: input.textContent,
  inputHTML: input.innerHTML,
  controls,
};
