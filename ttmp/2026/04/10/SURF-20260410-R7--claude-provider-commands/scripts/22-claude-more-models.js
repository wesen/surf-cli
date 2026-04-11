function realClick(el) {
  if (!el) return;
  for (const type of ['pointerdown', 'mousedown', 'pointerup', 'mouseup', 'click']) {
    el.dispatchEvent(new MouseEvent(type, { bubbles: true, cancelable: true, view: window }));
  }
}
function text(v) {
  return String(v || '').replace(/\s+/g, ' ').trim();
}
const button = document.querySelector('[data-testid="model-selector-dropdown"]');
if (!button) return { ok: false, error: 'model selector not found', href: location.href };
realClick(button);
await new Promise((resolve) => setTimeout(resolve, 300));
const items = Array.from(document.querySelectorAll('[role="menuitem"]'));
const more = items.find((el) => text(el.innerText || el.textContent).startsWith('More models'));
if (!more) {
  return {
    ok: true,
    href: location.href,
    current: text(button.innerText || button.textContent),
    topLevel: items.map((el) => text(el.innerText || el.textContent)),
    submenu: [],
  };
}
realClick(more);
await new Promise((resolve) => setTimeout(resolve, 500));
const all = Array.from(document.querySelectorAll('[role="menuitem"], [role="option"], button, [role="button"]')).map((el, i) => ({
  i,
  role: el.getAttribute('role'),
  ariaLabel: el.getAttribute('aria-label'),
  dataTestid: el.getAttribute('data-testid'),
  text: text(el.innerText || el.textContent),
  className: (el.className || '').toString().slice(0, 180),
})).filter((x) => /sonnet|opus|haiku|more models|thinking|extended|claude|model/i.test([x.role, x.ariaLabel, x.dataTestid, x.text, x.className].join(' ')));
return {
  ok: true,
  href: location.href,
  current: text(button.innerText || button.textContent),
  topLevel: items.map((el) => text(el.innerText || el.textContent)),
  submenu: all,
};
