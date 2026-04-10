const sleep = (ms) => new Promise((resolve) => setTimeout(resolve, ms));

function normalizeText(value) {
  return (value || '').replace(/\s+/g, ' ').trim();
}

async function waitFor(fn, timeoutMs = 5000, intervalMs = 100) {
  const started = Date.now();
  while (Date.now() - started < timeoutMs) {
    const value = fn();
    if (value) return { value, waitedMs: Date.now() - started };
    await sleep(intervalMs);
  }
  return null;
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
const ready = await waitFor(() => {
  return Array.from(document.querySelectorAll('dialog[open], [role="dialog"], .popover, .modal, .thread-more-menu, .tag-selector, .threads-pane, .tags-pane, form, input')).find((el) => {
    const text = normalizeText(el.textContent || '');
    const html = (el.outerHTML || '').slice(0, 500);
    return /tag|untagged|temporary|public|done|select threads/i.test(text) || /checkbox|radio|text/.test(html);
  });
}, 5000, 100);
const found = ready?.value || null;
return {
  ok: !!found,
  href: location.href,
  button: {
    text: normalizeText(button.textContent || ''),
    title: button.getAttribute('title'),
    cls: (button.className || '').toString().slice(0, 200),
  },
  waitedMs: ready?.waitedMs || null,
  foundTag: found ? found.tagName.toLowerCase() : null,
  foundClass: found ? (found.className || '').toString().slice(0, 200) : null,
  foundText: found ? normalizeText(found.textContent || '').slice(0, 2000) : null,
  foundHtml: found ? found.outerHTML.slice(0, 4000) : null,
};
