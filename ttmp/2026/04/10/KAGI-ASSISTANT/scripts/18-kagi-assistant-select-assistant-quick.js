const sleep = (ms) => new Promise((resolve) => setTimeout(resolve, ms));
const targetAssistant = 'Quick';
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
const btn = document.querySelector('#profile-select');
if (!btn) return { ok:false, error:'profile-select not found', href: location.href };
realClick(btn);
const ready = await waitFor(() => document.querySelector('dialog.promptOptionsSelector[open]'), 5000, 100);
const dialog = ready?.value;
if (!dialog) return { ok:false, error:'dialog not open', href: location.href };
const match = Array.from(dialog.querySelectorAll('li.option[role="option"]')).find((el) => normalizeText(el.textContent || '').startsWith(targetAssistant));
if (!match) return { ok:false, error:'target assistant not found', targetAssistant, href: location.href };
const before = normalizeText(btn.textContent || '');
realClick(match);
await sleep(1200);
return {
  ok:true,
  href: location.href,
  targetAssistant,
  selectedText: normalizeText(match.textContent || ''),
  before,
  after: normalizeText(btn.textContent || ''),
  dialogStillOpen: !!document.querySelector('dialog.promptOptionsSelector[open]'),
};
