const sleep = (ms) => new Promise((resolve) => setTimeout(resolve, ms));
function normalizeText(value) { return (value || '').replace(/\s+/g, ' ').trim(); }
function realClick(el) {
  el.focus();
  for (const type of ['pointerdown','mousedown','pointerup','mouseup','click']) {
    el.dispatchEvent(new MouseEvent(type, { bubbles: true, cancelable: true, view: window }));
  }
}
const input = document.querySelector('#internet-access input[type="checkbox"], label#internet-access input[type="checkbox"], input[aria-label="Web access"]');
const label = document.querySelector('#internet-access, label#internet-access');
if (!input || !label) return { ok:false, error:'web toggle not found', href: location.href };
const before = !!input.checked;
realClick(label);
await sleep(300);
return {
  ok: true,
  href: location.href,
  before,
  after: !!input.checked,
  labelText: normalizeText(label.textContent || ''),
};
