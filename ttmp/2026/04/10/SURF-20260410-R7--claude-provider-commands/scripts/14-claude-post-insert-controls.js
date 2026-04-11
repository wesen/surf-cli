const input = document.querySelector('[data-testid="chat-input"]');
if (!input || !input.editor) return { ok: false, error: 'editor not found', href: location.href };
input.editor.chain().focus().clearContent().insertContent('Please just say hello.').run();
const controls = Array.from(document.querySelectorAll('button, [role="button"], [aria-label], [data-testid]')).map((el, i) => ({
  i,
  tag: el.tagName.toLowerCase(),
  ariaLabel: el.getAttribute('aria-label'),
  role: el.getAttribute('role'),
  dataTestid: el.getAttribute('data-testid'),
  disabled: !!el.disabled || el.getAttribute('aria-disabled') === 'true',
  text: (el.textContent || '').trim().replace(/\s+/g, ' ').slice(0, 200),
  className: (el.className || '').toString().slice(0, 180),
})).filter((x) => /send|stop|upload|voice|model|claude|artifact|project|chat-input|extended|hello/i.test([x.ariaLabel, x.role, x.dataTestid, x.text, x.className].join(' ')));
return {
  ok: true,
  href: location.href,
  inputText: input.editor.getText(),
  inputHTML: input.editor.getHTML(),
  controls,
};
