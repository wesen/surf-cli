function realClick(el) {
  if (!el) return;
  for (const type of ['pointerdown', 'mousedown', 'pointerup', 'mouseup', 'click']) {
    el.dispatchEvent(new MouseEvent(type, { bubbles: true, cancelable: true, view: window }));
  }
}
const button = document.querySelector('[data-testid="model-selector-dropdown"]');
if (!button) return { ok: false, error: 'model selector not found', href: location.href };
realClick(button);
await new Promise((resolve) => setTimeout(resolve, 400));
const options = Array.from(document.querySelectorAll('[role="menuitem"], [role="option"], [data-testid], button, [role="button"]')).map((el, i) => ({
  i,
  tag: el.tagName.toLowerCase(),
  role: el.getAttribute('role'),
  ariaLabel: el.getAttribute('aria-label'),
  dataTestid: el.getAttribute('data-testid'),
  className: (el.className || '').toString().slice(0, 180),
  text: (el.textContent || '').trim().replace(/\s+/g, ' ').slice(0, 200),
})).filter((x) => /sonnet|opus|haiku|model|extended|think|menu|claude/i.test([x.role, x.ariaLabel, x.dataTestid, x.className, x.text].join(' '))).slice(0, 200);
return {
  ok: true,
  href: location.href,
  current: (button.textContent || '').trim().replace(/\s+/g, ' '),
  options,
};
