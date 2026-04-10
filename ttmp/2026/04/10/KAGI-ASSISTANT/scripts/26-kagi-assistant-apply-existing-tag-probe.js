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

function visibleTags() {
  return Array.from(document.querySelectorAll('#tags .untagged, #tags [class*="tag"], #tags .dialog-tag')).map((el) => normalizeText(el.textContent || '')).filter(Boolean);
}

realClick(document.querySelector('#tags-add'));
await sleep(500);
const dialog = document.querySelector('#tags dialog.promptOptionsSelector');
const label = Array.from(dialog?.querySelectorAll('label') || []).find((el) => normalizeText(el.textContent || '').toLowerCase() === 'temporary');
if (!dialog || !label) {
  return { ok: false, error: 'temporary tag label not found', href: location.href, visibleTags: visibleTags() };
}
const input = label.querySelector('input[type="checkbox"]');
const before = {
  checked: !!input?.checked,
  visibleTags: visibleTags(),
};
realClick(label);
await sleep(700);
return {
  ok: true,
  href: location.href,
  before,
  after: {
    checked: !!input?.checked,
    visibleTags: visibleTags(),
  },
};
