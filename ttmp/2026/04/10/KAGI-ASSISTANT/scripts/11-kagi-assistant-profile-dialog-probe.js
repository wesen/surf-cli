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
const btn = document.querySelector('#profile-select');
if (!btn) return { ok:false, error:'profile-select not found', href: location.href };
const before = {
  text: normalizeText(btn.textContent || ''),
  ariaExpanded: btn.getAttribute('aria-expanded'),
};
btn.click();
const ready = await waitFor(() => document.querySelector('dialog.promptOptionsSelector[open]'), 5000, 100);
const dialog = ready?.value || null;
return {
  ok: !!dialog,
  href: location.href,
  before,
  after: {
    text: normalizeText(btn.textContent || ''),
    ariaExpanded: btn.getAttribute('aria-expanded'),
  },
  waitedMs: ready?.waitedMs || null,
  dialogText: dialog ? normalizeText(dialog.textContent || '').slice(0, 4000) : null,
  dialogHTML: dialog ? dialog.outerHTML.slice(0, 4000) : null,
};
