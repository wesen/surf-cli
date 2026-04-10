const sleep = (ms) => new Promise((resolve) => setTimeout(resolve, ms));
function normalizeText(value) { return (value || '').replace(/\s+/g, ' ').trim(); }
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
  el.focus();
  for (const type of ['pointerdown','mousedown','pointerup','mouseup','click']) {
    el.dispatchEvent(new MouseEvent(type, { bubbles: true, cancelable: true, view: window }));
  }
}
const btn = document.querySelector('#lens-select');
if (!btn) return { ok:false, error:'lens-select not found', href: location.href };
realClick(btn);
const ready = await waitFor(() => document.querySelector('ul[role="listbox"][aria-label="Lens chooser"]'), 5000, 100);
const listbox = ready?.value || null;
return {
  ok: !!listbox,
  href: location.href,
  button: {
    text: normalizeText(btn.textContent || ''),
    title: btn.getAttribute('title'),
    ariaExpanded: btn.getAttribute('aria-expanded'),
  },
  waitedMs: ready?.waitedMs || null,
  options: listbox ? Array.from(listbox.querySelectorAll('li.option[role="option"]')).map((el) => ({
    text: normalizeText(el.textContent || ''),
    selected: el.getAttribute('aria-selected') === 'true',
  })) : [],
};
