const input = document.querySelector('[data-testid="chat-input"]');
if (!input) {
  return { ok: false, error: 'chat input not found', href: location.href };
}
const snapshot = Array.from(document.querySelectorAll('button, [role="button"], [data-testid], [aria-label]')).map((el, i) => ({
  i,
  tag: el.tagName.toLowerCase(),
  ariaLabel: el.getAttribute('aria-label'),
  dataTestid: el.getAttribute('data-testid'),
  disabled: !!el.disabled || el.getAttribute('aria-disabled') === 'true',
  text: (el.textContent || '').trim().replace(/\s+/g, ' ').slice(0, 160),
  className: (el.className || '').toString().slice(0, 120),
})).filter((x) => /send|stop|upload|voice|model|claude|project|artifact|chat-input/i.test([x.ariaLabel, x.dataTestid, x.text, x.className].join(' ')));
return {
  ok: true,
  href: location.href,
  inputText: input.textContent || '',
  inputHTML: input.innerHTML || '',
  controls: snapshot,
};
