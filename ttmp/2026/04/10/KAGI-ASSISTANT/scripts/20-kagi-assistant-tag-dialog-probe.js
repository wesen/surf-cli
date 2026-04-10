const sleep = (ms) => new Promise((resolve) => setTimeout(resolve, ms));

function normalizeText(value) {
  return (value || '').replace(/\s+/g, ' ').trim();
}

async function waitFor(fn, timeoutMs = 5000, intervalMs = 100) {
  const started = Date.now();
  while (Date.now() - started < timeoutMs) {
    const value = fn();
    if (value) {
      return { value, waitedMs: Date.now() - started };
    }
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

const editTrigger = Array.from(document.querySelectorAll('button, a, label, div, span')).find((el) => /edit tags/i.test(normalizeText(el.textContent || '')));
if (!editTrigger) {
  return {
    ok: false,
    error: 'Edit tags trigger not found',
    href: location.href,
  };
}
realClick(editTrigger);
const ready = await waitFor(() => Array.from(document.querySelectorAll('dialog[open], [role="dialog"], .modal, .popover')).find((el) => /tag|untagged|temporary|public/i.test(normalizeText(el.textContent || ''))), 5000, 100);
const popup = ready?.value || null;
return {
  ok: !!popup,
  href: location.href,
  trigger: {
    tag: editTrigger.tagName.toLowerCase(),
    text: normalizeText(editTrigger.textContent || ''),
    cls: (editTrigger.className || '').toString().slice(0, 200),
  },
  waitedMs: ready?.waitedMs || null,
  popupText: popup ? normalizeText(popup.textContent || '').slice(0, 4000) : null,
  popupHtml: popup ? popup.outerHTML.slice(0, 4000) : null,
};
