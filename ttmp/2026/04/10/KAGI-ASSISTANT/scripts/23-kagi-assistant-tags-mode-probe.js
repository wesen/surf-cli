const sleep = (ms) => new Promise((resolve) => setTimeout(resolve, ms));

function normalizeText(value) {
  return (value || '').replace(/\s+/g, ' ').trim();
}

function realClick(el) {
  if (!el) return;
  el.focus();
  for (const type of ['pointerdown', 'mousedown', 'pointerup', 'mouseup', 'click']) {
    el.dispatchEvent(new MouseEvent(type, { bubbles: true, cancelable: true, view: window }));
  }
}

const button = document.querySelector('#tags-add');
if (!button) {
  return { ok: false, error: 'tags-add button not found', href: location.href };
}
realClick(button);
await sleep(500);

const interesting = Array.from(document.querySelectorAll('input, button, label, div')).map((el, i) => ({
  i,
  tag: el.tagName.toLowerCase(),
  id: el.id || null,
  type: el.getAttribute('type'),
  hidden: el.hidden,
  ariaHidden: el.getAttribute('aria-hidden'),
  checked: el.checked === undefined ? null : !!el.checked,
  cls: (el.className || '').toString().slice(0, 200),
  text: normalizeText(el.textContent || '').slice(0, 240),
})).filter((x) => /tag|select threads|done|untagged|temporary|public|all|thread-bulk-ops|edit tags/i.test((x.id || '') + ' ' + (x.cls || '') + ' ' + (x.text || '')));

return {
  ok: true,
  href: location.href,
  title: document.title,
  interesting: interesting.slice(0, 200),
};
